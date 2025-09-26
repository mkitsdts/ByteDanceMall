package payment

type Amount struct {
	Total    int    `json:"total"`    // 金额，单位为分
	Currency string `json:"currency"` // 货币类型
}

type GoodsDetail struct {
	MerchantGoodsID  string `json:"merchant_goods_id"`  // 商户侧商品编码
	WechatpayGoodsID string `json:"wechatpay_goods_id"` // 微信支付商品编码
	GoodsName        string `json:"goods_name"`         // 商品名称
	Quantity         int    `json:"quantity"`           // 商品数量
	UnitPrice        int    `json:"unit_price"`         // 商品单价，单位为分
}

type Detail struct {
	CostPrice   int           `json:"cost_price"`   // 原价
	InvoiceID   string        `json:"invoice_id"`   // 商家收据ID
	GoodsDetail []GoodsDetail `json:"goods_detail"` // 商品详情
}

type StoreInfo struct {
}

type H5Info struct {
	Type        string `json:"type"`         // H5支付场景
	AppName     string `json:"app_name"`     // H5应用名称
	AppURL      string `json:"app_url"`      // H5页面URL
	BundleID    string `json:"bundle_id"`    // IOS 平台 BundleID
	PackageName string `json:"package_name"` // Android 平台 PackageName
}

type SceneInfo struct {
	PayerClientIP string    `json:"payer_client_ip"` // 用户终端IP
	DeviceID      string    `json:"device_id"`       // 设备ID
	StoreInfo     StoreInfo `json:"store_info"`      // 门店信息
	H5Info        H5Info    `json:"h5_info"`         // H5场景信息
}

type SettleInfo struct {
	ProfitSharing bool `json:"profit_sharing"` // 是否分账
}

type CreatePaymentRequestBody struct {
	AppID         string     `json:"appid"`
	MachineID     string     `json:"mchid"`
	Description   string     `json:"description"`
	OutTradeNo    string     `json:"out_trade_no"`
	TimeExpire    string     `json:"time_expire"`    // 格式要求 2018-06-08T10:34:56+08:00
	Attach        string     `json:"attach"`         // 附加数据，在查询API和支付通知中原样返回，该字段主要用于商户携带订单的自定义数据
	NotifyURL     string     `json:"notify_url"`     // 通知地址
	GoodsTag      string     `json:"goods_tag"`      // 商品标记
	SupportFaPiao bool       `json:"support_fapiao"` // 是否支持发票
	Amount        Amount     `json:"amount"`
	Detail        Detail     `json:"detail"`
	SettleInfo    SettleInfo `json:"settle_info"`
}

type Resource struct {
	OriginType     string `json:"origin_type"`
	Algorithm      string `json:"algorithm"`
	CipherText     string `json:"cipher_text"`
	AssociatedData string `json:"associated_data"`
	Nonce          string `json:"nonce"`
}

type CreatePaymentResponseBody struct {
	ID           string   `json:"id"`
	CreateTime   string   `json:"create_time"`
	ResourceType string   `json:"resource_type"`
	EventType    string   `json:"event_type"`
	Summary      string   `json:"summary"`
	Resource     Resource `json:"resource"`
}
