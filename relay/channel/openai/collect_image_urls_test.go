package openai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectOpenAIImageURLs(t *testing.T) {
	body := []byte(`{"created":1,"data":[{"url":"https://example.com/a.png"},{"url":"https://example.com/b.png"},{"b64_json":"AAAA"}]}`)
	assert.Equal(t, []string{
		"https://example.com/a.png",
		"https://example.com/b.png",
	}, collectOpenAIImageURLs(body))

	// 纯 b64 响应没有链接。
	assert.Empty(t, collectOpenAIImageURLs([]byte(`{"data":[{"b64_json":"AAAA"}]}`)))
	// 非数组 / 空响应不 panic。
	assert.Empty(t, collectOpenAIImageURLs([]byte(`{"data":{}}`)))
	assert.Empty(t, collectOpenAIImageURLs(nil))
}
