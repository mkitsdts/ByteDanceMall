package router

import (
	"bytedancemall/router/config"
	"bytedancemall/router/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	r := gin.New()
	gin.SetMode(gin.ReleaseMode)
	r.UseH2C = true
	r.Use(gin.Logger(), gin.Recovery())
	// 启用限流器
	r.Use(middleware.RateLimit(config.Get().Server.LimiterRate, config.Get().Server.LimiterWindow))

	// 允许本地访问
	r.SetTrustedProxies([]string{"127.0.0.1"})

	return r
}
