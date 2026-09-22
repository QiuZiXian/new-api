package middleware

import (
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/gin-gonic/gin"
)

func abortWithOpenAiMessage(c *gin.Context, statusCode int, message string, code ...types.ErrorCode) {
	codeStr := ""
	if len(code) > 0 {
		codeStr = string(code[0])
	}
	userId := c.GetInt("id")
	// 统一错误体：顶层扁平字段 + OpenAI 风格 error 包裹。
	c.JSON(statusCode, common.NewErrorBody(statusCode, codeStr, common.MessageWithRequestId(message, c.GetString(common.RequestIdKey))))
	c.Abort()
	logger.LogError(c.Request.Context(), fmt.Sprintf("user %d | %s", userId, message))
}

// abortWithMidjourneyMessage 沿用 Midjourney 渠道自己的 {description,type,code}
// 契约——它的调用方是按该格式解析的，不并入统一错误体。
func abortWithMidjourneyMessage(c *gin.Context, statusCode int, code int, description string) {
	c.JSON(statusCode, gin.H{
		"description": description,
		"type":        common.AppErrorType(),
		"code":        code,
	})
	c.Abort()
	logger.LogError(c.Request.Context(), description)
}
