package middleware

import (
	"bytedancemall/gateway/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthRequired(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
		refreshToken := strings.TrimSpace(c.GetHeader("X-Refresh-Token"))

		if err := authService.Verify(c.Request.Context(), token, refreshToken); err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Next()
	}
}
