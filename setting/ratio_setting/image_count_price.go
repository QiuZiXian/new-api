package ratio_setting

import (
	"fmt"
	"math"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/types"
)

// ImageCountPrice 是按张计费的单张价格：输入参考图按 Input 计价，输出生成图按
// Output 计价。价格口径与 ModelPrice 一致（单价 × QuotaPerUnit × 分组倍率），
// 因此站点以人民币为口径时 0.18 就表示 ¥0.18/张。
//
// 输入和输出是两条独立计价维度，必须加和，不能像 OtherRatios 那样连乘：
// 连乘会得到「单价 × 输入张数 × 输出张数」，语义完全不对。
type ImageCountPrice struct {
	Input  float64 `json:"input"`
	Output float64 `json:"output"`
}

// maxImageUnitPrice 限制单张价格，避免管理员误填超大值把配额乘法推到饱和区间。
const maxImageUnitPrice = 1_000_000

var imageCountPriceMap = types.NewRWMap[string, ImageCountPrice]()

// GetImageCountPrice 返回模型的按张计价配置。没有任何正价时视为未配置，
// 调用方继续走按次 / 按 token 的既有逻辑。
func GetImageCountPrice(model string) (ImageCountPrice, bool) {
	price, ok := imageCountPriceMap.Get(model)
	if !ok || !price.IsUsable() {
		return ImageCountPrice{}, false
	}
	return price.Sanitized(), true
}

// IsUsable 报告该配置是否至少有一个可用的正价。
func (p ImageCountPrice) IsUsable() bool {
	return (p.Input > 0 || p.Output > 0) &&
		validImageUnitPrice(p.Input) && validImageUnitPrice(p.Output)
}

// Sanitized 把负数 / NaN / 超限价格归零，保证费用永远是非负加和。
func (p ImageCountPrice) Sanitized() ImageCountPrice {
	if !validImageUnitPrice(p.Input) {
		p.Input = 0
	}
	if !validImageUnitPrice(p.Output) {
		p.Output = 0
	}
	return p
}

// ImageCountCost 返回 inputCount 张输入图 + outputCount 张输出图的原始费用。
// 张数在此处二次钳制，与 PriceData.SetImageCounts 形成纵深防御。
func ImageCountCost(price ImageCountPrice, inputCount, outputCount int) float64 {
	p := price.Sanitized()
	cost := float64(clampImageCount(inputCount))*p.Input + float64(clampImageCount(outputCount))*p.Output
	if cost < 0 || math.IsNaN(cost) || math.IsInf(cost, 0) {
		return 0
	}
	return cost
}

func ImageCountPrice2JSONString() string {
	return imageCountPriceMap.MarshalJSONString()
}

func GetImageCountPriceCopy() map[string]ImageCountPrice {
	return imageCountPriceMap.ReadAll()
}

func UpdateImageCountPriceByJSONString(jsonStr string) error {
	var parsed map[string]ImageCountPrice
	if err := common.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return fmt.Errorf("invalid image count price JSON: %w", err)
	}
	for model, price := range parsed {
		if !price.IsUsable() {
			return fmt.Errorf("invalid image count price for model %q: input/output must be non-negative finite numbers with at least one positive", model)
		}
	}
	return types.LoadFromJsonStringWithCallback(imageCountPriceMap, jsonStr, InvalidateExposedDataCache)
}

func validImageUnitPrice(price float64) bool {
	return price >= 0 && price <= maxImageUnitPrice && !math.IsNaN(price) && !math.IsInf(price, 0)
}

func clampImageCount(n int) int {
	if n <= 0 {
		return 0
	}
	if n > dto.MaxImageN {
		return dto.MaxImageN
	}
	return n
}
