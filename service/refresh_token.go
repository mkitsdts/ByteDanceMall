package service

import (
	"context"
	"log/slog"
	"time"

	rds "bytedancemall/auth/pkg/redis"
	pb "bytedancemall/auth/proto"
	"bytedancemall/auth/utils"

	"github.com/redis/go-redis/v9"
)

func (s *AuthService) ProlongRefreshToken(ctx context.Context, req *pb.ProlongRefreshTokenReq) (*pb.ProlongRefreshTokenResp, error) {
	// 检查 refresh token 合法性
	// 生成新的 refresh token 并返回
	// 异步删除旧 refresh token 并写入新的 refresh token

	refreshTokenKey := "refresh_token:" + req.RefreshToken
	cli := rds.GetCLI()
	maxRetries := 5
	var err error
	var val string
	for i := range maxRetries {
		if val, err = cli.Get(ctx, refreshTokenKey).Result(); err == nil {
			break
		}
		if err == redis.Nil {
			slog.Warn("Refresh token not found", "token", req.RefreshToken)
			return &pb.ProlongRefreshTokenResp{Result: false}, nil
		}
		if i == maxRetries-1 {
			slog.Error("Redis GET error", "error", err)
			return &pb.ProlongRefreshTokenResp{Result: false}, err
		}
		time.Sleep(10 << i * time.Millisecond)
	}

	newRefreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		slog.Error("Refresh token generation error", "error", err)
		return &pb.ProlongRefreshTokenResp{Result: false}, err
	}

	// 这里后续需要加入异常处理机制
	for i := range maxRetries {
		if _, err = cli.Del(ctx, refreshTokenKey).Result(); err == nil {
			break
		}
		if i == maxRetries-1 {
			slog.Error("Redis DEL error", "error", err)
		}
		time.Sleep(10 << i * time.Millisecond)
	}
	for i := range maxRetries {
		if _, err = cli.Set(ctx, "refresh_token:"+newRefreshToken, val, time.Hour*24*30).Result(); err == nil {
			break
		}
		if i == maxRetries-1 {
			slog.Error("Redis SET error", "error", err)
		}
		time.Sleep(10 << i * time.Millisecond)
	}

	return &pb.ProlongRefreshTokenResp{Result: true}, nil
}

func (s *AuthService) RemoveRefreshToken(ctx context.Context, req *pb.RemoveRefreshTokenReq) (*pb.RemoveRefreshTokenResp, error) {
	cli := rds.GetCLI()
	refreshToken := "refresh_token:" + req.RefreshToken
	maxRetries := 5
	var err error
	for i := range maxRetries {
		if _, err = cli.Del(ctx, refreshToken).Result(); err == nil {
			slog.Info("Refresh token removed", "token", req.RefreshToken)
			return &pb.RemoveRefreshTokenResp{Result: true}, nil
		}
		if i == maxRetries-1 {
			slog.Error("Redis DEL error", "error", err)
		}
		time.Sleep(10 << i * time.Millisecond)
	}
	return &pb.RemoveRefreshTokenResp{Result: false}, err
}
