package ali

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relay/channel"
	"github.com/QuantumNous/new-api/relay/channel/claude"
	"github.com/QuantumNous/new-api/relay/channel/openai"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/model_setting"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

type Adaptor struct {
	IsSyncImageModel bool
}

const aliAnthropicMessagesModelsEnv = "ALI_ANTHROPIC_MESSAGES_MODELS"
const defaultAliAnthropicMessagesModels = "qwen,deepseek-v4,kimi,glm,minimax-m"

// DashScope 内置图片端点。第三方网关（如京东云 modelservice）常把同一套
// Qwen-Image 协议挂在不同路径下、且不带 /api 前缀，此时用 ChannelOtherSettings
// 的 image_api_path 覆盖；默认值集中在这里，避免路径散落在各分支里。
const (
	defaultAliSyncImageApiPath  = "/api/v1/services/aigc/multimodal-generation/generation"
	defaultAliAsyncImageApiPath = "/api/v1/services/aigc/text2image/image-synthesis"
	defaultAliOldWanEditApiPath = "/api/v1/services/aigc/image2image/image-synthesis"
	defaultAliWanEditApiPath    = "/api/v1/services/aigc/image-generation/generation"
)

/*
	var syncModels = []string{
		"z-image",
		"qwen-image",
		"wan2.6",
	}
*/
func supportsAliAnthropicMessages(modelName string) bool {
	normalizedModelName := strings.ToLower(strings.TrimSpace(modelName))
	if normalizedModelName == "" {
		return false
	}

	return lo.SomeBy(aliAnthropicMessagesModelPatterns(), func(pattern string) bool {
		return strings.Contains(normalizedModelName, pattern)
	})
}

func aliAnthropicMessagesModelPatterns() []string {
	configuredModels := common.GetEnvOrDefaultString(aliAnthropicMessagesModelsEnv, defaultAliAnthropicMessagesModels)
	return lo.FilterMap(strings.Split(configuredModels, ","), func(item string, _ int) (string, bool) {
		pattern := strings.ToLower(strings.TrimSpace(item))
		return pattern, pattern != ""
	})
}

var syncModels = []string{
	"z-image",
	"qwen-image",
	"wan2.6",
}

func isSyncImageModel(modelName string) bool {
	return model_setting.IsSyncImageModel(modelName)
}

// imageApiPath 返回渠道自定义的图片 API 路径后缀，未配置或非法时返回 ""，
// 由 GetRequestURL 回落到 DashScope 内置路径。校验限制为纯相对路径，防止
// 管理员配置里夹带 scheme / query / ".." 造成路径注入。
func imageApiPath(info *relaycommon.RelayInfo) string {
	var p string
	if info != nil && info.HasChannelMeta() {
		p = strings.TrimSpace(info.ChannelOtherSettings.ImageApiPath)
	}
	if p == "" || len(p) > 256 {
		return ""
	}
	for _, r := range p {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '-', r == '.', r == '_', r == '~', r == '/':
		default:
			return ""
		}
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." {
			return ""
		}
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

func (a *Adaptor) ConvertGeminiRequest(*gin.Context, *relaycommon.RelayInfo, *dto.GeminiChatRequest) (any, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, info *relaycommon.RelayInfo, req *dto.ClaudeRequest) (any, error) {
	if supportsAliAnthropicMessages(info.UpstreamModelName) {
		return req, nil
	}

	result, err := service.ConvertRequest(c, info, types.RelayFormatOpenAI, req)
	if err != nil {
		return nil, err
	}
	oaiReq, ok := result.Value.(*dto.GeneralOpenAIRequest)
	if !ok {
		return nil, fmt.Errorf("expected OpenAI chat completions request, got %T", result.Value)
	}
	if info.SupportStreamOptions && info.IsStream {
		oaiReq.StreamOptions = &dto.StreamOptions{IncludeUsage: true}
	}
	return a.ConvertOpenAIRequest(c, info, oaiReq)
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo) {
	if info == nil {
		return
	}
	// IsSyncImageModel 决定 aliImageHandler 直接解析响应还是走任务轮询。
	// 必须在这里按上游模型算出来：请求体透传（pass_through_body_enabled）会
	// 跳过 ConvertImageRequest，而那里原本是唯一赋值点，导致同步图片模型在
	// 透传模式下被当成异步任务，拿不到 task_id 就返回空的「status code: 200」错误。
	a.IsSyncImageModel = isSyncImageModel(info.UpstreamModelName)
	if info.RelayMode == constant.RelayModeImagesEdits && isWanModel(info.UpstreamModelName) {
		// wan 系列图生图走异步端点，即使在同步名单里也不能同步解析。
		a.IsSyncImageModel = false
	}
}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	var fullRequestURL string
	switch info.RelayFormat {
	case types.RelayFormatClaude:
		if supportsAliAnthropicMessages(info.UpstreamModelName) {
			fullRequestURL = fmt.Sprintf("%s/apps/anthropic/v1/messages", info.ChannelBaseUrl)
		} else {
			fullRequestURL = fmt.Sprintf("%s/compatible-mode/v1/chat/completions", info.ChannelBaseUrl)
		}
	default:
		switch info.RelayMode {
		case constant.RelayModeEmbeddings:
			fullRequestURL = fmt.Sprintf("%s/compatible-mode/v1/embeddings", info.ChannelBaseUrl)
		case constant.RelayModeRerank:
			fullRequestURL = fmt.Sprintf("%s/api/v1/services/rerank/text-rerank/text-rerank", info.ChannelBaseUrl)
		case constant.RelayModeResponses:
			fullRequestURL = fmt.Sprintf("%s/api/v2/apps/protocols/compatible-mode/v1/responses", info.ChannelBaseUrl)
		case constant.RelayModeImagesGenerations:
			if customPath := imageApiPath(info); customPath != "" {
				fullRequestURL = fmt.Sprintf("%s%s", info.ChannelBaseUrl, customPath)
			} else if isSyncImageModel(info.UpstreamModelName) {
				fullRequestURL = fmt.Sprintf("%s%s", info.ChannelBaseUrl, defaultAliSyncImageApiPath)
			} else {
				fullRequestURL = fmt.Sprintf("%s%s", info.ChannelBaseUrl, defaultAliAsyncImageApiPath)
			}
		case constant.RelayModeImagesEdits:
			if customPath := imageApiPath(info); customPath != "" {
				fullRequestURL = fmt.Sprintf("%s%s", info.ChannelBaseUrl, customPath)
			} else if isOldWanModel(info.UpstreamModelName) {
				fullRequestURL = fmt.Sprintf("%s%s", info.ChannelBaseUrl, defaultAliOldWanEditApiPath)
			} else if isWanModel(info.UpstreamModelName) {
				fullRequestURL = fmt.Sprintf("%s%s", info.ChannelBaseUrl, defaultAliWanEditApiPath)
			} else {
				fullRequestURL = fmt.Sprintf("%s%s", info.ChannelBaseUrl, defaultAliSyncImageApiPath)
			}
		case constant.RelayModeCompletions:
			fullRequestURL = fmt.Sprintf("%s/compatible-mode/v1/completions", info.ChannelBaseUrl)
		default:
			fullRequestURL = fmt.Sprintf("%s/compatible-mode/v1/chat/completions", info.ChannelBaseUrl)
		}
	}

	return fullRequestURL, nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Header, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, req)
	req.Set("Authorization", "Bearer "+info.ApiKey)
	if info.IsStream {
		req.Set("X-DashScope-SSE", "enable")
	}
	if c.GetString("plugin") != "" {
		req.Set("X-DashScope-Plugin", c.GetString("plugin"))
	}
	if info.RelayMode == constant.RelayModeImagesGenerations {
		if isSyncImageModel(info.UpstreamModelName) {

		} else {
			req.Set("X-DashScope-Async", "enable")
		}
	}
	if info.RelayMode == constant.RelayModeImagesEdits {
		if isWanModel(info.UpstreamModelName) {
			req.Set("X-DashScope-Async", "enable")
		}
		req.Set("Content-Type", "application/json")
	}
	return nil
}

func (a *Adaptor) ConvertOpenAIRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	// docs: https://bailian.console.aliyun.com/?tab=api#/api/?type=model&url=2712216
	// fix: InternalError.Algo.InvalidParameter: The value of the enable_thinking parameter is restricted to True.
	//if strings.Contains(request.Model, "thinking") {
	//	request.EnableThinking = true
	//	request.Stream = true
	//	info.IsStream = true
	//}
	//// fix: ali parameter.enable_thinking must be set to false for non-streaming calls
	//if !info.IsStream {
	//	request.EnableThinking = false
	//}

	switch info.RelayMode {
	default:
		aliReq := requestOpenAI2Ali(*request, info.UpstreamModelName)
		return aliReq, nil
	}
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	if info.RelayMode == constant.RelayModeImagesGenerations {
		if isSyncImageModel(info.UpstreamModelName) {
			a.IsSyncImageModel = true
		}
		aliRequest, err := oaiImage2AliImageRequest(info, request, a.IsSyncImageModel)
		if err != nil {
			return nil, fmt.Errorf("convert image request to async ali image request failed: %w", err)
		}
		return aliRequest, nil
	} else if info.RelayMode == constant.RelayModeImagesEdits {
		if isOldWanModel(info.UpstreamModelName) {
			return oaiFormEdit2WanxImageEdit(c, info, request)
		}
		if isSyncImageModel(info.UpstreamModelName) {
			if isWanModel(info.UpstreamModelName) {
				a.IsSyncImageModel = false
			} else {
				a.IsSyncImageModel = true
			}
		}
		// ali image edit https://bailian.console.aliyun.com/?tab=api#/api/?type=model&url=2976416
		// 如果用户使用表单，则需要解析表单数据
		if strings.Contains(c.Request.Header.Get("Content-Type"), "multipart/form-data") {
			aliRequest, err := oaiFormEdit2AliImageEdit(c, info, request)
			if err != nil {
				return nil, fmt.Errorf("convert image edit form request failed: %w", err)
			}
			return aliRequest, nil
		} else {
			aliRequest, err := oaiImage2AliImageRequest(info, request, a.IsSyncImageModel)
			if err != nil {
				return nil, fmt.Errorf("convert image request to async ali image request failed: %w", err)
			}
			return aliRequest, nil
		}
	}
	return nil, fmt.Errorf("unsupported image relay mode: %d", info.RelayMode)
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return ConvertRerankRequest(request), nil
}

func (a *Adaptor) ConvertEmbeddingRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.EmbeddingRequest) (any, error) {
	return request, nil
}

func (a *Adaptor) ConvertAudioRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (io.Reader, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) ConvertOpenAIResponsesRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.OpenAIResponsesRequest) (any, error) {
	return request, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error) {
	return channel.DoApiRequest(a, c, info, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *types.NewAPIError) {
	switch info.RelayFormat {
	case types.RelayFormatClaude:
		if supportsAliAnthropicMessages(info.UpstreamModelName) {
			adaptor := claude.Adaptor{}
			return adaptor.DoResponse(c, resp, info)
		}

		adaptor := openai.Adaptor{}
		return adaptor.DoResponse(c, resp, info)
	default:
		switch info.RelayMode {
		case constant.RelayModeImagesGenerations:
			err, usage = aliImageHandler(a, c, resp, info)
		case constant.RelayModeImagesEdits:
			err, usage = aliImageHandler(a, c, resp, info)
		case constant.RelayModeRerank:
			err, usage = RerankHandler(c, resp, info)
		default:
			adaptor := openai.Adaptor{}
			usage, err = adaptor.DoResponse(c, resp, info)
		}
		return usage, err
	}
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
