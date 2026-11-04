package controller

import (
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func configureTokenGroupValidationTest(t *testing.T) {
	t.Helper()

	originalUsableGroups := setting.UserUsableGroups2JSONString()
	originalRatios := ratio_setting.GroupRatio2JSONString()
	require.NoError(t, setting.UpdateUserUsableGroupsByJSONString(`{"default":"Default","vip":"VIP"}`))
	require.NoError(t, ratio_setting.UpdateGroupRatioByJSONString(`{"default":1,"vip":1}`))
	t.Cleanup(func() {
		require.NoError(t, setting.UpdateUserUsableGroupsByJSONString(originalUsableGroups))
		require.NoError(t, ratio_setting.UpdateGroupRatioByJSONString(originalRatios))
	})
}

func TestAddTokenRejectsUnselectableGroup(t *testing.T) {
	configureTokenGroupValidationTest(t)
	user := setupTokenAutoGroupsControllerTest(t)
	request := baseAutoTokenRequest("reject-unselectable-group")
	request["group"] = "svip"

	ctx, recorder := newTokenAutoGroupsAuthenticatedContext(t, http.MethodPost, "/api/token/", request, user.Id)
	AddToken(ctx)

	response := decodeAPIResponse(t, recorder)
	require.False(t, response.Success)
	assert.Contains(t, response.Message, "svip")

	var count int64
	require.NoError(t, model.DB.Model(&model.Token{}).Where("name = ?", request["name"]).Count(&count).Error)
	assert.Zero(t, count)
}

func TestAddTokenAcceptsSelectableGroup(t *testing.T) {
	configureTokenGroupValidationTest(t)
	user := setupTokenAutoGroupsControllerTest(t)
	request := baseAutoTokenRequest("accept-selectable-group")
	request["group"] = "vip"

	ctx, recorder := newTokenAutoGroupsAuthenticatedContext(t, http.MethodPost, "/api/token/", request, user.Id)
	AddToken(ctx)

	require.True(t, decodeAPIResponse(t, recorder).Success)
	var token model.Token
	require.NoError(t, model.DB.Where("name = ?", request["name"]).First(&token).Error)
	assert.Equal(t, "vip", token.Group)
}

func TestAddTokenAllowsEmptyGroup(t *testing.T) {
	configureTokenGroupValidationTest(t)
	user := setupTokenAutoGroupsControllerTest(t)
	request := baseAutoTokenRequest("empty-group-token")
	request["group"] = ""

	ctx, recorder := newTokenAutoGroupsAuthenticatedContext(t, http.MethodPost, "/api/token/", request, user.Id)
	AddToken(ctx)

	require.True(t, decodeAPIResponse(t, recorder).Success)
	var token model.Token
	require.NoError(t, model.DB.Where("name = ?", request["name"]).First(&token).Error)
	assert.Empty(t, token.Group)
}

func TestUpdateTokenRejectsUnselectableGroup(t *testing.T) {
	configureTokenGroupValidationTest(t)
	user := setupTokenAutoGroupsControllerTest(t)
	token := seedToken(t, model.DB, user.Id, "update-reject-group", "update-reject-group-key")

	request := baseAutoTokenRequest("update-reject-group")
	request["id"] = token.Id
	request["status"] = common.TokenStatusEnabled
	request["group"] = "svip"

	ctx, recorder := newTokenAutoGroupsAuthenticatedContext(t, http.MethodPut, "/api/token/", request, user.Id)
	UpdateToken(ctx)

	require.False(t, decodeAPIResponse(t, recorder).Success)
	var reloaded model.Token
	require.NoError(t, model.DB.First(&reloaded, token.Id).Error)
	assert.Equal(t, "default", reloaded.Group)
}

func TestUpdateTokenSkipsValidationForUnchangedLegacyGroup(t *testing.T) {
	configureTokenGroupValidationTest(t)
	user := setupTokenAutoGroupsControllerTest(t)
	token := seedToken(t, model.DB, user.Id, "legacy-group-token", "legacy-group-token-key")
	token.Group = "retired"
	require.NoError(t, model.DB.Save(token).Error)

	request := baseAutoTokenRequest("legacy-group-renamed")
	request["id"] = token.Id
	request["status"] = common.TokenStatusEnabled
	request["group"] = "retired"

	ctx, recorder := newTokenAutoGroupsAuthenticatedContext(t, http.MethodPut, "/api/token/", request, user.Id)
	UpdateToken(ctx)

	require.True(t, decodeAPIResponse(t, recorder).Success)
	var reloaded model.Token
	require.NoError(t, model.DB.First(&reloaded, token.Id).Error)
	assert.Equal(t, "legacy-group-renamed", reloaded.Name)
	assert.Equal(t, "retired", reloaded.Group)
}
