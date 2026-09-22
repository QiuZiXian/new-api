package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
)

func RelayPanicRecover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				common.SysLog(fmt.Sprintf("panic detected: %v", err))
				common.SysLog(fmt.Sprintf("stacktrace from panic: %s", string(debug.Stack())))
				// 与其它 /v1 接口同形：扁平字段 + error 包裹双写。
				c.JSON(http.StatusInternalServerError,
					common.NewErrorBodyWithType(http.StatusInternalServerError, common.ErrorCodeInternal,
						fmt.Sprintf("Panic detected, error: %v. Please submit a issue here: https://github.com/Calcium-Ion/new-api", err),
						common.AppPanicType()))
				c.Abort()
			}
		}()
		c.Next()
	}
}
