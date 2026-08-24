package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/stretchr/testify/assert"
)

func TestPaymentReturnPathUsesDefaultDashboardRoutes(t *testing.T) {
	previousAddress := system_setting.ServerAddress
	system_setting.ServerAddress = "https://dashboard.example.com/"
	t.Cleanup(func() { system_setting.ServerAddress = previousAddress })

	assert.Equal(
		t,
		"https://dashboard.example.com/wallet?pay=success",
		paymentReturnPath("/wallet?pay=success"),
	)
	assert.Equal(
		t,
		"https://dashboard.example.com/usage-logs",
		paymentReturnPath("/usage-logs"),
	)
}

func TestPasswordResetLinkUsesCustomerFrontendAndEncodesParameters(t *testing.T) {
	t.Setenv("PASSWORD_RESET_BASE_URL", "https://interapi.example.com/")

	assert.Equal(
		t,
		"https://interapi.example.com/user/reset?email=operator%2Balerts%40example.com&token=token%2Fwith%3Fsymbols",
		passwordResetLink("operator+alerts@example.com", "token/with?symbols"),
	)
}

func TestPasswordResetLinkFallsBackToServerAddress(t *testing.T) {
	previousAddress := system_setting.ServerAddress
	system_setting.ServerAddress = "https://admin.example.com/"
	t.Cleanup(func() { system_setting.ServerAddress = previousAddress })
	t.Setenv("PASSWORD_RESET_BASE_URL", "")

	assert.Equal(
		t,
		"https://admin.example.com/user/reset?email=operator%40example.com&token=reset-token",
		passwordResetLink("operator@example.com", "reset-token"),
	)
}
