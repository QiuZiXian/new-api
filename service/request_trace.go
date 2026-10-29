package service

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/gin-gonic/gin"
)

// 请求诊断日志：把「请求被拒绝 / 上游失败」时的现场完整记录下来。
//
// 两条输出通道：
//  1. stdout（docker logs / 日志文件）—— 完整详情，便于当场排查；
//  2. logs 表 type=error —— 结构化详情，便于事后按用户/令牌/request id 检索。
//
// 脱敏是硬约束：Authorization、api key、cookie 一律不落原文；body 中出现
// password/secret/token 一类字段同样打码。其余参数（model、prompt、size 等）
// 是排查所需，原样保留。

const (
	// 单个 body 最多记录 4KB，避免 base64 图片把日志撑爆
	traceBodyLimit = 4096
	// body 超过该大小就不读取，避免为记日志把大请求拖到磁盘
	traceBodyReadLimit = 1 << 20
	// 鉴权拒绝日志每分钟上限：无效 key 扫描会放大成写库洪峰，超限直接丢弃
	maxRejectedLogsPerMinute = 300
)

// rejectedLogWindow 按分钟统计鉴权拒绝日志条数，避免被无效 key 刷爆日志表。
var rejectedLogWindow struct {
	mu     sync.Mutex
	minute int64
	count  int
}

func allowRejectedLog() bool {
	minute := time.Now().Unix() / 60
	rejectedLogWindow.mu.Lock()
	defer rejectedLogWindow.mu.Unlock()
	if rejectedLogWindow.minute != minute {
		rejectedLogWindow.minute = minute
		rejectedLogWindow.count = 0
	}
	if rejectedLogWindow.count >= maxRejectedLogsPerMinute {
		return false
	}
	rejectedLogWindow.count++
	return true
}

// 这些 header 的值一律打码
var sensitiveHeaders = map[string]bool{
	"authorization":       true,
	"proxy-authorization": true,
	"x-api-key":           true,
	"api-key":             true,
	"apikey":              true,
	"x-goog-api-key":      true,
	"mj-api-secret":       true,
	"cookie":              true,
	"set-cookie":          true,
	"x-auth-token":        true,
	"x-access-token":      true,
}

// IsErrorLogEnabled 错误日志总开关：环境变量与系统设置同时开启才记录。
func IsErrorLogEnabled() bool {
	return constant.ErrorLogEnabled && operation_setting.GetGeneralSetting().ErrorLogEnabled
}

// IsRequestDetailLogEnabled 是否记录请求/上游的完整详情。
func IsRequestDetailLogEnabled() bool {
	return IsErrorLogEnabled() && operation_setting.GetGeneralSetting().ErrorLogRequestDetail
}

// maskSecret 对凭据类字符串打码，保留长度特征便于判断「有没有配错」。
func maskSecret(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return fmt.Sprintf("*** (len=%d)", len(value))
	}
	return fmt.Sprintf("%s*** (len=%d)", value[:4], len(value))
}

func isSensitiveFieldName(name string) bool {
	lower := strings.ToLower(name)
	for _, keyword := range []string{"password", "secret", "token", "api_key", "apikey", "credential", "signature", "access_key"} {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}

func isSensitiveQueryKey(key string) bool {
	switch strings.ToLower(key) {
	case "key", "api_key", "apikey", "token", "access_token", "secret", "sig", "signature":
		return true
	}
	return false
}

func maskSensitiveJSON(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		for key, item := range typed {
			if text, ok := item.(string); ok && isSensitiveFieldName(key) {
				typed[key] = maskSecret(text)
				continue
			}
			typed[key] = maskSensitiveJSON(item)
		}
		return typed
	case []any:
		for i := range typed {
			typed[i] = maskSensitiveJSON(typed[i])
		}
		return typed
	default:
		return value
	}
}

func truncateTraceText(text string) string {
	if len(text) <= traceBodyLimit {
		return text
	}
	return fmt.Sprintf("%s... [truncated, original_length=%d]", text[:traceBodyLimit], len(text))
}

// sanitizeQuery 打码后按 key 排序，保证同一请求的日志可对比。
func sanitizeQuery(query url.Values) map[string]string {
	keys := make([]string, 0, len(query))
	for key := range query {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make(map[string]string, len(keys))
	for _, key := range keys {
		joined := strings.Join(query[key], ",")
		if isSensitiveQueryKey(key) {
			joined = maskSecret(joined)
		}
		result[key] = truncateTraceText(joined)
	}
	return result
}

func sortedHeaderMap(headers http.Header) map[string]string {
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make(map[string]string, len(keys))
	for _, key := range keys {
		values := headers[key]
		joined := strings.Join(values, ",")
		if sensitiveHeaders[strings.ToLower(key)] {
			joined = maskSecret(joined)
		}
		result[strings.ToLower(key)] = truncateTraceText(joined)
	}
	return result
}

// sanitizeBody 把请求/响应体转成可安全落盘的文本。
// JSON 走字段级脱敏，其余按文本截断，二进制只记体积。
func sanitizeBody(body []byte, contentType string) string {
	if len(body) == 0 {
		return ""
	}
	lowerType := strings.ToLower(contentType)
	if strings.Contains(lowerType, "multipart/") || strings.Contains(lowerType, "octet-stream") || strings.Contains(lowerType, "image/") {
		return fmt.Sprintf("(binary body, %d bytes)", len(body))
	}
	var parsed any
	if common.Unmarshal(body, &parsed) == nil {
		if encoded, err := common.Marshal(maskSensitiveJSON(parsed)); err == nil {
			return truncateTraceText(string(encoded))
		}
	}
	return truncateTraceText(string(body))
}

// requestTrace 是一次请求的现场快照。
type requestTrace struct {
	Method   string            `json:"method"`
	Path     string            `json:"path"`
	Query    map[string]string `json:"query,omitempty"`
	ClientIP string            `json:"client_ip"`
	Headers  map[string]string `json:"headers,omitempty"`
	Body     string            `json:"body,omitempty"`
	Model    string            `json:"model,omitempty"`
}

func jsonString(value any) string {
	encoded, err := common.Marshal(value)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func (t requestTrace) String() string {
	parts := []string{fmt.Sprintf("request %s %s", t.Method, t.Path), "client_ip=" + t.ClientIP}
	if len(t.Query) > 0 {
		parts = append(parts, "query="+jsonString(t.Query))
	}
	if len(t.Headers) > 0 {
		parts = append(parts, "headers="+jsonString(t.Headers))
	}
	if t.Body != "" {
		parts = append(parts, "body="+t.Body)
	}
	return strings.Join(parts, " | ")
}

func (t requestTrace) other() map[string]interface{} {
	other := map[string]interface{}{
		"method":    t.Method,
		"path":      t.Path,
		"client_ip": t.ClientIP,
	}
	if len(t.Query) > 0 {
		other["query"] = t.Query
	}
	if len(t.Headers) > 0 {
		other["headers"] = t.Headers
	}
	if t.Body != "" {
		other["body"] = t.Body
	}
	return other
}

// captureRequestTrace 采集当前客户端请求的现场信息。
// body 读取失败或体积过大时降级为不含 body 的摘要，不影响主流程。
func captureRequestTrace(c *gin.Context) requestTrace {
	trace := requestTrace{ClientIP: c.ClientIP()}
	if c.Request == nil {
		return trace
	}
	trace.Method = c.Request.Method
	if c.Request.URL != nil {
		trace.Path = c.Request.URL.Path
		trace.Query = sanitizeQuery(c.Request.URL.Query())
	}
	trace.Headers = sortedHeaderMap(c.Request.Header)
	if !IsRequestDetailLogEnabled() {
		return trace
	}
	if c.Request.ContentLength > traceBodyReadLimit {
		trace.Body = fmt.Sprintf("(body skipped, %d bytes)", c.Request.ContentLength)
		return trace
	}
	storage, err := common.GetBodyStorage(c)
	if err != nil {
		trace.Body = fmt.Sprintf("(body read failed: %s)", err.Error())
		return trace
	}
	if _, err := storage.Seek(0, io.SeekStart); err != nil {
		trace.Body = fmt.Sprintf("(body seek failed: %s)", err.Error())
		return trace
	}
	body, err := io.ReadAll(io.LimitReader(storage, traceBodyReadLimit))
	if err != nil {
		trace.Body = fmt.Sprintf("(body read failed: %s)", err.Error())
		return trace
	}
	// 复位，避免影响后续仍要读 body 的调用方
	_, _ = storage.Seek(0, io.SeekStart)

	trace.Body = sanitizeBody(body, c.Request.Header.Get("Content-Type"))
	if trace.Model == "" {
		var payload struct {
			Model string `json:"model"`
		}
		if common.Unmarshal(body, &payload) == nil {
			trace.Model = payload.Model
		}
	}
	return trace
}

// RecordRejectedRequest 记录被中间件拦下的 relay 请求（鉴权/限流/无可用渠道等）。
// 这类请求从未进入 relay，计费日志里看不到，只有这里留痕。
func RecordRejectedRequest(c *gin.Context, statusCode int, code string, message string) {
	if !IsErrorLogEnabled() || !allowRejectedLog() {
		return
	}
	// 记日志绝不能反过来把请求打成 500
	defer func() {
		if recovered := recover(); recovered != nil {
			logger.LogError(c, fmt.Sprintf("record rejected request log panic: %v", recovered))
		}
	}()
	trace := captureRequestTrace(c)

	startTime := common.GetContextKeyTime(c, constant.ContextKeyRequestStartTime)
	if startTime.IsZero() {
		startTime = time.Now()
	}
	other := trace.other()
	other["stage"] = "middleware"
	other["reject_reason"] = code
	other["status_code"] = statusCode

	logger.LogError(c, fmt.Sprintf("request rejected (status=%d, reason=%s): %s | %s",
		statusCode, code, message, trace.String()))

	model.RecordErrorLog(c,
		c.GetInt("id"),
		c.GetInt("channel_id"),
		trace.Model,
		c.GetString("token_name"),
		fmt.Sprintf("%s | %s", message, truncateTraceText(trace.String())),
		c.GetInt("token_id"),
		int(time.Since(startTime).Seconds()),
		common.GetContextKeyBool(c, constant.ContextKeyIsStream),
		c.GetString("group"),
		other,
	)
}

// traceLoggedKey 标记当前请求的现场详情已经输出过，避免重试/多段错误处理重复刷屏。
const traceLoggedKey = "request_trace_logged"

// LogRelayRequestDetail 输出进入 relay 之后的请求现场（上游错误、任务提交失败等）。
// 同一个请求只输出一次；返回的结构可直接并入 logs.other，由调用方决定是否落库。
func LogRelayRequestDetail(c *gin.Context, stage string) map[string]interface{} {
	if !IsErrorLogEnabled() {
		return nil
	}
	// 该函数会在错误响应的 defer 中调用，panic 会导致响应体发不出去
	defer func() {
		if recovered := recover(); recovered != nil {
			logger.LogError(c, fmt.Sprintf("log relay request detail panic: %v", recovered))
		}
	}()
	trace := captureRequestTrace(c)
	other := trace.other()
	other["stage"] = stage
	if !c.GetBool(traceLoggedKey) {
		logger.LogError(c, fmt.Sprintf("%s error request detail | %s", stage, trace.String()))
		c.Set(traceLoggedKey, true)
	}
	return other
}

// LogUpstreamExchange 记录发往上游的请求与上游响应摘要。
// 只在请求失败或上游返回 4xx/5xx 时调用，正常链路不产生任何输出。
func LogUpstreamExchange(c *gin.Context, req *http.Request, statusCode int, respBody []byte, requestBody []byte) {
	if !IsRequestDetailLogEnabled() || req == nil || req.URL == nil {
		return
	}

	loggedURL := *req.URL
	sanitizedQuery := sanitizeQuery(loggedURL.Query())
	queryValues := make(url.Values, len(sanitizedQuery))
	for key, value := range sanitizedQuery {
		queryValues[key] = []string{value}
	}
	loggedURL.RawQuery = queryValues.Encode()

	parts := []string{
		fmt.Sprintf("upstream %s %s", req.Method, loggedURL.String()),
		"headers=" + jsonString(sortedHeaderMap(req.Header)),
	}
	if len(requestBody) > 0 {
		parts = append(parts, "body="+sanitizeBody(requestBody, req.Header.Get("Content-Type")))
	}
	parts = append(parts, fmt.Sprintf("status=%d", statusCode))
	if len(respBody) > 0 {
		parts = append(parts, "response="+sanitizeBody(respBody, ""))
	}
	logger.LogError(c, strings.Join(parts, " | "))
}
