package dto

import (
	"strconv"
	"strings"
)

const (
	VideoStatusUnknown    = "unknown"
	VideoStatusQueued     = "queued"
	VideoStatusInProgress = "in_progress"
	VideoStatusCompleted  = "completed"
	VideoStatusFailed     = "failed"
)

// OpenAIVideo 视频任务的标准响应体。
// Duration / Resolution / Ratio / FramesPerSecond / Usage 与上游（如火山方舟
// Seedance）的终态字段对齐，便于调用方核对实际生成参数与 token 用量；
// 提交阶段未知，任务终态查询时才会填充。
type OpenAIVideo struct {
	ID                 string            `json:"id"`
	TaskID             string            `json:"task_id,omitempty"` //兼容旧接口 待废弃
	Object             string            `json:"object"`
	Model              string            `json:"model"`
	Status             string            `json:"status"` // Should use VideoStatus constants: VideoStatusQueued, VideoStatusInProgress, VideoStatusCompleted, VideoStatusFailed
	Progress           int               `json:"progress"`
	CreatedAt          int64             `json:"created_at"`
	CompletedAt        int64             `json:"completed_at,omitempty"`
	ExpiresAt          int64             `json:"expires_at,omitempty"`
	Seconds            string            `json:"seconds,omitempty"`
	Size               string            `json:"size,omitempty"`
	Duration           int               `json:"duration,omitempty"`
	Resolution         string            `json:"resolution,omitempty"`
	Ratio              string            `json:"ratio,omitempty"`
	FramesPerSecond    int               `json:"framespersecond,omitempty"`
	Usage              *OpenAIVideoUsage `json:"usage,omitempty"`
	RemixedFromVideoID string            `json:"remixed_from_video_id,omitempty"`
	Error              *OpenAIVideoError `json:"error,omitempty"`
	Metadata           map[string]any    `json:"metadata,omitempty"`
}

// OpenAIVideoUsage 视频任务的 token 用量。上游（如 Seedance）只回报
// completion_tokens / total_tokens，prompt_tokens 缺省时为 0。
type OpenAIVideoUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// SetUsage 写入 token 用量；全为 0 时不写，避免给未上报用量的渠道塞空对象。
func (m *OpenAIVideo) SetUsage(promptTokens int, completionTokens int, totalTokens int) {
	if promptTokens <= 0 && completionTokens <= 0 && totalTokens <= 0 {
		return
	}
	if totalTokens <= 0 {
		totalTokens = promptTokens + completionTokens
	}
	m.Usage = &OpenAIVideoUsage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
	}
}

func (m *OpenAIVideo) SetProgressStr(progress string) {
	progress = strings.TrimSuffix(progress, "%")
	m.Progress, _ = strconv.Atoi(progress)
}
func (m *OpenAIVideo) SetMetadata(k string, v any) {
	if m.Metadata == nil {
		m.Metadata = make(map[string]any)
	}
	m.Metadata[k] = v
}
func NewOpenAIVideo() *OpenAIVideo {
	return &OpenAIVideo{
		Object: "video",
		Status: VideoStatusQueued,
	}
}

type OpenAIVideoError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}
