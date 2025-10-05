package payment

import (
	"bytedancemall/payment/pkg/database"
	kfk "bytedancemall/payment/pkg/kafka"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
)

type Resource struct {
	Algorithm  string `json:"algorithm"`
	Ciphertext string `json:"ciphertext"`
	Nonce      string `json:"nonce"`
}

type Body struct {
	ID           string    `json:"id"`
	CreateTime   time.Time `json:"create_time"`
	ResourceType string    `json:"resource_type"`
	EventType    string    `json:"event_type"`
	Summary      string    `json:"summary"`
	Resource     Resource  `json:"resource"`
}

// 这里处理微信支付的异步通知
func HandleWechatPaymentNotify(c *gin.Context) {
	// 但这里的逻辑处理起来非常复杂，容我摸索清楚再写
	var body Body
	if err := c.BindJSON(&body); err != nil {
		slog.Error("bind json error", "error", err)
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	if body.EventType == "TRANSACTION.SUCCESS" {
		// 修改唯一订单状态为已支付
		// 绑定订单和支付记录
		database.DB().Table("payment_processes").Where("order_id = ?", body.Summary).Update("status", "paid").Update("id", body.ID)
		database.DB().Table("payment_records").Where("order_id = ?", body.Summary).Update("status", "paid")
		kfk.GetWriter("payment_success").WriteMessages(c, kafka.Message{
			Key:   []byte(body.ID),
			Value: []byte(body.Summary),
		})
		// 这里还要发个消息给订单服务，告诉它支付成功了
		slog.Info("payment success", "id", body.ID, "summary", body.Summary)
		c.JSON(200, gin.H{"status": "success"})
	}
	c.JSON(200, gin.H{"status": "ignored"})
}
