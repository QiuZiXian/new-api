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
