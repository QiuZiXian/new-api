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
	// 2.0 基准档的预扣还要乘实测补偿系数（applyPreConsumeFactor），此处直接按 2 倍给出期望值。
	expectTokens := func(seconds, width, height int) int {
		return 2 * int(math.Round(float64(seconds*width*height*24)/1024))
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

// TestConvertToRequestPayload_MapsGenerationParams 守护「客户端顶层参数必须到达上游」
// 这条契约：resolution / ratio / seed / watermark / camera_fixed / duration
// 曾经因为结构体缺字段被静默丢弃，上游只收到 model + content。
func TestConvertToRequestPayload_MapsGenerationParams(t *testing.T) {
	cfg := TaskAdaptor{}

	t.Run("top-level params reach upstream payload", func(t *testing.T) {
		req := &common.TaskSubmitReq{
			Model:       "doubao-seedance-2-0-260128",
			Prompt:      "a cat yawning",
			Duration:    10,
			Resolution:  "1080p",
			Ratio:       "9:16",
			Seed:        intPtr(11),
			Watermark:   boolPtr(true),
			CameraFixed: boolPtr(false),
		}

		payload, err := cfg.convertToRequestPayload(req)
		require.NoError(t, err)
		assert.Equal(t, "1080p", payload.Resolution)
		assert.Equal(t, "9:16", payload.Ratio)
		require.NotNil(t, payload.Seed)
		assert.Equal(t, 11, int(*payload.Seed))
		require.NotNil(t, payload.Watermark)
		assert.True(t, bool(*payload.Watermark))
		// 显式 false 不能被 omitempty 丢掉
		require.NotNil(t, payload.CameraFixed)
		assert.False(t, bool(*payload.CameraFixed))
		require.NotNil(t, payload.Duration)
		assert.Equal(t, 10, int(*payload.Duration))
		require.Len(t, payload.Content, 1)
		assert.Equal(t, "a cat yawning", payload.Content[0].Text)
	})

	t.Run("duration prefers int field over legacy seconds string", func(t *testing.T) {
		payload, err := cfg.convertToRequestPayload(&common.TaskSubmitReq{
			Model:    "doubao-seedance-2-0-260128",
			Prompt:   "p",
			Duration: 10,
			Seconds:  "5",
		})
		require.NoError(t, err)
		require.NotNil(t, payload.Duration)
		assert.Equal(t, 10, int(*payload.Duration))
	})

	t.Run("legacy seconds string still honored", func(t *testing.T) {
		payload, err := cfg.convertToRequestPayload(&common.TaskSubmitReq{
			Model:   "doubao-seedance-2-0-260128",
			Prompt:  "p",
			Seconds: "6",
		})
		require.NoError(t, err)
		require.NotNil(t, payload.Duration)
		assert.Equal(t, 6, int(*payload.Duration))
	})

	t.Run("duration -1 (smart selection) is forwarded", func(t *testing.T) {
		payload, err := cfg.convertToRequestPayload(&common.TaskSubmitReq{
			Model:    "doubao-seedance-2-0-260128",
			Prompt:   "p",
			Duration: -1,
		})
		require.NoError(t, err)
		require.NotNil(t, payload.Duration)
		assert.Equal(t, -1, int(*payload.Duration))
	})

	t.Run("top-level params override metadata workaround", func(t *testing.T) {
		payload, err := cfg.convertToRequestPayload(&common.TaskSubmitReq{
			Model:      "doubao-seedance-2-0-260128",
			Prompt:     "p",
			Resolution: "1080p",
			Ratio:      "1:1",
			Metadata: map[string]interface{}{
				"resolution": "720p",
				"ratio":      "16:9",
			},
		})
		require.NoError(t, err)
		assert.Equal(t, "1080p", payload.Resolution)
		assert.Equal(t, "1:1", payload.Ratio)
	})

	t.Run("metadata fills params that top level omits", func(t *testing.T) {
		payload, err := cfg.convertToRequestPayload(&common.TaskSubmitReq{
			Model:  "doubao-seedance-2-0-260128",
			Prompt: "p",
			Metadata: map[string]interface{}{
				"resolution": "720p",
			},
		})
		require.NoError(t, err)
		assert.Equal(t, "720p", payload.Resolution)
	})
}

// TestValidateGenerationParams 守护本地参数校验：非法枚举/超范围时长必须在
// 预扣费之前被拦下，而不是打到上游后由上游 4xx（常被余额不足等业务错误掩盖）。
func TestValidateGenerationParams(t *testing.T) {
	cases := []struct {
		name    string
		req     common.TaskSubmitReq
		wantErr string
	}{
		{"valid 2.0 request", common.TaskSubmitReq{Model: "doubao-seedance-2.0", Resolution: "1080p", Ratio: "16:9", Duration: 10}, ""},
		{"valid upper-case resolution", common.TaskSubmitReq{Model: "doubao-seedance-2.0", Resolution: "4K", Duration: 5}, ""},
		{"valid adaptive ratio", common.TaskSubmitReq{Model: "doubao-seedance-2.0", Ratio: "adaptive", Duration: 5}, ""},
		{"no optional params", common.TaskSubmitReq{Model: "doubao-seedance-2.0"}, ""},
		{"unknown resolution", common.TaskSubmitReq{Model: "doubao-seedance-2.0", Resolution: "2k"}, "invalid_resolution"},
		{"unknown ratio", common.TaskSubmitReq{Model: "doubao-seedance-2.0", Ratio: "7:3"}, "invalid_ratio"},
		{"2.0 duration above model range", common.TaskSubmitReq{Model: "doubao-seedance-2.0", Duration: 16}, "invalid_duration"},
		{"2.0 duration=999", common.TaskSubmitReq{Model: "doubao-seedance-2.0-fast", Duration: 999}, "invalid_duration"},
		{"2.0 duration below model range", common.TaskSubmitReq{Model: "doubao-seedance-2.0-mini", Duration: 3}, "invalid_duration"},
		{"2.5 allows 30s", common.TaskSubmitReq{Model: "doubao-seedance-2.5", Duration: 30}, ""},
		{"2.5 rejects 31s", common.TaskSubmitReq{Model: "doubao-seedance-2.5", Duration: 31}, "invalid_duration"},
		{"1.5-pro rejects 13s", common.TaskSubmitReq{Model: "doubao-seedance-1.5-pro", Duration: 13}, "invalid_duration"},
		{"unlisted model keeps generic bounds", common.TaskSubmitReq{Model: "doubao-seedance-1.0-pro", Duration: 60}, ""},
		{"seed -1 is allowed", common.TaskSubmitReq{Model: "doubao-seedance-2.0", Seed: intPtr(-1)}, ""},
		{"seed below -1 rejected", common.TaskSubmitReq{Model: "doubao-seedance-2.0", Seed: intPtr(-2)}, "invalid_seed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			taskErr := validateGenerationParams(&tc.req)
			if tc.wantErr == "" {
				assert.Nil(t, taskErr)
				return
			}
			require.NotNil(t, taskErr)
			assert.Equal(t, tc.wantErr, taskErr.Code)
			assert.Equal(t, http.StatusBadRequest, taskErr.StatusCode)
		})
	}
}

// TestApplyPreConsumeFactor 守护预扣补偿只作用于 2.0 基准档：
// 实测 2.0 的 token 是官方公式的 2 倍，fast / mini 与公式一致。
func TestApplyPreConsumeFactor(t *testing.T) {
	assert.Equal(t, 216000, applyPreConsumeFactor("doubao-seedance-2.0", 108000))
	assert.Equal(t, 216000, applyPreConsumeFactor("doubao-seedance-2-0-260128", 108000))
	assert.Equal(t, 108000, applyPreConsumeFactor("doubao-seedance-2.0-fast", 108000))
	assert.Equal(t, 108000, applyPreConsumeFactor("doubao-seedance-2.0-mini", 108000))
	assert.Equal(t, 0, applyPreConsumeFactor("doubao-seedance-2.0", 0))
}

func intPtr(v int) *int    { return &v }
func boolPtr(v bool) *bool { return &v }
