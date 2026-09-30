package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveSelfEmailRequiresSingleUseBoundProof(t *testing.T) {
	for _, scenario := range []string{"missing", "wrong scope", "expired", "consumed", "changed email", "success"} {
		t.Run(scenario, func(t *testing.T) {
			user, identity := setupSecurityEnrollmentTest(t)
			require.NoError(t, model.DB.Model(user).Update("email", "self@example.com").Error)
			operation := service.VerificationOperation{Scope: "account.email.remove", Context: []byte(`{"email":"self@example.com"}`)}
			proof := ""
			if scenario != "missing" {
				if scenario == "wrong scope" {
					operation = service.VerificationOperation{Scope: service.VerificationScopeAccountDelete}
				}
				proof = issueSecurityEnrollmentProof(t, identity, operation, service.VerificationMethodPassword)
			}
			switch scenario {
			case "expired":
				require.NoError(t, model.DB.Model(&model.AuthFlow{}).Where("purpose = ?", model.AuthFlowPurposeSecurityProof).Update("expires_at", time.Now().Add(-time.Minute)).Error)
			case "consumed":
				_, err := service.ConsumeOperationProof(proof, identity, operation)
				require.NoError(t, err)
			case "changed email":
				require.NoError(t, model.DB.Model(user).Update("email", "changed@example.com").Error)
			}
			response := securityEnrollmentRequest(http.MethodDelete, "/api/user/self/email", "", proof, identity, RemoveSelfEmail)
			stored, err := model.GetUserById(user.Id, false)
			require.NoError(t, err)
			if scenario != "success" {
				assert.Equal(t, http.StatusForbidden, response.Code)
				assert.NotEmpty(t, stored.Email)
				return
			}
			assert.Equal(t, http.StatusOK, response.Code)
			assert.Empty(t, stored.Email)
			// A new email must never be removed by replaying the old proof.
			require.NoError(t, model.DB.Model(user).Update("email", "self@example.com").Error)
			replay := securityEnrollmentRequest(http.MethodDelete, "/api/user/self/email", "", proof, identity, RemoveSelfEmail)
			assert.Equal(t, http.StatusForbidden, replay.Code)
			stored, err = model.GetUserById(user.Id, false)
			require.NoError(t, err)
			assert.Equal(t, "self@example.com", stored.Email)
		})
	}
}

func TestUpdateSelfPreservesFieldsThatCustomersCannotEdit(t *testing.T) {
	db := setupManageUserTestDB(t)
	user := model.User{
		Username: "profile-fields-user", Password: "password", DisplayName: "Original name",
		Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "vip", Remark: "operator note",
		AffCode: "profile-fields-aff",
	}
	require.NoError(t, db.Create(&user).Error)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPut, "/api/user/self", strings.NewReader(`{"display_name":"Updated name"}`))
	context.Request.Header.Set("Content-Type", "application/json")
	context.Set("id", user.Id)

	UpdateSelf(context)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"success":true`)
	var updated model.User
	require.NoError(t, db.First(&updated, user.Id).Error)
	assert.Equal(t, user.Username, updated.Username)
	assert.Equal(t, "Updated name", updated.DisplayName)
	assert.Equal(t, user.Group, updated.Group)
	assert.Equal(t, user.Remark, updated.Remark)
}
