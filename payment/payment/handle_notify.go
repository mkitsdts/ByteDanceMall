package payment

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
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
		// 处理支付成功的逻辑，例如更新订单状态
		slog.Info("payment success", "id", body.ID, "summary", body.Summary)
		c.JSON(200, gin.H{"status": "success"})
	}
	c.JSON(200, gin.H{"status": "ignored"})
}
