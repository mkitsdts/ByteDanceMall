package api

import (
	"bytedancemall/router/pkg/user"
	userpb "bytedancemall/router/proto/user"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	type LoginRequest struct {
		UserId   uint64 `json:"user_id" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	resp, err := user.GetClient().Login(c, &userpb.LoginReq{
		UserId:   req.UserId,
		Password: req.Password,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to login"})
		return
	}
	c.JSON(200, resp)
}
