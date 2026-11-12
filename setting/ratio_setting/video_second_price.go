package ratio_setting

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/types"
)

// VideoSecondDefaultVariant 是按秒计费的兜底档位，同时作为分档倍率的基准价。
// 管理员为每个视频模型配一组「分辨率档 -> 每秒单价」，其中 default 档用于
// 请求未指定分辨率（或档位未命中）时计价。
const VideoSecondDefaultVariant = "default"

// maxVideoSecondUnitPrice 限制每秒单价，避免管理员误填超大值把配额乘法推到饱和区间。
const maxVideoSecondUnitPrice = 1_000_000

// videoSecondPriceMap 按秒计费价格表：模型 -> 分辨率档 -> 每秒单价。
// 单价口径与 ModelPrice 一致（即「每 QuotaPerUnit 对应的价格」），
// 这里的一次等于一秒，任务链路再乘以视频时长得到最终额度。
var videoSecondPriceMap = types.NewRWMap[string, map[string]float64]()

// normalizeVideoSecondVariant 统一档位名。管理员在后台可能填 "720P"，
// 而适配器/上游回报的是 "720p"；不做归一化会让分档价静默退化成 default 档。
func normalizeVideoSecondVariant(variant string) string {
	return strings.ToLower(strings.TrimSpace(variant))
}

// lookupVideoSecondVariant 先精确查规范化后的档位名，再兼容历史配置里
// 大小写/空格不统一的档位名，保证旧数据不会因为一次归一化改动而失价。
func lookupVideoSecondVariant(variants map[string]float64, variant string) (float64, bool) {
	key := normalizeVideoSecondVariant(variant)
	if price, ok := variants[key]; ok {
		return price, true
	}
	for name, price := range variants {
		if normalizeVideoSecondVariant(name) == key {
			return price, true
		}
	}
	return 0, false
}

// parseVideoSecondPriceValue 兼容数值与字符串两种写法。
// 早期前端把单价按字符串写进 JSON，而 Go 的 map[string]float64 会直接拒绝，
// 导致整张表加载失败（配置看似保存成功、实际按秒计费完全不生效）。
func parseVideoSecondPriceValue(raw any) (float64, error) {
	switch v := raw.(type) {
	case float64:
		return v, nil
	case string:
		price, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return 0, fmt.Errorf("%q is not a number", v)
		}
		return price, nil
	case nil:
		return 0, fmt.Errorf("price is null")
	default:
		return 0, fmt.Errorf("price has unsupported type %T", raw)
	}
}

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
	if price, ok := lookupVideoSecondVariant(variants, VideoSecondDefaultVariant); ok && price > 0 {
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
	if price, ok := lookupVideoSecondVariant(variants, variant); ok && price > 0 {
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

// UpdateVideoSecondPriceByJSONString 校验并加载整张按秒价格表。
// 校验必须先于替换内存表完成：否则一条脏数据会把已生效的其它模型价格一起清空。
func UpdateVideoSecondPriceByJSONString(jsonStr string) error {
	var parsed map[string]map[string]any
	if err := common.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return fmt.Errorf("invalid video second price JSON: %w", err)
	}

	normalized := make(map[string]map[string]float64, len(parsed))
	for model, variants := range parsed {
		if strings.TrimSpace(model) == "" {
			return fmt.Errorf("invalid video second price: empty model name")
		}
		tiers := make(map[string]float64, len(variants))
		hasPositive := false
		for variant, raw := range variants {
			key := normalizeVideoSecondVariant(variant)
			if key == "" {
				return fmt.Errorf("invalid video second price for model %q: empty tier name", model)
			}
			price, err := parseVideoSecondPriceValue(raw)
			if err != nil {
				return fmt.Errorf("invalid video second price for model %q tier %q: %w", model, variant, err)
			}
			if price < 0 || price > maxVideoSecondUnitPrice || math.IsNaN(price) || math.IsInf(price, 0) {
				return fmt.Errorf("invalid video second price for model %q tier %q: price must be a finite number in [0, %d]", model, variant, maxVideoSecondUnitPrice)
			}
			// 大小写不同的同档名归一化后合并，保留较高价，避免低档价覆盖高档价。
			if existing, ok := tiers[key]; ok && existing > price {
				price = existing
			}
			tiers[key] = price
			if price > 0 {
				hasPositive = true
			}
		}
		if len(tiers) == 0 {
			continue
		}
		if !hasPositive {
			return fmt.Errorf("invalid video second price for model %q: at least one positive tier price is required", model)
		}
		normalized[model] = tiers
	}

	payload, err := common.Marshal(normalized)
	if err != nil {
		return fmt.Errorf("marshal video second price: %w", err)
	}
	return types.LoadFromJsonStringWithCallback(videoSecondPriceMap, string(payload), InvalidateExposedDataCache)
}

func GetVideoSecondPriceCopy() map[string]map[string]float64 {
	return videoSecondPriceMap.ReadAll()
}
