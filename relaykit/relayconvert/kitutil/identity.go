package kitutil

import (
	"strings"
	"sync/atomic"
)

// The gateway's identity leaks into outbound payloads through error type
// strings (e.g. "new_api_error"). The host mirrors its configured app slug here
// at startup; standalone relaykit users keep the historical default.

const defaultIdentitySlug = "new_api"

var identitySlug atomic.Pointer[string]

// SetIdentitySlug mirrors the host's configured app slug into the kit. Empty or
// separator-only values fall back to DefaultIdentitySlug so callers never emit
// a malformed type like "_error".
func SetIdentitySlug(slug string) {
	normalized := normalizeIdentitySlug(slug)
	identitySlug.Store(&normalized)
}

// DefaultIdentitySlug returns the slug used until the host injects one.
func DefaultIdentitySlug() string { return defaultIdentitySlug }

// IdentitySlug returns the current identity slug in snake_case.
func IdentitySlug() string {
	if slug := identitySlug.Load(); slug != nil {
		return *slug
	}
	return defaultIdentitySlug
}

// ErrorTypeSlug returns the type value for gateway-local errors, e.g. "new_api_error".
func ErrorTypeSlug() string { return IdentitySlug() + "_error" }

func normalizeIdentitySlug(slug string) string {
	slug = strings.Trim(slug, "-_. ")
	slug = strings.ToLower(slug)
	slug = strings.NewReplacer("-", "_", " ", "_", ".", "_").Replace(slug)
	if slug == "" {
		return defaultIdentitySlug
	}
	return slug
}
