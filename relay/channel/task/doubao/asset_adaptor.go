package doubao

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/service"
)

// 素材组 / 素材 上游 API 适配器
//
// 文档：dev-docs/cii-api.md
//   素材组: POST/GET/PUT/DELETE  /api/v1/asset-groups[/list]
//   素材  : POST/GET/PUT/DELETE  /api/v1/assets[/list]
//
// 文档上的字段命名在 create 和 list 之间不一致（create 用 AssetGroupId /
// AssetId，list 用 Id / AssetId），schema 也不同（create 平铺、list 包裹）。
// 本适配器负责做归一化、parse，并向上层返回与文档一致（透传）字段。
//
// 所有方法都是 best-effort：调用方拿到响应后用本地表作为事实源，
// 上游 4xx/5xx/超时由调用方按错误码语义翻译。

// defaultAssetGroupsApiPath 是上游"素材组"基础路径的默认值，匹配 dev-docs/cii-api.md。
const defaultAssetGroupsApiPath = "/api/v1/asset-groups"

// defaultAssetsApiPath 是上游"素材"基础路径的默认值，匹配 dev-docs/cii-api.md。
const defaultAssetsApiPath = "/api/v1/assets"

// normalizeAssetApiPath 校验 path 是否可安全地拼到 channel base URL 后面；
// 越界（过长、含 ..、含非法字符等）一律回退到 fallback。
// 与 normalizeTaskApiPath 同形；后续如有第三个 caller 再抽到 taskcommon。
func normalizeAssetApiPath(raw, fallback string) string {
	p := strings.TrimSpace(raw)
	if p == "" || len(p) > 256 {
		return fallback
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	for _, r := range p {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '.' || r == '_' || r == '~' || r == '/':
		default:
			return fallback
		}
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." {
			return fallback
		}
	}
	return p
}

// CreateGroup 调用上游 POST /api/v1/asset-groups。
//
//	body 由调用方构造（Name 必填 / Description 可选）。
//
// 返回上游响应 body（已读完整）和 HTTP 状态码；调用方负责 JSON parse。
func (a *TaskAdaptor) CreateGroup(baseURL, key, path string, body []byte, proxy string) ([]byte, int, error) {
	p := normalizeAssetApiPath(path, defaultAssetGroupsApiPath)
	return assetDoJSON(baseURL, p, key, http.MethodPost, body, proxy)
}

// ListGroups 调用上游 POST /api/v1/asset-groups/list（按 dev-docs，list 接口用 POST+body）。
func (a *TaskAdaptor) ListGroups(baseURL, key, path string, body []byte, proxy string) ([]byte, int, error) {
	p := normalizeAssetApiPath(path, defaultAssetGroupsApiPath)
	return assetDoJSON(baseURL, p+"/list", key, http.MethodPost, body, proxy)
}

// GetGroup 调用上游 GET /api/v1/asset-groups/{groupID}。
func (a *TaskAdaptor) GetGroup(baseURL, key, path, groupID string, proxy string) ([]byte, int, error) {
	p := normalizeAssetApiPath(path, defaultAssetGroupsApiPath)
	uri, err := joinAssetID(p, groupID)
	if err != nil {
		return nil, 0, err
	}
	return assetDoJSON(baseURL, uri, key, http.MethodGet, nil, proxy)
}

// UpdateGroup 调用上游 PUT /api/v1/asset-groups/{groupID}，body 由调用方构造。
func (a *TaskAdaptor) UpdateGroup(baseURL, key, path, groupID string, body []byte, proxy string) ([]byte, int, error) {
	p := normalizeAssetApiPath(path, defaultAssetGroupsApiPath)
	uri, err := joinAssetID(p, groupID)
	if err != nil {
		return nil, 0, err
	}
	return assetDoJSON(baseURL, uri, key, http.MethodPut, body, proxy)
}

// DeleteGroup 调用上游 DELETE /api/v1/asset-groups/{groupID}。
// 404 视为成功（幂等）；其他状态由调用方解析。
func (a *TaskAdaptor) DeleteGroup(baseURL, key, path, groupID string, proxy string) ([]byte, int, error) {
	p := normalizeAssetApiPath(path, defaultAssetGroupsApiPath)
	uri, err := joinAssetID(p, groupID)
	if err != nil {
		return nil, 0, err
	}
	return assetDoDelete(baseURL, uri, key, proxy)
}

// CreateAsset 调用上游 POST /api/v1/assets。
func (a *TaskAdaptor) CreateAsset(baseURL, key, path string, body []byte, proxy string) ([]byte, int, error) {
	p := normalizeAssetApiPath(path, defaultAssetsApiPath)
	return assetDoJSON(baseURL, p, key, http.MethodPost, body, proxy)
}

// ListAssets 调用上游 POST /api/v1/assets/list。
func (a *TaskAdaptor) ListAssets(baseURL, key, path string, body []byte, proxy string) ([]byte, int, error) {
	p := normalizeAssetApiPath(path, defaultAssetsApiPath)
	return assetDoJSON(baseURL, p+"/list", key, http.MethodPost, body, proxy)
}

// GetAsset 调用上游 GET /api/v1/assets/{assetID}。
func (a *TaskAdaptor) GetAsset(baseURL, key, path, assetID string, proxy string) ([]byte, int, error) {
	p := normalizeAssetApiPath(path, defaultAssetsApiPath)
	uri, err := joinAssetID(p, assetID)
	if err != nil {
		return nil, 0, err
	}
	return assetDoJSON(baseURL, uri, key, http.MethodGet, nil, proxy)
}

// UpdateAsset 调用上游 PUT /api/v1/assets/{assetID}，body 由调用方构造。
func (a *TaskAdaptor) UpdateAsset(baseURL, key, path, assetID string, body []byte, proxy string) ([]byte, int, error) {
	p := normalizeAssetApiPath(path, defaultAssetsApiPath)
	uri, err := joinAssetID(p, assetID)
	if err != nil {
		return nil, 0, err
	}
	return assetDoJSON(baseURL, uri, key, http.MethodPut, body, proxy)
}

// DeleteAsset 调用上游 DELETE /api/v1/assets/{assetID}。
// 404 视为成功（幂等）；其他状态由调用方解析。
func (a *TaskAdaptor) DeleteAsset(baseURL, key, path, assetID string, proxy string) ([]byte, int, error) {
	p := normalizeAssetApiPath(path, defaultAssetsApiPath)
	uri, err := joinAssetID(p, assetID)
	if err != nil {
		return nil, 0, err
	}
	return assetDoDelete(baseURL, uri, key, proxy)
}

// joinAssetID 把 path 和 id 拼起来（id 走与 path 类似的字符白名单）。
// 这样可以同时防止 path 注入和 id 注入。
func joinAssetID(path, id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("id is required")
	}
	if len(id) > 128 {
		return "", fmt.Errorf("id too long")
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '.' || r == '_':
		default:
			return "", fmt.Errorf("id contains invalid character: %q", r)
		}
	}
	return strings.TrimRight(path, "/") + "/" + id, nil
}

// assetDoJSON 通用方法：构造请求、附带 Bearer、读取上游响应 body。
// 返回 (body, statusCode, err)。HTTP 2xx 也返回 err=nil，仅负责传递状态码。
func assetDoJSON(baseURL, path, key, method string, body []byte, proxy string) ([]byte, int, error) {
	var rdr io.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, baseURL+path, rdr)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+key)

	client, err := service.GetHttpClientWithProxy(proxy)
	if err != nil {
		return nil, 0, fmt.Errorf("new proxy http client failed: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return respBody, resp.StatusCode, nil
}

// assetDoDelete 走 assetDoJSON + 把 404 视作成功的精简封装。
//
// 仍保留 raw body 透传：上游网关在 404 时返回 `{"error": {"code": 404, "message": "..."}}`
// 这种非标准结构，调用方可能要把原文/结构化错误回填给客户端。仅把 404
// 视作"目标已经不在了"，语义上等价于删除成功。
func assetDoDelete(baseURL, path, key, proxy string) (rawBody []byte, status int, err error) {
	body, st, err := assetDoJSON(baseURL, path, key, http.MethodDelete, nil, proxy)
	if err != nil {
		return body, st, err
	}
	if st == http.StatusNotFound {
		return body, st, nil
	}
	if st < 200 || st >= 300 {
		snippet := strings.TrimSpace(string(body))
		if len(snippet) > 500 {
			snippet = snippet[:500] + " ...(truncated)"
		}
		return body, st, fmt.Errorf("upstream delete failed: status=%d body=%s", st, snippet)
	}
	return body, st, nil
}
