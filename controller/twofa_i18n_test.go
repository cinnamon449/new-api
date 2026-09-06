/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTwoFACodeInvalidReturnsStableCodeAndLocalizedMessage(t *testing.T) {
	require.NoError(t, i18n.Init())
	gin.SetMode(gin.TestMode)

	tests := []struct {
		language string
		message  string
	}{
		{language: "en", message: "Verification code or backup code is incorrect"},
		{language: "zh-CN", message: "验证码或备用码不正确"},
	}

	for _, test := range tests {
		t.Run(test.language, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Request = httptest.NewRequest("POST", "/api/user/login/2fa", nil)
			context.Request.Header.Set("Accept-Language", test.language)
			middleware.I18n()(context)

			writeSecurityOperationError(context, model.ErrTwoFACodeInvalid)

			var response struct {
				Success bool   `json:"success"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}
			require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
			assert.False(t, response.Success)
			assert.Equal(t, "TWOFA_CODE_INVALID", response.Code)
			assert.Equal(t, test.message, response.Message)
		})
	}
}

func TestTwoFALoginInvalidBackupCodePreservesLocalizedErrorAndFailureAccounting(t *testing.T) {
	for _, path := range []string{"/api/user/login/2fa", "/api/user/login/verify"} {
		for _, test := range []struct {
			language string
			message  string
		}{
			{language: "en", message: "Verification code or backup code is incorrect"},
			{language: "zh-CN", message: "验证码或备用码不正确"},
		} {
			t.Run(path+"/"+test.language, func(t *testing.T) {
				user, _ := setupSecurityEnrollmentTest(t)
				factor := &model.TwoFA{UserId: user.Id, Secret: "JBSWY3DPEHPK3PXP", IsEnabled: true}
				require.NoError(t, model.DB.Create(factor).Error)
				challenge, err := service.StartLoginVerification(user, "password")
				require.NoError(t, err)
				require.NotNil(t, challenge)
				body, err := common.Marshal(map[string]string{"flow_token": challenge.FlowToken, "code": "ABCD-1234"})
				require.NoError(t, err)
				router := gin.New()
				router.Use(middleware.I18n())
				router.POST("/api/user/login/2fa", Verify2FALogin)
				router.POST("/api/user/login/verify", VerifyLogin)
				request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(string(body)))
				request.Header.Set("Content-Type", "application/json")
				request.Header.Set("Accept-Language", test.language)
				response := httptest.NewRecorder()

				router.ServeHTTP(response, request)

				var result securityEnrollmentResponse
				require.NoError(t, common.Unmarshal(response.Body.Bytes(), &result))
				assert.Equal(t, http.StatusOK, response.Code)
				assert.False(t, result.Success)
				assert.Equal(t, "TWOFA_CODE_INVALID", result.Code)
				assert.Equal(t, test.message, result.Message)
				assert.Empty(t, response.Header().Values("Set-Cookie"))
				stored, err := model.GetTwoFAByUserId(user.Id)
				require.NoError(t, err)
				assert.Equal(t, 1, stored.FailedAttempts)
				_, err = service.RequireLoginVerification(challenge.FlowToken, service.VerificationMethodTwoFA)
				assert.NoError(t, err, "an invalid code must leave the login challenge available for retry")
			})
		}
	}
}
