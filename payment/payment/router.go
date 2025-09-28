package payment

import "github.com/gin-gonic/gin"

func (pc *PaymentClient) initRouter() {
	pc.router = gin.Default()
	pc.router.POST("/payment/wechat/notify", HandleWechatPaymentNotify)
	go pc.router.Run(":8080")
}
