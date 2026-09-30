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

// RechargePlisio commits order completion and wallet credit together. The
// database lock, rather than a process-local mutex, prevents duplicate credit.
func RechargePlisio(tradeNo string, sourceAmount decimal.Decimal, callerIP string) (alreadyDone bool, err error) {
	var topUp TopUp
	var quota int
	err = DB.Transaction(func(tx *gorm.DB) error {
		if err := lockForUpdate(tx).Where("trade_no = ?", tradeNo).First(&topUp).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTopUpNotFound
			}
			return err
		}
		if topUp.PaymentProvider != PaymentProviderPlisio || topUp.PaymentMethod != PaymentMethodPlisio {
			return ErrPaymentMethodMismatch
		}
		// Match the exact fiat amount sent to Plisio, including float formatting
		// at the cent boundary; decimal rounding can produce a different invoice.
		expectedAmount, err := decimal.NewFromString(strconv.FormatFloat(topUp.Money, 'f', 2, 64))
		if err != nil || !sourceAmount.IsPositive() || !sourceAmount.Equal(expectedAmount) {
			return ErrTopUpAmountMismatch
		}
		if topUp.Status == common.TopUpStatusSuccess {
			alreadyDone = true
			return nil
		}
		if topUp.Status != common.TopUpStatusPending {
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
	syncCreditUserQuotaCache(topUp.UserId, quota, "plisio topup")
	RecordTopupLog(topUp.UserId, fmt.Sprintf("使用加密货币充值成功，充值金额：%s，支付金额：%f", logger.LogQuota(quota), topUp.Money), callerIP, PaymentMethodPlisio, PaymentProviderPlisio)
	return false, nil
}
