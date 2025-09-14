package service

import (
	rds "bytedancemall/auth/pkg/redis"
	pb "bytedancemall/auth/proto"
	"bytedancemall/auth/utils"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func (s *AuthService) DeliverToken(ctx context.Context, req *pb.DeliverTokenReq) (*pb.DeliveryTokenResp, error) {
	token, err := utils.GenerateToken(req.UserId, 5)
	if err != nil {
		return &pb.DeliveryTokenResp{Result: false}, err
	}
	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return &pb.DeliveryTokenResp{Result: false}, err
	}
	key := "refresh_token:" + refreshToken
	maxRetries := 10
	for i := range maxRetries {
		_, err := rds.GetCLI().Set(context.Background(), key, fmt.Sprint(req.UserId), time.Hour*24*30).Result()
		if err == nil {
			break
		}
		if i == maxRetries-1 {
			slog.Error("Redis SET error", "error", err)
			return &pb.DeliveryTokenResp{Result: false}, err
		}
		time.Sleep(10 << i * time.Millisecond)
	}

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
	userID, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		slog.Error("Token generation error", "error", err)
		return &pb.RefreshTokenResp{Result: false}, err
	}
	token, err := utils.GenerateToken(userID, 5)
	if err != nil {
		slog.Error("Token generation error", "error", err)
		return &pb.RefreshTokenResp{Result: false}, err
	}
	return &pb.RefreshTokenResp{Result: true, Token: token}, nil
}
