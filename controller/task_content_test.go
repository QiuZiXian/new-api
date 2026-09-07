package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentStatusToUpstream(t *testing.T) {
	cases := []struct {
		name   string
		status model.TaskStatus
		want   string
	}{
		{"not_start maps to queued", model.TaskStatusNotStart, "queued"},
		{"submitted maps to queued", model.TaskStatusSubmitted, "queued"},
		{"queued maps to queued", model.TaskStatusQueued, "queued"},
		{"in_progress maps to running", model.TaskStatusInProgress, "running"},
		{"success maps to succeeded", model.TaskStatusSuccess, "succeeded"},
		{"failure maps to failed", model.TaskStatusFailure, "failed"},
		{"cancelled maps to cancelled", model.TaskStatusCancelled, "cancelled"},
		{"unknown falls back to queued", model.TaskStatusUnknown, "queued"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, contentStatusToUpstream(tc.status))
		})
	}
}

func TestContentFilterStatusToInternal(t *testing.T) {
	cases := []struct {
		name   string
		filter string
		want   []model.TaskStatus
		ok     bool
	}{
		{"queued expands to pending set", "queued",
			[]model.TaskStatus{model.TaskStatusNotStart, model.TaskStatusSubmitted, model.TaskStatusQueued}, true},
		{"running maps to in_progress", "running",
			[]model.TaskStatus{model.TaskStatusInProgress}, true},
		{"cancelled maps to cancelled", "cancelled",
			[]model.TaskStatus{model.TaskStatusCancelled}, true},
		{"succeeded maps to success", "succeeded",
			[]model.TaskStatus{model.TaskStatusSuccess}, true},
		{"failed maps to failure", "failed",
			[]model.TaskStatus{model.TaskStatusFailure}, true},
		{"expired is valid but empty locally", "expired",
			[]model.TaskStatus{}, true},
		{"unknown filter is invalid", "mystery", nil, false},
		{"whitespace is invalid", "  ", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := contentFilterStatusToInternal(tc.filter)
			assert.Equal(t, tc.ok, ok)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestContentPageParam(t *testing.T) {
	assert.Equal(t, 1, contentPageParam("", 1))
	assert.Equal(t, 20, contentPageParam("", 20))
	assert.Equal(t, 3, contentPageParam("3", 1))
	assert.Equal(t, 500, contentPageParam("9000", 1))
	assert.Equal(t, 1, contentPageParam("-2", 1))
	assert.Equal(t, 20, contentPageParam("abc", 20))
}

func TestBuildContentTaskItem(t *testing.T) {
	t.Run("local fields override upstream snapshot", func(t *testing.T) {
		task := &model.Task{
			TaskID:     "task_local_1",
			Status:     model.TaskStatusSuccess,
			CreatedAt:  100,
			UpdatedAt:  200,
			Properties: model.Properties{OriginModelName: "doubao-seedance-2-0-260128"},
		}
		task.PrivateData.ResultURL = "https://cdn.example.com/out.mp4"
		// Upstream snapshot carries an id/model/url that must not win over the
		// local row's authoritative fields.
		task.SetData(map[string]any{
			"id":       "ct-4fc3d0ce9c084e4ab",
			"model":    "ep-20250403102709-h6c8n",
			"status":   "succeeded",
			"duration": 5,
			"content":  map[string]any{"video_url": "https://upstream.example.com/v.mp4"},
		})

		item := buildContentTaskItem(task)
		assert.Equal(t, "task_local_1", item.ID, "id must be the local public task id")
		assert.Equal(t, "doubao-seedance-2-0-260128", item.Model, "model must come from the local origin model")
		assert.Equal(t, "succeeded", item.Status)
		assert.Equal(t, int64(100), item.CreatedAt)
		assert.Equal(t, int64(200), item.UpdatedAt)
		require.NotNil(t, item.Content)
		assert.Equal(t, "https://cdn.example.com/out.mp4", item.Content.VideoURL, "result url must come from the local row")
		require.NotNil(t, item.Duration, "upstream snapshot fields not tracked locally are enriched")
		assert.Equal(t, 5, *item.Duration)
	})

	t.Run("failed task surfaces local reason", func(t *testing.T) {
		task := &model.Task{
			TaskID:     "task_local_2",
			Status:     model.TaskStatusFailure,
			FailReason: "content rejected",
			Properties: model.Properties{UpstreamModelName: "doubao-seedance-1-0-lite-t2v"},
		}
		task.SetData(map[string]any{
			"status": "failed",
			"error":  map[string]any{"code": "BadRequest", "message": "upstream said no"},
		})

		item := buildContentTaskItem(task)
		assert.Equal(t, "failed", item.Status)
		require.NotNil(t, item.Error)
		assert.Equal(t, "failed", item.Error.Code)
		assert.Equal(t, "content rejected", item.Error.Message, "local fail reason is authoritative")
		assert.Equal(t, "doubao-seedance-1-0-lite-t2v", item.Model, "upstream model name is used when origin is empty")
	})

	t.Run("cancelled task maps to cancelled", func(t *testing.T) {
		task := &model.Task{TaskID: "task_local_3", Status: model.TaskStatusCancelled}
		item := buildContentTaskItem(task)
		assert.Equal(t, "cancelled", item.Status)
	})

	t.Run("bad snapshot is ignored", func(t *testing.T) {
		task := &model.Task{TaskID: "task_local_4", Status: model.TaskStatusQueued}
		task.Data = []byte("{not json")
		item := buildContentTaskItem(task)
		assert.Equal(t, "task_local_4", item.ID)
		assert.Equal(t, "queued", item.Status)
	})
}
