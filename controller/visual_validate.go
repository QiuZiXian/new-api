package controller

import (
	"html"
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	vvservice "github.com/QuantumNous/new-api/service/visualvalidate"

	"github.com/gin-gonic/gin"
)

// 真人审核 (Visual Validate) 控制器层
//
// 路由由 router/video-router.go 注册：
//   - POST /v1/visual-validate/sessions         TokenAuth，拉起认证
//   - GET  /v1/visual-validate/sessions         TokenAuth，会话列表
//   - GET  /v1/visual-validate/sessions/:id     TokenAuth，会话详情（轮询用）
//   - POST /v1/visual-validate/results          TokenAuth，查询/换取结果
//   - GET  /v1/visual-validate/callback         公开（浏览器打开），回调落库
//
// 回调端点是终端用户认证完成后浏览器跳转打开的，因此不能挂鉴权；
// 防伪造靠 session 记录里的一次性 sign token（service 层校验）。

// CreateVisualValidateSessionReq 是 POST /v1/visual-validate/sessions 的入参。
type CreateVisualValidateSessionReq struct {
	CallbackURL string `json:"callback_url"`
	ProjectName string `json:"project_name"`
}

// VisualValidateResultReq 是 POST /v1/visual-validate/results 的入参。
type VisualValidateResultReq struct {
	SessionID   string `json:"session_id"`
	BytedToken  string `json:"byted_token"`
	ProjectName string `json:"project_name"`
}

// visualValidateSessionView 是会话的对外响应（不泄漏 sign token / 内部 ID）。
type visualValidateSessionView struct {
	ID                      string `json:"id"`
	BytedToken              string `json:"byted_token"`
	H5Link                  string `json:"h5_link"`
	CallbackURL             string `json:"callback_url"`
	Status                  string `json:"status"`
	ResultCode              string `json:"result_code,omitempty"`
	UpstreamGroupID         string `json:"upstream_group_id,omitempty"`
	LocalAssetGroupPublicID string `json:"local_asset_group_id,omitempty"`
	NotifyStatus            string `json:"notify_status,omitempty"`
	CreatedAt               int64  `json:"created_at"`
	UpdatedAt               int64  `json:"updated_at"`
}

func toVisualValidateSessionView(s *model.VisualValidateSession) *visualValidateSessionView {
	if s == nil {
		return nil
	}
	return &visualValidateSessionView{
		ID:                      s.PublicID,
		BytedToken:              s.UpstreamToken,
		H5Link:                  s.H5Link,
		CallbackURL:             s.CallbackURL,
		Status:                  s.Status,
		ResultCode:              s.ResultCode,
		UpstreamGroupID:         s.UpstreamGroupID,
		LocalAssetGroupPublicID: s.LocalAssetGroupPublicID,
		NotifyStatus:            s.NotifyStatus,
		CreatedAt:               s.CreatedAt,
		UpdatedAt:               s.UpdatedAt,
	}
}

// CreateVisualValidateSession 处理 POST /v1/visual-validate/sessions。
func CreateVisualValidateSession(c *gin.Context) {
	userID := c.GetInt("id")
	var req CreateVisualValidateSessionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAssetError(c, http.StatusBadRequest, common.ErrorCodeBadRequest, "请求体解析失败: "+err.Error())
		return
	}
	s, terr := vvservice.CreateSession(userID, vvservice.CreateSessionReq{
		CallbackURL: req.CallbackURL,
		ProjectName: req.ProjectName,
	})
	if terr != nil {
		respondTaskError(c, terr)
		return
	}
	c.JSON(http.StatusOK, toVisualValidateSessionView(s))
}

// ListVisualValidateSessions 处理 GET /v1/visual-validate/sessions。
func ListVisualValidateSessions(c *gin.Context) {
	userID := c.GetInt("id")
	pageNum := atoiOrDefault(c.Query("page_num"), 1)
	pageSize := atoiOrDefault(c.Query("page_size"), 20)
	items, total, terr := vvservice.ListSessions(userID, pageNum, pageSize)
	if terr != nil {
		respondTaskError(c, terr)
		return
	}
	views := make([]*visualValidateSessionView, 0, len(items))
	for _, s := range items {
		views = append(views, toVisualValidateSessionView(s))
	}
	c.JSON(http.StatusOK, gin.H{
		"items":     views,
		"total":     total,
		"page_num":  pageNum,
		"page_size": pageSize,
	})
}

// GetVisualValidateSession 处理 GET /v1/visual-validate/sessions/:id。
func GetVisualValidateSession(c *gin.Context) {
	userID := c.GetInt("id")
	s, terr := vvservice.GetSession(userID, c.Param("session_id"))
	if terr != nil {
		respondTaskError(c, terr)
		return
	}
	c.JSON(http.StatusOK, toVisualValidateSessionView(s))
}

// GetVisualValidateResult 处理 POST /v1/visual-validate/results。
// session_id 优先；只传 byted_token 时退化为纯透传（响应里只有 group_id）。
func GetVisualValidateResult(c *gin.Context) {
	userID := c.GetInt("id")
	var req VisualValidateResultReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAssetError(c, http.StatusBadRequest, common.ErrorCodeBadRequest, "请求体解析失败: "+err.Error())
		return
	}
	s, groupID, terr := vvservice.GetResult(userID, vvservice.GetResultReq{
		SessionID:   req.SessionID,
		BytedToken:  req.BytedToken,
		ProjectName: req.ProjectName,
	})
	if terr != nil {
		respondTaskError(c, terr)
		return
	}
	if s == nil {
		// 纯透传场景：没有本地会话。
		c.JSON(http.StatusOK, gin.H{"group_id": groupID})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"session":  toVisualValidateSessionView(s),
		"group_id": groupID,
	})
}

// VisualValidateCallback 处理 GET /v1/visual-validate/callback（公开端点）。
// 终端用户在 H5 完成认证后浏览器跳转到这里；返回一个极简 HTML 页面，
// 供终端用户看到认证受理结果。业务方通过轮询会话接口或下游回调拿结构化结果。
func VisualValidateCallback(c *gin.Context) {
	s, terr := vvservice.HandleCallback(
		c.Query("session"),
		c.Query("sign"),
		c.Query("bytedToken"),
		c.Query("resultCode"),
		c.Query("algorithmBaseRespCode"),
	)
	if terr != nil {
		// 浏览器打开的场景，直接回 HTML 错误页（不含内部细节）。
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8",
			[]byte(callbackHTML("认证回调处理失败", "请联系服务提供方排查。")))
		return
	}
	title, detail := callbackViewText(s)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(callbackHTML(title, detail)))
}

// callbackViewText 把会话状态翻译成给终端用户看的文案。
func callbackViewText(s *model.VisualValidateSession) (title, detail string) {
	switch s.Status {
	case model.VisualValidateStatusSucceeded:
		return "真人认证成功", "你可以关闭本页面并返回应用继续操作。"
	case model.VisualValidateStatusFailed:
		code := s.ResultCode
		if code == "" {
			code = "未知"
		}
		return "真人认证未通过", "结果码 " + html.EscapeString(code) + "，如需重试请返回应用重新发起认证。"
	case model.VisualValidateStatusExpired:
		return "认证已过期", "认证凭证超过有效期，请返回应用重新发起认证。"
	default:
		return "认证受理中", "请稍后在应用内刷新认证状态。"
	}
}

// callbackHTML 生成回调端点返回的极简 HTML 页。
func callbackHTML(title, detail string) string {
	return `<!DOCTYPE html><html lang="zh"><head><meta charset="utf-8">` +
		`<meta name="viewport" content="width=device-width,initial-scale=1">` +
		`<title>` + html.EscapeString(title) + `</title>` +
		`<style>body{font-family:system-ui,-apple-system,sans-serif;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0;background:#f5f5f5;color:#333}` +
		`.card{background:#fff;padding:32px 40px;border-radius:12px;box-shadow:0 1px 4px rgba(0,0,0,.08);text-align:center;max-width:360px}` +
		`h1{font-size:18px;margin:0 0 12px}p{font-size:14px;margin:0;color:#666}</style></head>` +
		`<body><div class="card"><h1>` + html.EscapeString(title) + `</h1><p>` + detail + `</p></div></body></html>`
}
