package controller

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// signPlisioBody mirrors Plisio's documented Node signing for json=true
// callbacks: JSON.stringify(body without verify_hash), HMAC-SHA1 with the
// secret key, hex-encoded.
func signPlisioBody(t *testing.T, bodyWithoutHash string, secret string) string {
	t.Helper()
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(bodyWithoutHash))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyPlisioSignature_AcceptsGenuineCallback(t *testing.T) {
	const secret = "plisio-secret-key"
	// Canonical (signing) payload: compact JSON, verify_hash omitted, fields in
	// the same order Plisio sends them. Plisio echoes order_number back so we
	// can match the local order.
	canonical := `{"txn_id":"abc123","ipn_type":"invoice","order_number":"PLSUSR1NOxyz1","order_name":"Top-up #5","status":"completed","amount":"0.0001","currency":"BTC","source_currency":"USD","source_amount":"5.00"}`
	sig := signPlisioBody(t, canonical, secret)

	// The delivered body is the canonical payload plus verify_hash appended.
	delivered := canonical[:len(canonical)-1] + `,"verify_hash":"` + sig + `"}`

	require.True(t, verifyPlisioSignature([]byte(delivered), sig, secret))
}

func TestVerifyPlisioSignature_RejectsTamperedAmount(t *testing.T) {
	const secret = "plisio-secret-key"
	canonical := `{"order_number":"PLSUSR1NOxyz1","status":"completed","source_amount":"5.00"}`
	sig := signPlisioBody(t, canonical, secret)

	tampered := `{"order_number":"PLSUSR1NOxyz1","status":"completed","source_amount":"5000.00","verify_hash":"` + sig + `"}`
	require.False(t, verifyPlisioSignature([]byte(tampered), sig, secret))
}

func TestVerifyPlisioSignature_RejectsBadSecret(t *testing.T) {
	canonical := `{"order_number":"PLSUSR1NOxyz1","status":"completed"}`
	sig := signPlisioBody(t, canonical, "real-secret")
	delivered := canonical[:len(canonical)-1] + `,"verify_hash":"` + sig + `"}`

	require.False(t, verifyPlisioSignature([]byte(delivered), sig, "wrong-secret"))
}

// The provider mismatch guard is enforced in PlisioWebhook; here we only verify
// the signature primitive handles an empty secret/hash defensively.
func TestVerifyPlisioSignature_RejectsEmptyInputs(t *testing.T) {
	require.False(t, verifyPlisioSignature([]byte(`{}`), "hash", ""))
	require.False(t, verifyPlisioSignature([]byte(`{}`), "", "secret"))
	require.False(t, verifyPlisioSignature([]byte(`{}`), "", ""))
}

func TestPlisioCanonicalBody_DropsVerifyHashAnyPosition(t *testing.T) {
	// verify_hash in the middle of the object must still be removed while every
	// other value keeps its original bytes and order.
	input := `{"a":"1","verify_hash":"zzz","b":"2","c":"3"}`
	got, err := plisioCanonicalBody([]byte(input))
	require.NoError(t, err)
	assert.Equal(t, `{"a":"1","b":"2","c":"3"}`, got)

	// verify_hash last (the common case).
	inputLast := `{"a":"1","b":"2","verify_hash":"zzz"}`
	got, err = plisioCanonicalBody([]byte(inputLast))
	require.NoError(t, err)
	assert.Equal(t, `{"a":"1","b":"2"}`, got)
}

func TestPlisioCanonicalBody_PreservesValueBytes(t *testing.T) {
	// Numbers and nested structures are preserved byte-for-byte (matches
	// JSON.stringify semantics on the receiver side).
	input := `{"n":42,"f":3.14,"arr":[1,2,3],"obj":{"k":"v"}}`
	got, err := plisioCanonicalBody([]byte(input))
	require.NoError(t, err)
	assert.Equal(t, input, got)
}

// Use the shared dialect fixture so these settlement invariants also run on
// real MySQL and PostgreSQL, including a separate log database.
func TestPlisioWebhookSettlement(t *testing.T) {
	for _, tc := range []struct {
		name, provider, status, amount string
		quota                          int
		storedAmount                   int64
		money                          float64
		missingUser, invalidSignature  bool
		wantStatus                     int
		wantCredit                     bool
	}{
		{name: "completed and duplicate", provider: "plisio", status: "completed", amount: `"5.00"`, storedAmount: 5, wantStatus: 200, wantCredit: true},
		{name: "invoice cent rounding", provider: "plisio", status: "completed", amount: `"5.00"`, storedAmount: 5, money: 5.005, wantStatus: 200, wantCredit: true},
		{name: "numeric amount", provider: "plisio", status: "completed", amount: `5`, storedAmount: 5, wantStatus: 200, wantCredit: true},
		{name: "amount mismatch", provider: "plisio", status: "completed", amount: `"4.00"`, storedAmount: 5, wantStatus: 400},
		{name: "missing amount", provider: "plisio", status: "completed", amount: `null`, storedAmount: 5, wantStatus: 400},
		{name: "wrong provider", provider: "epay", status: "completed", amount: `"5.00"`, storedAmount: 5, wantStatus: 400},
		{name: "pending invoice", provider: "plisio", status: "pending", amount: `"5.00"`, storedAmount: 5, wantStatus: 200},
		{name: "invalid signature", provider: "plisio", status: "completed", amount: `"5.00"`, storedAmount: 5, invalidSignature: true, wantStatus: 401},
		{name: "wallet limit rolls back completion", provider: "plisio", status: "completed", amount: `"5.00"`, storedAmount: 5, quota: common.MaxWalletQuota, wantStatus: 500},
		{name: "overflow rolls back completion", provider: "plisio", status: "completed", amount: `"5.00"`, storedAmount: math.MaxInt64, wantStatus: 500},
		{name: "missing user rolls back completion", provider: "plisio", status: "completed", amount: `"5.00"`, storedAmount: 5, missingUser: true, wantStatus: 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := setupManageUserTestDB(t)
			require.NoError(t, db.AutoMigrate(&model.TopUp{}))
			oldKey := setting.PlisioApiKey
			setting.PlisioApiKey = "qa-local-plisio-secret"
			payment := operation_setting.GetPaymentSetting()
			oldPayment := *payment
			payment.ComplianceConfirmed = true
			payment.ComplianceTermsVersion = operation_setting.CurrentComplianceTermsVersion
			t.Cleanup(func() { setting.PlisioApiKey = oldKey; *payment = oldPayment })
			user := model.User{Username: "qa-plisio", Password: "inert-fixture", AffCode: "qa-plisio", Quota: tc.quota}
			require.NoError(t, db.Create(&user).Error)
			order := model.TopUp{UserId: user.Id, TradeNo: "qa-plisio-order", Amount: tc.storedAmount, Money: 5, PaymentMethod: "plisio", PaymentProvider: tc.provider, Status: common.TopUpStatusPending}
			if tc.money != 0 {
				order.Money = tc.money
			}
			require.NoError(t, db.Create(&order).Error)
			if tc.missingUser {
				require.NoError(t, db.Unscoped().Delete(&user).Error)
			}
			canonical := fmt.Sprintf(`{"txn_id":"qa-txn","order_number":"qa-plisio-order","status":%q,"source_amount":%s}`, tc.status, tc.amount)
			sig := signPlisioBody(t, canonical, setting.PlisioApiKey)
			if tc.invalidSignature {
				sig = "invalid"
			}
			body := canonical[:len(canonical)-1] + `,"verify_hash":"` + sig + `"}`
			request := func() int {
				recorder := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(recorder)
				ctx.Request = httptest.NewRequest(http.MethodPost, "/api/plisio/webhook", strings.NewReader(body))
				PlisioWebhook(ctx)
				return recorder.Code
			}
			if tc.name == "completed and duplicate" {
				sqlDB, err := db.DB()
				require.NoError(t, err)
				if db.Dialector.Name() != "sqlite" {
					sqlDB.SetMaxOpenConns(4)
				}
				start := make(chan struct{})
				failures := make(chan error, 2)
				var workers sync.WaitGroup
				for range 2 {
					workers.Add(1)
					go func() {
						defer workers.Done()
						<-start
						_, err := model.RechargePlisio(order.TradeNo, decimal.NewFromInt(5), "127.0.0.1")
						failures <- err
					}()
				}
				close(start)
				workers.Wait()
				close(failures)
				for err := range failures {
					if db.Dialector.Name() == "sqlite" && err != nil {
						// SQLite may reject a competing writer; the subsequent callback
						// retries after rollback and must still credit exactly once.
						require.ErrorContains(t, err, "SQLITE_BUSY")
					} else {
						require.NoError(t, err)
					}
				}
			}
			require.Equal(t, tc.wantStatus, request())
			require.NoError(t, db.First(&order, order.Id).Error)
			if tc.wantCredit {
				require.Equal(t, common.TopUpStatusSuccess, order.Status)
				require.Equal(t, 200, request())
				require.NoError(t, db.First(&user, user.Id).Error)
				require.Equal(t, tc.quota+int(5*common.QuotaPerUnit), user.Quota)
				var count int64
				require.NoError(t, model.LOG_DB.Model(&model.Log{}).Where("user_id = ?", user.Id).Count(&count).Error)
				require.EqualValues(t, 1, count)
			} else {
				require.Equal(t, common.TopUpStatusPending, order.Status)
				require.Zero(t, order.CompleteTime)
				if !tc.missingUser {
					require.NoError(t, db.First(&user, user.Id).Error)
					require.Equal(t, tc.quota, user.Quota)
				}
			}
		})
	}
}

type plisioInvoiceTestTransport struct {
	body string
	err  error
}

func (transport plisioInvoiceTestTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if transport.err != nil {
		return nil, transport.err
	}
	return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(transport.body)), Request: request}, nil
}
func TestPlisioInvoiceErrorsDoNotExposeCredentials(t *testing.T) {
	oldKey, oldClient := setting.PlisioApiKey, plisioHTTPClient
	setting.PlisioApiKey = "qa-credential-that-must-not-be-logged"
	t.Cleanup(func() { setting.PlisioApiKey = oldKey; plisioHTTPClient = oldClient })
	for _, transport := range []plisioInvoiceTestTransport{{err: errors.New("unavailable")}, {body: `{"status":"error","data":{"message":"qa-credential-that-must-not-be-logged"}}`}} {
		plisioHTTPClient = &http.Client{Transport: transport}
		_, err := createPlisioInvoice(context.Background(), "qa-order", "qa-invoice", "5.00", "USD", "")
		require.Error(t, err)
		assert.NotContains(t, err.Error(), setting.PlisioApiKey)
		assert.NotContains(t, err.Error(), "api_key")
	}
}
func TestPlisioCheckoutRejectsUnsafeQuotaBeforeInvoice(t *testing.T) {
	db := setupManageUserTestDB(t)
	user := model.User{Username: "qa-plisio-limit", AffCode: "qa-plisio-limit", Quota: common.MaxWalletQuota}
	require.NoError(t, db.Create(&user).Error)
	for _, handler := range []gin.HandlerFunc{RequestPlisioAmount, RequestPlisioPay} {
		for _, amount := range []int64{getPlisioMinTopup(), math.MaxInt64} {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/api/user/plisio/pay", strings.NewReader(fmt.Sprintf(`{"amount":%d,"payment_method":"plisio"}`, amount)))
			ctx.Request.Header.Set("Content-Type", "application/json")
			ctx.Set("id", user.Id)
			handler(ctx)
			assert.Contains(t, recorder.Body.String(), `"message":"error"`)
		}
	}
}
