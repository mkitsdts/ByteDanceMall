package service

import (
	"github.com/gin-gonic/gin"
	"context"
	productpb "bytedancemall/router/proto/product"
	"bytedancemall/router/model"
	"strconv"
)

func (s *RouterService)HandleListProducts(c *gin.Context) {

	var page model.Page
	if err := c.BindJSON(page); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 调用 rpc 服务
	rsp, err := s.ProductClient.ListProducts(context.Background(), &productpb.ListProductsReq{
		Page: page.Page,
		PageSize: page.PageSize,
		CategoryName: page.CategoryName,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, rsp.Products)
}

func (s *RouterService)HandleGetProduct(c *gin.Context) {
	productId := c.Param("product_id")

	id , err := strconv.ParseUint(productId, 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	rsp, err := s.ProductClient.GetProduct(context.Background(), &productpb.GetProductReq{
		Id: uint32(id),
	})

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, rsp.Product)
}

func (s *RouterService)HandleSearchProducts(c *gin.Context) {
	var req model.SearchRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	rsp, err := s.ProductClient.SearchProducts(context.Background(), &productpb.SearchProductsReq{
		Query: req.Query,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, rsp.Results)
}