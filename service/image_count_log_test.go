package service

import (
	"testing"

	"github.com/QuantumNous/new-api/setting/ratio_setting"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setImageCountPriceForTest(t *testing.T, jsonStr string) {
	t.Helper()
	original := ratio_setting.ImageCountPrice2JSONString()
	t.Cleanup(func() {
		require.NoError(t, ratio_setting.UpdateImageCountPriceByJSONString(original))
	})
	require.NoError(t, ratio_setting.UpdateImageCountPriceByJSONString(jsonStr))
}

// 纯文生图：上游仍会返回 input_image_type，但 0 张输入不能写档位，
// 输出单价必须取实际命中的档位价而不是默认价。
func TestAppendImageCountLogInfoSkipsEmptyInputTier(t *testing.T) {
	setImageCountPriceForTest(t, `{"qwen-image-3.0-pro":{
		"input":0.02,"output":0.18,
		"input_2k":0.03,"output_2k":0.25,
		"free_first_output":true
	}}`)

	relayInfo := newImageCountRelayInfoWithTiers(0, 2, "2k", "2k")
	other := map[string]interface{}{}
	appendImageCountLogInfo(other, relayInfo, "qwen-image-3.0-pro")

	assert.Equal(t, true, other["image_count_billing"])
	assert.Equal(t, 0, other["input_image_count"])
	assert.Equal(t, 2, other["output_image_count"])
	assert.NotContains(t, other, "input_image_tier", "0 张输入不应记录档位")
	assert.Equal(t, "2k", other["output_image_tier"])
	assert.InDelta(t, 0.25, other["image_output_price"], 1e-9, "应记录实际档位单价")
	assert.NotContains(t, other, "image_input_price", "0 张输入没有生效单价")
	assert.Equal(t, true, other["image_free_first_output"])
	assert.Equal(t, 1, other["billable_output_image_count"], "首张免费后计费 1 张")
}

// 图生图：输入输出都有图，两侧档位与档位价都要落日志。
func TestAppendImageCountLogInfoRecordsBothSides(t *testing.T) {
	setImageCountPriceForTest(t, `{"qwen-image-3.0-pro":{
		"input":0.02,"output":0.18,
		"input_1k":0.015,"output_1k":0.12
	}}`)

	relayInfo := newImageCountRelayInfoWithTiers(1, 3, "1k", "1k")
	other := map[string]interface{}{}
	appendImageCountLogInfo(other, relayInfo, "qwen-image-3.0-pro")

	assert.Equal(t, "1k", other["input_image_tier"])
	assert.Equal(t, "1k", other["output_image_tier"])
	assert.InDelta(t, 0.015, other["image_input_price"], 1e-9)
	assert.InDelta(t, 0.12, other["image_output_price"], 1e-9)
	assert.NotContains(t, other, "image_free_first_output", "未开启首张免费时不应出现该字段")
}

// 模型没配按张价格时仍要记录张数（否则无法判断为什么没按张计费）。
func TestAppendImageCountLogInfoWithoutPriceConfig(t *testing.T) {
	setImageCountPriceForTest(t, `{}`)

	relayInfo := newImageCountRelayInfoWithTiers(0, 1, "", "2k")
	other := map[string]interface{}{}
	appendImageCountLogInfo(other, relayInfo, "unknown-model")

	assert.Equal(t, 1, other["output_image_count"])
	assert.NotContains(t, other, "image_output_price")
	assert.NotContains(t, other, "image_free_first_output")
}
