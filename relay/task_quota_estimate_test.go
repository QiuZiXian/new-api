package relay

import (
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	hosttypes "github.com/QuantumNous/new-api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaskBaseQuotaFromEstimatedTokens 守护视频任务的预扣基数：
// 按 token 计费时应使用官方公式估算的 token（5s/720p ≈ 108,900 tokens），
// 而不是「25 万 token × 时长」——后者会把冻结额度放大一个数量级。
func TestTaskBaseQuotaFromEstimatedTokens(t *testing.T) {
	newInfo := func() *relaycommon.RelayInfo {
		return &relaycommon.RelayInfo{
			PriceData: hosttypes.PriceData{
				EstimatedTokens: 108900,
				ModelRatio:      17.5,
				GroupRatioInfo:  hosttypes.GroupRatioInfo{GroupRatio: 1},
			},
		}
	}

	t.Run("token mode uses estimated tokens", func(t *testing.T) {
		quota, ok := taskBaseQuotaFromEstimatedTokens(newInfo(), "doubao-seedance-2-0-mini")
		require.True(t, ok)
		// 108900 × 17.5 = 1,905,750 quota（≈$3.81），与终态结算同一口径
		assert.Equal(t, 1905750, quota)
	})

	t.Run("group ratio still applies", func(t *testing.T) {
		info := newInfo()
		info.PriceData.GroupRatioInfo.GroupRatio = 2
		quota, ok := taskBaseQuotaFromEstimatedTokens(info, "doubao-seedance-2-0-mini")
		require.True(t, ok)
		assert.Equal(t, 3811500, quota)
	})

	t.Run("per-second and per-call modes are excluded", func(t *testing.T) {
		perSecond := newInfo()
		perSecond.PriceData.UseVideoSecond = true
		_, ok := taskBaseQuotaFromEstimatedTokens(perSecond, "m")
		assert.False(t, ok)

		perCall := newInfo()
		perCall.PriceData.UsePrice = true
		_, ok = taskBaseQuotaFromEstimatedTokens(perCall, "m")
		assert.False(t, ok)

		free := newInfo()
		free.PriceData.FreeModel = true
		_, ok = taskBaseQuotaFromEstimatedTokens(free, "m")
		assert.False(t, ok)
	})

	t.Run("no estimate available keeps legacy base", func(t *testing.T) {
		info := newInfo()
		info.PriceData.EstimatedTokens = 0
		_, ok := taskBaseQuotaFromEstimatedTokens(info, "m")
		assert.False(t, ok)
	})
}
