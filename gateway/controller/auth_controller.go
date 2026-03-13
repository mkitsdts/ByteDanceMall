package controller

import (
	"bytedancemall/gateway/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	userService *service.UserService
}

func NewAuthController(userService *service.UserService) *AuthController {
	return &AuthController{userService: userService}
}

type loginRequest struct {
	Email    string `json:"email"`
	UserID   uint64 `json:"user_id"`
	Password string `json:"password" binding:"required"`
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req loginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	resp, err := c.userService.Login(ctx.Request.Context(), service.LoginInput{
		Email:    req.Email,
		UserID:   req.UserID,
		Password: req.Password,
	})
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"message": "login failed"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"result": resp.GetResult(),
		"token":  resp.GetToken(),
	})
}

func (c *AuthController) RedirectTarget(ctx *gin.Context) {
	ctx.JSON(http.StatusUnauthorized, gin.H{"message": "auth required"})
}
