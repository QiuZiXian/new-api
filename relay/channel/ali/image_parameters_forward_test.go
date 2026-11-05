package ali

import (
	"encoding/json"
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// marshalUpstreamParameters 走完整链路（转换 + 序列化）后取出 parameters 对象，
// 这正是真正发给上游的字节，能验证「字段有没有被丢弃」。
func marshalUpstreamParameters(t *testing.T, info *relaycommon.RelayInfo, request dto.ImageRequest) map[string]json.RawMessage {
	t.Helper()
	converted, err := oaiImage2AliImageRequest(info, request, true)
	require.NoError(t, err)

	body, err := common.Marshal(converted)
	require.NoError(t, err)

	var payload struct {
		Parameters map[string]json.RawMessage `json:"parameters"`
	}
	require.NoError(t, json.Unmarshal(body, &payload))
	return payload.Parameters
}

func rawJSON(t *testing.T, s string) json.RawMessage {
	t.Helper()
	require.True(t, json.Valid([]byte(s)), "invalid test JSON: %s", s)
	return json.RawMessage(s)
}

// 手册第 6 节的两个参数必须真正发到上游，而不是在结构体反序列化时被丢掉。
func TestImageParametersForwardDocumentedFields(t *testing.T) {
	info := newImageRelayInfo("")
	request := dto.ImageRequest{
		Model:  "qwen-image-3.0-pro",
		Prompt: "一只戴帽子的猫",
		Extra: map[string]json.RawMessage{
			"parameters": rawJSON(t, `{
				"size": "2048*2048",
				"n": 2,
				"negative_prompt": "低分辨率, 模糊",
				"prompt_extend_mode": "agent",
				"watermark": false
			}`),
		},
	}

	params := marshalUpstreamParameters(t, info, request)

	assert.JSONEq(t, `"低分辨率, 模糊"`, string(params["negative_prompt"]))
	assert.JSONEq(t, `"agent"`, string(params["prompt_extend_mode"]))
	assert.JSONEq(t, `"2048*2048"`, string(params["size"]))
	assert.JSONEq(t, `2`, string(params["n"]))
	// 显式 false 不能被 omitempty 吞掉
	assert.JSONEq(t, `false`, string(params["watermark"]))
}

// 白名单之外的新参数也要原样透传，避免上游以后加字段又被静默吞掉。
func TestImageParametersForwardUnknownFields(t *testing.T) {
	info := newImageRelayInfo("")
	request := dto.ImageRequest{
		Model:  "qwen-image-3.0-pro",
		Prompt: "一只戴帽子的猫",
		Extra: map[string]json.RawMessage{
			"parameters": rawJSON(t, `{
				"n": 1,
				"future_option": {"level": 3},
				"quality_level": "high"
			}`),
		},
	}

	params := marshalUpstreamParameters(t, info, request)

	assert.JSONEq(t, `{"level": 3}`, string(params["future_option"]))
	assert.JSONEq(t, `"high"`, string(params["quality_level"]))
}

// 具名字段优先：Extra 里即便出现同名 key 也不能覆盖已解析的值。
func TestImageParametersKnownFieldsWinOverExtra(t *testing.T) {
	params := AliImageParameters{
		N: 2,
		Extra: map[string]json.RawMessage{
			"n":    rawJSON(t, `99`),
			"seed": rawJSON(t, `42`),
		},
	}

	body, err := common.Marshal(params)
	require.NoError(t, err)

	var decoded map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(body, &decoded))
	assert.JSONEq(t, `2`, string(decoded["n"]), "具名字段必须优先")
	assert.JSONEq(t, `42`, string(decoded["seed"]), "Extra 补上未声明字段")
}

// OpenAI 兼容入口的顶层 negative_prompt 要映射到 parameters 里。
func TestTopLevelNegativePromptMapsIntoParameters(t *testing.T) {
	info := newImageRelayInfo("")
	request := dto.ImageRequest{
		Model:  "qwen-image-3.0-pro",
		Prompt: "一只戴帽子的猫",
		Extra: map[string]json.RawMessage{
			"negative_prompt": rawJSON(t, `"低分辨率"`),
			"parameters":      rawJSON(t, `{"n":1}`),
		},
	}

	params := marshalUpstreamParameters(t, info, request)
	assert.JSONEq(t, `"低分辨率"`, string(params["negative_prompt"]))
}

// parameters 里已经写了 negative_prompt 时，顶层值不能覆盖它。
func TestParametersNegativePromptWinsOverTopLevel(t *testing.T) {
	info := newImageRelayInfo("")
	request := dto.ImageRequest{
		Model:  "qwen-image-3.0-pro",
		Prompt: "一只戴帽子的猫",
		Extra: map[string]json.RawMessage{
			"negative_prompt": rawJSON(t, `"顶层值"`),
			"parameters":      rawJSON(t, `{"negative_prompt":"参数内值"}`),
		},
	}

	params := marshalUpstreamParameters(t, info, request)
	assert.JSONEq(t, `"参数内值"`, string(params["negative_prompt"]))
}

// 空 parameters 仍要输出 {}：京东云这类端点要求该字段存在。
func TestEmptyParametersStillEmitsEmptyObject(t *testing.T) {
	info := newImageRelayInfo(testJDCloudPath)
	request := dto.ImageRequest{
		Model:  "qwen-image-3.0-pro",
		Prompt: "一只戴帽子的猫",
		Size:   "2048x2048",
	}

	converted, err := oaiImage2AliImageRequest(info, request, true)
	require.NoError(t, err)
	body, err := common.Marshal(converted)
	require.NoError(t, err)

	assert.Contains(t, string(body), `"parameters":{}`)
}
