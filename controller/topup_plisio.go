package controller

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// plisioHTTPClient is dedicated to Plisio API calls with a sane timeout so a
// slow upstream cannot hold a user's payment request open indefinitely.
var plisioHTTPClient = &http.Client{Timeout: 30 * time.Second}

const plisioAPIBase = "https://api.plisio.net/api/v1"

// PlisioPayRequest represents a payment request for a Plisio crypto checkout.
type PlisioPayRequest struct {
	// Amount is the quantity of units to purchase.
	Amount int64 `json:"amount"`
	// PaymentMethod specifies the payment method (expected "plisio").
	PaymentMethod string `json:"payment_method"`
}

// RequestPlisioAmount returns the fiat price preview for a given top-up amount,
// reusing the shared global price model (same as Epay).
func RequestPlisioAmount(c *gin.Context) {
	var req PlisioPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	if req.Amount < getPlisioMinTopup() {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", getPlisioMinTopup())})
		return
	}
	id := c.GetInt("id")
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

// RequestPlisioPay creates a Plisio invoice and redirects the user to the
// hosted invoice page where they choose a cryptocurrency and pay.
func RequestPlisioPay(c *gin.Context) {
	var req PlisioPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	if req.PaymentMethod != model.PaymentMethodPlisio {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "不支持的支付渠道"})
		return
	}
	if req.Amount < getPlisioMinTopup() {
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("充值数量不能小于 %d", getPlisioMinTopup()), "data": 10})
		return
	}

	id := c.GetInt("id")
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

	user, _ := model.GetUserById(id, false)
	tradeNo := fmt.Sprintf("PLSUSR%dNO%s%d", id, common.GetRandomString(6), time.Now().Unix())

	// Resolve amount stored on the order in display units (mirror Epay logic).
	amount := req.Amount
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		dAmount := decimal.NewFromInt(int64(amount))
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		amount = dAmount.Div(dQuotaPerUnit).IntPart()
	}

	// Persist the pending order before creating the upstream invoice, so a
	// failure to reach Plisio leaves a recoverable local record (same order as
	// Stripe/Creem).
	topUp := &model.TopUp{
		UserId:          id,
		Amount:          amount,
		Money:           payMoney,
		TradeNo:         tradeNo,
		PaymentMethod:   model.PaymentMethodPlisio,
		PaymentProvider: model.PaymentProviderPlisio,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}
	if err := topUp.Insert(); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("Plisio 创建充值订单失败 user_id=%d trade_no=%s amount=%d error=%q", id, tradeNo, req.Amount, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}

	sourceCurrency := strings.TrimSpace(setting.PlisioSourceCurrency)
	if sourceCurrency == "" {
		sourceCurrency = "USD"
	}
	invoiceURL, err := createPlisioInvoice(c.Request.Context(), tradeNo, fmt.Sprintf("Top-up #%d", req.Amount), strconv.FormatFloat(payMoney, 'f', 2, 64), sourceCurrency, user.Email)
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("Plisio 创建发票失败 user_id=%d trade_no=%s amount=%d error=%q", id, tradeNo, req.Amount, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}

	logger.LogInfo(c.Request.Context(), fmt.Sprintf("Plisio 充值订单创建成功 user_id=%d trade_no=%s amount=%d money=%.2f currency=%s", id, tradeNo, req.Amount, payMoney, sourceCurrency))
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"invoice_url": invoiceURL,
			"order_id":    tradeNo,
		},
	})
}

type plisioInvoiceResponse struct {
	Status string `json:"status"`
	Data   struct {
		TxnID           string `json:"txn_id"`
		InvoiceURL      string `json:"invoice_url"`
		InvoiceTotalSum string `json:"invoice_total_sum"`
	} `json:"data"`
}

// createPlisioInvoice calls Plisio's GET /invoices/new endpoint and returns the
// hosted invoice URL the buyer must be redirected to.
func createPlisioInvoice(ctx context.Context, orderNumber string, orderName string, sourceAmount string, sourceCurrency string, email string) (string, error) {
	apiKey := strings.TrimSpace(setting.PlisioApiKey)
	if apiKey == "" {
		return "", fmt.Errorf("Plisio api key 未配置")
	}

	callbackURL := service.GetCallbackAddress() + "/api/plisio/webhook?json=true"
	successURL := paymentReturnPath("/console/log")
	failURL := paymentReturnPath("/console/topup")

	params := url.Values{}
	params.Set("api_key", apiKey)
	params.Set("source_currency", sourceCurrency)
	params.Set("source_amount", sourceAmount)
	params.Set("order_number", orderNumber)
	params.Set("order_name", orderName)
	params.Set("callback_url", callbackURL)
	params.Set("success_invoice_url", successURL)
	params.Set("fail_invoice_url", failURL)
	if email != "" {
		params.Set("email", email)
	}

	requestURL := plisioAPIBase + "/invoices/new?" + params.Encode()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := plisioHTTPClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result plisioInvoiceResponse
	if err := common.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析 Plisio 响应失败: %w", err)
	}
	if result.Status != "success" || result.Data.InvoiceURL == "" {
		return "", fmt.Errorf("Plisio 返回非成功状态: status=%s body=%s", result.Status, string(body))
	}
	return result.Data.InvoiceURL, nil
}

// PlisioWebhook handles Plisio IPN callbacks. Plisio POSTs JSON (because
// callback_url carries json=true) and signs the payload with an HMAC-SHA1
// verify_hash computed over the JSON body minus verify_hash. Only a
// status=completed event credits the user's quota.
func PlisioWebhook(c *gin.Context) {
	ctx := c.Request.Context()
	if !isPlisioWebhookEnabled() {
		logger.LogWarn(ctx, fmt.Sprintf("Plisio webhook 被拒绝 reason=webhook_disabled path=%q client_ip=%s", c.Request.RequestURI, c.ClientIP()))
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("Plisio webhook 读取请求体失败 path=%q client_ip=%s error=%q", c.Request.RequestURI, c.ClientIP(), err.Error()))
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}

	var payload plisioIPN
	if err := common.Unmarshal(rawBody, &payload); err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("Plisio webhook 解析 JSON 失败 path=%q client_ip=%s error=%q body=%q", c.Request.RequestURI, c.ClientIP(), err.Error(), string(rawBody)))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if !verifyPlisioSignature(rawBody, payload.VerifyHash, setting.PlisioApiKey) {
		logger.LogWarn(ctx, fmt.Sprintf("Plisio webhook 验签失败 path=%q client_ip=%s order_number=%s status=%s", c.Request.RequestURI, c.ClientIP(), payload.OrderNumber, payload.Status))
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	callerIP := c.ClientIP()
	logger.LogInfo(ctx, fmt.Sprintf("Plisio webhook 验签成功 order_number=%s status=%s txn_id=%s client_ip=%s", payload.OrderNumber, payload.Status, payload.TxnID, callerIP))

	if payload.Status != "completed" {
		logger.LogInfo(ctx, fmt.Sprintf("Plisio webhook 忽略非完成状态 order_number=%s status=%s client_ip=%s", payload.OrderNumber, payload.Status, callerIP))
		c.Status(http.StatusOK)
		return
	}

	if payload.OrderNumber == "" {
		logger.LogWarn(ctx, fmt.Sprintf("Plisio webhook completed 缺少订单号 client_ip=%s", callerIP))
		c.Status(http.StatusOK)
		return
	}

	LockOrder(payload.OrderNumber)
	defer UnlockOrder(payload.OrderNumber)

	topUp := model.GetTopUpByTradeNo(payload.OrderNumber)
	if topUp == nil {
		logger.LogWarn(ctx, fmt.Sprintf("Plisio webhook completed 但本地订单不存在 trade_no=%s client_ip=%s", payload.OrderNumber, callerIP))
		c.Status(http.StatusOK)
		return
	}
	if topUp.PaymentProvider != model.PaymentProviderPlisio {
		logger.LogWarn(ctx, fmt.Sprintf("Plisio webhook 订单支付网关不匹配 trade_no=%s payment_provider=%s client_ip=%s", payload.OrderNumber, topUp.PaymentProvider, callerIP))
		c.Status(http.StatusOK)
		return
	}
	if topUp.Status != common.TopUpStatusPending {
		logger.LogInfo(ctx, fmt.Sprintf("Plisio webhook 订单状态非 pending，忽略 trade_no=%s status=%s client_ip=%s", payload.OrderNumber, topUp.Status, callerIP))
		c.Status(http.StatusOK)
		return
	}

	topUp.Status = common.TopUpStatusSuccess
	topUp.CompleteTime = time.Now().Unix()
	if err := topUp.Update(); err != nil {
		logger.LogError(ctx, fmt.Sprintf("Plisio 更新充值订单状态失败 trade_no=%s client_ip=%s error=%q", payload.OrderNumber, callerIP, err.Error()))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	dAmount := decimal.NewFromInt(int64(topUp.Amount))
	dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
	quotaToAdd := int(dAmount.Mul(dQuotaPerUnit).IntPart())
	if err := model.IncreaseUserQuota(topUp.UserId, quotaToAdd, true); err != nil {
		logger.LogError(ctx, fmt.Sprintf("Plisio 增加用户额度失败 trade_no=%s user_id=%d quota=%d error=%q", payload.OrderNumber, topUp.UserId, quotaToAdd, err.Error()))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	model.RecordTopupLog(topUp.UserId, fmt.Sprintf("使用加密货币充值成功，充值金额：%s，支付金额：%f", logger.LogQuota(quotaToAdd), topUp.Money), callerIP, model.PaymentMethodPlisio, "plisio")
	logger.LogInfo(ctx, fmt.Sprintf("Plisio 充值成功 trade_no=%s user_id=%d quota=%d client_ip=%s", payload.OrderNumber, topUp.UserId, quotaToAdd, callerIP))
	c.Status(http.StatusOK)
}

// plisioIPN captures the fields Plisio sends in an IPN callback. Only the
// fields used for routing/fulfillment are typed; the rest are ignored.
type plisioIPN struct {
	TxnID       string `json:"txn_id"`
	OrderNumber string `json:"order_number"`
	OrderName   string `json:"order_name"`
	Status      string `json:"status"`
	VerifyHash  string `json:"verify_hash"`
}

// verifyPlisioSignature recomputes the Plisio verify_hash (HMAC-SHA1 over the
// JSON body with verify_hash removed) and compares it in constant time.
//
// Plisio signs JSON.stringify(body without verify_hash), preserving the
// original field order. Because Go maps are unordered, the canonical string is
// rebuilt from the raw body using an ordered token scan that drops verify_hash
// and preserves every other value byte-for-byte, faithfully reproducing
// Plisio's JSON.stringify output.
func verifyPlisioSignature(rawBody []byte, verifyHash string, secret string) bool {
	if secret == "" || verifyHash == "" {
		return false
	}
	canonical, err := plisioCanonicalBody(rawBody)
	if err != nil {
		return false
	}
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(canonical))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(verifyHash), []byte(expected))
}

// plisioCanonicalBody rebuilds the JSON object in rawBody without its
// verify_hash field, preserving the original key order and each value's raw
// bytes (matching JSON.stringify semantics for non-PHP callbacks).
func plisioCanonicalBody(rawBody []byte) (string, error) {
	dec := json.NewDecoder(bytes.NewReader(rawBody))
	dec.UseNumber()

	t, err := dec.Token()
	if err != nil {
		return "", err
	}
	delim, ok := t.(json.Delim)
	if !ok || delim != '{' {
		return "", fmt.Errorf("plisio callback root is not a JSON object")
	}

	var sb strings.Builder
	sb.WriteByte('{')
	first := true
	for dec.More() {
		keyToken, err := dec.Token()
		if err != nil {
			return "", err
		}
		key, ok := keyToken.(string)
		if !ok {
			return "", fmt.Errorf("plisio callback expected string key")
		}
		var rawValue json.RawMessage
		if err := dec.Decode(&rawValue); err != nil {
			return "", err
		}
		if key == "verify_hash" {
			continue
		}
		if !first {
			sb.WriteByte(',')
		}
		first = false
		encodedKey, err := json.Marshal(key)
		if err != nil {
			return "", err
		}
		sb.Write(encodedKey)
		sb.WriteByte(':')
		sb.Write(rawValue)
	}
	if _, err := dec.Token(); err != nil {
		return "", err
	}
	sb.WriteByte('}')
	return sb.String(), nil
}

func getPlisioMinTopup() int64 {
	minTopup := setting.PlisioMinTopUp
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		dMinTopup := decimal.NewFromInt(int64(minTopup))
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		minTopup = int(dMinTopup.Mul(dQuotaPerUnit).IntPart())
	}
	return int64(minTopup)
}
