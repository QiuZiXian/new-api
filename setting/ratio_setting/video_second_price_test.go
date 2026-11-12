package ratio_setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withVideoSecondPrices(t *testing.T, jsonStr string) {
	t.Helper()
	original := VideoSecondPrice2JSONString()
	require.NoError(t, UpdateVideoSecondPriceByJSONString(jsonStr))
	t.Cleanup(func() {
		require.NoError(t, UpdateVideoSecondPriceByJSONString(original))
	})
}

func TestVideoSecondPriceNormalizesTiers(t *testing.T) {
	withVideoSecondPrices(t, `{"wan3.0-video":{"Default":0.5,"720P":0.5,"1080p":1.0}}`)

	variants, ok := GetVideoSecondPrice("wan3.0-video")
	require.True(t, ok)
	assert.Equal(t, map[string]float64{"default": 0.5, "720p": 0.5, "1080p": 1.0}, variants)

	base, ok := GetVideoSecondBasePrice("wan3.0-video")
	require.True(t, ok)
	assert.InDelta(t, 0.5, base, 1e-12)

	// 适配器/上游回报的是 "1080P" 这类写法，不能因此退化到 default 档。
	price, ok := GetVideoSecondVariantPrice("wan3.0-video", " 1080P ")
	require.True(t, ok)
	assert.InDelta(t, 1.0, price, 1e-12)

	ratio, ok := GetVideoSecondVariantRatio("wan3.0-video", "1080P")
	require.True(t, ok)
	assert.InDelta(t, 2.0, ratio, 1e-12)
}

func TestVideoSecondPriceAcceptsLegacyStringValues(t *testing.T) {
	// 早期前端把单价写成字符串，Go 的 map[string]float64 会整表拒绝加载，
	// 表现为「后台显示已保存、调用却提示未配置模型价格」。
	withVideoSecondPrices(t, `{"wan3.0-video":{"default":"0.5","720p":"0.5"}}`)

	base, ok := GetVideoSecondBasePrice("wan3.0-video")
	require.True(t, ok)
	assert.InDelta(t, 0.5, base, 1e-12)
	assert.JSONEq(t, `{"wan3.0-video":{"default":0.5,"720p":0.5}}`, VideoSecondPrice2JSONString())
}

func TestVideoSecondPriceRejectsInvalidConfigWithoutClobbering(t *testing.T) {
	withVideoSecondPrices(t, `{"wan3.0-video":{"default":0.5}}`)

	for _, raw := range []string{
		`{"m":{"default":-1}}`,    // 负单价会变成退费
		`{"m":{"default":"NaN"}}`, // 非有限数
		`{"m":{"default":2e9}}`,   // 超上限
		`{"m":{"default":true}}`,  // 类型错误
		`{"m":{"default":0}}`,     // 没有任何正价
		`{"":{"default":0.5}}`,    // 空模型名
		`{"m":{"  ":0.5}}`,        // 空档位名
		`not-json`,
	} {
		assert.Error(t, UpdateVideoSecondPriceByJSONString(raw), raw)
	}

	// 校验失败不能清空已生效的价格表。
	base, ok := GetVideoSecondBasePrice("wan3.0-video")
	require.True(t, ok)
	assert.InDelta(t, 0.5, base, 1e-12)
}
