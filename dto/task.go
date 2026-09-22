package dto

import (
	"encoding/json"

	"github.com/QuantumNous/new-api/common"
)

type TaskError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Data       any    `json:"data"`
	StatusCode int    `json:"-"`
	LocalError bool   `json:"-"`
	Error      error  `json:"-"`
}

// taskErrorEnvelope 是 TaskError 序列化后的统一响应形态：
// 顶层保留扁平字段（历史形态），并列的 error 包裹按 OpenAI 约定提供
// code/message/type。客户端用任一种姿势解析都可以。
type taskErrorEnvelope struct {
	Code       string           `json:"code"`
	Message    string           `json:"message"`
	Data       any              `json:"data,omitempty"`
	StatusCode int              `json:"status_code,omitempty"`
	Error      *taskErrorDetail `json:"error,omitempty"`
}

type taskErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// MarshalJSON 让所有把 TaskError 直接 c.JSON 出去的接口自动获得统一错误体，
// 不必逐个改调用点。StatusCode 未设置时不输出 status_code，避免误报 0。
// LocalError / Error 仅供内部流转，不出现在响应里。
func (t *TaskError) MarshalJSON() ([]byte, error) {
	if t == nil {
		return []byte("null"), nil
	}
	return common.Marshal(taskErrorEnvelope{
		Code:       t.Code,
		Message:    t.Message,
		Data:       t.Data,
		StatusCode: t.StatusCode,
		Error: &taskErrorDetail{
			Code:    t.Code,
			Message: t.Message,
			Type:    common.AppErrorType(),
		},
	})
}

type TaskData interface {
	SunoDataResponse | []SunoDataResponse | string | any
}

const TaskSuccessCode = "success"

type TaskResponse[T TaskData] struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func (t *TaskResponse[T]) IsSuccess() bool {
	return t.Code == TaskSuccessCode
}

type TaskDto struct {
	ID         int64           `json:"id"`
	CreatedAt  int64           `json:"created_at"`
	UpdatedAt  int64           `json:"updated_at"`
	TaskID     string          `json:"task_id"`
	Platform   string          `json:"platform"`
	UserId     int             `json:"user_id"`
	Group      string          `json:"group"`
	ChannelId  int             `json:"channel_id"`
	Quota      int             `json:"quota"`
	Action     string          `json:"action"`
	Status     string          `json:"status"`
	FailReason string          `json:"fail_reason"`
	ResultURL  string          `json:"result_url,omitempty"` // 任务结果 URL（视频地址等）
	SubmitTime int64           `json:"submit_time"`
	StartTime  int64           `json:"start_time"`
	FinishTime int64           `json:"finish_time"`
	Progress   string          `json:"progress"`
	Properties any             `json:"properties"`
	Username   string          `json:"username,omitempty"`
	Data       json.RawMessage `json:"data"`
}

type FetchReq struct {
	IDs []string `json:"ids"`
}
