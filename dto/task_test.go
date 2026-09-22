package dto

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
)

// TestTaskErrorMarshalDualWrite 序列化后的 TaskError 必须同时具备扁平字段与
// OpenAI 风格 error 包裹，这样 relay 链路与其它 /v1 接口的错误形态一致。
func TestTaskErrorMarshalDualWrite(t *testing.T) {
	taskErr := &TaskError{Code: "task_not_exist", Message: "任务不存在", StatusCode: 400}
	data, err := common.Marshal(taskErr)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	var raw map[string]any
	if err := common.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if raw["code"] != "task_not_exist" {
		t.Errorf("top-level code = %v", raw["code"])
	}
	if raw["message"] != "任务不存在" {
		t.Errorf("top-level message = %v", raw["message"])
	}
	if raw["status_code"] != float64(400) {
		t.Errorf("top-level status_code = %v, want 400", raw["status_code"])
	}
	detail, ok := raw["error"].(map[string]any)
	if !ok {
		t.Fatalf("error envelope missing or not an object: %v", raw["error"])
	}
	if detail["code"] != "task_not_exist" || detail["message"] != "任务不存在" {
		t.Errorf("error envelope content mismatch: %v", detail)
	}
	if detail["type"] != common.AppErrorType() {
		t.Errorf("error type = %v, want %v", detail["type"], common.AppErrorType())
	}
}

// TestTaskErrorMarshalHidesInternals 内部流转字段不得出现在响应里。
func TestTaskErrorMarshalHidesInternals(t *testing.T) {
	taskErr := &TaskError{Code: "failed", Message: "boom", StatusCode: 500, LocalError: true}
	data, err := common.Marshal(taskErr)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	var raw map[string]any
	if err := common.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	for _, hidden := range []string{"LocalError", "Error"} {
		if _, leaked := raw[hidden]; leaked {
			t.Errorf("internal field %q leaked into response: %s", hidden, data)
		}
	}
}

// TestTaskErrorMarshalWithoutStatusCode 未设置 StatusCode 时不输出 status_code，
// 避免给客户端一个误导性的 0。
func TestTaskErrorMarshalWithoutStatusCode(t *testing.T) {
	data, err := common.Marshal(&TaskError{Code: "failed", Message: "boom"})
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	var raw map[string]any
	if err := common.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if _, exists := raw["status_code"]; exists {
		t.Errorf("status_code should be omitted when unset: %s", data)
	}
}
