package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
