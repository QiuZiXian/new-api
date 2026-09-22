package controller

import (
	"errors"
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relaykit/types"
)

// TestOpenaiErrorBodyDualWrite 本地错误必须双写：顶层扁平字段 + OpenAI error 包裹。
func TestOpenaiErrorBodyDualWrite(t *testing.T) {
	newAPIError := types.NewError(errors.New("boom"), types.ErrorCodeInvalidRequest,
		types.ErrOptionWithStatusCode(http.StatusBadRequest))
	body := openaiErrorBody(newAPIError)

	data, err := common.Marshal(body)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	var raw map[string]any
	if err := common.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if raw["code"] != string(types.ErrorCodeInvalidRequest) {
		t.Errorf("top-level code = %v, want %q", raw["code"], types.ErrorCodeInvalidRequest)
	}
	if raw["status_code"] != float64(http.StatusBadRequest) {
		t.Errorf("top-level status_code = %v, want 400", raw["status_code"])
	}
	detail, ok := raw["error"].(map[string]any)
	if !ok {
		t.Fatalf("error envelope missing or not an object: %v", raw["error"])
	}
	if detail["type"] != string(types.ErrorTypeNewAPIError) {
		t.Errorf("error type = %v, want %v", detail["type"], types.ErrorTypeNewAPIError)
	}
	if detail["message"] == nil || detail["message"] == "" {
		t.Errorf("error envelope message should not be empty: %v", detail)
	}
}

// TestOpenaiErrorBodyPreservesUpstream 上游错误要原样保留自己的 type / code / param，
// 不能被本地错误类型覆盖掉。
func TestOpenaiErrorBodyPreservesUpstream(t *testing.T) {
	upstream := types.WithOpenAIError(types.OpenAIError{
		Message: "Rate limit reached",
		Type:    "upstream_error",
		Code:    "rate_limit_exceeded",
		Param:   "model",
	}, http.StatusTooManyRequests)

	body := openaiErrorBody(upstream)
	if body.Error.Type != "upstream_error" {
		t.Errorf("error type = %q, want upstream_error", body.Error.Type)
	}
	if body.Error.Param != "model" {
		t.Errorf("error param = %q, want model", body.Error.Param)
	}
	if body.Code != "rate_limit_exceeded" {
		t.Errorf("code = %q, want rate_limit_exceeded", body.Code)
	}
	if body.StatusCode != http.StatusTooManyRequests {
		t.Errorf("status_code = %d, want 429", body.StatusCode)
	}
}

// TestRelayNotImplementedShape 确认即便没有 NewAPIError，手写响应也是统一形态。
func TestRelayNotImplementedShape(t *testing.T) {
	body := common.NewErrorBody(http.StatusNotImplemented, common.ErrorCodeNotImplemented, "API not implemented")
	if body.Error.Code != body.Code || body.Error.Message != body.Message {
		t.Errorf("flat fields and error envelope diverged: %+v", body)
	}
	if body.Error.Type != common.AppErrorType() {
		t.Errorf("error type = %q, want %q", body.Error.Type, common.AppErrorType())
	}
}
