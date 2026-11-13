package model

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPricingCarriesCustomCardImage(t *testing.T) {
	resetPricingEndpointTestTables(t)

	insertPricingEndpointChannel(t, 211, constant.ChannelTypeAdvancedCustom, pricingEndpointAdvancedCustomConfig())
	insertPricingEndpointAbility(t, 211, "wan3.0-video")
	insertPricingEndpointAbility(t, 211, "gpt-4o")
	require.NoError(t, DB.Create(&Model{
		ModelName: "wan3.0-video",
		CardImage: "https://cdn.example.com/wan3-card.webp",
		Status:    1,
		NameRule:  NameRuleExact,
	}).Error)

	InitChannelCache()
	InvalidatePricingCache()
	byName := make(map[string]Pricing)
	for _, p := range GetPricing() {
		byName[p.ModelName] = p
	}

	require.Contains(t, byName, "wan3.0-video")
	assert.Equal(t, "https://cdn.example.com/wan3-card.webp", byName["wan3.0-video"].CardImage)
	// 未配置的模型保持空值，前端据此回落到按类别的默认配图。
	require.Contains(t, byName, "gpt-4o")
	assert.Empty(t, byName["gpt-4o"].CardImage)
}

func TestModelUpdatePersistsAndClearsCardImage(t *testing.T) {
	resetPricingEndpointTestTables(t)

	m := &Model{
		ModelName: "card-image-model",
		Status:    1,
		NameRule:  NameRuleExact,
	}
	require.NoError(t, m.Insert())

	m.CardImage = "https://cdn.example.com/custom.webp"
	require.NoError(t, m.Update())

	var stored Model
	require.NoError(t, DB.First(&stored, m.Id).Error)
	assert.Equal(t, "https://cdn.example.com/custom.webp", stored.CardImage)

	// Update 用 Select 白名单强制写零值：漏列会导致清空配图不生效。
	m.CardImage = ""
	require.NoError(t, m.Update())

	var cleared Model
	require.NoError(t, DB.First(&cleared, m.Id).Error)
	assert.Empty(t, cleared.CardImage)
}
