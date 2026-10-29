package doubao

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
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
	"github.com/QuantumNous/new-api/setting/ratio_setting"

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
		PromptTokens     int `json:"prompt_tokens"`
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
// 返回相对基准价的计费 OtherRatio。按秒 / 按 token 两种模式的乘数严格分开：
//   - 按秒模式（模型配置了分档秒价）：
//   - "video_variant"：命中档位价 / 基准秒价
//   - "seconds"：用户显式请求的视频时长（秒）
//     两者只参与「展示 + 预扣」，终态结算以上游 token 为准（见 MarkOtherRatioEstimateOnly）。
//   - 按 token 模式：
//   - "video_input"：分辨率档 + 是否含视频输入带来的单价差（来自价格表），参与结算
//   - "seconds"：仅在无法按官方公式估算 token 时兜底，用于放大预扣冻结额度，
//     结算不参与（上游 token 已含时长）
//
// 按 token 模式下若能估出 token（分辨率 + 时长可确定），会把预估 token 写入
// info.PriceData.EstimatedTokens，由 relay 层据此计算预扣基数。
func (a *TaskAdaptor) EstimateBilling(c *gin.Context, info *relaycommon.RelayInfo) map[string]float64 {
	req, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return nil
	}
	ratios := make(map[string]float64)
	resolution, _ := req.Metadata["resolution"].(string)

	// 按秒计费：分档价格本身已含分辨率差异，这里只补时长与「当前档位 / 基准档位」
	// 的倍率，不再叠加渠道内置的 video_input 倍率表（那是按 token 计费的口径）。
	if info.PriceData.UseVideoSecond {
		if ratio, ok := ratio_setting.GetVideoSecondVariantRatio(info.OriginModelName, resolution); ok {
			ratios["video_variant"] = ratio
		}
		if seconds := resolveTaskSeconds(&req); seconds > 0 {
			if seconds > relaycommon.MaxTaskDurationSeconds {
				seconds = relaycommon.MaxTaskDurationSeconds
			}
			ratios["seconds"] = float64(seconds)
		}
		if len(ratios) == 0 {
			return nil
		}
		// 按秒模式的乘数只服务于「展示 + 预扣」：终态结算以上游真实 token 为准，
		// 档位倍率与时长都不能再参与结算连乘。
		info.PriceData.MarkOtherRatioEstimateOnly("video_variant", "seconds")
		return ratios
	}

	hasVideo := hasVideoInMetadata(req.Metadata)
	if ratio, ok := GetVideoInputRatio(info.OriginModelName, resolution, hasVideo); ok && ratio != 1.0 {
		ratios["video_input"] = ratio
	}

	seconds := resolveTaskSeconds(&req)
	if seconds > relaycommon.MaxTaskDurationSeconds {
		// 与 sora remix 路径一致：作为计费乘数前必须钳制，避免 int 溢出为负
		seconds = relaycommon.MaxTaskDurationSeconds
	}
	ratio, _ := req.Metadata["ratio"].(string)
	// 按 token 计费：用官方公式直接估算 token 用量，替换「25 万 token」的默认粗估基数。
	// 预估 token 已包含时长维度，因此不再把 seconds 作为独立乘数。
	if tokens := estimateVideoTokens(resolution, ratio, seconds); tokens > 0 {
		info.PriceData.EstimatedTokens = tokens
		return ratiosIfAny(ratios)
	}

	// 无法估算（分辨率未知等）时回退到「时长放大」：仅影响预扣冻结额度，
	// 终态结算仍以上游真实 token 为准。
	if seconds > 0 {
		ratios["seconds"] = float64(seconds)
		info.PriceData.MarkOtherRatioEstimateOnly("seconds")
	}
	return ratiosIfAny(ratios)
}

// ratiosIfAny 避免返回空 map，调用方以 nil 表示「没有附加倍率」。
func ratiosIfAny(ratios map[string]float64) map[string]float64 {
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

// arkVideoTokenDivisor 是火山方舟视频模型的 token 换算除数：
//
//	token 用量 = (输入视频时长 + 输出视频时长) × 输出宽 × 输出高 × 输出帧率 / 1024
const arkVideoTokenDivisor = 1024

// defaultVideoOutputFPS 是 Seedance 未显式指定帧率时的输出帧率。
const defaultVideoOutputFPS = 24

// defaultVideoDurationSeconds 是请求未带时长时上游的默认生成时长（Seedance 默认 5 秒）。
const defaultVideoDurationSeconds = 5

// defaultVideoResolution 是请求未带分辨率时上游的默认输出档（Seedance 默认 720p）。
const defaultVideoResolution = "720p"

// videoResolutionBaseDims 是各分辨率档在 16:9 下的官方输出尺寸，
// 用于还原官方 token 公式里的「输出宽 × 输出高」。
var videoResolutionBaseDims = map[string][2]int{
	"480p":  {864, 480},
	"720p":  {1280, 720},
	"1080p": {1920, 1080},
	"4k":    {3840, 2160},
}

// estimateVideoTokens 按火山官方公式估算一次视频生成消耗的 token：
//
//	tokens = (输入视频时长 + 输出视频时长) × 输出宽 × 输出高 × 输出帧率 / 1024
//
// 请求里只有输出时长，输入视频时长未知：含视频输入的场景会略微低估，
// 预扣阶段可接受（终态仍以上游回报的真实 usage 结算）。
// 分辨率未知时按上游默认 720p、时长未知时按默认 5 秒估算，保证预扣不至于
// 用「25 万 token」的通用粗估冻结出十倍金额。
func estimateVideoTokens(resolution, ratio string, seconds int) int {
	if resolution == "" {
		resolution = defaultVideoResolution
	}
	if seconds <= 0 {
		seconds = defaultVideoDurationSeconds
	}
	if seconds > relaycommon.MaxTaskDurationSeconds {
		seconds = relaycommon.MaxTaskDurationSeconds
	}
	base, ok := videoResolutionBaseDims[strings.ToLower(strings.TrimSpace(resolution))]
	if !ok {
		return 0
	}
	width, height := resolveOutputDims(base[0], base[1], ratio)
	if width <= 0 || height <= 0 {
		return 0
	}
	tokens := float64(seconds) * float64(width) * float64(height) * defaultVideoOutputFPS / arkVideoTokenDivisor
	if tokens < 1 {
		return 0
	}
	return int(math.Round(tokens))
}

// resolveOutputDims 按宽高比把 16:9 基准尺寸换算成实际输出尺寸。
// ratio 形如 "16:9" / "9:16" / "1:1"；缺省或非数值（如 adaptive）时按 16:9 处理。
func resolveOutputDims(baseWidth, baseHeight int, ratio string) (int, int) {
	ratio = strings.ToLower(strings.TrimSpace(ratio))
	if ratio == "" || ratio == "adaptive" || ratio == "auto" {
		return baseWidth, baseHeight
	}
	parts := strings.Split(ratio, ":")
	if len(parts) != 2 {
		return baseWidth, baseHeight
	}
	w, errW := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	h, errH := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if errW != nil || errH != nil || w <= 0 || h <= 0 {
		return baseWidth, baseHeight
	}
	if w >= h {
		return int(math.Round(float64(baseHeight) * w / h)), baseHeight
	}
	// 竖屏：短边仍是 baseHeight（如 720p 9:16 = 720×1280）
	return baseHeight, int(math.Round(float64(baseHeight) * h / w))
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

	path := resolveTaskApiPath(body, "task_query_path", "task_path")
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

// isValidTaskApiPath reports whether raw is a safe relative URL path suffix:
// length-capped, only URL-safe characters, no ".." segment. Splitting this out
// of normalizeTaskApiPath lets callers tell "not configured" apart from
// "configured to something unusable", which is what path fallback needs.
func isValidTaskApiPath(p string) bool {
	if p == "" || len(p) > 256 {
		return false
	}
	for _, r := range p {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '.' || r == '_' || r == '~' || r == '/':
		default:
			return false
		}
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." {
			return false
		}
	}
	return true
}

// normalizeTaskApiPath restricts the path suffix to a safe relative URL path so
// it can be spliced onto the channel Base URL. Anything that is not a plain
// absolute path — e.g. containing ".." segments, backslashes, query/hash,
// schemes, whitespace or over 256 bytes — is replaced with the default. This
// protects against path injection from a misconfigured OtherSettings blob.
func normalizeTaskApiPath(raw string) string {
	p := strings.TrimSpace(raw)
	if !isValidTaskApiPath(p) {
		return defaultDoubaoTaskApiPath
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

// resolveTaskApiPath picks the path suffix named by primaryKey and falls back to
// fallbackKey when it is unset or unusable. Upstreams that serve create and
// query from one path leave the query path empty, so they transparently fall
// back to the create path.
func resolveTaskApiPath(body map[string]any, primaryKey, fallbackKey string) string {
	if raw := strings.TrimSpace(stringFromBody(body, primaryKey)); isValidTaskApiPath(raw) {
		return normalizeTaskApiPath(raw)
	}
	return normalizeTaskApiPath(stringFromBody(body, fallbackKey))
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
		// 解析 usage 信息用于按倍率计费与日志展示。
		// 上游只回报 completion_tokens / total_tokens，输入 token 用差值兜底。
		taskResult.CompletionTokens = resTask.Usage.CompletionTokens
		taskResult.TotalTokens = resTask.Usage.TotalTokens
		if taskResult.TotalTokens <= 0 {
			taskResult.TotalTokens = taskResult.CompletionTokens
		}
		taskResult.PromptTokens = resTask.Usage.PromptTokens
		if taskResult.PromptTokens <= 0 && taskResult.TotalTokens > taskResult.CompletionTokens {
			taskResult.PromptTokens = taskResult.TotalTokens - taskResult.CompletionTokens
		}
		// 真实时长与分辨率档，供按秒计费在终态按实际结果补扣
		taskResult.VideoSeconds = resTask.Duration
		taskResult.VideoVariant = resTask.Resolution
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
	// 生成参数与 token 用量对齐上游终态字段，便于调用方核对实际计费口径。
	openAIVideo.Duration = dResp.Duration
	openAIVideo.Resolution = dResp.Resolution
	openAIVideo.Ratio = dResp.Ratio
	openAIVideo.FramesPerSecond = dResp.FramesPerSecond
	promptTokens, completionTokens, totalTokens := originTask.PromptTokens, originTask.CompletionTokens, 0
	if promptTokens == 0 && completionTokens == 0 {
		// 历史任务没有持久化 token 用量，回退到任务 Data 里的上游 usage。
		promptTokens = dResp.Usage.PromptTokens
		completionTokens = dResp.Usage.CompletionTokens
		totalTokens = dResp.Usage.TotalTokens
	}
	openAIVideo.SetUsage(promptTokens, completionTokens, totalTokens)

	if dResp.Status == "failed" {
		openAIVideo.Error = &dto.OpenAIVideoError{
			Message: dResp.Error.Message,
			Code:    dResp.Error.Code,
		}
	}

	return common.Marshal(openAIVideo)
}
