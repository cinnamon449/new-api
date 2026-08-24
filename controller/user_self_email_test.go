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

func TestRemoveSelfEmailClearsOnlyCurrentUserAndIsIdempotent(t *testing.T) {
	db := setupManageUserTestDB(t)
	currentUser := model.User{
		Username: "self-email-user", Password: "password", Email: "self@example.com",
		Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default", AffCode: "self-email-aff",
	}
	otherUser := model.User{
		Username: "other-email-user", Password: "password", Email: "other@example.com",
		Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default", AffCode: "other-email-aff",
	}
	require.NoError(t, db.Create(&currentUser).Error)
	require.NoError(t, db.Create(&otherUser).Error)

	for range 2 {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)
		context.Request = httptest.NewRequest(http.MethodDelete, "/api/user/self/email", nil)
		context.Set("id", currentUser.Id)

		RemoveSelfEmail(context)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, recorder.Body.String(), `"success":true`)
	}

	var updatedCurrent model.User
	require.NoError(t, db.First(&updatedCurrent, currentUser.Id).Error)
	assert.Empty(t, updatedCurrent.Email)

	var unchangedOther model.User
	require.NoError(t, db.First(&unchangedOther, otherUser.Id).Error)
	assert.Equal(t, "other@example.com", unchangedOther.Email)
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
