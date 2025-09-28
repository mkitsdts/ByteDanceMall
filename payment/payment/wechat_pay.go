package payment

import (
	"bytedancemall/payment/config"
	"bytedancemall/payment/model"
	"bytedancemall/payment/pkg/database"
	"bytedancemall/payment/util"
	"context"
	"crypto/rsa"
	"fmt"
	"log/slog"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

type wechat struct {
	mchPrivateKey *rsa.PrivateKey
}

var w = &wechat{}

func init() {
	var err error
	w.mchPrivateKey, err = utils.LoadPrivateKeyWithPath(config.Cfg.Payment.WeChat.PrimaryKeyPath)
	if err != nil {
		slog.Error("load merchant private key error", "error", err)
		panic(err)
	}
}

func (w *wechat) wechat_pay(ctx context.Context, req *PaymentRequest) string {
	select {
	case <-ctx.Done():
		return ""
	default:
	}

	// 使用商户私钥等初始化 client，并使它具有自动定时获取微信支付平台证书的能力
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(config.Cfg.Payment.WeChat.MchID, config.Cfg.Payment.WeChat.MchCertificateSerialNumber, w.mchPrivateKey, config.Cfg.Payment.WeChat.MchAPIv3Key),
	}
	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		slog.Error("new wechat pay client error:", " ", err)
	}
	svc := native.NativeApiService{Client: client}
	// 发送请求
	for i := range 3 {
		resp, result, err := svc.Prepay(ctx,
			native.PrepayRequest{
				Appid:       core.String(config.Cfg.Payment.WeChat.AppID),
				Mchid:       core.String(config.Cfg.Payment.WeChat.MchID),
				Description: core.String(req.Description),
				OutTradeNo:  core.String(fmt.Sprint(req.OrderID)),
				Attach:      core.String(req.Attach),
				NotifyUrl:   core.String(config.Cfg.Payment.WeChat.NotifyURL),
				Amount: &native.Amount{
					Total: core.Int64(100),
				},
				TimeExpire: core.Time(time.Now().Add(5 * time.Minute)),
			},
		)
		if err != nil {
			time.Sleep(10 << i * time.Millisecond)
			continue
		}
		if result.Response.StatusCode != 200 {
			slog.Error("wechat prepay request failed", "status", result.Response.StatusCode, "body", result.Response.Body)
			return ""
		}
		if resp.CodeUrl == nil || *resp.CodeUrl == "" {
			slog.Error("wechat prepay response missing code_url")
			return ""
		}
		for i := range 3 {
			if err = database.DB().Table("payment_records").Create(&model.PaymentRecord{
				PaymentID: util.GenerateUUID(),
				OrderID:   req.OrderID,
				Method:    "wechat",
				Status:    model.CREATED,
				OrderStr:  resp.CodeUrl,
			}).Error; err != nil {
				slog.Error("Failed to create payment record, retrying...", "error", err)
				time.Sleep(10 << i * time.Millisecond)
			}
		}
		if err != nil {
			slog.Error("Failed to create payment record", "error", err)
			return ""
		}
		return *resp.CodeUrl
	}
	slog.Error("failed to create wechat prepay order after retries")
	return ""
}
