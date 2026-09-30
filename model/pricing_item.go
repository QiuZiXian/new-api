package model

import (
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
)

// 价格展示项（PriceItem）说明
//
// 模型广场详情页的「价格信息」表需要四类信息：分组文案（缓存 tokens / 输出）、
// 维度文案（基础价格 / Fast快速·480p）、单位（M Tokens / 张 / 秒）、价格。
// 前三类的承载体是本文件；第四类永远由计量配置实时推导，不在展示侧二次存储，
// 这样计量配置一改展示价就跟着变，不会出现两处配置漂移。
//
// 本文件只做展示层派生，不参与计费：不读 RelayInfo、不写 PriceData、不改用量。

// BillingRef 取值：把展示行绑定到既有计量配置上。
// 命名与截图里的「模型计费标识」一致，纯展示文案，不参与计费匹配。
const (
	BillingRefInputTokens    = "input_tokens"
	BillingRefOutputTokens   = "output_tokens"
	BillingRefCachedInput    = "cached_input_tokens"
	BillingRefCacheWrite     = "cache_write_tokens"
	BillingRefAudioInput     = "audio_input_tokens"
	BillingRefAudioOutput    = "audio_output_tokens"
	BillingRefPerRequest     = "per_request"
	BillingRefPerImageOutput = "image_output_count"
	BillingRefPerVideoSecond = "video_output_second"
)

// 展示单位取值。
const (
	PriceUnitTokens  = "tokens"
	PriceUnitImage   = "image"
	PriceUnitSecond  = "second"
	PriceUnitRequest = "request"
)

// GroupKey 取值：语义键，前端按 i18n 翻译。
// 管理员自定义分组时改用 GroupLabel，两者可共存（有 Key 优先走 i18n）。
const (
	PriceGroupKeyInputTokens    = "input_tokens"
	PriceGroupKeyOutputTokens   = "output_tokens"
	PriceGroupKeyCachedInput    = "cached_input_tokens"
	PriceGroupKeyCacheWrite     = "cache_write_tokens"
	PriceGroupKeyImageOutput    = "image_output"
	PriceGroupKeyAudioInput     = "audio_input"
	PriceGroupKeyAudioOutput    = "audio_output"
	PriceGroupKeyVideoOutput    = "video_output"
	PriceGroupKeyImageInputToks = "image_input_tokens"
)

// PriceItemConfig 是管理员在「模型（models/metadata）」里编排的展示项，
// 存 models.price_items（JSON 数组）。只描述「怎么展示」，不含价格数值。
type PriceItemConfig struct {
	// ID 即截图里的「模型计费标识」，纯展示文案
	ID string `json:"id"`
	// GroupKey 语义键，前端按 i18n 翻译后显示在「功能」列
	GroupKey string `json:"group_key,omitempty"`
	// GroupLabel 管理员自定义分组文案，填了则优先于 GroupKey
	GroupLabel string `json:"group_label,omitempty"`
	// Variant 显示在「维度」列，如「基础价格」「Fast快速·480p」
	Variant string `json:"variant,omitempty"`
	// BillingRef 计量绑定，决定这一行的价格从哪套配置推导
	BillingRef string `json:"billing_ref"`
	// Unit 单位：tokens / image / second / request
	Unit string `json:"unit,omitempty"`
	// OriginPrice 原价（与 Price 同单位）。>0 且大于现价时才渲染划线价与折扣
	OriginPrice *float64 `json:"origin_price,omitempty"`
	// Enabled 关闭后不渲染该行
	Enabled bool `json:"enabled"`
}

// PriceItem 是派生后下发给 /api/pricing 的展示行。
// Price 是「每单位价格」：tokens 单位为 USD/1M tokens（前端已有 K/M 换算），
// image 为 USD/张，second 为 USD/秒，request 为 USD/次。
type PriceItem struct {
	ID          string   `json:"id"`
	GroupKey    string   `json:"group_key,omitempty"`
	GroupLabel  string   `json:"group_label,omitempty"`
	Variant     string   `json:"variant,omitempty"`
	Unit        string   `json:"unit"`
	Price       float64  `json:"price"`
	OriginPrice *float64 `json:"origin_price,omitempty"`
	// DiscountLabel / SavedPercent 由原价与现价自动算出，不需要管理员维护
	DiscountLabel string `json:"discount_label,omitempty"`
	SavedPercent  string `json:"saved_percent,omitempty"`
}

// parsePriceItemConfigs 解析存储值。JSON 损坏时返回 nil 而不是报错，
// 保证单条模型配置写坏不会拖垮整个定价接口。
func parsePriceItemConfigs(raw string) []PriceItemConfig {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var items []PriceItemConfig
	if err := common.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	out := make([]PriceItemConfig, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.ID) == "" {
			continue
		}
		out = append(out, item)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// defaultPriceItemConfigs 为没有显式配置的模型派生标准 token 行，
// 让存量模型零配置就能渲染出「输入 / 输出 / 缓存 tokens」三行。
func defaultPriceItemConfigs(modelName string, hasCache, hasCacheWrite bool) []PriceItemConfig {
	items := []PriceItemConfig{
		{
			ID:         modelName + "." + BillingRefCachedInput,
			GroupKey:   PriceGroupKeyCachedInput,
			BillingRef: BillingRefCachedInput,
			Unit:       PriceUnitTokens,
			Enabled:    hasCache,
		},
		{
			ID:         modelName + "." + BillingRefInputTokens,
			GroupKey:   PriceGroupKeyInputTokens,
			BillingRef: BillingRefInputTokens,
			Unit:       PriceUnitTokens,
			Enabled:    true,
		},
		{
			ID:         modelName + "." + BillingRefOutputTokens,
			GroupKey:   PriceGroupKeyOutputTokens,
			BillingRef: BillingRefOutputTokens,
			Unit:       PriceUnitTokens,
			Enabled:    true,
		},
	}
	if hasCacheWrite {
		items = append(items, PriceItemConfig{
			ID:         modelName + "." + BillingRefCacheWrite,
			GroupKey:   PriceGroupKeyCacheWrite,
			BillingRef: BillingRefCacheWrite,
			Unit:       PriceUnitTokens,
			Enabled:    true,
		})
	}
	return items
}

// SplitMetaList 把 models 表里的逗号/分号分隔字段切成数组，
// 供前端 capabilities / modalities 直接渲染。
func SplitMetaList(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	parts := strings.FieldsFunc(trimmed, func(r rune) bool {
		return r == ',' || r == ';' || r == '、' || r == '|'
	})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// tokenPricePerMillion 把计量倍率换算成 USD/1M tokens。
// 与前端 lib/price.ts 的 `model_ratio * 2 * ratio` 保持同一换算口径：
// 1 倍率 === $0.002 / 1K tokens === $2 / 1M tokens。
func tokenPricePerMillion(baseRatio, multiplier float64) float64 {
	return baseRatio * 2 * multiplier
}

// PriceRatioSource 汇总派生展示价所需的既有计量配置。
// 字段与 model.Pricing 一一对应，由 updatePricing 填充后传入，这里不做任何读取。
type PriceRatioSource struct {
	Ratio                float64
	CompletionRatio      float64
	CacheRatio           *float64
	CreateCacheRatio     *float64
	AudioRatio           *float64
	AudioCompletionRatio *float64
	ModelPrice           float64
}

// resolveDisplayPrice 依据 BillingRef 从既有计量配置推导展示价。
// 返回 ok=false 表示这一行绑定的计量配置不存在，该行不渲染——
// 这样「配了展示项但没配计费」不会渲染出 0 元这种误导性价格。
func resolveDisplayPrice(ref string, src PriceRatioSource) (float64, bool) {
	switch ref {
	case BillingRefInputTokens:
		return tokenPricePerMillion(src.Ratio, 1), true
	case BillingRefOutputTokens:
		return tokenPricePerMillion(src.Ratio, src.CompletionRatio), true
	case BillingRefCachedInput:
		if src.CacheRatio == nil {
			return 0, false
		}
		return tokenPricePerMillion(src.Ratio, *src.CacheRatio), true
	case BillingRefCacheWrite:
		if src.CreateCacheRatio == nil {
			return 0, false
		}
		return tokenPricePerMillion(src.Ratio, *src.CreateCacheRatio), true
	case BillingRefAudioInput:
		if src.AudioRatio == nil {
			return 0, false
		}
		return tokenPricePerMillion(src.Ratio, *src.AudioRatio), true
	case BillingRefAudioOutput:
		if src.AudioCompletionRatio == nil {
			return 0, false
		}
		return tokenPricePerMillion(src.Ratio, *src.AudioCompletionRatio), true
	case BillingRefPerRequest, BillingRefPerImageOutput:
		// 按次 / 按张共用 model_price：图像按张的计费口径就是 model_price × 张数，
		// 故单张展示价即 model_price。
		if src.ModelPrice <= 0 {
			return 0, false
		}
		return src.ModelPrice, true
	}
	// 按秒（video_output_second）目前没有可推导的计量配置，整行不渲染，
	// 继续回落到前端既有的排版分支，不展示没有依据的价格。
	return 0, false
}

// BuildPriceItems 派生模型的展示行。
//
// ratio/completionRatio/cacheRatio/createCacheRatio/modelPrice 均来自既有计量配置
// （见 updatePricing 的填充逻辑），这里只做单位换算与折扣计算。
//
// deriveDefault 控制「没有显式配置时是否派生标准 token 行」。按次计费的模型
// 没有真实 token 倍率，凭空派生会展示出与实际扣费无关的价格，故只在其显式
// 编排过展示项时才渲染。
func BuildPriceItems(
	modelName string,
	storedPriceItems string,
	deriveDefault bool,
	src PriceRatioSource,
) []PriceItem {
	configs := parsePriceItemConfigs(storedPriceItems)
	if len(configs) == 0 {
		if !deriveDefault {
			return nil
		}
		configs = defaultPriceItemConfigs(modelName, src.CacheRatio != nil, src.CreateCacheRatio != nil)
	}

	items := make([]PriceItem, 0, len(configs))
	for _, cfg := range configs {
		if !cfg.Enabled {
			continue
		}
		price, ok := resolveDisplayPrice(cfg.BillingRef, src)
		if !ok || price <= 0 {
			continue
		}
		unit := cfg.Unit
		if unit == "" {
			unit = PriceUnitTokens
		}
		item := PriceItem{
			ID:         cfg.ID,
			GroupKey:   cfg.GroupKey,
			GroupLabel: cfg.GroupLabel,
			Variant:    cfg.Variant,
			Unit:       unit,
			Price:      price,
		}
		if cfg.OriginPrice != nil && *cfg.OriginPrice > price {
			origin := *cfg.OriginPrice
			item.OriginPrice = &origin
			item.DiscountLabel = fmt.Sprintf("%.1f折", price/origin*10)
			item.SavedPercent = fmt.Sprintf("%.1f%%", (1-price/origin)*100)
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil
	}
	return items
}
