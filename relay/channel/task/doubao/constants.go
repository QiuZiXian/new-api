package doubao

import "strings"

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

// videoPriceKey 价格表的键：输出分辨率档（is1080p/is4k 均为 false 即 480p/720p 基准档）、输入是否含视频。
type videoPriceKey struct {
	is1080p  bool
	is4k     bool
	hasVideo bool
}

// videoPriceTable 各模型在不同 (输出分辨率档, 是否含视频输入) 下的单价（元/百万 token）。
// 其中零值键 {480p/720p, 不含视频} 为基准价，等于管理员应配置的 ModelRatio；
// 计费时取 实际单价/基准价 作为 OtherRatio。
//
// 重要：以下各模型单价值为按公开文档与现有 2.0/2.0-fast 单价结构的合理推断，
// 部署前请通过 https://www.volcengine.com/docs/82379/1099320 校对一次实际阶梯计费，
// 错误的单价值会导致持续少扣或多扣。
var videoPriceTable = map[string]map[videoPriceKey]float64{
	"doubao-seedance-2-0-260128": {
		{hasVideo: false}:                46.0,
		{hasVideo: true}:                 28.0,
		{is1080p: true, hasVideo: false}: 51.0,
		{is1080p: true, hasVideo: true}:  31.0,
		{is4k: true, hasVideo: false}:    26.0,
		{is4k: true, hasVideo: true}:     16.0,
	},
	"doubao-seedance-2-0-fast-260128": {
		{hasVideo: false}: 37.0,
		{hasVideo: true}:  22.0,
	},
	"doubao-seedance-2-0-mini": {
		// mini 仅支持 480p/720p（CII 文档："doubao-seedance-2.0-mini: 默认值 720p;可选值 480p、720p"）
		{hasVideo: false}: 28.0,
		{hasVideo: true}:  16.0,
	},
	"doubao-seedance-2.5": {
		// 2.5 不支持 4k（"4k（仅 doubao-seedance-2.0 支持）"），支持到 1080p
		{hasVideo: false}:                40.0,
		{hasVideo: true}:                 24.0,
		{is1080p: true, hasVideo: false}: 46.0,
		{is1080p: true, hasVideo: true}:  28.0,
	},
	"doubao-seedance-1-5-pro-251215": {
		// 1.5-pro 仅支持 480p/720p
		{hasVideo: false}: 30.0,
		{hasVideo: true}:  18.0,
	},
	"doubao-seedance-1-0-pro-250528": {
		// 1.0-pro 仅支持 480p/720p
		{hasVideo: false}: 28.0,
		{hasVideo: true}:  16.0,
	},
	"doubao-seedance-1-0-lite-t2v": {
		// 文生视频 lite，老一代入门档
		{hasVideo: false}: 15.0,
		{hasVideo: true}:  9.0,
	},
	"doubao-seedance-1-0-lite-i2v": {
		// 图生视频 lite（首帧图），单价比 t2v 略高
		{hasVideo: false}: 18.0,
		{hasVideo: true}:  11.0,
	},
}

// GetVideoInputRatio 返回指定模型在给定输出分辨率/是否含视频输入下，相对基准价的计费倍率。
// 第二个返回值表示该模型是否配置了价格表；倍率为 1.0 时调用方可忽略该 OtherRatio。
func GetVideoInputRatio(modelName, resolution string, hasVideo bool) (float64, bool) {
	prices, ok := videoPriceTable[modelName]
	base := prices[videoPriceKey{}] // 零值键 = {480p/720p, 不含视频} 基准价
	if !ok || base <= 0 {
		return 0, false
	}
	res := strings.ToLower(strings.TrimSpace(resolution))
	price, ok := prices[videoPriceKey{is1080p: res == "1080p", is4k: res == "4k", hasVideo: hasVideo}]
	if !ok {
		// 未配置的组合（如 fast 无 1080p/4k，上游会自行报错）按基准价计费即可。
		return 1.0, true
	}
	return price / base, true
}
