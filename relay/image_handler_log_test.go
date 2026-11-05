package relay

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// 上游真实张数优先：客户端声明 1 张、实际出 2 张时日志必须显示 2。
func TestResolveImageLogCountPrefersUpstreamCount(t *testing.T) {
	assert.Equal(t, uint(2), resolveImageLogCount(1, 2))
	assert.Equal(t, uint(4), resolveImageLogCount(4, 4))
	// 上游没报张数时回退客户端声明值，不能把日志抹成 0
	assert.Equal(t, uint(3), resolveImageLogCount(3, 0))
	assert.Equal(t, uint(1), resolveImageLogCount(1, -1))
}

// 上游实际出图尺寸优先：透传模式下 request.Size 为空也能显示真实尺寸。
func TestResolveImageLogSizePrefersResultSize(t *testing.T) {
	assert.Equal(t, "1792x2400", resolveImageLogSize("", "1792x2400"))
	assert.Equal(t, "1792x2400", resolveImageLogSize("2048x2048", "1792x2400"))
	assert.Equal(t, "2048x2048", resolveImageLogSize("2048x2048", ""))
	assert.Equal(t, "", resolveImageLogSize("", ""))
}

// 日志摘要：品质只在客户端显式指定时出现，不再伪造 standard。
func TestBuildImageLogContentOmitsDefaultQuality(t *testing.T) {
	assert.Equal(t, []string{"大小 1792x2400", "生成数量 2"},
		buildImageLogContent("1792x2400", "", 2))

	assert.Equal(t, []string{"大小 1024x1024", "品质 hd", "生成数量 1"},
		buildImageLogContent("1024x1024", " hd ", 1))
}

// 三者都缺失时不应产出空条目（空字符串会被拼成 "大小 " 这类噪声）。
func TestBuildImageLogContentSkipsEmptyFields(t *testing.T) {
	assert.Empty(t, buildImageLogContent("", "", 0))
	assert.Equal(t, []string{"生成数量 1"}, buildImageLogContent("", "", 1))
}
