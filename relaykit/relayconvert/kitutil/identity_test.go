package kitutil

import "testing"

func TestErrorTypeSlugDefault(t *testing.T) {
	if got := IdentitySlug(); got != DefaultIdentitySlug() {
		t.Errorf("IdentitySlug() = %q, want %q", got, DefaultIdentitySlug())
	}
	if got := ErrorTypeSlug(); got != "new_api_error" {
		t.Errorf("ErrorTypeSlug() = %q, want %q", got, "new_api_error")
	}
}

func TestSetIdentitySlug(t *testing.T) {
	// 清空即可回退默认，避免污染同包其它测试。
	defer SetIdentitySlug("")

	SetIdentitySlug("cii-group")
	if got := IdentitySlug(); got != "cii_group" {
		t.Errorf("IdentitySlug() = %q, want %q", got, "cii_group")
	}
	if got := ErrorTypeSlug(); got != "cii_group_error" {
		t.Errorf("ErrorTypeSlug() = %q, want %q", got, "cii_group_error")
	}
}

// TestSetIdentitySlugFallback 保证误配成空值时不会产出 "_error" 这类畸形 type。
func TestSetIdentitySlugFallback(t *testing.T) {
	defer SetIdentitySlug("")

	for _, broken := range []string{"", "  ", "--", "_._"} {
		SetIdentitySlug(broken)
		if got := ErrorTypeSlug(); got != "new_api_error" {
			t.Errorf("ErrorTypeSlug() after %q = %q, want %q", broken, got, "new_api_error")
		}
	}
}
