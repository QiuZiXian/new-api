package relay

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaskErrorFromUpstream 守护「上游 code 透传」：客户端要能按上游业务 code
// 分支处理（如余额不足 vs 参数非法 vs 取任务失败），而不是一律 fail_to_fetch_task。
func TestTaskErrorFromUpstream(t *testing.T) {
	cases := []struct {
		name        string
		statusCode  int
		body        string
		wantCode    string
		wantMessage string
	}{
		{
			name:        "openai style error envelope",
			statusCode:  http.StatusPaymentRequired,
			body:        `{"error":{"code":"InsufficientBalance","message":"需要 100，余额 6.0000"}}`,
			wantCode:    "InsufficientBalance",
			wantMessage: "需要 100，余额 6.0000",
		},
		{
			name:        "flat code field",
			statusCode:  http.StatusBadRequest,
			body:        `{"code":"InvalidParameter","message":"resolution is invalid"}`,
			wantCode:    "InvalidParameter",
			wantMessage: "resolution is invalid",
		},
		{
			name:        "numeric code is stringified",
			statusCode:  http.StatusTooManyRequests,
			body:        `{"error":{"code":429,"message":"rate limited"}}`,
			wantCode:    "429",
			wantMessage: "rate limited",
		},
		{
			name:        "non-json body keeps fallback code and raw message",
			statusCode:  http.StatusBadGateway,
			body:        "upstream gateway error",
			wantCode:    "fail_to_fetch_task",
			wantMessage: "upstream gateway error",
		},
		{
			name:        "empty body falls back to status description",
			statusCode:  http.StatusNotFound,
			body:        "",
			wantCode:    "fail_to_fetch_task",
			wantMessage: "upstream returned http 404",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			taskErr := taskErrorFromUpstream(tc.statusCode, []byte(tc.body))
			require.NotNil(t, taskErr)
			assert.Equal(t, tc.wantCode, taskErr.Code)
			assert.Equal(t, tc.wantMessage, taskErr.Message)
			assert.Equal(t, tc.statusCode, taskErr.StatusCode)
			assert.False(t, taskErr.LocalError)
		})
	}
}
