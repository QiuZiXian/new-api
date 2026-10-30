package doubao

import (
	"strings"

	"github.com/QuantumNous/new-api/pkg/videopricing"
)

var ModelList = []string{
	"doubao-seedance-1-0-pro-250528",
	"doubao-seedance-1-0-lite-t2v",
	"doubao-seedance-1-0-lite-i2v",
	"doubao-seedance-1-5-pro-251215",
	"doubao-seedance-2-0-260128",
	"doubao-seedance-2-0-fast-260128",
	"doubao-seedance-2-0-mini",
	"doubao-seedance-2.5",
}

var ChannelName = "doubao-video"

// 档位价格表（分辨率档 × 是否含视频 → 元/百万 token）由 pkg/videopricing 统一维护：
// 计费（relay）与模型广场的展示价（model）必须读同一份数据，否则报价与实扣会漂移。

// supportedVideoResolutions 是上游 Seedance 接受的输出分辨率枚举，
// 同时也是价格表的分档键；本地尽早拒绝非法值，避免预扣后才被上游 4xx 打回。
var supportedVideoResolutions = []string{"480p", "720p", "1080p", "4k"}

// supportedVideoRatios 是上游 Seedance 接受的宽高比枚举；
// adaptive / auto 表示由上游跟随输入素材或自行决定。
var supportedVideoRatios = []string{"16:9", "4:3", "1:1", "3:4", "9:16", "21:9", "adaptive", "auto"}

// videoDurationBounds 返回模型家族允许的单次生成时长（秒）。
// 取值来自 CII / 方舟文档：2.5 → [4, 30]，2.0 系列 → [4, 15]，1.5-pro → [4, 12]。
// 未收录的模型返回 false，不额外限制（仍受 MaxTaskDurationSeconds 兜底）。
func videoDurationBounds(modelName string) (minSeconds, maxSeconds int, ok bool) {
	normalized := videopricing.NormalizeModelName(modelName)
	switch {
	case normalized == "doubao-seedance-2-5":
		return 4, 30, true
	case strings.HasPrefix(normalized, "doubao-seedance-2-0"):
		return 4, 15, true
	case normalized == "doubao-seedance-1-5-pro":
		return 4, 12, true
	}
	return 0, 0, false
}
