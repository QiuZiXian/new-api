package ratio_setting

import (
	"strings"

	"github.com/QuantumNous/new-api/types"
)

// VideoSecondDefaultVariant 是按秒计费的兜底档位，同时作为分档倍率的基准价。
// 管理员为每个视频模型配一组「分辨率档 -> 每秒单价」，其中 default 档用于
// 请求未指定分辨率（或档位未命中）时计价。
const VideoSecondDefaultVariant = "default"

// videoSecondPriceMap 按秒计费价格表：模型 -> 分辨率档 -> 每秒单价。
// 单价口径与 ModelPrice 一致（即「每 QuotaPerUnit 对应的价格」），
// 这里的一次等于一秒，任务链路再乘以视频时长得到最终额度。
var videoSecondPriceMap = types.NewRWMap[string, map[string]float64]()

// GetVideoSecondPrice 返回模型按分辨率分档的每秒单价。ok=false 表示该模型未配置
// 按秒计费，调用方应继续走按次 / 按量的既有逻辑。
func GetVideoSecondPrice(model string) (map[string]float64, bool) {
	variants, ok := videoSecondPriceMap.Get(model)
	if !ok || len(variants) == 0 {
		return nil, false
	}
	return variants, true
}

// GetVideoSecondBasePrice 返回按秒计费的基准单价：优先 default 档，缺失时取该模型
// 已配置档位中的最低价，保证总能算出可预扣的额度。
func GetVideoSecondBasePrice(model string) (float64, bool) {
	variants, ok := GetVideoSecondPrice(model)
	if !ok {
		return 0, false
	}
	if price, ok := variants[VideoSecondDefaultVariant]; ok && price > 0 {
		return price, true
	}
	var min float64
	for _, price := range variants {
		if price > 0 && (min == 0 || price < min) {
			min = price
		}
	}
	if min <= 0 {
		return 0, false
	}
	return min, true
}

// GetVideoSecondVariantPrice 取指定档位的每秒单价，未配置该档时回落到基准价。
func GetVideoSecondVariantPrice(model, variant string) (float64, bool) {
	variants, ok := GetVideoSecondPrice(model)
	if !ok {
		return 0, false
	}
	variant = strings.ToLower(strings.TrimSpace(variant))
	if price, ok := variants[variant]; ok && price > 0 {
		return price, true
	}
	return GetVideoSecondBasePrice(model)
}

// GetVideoSecondVariantRatio 返回档位价相对基准价的倍率，供任务链路以 OtherRatio
// 形式参与连乘。倍率为 1（即命中基准档）时返回 false，调用方跳过即可。
func GetVideoSecondVariantRatio(model, variant string) (float64, bool) {
	base, ok := GetVideoSecondBasePrice(model)
	if !ok || base <= 0 {
		return 0, false
	}
	price, ok := GetVideoSecondVariantPrice(model, variant)
	if !ok || price <= 0 {
		return 0, false
	}
	ratio := price / base
	if ratio <= 0 || ratio == 1 {
		return 0, false
	}
	return ratio, true
}

func VideoSecondPrice2JSONString() string {
	return videoSecondPriceMap.MarshalJSONString()
}

func UpdateVideoSecondPriceByJSONString(jsonStr string) error {
	return types.LoadFromJsonStringWithCallback(videoSecondPriceMap, jsonStr, InvalidateExposedDataCache)
}

func GetVideoSecondPriceCopy() map[string]map[string]float64 {
	return videoSecondPriceMap.ReadAll()
}
