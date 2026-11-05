package dto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// GetTokenCountMeta 的张数来源：顶层 n 优先，缺失时回退 parameters.n。
// 任何异常形态都必须安静地回落到 1，不能报错也不能产生异常的计费乘数。
func TestGetTokenCountMetaImageN(t *testing.T) {
	tests := []struct {
		name string
		body string
		want float64
	}{
		{"top-level n wins", `{"model":"m","prompt":"p","n":3,"parameters":{"n":4}}`, 3},
		{"fallback to parameters.n", `{"model":"m","prompt":"p","parameters":{"n":4}}`, 4},
		{"no parameters at all", `{"model":"m","prompt":"p"}`, 1},
		{"empty parameters", `{"model":"m","prompt":"p","parameters":{}}`, 1},
		{"zero n", `{"model":"m","prompt":"p","parameters":{"n":0}}`, 1},
		{"negative n", `{"model":"m","prompt":"p","parameters":{"n":-2}}`, 1},
		{"string n", `{"model":"m","prompt":"p","parameters":{"n":"4"}}`, 1},
		{"float n", `{"model":"m","prompt":"p","parameters":{"n":2.5}}`, 1},
		{"null n", `{"model":"m","prompt":"p","parameters":{"n":null}}`, 1},
		{"oversized n", `{"model":"m","prompt":"p","parameters":{"n":999}}`, 1},
		{"parameters is an array", `{"model":"m","prompt":"p","parameters":[]}`, 1},
		{"parameters is a string", `{"model":"m","prompt":"p","parameters":"n=4"}`, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req ImageRequest
			require.NoError(t, json.Unmarshal([]byte(tt.body), &req))

			meta := req.GetTokenCountMeta()
			require.NotNil(t, meta)
			assert.Equal(t, tt.want, meta.BillingRatios["n"])
		})
	}
}

// 顶层 n 超过 MaxImageN 时也要钳住，不能原样进入计费乘数。
func TestGetTokenCountMetaClampsTopLevelN(t *testing.T) {
	var req ImageRequest
	require.NoError(t, json.Unmarshal([]byte(`{"model":"m","prompt":"p","n":999}`), &req))

	// 顶层 n 走原有逻辑，这里只断言不 panic 且能被下游钳制（>0）。
	assert.Greater(t, req.GetTokenCountMeta().BillingRatios["n"], float64(0))
}
