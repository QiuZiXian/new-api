package common

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetImageResultURLsTrimsAndCaps(t *testing.T) {
	info := &RelayInfo{}

	info.SetImageResultURLs([]string{"  https://example.com/a.png  ", "", "   "})
	assert.Equal(t, []string{"https://example.com/a.png"}, info.ImageResultURLs)

	// 超出条数上限的部分丢弃，避免把消费日志撑大。
	long := make([]string, 0, maxLoggedImageURLs+5)
	for i := 0; i < maxLoggedImageURLs+5; i++ {
		long = append(long, "https://example.com/"+strings.Repeat("x", i+1)+".png")
	}
	info.SetImageResultURLs(long)
	assert.Len(t, info.ImageResultURLs, maxLoggedImageURLs)

	// 单条超长链接截断。
	oversized := strings.Repeat("a", maxLoggedImageURLLength+100)
	info.SetImageResultURLs([]string{oversized})
	assert.Len(t, info.ImageResultURLs[0], maxLoggedImageURLLength)
}

func TestSetImageResultURLsIgnoresEmptyInput(t *testing.T) {
	info := &RelayInfo{}
	info.SetImageResultURLs(nil)
	assert.Empty(t, info.ImageResultURLs)

	// b64_json 响应没有链接，不应写入空切片字段。
	info.SetImageResultURLs([]string{"", ""})
	assert.Empty(t, info.ImageResultURLs)

	// nil 接收者不应 panic。
	var nilInfo *RelayInfo
	nilInfo.SetImageResultURLs([]string{"https://example.com/a.png"})
}
