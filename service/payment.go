package service

import (
	"bytedancemall/router/model"
	paymentpb "bytedancemall/router/proto/payment"
	"context"

	"github.com/gin-gonic/gin"
)

func (s *RouterService)HandleCreatePayment(c *gin.Context) {
	// 获取用户ID
	userId , _ := c.Get("user_id")
	// 获取请求参数
	var req model.Payment
	req.UserId = userId.(uint32)
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	
	// 创建支付
	resp , err := s.PaymentClient.Charge(context.Background(), &paymentpb.ChargeReq{
		UserId: req.UserId,
		OrderId: req.OrderId,
		Amount: req.Amount,
		CreditCard: &paymentpb.CreditCardInfo{
			CreditCardNumber: req.CreditCard.CreditCardNumber,
			CreditCardExpirationMonth: req.CreditCard.CreditCardExpirationMonth,
			CreditCardExpirationYear: req.CreditCard.CreditCardExpirationYear,
			CreditCardCvv: req.CreditCard.CreditCardCvv,
		},
	})

	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"transaction_id": resp.TransactionId})
}
