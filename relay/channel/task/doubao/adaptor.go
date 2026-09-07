package doubao

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"

	"github.com/QuantumNous/new-api/constant"
	taskdto "github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay/channel"
	"github.com/QuantumNous/new-api/relay/channel/task/taskcommon"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

// ============================
// Request / Response structures
// ============================

type ContentItem struct {
	Type     string    `json:"type,omitempty"`
	Text     string    `json:"text,omitempty"`
	ImageURL *MediaURL `json:"image_url,omitempty"`
	VideoURL *MediaURL `json:"video_url,omitempty"`
	AudioURL *MediaURL `json:"audio_url,omitempty"`
	Role     string    `json:"role,omitempty"`
}

type MediaURL struct {
	URL string `json:"url,omitempty"`
}

type requestPayload struct {
	Model                 string         `json:"model"`
	Content               []ContentItem  `json:"content,omitempty"`
	CallbackURL           string         `json:"callback_url,omitempty"`
	ReturnLastFrame       *dto.BoolValue `json:"return_last_frame,omitempty"`
	ServiceTier           string         `json:"service_tier,omitempty"`
	ExecutionExpiresAfter *dto.IntValue  `json:"execution_expires_after,omitempty"`
	GenerateAudio         *dto.BoolValue `json:"generate_audio,omitempty"`
	Draft                 *dto.BoolValue `json:"draft,omitempty"`
	Tools                 []struct {
		Type string `json:"type,omitempty"`
	} `json:"tools,omitempty"`
	SafetyIdentifier string         `json:"safety_identifier,omitempty"`
	Priority         *dto.IntValue  `json:"priority,omitempty"`
	Resolution       string         `json:"resolution,omitempty"`
	Ratio            string         `json:"ratio,omitempty"`
	Duration         *dto.IntValue  `json:"duration,omitempty"`
	Frames           *dto.IntValue  `json:"frames,omitempty"`
	Seed             *dto.IntValue  `json:"seed,omitempty"`
	CameraFixed      *dto.BoolValue `json:"camera_fixed,omitempty"`
	Watermark        *dto.BoolValue `json:"watermark,omitempty"`
}

type responsePayload struct {
	ID string `json:"id"` // task_id
}

type responseTask struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Status  string `json:"status"`
	Content struct {
		VideoURL     string `json:"video_url"`
		LastFrameURL string `json:"last_frame_url"`
	} `json:"content"`
	Seed            int    `json:"seed"`
	Resolution      string `json:"resolution"`
	Duration        int    `json:"duration"`
	Ratio           string `json:"ratio"`
	Frames          int    `json:"frames"`
	FramesPerSecond int    `json:"framespersecond"`
	ServiceTier     string `json:"service_tier"`
	Tools           []struct {
		Type string `json:"type"`
	} `json:"tools"`
	Usage struct {
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
		ToolUsage        struct {
			WebSearch int `json:"web_search"`
		} `json:"tool_usage"`
	} `json:"usage"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}

// responseTaskList 是上游 GET /api/v1/contents/generations/tasks 的响应结构。
// items 字段是视频任务列表，total 是符合筛选条件的总任务数。
type responseTaskList struct {
	Items []responseTask `json:"items"`
	Total int            `json:"total"`
}

// ============================
// Adaptor implementation
// ============================

type TaskAdaptor struct {
	taskcommon.BaseBilling
	ChannelType int
	apiKey      string
	baseURL     string
}

func (a *TaskAdaptor) Init(info *relaycommon.RelayInfo) {
	a.ChannelType = info.ChannelType
	a.baseURL = info.ChannelBaseUrl
	a.apiKey = info.ApiKey
}

// ValidateRequestAndSetAction parses body, validates fields and sets default action.
func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) (taskErr *taskdto.TaskError) {
	// Accept only POST /v1/video/generations as "generate" action.
	return relaycommon.ValidateBasicTaskRequest(c, info, constant.TaskActionGenerate)
}

// BuildRequestURL constructs the upstream URL.
func (a *TaskAdaptor) BuildRequestURL(info *relaycommon.RelayInfo) (string, error) {
	return a.baseURL + a.taskApiPath(info), nil
}

// BuildRequestHeader sets required headers.
func (a *TaskAdaptor) BuildRequestHeader(_ *gin.Context, req *http.Request, _ *relaycommon.RelayInfo) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	return nil
}

// EstimateBilling 根据请求 metadata 中的输出分辨率、是否包含视频输入以及视频时长，
// 返回相对基准价的计费 OtherRatio。Seedance 按 token 计费，时长是主因子：
// - "video_input"：分辨率档 + 是否含视频输入（来自价格表）
// - "seconds"：用户显式请求的视频时长（秒），仅在客户端提供正数时写入；
//   未提供时（duration=-1 / 缺省）不写入，留给 AdjustBillingOnComplete 按
//   usage.completion_tokens 真实值结算。
func (a *TaskAdaptor) EstimateBilling(c *gin.Context, info *relaycommon.RelayInfo) map[string]float64 {
	req, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return nil
	}
	ratios := make(map[string]float64)

	hasVideo := hasVideoInMetadata(req.Metadata)
	resolution, _ := req.Metadata["resolution"].(string)
	if ratio, ok := GetVideoInputRatio(info.OriginModelName, resolution, hasVideo); ok && ratio != 1.0 {
		ratios["video_input"] = ratio
	}

	if seconds := resolveTaskSeconds(&req); seconds > 0 {
		// 与 sora remix 路径一致：作为计费乘数前必须钳制，避免 int 溢出为负
		if seconds > relaycommon.MaxTaskDurationSeconds {
			seconds = relaycommon.MaxTaskDurationSeconds
		}
		ratios["seconds"] = float64(seconds)
	}

	if len(ratios) == 0 {
		return nil
	}
	return ratios
}

// resolveTaskSeconds 从 TaskSubmitReq 中读出用户显式请求的视频时长（秒）。
// 优先取 int 字段 req.Duration，兼容字符串字段 req.Seconds；
// 与 convertToRequestPayload 中构造上游 duration 的取数策略保持一致。
func resolveTaskSeconds(req *relaycommon.TaskSubmitReq) int {
	if req == nil {
		return 0
	}
	if req.Duration > 0 {
		return req.Duration
	}
	if req.Seconds == "" {
		return 0
	}
	v, err := strconv.Atoi(req.Seconds)
	if err != nil {
		return 0
	}
	return v
}

// hasVideoInMetadata 直接检查 metadata 的 content 数组是否包含 video_url 条目，
// 避免构建完整的上游 requestPayload。
func hasVideoInMetadata(metadata map[string]interface{}) bool {
	if metadata == nil {
		return false
	}
	contentRaw, ok := metadata["content"]
	if !ok {
		return false
	}
	contentSlice, ok := contentRaw.([]interface{})
	if !ok {
		return false
	}
	for _, item := range contentSlice {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if itemMap["type"] == "video_url" {
			return true
		}
		if _, has := itemMap["video_url"]; has {
			return true
		}
	}
	return false
}

// BuildRequestBody converts request into Doubao specific format.
func (a *TaskAdaptor) BuildRequestBody(c *gin.Context, info *relaycommon.RelayInfo) (io.Reader, error) {
	req, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return nil, err
	}

	body, err := a.convertToRequestPayload(&req)
	if err != nil {
		return nil, errors.Wrap(err, "convert request payload failed")
	}
	if info.IsModelMapped {
		body.Model = info.UpstreamModelName
	} else {
		info.UpstreamModelName = body.Model
	}
	data, err := common.Marshal(body)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

// DoRequest delegates to common helper.
func (a *TaskAdaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	return channel.DoTaskApiRequest(a, c, info, requestBody)
}

// DoResponse handles upstream response, returns taskID etc.
func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (taskID string, taskData []byte, taskErr *taskdto.TaskError) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
		return
	}
	_ = resp.Body.Close()

	// Parse Doubao response
	var dResp responsePayload
	if err := common.Unmarshal(responseBody, &dResp); err != nil {
		taskErr = service.TaskErrorWrapper(errors.Wrapf(err, "body: %s", responseBody), "unmarshal_response_body_failed", http.StatusInternalServerError)
		return
	}

	if dResp.ID == "" {
		// Surface the upstream body instead of swallowing it, so a wrong path /
		// unexpected response shape is diagnosable from the error log.
		bodySnippet := strings.TrimSpace(string(responseBody))
		if len(bodySnippet) > 500 {
			bodySnippet = bodySnippet[:500] + " ...(truncated)"
		}
		if bodySnippet == "" {
			bodySnippet = "(empty body)"
		}
		taskErr = service.TaskErrorWrapper(fmt.Errorf("task_id is empty; upstream response: %s", bodySnippet), "invalid_response", http.StatusInternalServerError)
		return
	}

	ov := dto.NewOpenAIVideo()
	ov.ID = info.PublicTaskID
	ov.TaskID = info.PublicTaskID
	ov.CreatedAt = time.Now().Unix()
	ov.Model = info.OriginModelName

	c.JSON(http.StatusOK, ov)
	return dResp.ID, responseBody, nil
}

// FetchTask fetch task status
func (a *TaskAdaptor) FetchTask(baseUrl, key string, body map[string]any, proxy string) (*http.Response, error) {
	taskID, ok := body["task_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid task_id")
	}

	path := normalizeTaskApiPath(stringFromBody(body, "task_path"))
	uri := baseUrl + path + "/" + taskID

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	client, err := service.GetHttpClientWithProxy(proxy)
	if err != nil {
		return nil, fmt.Errorf("new proxy http client failed: %w", err)
	}
	return client.Do(req)
}

// ListTaskOptions 是 ListTask 的入参选项，零值即代表 "不传" ——
// 与上游 cii-group 平台的 query string 语义保持一致：
//   - 数值字段零值/负值/超过 [1,500] → 不带该参数
//   - 字符串字段空字符串 → 不带该参数
//
// 文档地址：https://www.cii-group.com/docs.html （Video 生成任务 / 列出任务）
type ListTaskOptions struct {
	PageNum     int      // 1-500，0 表示不带
	PageSize    int      // 1-500，0 表示不带
	Status      string   // 过滤状态：queued/running/cancelled/succeeded/failed/expired
	TaskIDs     []string // 精确搜索，支持多个
	Model       string   // 精确搜索（推理接入点 ID）
	ServiceTier string   // default/flex
}

// buildListQuery 根据 opts 拼出上游 query string。空值字段会被省略。
// page_num / page_size 限制在 [1, 500] 内，超出或零值会回退到不传。
func (opts ListTaskOptions) buildListQuery() string {
	values := url.Values{}
	if v := clampInt(opts.PageNum, 1, 500); v > 0 {
		values.Set("page_num", strconv.Itoa(v))
	}
	if v := clampInt(opts.PageSize, 1, 500); v > 0 {
		values.Set("page_size", strconv.Itoa(v))
	}
	if s := strings.TrimSpace(opts.Status); s != "" {
		values.Set("filter.status", s)
	}
	for _, id := range opts.TaskIDs {
		if id = strings.TrimSpace(id); id != "" {
			values.Add("filter.task_ids", id)
		}
	}
	if s := strings.TrimSpace(opts.Model); s != "" {
		values.Set("filter.model", s)
	}
	if s := strings.TrimSpace(opts.ServiceTier); s != "" {
		values.Set("filter.service_tier", s)
	}
	return values.Encode()
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return 0
	}
	if v > hi {
		return hi
	}
	return v
}

// ListTask 调用上游 GET /api/v1/contents/generations/tasks 接口，返回原始 JSON 响应体。
// 返回 []byte 是为了避免 service 包反向依赖 doubao 包的 responseTaskList 结构体；
// 上层 controller 在拿到字节体后，会按自己的 DTO（dto.UpstreamTaskListItem）反序列化。
// 仅在网络/反序列化/HTTP 非 2xx 时返回 error，成功时 body 即为上游的 JSON 原文。
func (a *TaskAdaptor) ListTask(baseURL, key, taskPath string, opts ListTaskOptions, proxy string) ([]byte, error) {
	path := normalizeTaskApiPath(taskPath)
	uri := baseURL + path
	if q := opts.buildListQuery(); q != "" {
		uri = uri + "?" + q
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	client, err := service.GetHttpClientWithProxy(proxy)
	if err != nil {
		return nil, fmt.Errorf("new proxy http client failed: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read list response body failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet := strings.TrimSpace(string(body))
		if len(snippet) > 500 {
			snippet = snippet[:500] + " ...(truncated)"
		}
		return nil, fmt.Errorf("upstream list task failed: status=%d body=%s", resp.StatusCode, snippet)
	}
	return body, nil
}

// DeleteTask 调用上游 DELETE /api/v1/contents/generations/tasks/{id}。
// 按任务状态不同，上游会产生不同副作用：
//   - queued      → 取消任务，状态变为 cancelled
//   - succeeded   → 删除任务记录
//   - failed      → 删除任务记录
//   - expired     → 删除任务记录
//   - running     → 不支持（上游返回错误）
//   - cancelled   → 不支持（上游返回错误）
//
// 返回 nil 表示上游处理成功（HTTP 2xx）；非 2xx 会把状态码 + body 摘要包成错误返回。
func (a *TaskAdaptor) DeleteTask(baseURL, key, taskPath, taskID string, proxy string) error {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return fmt.Errorf("task_id is required")
	}
	path := normalizeTaskApiPath(taskPath)
	uri := baseURL + path + "/" + taskID

	req, err := http.NewRequest(http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	client, err := service.GetHttpClientWithProxy(proxy)
	if err != nil {
		return fmt.Errorf("new proxy http client failed: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// 上游删除接口"无返回参数"——读完丢弃即可，但失败状态下需要把 body
	// 摘要附在错误信息里，方便排查为什么 queued 状态以外的删除会失败。
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		snippet := strings.TrimSpace(string(body))
		if len(snippet) > 500 {
			snippet = snippet[:500] + " ...(truncated)"
		}
		return fmt.Errorf("upstream delete task failed: status=%d body=%s", resp.StatusCode, snippet)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

func (a *TaskAdaptor) GetModelList() []string {
	return ModelList
}

// defaultDoubaoTaskApiPath is the upstream operations path assumed when the
// channel has not set TaskApiPath in OtherSettings. Kept as a single constant
// so future changes only touch one place.
const defaultDoubaoTaskApiPath = "/api/v3/contents/generations/tasks"

// taskApiPath returns the configured upstream path suffix (from the channel
// Base URL) for this adaptor. Falls back to defaultDoubaoTaskApiPath when the
// value is missing or fails validation.
func (a *TaskAdaptor) taskApiPath(info *relaycommon.RelayInfo) string {
	var raw string
	if info != nil && info.HasChannelMeta() {
		raw = info.ChannelOtherSettings.TaskApiPath
	}
	return normalizeTaskApiPath(raw)
}

// normalizeTaskApiPath restricts the path suffix to a safe relative URL path so
// it can be spliced onto the channel Base URL. Anything that is not a plain
// absolute path — e.g. containing ".." segments, backslashes, query/hash,
// schemes, whitespace or over 256 bytes — is replaced with the default. This
// protects against path injection from a misconfigured OtherSettings blob.
func normalizeTaskApiPath(raw string) string {
	p := strings.TrimSpace(raw)
	if p == "" || len(p) > 256 {
		return defaultDoubaoTaskApiPath
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
			return defaultDoubaoTaskApiPath
		}
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." {
			return defaultDoubaoTaskApiPath
		}
	}
	return p
}

// stringFromBody safely reads a string field from the FetchTask body map.
// Returns "" if the key is missing or the value is not a string.
func stringFromBody(body map[string]any, key string) string {
	if body == nil {
		return ""
	}
	if v, ok := body[key].(string); ok {
		return v
	}
	return ""
}

func (a *TaskAdaptor) GetChannelName() string {
	return ChannelName
}

func (a *TaskAdaptor) convertToRequestPayload(req *relaycommon.TaskSubmitReq) (*requestPayload, error) {
	r := requestPayload{
		Model:   req.Model,
		Content: []ContentItem{},
	}

	// Add images if present
	if req.HasImage() {
		for _, imgURL := range req.Images {
			r.Content = append(r.Content, ContentItem{
				Type: "image_url",
				ImageURL: &MediaURL{
					URL: imgURL,
				},
			})
		}
	}

	metadata := req.Metadata
	if err := taskcommon.UnmarshalMetadata(metadata, &r); err != nil {
		return nil, errors.Wrap(err, "unmarshal metadata failed")
	}

	if sec, _ := strconv.Atoi(req.Seconds); sec > 0 {
		r.Duration = lo.ToPtr(dto.IntValue(sec))
	}

	r.Content = lo.Reject(r.Content, func(c ContentItem, _ int) bool { return c.Type == "text" })
	r.Content = append(r.Content, ContentItem{
		Type: "text",
		Text: req.Prompt,
	})

	return &r, nil
}

func (a *TaskAdaptor) ParseTaskResult(respBody []byte) (*relaycommon.TaskInfo, error) {
	resTask := responseTask{}
	if err := common.Unmarshal(respBody, &resTask); err != nil {
		return nil, errors.Wrap(err, "unmarshal task result failed")
	}

	taskResult := relaycommon.TaskInfo{
		Code: 0,
	}

	// Map Doubao status to internal status
	switch resTask.Status {
	case "pending", "queued":
		taskResult.Status = model.TaskStatusQueued
		taskResult.Progress = "10%"
	case "processing", "running":
		taskResult.Status = model.TaskStatusInProgress
		taskResult.Progress = "50%"
	case "succeeded":
		taskResult.Status = model.TaskStatusSuccess
		taskResult.Progress = "100%"
		taskResult.Url = resTask.Content.VideoURL
		// 解析 usage 信息用于按倍率计费
		taskResult.CompletionTokens = resTask.Usage.CompletionTokens
		taskResult.TotalTokens = resTask.Usage.TotalTokens
	case "failed":
		taskResult.Status = model.TaskStatusFailure
		taskResult.Progress = "100%"
		taskResult.Reason = resTask.Error.Message
	default:
		// Unknown status, treat as processing
		taskResult.Status = model.TaskStatusInProgress
		taskResult.Progress = "30%"
	}

	return &taskResult, nil
}

func (a *TaskAdaptor) ConvertToOpenAIVideo(originTask *model.Task) ([]byte, error) {
	var dResp responseTask
	if err := common.Unmarshal(originTask.Data, &dResp); err != nil {
		return nil, errors.Wrap(err, "unmarshal doubao task data failed")
	}

	openAIVideo := dto.NewOpenAIVideo()
	openAIVideo.ID = originTask.TaskID
	openAIVideo.TaskID = originTask.TaskID
	openAIVideo.Status = originTask.Status.ToVideoStatus()
	openAIVideo.SetProgressStr(originTask.Progress)
	openAIVideo.SetMetadata("url", dResp.Content.VideoURL)
	openAIVideo.CreatedAt = originTask.CreatedAt
	openAIVideo.CompletedAt = originTask.UpdatedAt
	openAIVideo.Model = originTask.Properties.OriginModelName

	if dResp.Status == "failed" {
		openAIVideo.Error = &dto.OpenAIVideoError{
			Message: dResp.Error.Message,
			Code:    dResp.Error.Code,
		}
	}

	return common.Marshal(openAIVideo)
}
