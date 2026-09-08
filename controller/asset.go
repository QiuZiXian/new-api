package controller

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	assetservice "github.com/QuantumNous/new-api/service/asset"

	"github.com/gin-gonic/gin"
)

// 素材组 / 素材 控制器层
//
// 路由由 router/video-router.go 注册，全部走 middleware.TokenAuth()，
// userId 通过 c.GetInt("id") 拿到。所有读写都按 user_id 限定，
// 业务编排交给 service/asset 包；这里只负责 HTTP 入参解析和响应。
//
// 错误响应统一沿用 controller/task_content.go 的 taskdto.TaskError 结构，
// 便于客户端复用 task_content 的错误处理逻辑。

// CreateAssetGroupReq 是 POST /v1/asset-groups 的入参。
type CreateAssetGroupReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateAssetGroupReq 是 PUT /v1/asset-groups/:group_id 的入参。
// 字段可选——只传 name 只改 name；同时传 name/desc 都改；两个都空返回 400。
type UpdateAssetGroupReq struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// CreateAssetReq 是 POST /v1/asset-groups/:group_id/assets 的入参。
// ImageURL 必填——OSS 接入前，调用方需自己上传并提供可被上游访问的公网 URL
// （dev-docs/cii-api.md 要求 ImageUrl 必填）。
type CreateAssetReq struct {
	ImageURL  string `json:"image_url"`
	AssetType string `json:"asset_type"`
	Name      string `json:"name"`
}

// UpdateAssetReq 是 PUT /v1/assets/:asset_id 的入参。
// 当前 dev-docs 2.1 仅描述 name 可更新。
type UpdateAssetReq struct {
	Name *string `json:"name"`
}

// ============================
// 素材组
// ============================

// CreateAssetGroup 处理 POST /v1/asset-groups。
func CreateAssetGroup(c *gin.Context) {
	userID := c.GetInt("id")
	var req CreateAssetGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAssetError(c, http.StatusBadRequest, "bad_request", "请求体解析失败: "+err.Error())
		return
	}
	g, terr := assetservice.CreateAssetGroup(userID, req.Name, req.Description)
	if terr != nil {
		respondTaskError(c, terr)
		return
	}
	c.JSON(http.StatusOK, toAssetGroupView(g))
}

// ListAssetGroups 处理 GET /v1/asset-groups。
func ListAssetGroups(c *gin.Context) {
	userID := c.GetInt("id")
	pageNum := atoiOrDefault(c.Query("page_num"), 1)
	pageSize := atoiOrDefault(c.Query("page_size"), 20)
	items, total, terr := assetservice.ListAssetGroups(userID, pageNum, pageSize)
	if terr != nil {
		respondTaskError(c, terr)
		return
	}
	views := make([]*assetGroupView, 0, len(items))
	for _, g := range items {
		views = append(views, toAssetGroupView(g))
	}
	c.JSON(http.StatusOK, gin.H{
		"items":     views,
		"total":     total,
		"page_num":  pageNum,
		"page_size": pageSize,
	})
}

// GetAssetGroup 处理 GET /v1/asset-groups/:group_id。
func GetAssetGroup(c *gin.Context) {
	userID := c.GetInt("id")
	publicID := c.Param("group_id")
	g, terr := assetservice.GetAssetGroup(userID, publicID)
	if terr != nil {
		respondTaskError(c, terr)
		return
	}
	c.JSON(http.StatusOK, toAssetGroupView(g))
}

// UpdateAssetGroup 处理 PUT /v1/asset-groups/:group_id。
func UpdateAssetGroup(c *gin.Context) {
	userID := c.GetInt("id")
	publicID := c.Param("group_id")
	var req UpdateAssetGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAssetError(c, http.StatusBadRequest, "bad_request", "请求体解析失败: "+err.Error())
		return
	}
	if req.Name == nil && req.Description == nil {
		respondAssetError(c, http.StatusBadRequest, "bad_request", "name/description 至少需要传一个")
		return
	}
	name := ""
	if req.Name != nil {
		name = *req.Name
	}
	desc := ""
	if req.Description != nil {
		desc = *req.Description
	}
	g, terr := assetservice.UpdateAssetGroup(userID, publicID, name, desc)
	if terr != nil {
		respondTaskError(c, terr)
		return
	}
	c.JSON(http.StatusOK, toAssetGroupView(g))
}

// DeleteAssetGroup 处理 DELETE /v1/asset-groups/:group_id。
// 本地软删幂等；上游删除为 best-effort（在 service 内部异步执行）。
func DeleteAssetGroup(c *gin.Context) {
	userID := c.GetInt("id")
	publicID := c.Param("group_id")
	terr := assetservice.DeleteAssetGroup(userID, publicID)
	if terr != nil {
		respondTaskError(c, terr)
		return
	}
	c.Status(http.StatusNoContent)
}

// ============================
// 素材
// ============================

// CreateAsset 处理 POST /v1/asset-groups/:group_id/assets。
func CreateAsset(c *gin.Context) {
	userID := c.GetInt("id")
	groupID := c.Param("group_id")
	var req CreateAssetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAssetError(c, http.StatusBadRequest, "bad_request", "请求体解析失败: "+err.Error())
		return
	}
	a, terr := assetservice.CreateAsset(userID, groupID, req.ImageURL, req.AssetType, req.Name)
	if terr != nil {
		respondTaskError(c, terr)
		return
	}
	c.JSON(http.StatusOK, toAssetView(a))
}

// ListAssets 处理 GET /v1/asset-groups/:group_id/assets。
func ListAssets(c *gin.Context) {
	userID := c.GetInt("id")
	groupID := c.Param("group_id")
	pageNum := atoiOrDefault(c.Query("page_num"), 1)
	pageSize := atoiOrDefault(c.Query("page_size"), 20)
	status := c.Query("status")
	items, total, terr := assetservice.ListAssets(userID, pageNum, pageSize, groupID, status)
	if terr != nil {
		respondTaskError(c, terr)
		return
	}
	views := make([]*assetView, 0, len(items))
	for _, a := range items {
		views = append(views, toAssetView(a))
	}
	c.JSON(http.StatusOK, gin.H{
		"items":     views,
		"total":     total,
		"page_num":  pageNum,
		"page_size": pageSize,
	})
}

// GetAsset 处理 GET /v1/assets/:asset_id。
func GetAsset(c *gin.Context) {
	userID := c.GetInt("id")
	publicID := c.Param("asset_id")
	a, terr := assetservice.GetAsset(userID, publicID)
	if terr != nil {
		respondTaskError(c, terr)
		return
	}
	c.JSON(http.StatusOK, toAssetView(a))
}

// UpdateAsset 处理 PUT /v1/assets/:asset_id。
func UpdateAsset(c *gin.Context) {
	userID := c.GetInt("id")
	publicID := c.Param("asset_id")
	var req UpdateAssetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAssetError(c, http.StatusBadRequest, "bad_request", "请求体解析失败: "+err.Error())
		return
	}
	if req.Name == nil {
		respondAssetError(c, http.StatusBadRequest, "bad_request", "name 不能为空")
		return
	}
	a, terr := assetservice.UpdateAsset(userID, publicID, *req.Name)
	if terr != nil {
		respondTaskError(c, terr)
		return
	}
	c.JSON(http.StatusOK, toAssetView(a))
}

// DeleteAsset 处理 DELETE /v1/assets/:asset_id。
func DeleteAsset(c *gin.Context) {
	userID := c.GetInt("id")
	publicID := c.Param("asset_id")
	terr := assetservice.DeleteAsset(userID, publicID)
	if terr != nil {
		respondTaskError(c, terr)
		return
	}
	c.Status(http.StatusNoContent)
}

// ============================
// View / helpers
// ============================

// assetGroupView 把 model.AssetGroup 转成对外响应（不泄漏 channel 内部字段）。
type assetGroupView struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	GroupType       string `json:"group_type"`
	UpstreamAssetID string `json:"upstream_asset_group_id,omitempty"`
	CreatedAt       int64  `json:"created_at"`
	UpdatedAt       int64  `json:"updated_at"`
}

func toAssetGroupView(g *model.AssetGroup) *assetGroupView {
	if g == nil {
		return nil
	}
	return &assetGroupView{
		ID:              g.PublicID,
		Name:            g.Name,
		Description:     g.Description,
		GroupType:       g.GroupType,
		UpstreamAssetID: g.UpstreamAssetGroupID,
		CreatedAt:       g.CreatedAt,
		UpdatedAt:       g.UpdatedAt,
	}
}

// assetView 把 model.Asset 转成对外响应。GroupID 字段暂留空字符串
// ——客户端可以通过外层 group context 拿到 group 公开 ID，避免再
// 暴露本地表自增 ID。
type assetView struct {
	ID              string `json:"id"`
	GroupID         string `json:"group_id,omitempty"`
	AssetType       string `json:"asset_type"`
	Name            string `json:"name"`
	SourceURL       string `json:"source_url"`
	Status          string `json:"status"`
	SizeBytes       int64  `json:"size_bytes,omitempty"`
	MimeType        string `json:"mime_type,omitempty"`
	UpstreamAssetID string `json:"upstream_asset_id,omitempty"`
	CreatedAt       int64  `json:"created_at"`
	UpdatedAt       int64  `json:"updated_at"`
}

func toAssetView(a *model.Asset) *assetView {
	if a == nil {
		return nil
	}
	return &assetView{
		ID:              a.PublicID,
		GroupID:         "",
		AssetType:       a.AssetType,
		Name:            a.Name,
		SourceURL:       a.SourceURL,
		Status:          a.Status,
		SizeBytes:       a.SizeBytes,
		MimeType:        a.MimeType,
		UpstreamAssetID: a.UpstreamAssetID,
		CreatedAt:       a.CreatedAt,
		UpdatedAt:       a.UpdatedAt,
	}
}

// respondAssetError 写错误响应。
func respondAssetError(c *gin.Context, status int, code, message string) {
	c.JSON(status, &dto.TaskError{Code: code, Message: message, StatusCode: status})
}

// atoiOrDefault 把字符串转 int；失败/空时返回 def。
func atoiOrDefault(raw string, def int) int {
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 1 {
		return def
	}
	return v
}
