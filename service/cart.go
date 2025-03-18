package service

import (
	"github.com/gin-gonic/gin"
	"bytedancemall/router/model"
	"context"
	cartpb "bytedancemall/router/proto/cart"
)

var jwtKey = []byte("adgihioasxbfjkcbAEWIOFGHBIOHasegfWEAWEgWEARx")

// 获取购物车内容
func (s *RouterService)HandleGetCart(c *gin.Context) {
	// 获取用户ID
	userId , _ := c.Get("user_id")
	// 获取购物车内容
	resp , err:= s.CartClient.GetCart(context.Background(), &cartpb.GetCartReq{
		UserId: userId.(uint32),
	})
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}

	var cart model.Cart
	for i := 0; i < len(resp.Cart.Items); i++ {
		cart.Items = append(cart.Items, model.CartItem{
			ProductId: resp.Cart.Items[i].ProductId,
			Quantity:  resp.Cart.Items[i].Quantity,
		})
	}

}
func (s *RouterService)HandleAddCartItem(c *gin.Context) {
	// 获取用户ID
	userId , _ := c.Get("user_id")
	// 获取请求参数
	var req model.CartItem
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}

	// 添加购物车内容
	_, err := s.CartClient.AddItem(context.Background(), &cartpb.AddItemReq{
		UserId:    userId.(uint32),
		Item: &cartpb.CartItem{
			ProductId: req.ProductId,
			Quantity:  req.Quantity,
		},
	})

	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "success"})
}

func (s *RouterService)HandleRemoveCartItem(c *gin.Context) {
	// 获取用户ID
	userId , _ := c.Get("user_id")

	// 删除购物车内容
	_, err := s.CartClient.EmptyCart(context.Background(), &cartpb.EmptyCartReq{
		UserId:    userId.(uint32),
	})

	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "success"})
}