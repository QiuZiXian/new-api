package ciisubtitleerase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"

	"github.com/QuantumNous/new-api/constant"
	taskdto "github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay/channel"
	taskcommon "github.com/QuantumNous/new-api/relay/channel/task/taskcommon"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// ============================
// Request / Response structures
// ============================

// submitRequest 是 CII 字幕擦除提交请求的载荷：仅一个 video_url 字段。
// model 字段不由 CII 服务消费，仅用作平台层的路由键。
type submitRequest struct {
	VideoURL string `json:"video_url"`
}

// submitResponse 是 CII 提交接口的成功响应。失败响应会带 success=false 或非 2xx。
type submitResponse struct {
	Success bool   `json:"success"`
	TaskID  string `json:"task_id"`
}

// pollResponse 是 CII 轮询接口的响应。CII 的 success 字段是 API 调用层结果，
// 任务最终状态以 status 字段为准。
type pollResponse struct {
	Status     string `json:"status"`
	Result     struct {
		VideoURL string `json:"video_url"`
	} `json:"result"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	CreatedAt  int64 `json:"created_at"`
	FinishedAt int64 `json:"finished_at"`
	ExpiresAt  int64 `json:"expires_at"`
}

// ============================
// Adaptor implementation
// ============================

// TaskAdaptor 实现 channel.TaskAdaptor：把 CII 字幕擦除的"创建任务 / 轮询"映射
// 到平台统一的任务接口。计费在 AdjustBillingOnComplete 里按 finished_at -
// created_at 的实际秒数 × 1 积分结算（不预扣）。
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

func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) *taskdto.TaskError {
	return relaycommon.ValidateBasicTaskRequest(c, info, constant.TaskActionGenerate)
}

func (a *TaskAdaptor) BuildRequestURL(info *relaycommon.RelayInfo) (string, error) {
	return a.baseURL + a.taskApiPath(info), nil
}

func (a *TaskAdaptor) BuildRequestHeader(_ *gin.Context, req *http.Request, _ *relaycommon.RelayInfo) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	return nil
}

// BuildRequestBody 从 req.Metadata["video_url"] 读取视频 URL，序列化为
// {"video_url": "..."} 发往 CII。video_url 缺失或空字符串直接 400。
func (a *TaskAdaptor) BuildRequestBody(c *gin.Context, info *relaycommon.RelayInfo) (io.Reader, error) {
	req, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return nil, err
	}

	videoURL, err := extractVideoURL(req.Metadata)
	if err != nil {
		return nil, err
	}

	body := submitRequest{VideoURL: videoURL}
	data, err := common.Marshal(body)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

// extractVideoURL 兼容两种 metadata 形态：
//  1. {"video_url": "https://..."}（标准形态）
//  2. {"video_url": {"url": "https://..."}}（嵌套形态，防止用户写错）
//
// 缺失 / 空字符串 / 非字符串 → 报错（400）。
func extractVideoURL(metadata map[string]interface{}) (string, error) {
	if metadata == nil {
		return "", errors.New("metadata.video_url is required")
	}
	raw, ok := metadata["video_url"]
	if !ok || raw == nil {
		return "", errors.New("metadata.video_url is required")
	}
	switch v := raw.(type) {
	case string:
		v = strings.TrimSpace(v)
		if v == "" {
			return "", errors.New("metadata.video_url must be a non-empty string")
		}
		return v, nil
	case map[string]interface{}:
		if inner, ok := v["url"].(string); ok {
			inner = strings.TrimSpace(inner)
			if inner == "" {
				return "", errors.New("metadata.video_url.url must be a non-empty string")
			}
			return inner, nil
		}
		return "", errors.New("metadata.video_url.url must be a non-empty string")
	default:
		return "", errors.New("metadata.video_url must be a string")
	}
}

func (a *TaskAdaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	return channel.DoTaskApiRequest(a, c, info, requestBody)
}

// DoResponse 解析 CII 提交响应，把 CII 的 task_id 记为上游任务 ID 并返回
// OpenAIVideo 形态给客户端。success=false 或 task_id 缺失时包装为 TaskError。
func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (string, []byte, *taskdto.TaskError) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, service.TaskErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
	}
	_ = resp.Body.Close()

	var sResp submitResponse
	if err := common.Unmarshal(responseBody, &sResp); err != nil {
		return "", nil, service.TaskErrorWrapper(errors.Wrapf(err, "body: %s", responseBody), "unmarshal_response_body_failed", http.StatusInternalServerError)
	}

	if !sResp.Success {
		// 把上游 body 摘要拼到错误里，便于排查路径 / 鉴权 / 配额等问题。
		return "", nil, service.TaskErrorWrapper(
			fmt.Errorf("cii submit failed: %s", snippetBody(responseBody)),
			"invalid_response",
			http.StatusInternalServerError,
		)
	}

	if sResp.TaskID == "" {
		return "", nil, service.TaskErrorWrapper(
			fmt.Errorf("task_id is empty; upstream response: %s", snippetBody(responseBody)),
			"invalid_response",
			http.StatusInternalServerError,
		)
	}

	ov := dto.NewOpenAIVideo()
	ov.ID = info.PublicTaskID
	ov.TaskID = info.PublicTaskID
	ov.CreatedAt = time.Now().Unix()
	ov.Model = info.OriginModelName
	c.JSON(http.StatusOK, ov)
	return sResp.TaskID, responseBody, nil
}

// FetchTask 用 task_id 拼出 GET {baseURL + path}/{task_id} 拉取状态。
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
	req.Header.Set("Authorization", "Bearer "+key)

	client, err := service.GetHttpClientWithProxy(proxy)
	if err != nil {
		return nil, fmt.Errorf("new proxy http client failed: %w", err)
	}
	return client.Do(req)
}

func (a *TaskAdaptor) GetModelList() []string {
	return ModelList
}

func (a *TaskAdaptor) GetChannelName() string {
	return ChannelName
}

// taskApiPath 返回配置的 upstream path 后缀（来自渠道 OtherSettings.TaskApiPath），
// 缺省/非法时回退 defaultSubtitleEraseApiPath。
func (a *TaskAdaptor) taskApiPath(info *relaycommon.RelayInfo) string {
	var raw string
	if info != nil && info.HasChannelMeta() {
		raw = info.ChannelOtherSettings.TaskApiPath
	}
	return normalizeTaskApiPath(raw)
}

// normalizeTaskApiPath 限制 path 后缀是干净的相对 URL 路径：
//   - 必须是 "/" 开头
//   - 长度不超过 256 字节
//   - 只允许字母 / 数字 / "-_.~/"
//   - 不允许 ".." 段
//
// 任何不合规的值都回退到 defaultSubtitleEraseApiPath，
// 防止 TaskApiPath 被配置成攻击向量（路径注入）。
func normalizeTaskApiPath(raw string) string {
	p := strings.TrimSpace(raw)
	if p == "" || len(p) > 256 {
		return defaultSubtitleEraseApiPath
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
			return defaultSubtitleEraseApiPath
		}
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." {
			return defaultSubtitleEraseApiPath
		}
	}
	return p
}

func stringFromBody(body map[string]any, key string) string {
	if body == nil {
		return ""
	}
	if v, ok := body[key].(string); ok {
		return v
	}
	return ""
}

// snippetBody 截取上游 body 的前 500 字节，便于在错误信息里展示但不暴露过多内容。
func snippetBody(body []byte) string {
	s := strings.TrimSpace(string(body))
	if len(s) > 500 {
		return s[:500] + " ...(truncated)"
	}
	if s == "" {
		return "(empty body)"
	}
	return s
}

// ParseTaskResult 映射 CII status 字符串到平台 TaskStatus。
// CII 的 success 是 API 调用层结果，与任务最终状态无关——必须以 status 为准。
func (a *TaskAdaptor) ParseTaskResult(respBody []byte) (*relaycommon.TaskInfo, error) {
	resTask := pollResponse{}
	if err := common.Unmarshal(respBody, &resTask); err != nil {
		return nil, errors.Wrap(err, "unmarshal task result failed")
	}

	taskResult := relaycommon.TaskInfo{Code: 0}

	switch resTask.Status {
	case "running":
		taskResult.Status = model.TaskStatusInProgress
		taskResult.Progress = "50%"
	case "completed":
		taskResult.Status = model.TaskStatusSuccess
		taskResult.Progress = "100%"
		taskResult.Url = resTask.Result.VideoURL
	case "failed":
		taskResult.Status = model.TaskStatusFailure
		taskResult.Progress = "100%"
		taskResult.Reason = resTask.Error.Message
		if taskResult.Reason == "" {
			taskResult.Reason = resTask.Error.Code
		}
	default:
		// 未知状态：保守地按"进行中"处理，避免误判为失败。
		taskResult.Status = model.TaskStatusInProgress
		taskResult.Progress = "30%"
	}

	return &taskResult, nil
}

// ConvertToOpenAIVideo 把本地 Task 翻成 OpenAIVideo。状态由
// originTask.Status.ToVideoStatus() 决定，metadata.url 来自 result.video_url。
func (a *TaskAdaptor) ConvertToOpenAIVideo(originTask *model.Task) ([]byte, error) {
	var pResp pollResponse
	if err := common.Unmarshal(originTask.Data, &pResp); err != nil {
		return nil, errors.Wrap(err, "unmarshal cii subtitle-erase task data failed")
	}

	openAIVideo := dto.NewOpenAIVideo()
	openAIVideo.ID = originTask.TaskID
	openAIVideo.TaskID = originTask.TaskID
	openAIVideo.Status = originTask.Status.ToVideoStatus()
	openAIVideo.SetProgressStr(originTask.Progress)
	openAIVideo.SetMetadata("url", pResp.Result.VideoURL)
	openAIVideo.CreatedAt = originTask.CreatedAt
	openAIVideo.CompletedAt = pResp.FinishedAt
	openAIVideo.ExpiresAt = pResp.ExpiresAt
	openAIVideo.Model = originTask.Properties.OriginModelName

	if pResp.Status == "failed" {
		openAIVideo.Error = &dto.OpenAIVideoError{
			Message: pResp.Error.Message,
			Code:    pResp.Error.Code,
		}
	}

	return common.Marshal(openAIVideo)
}

// ============================
// Billing
// ============================

// EstimateBilling 不预扣。subtitle-erase 渠道的费率基于 finished_at - created_at
// 的实际秒数，提交时无法预知。返回 nil 让调用方按基础 ModelRatio 计算 0 扣费。
func (a *TaskAdaptor) EstimateBilling(_ *gin.Context, _ *relaycommon.RelayInfo) map[string]float64 {
	return nil
}

// AdjustBillingOnSubmit 提交阶段不需要调整（无预扣）。
func (a *TaskAdaptor) AdjustBillingOnSubmit(_ *relaycommon.RelayInfo, _ []byte) map[string]float64 {
	return nil
}

// AdjustBillingOnComplete 在终态被轮询循环调用：
//  1. 解析 task.Data 取 created_at / finished_at；
//  2. actualSeconds = finished_at - created_at；
//  3. 异常值（< 0 或 > 4 * MaxTaskDurationSeconds）写 task.FailReason 并返回 0；
//  4. 合法范围返回 common.QuotaFromFloat(actualSeconds * 1)。
//
// actualSeconds == 0 也算合法（返回 0 即可）。
func (a *TaskAdaptor) AdjustBillingOnComplete(task *model.Task, taskResult *relaycommon.TaskInfo) int {
	if task == nil || taskResult == nil {
		return 0
	}

	var pResp pollResponse
	if err := common.Unmarshal(task.Data, &pResp); err != nil {
		// 解析失败：记日志、记 FailReason，让轮询循环走失败分支。
		a.markDurationInvalid(task, "cannot parse upstream poll response: "+err.Error())
		return 0
	}

	// finished_at 缺失的极端情况下：按 0 秒处理（不扣费）。
	actualSeconds := pResp.FinishedAt - pResp.CreatedAt

	maxSeconds := int64(4 * relaycommon.MaxTaskDurationSeconds) // 4 * 3600 = 14400
	if actualSeconds < 0 {
		a.markDurationInvalid(task, fmt.Sprintf("negative: %d", actualSeconds))
		return 0
	}
	if actualSeconds > maxSeconds {
		a.markDurationInvalid(task, fmt.Sprintf("exceeded %d seconds", maxSeconds))
		return 0
	}

	return common.QuotaFromFloat(float64(actualSeconds) * quotaPerSecond)
}

// markDurationInvalid 记录失败原因（供轮询循环走失败分支）+ 日志。
// 不直接修改 task.Status，避免与轮询主流程的状态机冲突；
// 由 task_polling 读取 FailReason 后再决定是否置失败。
func (a *TaskAdaptor) markDurationInvalid(task *model.Task, detail string) {
	if task == nil {
		return
	}
	task.FailReason = "subtitle erase duration invalid: " + detail
	common.SysError(fmt.Sprintf("cii subtitle-erase billing rejected: task_id=%s reason=%s", task.TaskID, task.FailReason))
	logger.LogWarn(context.Background(), fmt.Sprintf("cii subtitle-erase billing rejected: task_id=%s reason=%s", task.TaskID, task.FailReason))
}
