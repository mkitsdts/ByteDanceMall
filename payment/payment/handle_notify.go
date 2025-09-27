package payment

import (
	"github.com/gin-gonic/gin"
)

type Body struct {
	OrderID uint64 `json:"order_id"`
	UserID  uint64 `json:"user_id"`
}

func HandlePaymentNotify(c *gin.Context) {

}
