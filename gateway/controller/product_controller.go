package controller

import (
	"bytedancemall/gateway/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	productService *service.ProductService
}

func NewProductController(productService *service.ProductService) *ProductController {
	return &ProductController{productService: productService}
}

func (c *ProductController) List(ctx *gin.Context) {
	page := int32(1)
	if raw := ctx.Query("page"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || parsed <= 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid page"})
			return
		}
		page = int32(parsed)
	}

	pageSize := int64(20)
	if raw := ctx.Query("page_size"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid page_size"})
			return
		}
		pageSize = parsed
	}

	resp, err := c.productService.List(ctx.Request.Context(), service.ListProductsInput{
		Page:         page,
		PageSize:     pageSize,
		CategoryName: ctx.Query("category"),
	})
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"message": "list products failed"})
		return
	}

	products := make([]gin.H, 0, len(resp.GetProducts()))
	for _, product := range resp.GetProducts() {
		products = append(products, gin.H{
			"id":          product.GetId(),
			"name":        product.GetName(),
			"description": product.GetDescription(),
			"picture":     product.GetPicture(),
			"price":       product.GetPrice(),
			"categories":  product.GetCategories(),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"products": products})
}

func (c *ProductController) GetDetail(ctx *gin.Context) {
	productID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || productID == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid product id"})
		return
	}

	resp, err := c.productService.GetDetail(ctx.Request.Context(), productID)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"message": "query product failed"})
		return
	}

	product := resp.GetProduct()
	ctx.JSON(http.StatusOK, gin.H{
		"id":          product.GetId(),
		"name":        product.GetName(),
		"description": product.GetDescription(),
		"picture":     product.GetPicture(),
		"price":       product.GetPrice(),
		"categories":  product.GetCategories(),
	})
}
