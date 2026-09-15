package ciisubtitleerase

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeTaskApiPath(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty falls back to default", "", defaultSubtitleEraseApiPath},
		{"whitespace falls back to default", "   ", defaultSubtitleEraseApiPath},
		{"default path passes through", defaultSubtitleEraseApiPath, defaultSubtitleEraseApiPath},
		{"missing leading slash gets one prepended", "api/v1/videos/subtitle-erase/tasks", defaultSubtitleEraseApiPath},
		{"trims surrounding whitespace", "  " + defaultSubtitleEraseApiPath + "  ", defaultSubtitleEraseApiPath},
		{"dot segment falls back to default", "/api/../v1/videos/subtitle-erase/tasks", defaultSubtitleEraseApiPath},
		{"leading dot segment falls back to default", "../etc/passwd", defaultSubtitleEraseApiPath},
		{"backslash falls back to default", "/api\\v1\\tasks", defaultSubtitleEraseApiPath},
		{"query string falls back to default", defaultSubtitleEraseApiPath + "?page=1", defaultSubtitleEraseApiPath},
		{"fragment falls back to default", defaultSubtitleEraseApiPath + "#frag", defaultSubtitleEraseApiPath},
		{"scheme falls back to default", "https://evil.example.com" + defaultSubtitleEraseApiPath, defaultSubtitleEraseApiPath},
		{"space inside falls back to default", "/api/ v1/tasks", defaultSubtitleEraseApiPath},
		{"tab inside falls back to default", "/api/v1\ttasks", defaultSubtitleEraseApiPath},
		{"too long falls back to default", "/" + strings.Repeat("a", 257), defaultSubtitleEraseApiPath},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, normalizeTaskApiPath(tc.in))
		})
	}
}

func TestTaskApiPath_FallsBackWithoutChannelMeta(t *testing.T) {
	a := &TaskAdaptor{}
	assert.Equal(t, defaultSubtitleEraseApiPath, a.taskApiPath(nil))
}

func TestExtractVideoURL(t *testing.T) {
	cases := []struct {
		name    string
		meta    map[string]interface{}
		wantURL string
		wantErr bool
	}{
		{
			name:    "valid string url",
			meta:    map[string]interface{}{"video_url": "https://example.com/source.mp4"},
			wantURL: "https://example.com/source.mp4",
		},
		{
			name:    "valid nested url",
			meta:    map[string]interface{}{"video_url": map[string]interface{}{"url": "https://example.com/source.mp4"}},
			wantURL: "https://example.com/source.mp4",
		},
		{
			name:    "trims whitespace",
			meta:    map[string]interface{}{"video_url": "  https://example.com/x.mp4  "},
			wantURL: "https://example.com/x.mp4",
		},
		{
			name:    "empty string is rejected",
			meta:    map[string]interface{}{"video_url": "   "},
			wantErr: true,
		},
		{
			name:    "missing key is rejected",
			meta:    map[string]interface{}{"other": "x"},
			wantErr: true,
		},
		{
			name:    "nil metadata is rejected",
			meta:    nil,
			wantErr: true,
		},
		{
			name:    "wrong type is rejected",
			meta:    map[string]interface{}{"video_url": 123},
			wantErr: true,
		},
		{
			name:    "nested map missing url is rejected",
			meta:    map[string]interface{}{"video_url": map[string]interface{}{"foo": "bar"}},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := extractVideoURL(tc.meta)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantURL, got)
		})
	}
}

func TestParseTaskResult_StatusMapping(t *testing.T) {
	a := &TaskAdaptor{}
	cases := []struct {
		name         string
		body         string
		wantStatus   string
		wantProgress string
		wantURL      string
		wantReason   string
	}{
		{
			name:         "running maps to in_progress 50%",
			body:         `{"status":"running"}`,
			wantStatus:   string(model.TaskStatusInProgress),
			wantProgress: "50%",
		},
		{
			name:         "completed maps to success 100% with url",
			body:         `{"status":"completed","result":{"video_url":"https://example.com/result.mp4"}}`,
			wantStatus:   string(model.TaskStatusSuccess),
			wantProgress: "100%",
			wantURL:      "https://example.com/result.mp4",
		},
		{
			name:         "failed maps to failure 100% with error message",
			body:         `{"status":"failed","error":{"code":"E001","message":"bad source"}}`,
			wantStatus:   string(model.TaskStatusFailure),
			wantProgress: "100%",
			wantReason:   "bad source",
		},
		{
			name:         "failed with empty message falls back to code",
			body:         `{"status":"failed","error":{"code":"E001","message":""}}`,
			wantStatus:   string(model.TaskStatusFailure),
			wantProgress: "100%",
			wantReason:   "E001",
		},
		{
			name:         "unknown status falls back to in_progress 30%",
			body:         `{"status":"queued-bizarre-state"}`,
			wantStatus:   string(model.TaskStatusInProgress),
			wantProgress: "30%",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := a.ParseTaskResult([]byte(tc.body))
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tc.wantStatus, got.Status)
			assert.Equal(t, tc.wantProgress, got.Progress)
			assert.Equal(t, tc.wantURL, got.Url)
			assert.Equal(t, tc.wantReason, got.Reason)
		})
	}
}

func TestAdjustBillingOnComplete(t *testing.T) {
	a := &TaskAdaptor{}
	maxSeconds := int64(4 * relaycommon.MaxTaskDurationSeconds) // 14400

	cases := []struct {
		name           string
		createdAt      int64
		finishedAt     int64
		wantQuota      int
		wantFailReason string
	}{
		{
			name:       "normal range (10 seconds)",
			createdAt:  100, finishedAt: 110,
			wantQuota: 10,
		},
		{
			name:       "exactly 14400s is allowed (boundary)",
			createdAt:  100, finishedAt: 100 + maxSeconds,
			wantQuota: int(maxSeconds),
		},
		{
			name:       "zero seconds (no charge)",
			createdAt:  100, finishedAt: 100,
			wantQuota: 0,
		},
		{
			name:       "one second",
			createdAt:  100, finishedAt: 101,
			wantQuota: 1,
		},
		{
			name:           "negative duration is rejected",
			createdAt:      100, finishedAt: 50,
			wantQuota:      0,
			wantFailReason: "subtitle erase duration invalid: negative: -50",
		},
		{
			name:           "exceeds 14400s is rejected",
			createdAt:      100, finishedAt: 100 + maxSeconds + 1,
			wantQuota:      0,
			wantFailReason: "subtitle erase duration invalid: exceeded 14400 seconds",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body, err := common.Marshal(pollResponse{
				CreatedAt:  tc.createdAt,
				FinishedAt: tc.finishedAt,
			})
			require.NoError(t, err)
			task := &model.Task{
				TaskID: "se-test",
				Data:   body,
			}
			taskResult := &relaycommon.TaskInfo{}
			got := a.AdjustBillingOnComplete(task, taskResult)
			assert.Equal(t, tc.wantQuota, got)
			if tc.wantFailReason != "" {
				assert.Equal(t, tc.wantFailReason, task.FailReason)
			} else {
				assert.Empty(t, task.FailReason)
			}
		})
	}
}

func TestAdjustBillingOnComplete_ParseErrorMarksInvalid(t *testing.T) {
	a := &TaskAdaptor{}
	task := &model.Task{
		TaskID: "se-parse-err",
		Data:   []byte(`{not json`),
	}
	taskResult := &relaycommon.TaskInfo{}
	got := a.AdjustBillingOnComplete(task, taskResult)
	assert.Equal(t, 0, got)
	assert.Contains(t, task.FailReason, "subtitle erase duration invalid: cannot parse upstream poll response")
}

func TestAdjustBillingOnComplete_NilArgsSafe(t *testing.T) {
	a := &TaskAdaptor{}
	assert.Equal(t, 0, a.AdjustBillingOnComplete(nil, &relaycommon.TaskInfo{}))
	assert.Equal(t, 0, a.AdjustBillingOnComplete(&model.Task{}, nil))
}

func TestConvertToOpenAIVideo_Success(t *testing.T) {
	a := &TaskAdaptor{}
	task := &model.Task{
		TaskID:    "task_abc",
		Status:    model.TaskStatusSuccess,
		Progress:  "100%",
		CreatedAt: 1000,
		Properties: model.Properties{
			OriginModelName: "cii-subtitle-erase",
		},
		Data: []byte(`{"status":"completed","result":{"video_url":"https://example.com/out.mp4"},"finished_at":1100,"expires_at":2000}`),
	}
	out, err := a.ConvertToOpenAIVideo(task)
	require.NoError(t, err)

	parsed := map[string]interface{}{}
	require.NoError(t, common.Unmarshal(out, &parsed))

	assert.Equal(t, "task_abc", parsed["id"])
	assert.Equal(t, "task_abc", parsed["task_id"])
	assert.Equal(t, "completed", parsed["status"])
	assert.EqualValues(t, 100, parsed["progress"])
	assert.EqualValues(t, 1100, parsed["completed_at"])
	assert.EqualValues(t, 2000, parsed["expires_at"])
	assert.Equal(t, "cii-subtitle-erase", parsed["model"])
	meta, ok := parsed["metadata"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "https://example.com/out.mp4", meta["url"])
}

func TestConvertToOpenAIVideo_Failed(t *testing.T) {
	a := &TaskAdaptor{}
	task := &model.Task{
		TaskID:   "task_fail",
		Status:   model.TaskStatusFailure,
		Progress: "100%",
		Data:     []byte(`{"status":"failed","error":{"code":"E001","message":"bad video"}}`),
	}
	out, err := a.ConvertToOpenAIVideo(task)
	require.NoError(t, err)

	parsed := map[string]interface{}{}
	require.NoError(t, common.Unmarshal(out, &parsed))

	assert.Equal(t, "failed", parsed["status"])
	errObj, ok := parsed["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "bad video", errObj["message"])
	assert.Equal(t, "E001", errObj["code"])
}

// 下列测试覆盖 DoResponse 的三种分支（success=true+task_id / success=true+空 / success=false）。
// 通过 http.Response + 自实现的 io.ReadCloser 构造假的响应体，绕开真实网络。

type nopCloser struct{ io.Reader }

func (nopCloser) Close() error { return nil }

func newReadCloser(s string) io.ReadCloser {
	return nopCloser{Reader: strings.NewReader(s)}
}

func newTestGinContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos/subtitle-erase/tasks", nil)
	return c
}

func TestDoResponse_EmptyTaskID(t *testing.T) {
	a := &TaskAdaptor{}
	c := newTestGinContext()
	info := &relaycommon.RelayInfo{
		OriginModelName: "cii-subtitle-erase",
		TaskRelayInfo:   &relaycommon.TaskRelayInfo{PublicTaskID: "task_xyz"},
	}

	resp := &http.Response{
		StatusCode: 200,
		Body:       newReadCloser(`{"success":true,"task_id":""}`),
	}
	taskID, _, taskErr := a.DoResponse(c, resp, info)
	assert.Empty(t, taskID)
	require.NotNil(t, taskErr)
	assert.Equal(t, "invalid_response", taskErr.Code)
}

func TestDoResponse_SuccessFalse(t *testing.T) {
	a := &TaskAdaptor{}
	c := newTestGinContext()
	info := &relaycommon.RelayInfo{
		OriginModelName: "cii-subtitle-erase",
		TaskRelayInfo:   &relaycommon.TaskRelayInfo{PublicTaskID: "task_xyz"},
	}

	resp := &http.Response{
		StatusCode: 200,
		Body:       newReadCloser(`{"success":false,"task_id":"se-1"}`),
	}
	taskID, _, taskErr := a.DoResponse(c, resp, info)
	assert.Empty(t, taskID)
	require.NotNil(t, taskErr)
	assert.Equal(t, "invalid_response", taskErr.Code)
}

func TestDoResponse_OK(t *testing.T) {
	a := &TaskAdaptor{}
	c := newTestGinContext()
	info := &relaycommon.RelayInfo{
		OriginModelName: "cii-subtitle-erase",
		TaskRelayInfo:   &relaycommon.TaskRelayInfo{PublicTaskID: "task_pub"},
	}

	resp := &http.Response{
		StatusCode: 200,
		Body:       newReadCloser(`{"success":true,"task_id":"se-123"}`),
	}
	taskID, body, taskErr := a.DoResponse(c, resp, info)
	assert.Equal(t, "se-123", taskID)
	assert.Nil(t, taskErr)
	assert.Contains(t, string(body), "se-123")
}
