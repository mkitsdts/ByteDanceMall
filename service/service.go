package service

import (
	pb "bytedancemall/payment/proto"
	"context"
	"fmt"
	"github.com/google/uuid"
	"bytedancemall/payment/utils"
)

type PaymentService struct {
	pb.UnimplementedPaymentServiceServer
}

// 创建一个新的用户服务实例
func NewPaymentService() *PaymentService {
	return &PaymentService{}
}

// 支付服务的Charge方法
func (s *PaymentService) Charge(ctx context.Context, in *pb.ChargeReq) (*pb.ChargeResp, error) {
	// 获取信用卡信息
	cvv := in.CreditCard.CreditCardCvv
	cardNumber := in.CreditCard.CreditCardNumber
	expirationYear := in.CreditCard.CreditCardExpirationYear
	expirationMonth := in.CreditCard.CreditCardExpirationMonth
	// 判断信用卡信息是否合法
	if utils.CheckCredit(cvv, cardNumber, expirationYear, expirationMonth) {
		// 生成支付成功的响应
		resp := pb.ChargeResp{
			TransactionId: uuid.New().String(),
		}
		return &resp, nil
	} else {
		// 生成支付失败的响应
		return &pb.ChargeResp{
			TransactionId: "ERROR",
		}, fmt.Errorf("Charge failed")
	}

}