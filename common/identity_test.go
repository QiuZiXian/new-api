package common

import (
	"testing"
)

func TestAppSlugDerivationsDefault(t *testing.T) {
	defer func(slug string) { AppSlug = slug }(AppSlug)

	AppSlug = DefaultAppSlug
	if got := AppSlugHeaderName(); got != "New-Api" {
		t.Errorf("AppSlugHeaderName() = %q, want %q", got, "New-Api")
	}
	if got := AppSlugSnake(); got != "new_api" {
		t.Errorf("AppSlugSnake() = %q, want %q", got, "new_api")
	}
	if got := AppErrorType(); got != "new_api_error" {
		t.Errorf("AppErrorType() = %q, want %q", got, "new_api_error")
	}
	if got := AppPanicType(); got != "new_api_panic" {
		t.Errorf("AppPanicType() = %q, want %q", got, "new_api_panic")
	}
}

// TestAppSlugDerivationsCustom 验证白标部署改配置后，响应头名与错误 type 一起变。
func TestAppSlugDerivationsCustom(t *testing.T) {
	defer func(slug string) { AppSlug = slug }(AppSlug)

	AppSlug = "cii_group"
	if got := AppSlugHeaderName(); got != "Cii-Group" {
		t.Errorf("AppSlugHeaderName() = %q, want %q", got, "Cii-Group")
	}
	if got := AppSlugSnake(); got != "cii_group" {
		t.Errorf("AppSlugSnake() = %q, want %q", got, "cii_group")
	}
	if got := AppErrorType(); got != "cii_group_error" {
		t.Errorf("AppErrorType() = %q, want %q", got, "cii_group_error")
	}
}

// TestAppSlugDerivationsFallback 保证管理员误配成空串时不会拼出畸形头名或 "_error"。
func TestAppSlugDerivationsFallback(t *testing.T) {
	defer func(slug string) { AppSlug = slug }(AppSlug)

	for _, broken := range []string{"", "   ", "--", "_ . "} {
		AppSlug = broken
		if got := AppSlugHeaderName(); got != "New-Api" {
			t.Errorf("AppSlugHeaderName() with %q = %q, want %q", broken, got, "New-Api")
		}
		if got := AppErrorType(); got != "new_api_error" {
			t.Errorf("AppErrorType() with %q = %q, want %q", broken, got, "new_api_error")
		}
	}
}

// TestNewErrorBodyShape 验证统一错误体是「扁平 + error 包裹」双写。
func TestNewErrorBodyShape(t *testing.T) {
	body := NewErrorBody(404, ErrorCodeNotFound, "任务不存在")
	data, err := Marshal(body)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	var raw map[string]any
	if err := Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if raw["code"] != ErrorCodeNotFound {
		t.Errorf("top-level code = %v, want %v", raw["code"], ErrorCodeNotFound)
	}
	if raw["message"] != "任务不存在" {
		t.Errorf("top-level message = %v", raw["message"])
	}
	if raw["status_code"] != float64(404) {
		t.Errorf("top-level status_code = %v, want 404", raw["status_code"])
	}
	detail, ok := raw["error"].(map[string]any)
	if !ok {
		t.Fatalf("error envelope missing or not an object: %v", raw["error"])
	}
	if detail["code"] != ErrorCodeNotFound || detail["message"] != "任务不存在" {
		t.Errorf("error envelope content mismatch: %v", detail)
	}
	if detail["type"] != AppErrorType() {
		t.Errorf("error type = %v, want %v", detail["type"], AppErrorType())
	}
}

// TestDefaultCodeForStatus 保证同一 HTTP 状态在各接口回落到同一个 code。
func TestDefaultCodeForStatus(t *testing.T) {
	cases := map[int]string{
		400: ErrorCodeBadRequest,
		401: ErrorCodeUnauthorized,
		403: ErrorCodeForbidden,
		404: ErrorCodeNotFound,
		409: ErrorCodeConflict,
		429: ErrorCodeRateLimited,
		500: ErrorCodeInternal,
		503: ErrorCodeInternal,
		501: ErrorCodeNotImplemented,
	}
	for status, want := range cases {
		if got := DefaultCodeForStatus(status); got != want {
			t.Errorf("DefaultCodeForStatus(%d) = %q, want %q", status, got, want)
		}
	}
}
