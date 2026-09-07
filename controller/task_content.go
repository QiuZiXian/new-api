package controller

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	taskdto "github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	taskdoubao "github.com/QuantumNous/new-api/relay/channel/task/doubao"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

// contentTaskPageMax 与上游 Seedance/CII 的分页上限保持一致。
const contentTaskPageMax = 500

// contentTaskItem 是对外内容生成任务（Seedance）列表项的对外形态：
// 字段名与上游 GET /contents/generations/tasks 返回的 task 结构对齐，
// 但 id 使用 new-api 本地的公开 task_xxx ID（创建/查询接口对外返回的就是它）。
type contentTaskItem struct {
	ID             string            `json:"id"`
	Model          string            `json:"model,omitempty"`
	Status         string            `json:"status"`
	Error          *contentTaskError `json:"error,omitempty"`
	Content        *contentTaskBody  `json:"content,omitempty"`
	Seed           *int              `json:"seed,omitempty"`
	Resolution     string            `json:"resolution,omitempty"`
	Ratio          string            `json:"ratio,omitempty"`
	Duration       *int              `json:"duration,omitempty"`
	Frames         *int              `json:"frames,omitempty"`
	FramePerSecond *int              `json:"framespersecond,omitempty"`
	ServiceTier    string            `json:"service_tier,omitempty"`
	Usage          *contentTaskUsage `json:"usage,omitempty"`
	CreatedAt      int64             `json:"created_at,omitempty"`
	UpdatedAt      int64             `json:"updated_at,omitempty"`
}

type contentTaskBody struct {
	VideoURL     string `json:"video_url,omitempty"`
	LastFrameURL string `json:"last_frame_url,omitempty"`
}

type contentTaskError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type contentTaskUsage struct {
	CompletionTokens *int `json:"completion_tokens,omitempty"`
	TotalTokens      *int `json:"total_tokens,omitempty"`
}

// doubaoContentPlatforms 是走 doubao/Seedance 内容生成任务协议的渠道类型对应的
// tasks.platform 存储值（平台列存的即渠道类型字符串）。
func doubaoContentPlatforms() []string {
	return []string{
		strconv.Itoa(constant.ChannelTypeDoubaoVideo),
		strconv.Itoa(constant.ChannelTypeVolcEngine),
	}
}

// ListContentTasks 处理 GET /v1/contents/generations/tasks。
//
// 与直接代理上游不同：上游账号是网关多租户共享的，直接透传上游列表会泄漏其它
// 用户的任务。因此这里读取调用者自己的本地 tasks 数据，并按上游风格返回。
//
// 请求参数：page_num/page_size（默认 1/20，上限 500）、filter.status、
// filter.task_ids（可重复）、filter.model。响应为 {"items":[...],"total":N}。
func ListContentTasks(c *gin.Context) {
	userId := c.GetInt("id")

	var statuses []model.TaskStatus
	if raw := strings.TrimSpace(c.Query("filter.status")); raw != "" {
		st, ok := contentFilterStatusToInternal(raw)
		if !ok {
			respondContentTaskError(c, http.StatusBadRequest, "invalid_status", "不支持的 filter.status："+raw)
			return
		}
		statuses = st
	}
	var taskIDs []string
	for _, v := range c.QueryArray("filter.task_ids") {
		if v = strings.TrimSpace(v); v != "" {
			taskIDs = append(taskIDs, v)
		}
	}
	modelName := strings.TrimSpace(c.Query("filter.model"))

	tasks, err := model.TaskGetUserVideoTasks(userId, doubaoContentPlatforms(), statuses, taskIDs)
	if err != nil {
		common.SysError("TaskGetUserVideoTasks error: " + err.Error())
		respondContentTaskError(c, http.StatusInternalServerError, "list_tasks_failed", "查询任务失败")
		return
	}

	// 模型名只存在 Properties JSON 里（跨库 JSON 查询不可靠），过滤放内存。
	if modelName != "" {
		filtered := tasks[:0]
		for _, t := range tasks {
			if strings.EqualFold(t.Properties.OriginModelName, modelName) ||
				strings.EqualFold(t.Properties.UpstreamModelName, modelName) {
				filtered = append(filtered, t)
			}
		}
		tasks = filtered
	}

	total := len(tasks)
	pageNum := contentPageParam(c.Query("page_num"), 1)
	pageSize := contentPageParam(c.Query("page_size"), 20)
	start := (pageNum - 1) * pageSize
	if start < 0 {
		start = 0
	}
	if start >= total {
		c.JSON(http.StatusOK, gin.H{"items": []contentTaskItem{}, "total": total})
		return
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	items := make([]contentTaskItem, 0, end-start)
	for _, t := range tasks[start:end] {
		items = append(items, buildContentTaskItem(t))
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

// CancelContentTask 处理 DELETE /v1/contents/generations/tasks/{task_id}。
//
// task_id 是 new-api 对外的公开 ID（task_xxx），等价于上游任务在本地平台上的标识。
// 语义为"先取消本地、再请求上游"：
//   - 仅排队/待提交（NOT_START/SUBMITTED/QUEUED）任务可取消：本地 CAS 置为终态
//     CANCELLED 并回收预扣额度，成功后调上游 DELETE 取消排队任务。
//   - 运行中（IN_PROGRESS）任务不支持（与上游规则一致；卡死任务由超时清理退款）。
//   - 已结束（SUCCESS/FAILURE/CANCELLED）任务返回 400。
func CancelContentTask(c *gin.Context) {
	taskID := strings.TrimSpace(c.Param("task_id"))
	if taskID == "" {
		respondContentTaskError(c, http.StatusBadRequest, "invalid_request", "缺少任务 ID")
		return
	}
	userId := c.GetInt("id")

	task, exist, err := model.GetByTaskId(userId, taskID)
	if err != nil {
		common.SysError("GetByTaskId error: " + err.Error())
		respondContentTaskError(c, http.StatusInternalServerError, "get_task_failed", "查询任务失败")
		return
	}
	if !exist {
		respondContentTaskError(c, http.StatusBadRequest, "task_not_exist", "任务不存在")
		return
	}

	prevStatus := task.Status
	switch prevStatus {
	case model.TaskStatusNotStart, model.TaskStatusSubmitted, model.TaskStatusQueued:
		// pending，可取消
	case model.TaskStatusInProgress:
		respondContentTaskError(c, http.StatusBadRequest, "task_running", "运行中的任务暂不支持取消")
		return
	default:
		respondContentTaskError(c, http.StatusBadRequest, "task_finished", "任务已结束，无法取消")
		return
	}

	ch, err := model.CacheGetChannel(task.ChannelId)
	if err != nil {
		respondContentTaskError(c, http.StatusBadRequest, "channel_not_found", "任务渠道不存在")
		return
	}
	if ch.Type != constant.ChannelTypeDoubaoVideo && ch.Type != constant.ChannelTypeVolcEngine {
		respondContentTaskError(c, http.StatusBadRequest, "unsupported_channel", "该任务渠道不支持取消")
		return
	}
	upstreamID := task.GetUpstreamTaskID()
	if strings.TrimSpace(upstreamID) == "" {
		respondContentTaskError(c, http.StatusBadRequest, "missing_upstream_task_id", "任务缺少上游 ID，无法取消")
		return
	}

	// 1) 本地先落终态 CANCELLED。用 CAS 防止与轮询/超时清理并发导致重复退款或覆盖终态。
	task.Status = model.TaskStatusCancelled
	task.Progress = "100%"
	task.FinishTime = time.Now().Unix()
	task.FailReason = "任务已被用户取消"
	won, err := task.UpdateWithStatus(prevStatus)
	if err != nil {
		common.SysError("cancel task CAS update error: " + err.Error())
		respondContentTaskError(c, http.StatusInternalServerError, "update_task_failed", "更新任务状态失败")
		return
	}
	if !won {
		respondContentTaskError(c, http.StatusConflict, "task_state_changed", "任务状态已变化，请刷新后重试")
		return
	}
	// CAS 成功后回收提交时的预扣额度
	service.RefundTaskQuota(context.Background(), task, task.FailReason)

	// 2) 再请求上游取消/删除排队任务。本地已取消，上游失败仅作为告警随错误返回。
	if err := (&taskdoubao.TaskAdaptor{}).DeleteTask(
		ch.GetBaseURL(),
		ch.Key,
		ch.GetOtherSettings().TaskApiPath,
		upstreamID,
		ch.GetSetting().Proxy,
	); err != nil {
		respondContentTaskError(c, http.StatusBadGateway, "upstream_delete_failed",
			"本地任务已取消，但上游删除失败："+err.Error())
		return
	}

	c.Status(http.StatusOK)
}

// buildContentTaskItem 以 task 行为权威来源构建上游形态的列表项：先用本地保存的
// 上游快照 task.Data 填充本地表不记录的字段（resolution/duration/usage 等），
// 再以本地字段覆盖标识/模型/状态/时间等权威信息。
func buildContentTaskItem(task *model.Task) contentTaskItem {
	item := contentTaskItem{}
	if len(task.Data) > 0 {
		// 快照解析失败时保留零值，仅走本地字段，不影响列表。
		_ = common.Unmarshal(task.Data, &item)
	}

	item.ID = task.TaskID
	if m := contentTaskModelName(task); m != "" {
		item.Model = m
	}
	item.Status = contentStatusToUpstream(task.Status)
	if task.CreatedAt != 0 {
		item.CreatedAt = task.CreatedAt
	}
	if task.UpdatedAt != 0 {
		item.UpdatedAt = task.UpdatedAt
	}
	if url := task.GetResultURL(); url != "" {
		if item.Content == nil {
			item.Content = &contentTaskBody{}
		}
		item.Content.VideoURL = url
	}
	if task.Status == model.TaskStatusFailure && task.FailReason != "" {
		item.Error = &contentTaskError{Code: "failed", Message: task.FailReason}
	}
	return item
}

// contentTaskModelName 返回对外展示的模型名：优先本地记录的原始（客户端侧）模型名，
// 兼容上游（映射后）模型名，两者都没有时返回空交由上游快照兜底。
func contentTaskModelName(task *model.Task) string {
	if n := strings.TrimSpace(task.Properties.OriginModelName); n != "" {
		return n
	}
	return strings.TrimSpace(task.Properties.UpstreamModelName)
}

// contentStatusToUpstream 把本地内部状态映射为上游形态的状态字符串。
func contentStatusToUpstream(status model.TaskStatus) string {
	switch status {
	case model.TaskStatusNotStart, model.TaskStatusSubmitted, model.TaskStatusQueued:
		return "queued"
	case model.TaskStatusInProgress:
		return "running"
	case model.TaskStatusSuccess:
		return "succeeded"
	case model.TaskStatusFailure:
		return "failed"
	case model.TaskStatusCancelled:
		return "cancelled"
	default:
		return "queued"
	}
}

// contentFilterStatusToInternal 把上游 filter.status 映射到内部状态集合。
// 第二个返回值标识入参是否合法：expired 在本地没有对应状态（超时任务已转 FAILURE），
// 视为合法但返回空集合（结果为空）。
func contentFilterStatusToInternal(s string) ([]model.TaskStatus, bool) {
	switch strings.TrimSpace(s) {
	case "queued":
		return []model.TaskStatus{model.TaskStatusNotStart, model.TaskStatusSubmitted, model.TaskStatusQueued}, true
	case "running":
		return []model.TaskStatus{model.TaskStatusInProgress}, true
	case "cancelled":
		return []model.TaskStatus{model.TaskStatusCancelled}, true
	case "succeeded":
		return []model.TaskStatus{model.TaskStatusSuccess}, true
	case "failed":
		return []model.TaskStatus{model.TaskStatusFailure}, true
	case "expired":
		return []model.TaskStatus{}, true
	default:
		return nil, false
	}
}

func contentPageParam(raw string, def int) int {
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return def
	}
	if v < 1 {
		return def
	}
	if v > contentTaskPageMax {
		return contentTaskPageMax
	}
	return v
}

func respondContentTaskError(c *gin.Context, status int, code, message string) {
	c.JSON(status, &taskdto.TaskError{Code: code, Message: message, StatusCode: status})
}
