package service

import (
	"github.com/gin-gonic/gin"
	cartpb "bytedancemall/router/proto/cart"
	authpb "bytedancemall/router/proto/auth"
	userpb "bytedancemall/router/proto/user"
	productpb "bytedancemall/router/proto/product"
	paymentpb "bytedancemall/router/proto/payment"
	orderpb "bytedancemall/router/proto/order"
	"bytedancemall/router/model"
	"google.golang.org/grpc"
	"encoding/json"
	"os"
)

type RouterService struct {
	CartClient cartpb.CartServiceClient
	AuthClient authpb.AuthServiceClient
	UserClient userpb.UserServiceClient
	ProductClient productpb.ProductCatalogServiceClient
	PaymentClient paymentpb.PaymentServiceClient
	OrderClient orderpb.OrderServiceClient
	Router *gin.Engine
}

// 认证服务中间件
func (r *RouterService) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取 Token
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// 调用认证服务的 VerifyToken 方法验证 Token
		resp, err := r.AuthClient.VerifyTokenByRPC(c, &authpb.VerifyTokenReq{Token: token})
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// 如果认证失败，返回错误响应
		if !resp.Res {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}


	}
}

func (r *RouterService) InitService () {
	// 读取json
	file, err := os.Open("configs.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	configs := model.Configs{}
	err = decoder.Decode(&configs)
	if err != nil {
		panic(err)
	}

	{
		conn, err := grpc.NewClient(configs.AuthService.Host)
		if err != nil {
			panic(err)
		}
		r.AuthClient = authpb.NewAuthServiceClient(conn)
	}
	
	{
		conn, err := grpc.NewClient(configs.CartService.Host)
		if err != nil {
			panic(err)
		}
		r.CartClient = cartpb.NewCartServiceClient(conn)
	}

	{
		conn, err := grpc.NewClient(configs.OrderService.Host)
		if err != nil {
			panic(err)
		}
		r.OrderClient = orderpb.NewOrderServiceClient(conn)
	}
	
	{
		conn, err := grpc.NewClient(configs.PaymentService.Host)
		if err != nil {
			panic(err)
		}
		r.PaymentClient = paymentpb.NewPaymentServiceClient(conn)
	}

	{
		conn, err := grpc.NewClient(configs.ProductService.Host)
		if err != nil {
			panic(err)
		}
		r.ProductClient = productpb.NewProductCatalogServiceClient(conn)
	}

	{
		conn, err := grpc.NewClient(configs.UserService.Host)
		if err != nil {
			panic(err)
		}
		r.UserClient = userpb.NewUserServiceClient(conn)
	}
}

func (r *RouterService) InitRouter() {
	// 初始化 Gin 路由
    r.Router = gin.Default()

    // 用户相关路由
    userGroup := r.Router.Group("/users")
    {
        userGroup.POST("/login", HandleLogin)
		userGroup.POST("/register", HandleRegister)
    }

    // 商品相关路由
    productGroup := r.Router.Group("/products")
    {
        productGroup.GET("/", HandleListProducts)
        productGroup.GET("/:id", HandleGetProduct)
        productGroup.GET("/search", HandleSearchProducts)
    }

    // 购物车相关路由
    cartGroup := r.Router.Group("/cart", r.AuthMiddleware())
    {
        cartGroup.GET("/", HandleGetCart)
        cartGroup.POST("/items", HandleAddCartItem)
        cartGroup.DELETE("/items/:id", HandleRemoveCartItem)
    }

    // 订单相关路由
    orderGroup := r.Router.Group("/orders", r.AuthMiddleware())
    {
        orderGroup.POST("", HandleCreateOrder)
        orderGroup.GET("", HandleListOrders)
        orderGroup.GET("/:id", HandleGetOrder)
    }

    // 支付相关路由
    paymentGroup := r.Router.Group("/payments", r.AuthMiddleware())
    {
        paymentGroup.POST("", HandleCreatePayment)
        paymentGroup.GET("/:id", HandleGetPaymentStatus)
    }
}

func InitRouterService() *RouterService{
	var r RouterService
	r.InitService()
	r.InitRouter()
	return &r
}