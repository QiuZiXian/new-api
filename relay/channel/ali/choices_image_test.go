package ali

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

// decodeAliOutput 用真实上游响应体构造 AliOutput，避免手搓匿名结构体出错。
func decodeAliOutput(t *testing.T, body string) *AliOutput {
	t.Helper()
	var out AliOutput
	require.NoError(t, json.Unmarshal([]byte(body), &out))
	return &out
}

// 上游在一条 message 里返回多张图时必须逐张转出，不能相互覆盖。
func TestChoicesToOpenAIImageDateKeepsEveryImage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	output := decodeAliOutput(t, `{
		"choices": [{
			"message": {
				"role": "assistant",
				"content": [
					{"type": "image", "image": "https://example.com/output_0.png"},
					{"type": "image", "image": "https://example.com/output_1.png"}
				]
			},
			"finish_reason": "stop"
		}]
	}`)

	data := output.ChoicesToOpenAIImageDate(c, "url")
	require.Len(t, data, 2, "n=2 必须返回两条结果")
	assert.Equal(t, "https://example.com/output_0.png", data[0].Url, "顺序要与上游一致")
	assert.Equal(t, "https://example.com/output_1.png", data[1].Url)
}

// 文本内容作为同一 choice 内所有图片的 revised_prompt，图片顺序不变。
func TestChoicesToOpenAIImageDateAppliesRevisedPromptToAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	output := decodeAliOutput(t, `{
		"choices": [{
			"message": {
				"role": "assistant",
				"content": [
					{"type": "text", "text": "一只戴帽子的猫"},
					{"type": "image", "image": "https://example.com/a.png"},
					{"type": "image", "image": "https://example.com/b.png"}
				]
			}
		}]
	}`)

	data := output.ChoicesToOpenAIImageDate(c, "url")
	require.Len(t, data, 2)
	for _, item := range data {
		assert.Equal(t, "一只戴帽子的猫", item.RevisedPrompt)
	}
	assert.Equal(t, "https://example.com/a.png", data[0].Url)
	assert.Equal(t, "https://example.com/b.png", data[1].Url)
}

// 多个 choice 各自展开，结果按 choice 顺序拼接。
func TestChoicesToOpenAIImageDateHandlesMultipleChoices(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	output := decodeAliOutput(t, `{
		"choices": [
			{"message": {"content": [{"image": "https://example.com/1.png"}]}},
			{"message": {"content": [{"image": "https://example.com/2.png"}, {"image": "https://example.com/3.png"}]}}
		]
	}`)

	data := output.ChoicesToOpenAIImageDate(c, "url")
	require.Len(t, data, 3)
	assert.Equal(t, []string{
		"https://example.com/1.png",
		"https://example.com/2.png",
		"https://example.com/3.png",
	}, []string{data[0].Url, data[1].Url, data[2].Url})
}

// 只有文本、没有图片的 choice 不产出空结果，避免污染 data 数组。
func TestChoicesToOpenAIImageDateSkipsChoiceWithoutImage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	output := decodeAliOutput(t, `{
		"choices": [{"message": {"content": [{"text": "only text"}]}}]
	}`)

	assert.Empty(t, output.ChoicesToOpenAIImageDate(c, "url"))
}

// 非 http 前缀视为 base64 内联图，不当作 URL。
func TestChoicesToOpenAIImageDateHandlesInlineBase64(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	output := decodeAliOutput(t, `{
		"choices": [{"message": {"content": [{"image": "iVBORw0KGgoAAA=="}]}}]
	}`)

	data := output.ChoicesToOpenAIImageDate(c, "url")
	require.Len(t, data, 1)
	assert.Empty(t, data[0].Url)
	assert.Equal(t, "iVBORw0KGgoAAA==", data[0].B64Json)
}

// 端到端断言最终序列化的 JSON：data 数组必须与上游图片数一致。
func TestResponseAli2OpenAIImageSerializesAllImages(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := []byte(`{
		"output": {"choices": [{"message": {"content": [
			{"type": "image", "image": "https://example.com/output_0.png"},
			{"type": "image", "image": "https://example.com/output_1.png"}
		]}}]},
		"usage": {"output_image_count": 2, "output_width": 1792, "output_height": 2400}
	}`)
	var aliResp AliResponse
	require.NoError(t, json.Unmarshal(body, &aliResp))
	require.Equal(t, 2, aliResp.Usage.OutputImageCount)

	info := &relaycommon.RelayInfo{StartTime: time.Now()}
	imageResponse := responseAli2OpenAIImage(c, &aliResp, body, info, "url")
	require.NotNil(t, imageResponse)
	require.Len(t, imageResponse.Data, 2)

	raw, err := json.Marshal(imageResponse)
	require.NoError(t, err)
	var parsed struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(raw, &parsed))
	require.Len(t, parsed.Data, 2, "响应体里应有两条图片")
	assert.Equal(t, "https://example.com/output_0.png", parsed.Data[0].URL)
	assert.Equal(t, "https://example.com/output_1.png", parsed.Data[1].URL)
}
