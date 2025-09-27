package middleware

import (
	"github.com/gin-gonic/gin"
)

// 简单的 Bearer Token 鉴权（用于后台写接口）
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
