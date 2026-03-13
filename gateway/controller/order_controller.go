package controller

import (
	"bytedancemall/gateway/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OrderController struct {
	orderService *service.OrderService
}

func NewOrderController(orderService *service.OrderService) *OrderController {
	return &OrderController{orderService: orderService}
}

type createOrderRequest struct {
	UserID  uint64  `json:"user_id" binding:"required"`
	Product uint64  `json:"product_id" binding:"required"`
	Amount  uint64  `json:"amount" binding:"required"`
	OrderID uint64  `json:"order_id"`
	Cost    float32 `json:"cost"`
	Address struct {
		StreetAddress string `json:"street_address" binding:"required"`
		City          string `json:"city" binding:"required"`
		State         string `json:"state" binding:"required"`
	} `json:"address" binding:"required"`
}

func (c *OrderController) Create(ctx *gin.Context) {
	var req createOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	resp, err := c.orderService.Create(ctx.Request.Context(), service.CreateOrderInput{
		UserID:    req.UserID,
		ProductID: req.Product,
		Amount:    req.Amount,
		OrderID:   req.OrderID,
		Cost:      req.Cost,
		Street:    req.Address.StreetAddress,
		City:      req.Address.City,
		State:     req.Address.State,
	})
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"message": "create order failed"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"result": resp.GetResult()})
}
