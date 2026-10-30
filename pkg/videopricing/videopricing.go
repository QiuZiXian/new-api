package videopricing

import "strings"

// Tier 描述视频生成的一个价格档位：输出分辨率档 + 输入是否含视频。
// 零值 Tier{Resolution: "", HasVideo: false} 是模型的基准档（480p/720p 不含视频）。
type Tier struct {
	Resolution string
	HasVideo   bool
}

// tierPrices 各模型在各档位的单价（元/百万 token）。
//
// 基准档（零值 Tier）的单价等于该模型的 model_ratio × 2，是「档位倍率」的分母：
// 计费与展示都用 档位单价 / 基准价 作为倍率，管理员调整 model_ratio 时两边同步跟随。
//
// 重要：以下单价值按公开文档与现有 2.0/2.0-fast 单价结构整理，部署前请通过
// https://www.volcengine.com/docs/82379/1099320 校对一次实际阶梯计费。
var tierPrices = map[string]map[Tier]float64{
	"doubao-seedance-2-0-260128": {
		{}:                                    46.0,
		{HasVideo: true}:                      28.0,
		{Resolution: "1080p"}:                 51.0,
		{Resolution: "1080p", HasVideo: true}: 31.0,
		{Resolution: "4k"}:                    26.0,
		{Resolution: "4k", HasVideo: true}:    16.0,
	},
	"doubao-seedance-2-0-fast-260128": {
		{}:               37.0,
		{HasVideo: true}: 22.0,
	},
	"doubao-seedance-2-0-mini": {
		// mini 仅支持 480p/720p（CII 文档："doubao-seedance-2.0-mini: 默认值 720p;可选值 480p、720p"）
		// 单价值对齐定价表：基准 23 元/百万 token、含视频 14 元/百万 token。
		{}:               23.0,
		{HasVideo: true}: 14.0,
	},
	"doubao-seedance-2.5": {
		// 2.5 不支持 4k（"4k（仅 doubao-seedance-2.0 支持）"），支持到 1080p。
		// 该组以 40 为基准归一（实际销售价 85.75 / 51.45），只有倍率参与计费与展示，
		// 绝对单价由 model_ratio × 2 决定。
		{}:                                    40.0,
		{HasVideo: true}:                      24.0,
		{Resolution: "1080p"}:                 46.0,
		{Resolution: "1080p", HasVideo: true}: 28.0,
	},
	"doubao-seedance-1-5-pro-251215": {
		// 1.5-pro 仅支持 480p/720p
		{}:               30.0,
		{HasVideo: true}: 18.0,
	},
	"doubao-seedance-1-0-pro-250528": {
		// 1.0-pro 仅支持 480p/720p
		{}:               28.0,
		{HasVideo: true}: 16.0,
	},
	"doubao-seedance-1-0-lite-t2v": {
		// 文生视频 lite，老一代入门档
		{}:               15.0,
		{HasVideo: true}: 9.0,
	},
	"doubao-seedance-1-0-lite-i2v": {
		// 图生视频 lite（首帧图），单价比 t2v 略高
		{}:               18.0,
		{HasVideo: true}: 11.0,
	},
}

// tierPricesByModel 是 tierPrices 的归一化索引：表的键是渠道配置的上游模型名，
// 而计费/展示链路拿到的是对外模型名（OriginModelName），需要归一化后才能命中。
var tierPricesByModel = func() map[string]map[Tier]float64 {
	index := make(map[string]map[Tier]float64, len(tierPrices))
	for name, prices := range tierPrices {
		index[NormalizeModelName(name)] = prices
	}
	return index
}()

// NormalizeModelName 把模型名归一化为价格表的比较键，兼容对外名与上游名两种写法：
// "doubao-seedance-2.0" / "doubao-seedance-2-0-260128" 都归一为 "doubao-seedance-2-0"。
// 规则：小写 → 分隔符（. _）统一为 "-" → 去掉末尾 6 位日期版本后缀。
func NormalizeModelName(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.NewReplacer(".", "-", "_", "-").Replace(s)
	parts := strings.Split(s, "-")
	if len(parts) > 1 && isVersionDateSuffix(parts[len(parts)-1]) {
		parts = parts[:len(parts)-1]
	}
	return strings.Join(parts, "-")
}

// isVersionDateSuffix 判断末段是否为版本日期（如 260128），用于从上游模型名还原模型家族。
func isVersionDateSuffix(segment string) bool {
	if len(segment) != 6 {
		return false
	}
	for _, r := range segment {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// lookupTierPrices 按模型名取价格表，先精确匹配再归一化匹配。
func lookupTierPrices(modelName string) (map[Tier]float64, bool) {
	if prices, ok := tierPrices[modelName]; ok {
		return prices, true
	}
	prices, ok := tierPricesByModel[NormalizeModelName(modelName)]
	return prices, ok
}

// tierKey 把分辨率档归一到价格表的档位键：480p 与 720p 同属基准档，
// 只有 1080p / 4k 是独立档位；未知分辨率按基准档处理（上游会自行报错）。
func tierKey(resolution string, hasVideo bool) Tier {
	switch strings.ToLower(strings.TrimSpace(resolution)) {
	case "1080p":
		return Tier{Resolution: "1080p", HasVideo: hasVideo}
	case "4k":
		return Tier{Resolution: "4k", HasVideo: hasVideo}
	}
	return Tier{HasVideo: hasVideo}
}

// GetTierPrice 返回模型在指定档位的单价与基准档单价（元/百万 token）。
// 第二个返回值同时作为「该模型是否配置了价格表」的判据；未配置的组合按基准档返回。
func GetTierPrice(modelName, resolution string, hasVideo bool) (price, basePrice float64, ok bool) {
	prices, ok := lookupTierPrices(modelName)
	if !ok {
		return 0, 0, false
	}
	basePrice = prices[Tier{}]
	if basePrice <= 0 {
		return 0, 0, false
	}
	if tierPrice, found := prices[tierKey(resolution, hasVideo)]; found {
		return tierPrice, basePrice, true
	}
	// 未配置的组合（如 fast 无 1080p/4k，上游会自行报错）按基准价计费即可。
	return basePrice, basePrice, true
}

// GetTierRatio 返回模型在指定档位相对基准档的倍率。
// 第二个返回值表示该模型是否配置了价格表；倍率为 1.0 时调用方可忽略该倍率。
func GetTierRatio(modelName, resolution string, hasVideo bool) (float64, bool) {
	price, basePrice, ok := GetTierPrice(modelName, resolution, hasVideo)
	if !ok {
		return 0, false
	}
	return price / basePrice, true
}
