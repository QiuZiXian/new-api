package ali

import (
	"encoding/json"
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAliBaseURL = "https://modelservice.jdcloud.com"
const testDefaultSyncPath = "/api/v1/services/aigc/multimodal-generation/generation"
const testJDCloudPath = "/v1/services/aigc/multimodal-generation/generation"

func newImageRelayInfo(rawPath string) *relaycommon.RelayInfo {
	return &relaycommon.RelayInfo{
		RelayMode:       constant.RelayModeImagesGenerations,
		OriginModelName: "qwen-image-3.0-pro",
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelBaseUrl:    testAliBaseURL,
			UpstreamModelName: "qwen-image-3.0-pro",
			ChannelOtherSettings: dto.ChannelOtherSettings{
				ImageApiPath: rawPath,
			},
		},
	}
}

func TestAliImageApiPathOverride(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"dashscope default when unset", "", testAliBaseURL + testDefaultSyncPath},
		{"jdcloud path without /api prefix", testJDCloudPath, testAliBaseURL + testJDCloudPath},
		{"missing leading slash is normalized", "v1/services/aigc/multimodal-generation/generation", testAliBaseURL + testJDCloudPath},
		{"path traversal falls back", "/../../v1/secret", testAliBaseURL + testDefaultSyncPath},
		{"full URL falls back", "https://evil.example/v1/x", testAliBaseURL + testDefaultSyncPath},
		{"query string falls back", "/v1/x?key=a", testAliBaseURL + testDefaultSyncPath},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := (&Adaptor{}).GetRequestURL(newImageRelayInfo(tt.raw))
			require.NoError(t, err)
			assert.Equal(t, tt.want, url)
		})
	}
}

func TestAliCustomImagePathSendsEmptyParameters(t *testing.T) {
	info := newImageRelayInfo(testJDCloudPath)
	// 客户端带的 size 等未知字段会进 Extra，但不能被自动映射成 n/size，
	// 否则京东云这类只接受 parameters:{} 的端点会返回「参数解析失败」。
	request := dto.ImageRequest{
		Model:  "qwen-image-3.0-pro",
		Prompt: "一只戴帽子的猫",
		Size:   "2048x2048",
		Extra:  map[string]json.RawMessage{"sequential_image_generation": json.RawMessage(`"auto"`)},
	}

	converted, err := oaiImage2AliImageRequest(info, request, true)
	require.NoError(t, err)
	assert.Equal(t, 0, converted.Parameters.N)
	assert.Empty(t, converted.Parameters.Size)
	require.NotNil(t, converted.Input)
	assert.Equal(t, 1, info.PriceData.OutputImageCount, "预估张数仍要保留给计费")

	// 自定义路径下 parameters 是空的，不能因此把 price.go 按客户端 size
	// 估出的输出档位覆盖成空值。
	info2 := newImageRelayInfo(testJDCloudPath)
	info2.PriceData.SetImageTiers("", "2k")
	_, err = oaiImage2AliImageRequest(info2, request, true)
	require.NoError(t, err)
	assert.Equal(t, "2k", info2.PriceData.OutputImageTier)
}
