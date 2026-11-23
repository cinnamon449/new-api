package controller

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
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
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const nowPaymentsTestSecret = "qa-local-nowpayments-ipn-secret"

// signNowPaymentsBody mirrors NowPayments' documented signing: HMAC-SHA512 of
// JSON.stringify(sortObject(body)) with the IPN secret, hex-encoded.
func signNowPaymentsBody(sortedBody string, secret string) string {
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write([]byte(sortedBody))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyNowPaymentsSignature(t *testing.T) {
	sorted := `{"actually_paid":5,"order_id":"qa-order","payment_id":123,"payment_status":"finished","price_amount":5}`
	unordered := `{"payment_status":"finished","price_amount":5,"order_id":"qa-order","payment_id":123,"actually_paid":5}`
	sig := signNowPaymentsBody(sorted, nowPaymentsTestSecret)

	assert.True(t, verifyNowPaymentsSignature([]byte(unordered), sig, nowPaymentsTestSecret), "field order in the delivered body must not matter")
	assert.True(t, verifyNowPaymentsSignature([]byte(unordered), strings.ToUpper(sig), nowPaymentsTestSecret))
	assert.False(t, verifyNowPaymentsSignature([]byte(strings.Replace(unordered, `"price_amount":5`, `"price_amount":50`, 1)), sig, nowPaymentsTestSecret), "tampered amount")
	assert.False(t, verifyNowPaymentsSignature([]byte(unordered), sig, "another-secret"))
	assert.False(t, verifyNowPaymentsSignature([]byte(unordered), "", nowPaymentsTestSecret))
	assert.False(t, verifyNowPaymentsSignature([]byte(unordered), sig, ""))
	assert.False(t, verifyNowPaymentsSignature([]byte(`not json`), sig, nowPaymentsTestSecret))
}

// Crypto transfers rarely land on the exact invoice amount. UnoRouter found
// verified-paid orders stuck pending on 0.25% shortfalls; these cases come from
// its regression table.
func TestWithinUnderpaymentTolerance(t *testing.T) {
	for _, tc := range []struct {
		name     string
		pay      float64
		actual   float64
		expected bool
	}{
		{name: "tiny shortfall credits", pay: 20.20, actual: 20.183446, expected: true},
		{name: "overpayment credits", pay: 20.00, actual: 20.50, expected: true},
		{name: "exact amount credits", pay: 20.00, actual: 20.00, expected: true},
		{name: "shortfall at the 2 percent edge credits", pay: 100.00, actual: 98.00, expected: true},
		{name: "shortfall beyond tolerance stays pending", pay: 100.00, actual: 97.99, expected: false},
		{name: "half payment stays pending", pay: 20.00, actual: 10.00, expected: false},
		{name: "zero paid stays pending", pay: 20.00, actual: 0, expected: false},
		{name: "zero invoice stays pending", pay: 0, actual: 20.00, expected: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, withinUnderpaymentTolerance(&nowPaymentsPaymentEvent{PayAmount: tc.pay, ActuallyPaid: tc.actual}))
		})
	}
	assert.False(t, withinUnderpaymentTolerance(nil))
}

func enableNowPaymentsForTest(t *testing.T) {
	t.Helper()
	old := []any{setting.NowPaymentsEnabled, setting.NowPaymentsApiKey, setting.NowPaymentsIpnSecret}
	setting.NowPaymentsEnabled = true
	setting.NowPaymentsApiKey = "qa-local-nowpayments-api-key"
	setting.NowPaymentsIpnSecret = nowPaymentsTestSecret
	payment := operation_setting.GetPaymentSetting()
	oldPayment := *payment
	payment.ComplianceConfirmed = true
	payment.ComplianceTermsVersion = operation_setting.CurrentComplianceTermsVersion
	t.Cleanup(func() {
		setting.NowPaymentsEnabled = old[0].(bool)
		setting.NowPaymentsApiKey = old[1].(string)
		setting.NowPaymentsIpnSecret = old[2].(string)
		*payment = oldPayment
	})
}

type nowPaymentsFixture struct {
	db    *gorm.DB
	user  model.User
	order model.TopUp
}

func newNowPaymentsFixture(t *testing.T, quota int, storedAmount int64, provider string) nowPaymentsFixture {
	t.Helper()
	db := setupManageUserTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.TopUp{}, &model.NowPaymentsPayment{}))
	enableNowPaymentsForTest(t)
	user := model.User{Username: "qa-nowpayments", Password: "inert-fixture", AffCode: "qa-nowpayments", Quota: quota}
	require.NoError(t, db.Create(&user).Error)
	order := model.TopUp{UserId: user.Id, TradeNo: "qa-nowpayments-order", Amount: storedAmount, Money: 5, PaymentMethod: model.PaymentMethodNowPayments, PaymentProvider: provider, CreateTime: time.Now().Unix(), Status: common.TopUpStatusPending}
	require.NoError(t, model.CreateNowPaymentsTopUp(&order))
	return nowPaymentsFixture{db: db, user: user, order: order}
}

func nowPaymentsEventBody(orderId string, status string, price string, payAmount float64, actuallyPaid float64) string {
	return fmt.Sprintf(`{"actually_paid":%v,"order_id":%q,"pay_amount":%v,"pay_currency":"usdttrc20","payment_id":4242,"payment_status":%q,"price_amount":%s,"price_currency":"usd"}`, actuallyPaid, orderId, payAmount, status, price)
}

func postNowPaymentsWebhook(body string, sig string) int {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/nowpayments/webhook", strings.NewReader(body))
	ctx.Request.Header.Set(nowPaymentsSignatureHeader, sig)
	NowPaymentsWebhook(ctx)
	return recorder.Code
}

// Use the shared dialect fixture so these settlement invariants also run on
// real MySQL and PostgreSQL, including a separate log database.
func TestNowPaymentsWebhookSettlement(t *testing.T) {
	for _, tc := range []struct {
		name, provider, status, price string
		payAmount, actuallyPaid       float64
		quota                         int
		storedAmount                  int64
		invalidSignature              bool
		wantStatus                    int
		wantOrderStatus               string
	}{
		{name: "finished and duplicate", provider: "nowpayments", status: "finished", price: `5`, payAmount: 5, actuallyPaid: 5, storedAmount: 5, wantStatus: 200, wantOrderStatus: common.TopUpStatusSuccess},
		{name: "string price", provider: "nowpayments", status: "finished", price: `"5.00"`, payAmount: 5, actuallyPaid: 5, storedAmount: 5, wantStatus: 200, wantOrderStatus: common.TopUpStatusSuccess},
		{name: "partial within tolerance", provider: "nowpayments", status: "partially_paid", price: `5`, payAmount: 20.2, actuallyPaid: 20.183446, storedAmount: 5, wantStatus: 200, wantOrderStatus: common.TopUpStatusSuccess},
		{name: "partial beyond tolerance", provider: "nowpayments", status: "partially_paid", price: `5`, payAmount: 20, actuallyPaid: 10, storedAmount: 5, wantStatus: 200, wantOrderStatus: common.TopUpStatusPending},
		{name: "price mismatch", provider: "nowpayments", status: "finished", price: `4`, payAmount: 4, actuallyPaid: 4, storedAmount: 5, wantStatus: 400, wantOrderStatus: common.TopUpStatusPending},
		{name: "missing price", provider: "nowpayments", status: "finished", price: `null`, payAmount: 5, actuallyPaid: 5, storedAmount: 5, wantStatus: 400, wantOrderStatus: common.TopUpStatusPending},
		{name: "wrong provider", provider: "epay", status: "finished", price: `5`, payAmount: 5, actuallyPaid: 5, storedAmount: 5, wantStatus: 400, wantOrderStatus: common.TopUpStatusPending},
		{name: "waiting", provider: "nowpayments", status: "waiting", price: `5`, payAmount: 5, storedAmount: 5, wantStatus: 200, wantOrderStatus: common.TopUpStatusPending},
		{name: "expired", provider: "nowpayments", status: "expired", price: `5`, payAmount: 5, storedAmount: 5, wantStatus: 200, wantOrderStatus: common.TopUpStatusFailed},
		{name: "invalid signature", provider: "nowpayments", status: "finished", price: `5`, payAmount: 5, actuallyPaid: 5, storedAmount: 5, invalidSignature: true, wantStatus: 401, wantOrderStatus: common.TopUpStatusPending},
		{name: "wallet limit rolls back completion", provider: "nowpayments", status: "finished", price: `5`, payAmount: 5, actuallyPaid: 5, storedAmount: 5, quota: common.MaxWalletQuota, wantStatus: 500, wantOrderStatus: common.TopUpStatusPending},
		{name: "overflow rolls back completion", provider: "nowpayments", status: "finished", price: `5`, payAmount: 5, actuallyPaid: 5, storedAmount: math.MaxInt64, wantStatus: 500, wantOrderStatus: common.TopUpStatusPending},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newNowPaymentsFixture(t, tc.quota, tc.storedAmount, tc.provider)
			body := nowPaymentsEventBody(f.order.TradeNo, tc.status, tc.price, tc.payAmount, tc.actuallyPaid)
			sig := signNowPaymentsBody(body, nowPaymentsTestSecret)
			if tc.invalidSignature {
				sig = "invalid"
			}
			if tc.name == "finished and duplicate" {
				sqlDB, err := f.db.DB()
				require.NoError(t, err)
				if f.db.Dialector.Name() != "sqlite" {
					sqlDB.SetMaxOpenConns(4)
				}
				start := make(chan struct{})
				failures := make(chan error, 2)
				var workers sync.WaitGroup
				for range 2 {
					workers.Go(func() {
						<-start
						_, err := model.RechargeNowPayments(f.order.TradeNo, decimal.NewFromInt(5), "127.0.0.1")
						failures <- err
					})
				}
				close(start)
				workers.Wait()
				close(failures)
				for err := range failures {
					if f.db.Dialector.Name() == "sqlite" && err != nil {
						// SQLite may reject a competing writer; the webhook below
						// retries after rollback and must still credit exactly once.
						require.ErrorContains(t, err, "SQLITE_BUSY")
					} else {
						require.NoError(t, err)
					}
				}
			}

			require.Equal(t, tc.wantStatus, postNowPaymentsWebhook(body, sig))
			require.NoError(t, f.db.First(&f.order, f.order.Id).Error)
			assert.Equal(t, tc.wantOrderStatus, f.order.Status)
			require.NoError(t, f.db.First(&f.user, f.user.Id).Error)

			if tc.wantOrderStatus != common.TopUpStatusSuccess {
				assert.Equal(t, tc.quota, f.user.Quota)
				return
			}
			// A replayed IPN is acknowledged without crediting again.
			require.Equal(t, 200, postNowPaymentsWebhook(body, sig))
			require.NoError(t, f.db.First(&f.user, f.user.Id).Error)
			assert.Equal(t, tc.quota+int(5*common.QuotaPerUnit), f.user.Quota)
			var count int64
			require.NoError(t, model.LOG_DB.Model(&model.Log{}).Where("user_id = ?", f.user.Id).Count(&count).Error)
			assert.EqualValues(t, 1, count)
			var topupLog model.Log
			require.NoError(t, model.LOG_DB.Where("user_id = ?", f.user.Id).First(&topupLog).Error)
			assert.Equal(t, model.LogTypeTopup, topupLog.Type)
			assert.Contains(t, topupLog.Content, "NowPayments")
		})
	}
}

func TestNowPaymentsWebhookRecordsProgressMonotonically(t *testing.T) {
	f := newNowPaymentsFixture(t, 0, 5, model.PaymentProviderNowPayments)
	for _, paid := range []float64{3, 1} {
		body := nowPaymentsEventBody(f.order.TradeNo, "confirming", `5`, 5, paid)
		require.Equal(t, 200, postNowPaymentsWebhook(body, signNowPaymentsBody(body, nowPaymentsTestSecret)))
	}
	var payment model.NowPaymentsPayment
	require.NoError(t, f.db.Where("trade_no = ?", f.order.TradeNo).First(&payment).Error)
	assert.Equal(t, "4242", payment.PaymentId)
	assert.Equal(t, 3.0, payment.PaidAmount)
}

// Turning the gateway off stops new invoices, never settlement of invoices
// already issued; only a missing IPN secret refuses callbacks.
func TestNowPaymentsWebhookSettlesWhileGatewayDisabled(t *testing.T) {
	f := newNowPaymentsFixture(t, 0, 5, model.PaymentProviderNowPayments)
	setting.NowPaymentsEnabled = false
	payment := operation_setting.GetPaymentSetting()
	payment.ComplianceConfirmed = false
	body := nowPaymentsEventBody(f.order.TradeNo, "finished", `5`, 5, 5)
	sig := signNowPaymentsBody(body, nowPaymentsTestSecret)
	require.Equal(t, http.StatusOK, postNowPaymentsWebhook(body, sig))
	require.NoError(t, f.db.First(&f.order, f.order.Id).Error)
	assert.Equal(t, common.TopUpStatusSuccess, f.order.Status)

	setting.NowPaymentsIpnSecret = ""
	assert.Equal(t, http.StatusForbidden, postNowPaymentsWebhook(body, sig))
}

// One invoice can carry several payment attempts. An expired first attempt
// fails the order, but a later attempt must reopen and settle it, and a late
// IPN for the dead attempt must not displace the live payment id.
func TestNowPaymentsRetryAfterExpiredAttemptSettles(t *testing.T) {
	f := newNowPaymentsFixture(t, 0, 5, model.PaymentProviderNowPayments)
	post := func(paymentId int, status string, paid float64) {
		body := strings.Replace(nowPaymentsEventBody(f.order.TradeNo, status, `5`, 5, paid), `"payment_id":4242`, fmt.Sprintf(`"payment_id":%d`, paymentId), 1)
		require.Equal(t, http.StatusOK, postNowPaymentsWebhook(body, signNowPaymentsBody(body, nowPaymentsTestSecret)), status)
	}
	orderStatus := func() string {
		require.NoError(t, f.db.First(&f.order, f.order.Id).Error)
		return f.order.Status
	}
	paymentId := func() string {
		var payment model.NowPaymentsPayment
		require.NoError(t, f.db.Where("trade_no = ?", f.order.TradeNo).First(&payment).Error)
		return payment.PaymentId
	}

	post(1001, "waiting", 0)
	post(1001, "expired", 0)
	assert.Equal(t, common.TopUpStatusFailed, orderStatus())

	post(1002, "waiting", 0)
	assert.Equal(t, common.TopUpStatusPending, orderStatus(), "a new attempt reopens the order for reconciliation")
	assert.Equal(t, "1002", paymentId())
	post(1001, "expired", 0)
	assert.Equal(t, "1002", paymentId(), "a late IPN for the dead attempt keeps the live id")
	assert.Equal(t, common.TopUpStatusFailed, orderStatus())

	post(1002, "finished", 5)
	assert.Equal(t, common.TopUpStatusSuccess, orderStatus(), "the paid retry settles even after the order was failed")
	require.NoError(t, f.db.First(&f.user, f.user.Id).Error)
	assert.Equal(t, int(5*common.QuotaPerUnit), f.user.Quota)
}

// Abandoned invoices never get a payment id and stay pending; they must not
// fill the reconcile batch ahead of a paid order whose IPN was missed.
func TestNowPaymentsReconcileBatchSkipsAbandonedCheckouts(t *testing.T) {
	f := newNowPaymentsFixture(t, 0, 5, model.PaymentProviderNowPayments)
	old := time.Now().Add(-2 * time.Hour).Unix()
	abandoned := model.TopUp{UserId: f.user.Id, TradeNo: "qa-nowpayments-abandoned", Amount: 5, Money: 5, PaymentMethod: model.PaymentMethodNowPayments, PaymentProvider: model.PaymentProviderNowPayments, CreateTime: old, Status: common.TopUpStatusPending}
	require.NoError(t, model.CreateNowPaymentsTopUp(&abandoned))
	require.NoError(t, f.db.Model(&model.TopUp{}).Where("id = ?", f.order.Id).Update("create_time", old+60).Error)
	require.NoError(t, model.SetNowPaymentsPaymentId(f.order.TradeNo, "4242"))

	pending, err := model.GetPendingNowPaymentsOrders(old-1, time.Now().Unix(), 1)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	assert.Equal(t, f.order.TradeNo, pending[0].TradeNo)
	assert.Equal(t, "4242", pending[0].PaymentId)
}

// A missed IPN must still settle: the reconcile task reads the payment back by
// the payment id an earlier IPN recorded.
func TestNowPaymentsReconcileSettlesMissedWebhook(t *testing.T) {
	f := newNowPaymentsFixture(t, 0, 5, model.PaymentProviderNowPayments)
	require.NoError(t, f.db.Model(&model.TopUp{}).Where("id = ?", f.order.Id).Update("create_time", time.Now().Add(-time.Hour).Unix()).Error)
	require.NoError(t, model.SetNowPaymentsPaymentId(f.order.TradeNo, "4242"))

	var requestedPath, requestedKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath, requestedKey = r.URL.Path, r.Header.Get("x-api-key")
		_, _ = io.WriteString(w, nowPaymentsEventBody(f.order.TradeNo, "finished", `5`, 5, 5))
	}))
	t.Cleanup(server.Close)
	nowPaymentsAPIBaseOverride = server.URL
	t.Cleanup(func() { nowPaymentsAPIBaseOverride = "" })

	runNowPaymentsReconcileOnce()

	assert.Equal(t, "/payment/4242", requestedPath)
	assert.Equal(t, setting.NowPaymentsApiKey, requestedKey)
	require.NoError(t, f.db.First(&f.order, f.order.Id).Error)
	assert.Equal(t, common.TopUpStatusSuccess, f.order.Status)
	require.NoError(t, f.db.First(&f.user, f.user.Id).Error)
	assert.Equal(t, int(5*common.QuotaPerUnit), f.user.Quota)
}

type nowPaymentsTestTransport struct {
	status int
	body   string
	err    error
}

func (transport nowPaymentsTestTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if transport.err != nil {
		return nil, transport.err
	}
	return &http.Response{StatusCode: transport.status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(transport.body)), Request: request}, nil
}

func TestNowPaymentsInvoiceErrorsDoNotExposeCredentials(t *testing.T) {
	oldKey, oldClient := setting.NowPaymentsApiKey, nowPaymentsHTTPClient
	setting.NowPaymentsApiKey = "qa-credential-that-must-not-be-logged"
	t.Cleanup(func() { setting.NowPaymentsApiKey = oldKey; nowPaymentsHTTPClient = oldClient })
	for _, transport := range []nowPaymentsTestTransport{
		{err: errors.New("unavailable qa-credential-that-must-not-be-logged")},
		{status: 401, body: `{"message":"qa-credential-that-must-not-be-logged"}`},
		{status: 200, body: `{"invoice_url":""}`},
	} {
		nowPaymentsHTTPClient = &http.Client{Transport: transport}
		_, err := createNowPaymentsInvoice(context.Background(), "qa-order", 5, "", "", "qa-invoice")
		require.Error(t, err)
		assert.NotContains(t, err.Error(), setting.NowPaymentsApiKey)
	}
}

func TestNowPaymentsCheckoutRejectsUnsafeRequestsBeforeInvoice(t *testing.T) {
	db := setupManageUserTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.TopUp{}, &model.NowPaymentsPayment{}))
	enableNowPaymentsForTest(t)
	oldClient := nowPaymentsHTTPClient
	nowPaymentsHTTPClient = &http.Client{Transport: nowPaymentsTestTransport{err: errors.New("no invoice may be requested")}}
	t.Cleanup(func() { nowPaymentsHTTPClient = oldClient })

	limited := model.User{Username: "qa-nowpayments-limit", AffCode: "qa-nowpayments-limit", Quota: common.MaxWalletQuota}
	require.NoError(t, db.Create(&limited).Error)
	ordinary := model.User{Username: "qa-nowpayments-ordinary", AffCode: "qa-nowpayments-ordinary"}
	require.NoError(t, db.Create(&ordinary).Error)

	handlers := []gin.HandlerFunc{RequestNowPaymentsAmount, RequestNowPaymentsPay}
	for _, handler := range handlers {
		for _, amount := range []int64{getNowPaymentsMinTopup(), math.MaxInt64} {
			assert.Contains(t, callNowPaymentsCheckout(handler, limited.Id, fmt.Sprintf(`{"amount":%d,"payment_method":"nowpayments"}`, amount)), `"message":"error"`)
		}
	}
	// A disabled gateway or unconfirmed compliance refuses new checkouts.
	payment := operation_setting.GetPaymentSetting()
	for _, disable := range []func(){
		func() { setting.NowPaymentsEnabled = false },
		func() { payment.ComplianceConfirmed = false },
	} {
		enableNowPaymentsForTest(t)
		disable()
		for _, handler := range handlers {
			assert.Contains(t, callNowPaymentsCheckout(handler, ordinary.Id, `{"amount":5,"payment_method":"nowpayments"}`), "NowPayments top-up is not enabled")
		}
	}
	var orders int64
	require.NoError(t, db.Model(&model.TopUp{}).Count(&orders).Error)
	assert.Zero(t, orders, "rejected checkouts must not create orders")
}

// An untrusted return page is never sent to NowPayments, but it must not block
// the checkout either: the default return path is used instead.
func TestTrustedNowPaymentsReturnURL(t *testing.T) {
	oldDomains, oldAddress := constant.TrustedRedirectDomains, system_setting.ServerAddress
	constant.TrustedRedirectDomains = []string{"interapi.test"}
	system_setting.ServerAddress = "https://api.interapi.test"
	t.Cleanup(func() { constant.TrustedRedirectDomains, system_setting.ServerAddress = oldDomains, oldAddress })
	for _, tc := range []struct{ requested, want string }{
		{requested: "https://interapi.test/dashboard/billing", want: "https://interapi.test/dashboard/billing"},
		{requested: "https://evil.example/return", want: "https://api.interapi.test/console/log"},
		{requested: "javascript:alert(1)", want: "https://api.interapi.test/console/log"},
		{requested: "", want: "https://api.interapi.test/console/log"},
	} {
		assert.Equal(t, tc.want, trustedNowPaymentsReturnURL(context.Background(), tc.requested, "/console/log"), tc.requested)
	}
}

func callNowPaymentsCheckout(handler gin.HandlerFunc, userId int, body string) string {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/user/nowpayments/pay", strings.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("id", userId)
	handler(ctx)
	return recorder.Body.String()
}

// Admin-console saves arrive as option strings; every NowPayments setting must
// reach the runtime value the gateway reads.
func TestNowPaymentsOptionsApply(t *testing.T) {
	db := setupManageUserTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Option{}, &model.SystemTask{}))
	enableNowPaymentsForTest(t)
	common.OptionMapRWMutex.Lock()
	originalOptionMap := common.OptionMap
	common.OptionMap = map[string]string{}
	common.OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		common.OptionMapRWMutex.Lock()
		common.OptionMap = originalOptionMap
		common.OptionMapRWMutex.Unlock()
	})
	oldValues := []any{setting.NowPaymentsSandbox, setting.NowPaymentsMinTopUp, setting.NowPaymentsFeePaidByUser, setting.NowPaymentsIsFixedRate}
	t.Cleanup(func() {
		setting.NowPaymentsSandbox = oldValues[0].(bool)
		setting.NowPaymentsMinTopUp = oldValues[1].(int)
		setting.NowPaymentsFeePaidByUser = oldValues[2].(bool)
		setting.NowPaymentsIsFixedRate = oldValues[3].(bool)
	})
	for key, value := range map[string]string{
		"NowPaymentsEnabled":       "false",
		"NowPaymentsApiKey":        "qa-updated-api-key",
		"NowPaymentsIpnSecret":     "qa-updated-ipn-secret",
		"NowPaymentsSandbox":       "true",
		"NowPaymentsMinTopUp":      "7",
		"NowPaymentsFeePaidByUser": "true",
		"NowPaymentsIsFixedRate":   "true",
	} {
		require.NoError(t, model.UpdateOption(key, value), key)
	}
	assert.False(t, setting.NowPaymentsEnabled)
	assert.Equal(t, "qa-updated-api-key", setting.NowPaymentsApiKey)
	assert.Equal(t, "qa-updated-ipn-secret", setting.NowPaymentsIpnSecret)
	assert.True(t, setting.NowPaymentsSandbox)
	assert.Equal(t, 7, setting.NowPaymentsMinTopUp)
	assert.True(t, setting.NowPaymentsFeePaidByUser)
	assert.True(t, setting.NowPaymentsIsFixedRate)
	assert.Equal(t, nowPaymentsAPIBaseSandbox, nowPaymentsAPIBase())
	assert.False(t, isNowPaymentsTopUpEnabled())
}
