package model

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var ErrTopUpAmountMismatch = errors.New("top-up amount mismatch")

// NowPaymentsPayment tracks a NowPayments top-up's provider state beside its
// TopUp row. NowPayments payments can be queried only by payment_id, which
// arrives with the first IPN; recording it (and what actually landed) lets the
// reconcile task settle an order whose final IPN never arrived. Keeping the
// fields in a separate table leaves the upstream TopUp model unchanged.
type NowPaymentsPayment struct {
	Id         int     `json:"id"`
	TradeNo    string  `json:"trade_no" gorm:"unique;type:varchar(255)"`
	PaymentId  string  `json:"payment_id" gorm:"type:varchar(64);default:'';index"`
	PaidAmount float64 `json:"paid_amount" gorm:"default:0"`
}

func (NowPaymentsPayment) TableName() string {
	return "nowpayments_payments"
}

// CreateNowPaymentsTopUp stores a pending NowPayments top-up and its tracking
// row together, so every order the reconcile task can see has both.
func CreateNowPaymentsTopUp(topUp *TopUp) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(topUp).Error; err != nil {
			return err
		}
		return tx.Create(&NowPaymentsPayment{TradeNo: topUp.TradeNo}).Error
	})
}

// SetNowPaymentsPaymentId records the provider payment id the reconcile task
// polls. One invoice can carry several payment attempts, so the latest live
// attempt replaces an earlier one; callers skip expired or failed attempts so
// a late IPN for a dead attempt cannot displace a live one.
func SetNowPaymentsPaymentId(tradeNo string, paymentId string) error {
	if tradeNo == "" || paymentId == "" {
		return nil
	}
	return DB.Model(&NowPaymentsPayment{}).
		Where("trade_no = ? AND payment_id <> ?", tradeNo, paymentId).
		Update("payment_id", paymentId).Error
}

// ReopenNowPaymentsTopUp returns a NowPayments order that an expired or failed
// payment attempt marked failed to pending, because the customer started a new
// payment on the same invoice. Other orders and statuses are left unchanged.
func ReopenNowPaymentsTopUp(tradeNo string) error {
	return DB.Model(&TopUp{}).
		Where("trade_no = ? AND payment_provider = ? AND status = ?", tradeNo, PaymentProviderNowPayments, common.TopUpStatusFailed).
		Update("status", common.TopUpStatusPending).Error
}

// SetNowPaymentsPaidAmount records funds the provider reports as received. It
// is monotonic: an out-of-order IPN reporting less cannot erase a higher figure.
func SetNowPaymentsPaidAmount(tradeNo string, paidAmount float64) error {
	if tradeNo == "" || paidAmount <= 0 {
		return nil
	}
	return DB.Model(&NowPaymentsPayment{}).
		Where("trade_no = ? AND paid_amount < ?", tradeNo, paidAmount).
		Update("paid_amount", paidAmount).Error
}

// GetPendingNowPaymentsOrders lists pending NowPayments top-ups created within
// [newerThan, olderThan] whose customer reached a payment (a payment id was
// recorded), oldest first.
func GetPendingNowPaymentsOrders(newerThan int64, olderThan int64, limit int) ([]NowPaymentsPayment, error) {
	// Filter on a recorded payment before the limit: abandoned invoices never
	// get a payment id and stay pending, so they must not fill the batch.
	withPayment := DB.Model(&NowPaymentsPayment{}).Select("trade_no").Where("payment_id <> ?", "")
	var tradeNos []string
	err := DB.Model(&TopUp{}).
		Where("payment_provider = ? AND status = ? AND create_time BETWEEN ? AND ?", PaymentProviderNowPayments, common.TopUpStatusPending, newerThan, olderThan).
		Where("trade_no IN (?)", withPayment).
		Order("create_time asc").Limit(limit).Pluck("trade_no", &tradeNos).Error
	if err != nil || len(tradeNos) == 0 {
		return nil, err
	}
	var payments []NowPaymentsPayment
	err = DB.Where("trade_no IN ?", tradeNos).Find(&payments).Error
	return payments, err
}

// RechargeNowPayments commits order completion and wallet credit together. The
// database row lock, rather than a process-local mutex, prevents duplicate
// credit. priceAmount is the fiat invoice price NowPayments reports and must
// match the order exactly.
func RechargeNowPayments(tradeNo string, priceAmount decimal.Decimal, callerIP string) (alreadyDone bool, err error) {
	var topUp TopUp
	var quota int
	err = DB.Transaction(func(tx *gorm.DB) error {
		if err := lockForUpdate(tx).Where("trade_no = ?", tradeNo).First(&topUp).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTopUpNotFound
			}
			return err
		}
		if topUp.PaymentProvider != PaymentProviderNowPayments || topUp.PaymentMethod != PaymentMethodNowPayments {
			return ErrPaymentMethodMismatch
		}
		// Match the exact fiat amount sent to NowPayments, including float
		// formatting at the cent boundary.
		expectedAmount, err := decimal.NewFromString(strconv.FormatFloat(topUp.Money, 'f', 2, 64))
		if err != nil || !priceAmount.IsPositive() || !priceAmount.Equal(expectedAmount) {
			return ErrTopUpAmountMismatch
		}
		if topUp.Status == common.TopUpStatusSuccess {
			alreadyDone = true
			return nil
		}
		// A failed order was failed only by an expired or failed payment attempt;
		// a later attempt on the same invoice that is paid still settles it.
		if topUp.Status != common.TopUpStatusPending && topUp.Status != common.TopUpStatusFailed {
			return ErrTopUpStatusInvalid
		}
		quota, err = common.WalletQuotaFromDecimalStrict(decimal.NewFromInt(topUp.Amount).Mul(decimal.NewFromFloat(common.QuotaPerUnit)))
		if err != nil || quota <= 0 {
			return ErrInvalidTopUpQuota
		}
		topUp.Status = common.TopUpStatusSuccess
		topUp.CompleteTime = common.GetTimestamp()
		if err := tx.Save(&topUp).Error; err != nil {
			return err
		}
		return creditTopUpQuota(tx, topUp.UserId, quota, nil)
	})
	if err != nil || alreadyDone {
		return alreadyDone, err
	}
	syncCreditUserQuotaCache(topUp.UserId, quota, "nowpayments topup")
	RecordTopupLog(topUp.UserId, common.NewMessage("{{provider}} top-up succeeded, amount added: {{quota}}, payment amount: {{amount}}", map[string]any{"provider": "NowPayments", "quota": logger.FormatQuota(quota), "amount": fmt.Sprintf("%f", topUp.Money)}), callerIP, PaymentMethodNowPayments, PaymentProviderNowPayments)
	return false, nil
}
