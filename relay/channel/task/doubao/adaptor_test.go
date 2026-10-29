package doubao

import (
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
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
	newInfo := func() *common.RelayInfo {
		return &common.RelayInfo{OriginModelName: "doubao-seedance-2-0-260128"}
	}
	// 官方公式：tokens = (输入时长 + 输出时长) × 宽 × 高 × 帧率 / 1024
	expectTokens := func(seconds, width, height int) int {
		return int(math.Round(float64(seconds*width*height*24) / 1024))
	}

	t.Run("no duration and no video still estimates tokens at defaults", func(t *testing.T) {
		info := newInfo()
		assert.Nil(t, a.EstimateBilling(newCtxWithReq(t, common.TaskSubmitReq{}), info))
		// 上游默认 720p / 5s，预扣按官方公式估算，而不是固定 25 万 token
		assert.Equal(t, expectTokens(5, 1280, 720), info.PriceData.EstimatedTokens)
	})

	t.Run("positive duration estimates tokens instead of seconds ratio", func(t *testing.T) {
		info := newInfo()
		ratios := a.EstimateBilling(newCtxWithReq(t, common.TaskSubmitReq{Duration: 10}), info)
		assert.Nil(t, ratios, "no video input → no other ratios")
		assert.Equal(t, expectTokens(10, 1280, 720), info.PriceData.EstimatedTokens)
		assert.False(t, info.PriceData.IsOtherRatioEstimateOnly("seconds"))
	})

	t.Run("video input adds video_input ratio and keeps token estimate", func(t *testing.T) {
		info := newInfo()
		ratios := a.EstimateBilling(newCtxWithReq(t, common.TaskSubmitReq{
			Duration: 5,
			Metadata: map[string]interface{}{
				"content": []interface{}{
					map[string]interface{}{"type": "video_url"},
				},
			},
		}), info)
		require.NotNil(t, ratios)
		assert.InDelta(t, 28.0/46.0, ratios["video_input"], 1e-9)
		_, hasSeconds := ratios["seconds"]
		assert.False(t, hasSeconds, "token 已含时长，不能再乘 seconds")
		assert.Equal(t, expectTokens(5, 1280, 720), info.PriceData.EstimatedTokens)
	})

	t.Run("1080p+video uses 1080p pixels", func(t *testing.T) {
		info := newInfo()
		ratios := a.EstimateBilling(newCtxWithReq(t, common.TaskSubmitReq{
			Duration: 6,
			Metadata: map[string]interface{}{
				"resolution": "1080p",
				"content": []interface{}{
					map[string]interface{}{"type": "video_url"},
				},
			},
		}), info)
		require.NotNil(t, ratios)
		assert.InDelta(t, 31.0/46.0, ratios["video_input"], 1e-9)
		assert.Equal(t, expectTokens(6, 1920, 1080), info.PriceData.EstimatedTokens)
	})

	t.Run("portrait ratio swaps width and height", func(t *testing.T) {
		info := newInfo()
		a.EstimateBilling(newCtxWithReq(t, common.TaskSubmitReq{
			Duration: 5,
			Metadata: map[string]interface{}{
				"resolution": "720p",
				"ratio":      "9:16",
			},
		}), info)
		assert.Equal(t, expectTokens(5, 720, 1280), info.PriceData.EstimatedTokens)
	})

	t.Run("seconds string field is honored when duration is zero", func(t *testing.T) {
		info := newInfo()
		a.EstimateBilling(newCtxWithReq(t, common.TaskSubmitReq{Seconds: "8"}), info)
		assert.Equal(t, expectTokens(8, 1280, 720), info.PriceData.EstimatedTokens)
	})

	t.Run("oversized duration is clamped to MaxTaskDurationSeconds", func(t *testing.T) {
		info := newInfo()
		a.EstimateBilling(newCtxWithReq(t, common.TaskSubmitReq{Duration: 1_000_000}), info)
		assert.Equal(t, expectTokens(common.MaxTaskDurationSeconds, 1280, 720), info.PriceData.EstimatedTokens)
	})

	t.Run("unknown resolution falls back to seconds ratio", func(t *testing.T) {
		info := newInfo()
		ratios := a.EstimateBilling(newCtxWithReq(t, common.TaskSubmitReq{
			Duration: 4,
			Metadata: map[string]interface{}{"resolution": "8k"},
		}), info)
		require.NotNil(t, ratios)
		assert.InDelta(t, 4.0, ratios["seconds"], 1e-9)
		assert.True(t, info.PriceData.IsOtherRatioEstimateOnly("seconds"))
		assert.Zero(t, info.PriceData.EstimatedTokens)
	})
}

// TestEstimateBilling_EstimateOnlyRatios 守护「按秒」与「按 token」两种乘数不混用：
// seconds 只放大预扣冻结额度，终态结算不再乘；video_input 属于按 token 单价档，
// 两个口径都要参与结算。
func TestEstimateBilling_EstimateOnlyRatios(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newCtx := func(t *testing.T) *gin.Context {
		t.Helper()
		request := httptest.NewRequest(http.MethodPost, "/v1/video/generations", strings.NewReader("{}"))
		request.Header.Set("Content-Type", "application/json")
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = request
		ctx.Set("task_request", common.TaskSubmitReq{
			Duration: 5,
			Metadata: map[string]interface{}{
				"resolution": "1080p",
				"content": []interface{}{
					map[string]interface{}{"type": "video_url"},
				},
			},
		})
		return ctx
	}

	a := &TaskAdaptor{}

	t.Run("token mode", func(t *testing.T) {
		info := &common.RelayInfo{OriginModelName: "doubao-seedance-2-0-260128"}
		ratios := a.EstimateBilling(newCtx(t), info)
		require.NotNil(t, ratios)
		// 调用方（RelayTaskSubmit）负责把返回的 ratios 写回 PriceData
		for key, ratio := range ratios {
			info.PriceData.AddOtherRatio(key, ratio)
		}

		// 预估 token 已含时长，token 模式不再产生 seconds 倍率；
		// video_input 是「每 token 单价档」，预扣和结算都要乘。
		assert.False(t, info.PriceData.IsOtherRatioEstimateOnly("video_input"))
		assert.Greater(t, info.PriceData.EstimatedTokens, 0)
		assert.InDelta(t, 31.0/46.0, info.PriceData.SettlementOtherRatioMultiplier(), 1e-9)
		assert.InDelta(t, 31.0/46.0, info.PriceData.OtherRatioMultiplier(), 1e-9)
	})

	t.Run("per-second mode excludes both second-dimension ratios", func(t *testing.T) {
		original := ratio_setting.VideoSecondPrice2JSONString()
		require.NoError(t, ratio_setting.UpdateVideoSecondPriceByJSONString(
			`{"doubao-seedance-2-0-260128":{"default":1,"1080p":1.51}}`))
		t.Cleanup(func() {
			require.NoError(t, ratio_setting.UpdateVideoSecondPriceByJSONString(original))
		})

		info := &common.RelayInfo{OriginModelName: "doubao-seedance-2-0-260128"}
		info.PriceData.UseVideoSecond = true
		ratios := a.EstimateBilling(newCtx(t), info)
		require.NotNil(t, ratios)
		for key, ratio := range ratios {
			info.PriceData.AddOtherRatio(key, ratio)
		}

		assert.InDelta(t, 5.0, ratios["seconds"], 1e-9)
		assert.InDelta(t, 1.51, ratios["video_variant"], 1e-9)
		assert.True(t, info.PriceData.IsOtherRatioEstimateOnly("seconds"))
		assert.True(t, info.PriceData.IsOtherRatioEstimateOnly("video_variant"))
		// 预扣乘满，结算不乘
		assert.InDelta(t, 5*1.51, info.PriceData.OtherRatioMultiplier(), 1e-9)
		assert.InDelta(t, 1.0, info.PriceData.SettlementOtherRatioMultiplier(), 1e-9)
	})
}
