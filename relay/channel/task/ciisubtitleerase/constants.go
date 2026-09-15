package ciisubtitleerase

// ModelList 渠道支持的模型。CII 字幕擦除服务是单一模型，本字段用于
// Distribute() 按 model 字段选择渠道。
var ModelList = []string{
	"cii-subtitle-erase",
}

// ChannelName 渠道对外展示名，与 constant.ChannelTypeNames[ChannelTypeCiiSubtitleErase] 保持一致。
var ChannelName = "cii-subtitle-erase"

// defaultSubtitleEraseApiPath 缺省上游 API 路径。TaskApiPath 留空时使用它，
// 集中在常量里以便后续调整。
const defaultSubtitleEraseApiPath = "/api/v1/videos/subtitle-erase/tasks"

// quotaPerSecond 计费基数：第三方视频 1 积分/秒（不足 1 秒按 1 秒）。
// 实际换算走 common.QuotaFromFloat，1 积分即 1 quota 基数。
const quotaPerSecond = 1
