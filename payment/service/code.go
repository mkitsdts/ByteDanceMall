package service

const (
	WECHAT_PAY = 1
	ALIPAY     = 2
	PAYPAL     = 3
)

const (
	WAITTING   = 1
	PROCESSING = 2
	COMPLETED  = 3
	FAILED     = 4
)

func getPaymentMethodName(method int32) string {
	switch method {
	case WECHAT_PAY:
		return "WECHAT_PAY"
	case ALIPAY:
		return "ALIPAY"
	case PAYPAL:
		return "PAYPAL"
	default:
		return "UNKNOWN"
	}
}
