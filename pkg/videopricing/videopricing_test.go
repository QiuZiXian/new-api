package videopricing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetTierRatio 守护档位倍率表：计费与模型广场展示共用同一份数据，
// 对外模型名与上游模型名必须命中同一张表。
func TestGetTierRatio(t *testing.T) {
	cases := []struct {
		name       string
		model      string
		resolution string
		hasVideo   bool
		wantRatio  float64
		wantOK     bool
	}{
		// 2.0 全档
		{"2-0 480p no video", "doubao-seedance-2-0-260128", "", false, 1.0, true},
		{"2-0 720p no video", "doubao-seedance-2-0-260128", "720p", false, 1.0, true},
		{"2-0 480p with video", "doubao-seedance-2-0-260128", "", true, 28.0 / 46.0, true},
		{"2-0 1080p no video", "doubao-seedance-2-0-260128", "1080p", false, 51.0 / 46.0, true},
		{"2-0 4k with video", "doubao-seedance-2-0-260128", "4k", true, 16.0 / 46.0, true},
		{"2-0 unknown resolution", "doubao-seedance-2-0-260128", "999p", false, 1.0, true},

		// 2.0-fast 仅 480p/720p
		{"2-0-fast 480p with video", "doubao-seedance-2-0-fast-260128", "", true, 22.0 / 37.0, true},
		{"2-0-fast 1080p falls back to base", "doubao-seedance-2-0-fast-260128", "1080p", false, 1.0, true},

		// 2.0-mini 仅 480p/720p
		{"2-0-mini 480p no video", "doubao-seedance-2-0-mini", "", false, 1.0, true},
		{"2-0-mini with video", "doubao-seedance-2-0-mini", "", true, 14.0 / 23.0, true},
		{"2-0-mini 1080p falls back to base", "doubao-seedance-2-0-mini", "1080p", false, 1.0, true},

		// 对外名（OriginModelName）必须命中同一张价格表
		{"external 2.0 name hits 1080p tier", "doubao-seedance-2.0", "1080p", false, 51.0 / 46.0, true},
		{"external 2.0 name with video", "doubao-seedance-2.0", "", true, 28.0 / 46.0, true},
		{"external 2.0-fast name with video", "doubao-seedance-2.0-fast", "", true, 22.0 / 37.0, true},
		{"external 2.0-mini name with video", "doubao-seedance-2.0-mini", "", true, 14.0 / 23.0, true},
		{"external 2.5 name hits 1080p tier", "doubao-seedance-2.5", "1080p", false, 46.0 / 40.0, true},

		// 2.5 支持到 1080p
		{"2.5 480p no video", "doubao-seedance-2.5", "", false, 1.0, true},
		{"2.5 with video", "doubao-seedance-2.5", "", true, 24.0 / 40.0, true},
		{"2.5 1080p", "doubao-seedance-2.5", "1080p", false, 46.0 / 40.0, true},
		{"2.5 4k falls back to base", "doubao-seedance-2.5", "4k", false, 1.0, true},

		// 1.5-pro / 1.0-pro / lite 系列
		{"1-5-pro 480p", "doubao-seedance-1-5-pro-251215", "", false, 1.0, true},
		{"1-5-pro with video", "doubao-seedance-1-5-pro-251215", "", true, 18.0 / 30.0, true},
		{"1-0-pro 480p", "doubao-seedance-1-0-pro-250528", "", false, 1.0, true},
		{"1-0-lite-t2v with video", "doubao-seedance-1-0-lite-t2v", "", true, 9.0 / 15.0, true},
		{"1-0-lite-i2v with video", "doubao-seedance-1-0-lite-i2v", "", true, 11.0 / 18.0, true},

		// 未知模型
		{"unknown model returns false", "doubao-seedance-9-9-999999", "", false, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ratio, ok := GetTierRatio(tc.model, tc.resolution, tc.hasVideo)
			assert.Equal(t, tc.wantOK, ok)
			assert.InDelta(t, tc.wantRatio, ratio, 1e-9)
		})
	}
}

// TestGetTierPrice 守护展示用单价：返回档位单价与基准价，
// 展示层据此算出「与实扣口径一致」的每百万 token 价格。
func TestGetTierPrice(t *testing.T) {
	price, base, ok := GetTierPrice("doubao-seedance-2.0", "1080p", false)
	assert.True(t, ok)
	assert.InDelta(t, 51.0, price, 1e-9)
	assert.InDelta(t, 46.0, base, 1e-9)

	price, base, ok = GetTierPrice("doubao-seedance-2.0", "", false)
	assert.True(t, ok)
	assert.InDelta(t, 46.0, price, 1e-9)
	assert.InDelta(t, 46.0, base, 1e-9)

	_, _, ok = GetTierPrice("gpt-4o", "", false)
	assert.False(t, ok)
}

func TestNormalizeModelName(t *testing.T) {
	cases := map[string]string{
		"doubao-seedance-2.0":            "doubao-seedance-2-0",
		"doubao-seedance-2-0-260128":     "doubao-seedance-2-0",
		"doubao-seedance-2.0-fast":       "doubao-seedance-2-0-fast",
		"doubao-seedance-2.0-mini":       "doubao-seedance-2-0-mini",
		"doubao-seedance-1-5-pro-251215": "doubao-seedance-1-5-pro",
		"doubao-seedance-2.5":            "doubao-seedance-2-5",
		"  DOUBAO-SEEDANCE-2.0  ":        "doubao-seedance-2-0",
	}
	for in, want := range cases {
		assert.Equal(t, want, NormalizeModelName(in), in)
	}
}
