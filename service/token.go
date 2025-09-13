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

func (s *AuthService) DeliverToken(ctx context.Context, req *pb.DeliverTokenReq) (*pb.DeliveryTokenResp, error) {
	token, err := utils.GenerateToken(req.UserId, 5)
	if err != nil {
		return &pb.DeliveryTokenResp{Result: false}, err
	}
	refreshToken, err := utils.GenerateToken(req.UserId, 60*24*30)
	if err != nil {
		return &pb.DeliveryTokenResp{Result: false}, err
	}
	key := "refresh_token:" + refreshToken
	go func() {
		for i := range 5 {
			_, err := rds.GetCLI().Set(context.Background(), key, fmt.Sprint(req.UserId), time.Hour*24*30).Result()
			if err == nil {
				break
			}
			if i == 4 {
				slog.Error("Redis SET error", "error", err)
			}
			time.Sleep(10 << i * time.Millisecond)
		}
	}()
	slog.Info(fmt.Sprintf("Generated token for user %d: %s", req.UserId, token))
	return &pb.DeliveryTokenResp{Token: token, RefreshToken: refreshToken, Result: true}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *pb.RefreshTokenReq) (*pb.RefreshTokenResp, error) {
	cli := rds.GetCLI()
	refreshToken := "refresh_token:" + req.RefreshToken
	maxRetries := 3
	var err error
	var val string
	for i := range maxRetries {
		if val, err = cli.Get(ctx, refreshToken).Result(); err == nil {
			break
		}
		if err == redis.Nil {
			slog.Warn("Refresh token not found", "token", req.RefreshToken)
			return &pb.RefreshTokenResp{Result: false}, nil
		}
		if val == "" {
			slog.Warn("Refresh token not found", "token", req.RefreshToken)
			return &pb.RefreshTokenResp{Result: false}, nil
		}
		if i == maxRetries-1 {
			// Redis 发生错误，直接返回
			slog.Error("Redis GET error", "error", err)
			return &pb.RefreshTokenResp{Result: false}, err
		}
		time.Sleep(10 << i * time.Millisecond)
	}
	// 解析 refresh token
	newToken := ""
	if claim, err := utils.ParseToken(req.RefreshToken); err == nil {
		// refresh token 未过期，生成新的 token
		if !claim.ExpiresAt.Before(time.Now().UTC()) {
			newToken, err = utils.GenerateToken(claim.UserId, 5)
			if err != nil {
				slog.Error("Refresh token generation error", "error", err)
				return &pb.RefreshTokenResp{Result: false}, err
			}
			for i := range maxRetries {
				if _, err = cli.Set(ctx, refreshToken, fmt.Sprint(claim.UserId), time.Hour*24*30).Result(); err == nil {
					slog.Info("Refresh token renewed", "token", req.RefreshToken)
					return &pb.RefreshTokenResp{Token: newToken, Result: true}, nil
				}
				if i == maxRetries-1 {
					slog.Error("Redis SET error", "error", err)
					return &pb.RefreshTokenResp{Result: false}, err
				}
				time.Sleep(10 << i * time.Millisecond)
			}
		}
		return &pb.RefreshTokenResp{Result: false}, fmt.Errorf("refresh token not expired yet")
	} else {
		slog.Error("Refresh token parsing error", "error", err)
		return &pb.RefreshTokenResp{Result: false}, nil
	}
}
