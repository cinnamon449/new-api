package controller

import (
	"net/url"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/system_setting"
)

func paymentReturnPath(suffix string) string {
	base := strings.TrimRight(system_setting.ServerAddress, "/")
	return base + suffix
}

func passwordResetLink(email string, token string) string {
	base := common.GetEnvOrDefaultString("PASSWORD_RESET_BASE_URL", system_setting.ServerAddress)
	query := url.Values{}
	query.Set("email", email)
	query.Set("token", token)
	return strings.TrimRight(base, "/") + "/user/reset?" + query.Encode()
}
