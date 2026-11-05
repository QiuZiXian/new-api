package ali

import (
	"encoding/json"
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// 渠道级「参数覆盖」在 ConvertImageRequest 之后执行，因此可以用
// operations 往阿里格式的 body 里注入 parameters.n，实现固定张数出图。
// 这条用例锁定「配置写法 → 最终请求体」的对应关系。
func TestParamOverrideInjectsParametersN(t *testing.T) {
	info := newImageRelayInfo(testJDCloudPath)
	info.ChannelMeta.ParamOverride = map[string]interface{}{
		"operations": []interface{}{
			map[string]interface{}{
				"path":  "parameters.n",
				"mode":  "set",
				"value": 4,
				"conditions": []interface{}{
					map[string]interface{}{
						"path":  "model",
						"mode":  "prefix",
						"value": "qwen-image",
					},
				},
			},
		},
	}

	request := dto.ImageRequest{
		Model:  "qwen-image-3.0-pro",
		Prompt: "一只在咖啡馆窗边的橘猫",
		Size:   "2048x2048",
	}
	converted, err := oaiImage2AliImageRequest(info, request, true)
	require.NoError(t, err)
	body, err := json.Marshal(converted)
	require.NoError(t, err)

	patched, err := relaycommon.ApplyParamOverrideWithRelayInfo(body, info)
	require.NoError(t, err)
	assert.Equal(t, int64(4), gjson.GetBytes(patched, "parameters.n").Int())
	// 注入 n 不能破坏原有结构。
	assert.Equal(t, "qwen-image-3.0-pro", gjson.GetBytes(patched, "model").String())
	assert.Contains(t, gjson.GetBytes(patched, "input").Raw, "一只在咖啡馆窗边的橘猫")
}

// 条件不满足（同渠道的对话模型）时不能注入，否则会把 chat 请求体写坏。
func TestParamOverrideConditionSkipsNonImageModel(t *testing.T) {
	info := newImageRelayInfo(testJDCloudPath)
	info.ChannelMeta.ParamOverride = map[string]interface{}{
		"operations": []interface{}{
			map[string]interface{}{
				"path":  "parameters.n",
				"mode":  "set",
				"value": 4,
				"conditions": []interface{}{
					map[string]interface{}{
						"path":  "model",
						"mode":  "prefix",
						"value": "qwen-image",
					},
				},
			},
		},
	}

	request := dto.ImageRequest{
		Model:  "qwen-vl-max",
		Prompt: "描述这张图",
		Size:   "1024x1024",
	}
	converted, err := oaiImage2AliImageRequest(info, request, true)
	require.NoError(t, err)
	body, err := json.Marshal(converted)
	require.NoError(t, err)

	patched, err := relaycommon.ApplyParamOverrideWithRelayInfo(body, info)
	require.NoError(t, err)
	assert.False(t, gjson.GetBytes(patched, "parameters.n").Exists())
}
