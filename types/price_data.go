package types

import (
	"fmt"
	"math"
	"sort"

	"github.com/shopspring/decimal"
)

type GroupRatioInfo struct {
	GroupRatio        float64
	GroupSpecialRatio float64
	HasSpecialRatio   bool
}

type PriceData struct {
	FreeModel            bool
	ModelPrice           float64
	ModelRatio           float64
	CompletionRatio      float64
	CacheRatio           float64
	CacheCreationRatio   float64
	CacheCreation5mRatio float64
	CacheCreation1hRatio float64
	ImageRatio           float64
	ImageCompletionRatio float64
	AudioRatio           float64
	AudioCompletionRatio float64
	otherRatios          map[string]float64
	// estimateOnlyRatios 记录只参与预扣估算、不参与终态结算的 OtherRatios key，
	// 具体语义见 MarkOtherRatioEstimateOnly。
	estimateOnlyRatios map[string]struct{}
	UsePrice           bool
	// UseVideoSecond 标记本次按「每秒单价 × 视频时长」计费（视频按秒定价模式），
	// 与 UsePrice 互斥：按秒时不走按次固定价，也不走按量倍率。
	UseVideoSecond bool
	// EstimatedTokens 是适配器按官方公式预估的本次请求 token 用量。
	// 目前用于按 token 计费的视频任务：上游 token 公式为
	// (输入视频时长 + 输出视频时长) × 输出宽 × 输出高 × 输出帧率 / 1024。
	// >0 时用它替代「25 万 token」的默认粗估基数计算预扣额度，
	// 避免冻结额度与终态结算差出一个数量级。
	EstimatedTokens   int
	Quota             int // 按次计费的最终额度（MJ / Task）
	QuotaToPreConsume int // 按量计费的预消耗额度
	GroupRatioInfo    GroupRatioInfo
}

func (p *PriceData) AddOtherRatio(key string, ratio float64) {
	if !isValidOtherRatio(ratio) {
		return
	}
	if p.otherRatios == nil {
		p.otherRatios = make(map[string]float64)
	}
	p.otherRatios[key] = ratio
}

func (p *PriceData) ReplaceOtherRatios(ratios map[string]float64) bool {
	p.otherRatios = nil
	for key, ratio := range ratios {
		p.AddOtherRatio(key, ratio)
	}
	return len(p.otherRatios) > 0
}

func (p *PriceData) HasOtherRatio(key string) bool {
	ratio, ok := p.otherRatios[key]
	return ok && isValidOtherRatio(ratio)
}

func (p *PriceData) OtherRatios() map[string]float64 {
	if len(p.otherRatios) == 0 {
		return nil
	}
	ratios := make(map[string]float64, len(p.otherRatios))
	for key, ratio := range p.otherRatios {
		if isValidOtherRatio(ratio) {
			ratios[key] = ratio
		}
	}
	if len(ratios) == 0 {
		return nil
	}
	return ratios
}

func (p *PriceData) OtherRatioMultiplier() float64 {
	multiplier := 1.0
	for _, ratio := range p.otherRatios {
		if isValidOtherRatio(ratio) && ratio != 1.0 {
			multiplier *= ratio
		}
	}
	return multiplier
}

// MarkOtherRatioEstimateOnly 把某个 OtherRatio 标记为「只用于预扣估算」。
//
// 典型场景：按 token 计费的视频模型，上游回报的 token 数已经包含视频时长
// （官方公式 (输入时长+输出时长)×宽×高×帧率/1024），结算时若再乘 seconds
// 就会把时长重复计费；但预扣阶段还没有 token 数，仍需要 seconds 让冻结额度
// 随请求时长增长。标记后：预扣照旧连乘，终态结算跳过。
func (p *PriceData) MarkOtherRatioEstimateOnly(keys ...string) {
	for _, key := range keys {
		if key == "" {
			continue
		}
		if p.estimateOnlyRatios == nil {
			p.estimateOnlyRatios = make(map[string]struct{})
		}
		p.estimateOnlyRatios[key] = struct{}{}
	}
}

// ReplaceEstimateOnlyRatios 覆盖估算专用 key 集合（用于从任务计费快照恢复）。
func (p *PriceData) ReplaceEstimateOnlyRatios(keys []string) {
	p.estimateOnlyRatios = nil
	p.MarkOtherRatioEstimateOnly(keys...)
}

// EstimateOnlyRatios 返回估算专用 key 列表，顺序稳定，供任务计费快照落库。
func (p *PriceData) EstimateOnlyRatios() []string {
	if len(p.estimateOnlyRatios) == 0 {
		return nil
	}
	keys := make([]string, 0, len(p.estimateOnlyRatios))
	for key := range p.estimateOnlyRatios {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// IsOtherRatioEstimateOnly 判断某个 OtherRatio 是否只参与预扣估算。
func (p *PriceData) IsOtherRatioEstimateOnly(key string) bool {
	_, ok := p.estimateOnlyRatios[key]
	return ok
}

// SettlementOtherRatioMultiplier 返回终态结算应使用的倍率连乘：
// 跳过标记为估算专用的项，其余与预扣一致。
func (p *PriceData) SettlementOtherRatioMultiplier() float64 {
	multiplier := 1.0
	for key, ratio := range p.otherRatios {
		if p.IsOtherRatioEstimateOnly(key) {
			continue
		}
		if isValidOtherRatio(ratio) && ratio != 1.0 {
			multiplier *= ratio
		}
	}
	return multiplier
}

func (p *PriceData) ApplyOtherRatiosToFloat(value float64) float64 {
	return value * p.OtherRatioMultiplier()
}

func (p *PriceData) ApplyOtherRatiosToDecimal(value decimal.Decimal) decimal.Decimal {
	for _, ratio := range p.otherRatios {
		if isValidOtherRatio(ratio) && ratio != 1.0 {
			value = value.Mul(decimal.NewFromFloat(ratio))
		}
	}
	return value
}

func (p *PriceData) RemoveOtherRatiosFromFloat(value float64) float64 {
	for _, ratio := range p.otherRatios {
		if isValidOtherRatio(ratio) && ratio != 1.0 {
			value /= ratio
		}
	}
	return value
}

func isValidOtherRatio(ratio float64) bool {
	return ratio > 0 && !math.IsInf(ratio, 1)
}

func (p *PriceData) ToSetting() string {
	return fmt.Sprintf("ModelPrice: %f, ModelRatio: %f, CompletionRatio: %f, CacheRatio: %f, GroupRatio: %f, UsePrice: %t, CacheCreationRatio: %f, CacheCreation5mRatio: %f, CacheCreation1hRatio: %f, QuotaToPreConsume: %d, ImageRatio: %f, ImageCompletionRatio: %f, AudioRatio: %f, AudioCompletionRatio: %f", p.ModelPrice, p.ModelRatio, p.CompletionRatio, p.CacheRatio, p.GroupRatioInfo.GroupRatio, p.UsePrice, p.CacheCreationRatio, p.CacheCreation5mRatio, p.CacheCreation1hRatio, p.QuotaToPreConsume, p.ImageRatio, p.ImageCompletionRatio, p.AudioRatio, p.AudioCompletionRatio)
}
