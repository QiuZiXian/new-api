package doubao

import "net/http"

// 真人审核 (Visual Validate) 上游 API 适配器
//
// 文档：dev-docs/cii-api.md「真人审核 (Visual Validate)」
//   拉起认证: POST {base}{path}/sessions   入参 CallbackURL(必填)/ProjectName(可选)
//             出参 BytedToken / H5Link / CallbackURL
//   获取结果: POST {base}{path}/results    入参 BytedToken(必填)/ProjectName(可选)
//             出参 GroupId（上游真人素材组 ID）
//
// 两个接口都是 POST + JSON body + Bearer 鉴权，与素材组/素材同构，
// 因此直接复用 normalizeAssetApiPath（路径白名单）和 assetDoJSON（请求封装）。
// 注意 BytedToken 可能含 '*'，但它只出现在 body 里、不参与 path 拼接，
// 不受 joinAssetID 白名单约束。

// defaultVisualValidateApiPath 是上游"真人审核"基础路径的默认值，匹配 dev-docs/cii-api.md。
const defaultVisualValidateApiPath = "/api/v1/visual-validate"

// CreateVisualValidateSession 调用上游 POST {path}/sessions 拉起真人认证。
//
//	body 由调用方构造（CallbackURL 必填 / ProjectName 可选）。
//	返回上游响应 body（已读完整）和 HTTP 状态码；调用方负责 JSON parse。
func (a *TaskAdaptor) CreateVisualValidateSession(baseURL, key, path string, body []byte, proxy string) ([]byte, int, error) {
	p := normalizeAssetApiPath(path, defaultVisualValidateApiPath)
	return assetDoJSON(baseURL, p+"/sessions", key, http.MethodPost, body, proxy)
}

// GetVisualValidateResult 调用上游 POST {path}/results 凭 BytedToken 换 GroupId。
func (a *TaskAdaptor) GetVisualValidateResult(baseURL, key, path string, body []byte, proxy string) ([]byte, int, error) {
	p := normalizeAssetApiPath(path, defaultVisualValidateApiPath)
	return assetDoJSON(baseURL, p+"/results", key, http.MethodPost, body, proxy)
}
