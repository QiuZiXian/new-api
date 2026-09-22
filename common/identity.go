package common

import "strings"

// DefaultAppSlug 是网关对外的标识缩写（也成为响应头名和错误 type 的前缀）。
// 历史版本把这些值写死成 "new-ai"/"new_api"，这里把它收敛成唯一来源，
// 便于白标部署时统一替换。
const DefaultAppSlug = "new-api"

// AppSlug 是对外接口面上的产品标识，决定：
//   - 响应头名：X-<Slug>-Version、X-<Slug>-Other-Ratios
//   - 错误 type：<snake>_error、<snake>_panic
//
// 默认 DefaultAppSlug，此时对外输出与历史版本逐字节一致；管理员通过
// "AppSlug" 选项改写后才会变化。
var AppSlug = DefaultAppSlug

// normalizedAppSlug 返回可直接用于拼装的 slug；只含分隔符/空白时回退到默认值，
// 避免管理员误配导致拼出 X--Version 或 "_error" 这类畸形串。
func normalizedAppSlug() string {
	slug := strings.TrimSpace(AppSlug)
	slug = strings.Trim(slug, slugSeparatorChars)
	if slug == "" {
		return DefaultAppSlug
	}
	return slug
}

// slugSeparatorChars 是 slug 内部允许的分隔符集合，写法可互换：
// "new-api" / "new_api" / "New API" 都得到相同结果。
const slugSeparatorChars = "-_. "

// AppSlugSplitFields 把 slug 按分隔符切成字段：
// "new-api" / "new_api" / "New API" 都得到 ["new", "api"]。
// 空结果一律回退到默认 slug 的字段，确保任何配置都拼得出合法串。
func AppSlugSplitFields() []string {
	split := func(r rune) bool { return strings.ContainsRune(slugSeparatorChars, r) }
	fields := strings.FieldsFunc(normalizedAppSlug(), split)
	if len(fields) == 0 {
		return strings.FieldsFunc(DefaultAppSlug, split)
	}
	return fields
}

// AppSlugHeaderName 返回 slug 的 HTTP 头写法，用于拼出 X-<Name>-Version 这类头名。
// "new-api" -> "New-Api"（与历史上的 X-New-Api-Version 保持一致）。
func AppSlugHeaderName() string {
	fields := AppSlugSplitFields()
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		lower := strings.ToLower(field)
		if lower == "" {
			continue
		}
		parts = append(parts, strings.ToUpper(lower[:1])+lower[1:])
	}
	return strings.Join(parts, "-")
}

// AppSlugSnake 返回 slug 的下划线写法，用于拼错误 type。
// "new-api" -> "new_api"。
func AppSlugSnake() string {
	return strings.Join(AppSlugSplitFields(), "_")
}

// AppErrorType 返回本地错误响应里的 type 值；默认 "new_api_error"。
func AppErrorType() string {
	return AppSlugSnake() + "_error"
}

// AppPanicType 返回 panic 兜底响应里的 type 值；默认 "new_api_panic"。
func AppPanicType() string {
	return AppSlugSnake() + "_panic"
}
