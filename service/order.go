package service

import (
	"github.com/gin-gonic/gin"
	"bytedancemall/router/model"
	"context"
	orderpb "bytedancemall/router/proto/order"
)

func (s *RouterService)HandleCreateOrder(c *gin.Context) {

	var req model.Order
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}

	// 构造请求参数
	var orderItems []*orderpb.OrderItem
	for i := 0; i < len(req.Items); i++ {
		orderItems = append(orderItems, &orderpb.OrderItem{
			Cost: req.Items[i].Cost,
			Item: &orderpb.CartItem{
				ProductId: req.Items[i].ProductId,
				Quantity: req.Items[i].Quantity,
			},
		})
	}
	
	resp , err := s.OrderClient.PlaceOrder(context.Background(), &orderpb.PlaceOrderReq{
		UserId: req.UserId,
		Address: &orderpb.Address{
			Country: req.Address.Country,
			StreetAddress: req.Address.StreetAddress,
			State: req.Address.State,
			City: req.Address.City,
			ZipCode: req.Address.ZipCode,
		},
		OrderItems: orderItems,
	})

	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"order_id": resp.Order.OrderId})
}

func  (s *RouterService)HandleListOrders(c *gin.Context) {
	userId , _ := c.Get("user_id")
	resp , err := s.OrderClient.ListOrder(context.Background(), &orderpb.ListOrderReq{
		UserId: userId.(uint32),
	})
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}

	var orders []model.Order
	for i := 0; i < len(resp.Orders); i++ {
		var orderItems []model.OrderItem
		for j := 0; j < len(resp.Orders[i].OrderItems); j++ {
			orderItems = append(orderItems, model.OrderItem{
				ProductId: resp.Orders[i].OrderItems[j].Item.ProductId,
				Quantity:  resp.Orders[i].OrderItems[j].Item.Quantity,
				Cost: resp.Orders[i].OrderItems[j].Cost,
			})
		}
		orders = append(orders, model.Order{
			Items: orderItems,
			Address: model.Address{
				Country: resp.Orders[i].Address.Country,
				StreetAddress: resp.Orders[i].Address.StreetAddress,
				State: resp.Orders[i].Address.State,
				City: resp.Orders[i].Address.City,
				ZipCode: resp.Orders[i].Address.ZipCode,
			},
			UserId: resp.Orders[i].UserId,
			UserCurrency: resp.Orders[i].UserCurrency,
			Email: resp.Orders[i].Email,
		})
	}

	c.JSON(200, gin.H{"orders": orders})
}