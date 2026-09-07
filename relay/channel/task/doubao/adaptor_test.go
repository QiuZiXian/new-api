package doubao

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeTaskApiPath(t *testing.T) {
	validSeedanceV1 := "/api/v1/contents/generations/tasks"
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty falls back to default", "", defaultDoubaoTaskApiPath},
		{"whitespace falls back to default", "   ", defaultDoubaoTaskApiPath},
		{"ark v3 path passes through", defaultDoubaoTaskApiPath, defaultDoubaoTaskApiPath},
		{"cii v1 path passes through", validSeedanceV1, validSeedanceV1},
		{"missing leading slash gets one prepended", "api/v1/contents/generations/tasks", validSeedanceV1},
		{"trims surrounding whitespace", "  " + validSeedanceV1 + "  ", validSeedanceV1},
		{"dot segment falls back to default", "/api/../v1/contents/generations/tasks", defaultDoubaoTaskApiPath},
		{"leading dot segment falls back to default", "../etc/passwd", defaultDoubaoTaskApiPath},
		{"backslash falls back to default", "/api\\v1\\tasks", defaultDoubaoTaskApiPath},
		{"query string falls back to default", validSeedanceV1 + "?page=1", defaultDoubaoTaskApiPath},
		{"fragment falls back to default", validSeedanceV1 + "#frag", defaultDoubaoTaskApiPath},
		{"scheme falls back to default", "https://evil.example.com" + validSeedanceV1, defaultDoubaoTaskApiPath},
		{"colon falls back to default", "/api:v1/tasks", defaultDoubaoTaskApiPath},
		{"space inside falls back to default", "/api/ v1/tasks", defaultDoubaoTaskApiPath},
		{"tab inside falls back to default", "/api/v1\ttasks", defaultDoubaoTaskApiPath},
		{"too long falls back to default", "/" + strings.Repeat("a", 257), defaultDoubaoTaskApiPath},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, normalizeTaskApiPath(tc.in))
		})
	}
}

func TestTaskApiPath_FallsBackWithoutChannelMeta(t *testing.T) {
	a := &TaskAdaptor{}
	assert.Equal(t, defaultDoubaoTaskApiPath, a.taskApiPath(nil))
}

func TestGetVideoInputRatio(t *testing.T) {
	cases := []struct {
		name       string
		model      string
		resolution string
		hasVideo   bool
		wantRatio  float64
		wantOK     bool
	}{
		// 2.0 全档
		{"2-0 480p no video", "doubao-seedance-2-0-260128", "", false, 1.0, true},
		{"2-0 480p with video", "doubao-seedance-2-0-260128", "", true, 28.0 / 46.0, true},
		{"2-0 1080p no video", "doubao-seedance-2-0-260128", "1080p", false, 51.0 / 46.0, true},
		{"2-0 4k with video", "doubao-seedance-2-0-260128", "4k", true, 16.0 / 46.0, true},
		{"2-0 unknown resolution", "doubao-seedance-2-0-260128", "999p", false, 1.0, true},

		// 2.0-fast 仅 480p
		{"2-0-fast 480p with video", "doubao-seedance-2-0-fast-260128", "", true, 22.0 / 37.0, true},
		{"2-0-fast 1080p falls back to base", "doubao-seedance-2-0-fast-260128", "1080p", false, 1.0, true},

		// 2.0-mini 仅 480p
		{"2-0-mini 480p no video", "doubao-seedance-2-0-mini", "", false, 1.0, true},
		{"2-0-mini with video", "doubao-seedance-2-0-mini", "", true, 16.0 / 28.0, true},
		{"2-0-mini 1080p falls back to base", "doubao-seedance-2-0-mini", "1080p", false, 1.0, true},

		// 2.5 支持到 1080p
		{"2.5 480p no video", "doubao-seedance-2.5", "", false, 1.0, true},
		{"2.5 with video", "doubao-seedance-2.5", "", true, 24.0 / 40.0, true},
		{"2.5 1080p", "doubao-seedance-2.5", "1080p", false, 46.0 / 40.0, true},
		{"2.5 4k falls back to base", "doubao-seedance-2.5", "4k", false, 1.0, true},

		// 1.5-pro / 1.0-pro / lite 系列
		{"1-5-pro 480p", "doubao-seedance-1-5-pro-251215", "", false, 1.0, true},
		{"1-5-pro with video", "doubao-seedance-1-5-pro-251215", "", true, 18.0 / 30.0, true},
		{"1-0-pro 480p", "doubao-seedance-1-0-pro-250528", "", false, 1.0, true},
		{"1-0-lite-t2v with video", "doubao-seedance-1-0-lite-t2v", "", true, 9.0 / 15.0, true},
		{"1-0-lite-i2v with video", "doubao-seedance-1-0-lite-i2v", "", true, 11.0 / 18.0, true},

		// 未知模型
		{"unknown model returns false", "doubao-seedance-9-9-999999", "", false, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ratio, ok := GetVideoInputRatio(tc.model, tc.resolution, tc.hasVideo)
			assert.Equal(t, tc.wantOK, ok)
			assert.InDelta(t, tc.wantRatio, ratio, 1e-9)
		})
	}
}

// TestEstimateBilling_Duration guards the billing invariant that user-supplied
// duration is bounded and added as the "seconds" OtherRatio. It also verifies
// that video_input ratio is composed with the seconds ratio, and that missing
// duration (relay/common.TaskSubmitReq.Duration == 0 && Seconds == "") does
// not add a seconds multiplier (the real bill is settled from completion_tokens).
func TestEstimateBilling_Duration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newCtxWithReq := func(t *testing.T, req common.TaskSubmitReq) *gin.Context {
		t.Helper()
		request := httptest.NewRequest(http.MethodPost, "/v1/video/generations", strings.NewReader("{}"))
		request.Header.Set("Content-Type", "application/json")
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = request
		ctx.Set("task_request", req)
		return ctx
	}

	a := &TaskAdaptor{}
	info := &common.RelayInfo{OriginModelName: "doubao-seedance-2-0-260128"}

	t.Run("no duration and no video gives nil (no multipliers)", func(t *testing.T) {
		ctx := newCtxWithReq(t, common.TaskSubmitReq{})
		assert.Nil(t, a.EstimateBilling(ctx, info))
	})

	t.Run("positive duration adds seconds ratio", func(t *testing.T) {
		ctx := newCtxWithReq(t, common.TaskSubmitReq{Duration: 10})
		ratios := a.EstimateBilling(ctx, info)
		require.NotNil(t, ratios)
		assert.InDelta(t, 10.0, ratios["seconds"], 1e-9)
		_, hasVideo := ratios["video_input"]
		assert.False(t, hasVideo, "no video input → no video_input multiplier")
	})

	t.Run("video input adds video_input ratio on top of seconds", func(t *testing.T) {
		ctx := newCtxWithReq(t, common.TaskSubmitReq{
			Duration: 5,
			Metadata: map[string]interface{}{
				"content": []interface{}{
					map[string]interface{}{"type": "video_url"},
				},
			},
		})
		ratios := a.EstimateBilling(ctx, info)
		require.NotNil(t, ratios)
		assert.InDelta(t, 5.0, ratios["seconds"], 1e-9)
		assert.InDelta(t, 28.0/46.0, ratios["video_input"], 1e-9)
	})

	t.Run("1080p+video adds composed ratio", func(t *testing.T) {
		ctx := newCtxWithReq(t, common.TaskSubmitReq{
			Duration: 6,
			Metadata: map[string]interface{}{
				"resolution": "1080p",
				"content": []interface{}{
					map[string]interface{}{"type": "video_url"},
				},
			},
		})
		ratios := a.EstimateBilling(ctx, info)
		require.NotNil(t, ratios)
		assert.InDelta(t, 6.0, ratios["seconds"], 1e-9)
		assert.InDelta(t, 31.0/46.0, ratios["video_input"], 1e-9)
	})

	t.Run("seconds string field is honored when duration is zero", func(t *testing.T) {
		ctx := newCtxWithReq(t, common.TaskSubmitReq{Seconds: "8"})
		ratios := a.EstimateBilling(ctx, info)
		require.NotNil(t, ratios)
		assert.InDelta(t, 8.0, ratios["seconds"], 1e-9)
	})

	t.Run("negative seconds is ignored", func(t *testing.T) {
		ctx := newCtxWithReq(t, common.TaskSubmitReq{Seconds: "-1"})
		ratios := a.EstimateBilling(ctx, info)
		assert.Nil(t, ratios, "duration=-1 means auto → no pre-consume multiplier")
	})

	t.Run("non-numeric seconds is ignored", func(t *testing.T) {
		ctx := newCtxWithReq(t, common.TaskSubmitReq{Seconds: "auto"})
		assert.Nil(t, a.EstimateBilling(ctx, info))
	})

	t.Run("oversized duration is clamped to MaxTaskDurationSeconds", func(t *testing.T) {
		ctx := newCtxWithReq(t, common.TaskSubmitReq{Duration: 1_000_000})
		ratios := a.EstimateBilling(ctx, info)
		require.NotNil(t, ratios)
		assert.InDelta(t, float64(common.MaxTaskDurationSeconds), ratios["seconds"], 1e-9)
	})
}
