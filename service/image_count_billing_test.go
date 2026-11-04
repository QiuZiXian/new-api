package service

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	hosttypes "github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newImageCountRelayInfo(input, output int) *relaycommon.RelayInfo {
	return newImageCountRelayInfoWithTiers(input, output, "", "")
}

func newImageCountRelayInfoWithTiers(input, output int, inputTier, outputTier string) *relaycommon.RelayInfo {
	priceData := hosttypes.PriceData{
		UseImageCount:  true,
		GroupRatioInfo: hosttypes.GroupRatioInfo{GroupRatio: 1},
	}
	priceData.SetImageCounts(input, output)
	priceData.SetImageTiers(inputTier, outputTier)
	return &relaycommon.RelayInfo{
		RelayFormat:             types.RelayFormatOpenAI,
		FinalRequestRelayFormat: types.RelayFormatOpenAI,
		OriginModelName:         "qwen-image-3.0-pro",
		PriceData:               priceData,
		StartTime:               time.Now(),
	}
}

func TestCalculateTextQuotaSummaryImageCountBilling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	original := ratio_setting.ImageCountPrice2JSONString()
	t.Cleanup(func() {
		require.NoError(t, ratio_setting.UpdateImageCountPriceByJSONString(original))
	})
	require.NoError(t, ratio_setting.UpdateImageCountPriceByJSONString(
		`{"qwen-image-3.0-pro":{"input":0.02,"output":0.18}}`))

	tests := []struct {
		name string
		in   int
		out  int
		want int
	}{
		{"output only", 0, 1, int(0.18 * common.QuotaPerUnit)},
		{"input and output are added", 2, 3, int(0.58 * common.QuotaPerUnit)},
		{"no images", 0, 0, 0},
		{"counts clamped to MaxImageN", 0, 999999, int(float64(dto.MaxImageN) * 0.18 * common.QuotaPerUnit)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			relayInfo := newImageCountRelayInfo(tt.in, tt.out)
			// 图片请求的 usage 通常没有 token，按张计费不能因此被清零
			summary := calculateTextQuotaSummary(ctx, relayInfo, &dto.Usage{})
			assert.Equal(t, tt.want, summary.Quota)
		})
	}
}

func TestImageCountBillingDoesNotMultiplyOtherRatioN(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	original := ratio_setting.ImageCountPrice2JSONString()
	t.Cleanup(func() {
		require.NoError(t, ratio_setting.UpdateImageCountPriceByJSONString(original))
	})
	require.NoError(t, ratio_setting.UpdateImageCountPriceByJSONString(
		`{"qwen-image-3.0-pro":{"input":0.02,"output":0.18}}`))

	relayInfo := newImageCountRelayInfo(0, 1)
	// 按次固定价链路遗留的 "n" 乘数不能再作用到按张模型上，否则重复扣费
	relayInfo.PriceData.AddOtherRatio("n", 3)
	summary := calculateTextQuotaSummary(ctx, relayInfo, &dto.Usage{})
	assert.Equal(t, int(0.18*common.QuotaPerUnit), summary.Quota)
	assert.True(t, summary.ImageCountBilling)
}

func TestImageCountBillingUsesResolutionTier(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	original := ratio_setting.ImageCountPrice2JSONString()
	t.Cleanup(func() {
		require.NoError(t, ratio_setting.UpdateImageCountPriceByJSONString(original))
	})
	require.NoError(t, ratio_setting.UpdateImageCountPriceByJSONString(
		`{"qwen-image-3.0-pro":{"input":0.02,"output":0.18,"input_2k":0.04,"output_1k":0.12,"output_2k":0.36}}`))

	tests := []struct {
		name       string
		inputTier  string
		outputTier string
		want       int
	}{
		{"1K tier", "1k", "1k", int(0.14 * common.QuotaPerUnit)},
		{"2K tier", "2k", "2k", int(0.40 * common.QuotaPerUnit)},
		{"unknown tier falls back to default", "", "", int(0.20 * common.QuotaPerUnit)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			relayInfo := newImageCountRelayInfoWithTiers(1, 1, tt.inputTier, tt.outputTier)
			summary := calculateTextQuotaSummary(ctx, relayInfo, &dto.Usage{})
			assert.Equal(t, tt.want, summary.Quota)
		})
	}
}

func TestImageCountBillingFirstOutputFree(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	original := ratio_setting.ImageCountPrice2JSONString()
	t.Cleanup(func() {
		require.NoError(t, ratio_setting.UpdateImageCountPriceByJSONString(original))
	})
	require.NoError(t, ratio_setting.UpdateImageCountPriceByJSONString(
		`{"qwen-image-3.0-pro":{"output":0.18,"free_first_output":true}}`))

	tests := []struct {
		name string
		out  int
		want int
	}{
		{"single image is free", 1, 0},
		{"only the first image is free", 3, int(0.36 * common.QuotaPerUnit)},
		{"no image means no charge", 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			relayInfo := newImageCountRelayInfo(0, tt.out)
			summary := calculateTextQuotaSummary(ctx, relayInfo, &dto.Usage{})
			assert.Equal(t, tt.want, summary.Quota)
		})
	}
}
