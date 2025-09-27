package service

import (
	"bytedancemall/router/config"
	"bytedancemall/router/pkg/router"
	"bytedancemall/router/service/api"
	"fmt"

	"github.com/gin-gonic/gin"
)

type RouterService struct {
	router *gin.Engine
}

func NewRouterService() *RouterService {
	s := &RouterService{}
	s.router = router.New()

	apiGroup := s.router.Group("/api")
	{
		apiGroup.POST("/login", api.Login) // 用户登录
	}

	return s
}

func (s *RouterService) Run() error {
	s.router.Run(fmt.Sprintf(":%d", config.Get().Server.Port))
	return nil
}
