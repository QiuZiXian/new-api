package controller

import (
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relaykit/types"
)

// openaiErrorBody 把 relay 链路的本地错误转成统一错误体：
// 顶层 code / message / status_code 与 OpenAI 风格 error 包裹双写，
// 使 chat/completions 与素材、任务类接口可以用同一套错误解析逻辑。
func openaiErrorBody(newAPIError *types.NewAPIError) common.ErrorBody {
	openaiErr := newAPIError.ToOpenAIError()
	message := openaiErr.Message
	if message == "" {
		message = string(newAPIError.GetErrorCode())
	}
	code := string(newAPIError.GetErrorCode())
	if openaiErr.Code != nil {
		if rawCode, ok := openaiErr.Code.(string); ok {
			code = rawCode
		} else {
			code = fmt.Sprintf("%v", openaiErr.Code)
		}
	}
	if code == "" {
		code = common.DefaultCodeForStatus(newAPIError.StatusCode)
	}
	errorType := openaiErr.Type
	if errorType == "" {
		errorType = common.AppErrorType()
	}
	body := common.NewErrorBodyWithType(newAPIError.StatusCode, code, message, errorType)
	body.Error.Param = openaiErr.Param
	return body
}
