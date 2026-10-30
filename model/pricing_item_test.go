package model

import (
	"math"
	"testing"
)

// 换算口径与前端 web/src/features/pricing/lib/price.ts 保持一致：
//
//	base   = model_ratio * 2 * groupRatio        // USD / 1M tokens
//	input  = base
//	output = base * completion_ratio
//	cache  = base * cache_ratio
//
// 本测试在 groupRatio = 1 的前提下核对后端派生值。
func TestBuildPriceItemsDefaults(t *testing.T) {
	cacheRatio := 0.25
	items := BuildPriceItems("glm-5.3", "", true, PriceRatioSource{
		Ratio: 1, CompletionRatio: 4, CacheRatio: &cacheRatio,
	})
	if len(items) != 3 {
		t.Fatalf("want 3 default items, got %d", len(items))
	}

	want := map[string]float64{
		"glm-5.3.cached_input_tokens": 0.5, // 1 * 2 * 0.25
		"glm-5.3.input_tokens":        2,   // 1 * 2
		"glm-5.3.output_tokens":       8,   // 1 * 2 * 4
	}
	for _, item := range items {
		if got, ok := want[item.ID]; !ok {
			t.Fatalf("unexpected item id %q", item.ID)
		} else if item.Price != got {
			t.Errorf("item %s: want price %v, got %v", item.ID, got, item.Price)
		}
		if item.Unit != PriceUnitTokens {
			t.Errorf("item %s: want unit %q, got %q", item.ID, PriceUnitTokens, item.Unit)
		}
		if item.DiscountLabel != "" || item.SavedPercent != "" {
			t.Errorf("item %s: no origin price configured, discount must stay empty", item.ID)
		}
	}
}

// 缓存倍率没配时不应渲染出缓存行（否则会展示误导性的 0 元或全额价）。
func TestBuildPriceItemsSkipsUnconfiguredCache(t *testing.T) {
	items := BuildPriceItems("m", "", true, PriceRatioSource{
		Ratio: 1, CompletionRatio: 4,
	})
	for _, item := range items {
		if item.GroupKey == PriceGroupKeyCachedInput {
			t.Fatalf("cache row must be skipped when cache_ratio is unset")
		}
	}
	if len(items) != 2 {
		t.Fatalf("want 2 items, got %d", len(items))
	}
}

// 管理员编排：多行同一分组即截图的「视频多档」排版；价格仍由计量配置推导。
func TestBuildPriceItemsFromStoredConfig(t *testing.T) {
	origin := 2.0
	stored := `[{"id":"seedream.image","group_key":"image_output","variant":"基础价格",` +
		`"billing_ref":"output_tokens","unit":"tokens","origin_price":2,"enabled":true},` +
		`{"id":"seedream.image.off","billing_ref":"output_tokens","enabled":false}]`
	items := BuildPriceItems("seedream", stored, false, PriceRatioSource{
		Ratio: 1, CompletionRatio: 0.35,
	})
	if len(items) != 1 {
		t.Fatalf("disabled row must be filtered out, got %d", len(items))
	}
	item := items[0]
	if item.Price != 0.7 { // 1 * 2 * 0.35
		t.Errorf("want price 0.7, got %v", item.Price)
	}
	if item.OriginPrice == nil || *item.OriginPrice != origin {
		t.Fatalf("want origin price 2, got %v", item.OriginPrice)
	}
	if item.DiscountLabel != "3.5折" {
		t.Errorf("want discount 3.5折, got %q", item.DiscountLabel)
	}
	if item.SavedPercent != "65.0%" {
		t.Errorf("want saved 65.0%%, got %q", item.SavedPercent)
	}
}

// 展示项绑定的计量配置不存在时，该行不渲染，避免展示 0 元。
// 按秒（video_output_second）目前没有可推导的计量配置，同样整行跳过。
func TestBuildPriceItemsSkipsUnresolvedRef(t *testing.T) {
	stored := `[{"id":"x.per_request","billing_ref":"per_request","unit":"request","enabled":true},` +
		`{"id":"x.video","billing_ref":"video_output_second","unit":"second","enabled":true}]`
	items := BuildPriceItems("x", stored, false, PriceRatioSource{})
	if len(items) != 0 {
		t.Fatalf("unresolved billing_ref must be skipped, got %d", len(items))
	}
}

// 按次计费且没编排展示项时不应派生 token 行（会展示与实际扣费无关的价格）。
func TestBuildPriceItemsNoDefaultForPerRequest(t *testing.T) {
	if items := BuildPriceItems("m", "", false, PriceRatioSource{
		Ratio: 1, CompletionRatio: 4, ModelPrice: 0.5,
	}); len(items) != 0 {
		t.Fatalf("want no items, got %d", len(items))
	}
	// 显式编排后仍要渲染，图像按张的价格从 model_price 推导
	stored := `[{"id":"m.image","group_key":"image_output","billing_ref":"image_output_count","unit":"image","enabled":true}]`
	items := BuildPriceItems("m", stored, false, PriceRatioSource{
		Ratio: 1, CompletionRatio: 4, ModelPrice: 0.5,
	})
	if len(items) != 1 || items[0].Price != 0.5 || items[0].Unit != PriceUnitImage {
		t.Fatalf("want one per-image item at 0.5, got %+v", items)
	}
}

// 音频行从既有 AudioRatio / AudioCompletionRatio 推导。
func TestBuildPriceItemsAudioRefs(t *testing.T) {
	audioIn, audioOut := 0.5, 1.5
	stored := `[{"id":"m.audio_in","billing_ref":"audio_input_tokens","unit":"tokens","enabled":true},` +
		`{"id":"m.audio_out","billing_ref":"audio_output_tokens","unit":"tokens","enabled":true}]`
	items := BuildPriceItems("m", stored, false, PriceRatioSource{
		Ratio: 1, AudioRatio: &audioIn, AudioCompletionRatio: &audioOut,
	})
	if len(items) != 2 {
		t.Fatalf("want 2 audio items, got %d", len(items))
	}
	if items[0].Price != 1 || items[1].Price != 3 { // 1*2*0.5, 1*2*1.5
		t.Fatalf("want prices 1 / 3, got %v / %v", items[0].Price, items[1].Price)
	}
}

// 视频按 token 计费的分档展示：同一模型六个档位必须算出不同价格。
// 存量配置把档位写在「维度」文案里，ref 仍绑 input_tokens / output_tokens。
func TestBuildPriceItemsVideoTierFromVariant(t *testing.T) {
	stored := `[{"id":"v.720p","group_key":"input_tokens","variant":"480p/720p不含视频","billing_ref":"input_tokens","unit":"tokens","enabled":true},` +
		`{"id":"v.720p_video","group_key":"input_tokens","variant":"480p/720p含视频","billing_ref":"input_tokens","unit":"tokens","enabled":true},` +
		`{"id":"v.1080p","group_key":"input_tokens","variant":"1080p不含视频","billing_ref":"input_tokens","unit":"tokens","origin_price":51,"enabled":true},` +
		`{"id":"v.1080p_video","group_key":"output_tokens","variant":"1080p含视频","billing_ref":"output_tokens","unit":"tokens","enabled":true},` +
		`{"id":"v.4k","group_key":"output_tokens","variant":"4K不含视频","billing_ref":"output_tokens","unit":"tokens","enabled":true},` +
		`{"id":"v.4k_video","group_key":"output_tokens","variant":"4K含视频","billing_ref":"output_tokens","unit":"tokens","enabled":true}]`
	items := BuildPriceItems("doubao-seedance-2.0", stored, false, PriceRatioSource{
		Ratio: 23, CompletionRatio: 1,
	})
	if len(items) != 6 {
		t.Fatalf("want 6 tier items, got %d", len(items))
	}
	// 与 doubao 计费档位一致：46 / 28 / 51 / 31 / 26 / 16（元/百万 token）
	want := []float64{46, 28, 51, 31, 26, 16}
	for i, item := range items {
		if math.Abs(item.Price-want[i]) > 1e-9 {
			t.Errorf("item %s: want price %v, got %v", item.ID, want[i], item.Price)
		}
	}
	// 1080p 档价已是 51，与原价相同 → 不再渲染划线价
	if items[2].OriginPrice != nil || items[2].DiscountLabel != "" {
		t.Errorf("1080p tier: origin price equals price, discount must be empty")
	}
}

// 显式档位后缀优先于文案推断，未配置的档位按基准价展示。
func TestBuildPriceItemsVideoTierFromBillingRef(t *testing.T) {
	stored := `[{"id":"v.base","billing_ref":"video_input_tokens","unit":"tokens","enabled":true},` +
		`{"id":"v.4k_video","billing_ref":"video_input_tokens:4k+video","unit":"tokens","enabled":true},` +
		`{"id":"v.fast1080","billing_ref":"video_input_tokens:1080p","unit":"tokens","enabled":true}]`
	items := BuildPriceItems("doubao-seedance-2.0-fast", stored, false, PriceRatioSource{
		Ratio: 18.5, CompletionRatio: 1,
	})
	if len(items) != 3 {
		t.Fatalf("want 3 items, got %d", len(items))
	}
	if items[0].Price != 37 {
		t.Errorf("base tier: want 37, got %v", items[0].Price)
	}
	if items[1].Price != 37 {
		t.Errorf("4k unsupported by fast → base price, got %v", items[1].Price)
	}
	if items[2].Price != 37 {
		t.Errorf("1080p unsupported by fast → base price, got %v", items[2].Price)
	}
}

// 非视频模型不受档位逻辑影响：没有档位价表时仍按 token 倍率换算。
func TestBuildPriceItemsVideoTierIgnoredForOtherModels(t *testing.T) {
	stored := `[{"id":"m.4k","billing_ref":"input_tokens","variant":"4K 高清","unit":"tokens","enabled":true}]`
	items := BuildPriceItems("some-llm", stored, false, PriceRatioSource{
		Ratio: 1, CompletionRatio: 1,
	})
	if len(items) != 1 || items[0].Price != 2 {
		t.Fatalf("want one item at 2, got %+v", items)
	}
}

func TestSplitMetaList(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"vision", 1},
		{"vision, tools", 2},
		{"text;image、video|audio", 4},
	}
	for _, c := range cases {
		got := SplitMetaList(c.in)
		if len(got) != c.want {
			t.Errorf("SplitMetaList(%q): want %d, got %d (%v)", c.in, c.want, len(got), got)
		}
	}
}
