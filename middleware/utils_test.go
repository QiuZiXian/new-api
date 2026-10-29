package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// 鉴权失败必须留下一条 type=error 的日志，否则线上只能靠 stdout 猜。
func TestAbortWithOpenAiMessageRecordsRelayRejection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	previousLogDB := model.LOG_DB
	previousEnabled := constant.ErrorLogEnabled
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Log{}))
	model.LOG_DB = db
	constant.ErrorLogEnabled = true
	t.Cleanup(func() {
		model.LOG_DB = previousLogDB
		constant.ErrorLogEnabled = previousEnabled
	})

	body := `{"model":"sora-2","prompt":"a cat","api_key":"sk-should-not-leak"}`
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/video/generations", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Authorization", "Bearer sk-invalid-token-value")
	c.Set(RouteTagKey, "relay")
	c.Set(common.RequestIdKey, "req-test-1")

	abortWithOpenAiMessage(c, http.StatusUnauthorized, "Invalid token")

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)

	var logs []model.Log
	require.NoError(t, db.Find(&logs).Error)
	require.Len(t, logs, 1, "鉴权失败应写入一条错误日志")

	record := logs[0]
	assert.Equal(t, model.LogTypeError, record.Type)
	assert.Equal(t, "sora-2", record.ModelName)
	assert.Equal(t, "req-test-1", record.RequestId)
	assert.Contains(t, record.Content, "Invalid token")

	assert.Contains(t, record.Other, "/v1/video/generations")
	assert.Contains(t, record.Other, "a cat", "请求参数应保留下来供排查")
	assert.NotContains(t, record.Other, "sk-should-not-leak", "请求体里的凭据必须脱敏")
	assert.NotContains(t, record.Other, "sk-invalid-token-value", "Authorization 必须脱敏")
}

// 管理端接口的鉴权失败不进 logs 表，避免噪音。
func TestAbortWithOpenAiMessageSkipsNonRelayRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	previousLogDB := model.LOG_DB
	previousEnabled := constant.ErrorLogEnabled
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Log{}))
	model.LOG_DB = db
	constant.ErrorLogEnabled = true
	t.Cleanup(func() {
		model.LOG_DB = previousLogDB
		constant.ErrorLogEnabled = previousEnabled
	})

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/user/self", nil)
	c.Set(RouteTagKey, "api")

	abortWithOpenAiMessage(c, http.StatusUnauthorized, "Invalid token")

	var count int64
	require.NoError(t, db.Model(&model.Log{}).Count(&count).Error)
	assert.Zero(t, count)
}
