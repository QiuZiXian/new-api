package ratio_setting

import (
	"testing"

	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withImageCountPrices(t *testing.T, jsonStr string) {
	original := ImageCountPrice2JSONString()
	t.Cleanup(func() {
		require.NoError(t, UpdateImageCountPriceByJSONString(original))
	})
	require.NoError(t, UpdateImageCountPriceByJSONString(jsonStr))
}

func TestImageCountPriceRoundTrip(t *testing.T) {
	withImageCountPrices(t, `{"qwen-image-3.0-pro":{"input":0.02,"output":0.18}}`)

	price, ok := GetImageCountPrice("qwen-image-3.0-pro")
	require.True(t, ok)
	assert.InDelta(t, 0.02, price.Input, 1e-12)
	assert.InDelta(t, 0.18, price.Output, 1e-12)

	_, ok = GetImageCountPrice("not-configured")
	assert.False(t, ok)
	assert.JSONEq(t, `{"qwen-image-3.0-pro":{"input":0.02,"output":0.18}}`, ImageCountPrice2JSONString())
}

func TestImageCountPriceRejectsInvalidConfig(t *testing.T) {
	withImageCountPrices(t, `{}`)

	for _, raw := range []string{
		`{"m":{"input":0,"output":0}}`,        // 没有任何正价
		`{"m":{"input":-1,"output":0.18}}`,    // 负单价会变成退费
		`{"m":{"input":0.02,"output":2e9}}`,   // 超上限
		`{"m":{"input":"bad","output":0.18}}`, // 类型错误
		`not-json`,
	} {
		assert.Error(t, UpdateImageCountPriceByJSONString(raw), raw)
	}
	// 校验失败不能污染已生效的配置表
	_, ok := GetImageCountPrice("m")
	assert.False(t, ok)
}

func TestImageCountCostAddsInputAndOutput(t *testing.T) {
	price := ImageCountPrice{Input: 0.02, Output: 0.18}

	// 2 张输入 + 3 张输出 = 0.04 + 0.54，绝不是连乘的 0.02*2*0.18*3
	assert.InDelta(t, 0.58, ImageCountCost(price, 2, 3), 1e-9)
	assert.InDelta(t, 0.18, ImageCountCost(price, 0, 1), 1e-9)
	assert.Equal(t, float64(0), ImageCountCost(price, 0, 0))

	// 张数是用户/上游可控的乘数，必须钳到 MaxImageN
	assert.InDelta(t, float64(dto.MaxImageN)*0.18, ImageCountCost(price, 0, 999999), 1e-9)
	assert.Equal(t, float64(0), ImageCountCost(price, -5, -5))

	// 脏配置不能产生负费用
	assert.Equal(t, float64(0), ImageCountCost(ImageCountPrice{Input: -1, Output: -1}, 3, 3))
}
