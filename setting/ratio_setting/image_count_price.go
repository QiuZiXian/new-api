package ratio_setting

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/types"
)

const (
	// ImageTierDefault 表示「未分档」的兜底单价，不写进配置 JSON。
	ImageTierDefault = "default"
	// ImageTier1K / ImageTier2K 是分辨率档位名，对应上游 usage 里的
	// qima_input_1k / qima_output_2k 这类标签。
	ImageTier1K = "1k"
	ImageTier2K = "2k"

	// imageTier1KMaxArea 是 1K 与 2K 档位的分界线（像素面积）。
	// 来源：Qwen-Image-3.0/3.0-pro 使用手册——「1K/2K 档位按输出分辨率面积判定
	// （≤2,250,000 为 1K，>2,250,000 为 2K），影响计费档位」。
	imageTier1KMaxArea = 2_250_000
)

// ImageCountPrice 是按张计费的单张价格。输入参考图与输出生成图是两条独立
// 计价维度，费用为「输入张数 × 输入单价 + 输出张数 × 输出单价」，必须加和，
// 不能像 OtherRatios 那样连乘（连乘会算成 单价 × 输入张数 × 输出张数）。
//
// 单价口径与 ModelPrice 一致：单价 × QuotaPerUnit × 分组倍率，因此站点以
// 人民币为口径时 0.18 就表示 ¥0.18/张。
//
// 分档：Input/Output 是未分档的兜底价，*_1k / *_2k 是档位覆盖价。档位价缺失
// 时回退到兜底价，所以旧的 {"input":0.02,"output":0.18} 配置继续可用。
type ImageCountPrice struct {
	Input  float64 `json:"input,omitempty"`
	Output float64 `json:"output,omitempty"`

	Input1K  float64 `json:"input_1k,omitempty"`
	Input2K  float64 `json:"input_2k,omitempty"`
	Output1K float64 `json:"output_1k,omitempty"`
	Output2K float64 `json:"output_2k,omitempty"`

	// FreeFirstOutput 表示每次请求的第一个输出张不计费（首张免费）。
	// 上游仍按成功张数收费，这里只在服务端计费侧打折，请求参数不变。
	FreeFirstOutput bool `json:"free_first_output,omitempty"`
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

// IsUsable 报告该配置是否合法且至少有一个可用的正价。
// 这里保持严格：负价 / 超限价直接判为非法，由保存接口报错，
// 不能静默清零——否则管理员会把「配错了」误当成「已生效」。
func (p ImageCountPrice) IsUsable() bool {
	return p.HasPositivePrice() && p.isValid()
}

// isValid 校验所有档位价都是非负有限且在允许范围内。
func (p ImageCountPrice) isValid() bool {
	for _, v := range p.prices() {
		if v < 0 || v > maxImageUnitPrice || math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}
	return true
}

// HasPositivePrice 报告是否至少配置了一个大于 0 的单价。
func (p ImageCountPrice) HasPositivePrice() bool {
	for _, v := range p.prices() {
		if v > 0 {
			return true
		}
	}
	return false
}

// Sanitized 把负数 / NaN / 超限价格归零，保证费用永远是非负加和。
func (p ImageCountPrice) Sanitized() ImageCountPrice {
	p.Input = sanitizeImageUnitPrice(p.Input)
	p.Output = sanitizeImageUnitPrice(p.Output)
	p.Input1K = sanitizeImageUnitPrice(p.Input1K)
	p.Input2K = sanitizeImageUnitPrice(p.Input2K)
	p.Output1K = sanitizeImageUnitPrice(p.Output1K)
	p.Output2K = sanitizeImageUnitPrice(p.Output2K)
	return p
}

// InputPriceFor 返回指定档位的输入单价，档位价缺失时回退到未分档单价。
func (p ImageCountPrice) InputPriceFor(tier string) float64 {
	switch normalizeImageTier(tier) {
	case ImageTier1K:
		if p.Input1K > 0 {
			return p.Input1K
		}
	case ImageTier2K:
		if p.Input2K > 0 {
			return p.Input2K
		}
	}
	return p.Input
}

// OutputPriceFor 返回指定档位的输出单价，档位价缺失时回退到未分档单价。
func (p ImageCountPrice) OutputPriceFor(tier string) float64 {
	switch normalizeImageTier(tier) {
	case ImageTier1K:
		if p.Output1K > 0 {
			return p.Output1K
		}
	case ImageTier2K:
		if p.Output2K > 0 {
			return p.Output2K
		}
	}
	return p.Output
}

// MaxOutputPrice 返回输出版最高档单价。/api/pricing 的 ModelPrice 与模型广场
// 列表排序都复用它，保证展示价不低于实际可能扣到的价格。
func (p ImageCountPrice) MaxOutputPrice() float64 {
	return math.Max(p.Output, math.Max(p.Output1K, p.Output2K))
}

func (p ImageCountPrice) prices() []float64 {
	return []float64{p.Input, p.Output, p.Input1K, p.Input2K, p.Output1K, p.Output2K}
}

// ImageTierByArea 按分辨率面积判定档位，与上游口径一致。
// 任一维度缺失时返回空串，表示「未知档位」，调用方回退到未分档单价。
func ImageTierByArea(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	area := int64(width) * int64(height)
	if area <= imageTier1KMaxArea {
		return ImageTier1K
	}
	return ImageTier2K
}

// ImageTierBySize 解析 "2048*2048" / "2048x2048" 这类尺寸字符串并判定档位。
// 阿里渠道的 parameters.size 用的就是「宽*高」格式。
func ImageTierBySize(size string) string {
	width, height, ok := parseImageSize(size)
	if !ok {
		return ""
	}
	return ImageTierByArea(width, height)
}

// NormalizeImageTier 从上游标签里识别档位，例如 qima_output_2k → 2k。
// 上游没有该字段或格式不认识时返回空串。
func NormalizeImageTier(label string) string {
	tier := strings.ToLower(strings.TrimSpace(label))
	switch {
	case tier == "":
		return ""
	case tier == ImageTier1K, strings.HasSuffix(tier, "_1k"), strings.HasSuffix(tier, "-1k"):
		return ImageTier1K
	case tier == ImageTier2K, strings.HasSuffix(tier, "_2k"), strings.HasSuffix(tier, "-2k"):
		return ImageTier2K
	}
	return ""
}

// ImageCountCost 返回 inputCount 张输入图 + outputCount 张输出图的原始费用。
// 张数在此处二次钳制，与 PriceData.SetImageCounts 形成纵深防御；
// 开启首张免费时输出侧按 max(0, n-1) 计费。
func ImageCountCost(price ImageCountPrice, inputCount, outputCount int, inputTier, outputTier string) float64 {
	p := price.Sanitized()
	billableOutput := clampImageCount(outputCount)
	if p.FreeFirstOutput && billableOutput > 0 {
		billableOutput--
	}
	cost := float64(clampImageCount(inputCount))*p.InputPriceFor(inputTier) +
		float64(billableOutput)*p.OutputPriceFor(outputTier)
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
			return fmt.Errorf("invalid image count price for model %q: prices must be non-negative finite numbers with at least one positive", model)
		}
	}
	return types.LoadFromJsonStringWithCallback(imageCountPriceMap, jsonStr, InvalidateExposedDataCache)
}

func sanitizeImageUnitPrice(price float64) float64 {
	if price < 0 || price > maxImageUnitPrice || math.IsNaN(price) || math.IsInf(price, 0) {
		return 0
	}
	return price
}

func normalizeImageTier(tier string) string {
	switch strings.ToLower(strings.TrimSpace(tier)) {
	case ImageTier1K:
		return ImageTier1K
	case ImageTier2K:
		return ImageTier2K
	}
	return ""
}

// parseImageSize 解析「宽*高」/「宽x高」（大小写不敏感）。
func parseImageSize(size string) (int, int, bool) {
	normalized := strings.ToLower(strings.TrimSpace(size))
	if normalized == "" {
		return 0, 0, false
	}
	parts := strings.FieldsFunc(normalized, func(r rune) bool {
		return r == '*' || r == 'x'
	})
	if len(parts) != 2 {
		return 0, 0, false
	}
	width, errWidth := strconv.Atoi(strings.TrimSpace(parts[0]))
	height, errHeight := strconv.Atoi(strings.TrimSpace(parts[1]))
	if errWidth != nil || errHeight != nil || width <= 0 || height <= 0 {
		return 0, 0, false
	}
	return width, height, true
}

// clampImageCount 把张数钳到 [0, MaxImageN]，与 PriceData 侧的钳制互为纵深防御。
func clampImageCount(n int) int {
	if n <= 0 {
		return 0
	}
	if n > dto.MaxImageN {
		return dto.MaxImageN
	}
	return n
}
