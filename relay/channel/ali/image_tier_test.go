package ali

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveImageTiersPrefersOutputArea(t *testing.T) {
	// 1K：面积正好 2,250,000
	usage := AliUsage{OutputWidth: 1500, OutputHeight: 1500, OutputImageType: "qima_output_2k"}
	inputTier, outputTier := resolveImageTiers(usage, "", "")
	assert.Equal(t, "1k", outputTier, "宽高面积优先于 image_type 标签")
	assert.Equal(t, "", inputTier)

	// 2K：面积 2,251,500
	usage = AliUsage{OutputWidth: 1500, OutputHeight: 1501}
	_, outputTier = resolveImageTiers(usage, "", "")
	assert.Equal(t, "2k", outputTier)
}

func TestResolveImageTiersFallsBackToLabels(t *testing.T) {
	// 没有宽高时读 output_image_type / input_image_type 的档位后缀。
	usage := AliUsage{
		InputImageType:  "qima_input_2k",
		OutputImageType: "qima_output_1k",
	}
	inputTier, outputTier := resolveImageTiers(usage, "", "")
	assert.Equal(t, "2k", inputTier)
	assert.Equal(t, "1k", outputTier)
}

func TestResolveImageTiersKeepsRequestEstimate(t *testing.T) {
	// 上游完全不报档位信息时，保留请求 size 估算出的输出档位。
	inputTier, outputTier := resolveImageTiers(AliUsage{}, "", "2k")
	assert.Equal(t, "", inputTier)
	assert.Equal(t, "2k", outputTier)

	// 无法识别的标签不能让请求侧估算失效。
	inputTier, outputTier = resolveImageTiers(
		AliUsage{OutputImageType: "qima_output_4k"},
		"1k",
		"2k",
	)
	assert.Equal(t, "1k", inputTier)
	assert.Equal(t, "2k", outputTier)
}
