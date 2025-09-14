package service

import (
	rds "bytedancemall/auth/pkg/redis"
	pb "bytedancemall/auth/proto"
	"bytedancemall/auth/utils"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

func (s *AuthService) VerifyToken(ctx context.Context, req *pb.VerifyTokenReq) (*pb.VerifyTokenResp, error) {
	var claims *utils.Claims
	var err error

	if claims, err = utils.ParseToken(req.Token); err != nil {
		if err.Error() == "invalid token" {
			slog.Info("Invalid", "token:", req.Token)
			return &pb.VerifyTokenResp{Result: false}, nil
		}
		slog.Error("Token parsing error", "error", err)
		return &pb.VerifyTokenResp{Result: false}, nil
	}

	if claims == nil {
		slog.Info("Token claims are nil", "token:", req.Token)
		return &pb.VerifyTokenResp{Result: false}, nil
	}
	if claims.ExpiresAt.Before(time.Now().UTC()) {
		slog.Info("Token expired", "token:", req.Token)
		return &pb.VerifyTokenResp{Result: false}, nil
	}

	val := ""
	refreshTokenKey := "refresh_token:" + req.RefreshToken
	maxRetries := 3
	for i := range maxRetries {
		if val, err = rds.GetCLI().Get(ctx, refreshTokenKey).Result(); err == nil {
			slog.Info("Refresh token found", "token:", req.RefreshToken)
			if val != fmt.Sprint(claims.UserId) {
				slog.Warn("Refresh token user ID mismatch", "token", req.RefreshToken, "userID_in_token", claims.UserId, "userID_in_redis", val)
				return &pb.VerifyTokenResp{Result: false}, nil
			}
			return &pb.VerifyTokenResp{Result: true}, nil
		}
		if err == redis.Nil {
			slog.Warn("Refresh token not found", "token", req.RefreshToken)
			return &pb.VerifyTokenResp{Result: false}, nil
		}
		if i == maxRetries-1 {
			slog.Error("Redis GET error", "error", err)
			return &pb.VerifyTokenResp{Result: false}, err
		}
		time.Sleep(10 << i * time.Millisecond)
	}

	slog.Info("Unknown situation", "token:", req.Token)
	return &pb.VerifyTokenResp{Result: false}, nil
}
