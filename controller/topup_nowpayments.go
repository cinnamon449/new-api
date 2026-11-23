package controller

// NowPayments crypto top-ups. The provider protocol (invoice creation, IPN
// signature, status handling, underpayment tolerance, payment lookup) is
// adapted from UnoRouter's production new-api fork
// (https://github.com/unorouter/new-api), on InterAPI's gin handlers and
// shared top-up pricing.

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

const (
	nowPaymentsSignatureHeader = "x-nowpayments-sig"
	nowPaymentsAPIBaseProd     = "https://api.nowpayments.io/v1"
	nowPaymentsAPIBaseSandbox  = "https://api-sandbox.nowpayments.io/v1"
)

// nowPaymentsHTTPClient is dedicated to NowPayments API calls with a timeout so
// a slow upstream cannot hold a user's payment request open indefinitely.
var nowPaymentsHTTPClient = &http.Client{Timeout: 30 * time.Second}

// nowPaymentsAPIBaseOverride lets tests point the client at a local server.
var nowPaymentsAPIBaseOverride = ""

func nowPaymentsAPIBase() string {
	if nowPaymentsAPIBaseOverride != "" {
		return nowPaymentsAPIBaseOverride
	}
	if setting.NowPaymentsSandbox {
		return nowPaymentsAPIBaseSandbox
	}
	return nowPaymentsAPIBaseProd
}

type NowPaymentsPayRequest struct {
	// Amount is the quantity of units to purchase.
	Amount int64 `json:"amount"`
	// PaymentMethod must be "nowpayments".
	PaymentMethod string `json:"payment_method"`
	// SuccessURL and CancelURL are optional return pages; each must be on a
	// TRUSTED_REDIRECT_DOMAINS host.
	SuccessURL string `json:"success_url,omitempty"`
	CancelURL  string `json:"cancel_url,omitempty"`
}

type nowPaymentsInvoiceRequest struct {
	PriceAmount      float64 `json:"price_amount"`
	PriceCurrency    string  `json:"price_currency"`
	OrderId          string  `json:"order_id"`
	OrderDescription string  `json:"order_description"`
	IpnCallbackURL   string  `json:"ipn_callback_url"`
	SuccessURL       string  `json:"success_url"`
	CancelURL        string  `json:"cancel_url"`
	IsFixedRate      bool    `json:"is_fixed_rate"`
	IsFeePaidByUser  bool    `json:"is_fee_paid_by_user"`
}

type nowPaymentsInvoiceResponse struct {
	Id         string `json:"id"`
	InvoiceURL string `json:"invoice_url"`
	OrderId    string `json:"order_id"`
}

// nowPaymentsPaymentEvent is the payment shape NowPayments sends in an IPN and
// returns from GET /payment/{id}.
type nowPaymentsPaymentEvent struct {
	PaymentId     int64             `json:"payment_id"`
	PaymentStatus string            `json:"payment_status"`
	PriceAmount   common.RawMessage `json:"price_amount"`
	PriceCurrency string            `json:"price_currency"`
	PayAmount     float64           `json:"pay_amount"`
	PayCurrency   string            `json:"pay_currency"`
	OrderId       string            `json:"order_id"`
	ActuallyPaid  float64           `json:"actually_paid"`
}

// RequestNowPaymentsAmount returns the fiat price preview for a top-up amount,
// using the shared top-up price model (same as Epay).
func RequestNowPaymentsAmount(c *gin.Context) {
	var req NowPaymentsPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	if req.Amount < getNowPaymentsMinTopup() {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", getNowPaymentsMinTopup())})
		return
	}
	id := c.GetInt("id")
	if rejectInvalidTopUpQuota(c, id, req.Amount) {
		return
	}
	group, err := model.GetUserGroup(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "获取用户分组失败"})
		return
	}
	payMoney := getPayMoney(req.Amount, group)
	if payMoney < 0.01 {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "充值金额过低"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": strconv.FormatFloat(payMoney, 'f', 2, 64)})
}

// RequestNowPaymentsPay creates a NowPayments invoice and returns the hosted
// invoice page where the buyer chooses a cryptocurrency and pays.
func RequestNowPaymentsPay(c *gin.Context) {
	var req NowPaymentsPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	if req.PaymentMethod != model.PaymentMethodNowPayments {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "不支持的支付渠道"})
		return
	}
	if req.Amount < getNowPaymentsMinTopup() {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", getNowPaymentsMinTopup())})
		return
	}
	if req.SuccessURL != "" && common.ValidateRedirectURL(req.SuccessURL) != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "Success redirect URL is not in the trusted domain list"})
		return
	}
	if req.CancelURL != "" && common.ValidateRedirectURL(req.CancelURL) != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "Cancel redirect URL is not in the trusted domain list"})
		return
	}

	id := c.GetInt("id")
	if rejectInvalidTopUpQuota(c, id, req.Amount) {
		return
	}
	group, err := model.GetUserGroup(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "获取用户分组失败"})
		return
	}
	payMoney := getPayMoney(req.Amount, group)
	if payMoney < 0.01 {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "充值金额过低"})
		return
	}

	tradeNo := fmt.Sprintf("NOWUSR%dNO%s%d", id, common.GetRandomString(6), time.Now().Unix())

	// Resolve amount stored on the order in display units (mirror Epay logic).
	amount := req.Amount
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		dAmount := decimal.NewFromInt(amount)
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		amount = dAmount.Div(dQuotaPerUnit).IntPart()
	}

	// Persist the pending order before creating the upstream invoice, so a
	// failure to reach NowPayments leaves a recoverable local record.
	topUp := &model.TopUp{
		UserId:          id,
		Amount:          amount,
		Money:           payMoney,
		TradeNo:         tradeNo,
		PaymentMethod:   model.PaymentMethodNowPayments,
		PaymentProvider: model.PaymentProviderNowPayments,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}
	if err := model.CreateNowPaymentsTopUp(topUp); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("NowPayments 创建充值订单失败 user_id=%d trade_no=%s amount=%d error=%q", id, tradeNo, req.Amount, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}

	// Invoice the cent-rounded price; the webhook requires NowPayments to echo
	// exactly this amount back before crediting.
	invoicePrice, _ := strconv.ParseFloat(strconv.FormatFloat(payMoney, 'f', 2, 64), 64)
	invoiceURL, err := createNowPaymentsInvoice(c.Request.Context(), tradeNo, invoicePrice, req.SuccessURL, req.CancelURL, fmt.Sprintf("Top-up #%d", req.Amount))
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("NowPayments 创建发票失败 user_id=%d trade_no=%s amount=%d error=%q", id, tradeNo, req.Amount, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}

	logger.LogInfo(c.Request.Context(), fmt.Sprintf("NowPayments 充值订单创建成功 user_id=%d trade_no=%s amount=%d money=%.2f", id, tradeNo, req.Amount, payMoney))
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"invoice_url": invoiceURL,
			"order_id":    tradeNo,
		},
	})
}

// createNowPaymentsInvoice calls POST /invoice and returns the hosted invoice
// URL. Errors never include the API key or the provider response body.
func createNowPaymentsInvoice(ctx context.Context, orderId string, priceAmount float64, successURL string, cancelURL string, description string) (string, error) {
	apiKey := strings.TrimSpace(setting.NowPaymentsApiKey)
	if apiKey == "" {
		return "", errors.New("NowPayments API key is not configured")
	}
	if successURL == "" {
		successURL = paymentReturnPath("/console/log")
	}
	if cancelURL == "" {
		cancelURL = paymentReturnPath("/console/topup")
	}

	body, err := common.Marshal(nowPaymentsInvoiceRequest{
		PriceAmount:      priceAmount,
		PriceCurrency:    "usd",
		OrderId:          orderId,
		OrderDescription: description,
		IpnCallbackURL:   service.GetCallbackAddress() + "/api/nowpayments/webhook",
		SuccessURL:       successURL,
		CancelURL:        cancelURL,
		IsFixedRate:      setting.NowPaymentsIsFixedRate,
		IsFeePaidByUser:  setting.NowPaymentsFeePaidByUser,
	})
	if err != nil {
		return "", err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, nowPaymentsAPIBase()+"/invoice", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)

	resp, err := nowPaymentsHTTPClient.Do(httpReq)
	if err != nil {
		return "", errors.New("NowPayments request failed")
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("failed to read NowPayments response")
	}
	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("NowPayments invoice creation failed status=%d", resp.StatusCode)
	}
	var invoice nowPaymentsInvoiceResponse
	if err := common.Unmarshal(respBody, &invoice); err != nil {
		return "", errors.New("failed to parse NowPayments response")
	}
	if invoice.InvoiceURL == "" {
		return "", errors.New("NowPayments returned an empty invoice_url")
	}
	return invoice.InvoiceURL, nil
}

// canonicalNowPaymentsJSON re-marshals an IPN payload with sorted keys, which
// is what NowPayments signs. Direct encoding/json use is an intentional
// exception to the common JSON wrappers: encoding/json sorts map keys
// deterministically, and byte-exact output is required for the HMAC.
func canonicalNowPaymentsJSON(payload []byte) ([]byte, error) {
	dec := stdjson.NewDecoder(bytes.NewReader(payload))
	dec.UseNumber()
	var obj map[string]any
	if err := dec.Decode(&obj); err != nil {
		return nil, err
	}
	return stdjson.Marshal(obj)
}

// verifyNowPaymentsSignature checks x-nowpayments-sig: hex HMAC-SHA512 of the
// sorted-key JSON body with the IPN secret, compared in constant time.
func verifyNowPaymentsSignature(payload []byte, sig string, secret string) bool {
	if secret == "" || sig == "" {
		return false
	}
	sorted, err := canonicalNowPaymentsJSON(payload)
	if err != nil {
		return false
	}
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write(sorted)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(strings.ToLower(sig)), []byte(expected))
}

// NowPaymentsWebhook handles NowPayments IPN callbacks.
func NowPaymentsWebhook(c *gin.Context) {
	ctx := c.Request.Context()
	callerIP := c.ClientIP()
	if !isNowPaymentsWebhookEnabled() {
		logger.LogWarn(ctx, fmt.Sprintf("NowPayments webhook rejected reason=webhook_disabled client_ip=%s", callerIP))
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("NowPayments webhook failed to read payload client_ip=%s error=%q", callerIP, err.Error()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if !verifyNowPaymentsSignature(body, c.GetHeader(nowPaymentsSignatureHeader), setting.NowPaymentsIpnSecret) {
		logger.LogWarn(ctx, fmt.Sprintf("NowPayments webhook signature verification failed client_ip=%s", callerIP))
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	var event nowPaymentsPaymentEvent
	if err := common.Unmarshal(body, &event); err != nil {
		logger.LogError(ctx, fmt.Sprintf("NowPayments webhook failed to parse payload client_ip=%s error=%q", callerIP, err.Error()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	logger.LogInfo(ctx, fmt.Sprintf("NowPayments webhook received order_id=%s status=%s pay_currency=%s actually_paid=%v client_ip=%s", event.OrderId, event.PaymentStatus, event.PayCurrency, event.ActuallyPaid, callerIP))
	if event.OrderId == "" {
		logger.LogWarn(ctx, fmt.Sprintf("NowPayments webhook missing order_id client_ip=%s", callerIP))
		c.Status(http.StatusOK)
		return
	}

	status := settleNowPaymentsEvent(ctx, &event, callerIP)
	if status >= http.StatusBadRequest {
		c.AbortWithStatus(status)
		return
	}
	c.Status(status)
}

// settleNowPaymentsEvent applies one NowPayments payment state to its order and
// returns the HTTP status the webhook answers with. The reconcile task feeds it
// the same shape read back from GET /payment/{id}, so an order whose IPN never
// arrived settles by exactly the rules an IPN would have used.
func settleNowPaymentsEvent(ctx context.Context, event *nowPaymentsPaymentEvent, callerIP string) int {
	orderId := event.OrderId

	// Record the provider payment id and what actually landed on every event:
	// NowPayments payments can only be looked up by payment_id, and the paid
	// figure is what separates a stalled paid order from an abandoned checkout.
	if event.PaymentId != 0 {
		if err := model.SetNowPaymentsPaymentId(orderId, strconv.FormatInt(event.PaymentId, 10)); err != nil {
			logger.LogError(ctx, fmt.Sprintf("NowPayments failed to record payment id trade_no=%s payment_id=%d error=%q", orderId, event.PaymentId, err.Error()))
		}
	}
	if event.ActuallyPaid > 0 {
		if err := model.SetNowPaymentsPaidAmount(orderId, event.ActuallyPaid); err != nil {
			logger.LogError(ctx, fmt.Sprintf("NowPayments failed to record paid amount trade_no=%s actually_paid=%v error=%q", orderId, event.ActuallyPaid, err.Error()))
		}
	}

	switch event.PaymentStatus {
	case "finished":
		return creditNowPaymentsOrder(ctx, event, callerIP)
	case "partially_paid":
		// A crypto transfer rarely lands on the exact invoice amount: gas,
		// slippage and rounding leave a few cents either way. Settle a small
		// shortfall here instead of depending on a dashboard setting.
		if withinUnderpaymentTolerance(event) {
			logger.LogInfo(ctx, fmt.Sprintf("NowPayments crediting partial payment within tolerance trade_no=%s pay_amount=%v actually_paid=%v", orderId, event.PayAmount, event.ActuallyPaid))
			return creditNowPaymentsOrder(ctx, event, callerIP)
		}
		logger.LogWarn(ctx, fmt.Sprintf("NowPayments underpaid beyond tolerance, left pending trade_no=%s pay_amount=%v actually_paid=%v", orderId, event.PayAmount, event.ActuallyPaid))
		return http.StatusOK
	case "failed", "expired", "refunded":
		LockOrder(orderId)
		defer UnlockOrder(orderId)
		err := model.UpdatePendingTopUpStatus(orderId, model.PaymentProviderNowPayments, common.TopUpStatusFailed)
		if err != nil && !errors.Is(err, model.ErrTopUpNotFound) && !errors.Is(err, model.ErrTopUpStatusInvalid) {
			logger.LogError(ctx, fmt.Sprintf("NowPayments failed to mark failure status trade_no=%s status=%s error=%q", orderId, event.PaymentStatus, err.Error()))
		} else {
			logger.LogInfo(ctx, fmt.Sprintf("NowPayments order marked status=%s trade_no=%s", event.PaymentStatus, orderId))
		}
		return http.StatusOK
	case "waiting", "confirming", "confirmed", "sending":
		logger.LogInfo(ctx, fmt.Sprintf("NowPayments awaiting confirmation trade_no=%s status=%s", orderId, event.PaymentStatus))
		return http.StatusOK
	default:
		logger.LogWarn(ctx, fmt.Sprintf("NowPayments unknown status ignored trade_no=%s status=%s", orderId, event.PaymentStatus))
		return http.StatusOK
	}
}

func creditNowPaymentsOrder(ctx context.Context, event *nowPaymentsPaymentEvent, callerIP string) int {
	LockOrder(event.OrderId)
	defer UnlockOrder(event.OrderId)

	priceAmount, err := decimal.NewFromString(common.JsonRawMessageToString(event.PriceAmount))
	if err != nil || !priceAmount.IsPositive() {
		logger.LogWarn(ctx, fmt.Sprintf("NowPayments event has no valid price_amount trade_no=%s", event.OrderId))
		return http.StatusBadRequest
	}
	alreadyDone, err := model.RechargeNowPayments(event.OrderId, priceAmount, callerIP)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("NowPayments credit failed trade_no=%s error=%q", event.OrderId, err.Error()))
		if errors.Is(err, model.ErrTopUpAmountMismatch) || errors.Is(err, model.ErrPaymentMethodMismatch) || errors.Is(err, model.ErrTopUpStatusInvalid) || errors.Is(err, model.ErrTopUpNotFound) {
			return http.StatusBadRequest
		}
		return http.StatusInternalServerError
	}
	logger.LogInfo(ctx, fmt.Sprintf("NowPayments credit completed trade_no=%s status=%s pay_currency=%s actually_paid=%v already_done=%t", event.OrderId, event.PaymentStatus, event.PayCurrency, event.ActuallyPaid, alreadyDone))
	return http.StatusOK
}

// fetchNowPaymentsPayment reads a payment back from NowPayments. The response
// carries the same field names as an IPN, so it settles through the same path.
func fetchNowPaymentsPayment(ctx context.Context, paymentId string) (*nowPaymentsPaymentEvent, error) {
	apiKey := strings.TrimSpace(setting.NowPaymentsApiKey)
	if apiKey == "" {
		return nil, errors.New("NowPayments API key is not configured")
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, nowPaymentsAPIBase()+"/payment/"+paymentId, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("x-api-key", apiKey)
	resp, err := nowPaymentsHTTPClient.Do(httpReq)
	if err != nil {
		return nil, errors.New("NowPayments payment lookup failed")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NowPayments payment lookup returned HTTP %d", resp.StatusCode)
	}
	var event nowPaymentsPaymentEvent
	if err := common.Unmarshal(body, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

func getNowPaymentsMinTopup() int64 {
	minTopup := setting.NowPaymentsMinTopUp
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		dMinTopup := decimal.NewFromInt(int64(minTopup))
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		quota, err := common.WalletQuotaFromDecimalStrict(dMinTopup.Mul(dQuotaPerUnit))
		if err != nil {
			return common.MaxWalletQuota
		}
		minTopup = quota
	}
	return int64(minTopup)
}

// nowPaymentsUnderpaymentTolerance should match the "payment covering"
// percentage configured on the NowPayments merchant account, so both sides
// agree on what counts as paid.
const nowPaymentsUnderpaymentTolerance = 0.02

// withinUnderpaymentTolerance reports whether a partially paid deposit is close
// enough to the invoice to credit. Overpayments always qualify; a shortfall
// qualifies while it stays inside the tolerance. A zero pay_amount is never
// credited, since it carries no invoice to compare against.
func withinUnderpaymentTolerance(event *nowPaymentsPaymentEvent) bool {
	if event == nil || event.PayAmount <= 0 || event.ActuallyPaid <= 0 {
		return false
	}
	if event.ActuallyPaid >= event.PayAmount {
		return true
	}
	shortfall := (event.PayAmount - event.ActuallyPaid) / event.PayAmount
	return shortfall <= nowPaymentsUnderpaymentTolerance
}
