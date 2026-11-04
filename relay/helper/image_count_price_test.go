package helper

import (
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/setting/ratio_setting"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelPriceHelperImageCountPreConsume(t *testing.T) {
	gin.SetMode(gin.TestMode)
	saved := ratio_setting.ImageCountPrice2JSONString()
	t.Cleanup(func() {
		require.NoError(t, ratio_setting.UpdateImageCountPriceByJSONString(saved))
	})
	require.NoError(t, ratio_setting.UpdateImageCountPriceByJSONString(
		`{"image-count-model":{"input":0.02,"output":0.18}}`))

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Set("group", "default")
	info := &relaycommon.RelayInfo{
		OriginModelName: "image-count-model",
		UserGroup:       "default",
		UsingGroup:      "default",
	}
	meta := &types.TokenCountMeta{BillingRatios: map[string]float64{"n": 3}}

	priceData, err := ModelPriceHelper(ctx, info, 1000, meta)
	require.NoError(t, err)
	assert.True(t, priceData.UseImageCount)
	assert.False(t, priceData.UsePrice, "按张计费不得同时走按次固定价")
	assert.Equal(t, 3, priceData.OutputImageCount)
	assert.Equal(t, 0, priceData.InputImageCount, "预扣阶段尚无上游 usage")
	assert.Equal(t, int(0.18*3*common.QuotaPerUnit), priceData.QuotaToPreConsume)
	assert.True(t, info.PriceData.UseImageCount)
}

func TestModelPriceHelperImageCountIgnoresTokenRatios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	saved := ratio_setting.ImageCountPrice2JSONString()
	t.Cleanup(func() {
		require.NoError(t, ratio_setting.UpdateImageCountPriceByJSONString(saved))
	})
	require.NoError(t, ratio_setting.UpdateImageCountPriceByJSONString(
		`{"image-count-model":{"input":0.02,"output":0.18}}`))

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Set("group", "default")
	info := &relaycommon.RelayInfo{
		OriginModelName: "image-count-model",
		UserGroup:       "default",
		UsingGroup:      "default",
	}

	// 即使没传 n，也要按 1 张输出预扣，且不会退化成 token 计费
	priceData, err := ModelPriceHelper(ctx, info, 1000, &types.TokenCountMeta{})
	require.NoError(t, err)
	assert.True(t, priceData.UseImageCount)
	assert.Equal(t, 1, priceData.OutputImageCount)
	assert.Zero(t, priceData.ModelRatio)
	assert.Equal(t, int(0.18*common.QuotaPerUnit), priceData.QuotaToPreConsume)
}

func TestModelPriceHelperImageCountUsesRequestSizeTier(t *testing.T) {
	gin.SetMode(gin.TestMode)
	saved := ratio_setting.ImageCountPrice2JSONString()
	t.Cleanup(func() {
		require.NoError(t, ratio_setting.UpdateImageCountPriceByJSONString(saved))
	})
	require.NoError(t, ratio_setting.UpdateImageCountPriceByJSONString(
		`{"image-count-model":{"output":0.18,"output_1k":0.12,"output_2k":0.36}}`))

	tests := []struct {
		name     string
		size     string
		wantTier string
		wantEach float64
	}{
		{"1K size", "1024x1024", "1k", 0.12},
		{"2K size", "2048x2048", "2k", 0.36},
		{"exact 1K boundary", "1500x1500", "1k", 0.12},
		{"just above 1K boundary", "1500x1501", "2k", 0.36},
		{"missing size falls back to default price", "", "", 0.18},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Set("group", "default")
			info := &relaycommon.RelayInfo{
				OriginModelName: "image-count-model",
				UserGroup:       "default",
				UsingGroup:      "default",
			}
			meta := &types.TokenCountMeta{
				ImageSize:     tt.size,
				BillingRatios: map[string]float64{"n": 2},
			}

			priceData, err := ModelPriceHelper(ctx, info, 1000, meta)
			require.NoError(t, err)
			assert.Equal(t, tt.wantTier, priceData.OutputImageTier)
			assert.Equal(t, 2, priceData.OutputImageCount)
			assert.Equal(t, int(tt.wantEach*2*common.QuotaPerUnit), priceData.QuotaToPreConsume)
		})
	}
}

func TestModelPriceHelperImageCountFirstOutputFree(t *testing.T) {
	gin.SetMode(gin.TestMode)
	saved := ratio_setting.ImageCountPrice2JSONString()
	t.Cleanup(func() {
		require.NoError(t, ratio_setting.UpdateImageCountPriceByJSONString(saved))
	})
	require.NoError(t, ratio_setting.UpdateImageCountPriceByJSONString(
		`{"image-count-model":{"output":0.18,"free_first_output":true}}`))

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Set("group", "default")
	info := &relaycommon.RelayInfo{
		OriginModelName: "image-count-model",
		UserGroup:       "default",
		UsingGroup:      "default",
	}

	// 单张请求命中首张免费，预扣为 0，不能冻结用户额度。
	priceData, err := ModelPriceHelper(ctx, info, 1000, &types.TokenCountMeta{})
	require.NoError(t, err)
	assert.Equal(t, 1, priceData.OutputImageCount)
	assert.Equal(t, 0, priceData.QuotaToPreConsume)
}
