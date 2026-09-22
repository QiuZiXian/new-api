package common

// 对外接口的错误响应统一形态。
//
// 历史上不同接口各自为政：relay 链路输出 OpenAI 的 {"error":{...}}，
// 任务/素材类接口输出扁平的 {"code","message"}。这里把两者收敛成一个体：
// 顶层保留扁平字段，同时并列 OpenAI 风格的 error 包裹，
// 老客户端不用改，新客户端可以按 OpenAI 约定解析。

// ErrorDetail 是 OpenAI 风格的错误包裹体。Param 为可选项，
// 只有能定位到具体请求参数时才输出。
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param,omitempty"`
}

// ErrorBody 是对外统一的错误响应体，扁平字段与 error 包裹双写。
type ErrorBody struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	StatusCode int         `json:"status_code"`
	Error      ErrorDetail `json:"error"`
}

// NewErrorBody 构造统一错误响应体。code 为空时按 HTTP 状态回落，
// type 默认走当前 AppSlug 派生的本地错误类型（默认 "new_api_error"）。
func NewErrorBody(statusCode int, code, message string) ErrorBody {
	return NewErrorBodyWithType(statusCode, code, message, AppErrorType())
}

// NewErrorBodyWithType 允许指定 type，用于 panic 兜底、上游透传等场景。
func NewErrorBodyWithType(statusCode int, code, message, errorType string) ErrorBody {
	if code == "" {
		code = DefaultCodeForStatus(statusCode)
	}
	return ErrorBody{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Type:    errorType,
		},
	}
}

// DefaultCodeForStatus 给没带业务码的错误回落一个稳定的 code，
// 避免同一个 HTTP 状态在不同接口冒出不同字符串。
func DefaultCodeForStatus(statusCode int) string {
	switch {
	case statusCode == 400:
		return ErrorCodeBadRequest
	case statusCode == 401:
		return ErrorCodeUnauthorized
	case statusCode == 403:
		return ErrorCodeForbidden
	case statusCode == 404:
		return ErrorCodeNotFound
	case statusCode == 409:
		return ErrorCodeConflict
	case statusCode == 429:
		return ErrorCodeRateLimited
	case statusCode == 501:
		return ErrorCodeNotImplemented
	case statusCode >= 500:
		return ErrorCodeInternal
	default:
		return ErrorCodeBadRequest
	}
}

// 对外统一错误码。业务可以通过更细的码覆盖，但回落值始终取自这里。
const (
	ErrorCodeBadRequest     = "bad_request"
	ErrorCodeUnauthorized   = "unauthorized"
	ErrorCodeForbidden      = "access_denied"
	ErrorCodeNotFound       = "not_found"
	ErrorCodeConflict       = "conflict"
	ErrorCodeRateLimited    = "rate_limited"
	ErrorCodeInternal       = "internal_error"
	ErrorCodeNotImplemented = "not_implemented"
	ErrorCodeUpstreamFailed = "upstream_error"
)
