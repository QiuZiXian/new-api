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

func TestImageCountPriceTierRoundTrip(t *testing.T) {
	withImageCountPrices(t, `{"qwen-image-3.0-pro":{"input":0.02,"output":0.1,"input_2k":0.04,"output_1k":0.18,"output_2k":0.36,"free_first_output":true}}`)

	price, ok := GetImageCountPrice("qwen-image-3.0-pro")
	require.True(t, ok)
	assert.InDelta(t, 0.04, price.Input2K, 1e-12)
	assert.InDelta(t, 0.18, price.Output1K, 1e-12)
	assert.InDelta(t, 0.36, price.Output2K, 1e-12)
	assert.True(t, price.FreeFirstOutput)
	// 最高输出档用于列表排序与 ModelPrice 兜底。
	assert.InDelta(t, 0.36, price.MaxOutputPrice(), 1e-12)
	assert.JSONEq(t,
		`{"qwen-image-3.0-pro":{"input":0.02,"output":0.1,"input_2k":0.04,"output_1k":0.18,"output_2k":0.36,"free_first_output":true}}`,
		ImageCountPrice2JSONString())
}

func TestImageCountPriceRejectsInvalidConfig(t *testing.T) {
	withImageCountPrices(t, `{}`)

	for _, raw := range []string{
		`{"m":{"input":0,"output":0}}`,        // 没有任何正价
		`{"m":{"input":-1,"output":0.18}}`,    // 负单价会变成退费
		`{"m":{"input":0.02,"output":2e9}}`,   // 超上限
		`{"m":{"input":"bad","output":0.18}}`, // 类型错误
		`{"m":{"output_2k":-0.3}}`,            // 档位价同样不允许为负
		`not-json`,
	} {
		assert.Error(t, UpdateImageCountPriceByJSONString(raw), raw)
	}
	// 校验失败不能污染已生效的配置表
	_, ok := GetImageCountPrice("m")
	assert.False(t, ok)
}

func TestImageTierByAreaBoundary(t *testing.T) {
	// 官方口径：面积 ≤2,250,000 为 1K，>2,250,000 为 2K。
	assert.Equal(t, ImageTier1K, ImageTierByArea(1500, 1500)) // 正好 2,250,000
	assert.Equal(t, ImageTier1K, ImageTierByArea(1024, 1024))
	assert.Equal(t, ImageTier1K, ImageTierByArea(1500, 1499))
	assert.Equal(t, ImageTier2K, ImageTierByArea(1500, 1501)) // 2,251,500
	assert.Equal(t, ImageTier2K, ImageTierByArea(2048, 2048))

	// 任一维度缺失时视为未知档位，调用方回退未分档单价。
	assert.Equal(t, "", ImageTierByArea(0, 1024))
	assert.Equal(t, "", ImageTierByArea(-10, 1024))
}

func TestImageTierBySize(t *testing.T) {
	assert.Equal(t, ImageTier1K, ImageTierBySize("1024*1024"))
	assert.Equal(t, ImageTier1K, ImageTierBySize("1024x1024"))
	assert.Equal(t, ImageTier2K, ImageTierBySize("2048*2048"))
	assert.Equal(t, ImageTier2K, ImageTierBySize("2048X2048"))
	assert.Equal(t, ImageTier2K, ImageTierBySize(" 1792*2400 "))
	assert.Equal(t, "", ImageTierBySize(""))
	assert.Equal(t, "", ImageTierBySize("auto"))
	assert.Equal(t, "", ImageTierBySize("1024"))
}

func TestNormalizeImageTier(t *testing.T) {
	assert.Equal(t, ImageTier1K, NormalizeImageTier("qima_input_1k"))
	assert.Equal(t, ImageTier2K, NormalizeImageTier("qima_output_2k"))
	assert.Equal(t, ImageTier2K, NormalizeImageTier("2K"))
	assert.Equal(t, ImageTier1K, NormalizeImageTier("1k"))
	assert.Equal(t, "", NormalizeImageTier(""))
	assert.Equal(t, "", NormalizeImageTier("qima_output_4k"))
}

func TestImageCountPriceTierFallback(t *testing.T) {
	// 只配默认价：任何档位都取默认价。
	flat := ImageCountPrice{Input: 0.02, Output: 0.18}
	assert.InDelta(t, 0.02, flat.InputPriceFor(ImageTier1K), 1e-12)
	assert.InDelta(t, 0.02, flat.InputPriceFor(ImageTier2K), 1e-12)
	assert.InDelta(t, 0.18, flat.OutputPriceFor(ImageTier2K), 1e-12)
	assert.InDelta(t, 0.18, flat.OutputPriceFor(""), 1e-12)

	// 配了 2K 输出档：2K 取档位价，1K/未知档位仍回退默认价。
	tiered := ImageCountPrice{Output: 0.18, Output2K: 0.36}
	assert.InDelta(t, 0.18, tiered.OutputPriceFor(ImageTier1K), 1e-12)
	assert.InDelta(t, 0.36, tiered.OutputPriceFor(ImageTier2K), 1e-12)
	assert.InDelta(t, 0.18, tiered.OutputPriceFor(""), 1e-12)
}

func TestImageCountCostAddsInputAndOutput(t *testing.T) {
	price := ImageCountPrice{Input: 0.02, Output: 0.18}

	// 2 张输入 + 3 张输出 = 0.04 + 0.54，绝不是连乘的 0.02*2*0.18*3
	assert.InDelta(t, 0.58, ImageCountCost(price, 2, 3, "", ""), 1e-9)
	assert.InDelta(t, 0.18, ImageCountCost(price, 0, 1, "", ""), 1e-9)
	assert.Equal(t, float64(0), ImageCountCost(price, 0, 0, "", ""))

	// 张数是用户/上游可控的乘数，必须钳到 MaxImageN
	assert.InDelta(t, float64(dto.MaxImageN)*0.18, ImageCountCost(price, 0, 999999, "", ""), 1e-9)
	assert.Equal(t, float64(0), ImageCountCost(price, -5, -5, "", ""))

	// 脏配置不能产生负费用
	assert.Equal(t, float64(0), ImageCountCost(ImageCountPrice{Input: -1, Output: -1}, 3, 3, "", ""))
}

func TestImageCountCostUsesTierPrice(t *testing.T) {
	price := ImageCountPrice{
		Input:    0.02,
		Output:   0.18,
		Input2K:  0.04,
		Output1K: 0.12,
		Output2K: 0.36,
	}

	// 1K：1 张输入 + 1 张输出 = 0.02 + 0.12
	assert.InDelta(t, 0.14, ImageCountCost(price, 1, 1, ImageTier1K, ImageTier1K), 1e-9)
	// 2K：1 张输入 + 1 张输出 = 0.04 + 0.36
	assert.InDelta(t, 0.40, ImageCountCost(price, 1, 1, ImageTier2K, ImageTier2K), 1e-9)
	// 未知档位回退默认价
	assert.InDelta(t, 0.20, ImageCountCost(price, 1, 1, "", ""), 1e-9)
}

func TestImageCountCostFirstOutputFree(t *testing.T) {
	price := ImageCountPrice{Output: 0.18, FreeFirstOutput: true}

	// 首张免费：1 张 = 0 元，3 张 = 2 张的钱。
	assert.Equal(t, float64(0), ImageCountCost(price, 0, 1, "", ""))
	assert.InDelta(t, 0.36, ImageCountCost(price, 0, 3, "", ""), 1e-9)
	// 输入侧不受首张免费影响。
	assert.InDelta(t, 0.02, ImageCountCost(ImageCountPrice{Input: 0.02, Output: 0.18, FreeFirstOutput: true}, 1, 1, "", ""), 1e-9)
	// 0 张输出不能被减成负数。
	assert.Equal(t, float64(0), ImageCountCost(price, 0, 0, "", ""))
}
