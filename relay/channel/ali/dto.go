package ali

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

type AliMessage struct {
	Content any    `json:"content"`
	Role    string `json:"role"`
}

type AliMediaContent struct {
	Image string `json:"image,omitempty"`
	Text  string `json:"text,omitempty"`
}

type AliInput struct {
	Prompt string `json:"prompt,omitempty"`
	//History []AliMessage `json:"history,omitempty"`
	Messages []AliMessage `json:"messages"`
}

type AliParameters struct {
	TopP              float64 `json:"top_p,omitempty"`
	TopK              int     `json:"top_k,omitempty"`
	Seed              uint64  `json:"seed,omitempty"`
	EnableSearch      bool    `json:"enable_search,omitempty"`
	IncrementalOutput bool    `json:"incremental_output,omitempty"`
}

type AliChatRequest struct {
	Model      string        `json:"model"`
	Input      AliInput      `json:"input,omitempty"`
	Parameters AliParameters `json:"parameters,omitempty"`
}

type AliEmbeddingRequest struct {
	Model string `json:"model"`
	Input struct {
		Texts []string `json:"texts"`
	} `json:"input"`
	Parameters *struct {
		TextType string `json:"text_type,omitempty"`
	} `json:"parameters,omitempty"`
}

type AliEmbedding struct {
	Embedding []float64 `json:"embedding"`
	TextIndex int       `json:"text_index"`
}

type AliEmbeddingResponse struct {
	Output struct {
		Embeddings []AliEmbedding `json:"embeddings"`
	} `json:"output"`
	Usage AliUsage `json:"usage"`
	AliError
}

type AliError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestId string `json:"request_id"`
}

type AliUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
	ImageCount   int `json:"image_count,omitempty"`
	// 同步多模态生成端点（qwen-image / 京东云 modelservice 等）回报的是下面这组
	// 字段，而不是 image_count。OutputImageCount 只统计成功生成的图片，正是
	// 「仅对成功生成图片按张数计费」的可靠来源。
	InputImageCount  int    `json:"input_image_count,omitempty"`
	OutputImageCount int    `json:"output_image_count,omitempty"`
	InputImageType   string `json:"input_image_type,omitempty"`
	OutputImageType  string `json:"output_image_type,omitempty"`
	OutputWidth      int    `json:"output_width,omitempty"`
	OutputHeight     int    `json:"output_height,omitempty"`
}

type TaskResult struct {
	B64Image string `json:"b64_image,omitempty"`
	Url      string `json:"url,omitempty"`
	Code     string `json:"code,omitempty"`
	Message  string `json:"message,omitempty"`
}

type AliOutput struct {
	TaskId       string       `json:"task_id,omitempty"`
	TaskStatus   string       `json:"task_status,omitempty"`
	Text         string       `json:"text"`
	FinishReason string       `json:"finish_reason"`
	Message      string       `json:"message,omitempty"`
	Code         string       `json:"code,omitempty"`
	Results      []TaskResult `json:"results,omitempty"`
	Choices      []struct {
		FinishReason string `json:"finish_reason,omitempty"`
		Message      struct {
			Role             string            `json:"role,omitempty"`
			Content          []AliMediaContent `json:"content,omitempty"`
			ReasoningContent string            `json:"reasoning_content,omitempty"`
		} `json:"message,omitempty"`
	} `json:"choices,omitempty"`
}

func (o *AliOutput) ChoicesToOpenAIImageDate(c *gin.Context, responseFormat string) []dto.ImageData {
	var imageData []dto.ImageData
	for _, choice := range o.Choices {
		// n > 1 时上游会把多张图塞进同一条 message 的 content 数组里，
		// 必须逐张产出结果；早期实现每个 choice 只建一个 ImageData，
		// 循环中被后一张覆盖，导致请求 2 张只回 1 张。
		revisedPrompt := ""
		for _, content := range choice.Message.Content {
			if content.Image == "" && content.Text != "" {
				revisedPrompt = content.Text
				break
			}
		}

		for _, content := range choice.Message.Content {
			if content.Image == "" {
				continue
			}
			item := dto.ImageData{RevisedPrompt: revisedPrompt}
			if strings.HasPrefix(content.Image, "http") {
				item.Url = content.Image
				if responseFormat == "b64_json" {
					_, b64, err := service.GetImageFromUrl(content.Image)
					if err != nil {
						logger.LogError(c, "get_image_data_failed: "+err.Error())
						continue
					}
					item.B64Json = b64
				}
			} else {
				item.B64Json = content.Image
			}
			imageData = append(imageData, item)
		}
	}

	return imageData
}

func (o *AliOutput) ResultToOpenAIImageDate(c *gin.Context, responseFormat string) []dto.ImageData {
	var imageData []dto.ImageData
	for _, data := range o.Results {
		var b64Json string
		if responseFormat == "b64_json" {
			_, b64, err := service.GetImageFromUrl(data.Url)
			if err != nil {
				logger.LogError(c, "get_image_data_failed: "+err.Error())
				continue
			}
			b64Json = b64
		} else {
			b64Json = data.B64Image
		}

		imageData = append(imageData, dto.ImageData{
			Url:           data.Url,
			B64Json:       b64Json,
			RevisedPrompt: "",
		})
	}
	return imageData
}

type AliResponse struct {
	Output AliOutput `json:"output"`
	Usage  AliUsage  `json:"usage"`
	AliError
}

type AliImageRequest struct {
	Model          string             `json:"model"`
	Input          any                `json:"input"`
	Parameters     AliImageParameters `json:"parameters,omitempty"`
	ResponseFormat string             `json:"response_format,omitempty"`
}

type AliImageParameters struct {
	Size         string `json:"size,omitempty"`
	N            int    `json:"n,omitempty"`
	Steps        string `json:"steps,omitempty"`
	Scale        string `json:"scale,omitempty"`
	Watermark    *bool  `json:"watermark,omitempty"`
	PromptExtend *bool  `json:"prompt_extend,omitempty"`
	ThinkingMode *bool  `json:"thinking_mode,omitempty"`
	// PromptExtendMode 提示词改写模式：direct（默认）/ agent（APE，仅 T2I）。
	PromptExtendMode string `json:"prompt_extend_mode,omitempty"`
	// NegativePrompt 反向提示词。
	NegativePrompt   string `json:"negative_prompt,omitempty"`
	EnableSequential *bool  `json:"enable_sequential,omitempty"`
	BboxList         any    `json:"bbox_list,omitempty"`
	ColorPalette     any    `json:"color_palette,omitempty"`
	Seed             *int   `json:"seed,omitempty"`

	// Extra 保存上面未声明的 parameters 字段，序列化时平铺回 parameters。
	// 上游参数会持续增加，只靠白名单结构体的话新参数会被静默丢弃
	// （不报错、无日志），调用方会误以为已经生效。
	Extra map[string]json.RawMessage `json:"-"`
}

// aliImageParameterKnownKeys 由结构体 tag 推导，新增字段无需同步维护。
var aliImageParameterKnownKeys = func() map[string]struct{} {
	keys := make(map[string]struct{})
	t := reflect.TypeOf(AliImageParameters{})
	for i := 0; i < t.NumField(); i++ {
		name, _, _ := strings.Cut(t.Field(i).Tag.Get("json"), ",")
		if name == "" || name == "-" {
			continue
		}
		keys[name] = struct{}{}
	}
	return keys
}()

func (p *AliImageParameters) UnmarshalJSON(data []byte) error {
	type plain AliImageParameters
	var decoded plain
	if err := common.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var raw map[string]json.RawMessage
	if err := common.Unmarshal(data, &raw); err != nil {
		return err
	}
	decoded.Extra = nil
	for key := range aliImageParameterKnownKeys {
		delete(raw, key)
	}
	if len(raw) > 0 {
		decoded.Extra = raw
	}
	*p = AliImageParameters(decoded)
	return nil
}

func (p AliImageParameters) MarshalJSON() ([]byte, error) {
	type plain AliImageParameters
	base, err := common.Marshal(plain(p))
	if err != nil {
		return nil, err
	}
	if len(p.Extra) == 0 {
		return base, nil
	}
	var merged map[string]json.RawMessage
	if err := common.Unmarshal(base, &merged); err != nil {
		return nil, err
	}
	// 具名字段优先：Extra 只补结构体没有声明的 key，不会覆盖已知字段。
	for key, value := range p.Extra {
		if _, exists := merged[key]; !exists {
			merged[key] = value
		}
	}
	return common.Marshal(merged)
}

func (p *AliImageParameters) PromptExtendValue() bool {
	if p != nil && p.PromptExtend != nil {
		return *p.PromptExtend
	}
	return false
}

type AliImageInput struct {
	Prompt         string       `json:"prompt,omitempty"`
	NegativePrompt string       `json:"negative_prompt,omitempty"`
	Messages       []AliMessage `json:"messages,omitempty"`
}

type WanImageInput struct {
	Prompt         string   `json:"prompt"`                    // 必需：文本提示词，描述生成图像中期望包含的元素和视觉特点
	Images         []string `json:"images"`                    // 必需：图像URL数组，长度不超过2，支持HTTP/HTTPS URL或Base64编码
	NegativePrompt string   `json:"negative_prompt,omitempty"` // 可选：反向提示词，描述不希望在画面中看到的内容
}

type WanImageParameters struct {
	N         int     `json:"n,omitempty"`         // 生成图片数量，取值范围1-4，默认4
	Watermark *bool   `json:"watermark,omitempty"` // 是否添加水印标识，默认false
	Seed      int     `json:"seed,omitempty"`      // 随机数种子，取值范围[0, 2147483647]
	Strength  float64 `json:"strength,omitempty"`  // 修改幅度 0.0-1.0，默认0.5（部分模型支持）
}

type AliRerankParameters struct {
	TopN            *int  `json:"top_n,omitempty"`
	ReturnDocuments *bool `json:"return_documents,omitempty"`
}

type AliRerankInput struct {
	Query     string `json:"query"`
	Documents []any  `json:"documents"`
}

type AliRerankRequest struct {
	Model      string              `json:"model"`
	Input      AliRerankInput      `json:"input"`
	Parameters AliRerankParameters `json:"parameters,omitempty"`
}

type AliRerankResponse struct {
	Output struct {
		Results []dto.RerankResponseResult `json:"results"`
	} `json:"output"`
	Usage     AliUsage `json:"usage"`
	RequestId string   `json:"request_id"`
	AliError
}
