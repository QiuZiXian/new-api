package ali

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRelayInfo() *relaycommon.RelayInfo {
	return &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{},
	}
}

func TestConvertToAliRequestWan27I2VBuildsMediaFromImage(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:    "wan2.7-i2v",
		Prompt:   "animate the first frame",
		Image:    "https://example.com/first.png",
		Size:     "720p",
		Duration: 10,
	}

	aliReq, err := adaptor.convertToAliRequest(testRelayInfo(), req)

	require.NoError(t, err)
	require.Equal(t, "wan2.7-i2v", aliReq.Model)
	require.Equal(t, "720P", aliReq.Parameters.Resolution)
	require.Equal(t, 10, aliReq.Parameters.Duration)
	require.Equal(t, []AliVideoMedia{
		{Type: "first_frame", URL: "https://example.com/first.png"},
	}, aliReq.Input.Media)
	require.Empty(t, aliReq.Input.ImgURL)

	body, err := common.Marshal(aliReq)
	require.NoError(t, err)
	require.Contains(t, string(body), `"media"`)
	require.NotContains(t, string(body), `"img_url"`)
}

func TestConvertToAliRequestWan27I2VBuildsFirstAndLastFrameFromImages(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:  "wan2.7-i2v",
		Prompt: "interpolate between frames",
		Images: []string{
			"https://example.com/first.png",
			"https://example.com/last.png",
		},
	}

	aliReq, err := adaptor.convertToAliRequest(testRelayInfo(), req)

	require.NoError(t, err)
	require.Equal(t, []AliVideoMedia{
		{Type: "first_frame", URL: "https://example.com/first.png"},
		{Type: "last_frame", URL: "https://example.com/last.png"},
	}, aliReq.Input.Media)
}

func TestConvertToAliRequestWan27I2VPrefersImageBeforeImagesAndInputReference(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:          "wan2.7-i2v",
		Prompt:         "use the direct image",
		Image:          " https://example.com/direct.png ",
		Images:         []string{"https://example.com/images-first.png", " https://example.com/images-last.png "},
		InputReference: "https://example.com/input-reference.png",
	}

	aliReq, err := adaptor.convertToAliRequest(testRelayInfo(), req)

	require.NoError(t, err)
	require.Equal(t, []AliVideoMedia{
		{Type: "first_frame", URL: "https://example.com/direct.png"},
		{Type: "last_frame", URL: "https://example.com/images-last.png"},
	}, aliReq.Input.Media)
}

func TestConvertToAliRequestWan27I2VFallsBackToFirstNonEmptyImage(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:  "wan2.7-i2v",
		Prompt: "skip blank images",
		Image:  " ",
		Images: []string{
			" ",
			" https://example.com/first.png ",
			" https://example.com/last.png ",
		},
		InputReference: "https://example.com/input-reference.png",
	}

	aliReq, err := adaptor.convertToAliRequest(testRelayInfo(), req)

	require.NoError(t, err)
	require.Equal(t, []AliVideoMedia{
		{Type: "first_frame", URL: "https://example.com/first.png"},
		{Type: "last_frame", URL: "https://example.com/last.png"},
	}, aliReq.Input.Media)
}

func TestConvertToAliRequestWan27I2VKeepsExplicitMetadataMedia(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:          "wan2.7-i2v",
		Prompt:         "continue the clip",
		Image:          "https://example.com/direct.png",
		Images:         []string{"https://example.com/images-first.png", "https://example.com/images-last.png"},
		InputReference: "https://example.com/input-reference.png",
		Metadata: map[string]interface{}{
			"input": map[string]interface{}{
				"media": []interface{}{
					map[string]interface{}{
						"type": "first_clip",
						"url":  "https://example.com/input.mp4",
					},
				},
			},
		},
	}

	aliReq, err := adaptor.convertToAliRequest(testRelayInfo(), req)

	require.NoError(t, err)
	require.Equal(t, []AliVideoMedia{
		{Type: "first_clip", URL: "https://example.com/input.mp4"},
	}, aliReq.Input.Media)
	require.Empty(t, aliReq.Input.ImgURL)

	body, err := common.Marshal(aliReq)
	require.NoError(t, err)
	require.Contains(t, string(body), `"media"`)
	require.NotContains(t, string(body), `"img_url"`)
}

func TestConvertToAliRequestWan27I2VRequiresMedia(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:  "wan2.7-i2v",
		Prompt: "animate without a frame",
	}

	_, err := adaptor.convertToAliRequest(testRelayInfo(), req)

	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "requires image"))
}

func TestConvertToAliRequestWan25I2VKeepsLegacyImgURL(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:  "wan2.5-i2v-preview",
		Prompt: "animate the first frame",
		Image:  "https://example.com/first.png",
	}

	aliReq, err := adaptor.convertToAliRequest(testRelayInfo(), req)

	require.NoError(t, err)
	require.Equal(t, "https://example.com/first.png", aliReq.Input.ImgURL)
	require.Empty(t, aliReq.Input.Media)

	body, err := common.Marshal(aliReq)
	require.NoError(t, err)
	require.Contains(t, string(body), `"img_url"`)
	require.NotContains(t, string(body), `"media"`)
}

// ---------------------------------------------------------------------------
// wan3.0-video 按秒计费链路
// ---------------------------------------------------------------------------

func withVideoSecondPrice(t *testing.T, jsonStr string) {
	t.Helper()
	original := ratio_setting.VideoSecondPrice2JSONString()
	require.NoError(t, ratio_setting.UpdateVideoSecondPriceByJSONString(jsonStr))
	t.Cleanup(func() {
		require.NoError(t, ratio_setting.UpdateVideoSecondPriceByJSONString(original))
	})
}

func taskContext(req relaycommon.TaskSubmitReq) *gin.Context {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/video/generations", nil)
	ctx.Set("task_request", req)
	return ctx
}

func perSecondRelayInfo() *relaycommon.RelayInfo {
	info := testRelayInfo()
	info.OriginModelName = "wan3.0-video"
	info.PriceData.UseVideoSecond = true
	return info
}

func TestConvertToAliRequestUsesTopLevelResolution(t *testing.T) {
	adaptor := &TaskAdaptor{}
	aliReq, err := adaptor.convertToAliRequest(testRelayInfo(), relaycommon.TaskSubmitReq{
		Model:      "wan3.0-video",
		Prompt:     "小猫对着镜头打哈欠",
		Duration:   5,
		Resolution: "720p",
		Ratio:      "16:9",
	})

	require.NoError(t, err)
	require.Equal(t, "720P", aliReq.Parameters.Resolution)
	require.Equal(t, 5, aliReq.Parameters.Duration)
}

func TestEstimateBillingUsesVideoSecondTiersForWan3(t *testing.T) {
	withVideoSecondPrice(t, `{"wan3.0-video":{"default":0.5,"720p":0.5,"1080p":1.0}}`)

	ctx := taskContext(relaycommon.TaskSubmitReq{
		Model:      "wan3.0-video",
		Prompt:     "小猫对着镜头打哈欠",
		Duration:   5,
		Resolution: "1080p",
		Ratio:      "16:9",
	})
	info := perSecondRelayInfo()

	ratios := (&TaskAdaptor{}).EstimateBilling(ctx, info)

	require.InDelta(t, 5.0, ratios["seconds"], 1e-9)
	// 1080p 是 default 档价的 2 倍。
	require.InDelta(t, 2.0, ratios["video_variant"], 1e-9)
	// 按秒分档价已含分辨率差异，不能再叠加按 token 口径的硬编码倍率表。
	assert.NotContains(t, ratios, "resolution-1080P")
	assert.True(t, info.PriceData.IsOtherRatioEstimateOnly("seconds"))
	assert.True(t, info.PriceData.IsOtherRatioEstimateOnly("video_variant"))
}

func TestEstimateBillingKeepsLegacyRatiosWhenNotPerSecond(t *testing.T) {
	ctx := taskContext(relaycommon.TaskSubmitReq{
		Model:    "wan2.5-t2v-preview",
		Prompt:   "a cat",
		Duration: 4,
		Size:     "1280*720",
	})
	info := testRelayInfo()
	info.OriginModelName = "wan2.5-t2v-preview"

	ratios := (&TaskAdaptor{}).EstimateBilling(ctx, info)

	assert.InDelta(t, 4.0, ratios["seconds"], 1e-9)
	assert.InDelta(t, 2.0, ratios["resolution-720P"], 1e-9)
	assert.False(t, info.PriceData.IsOtherRatioEstimateOnly("seconds"))
}

func TestParseTaskResultReportsRealVideoDuration(t *testing.T) {
	adaptor := &TaskAdaptor{}

	// 上游 usage 同时给 duration 与 output_video_duration，以输出时长为准。
	result, err := adaptor.ParseTaskResult([]byte(`{
		"output":{"task_id":"t1","task_status":"SUCCEEDED","video_url":"https://example.com/a.mp4"},
		"usage":{"SR":720,"fps":30,"ratio":"16:9","duration":5,"video_count":1,
		         "input_video_duration":0,"output_video_duration":5},
		"request_id":"r1"
	}`))
	require.NoError(t, err)
	assert.Equal(t, 5, result.VideoSeconds)
	assert.Equal(t, "720p", result.VideoVariant)
	assert.Equal(t, "https://example.com/a.mp4", result.Url)

	// 老版本只回报 duration 时同样要能按秒结算。
	legacy, err := adaptor.ParseTaskResult([]byte(`{
		"output":{"task_status":"SUCCEEDED"},
		"usage":{"duration":8}
	}`))
	require.NoError(t, err)
	assert.Equal(t, 8, legacy.VideoSeconds)
	assert.Empty(t, legacy.VideoVariant)

	// 没有 usage 时不猜时长，结算层保留预扣额度。
	noUsage, err := adaptor.ParseTaskResult([]byte(`{"output":{"task_status":"SUCCEEDED"}}`))
	require.NoError(t, err)
	assert.Zero(t, noUsage.VideoSeconds)
}
