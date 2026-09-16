package visualvalidate

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	taskdto "github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	taskdoubao "github.com/QuantumNous/new-api/relay/channel/task/doubao"
	"github.com/QuantumNous/new-api/setting/system_setting"

	assetservice "github.com/QuantumNous/new-api/service/asset"
)

// 真人审核 (Visual Validate) 业务编排层
//
// 文档：dev-docs/cii-api.md「真人审核 (Visual Validate)」
//
// 与素材体系（service/asset）对齐的策略：
//   - 本地表 visual_validate_sessions 是事实源（会话状态、拥有权、通知状态）
//   - 上游 BytedToken / GroupId 是缓存映射
//   - 创建会话：先调上游 sessions，拿到 BytedToken/H5Link 后再写本地；
//     CallbackURL 一律指向 new-api 自己的 /v1/visual-validate/callback 端点
//     （带 session public id + 一次性 sign token），保证回调结果能落库
//   - 回调处理：sign 校验 → 落库 resultCode → 成功时调上游 results 换
//     GroupId → 幂等注册本地 AssetGroup → best-effort 通知下游 callback_url
//   - 结果查询：幂等。已成功的会话直接返回本地映射，不再打上游

// vvSignTokenLen 回调防伪造 token 长度；与 randomIDKey 的 32 位对齐。
const vvSignTokenLen = 32

// vvCallbackPath 是 new-api 对外提供的回调端点路径（router 同步注册）。
const vvCallbackPath = "/v1/visual-validate/callback"

// vvNotifyTimeout 下游通知 / 上游调用的超时。
const vvNotifyTimeout = 10 * time.Second

// resultCodePassed 是上游约定的"认证通过"结果码。
const resultCodePassed = "10000"

// CreateSessionReq 是 POST /v1/visual-validate/sessions 的业务入参。
type CreateSessionReq struct {
	// CallbackURL 可选。调用方自己的回调地址；认证完成后 new-api 会
	// best-effort POST 一次通知。不传时调用方可轮询会话状态。
	CallbackURL string `json:"callback_url"`
	// ProjectName 可选，透传上游（默认 "default"）。
	ProjectName string `json:"project_name"`
}

// CreateSession 拉起一次真人认证：
//  1. 选定上游渠道
//  2. 生成本地 public id + sign token，拼出指向 new-api 的上游 CallbackURL
//  3. POST 上游 sessions
//  4. 落本地会话记录
func CreateSession(userID int, req CreateSessionReq) (*model.VisualValidateSession, *taskdto.TaskError) {
	callbackURL := strings.TrimSpace(req.CallbackURL)
	if callbackURL != "" && !strings.HasPrefix(callbackURL, "http://") && !strings.HasPrefix(callbackURL, "https://") {
		return nil, errBadRequest("callback_url 必须是 http(s) 地址")
	}

	ch, err := assetservice.PickAssetChannel()
	if err != nil {
		common.SysError("visualvalidate: pick channel failed: " + err.Error())
		return nil, errBadGateway("无可用 doubao/volcengine 渠道")
	}

	signToken, terr := generateSignToken()
	if terr != nil {
		return nil, terr
	}

	publicID := model.GenerateVisualValidateSessionID()
	upstreamCallback := buildUpstreamCallbackURL(publicID, signToken)

	body := mustMarshal(sessionReqBody(upstreamCallback, strings.TrimSpace(req.ProjectName)))
	upstreamBody, status, err := (&taskdoubao.TaskAdaptor{}).CreateVisualValidateSession(
		ch.GetBaseURL(), ch.Key, resolveVisualValidateApiPath(ch), body, ch.GetSetting().Proxy,
	)
	if err != nil {
		common.SysError("visualvalidate: upstream create session failed: " + err.Error())
		return nil, errBadGateway("上游创建真人认证会话失败")
	}
	if status < 200 || status >= 300 {
		return nil, errBadGateway("上游创建真人认证会话失败: " + snippetFromBody(upstreamBody))
	}

	token, h5link, perr := parseSessionResponse(upstreamBody)
	if perr != nil {
		common.SysError("visualvalidate: parse session response failed: " + perr.Error())
		return nil, errBadGateway("上游创建真人认证会话失败: " + perr.Error())
	}

	now := time.Now().Unix()
	s := &model.VisualValidateSession{
		PublicID:              publicID,
		SignToken:             signToken,
		UserID:                userID,
		ChannelID:             ch.Id,
		ChannelType:           ch.Type,
		UpstreamToken:         token,
		H5Link:                h5link,
		CallbackURL:           upstreamCallback,
		DownstreamCallbackURL: callbackURL,
		Status:                model.VisualValidateStatusPending,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	if err := s.Insert(); err != nil {
		common.SysError("visualvalidate: insert session failed: " + err.Error())
		return nil, errBadGateway("本地写入认证会话失败")
	}
	return s, nil
}

// CallbackResult 是回调处理的产出，供 controller 返回 HTML/JSON。
type CallbackResult struct {
	Session *model.VisualValidateSession
}

// HandleCallback 处理浏览器回调（GET /v1/visual-validate/callback）：
// sign 校验 → 首次处理时落库 → 成功时换 GroupId 并注册本地素材组 → 通知下游。
// 幂等：非 pending 状态的会话直接返回当前状态，不重复处理。
func HandleCallback(sessionPublicID, signToken, bytedToken, resultCode, algorithmBaseRespCode string) (*model.VisualValidateSession, *taskdto.TaskError) {
	sessionPublicID = strings.TrimSpace(sessionPublicID)
	signToken = strings.TrimSpace(signToken)
	if sessionPublicID == "" || signToken == "" {
		return nil, errBadRequest("session/sign 不能为空")
	}
	s, err := model.GetVisualValidateSessionByPublicID(sessionPublicID, 0)
	if err != nil {
		return nil, errNotFound("认证会话不存在")
	}
	// sign token 恒定比较，防伪造回调。
	if s.SignToken == "" || s.SignToken != signToken {
		return nil, errBadRequest("回调校验失败")
	}

	if s.Status != model.VisualValidateStatusPending {
		// 重复回调（浏览器刷新、H5 重试）按幂等处理。
		return s, nil
	}

	// 回调里的 bytedToken 与创建时返回的一致性仅做记录，不做强校验
	// （上游只在 resultCode=10000 的回调里带 token）。
	if resultCode != resultCodePassed {
		s.Status = model.VisualValidateStatusFailed
		s.ResultCode = resultCode
		s.UpdatedAt = time.Now().Unix()
		if err := s.Update(); err != nil {
			common.SysError("visualvalidate: update session failed: " + err.Error())
		}
		notifyDownstream(s)
		return s, nil
	}

	// 认证通过：凭 BytedToken 换 GroupId（上游素材组 ID）。
	groupID, upstreamErr := fetchUpstreamGroupID(s, bytedToken)
	if upstreamErr != nil {
		// 404（token 过期/已使用）→ expired；其余 → failed。
		if errors.Is(upstreamErr, errUpstreamTokenInvalid) {
			s.Status = model.VisualValidateStatusExpired
		} else {
			s.Status = model.VisualValidateStatusFailed
		}
		s.ResultCode = resultCode
		s.UpdatedAt = time.Now().Unix()
		if err := s.Update(); err != nil {
			common.SysError("visualvalidate: update session failed: " + err.Error())
		}
		notifyDownstream(s)
		return s, nil
	}

	s.Status = model.VisualValidateStatusSucceeded
	s.ResultCode = resultCode
	s.UpstreamGroupID = groupID
	s.UpdatedAt = time.Now().Unix()

	// 把上游真人素材组注册到本地，打通"认证 → 传素材 → 生成"闭环。
	// 失败不阻断会话成功状态——本地映射可由人工/重试补齐。
	var ch *model.Channel
	if full, cerr := model.GetChannelById(s.ChannelID, true); cerr == nil {
		ch = full
	}
	name := "真人素材组 " + groupID
	g, regErr := assetservice.RegisterUpstreamGroup(s.UserID, ch, groupID, name, "由真人审核自动创建")
	if regErr != nil {
		common.SysError("visualvalidate: register upstream group failed: " + regErr.Message)
	} else {
		s.LocalAssetGroupPublicID = g.PublicID
	}

	if err := s.Update(); err != nil {
		common.SysError("visualvalidate: update session failed: " + err.Error())
	}
	notifyDownstream(s)
	return s, nil
}

// GetSession 查询单个会话（拥有者校验）。
func GetSession(userID int, publicID string) (*model.VisualValidateSession, *taskdto.TaskError) {
	s, err := model.GetVisualValidateSessionByPublicID(publicID, userID)
	if err != nil {
		return nil, errNotFound("认证会话不存在")
	}
	return s, nil
}

// ListSessions 列出某个用户的认证会话。
func ListSessions(userID, pageNum, pageSize int) ([]*model.VisualValidateSession, int64, *taskdto.TaskError) {
	items, total, err := model.ListVisualValidateSessions(userID, pageNum, pageSize)
	if err != nil {
		common.SysError("visualvalidate: list sessions failed: " + err.Error())
		return nil, 0, errBadGateway("读取认证会话失败")
	}
	return items, total, nil
}

// GetResultReq 是 POST /v1/visual-validate/results 的业务入参。
// session_id 优先（走本地映射、幂等、自动注册素材组）；
// 只传 byted_token 时退化为纯透传（不落库、不注册素材组）。
type GetResultReq struct {
	SessionID   string `json:"session_id"`
	BytedToken  string `json:"byted_token"`
	ProjectName string `json:"project_name"`
}

// GetResult 查询认证结果（幂等）：
//   - 已成功会话：直接返回本地映射，不再打上游
//   - pending 会话：立即凭 token 换 GroupId（调用方不等回调也能拿结果）
//   - 仅 byted_token：纯透传
func GetResult(userID int, req GetResultReq) (*model.VisualValidateSession, string, *taskdto.TaskError) {
	sessionID := strings.TrimSpace(req.SessionID)
	token := strings.TrimSpace(req.BytedToken)

	if sessionID == "" && token == "" {
		return nil, "", errBadRequest("session_id / byted_token 至少传一个")
	}

	if sessionID == "" {
		// 纯透传：不落库。
		groupID, err := fetchUpstreamGroupIDByToken(token, strings.TrimSpace(req.ProjectName), pickAnyChannel)
		if err != nil {
			return nil, "", translateUpstreamResultError(err)
		}
		return nil, groupID, nil
	}

	s, terr := GetSession(userID, sessionID)
	if terr != nil {
		return nil, "", terr
	}
	if s.Status == model.VisualValidateStatusSucceeded && s.UpstreamGroupID != "" {
		return s, s.UpstreamGroupID, nil
	}
	if s.Status != model.VisualValidateStatusPending {
		return s, "", errBadRequest("会话已结束，状态: " + s.Status)
	}
	if s.UpstreamToken == "" {
		return s, "", errBadRequest("会话缺少 BytedToken，无法查询结果")
	}

	groupID, upstreamErr := fetchUpstreamGroupID(s, s.UpstreamToken)
	if upstreamErr != nil {
		return s, "", translateUpstreamResultError(upstreamErr)
	}

	s.Status = model.VisualValidateStatusSucceeded
	s.ResultCode = resultCodePassed
	s.UpstreamGroupID = groupID
	s.UpdatedAt = time.Now().Unix()

	var ch *model.Channel
	if full, cerr := model.GetChannelById(s.ChannelID, true); cerr == nil {
		ch = full
	}
	name := "真人素材组 " + groupID
	g, regErr := assetservice.RegisterUpstreamGroup(s.UserID, ch, groupID, name, "由真人审核自动创建")
	if regErr != nil {
		common.SysError("visualvalidate: register upstream group failed: " + regErr.Message)
	} else {
		s.LocalAssetGroupPublicID = g.PublicID
	}

	if err := s.Update(); err != nil {
		common.SysError("visualvalidate: update session failed: " + err.Error())
	}
	return s, groupID, nil
}

// ============================
// 上游调用 helpers
// ============================

// errUpstreamTokenInvalid 表示 BytedToken 过期/已使用（上游 404）。
var errUpstreamTokenInvalid = errors.New("byted token invalid or expired")

// pickAnyChannel 挑一个可用渠道（纯透传场景用）。
func pickAnyChannel() (*model.Channel, error) {
	return assetservice.PickAssetChannel()
}

// fetchUpstreamGroupIDByToken 用指定渠道选择器调上游 results。
func fetchUpstreamGroupIDByToken(token, projectName string, pickChannel func() (*model.Channel, error)) (string, error) {
	if token == "" {
		return "", errors.New("byted_token is required")
	}
	ch, err := pickChannel()
	if err != nil {
		return "", fmt.Errorf("pick channel: %w", err)
	}
	body := mustMarshal(resultReqBody(token, projectName))
	upstreamBody, status, err := (&taskdoubao.TaskAdaptor{}).GetVisualValidateResult(
		ch.GetBaseURL(), ch.Key, resolveVisualValidateApiPath(ch), body, ch.GetSetting().Proxy,
	)
	if err != nil {
		return "", fmt.Errorf("upstream results: %w", err)
	}
	if status == http.StatusNotFound {
		return "", errUpstreamTokenInvalid
	}
	if status < 200 || status >= 300 {
		return "", fmt.Errorf("upstream results status=%d body=%s", status, snippetFromBody(upstreamBody))
	}
	groupID, perr := parseGroupID(upstreamBody)
	if perr != nil {
		return "", perr
	}
	return groupID, nil
}

// fetchUpstreamGroupID 用会话记录的渠道调上游 results（回调路径用）。
func fetchUpstreamGroupID(s *model.VisualValidateSession, bytedToken string) (string, error) {
	token := strings.TrimSpace(bytedToken)
	if token == "" {
		token = s.UpstreamToken
	}
	return fetchUpstreamGroupIDByToken(token, "", func() (*model.Channel, error) {
		if ch, err := model.GetChannelById(s.ChannelID, true); err == nil && ch.Key != "" {
			return ch, nil
		}
		return assetservice.PickAssetChannel()
	})
}

// translateUpstreamResultError 把上游错误翻译成对外的 TaskError。
func translateUpstreamResultError(err error) *taskdto.TaskError {
	if err == nil {
		return nil
	}
	if errors.Is(err, errUpstreamTokenInvalid) {
		return errBadRequest("BytedToken 无效或已过期（有效期 30 分钟且仅能认证一次）")
	}
	common.SysError("visualvalidate: upstream result failed: " + err.Error())
	return errBadGateway("上游查询真人认证结果失败: " + err.Error())
}

// resolveVisualValidateApiPath 取 channel setting 里的 VisualValidateApiPath 覆盖值；
// 为空时由 doubao adaptor 回退到默认路径 /api/v1/visual-validate。
func resolveVisualValidateApiPath(ch *model.Channel) string {
	if ch == nil {
		return ""
	}
	return ch.GetOtherSettings().VisualValidateApiPath
}

// generateSignToken 生成回调端点防伪造 token。
func generateSignToken() (string, *taskdto.TaskError) {
	token, err := common.GenerateRandomCharsKey(vvSignTokenLen)
	if err != nil {
		common.SysError("visualvalidate: generate sign token failed: " + err.Error())
		return "", errBadGateway("生成回调校验 token 失败")
	}
	return token, nil
}

// buildUpstreamCallbackURL 拼出指向 new-api 的上游 CallbackURL。
// system_setting.ServerAddress 为空时回退 http://localhost:3000（与默认值一致）。
func buildUpstreamCallbackURL(sessionPublicID, signToken string) string {
	base := strings.TrimRight(system_setting.ServerAddress, "/")
	if base == "" {
		base = "http://localhost:3000"
	}
	return fmt.Sprintf("%s%s?session=%s&sign=%s", base, vvCallbackPath, sessionPublicID, signToken)
}

// sessionReqBody / resultReqBody 构造上游请求体；可选字段为空时不传，
// 让上游按文档默认值处理（如 ProjectName 默认 "default"）。
func sessionReqBody(callbackURL, projectName string) map[string]any {
	m := map[string]any{"CallbackURL": callbackURL}
	if projectName != "" {
		m["ProjectName"] = projectName
	}
	return m
}

func resultReqBody(token, projectName string) map[string]any {
	m := map[string]any{"BytedToken": token}
	if projectName != "" {
		m["ProjectName"] = projectName
	}
	return m
}

// notifyDownstream best-effort POST 通知调用方注册的 callback_url。
// body 为会话的对外 JSON 视图；失败仅记日志并更新 notify_status。
func notifyDownstream(s *model.VisualValidateSession) {
	url := strings.TrimSpace(s.DownstreamCallbackURL)
	if url == "" {
		return
	}
	body := mustMarshal(map[string]any{
		"id":                   s.PublicID,
		"status":               s.Status,
		"result_code":          s.ResultCode,
		"upstream_group_id":    s.UpstreamGroupID,
		"local_asset_group_id": s.LocalAssetGroupPublicID,
	})
	client := &http.Client{Timeout: vvNotifyTimeout}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		common.SysError("visualvalidate: notify downstream failed: " + err.Error())
		s.NotifyStatus = model.VisualValidateNotifyFailed
		s.UpdatedAt = time.Now().Unix()
		_ = s.Update()
		return
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		common.SysError(fmt.Sprintf("visualvalidate: notify downstream status=%d", resp.StatusCode))
		s.NotifyStatus = model.VisualValidateNotifyFailed
	} else {
		s.NotifyStatus = model.VisualValidateNotifyNotified
	}
	s.UpdatedAt = time.Now().Unix()
	_ = s.Update()
}

// parseSessionResponse 从上游 sessions 响应中解析 BytedToken / H5Link。
func parseSessionResponse(body []byte) (token, h5link string, err error) {
	var m map[string]any
	if err := common.Unmarshal(body, &m); err != nil {
		return "", "", err
	}
	if errMsg := extractEnvelopeError(m); errMsg != "" {
		return "", "", errors.New(errMsg)
	}
	if v, ok := m["BytedToken"].(string); ok {
		token = v
	}
	if v, ok := m["H5Link"].(string); ok {
		h5link = v
	}
	if token == "" || h5link == "" {
		return "", "", fmt.Errorf("BytedToken/H5Link missing, body=%s", snippetFromBody(body))
	}
	return token, h5link, nil
}

// parseGroupID 从上游 results 响应中解析 GroupId。
func parseGroupID(body []byte) (string, error) {
	var m map[string]any
	if err := common.Unmarshal(body, &m); err != nil {
		return "", err
	}
	if v, ok := m["GroupId"].(string); ok && v != "" {
		return v, nil
	}
	if result, ok := m["Result"].(map[string]any); ok {
		if v, ok := result["GroupId"].(string); ok && v != "" {
			return v, nil
		}
	}
	if errMsg := extractEnvelopeError(m); errMsg != "" {
		return "", errors.New(errMsg)
	}
	return "", fmt.Errorf("GroupId missing, body=%s", snippetFromBody(body))
}

// extractEnvelopeError 兼容 CII app-api 的信封错误格式：
// {"ResponseMetadata": {"Error": {"Code": "...", "Message": "..."}}}
// 返回空串表示不是信封错误。
func extractEnvelopeError(m map[string]any) string {
	meta, ok := m["ResponseMetadata"].(map[string]any)
	if !ok {
		return ""
	}
	errObj, ok := meta["Error"].(map[string]any)
	if !ok {
		return ""
	}
	code, _ := errObj["Code"].(string)
	message, _ := errObj["Message"].(string)
	return strings.TrimSpace(code + " " + message)
}

// snippetFromBody 把响应体裁短到 256 字符。
func snippetFromBody(body []byte) string {
	s := strings.TrimSpace(string(body))
	if len(s) > 256 {
		s = s[:256] + " ...(truncated)"
	}
	return s
}

// mustMarshal 序列化请求体；失败时记日志并返回 nil。
func mustMarshal(v any) []byte {
	b, err := common.Marshal(v)
	if err != nil {
		common.SysError("visualvalidate: marshal request body failed: " + err.Error())
		return nil
	}
	return b
}

// errBadRequest 构造 400 错误。
func errBadRequest(msg string) *taskdto.TaskError {
	return &taskdto.TaskError{Code: "bad_request", Message: msg, StatusCode: 400}
}

// errNotFound 构造 404 错误。
func errNotFound(msg string) *taskdto.TaskError {
	return &taskdto.TaskError{Code: "not_found", Message: msg, StatusCode: 404}
}

// errBadGateway 构造 502 错误。
func errBadGateway(msg string) *taskdto.TaskError {
	return &taskdto.TaskError{Code: "upstream_error", Message: msg, StatusCode: 502}
}
