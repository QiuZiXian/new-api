package helper

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withVideoSecondPriceConfig(t *testing.T, jsonStr string) {
	t.Helper()
	original := ratio_setting.VideoSecondPrice2JSONString()
	require.NoError(t, ratio_setting.UpdateVideoSecondPriceByJSONString(jsonStr))
	t.Cleanup(func() {
		require.NoError(t, ratio_setting.UpdateVideoSecondPriceByJSONString(original))
	})
}

// TestModelPriceHelperPerCallUsesVideoSecondPrice 守护「按秒配置本身就是价格」：
// 管理员只配了按秒档价、没有配按次 ModelPrice，也不该再报「请先配置模型价格」。
func TestModelPriceHelperPerCallUsesVideoSecondPrice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	withVideoSecondPriceConfig(t, `{"wan3.0-video":{"default":0.5,"720p":0.5}}`)

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/video/generations", nil)
	ctx.Set("group", "default")

	info := &relaycommon.RelayInfo{
		OriginModelName: "wan3.0-video",
		UserGroup:       "default",
		UsingGroup:      "default",
	}

	priceData, err := ModelPriceHelperPerCall(ctx, info)
	require.NoError(t, err)
	assert.True(t, priceData.UseVideoSecond)
	assert.False(t, priceData.UsePrice)
	assert.InDelta(t, 0.5, priceData.ModelPrice, 1e-12)

	expected, err := common.QuotaFromFloatStrict(0.5 * common.QuotaPerUnit * priceData.GroupRatioInfo.GroupRatio)
	require.NoError(t, err)
	assert.Equal(t, expected, priceData.Quota)

	// 只配按秒价格的模型也必须出现在 /v1/models 里。
	assert.True(t, HasModelBillingConfig("wan3.0-video"))
}

func TestHasModelBillingConfigRecognizesPerImagePrice(t *testing.T) {
	original := ratio_setting.ImageCountPrice2JSONString()
	require.NoError(t, ratio_setting.UpdateImageCountPriceByJSONString(`{"image-count-only":{"output":0.18}}`))
	t.Cleanup(func() {
		require.NoError(t, ratio_setting.UpdateImageCountPriceByJSONString(original))
	})
	assert.True(t, HasModelBillingConfig("image-count-only"))
}
