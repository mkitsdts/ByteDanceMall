package app

import (
	"bytedancemall/gateway/config"
	"bytedancemall/gateway/controller"
	"bytedancemall/gateway/middleware"
	"bytedancemall/gateway/repository"
	"bytedancemall/gateway/router"
	"bytedancemall/gateway/service"
	"fmt"

	"github.com/gin-gonic/gin"
)

type Server struct {
	engine *gin.Engine
}

func NewServer() *Server {
	authRepo := repository.NewAuthRepository(config.Get().Auth)
	userRepo := repository.NewUserRepository(config.Get().User)
	orderRepo := repository.NewOrderRepository(config.Get().Order)
	productRepo := repository.NewProductRepository(config.Get().Product)

	authService := service.NewAuthService(authRepo)
	userService := service.NewUserService(userRepo)
	orderService := service.NewOrderService(orderRepo)
	productService := service.NewProductService(productRepo)

	authController := controller.NewAuthController(userService)
	userController := controller.NewUserController(userService)
	orderController := controller.NewOrderController(orderService)
	productController := controller.NewProductController(productService)

	engine := router.New(
		middleware.RateLimit(config.Get().Server.LimiterRate, config.Get().Server.LimiterWindow),
		middleware.AuthRequired(authService),
		authController,
		userController,
		orderController,
		productController,
	)

	return &Server{engine: engine}
}

func (s *Server) Run() error {
	port := config.Get().Server.Port
	if port == 0 {
		port = 8080
	}
	return s.engine.Run(fmt.Sprintf(":%d", port))
}
