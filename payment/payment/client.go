package payment

import (
	"bytedancemall/payment/config"
	"context"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type PaymentClient struct {
	clients         sync.Pool
	router          *gin.Engine
	templateRequest CreatePaymentRequestBody
}

func NewPaymentClient() *PaymentClient {
	pc := &PaymentClient{}

	pc.initRouter()
	pc.clients = sync.Pool{
		New: func() any {
			return &http.Client{}
		},
	}

	pc.templateRequest = CreatePaymentRequestBody{
		AppID:         config.Cfg.Payment.AppID,
		MachineID:     config.Cfg.Payment.MachineID,
		NotifyURL:     config.Cfg.Server.Host + PaymentNotifyURL,
		SupportFaPiao: config.Cfg.Payment.SupportFaPiao,
	}

	return pc
}

// 返回创建支付请求的URL
func (pc *PaymentClient) CreatePaymentRequest(ctx context.Context, order_id string) string {
	select {
	case <-ctx.Done():
		return ""
	default:
	}
	return "http://payment.example.com/create?order_id=" + order_id
}
