package ali

import (
	"testing"

	"github.com/QuantumNous/new-api/relay/constant"

	"github.com/stretchr/testify/assert"
)

// 请求体透传会跳过 ConvertImageRequest，而 IsSyncImageModel 原本只在其中赋值。
// 若不在 Init 里补上，同步图片模型会走异步轮询分支，拿不到 task_id 就报
// 「channel error (status code: 200)」且消息为空。
func TestInitDetectsSyncImageModelWithoutConvertRequest(t *testing.T) {
	info := newImageRelayInfo(testJDCloudPath)
	adaptor := &Adaptor{}

	adaptor.Init(info)

	assert.True(t, adaptor.IsSyncImageModel, "透传模式下也要能识别同步图片模型")
}

func TestInitKeepsWanEditAsync(t *testing.T) {
	info := newImageRelayInfo(testJDCloudPath)
	info.RelayMode = constant.RelayModeImagesEdits
	info.ChannelMeta.UpstreamModelName = "wan2.6-image-edit"
	adaptor := &Adaptor{}

	adaptor.Init(info)

	assert.False(t, adaptor.IsSyncImageModel, "wan 图生图走异步端点，不能按同步解析")
}

func TestInitIgnoresNonImageModel(t *testing.T) {
	info := newImageRelayInfo(testJDCloudPath)
	info.ChannelMeta.UpstreamModelName = "qwen-max"
	adaptor := &Adaptor{}

	adaptor.Init(info)

	assert.False(t, adaptor.IsSyncImageModel)
}
