package visualvalidate

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 上游"BytedToken 不存在"的真实报文（2026-09-24 实测，cii-group → ark）：
// HTTP 200 + ResponseMetadata.Error.Code = "NotFound.<BytedToken>"。
const upstreamTokenNotFoundBody = `{"ResponseMetadata":{"RequestId":"202609241535307EA704DAF93A7E49A5CD","Action":"GetVisualValidateResult","Version":"2024-01-01","Service":"ark","Region":"cn-beijing","Error":{"Code":"NotFound.202609241510228EDD847E5211BEF3B31C","Message":"The specified token Visual Face token is not found.","Data":{"__Message.parameter":"202609241510228EDD847E5211BEF3B31C","__Message.resourceContent":"Visual Face token","__Message.resourceType":"token"}}}}`

// 上游认证成功的真实报文：HTTP 200 + Result.GroupId。
const upstreamSuccessBody = `{"ResponseMetadata":{"RequestId":"2026092415355982347304681CE9CD150E","Action":"GetVisualValidateResult","Version":"2024-01-01","Service":"ark","Region":"cn-beijing"},"Result":{"GroupId":"group-20260924153205-lx298"}}`

// TestIsUpstreamTokenNotFound_EnvelopeNotFound 上游 200 + 信封 NotFound
// 必须识别为 token 失效（而不是 502）。
func TestIsUpstreamTokenNotFound_EnvelopeNotFound(t *testing.T) {
	assert.True(t, isUpstreamTokenNotFound([]byte(upstreamTokenNotFoundBody), http.StatusOK))
}

// TestIsUpstreamTokenNotFound_EmptyBody404 上游对"还没刷脸"的标准回应是
// HTTP 404 + 空 body，同样算 token 未认证。
func TestIsUpstreamTokenNotFound_EmptyBody404(t *testing.T) {
	assert.True(t, isUpstreamTokenNotFound(nil, http.StatusNotFound))
	assert.True(t, isUpstreamTokenNotFound([]byte("  "), http.StatusNotFound))
}

// TestIsUpstreamTokenNotFound_Gateway404NotTokenError 网关/路径错导致的 404
// （HTML 或其它非 token 报文）不能被误判成 token 失效，否则会掩盖配置错误。
func TestIsUpstreamTokenNotFound_Gateway404NotTokenError(t *testing.T) {
	html := []byte(`<html><head><title>404 Not Found</title></head><body>nginx</body></html>`)
	assert.False(t, isUpstreamTokenNotFound(html, http.StatusNotFound))
	assert.False(t, isUpstreamTokenNotFound([]byte(`{"code":"bad_request","message":"path not registered"}`), http.StatusNotFound))
}

// TestIsUpstreamTokenNotFound_SuccessBody 成功报文不能被误判。
func TestIsUpstreamTokenNotFound_SuccessBody(t *testing.T) {
	assert.False(t, isUpstreamTokenNotFound([]byte(upstreamSuccessBody), http.StatusOK))
	assert.False(t, isUpstreamTokenNotFound([]byte(`{"GroupId":"group-1"}`), http.StatusOK))
}

// TestExtractUpstreamErrorMessage 覆盖信封与常见错误结构。
func TestExtractUpstreamErrorMessage(t *testing.T) {
	assert.Contains(t, extractUpstreamErrorMessage([]byte(upstreamTokenNotFoundBody)), "NotFound.")
	assert.Contains(t, extractUpstreamErrorMessage([]byte(`{"error":{"code":"NotFound","message":"token missing"}}`)), "NotFound")
	assert.Equal(t, "", extractUpstreamErrorMessage([]byte(upstreamSuccessBody)))
	assert.Equal(t, "", extractUpstreamErrorMessage([]byte("<html>404</html>")))
}

// TestParseGroupID_ResultEnvelope 上游把 GroupId 放在 Result 里，必须能解出来。
func TestParseGroupID_ResultEnvelope(t *testing.T) {
	groupID, err := parseGroupID([]byte(upstreamSuccessBody))
	assert.NoError(t, err)
	assert.Equal(t, "group-20260924153205-lx298", groupID)
}
