package payment

import (
	"context"

	"github.com/gin-gonic/gin"
)

type PaymentClient struct {
	w      *wechat
	router *gin.Engine
}

func NewPaymentClient() *PaymentClient {
	pc := &PaymentClient{}
	pc.initRouter()
	return pc
}

// 返回创建支付请求的URL
func (pc *PaymentClient) CreatePaymentRequest(ctx context.Context, method string, req *PaymentRequest) string {
	select {
	case <-ctx.Done():
		return ""
	default:
	}
	if method == "wechat" {
		return pc.w.wechat_pay(ctx, req)
	}
	return ""
}
