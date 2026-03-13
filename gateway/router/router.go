package router

import (
	"bytedancemall/gateway/controller"

	"github.com/gin-gonic/gin"
)

func New(
	rateLimiter gin.HandlerFunc,
	authMiddleware gin.HandlerFunc,
	authController *controller.AuthController,
	userController *controller.UserController,
	orderController *controller.OrderController,
	productController *controller.ProductController,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.UseH2C = true
	engine.Use(gin.Logger(), gin.Recovery())
	if rateLimiter != nil {
		engine.Use(rateLimiter)
	}
	_ = engine.SetTrustedProxies([]string{"127.0.0.1"})

	engine.GET("/login", authController.RedirectTarget)

	v1 := engine.Group("/api/v1")
	v1.POST("/auth/login", authController.Login)

	protected := v1.Group("/")
	if authMiddleware != nil {
		protected.Use(authMiddleware)
	}
	protected.POST("/users/register", userController.Register)
	protected.POST("/orders", orderController.Create)
	protected.GET("/products", productController.List)
	protected.GET("/products/:id", productController.GetDetail)

	return engine
}
