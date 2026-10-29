package service

import (
	"strings"
	"testing"
)

func TestMaskSecret(t *testing.T) {
	if got := maskSecret(""); got != "" {
		t.Fatalf("empty secret should stay empty, got %q", got)
	}
	if got := maskSecret("short"); !strings.Contains(got, "***") {
		t.Fatalf("short secret should be masked, got %q", got)
	}
	got := maskSecret("sk-abcdefghijklmnop")
	if !strings.HasPrefix(got, "sk-a***") || !strings.Contains(got, "len=19") {
		t.Fatalf("unexpected masked secret: %q", got)
	}
}

func TestSanitizeBodyMasksCredentialsOnly(t *testing.T) {
	body := []byte(`{"api_key":"sk-1234567890abcdef","model":"sora-2","prompt":"a cat","nested":{"access_token":"abcdefghijk"}}`)
	got := sanitizeBody(body, "application/json")

	for _, leaked := range []string{"sk-1234567890abcdef", "abcdefghijk"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("credential %q leaked into log: %s", leaked, got)
		}
	}
	for _, kept := range []string{"sora-2", "a cat"} {
		if !strings.Contains(got, kept) {
			t.Fatalf("diagnostic field %q should be preserved: %s", kept, got)
		}
	}
}

func TestSanitizeBodyTruncatesLargeText(t *testing.T) {
	got := sanitizeBody([]byte(strings.Repeat("a", traceBodyLimit+100)), "text/plain")
	if !strings.Contains(got, "truncated") {
		t.Fatalf("large body should be truncated, got %d bytes", len(got))
	}
	if len(got) > traceBodyLimit+128 {
		t.Fatalf("truncated body too long: %d", len(got))
	}
}

func TestSanitizeBodyMarksBinary(t *testing.T) {
	got := sanitizeBody([]byte{1, 2, 3}, "multipart/form-data; boundary=x")
	if !strings.Contains(got, "binary body") {
		t.Fatalf("multipart body should be summarized, got %q", got)
	}
}

func TestSortedHeaderMapMasksAuthorization(t *testing.T) {
	headers := map[string][]string{
		"Authorization": {"Bearer sk-secret-token-value"},
		"Content-Type":  {"application/json"},
	}
	got := sortedHeaderMap(headers)
	if strings.Contains(got["authorization"], "sk-secret-token-value") {
		t.Fatalf("authorization header leaked: %q", got["authorization"])
	}
	if got["content-type"] != "application/json" {
		t.Fatalf("content-type should be preserved, got %q", got["content-type"])
	}
}

func TestSanitizeQueryMasksKeyParam(t *testing.T) {
	query := map[string][]string{
		"key":    {"AIza-secret"},
		"alt":    {"sse"},
		"filter": {"a,b"},
	}
	got := sanitizeQuery(query)
	if strings.Contains(got["key"], "AIza-secret") {
		t.Fatalf("query key leaked: %q", got["key"])
	}
	if got["alt"] != "sse" {
		t.Fatalf("non-sensitive query should be preserved, got %q", got["alt"])
	}
}
