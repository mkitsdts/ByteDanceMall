package service

import (
	"bytedancemall/auth/pkg/redis"
	pb "bytedancemall/auth/proto"
	"bytedancemall/auth/utils"
	"context"
	"fmt"
	"log/slog"
	"time"
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
	slog.Info(fmt.Sprintf("Generated token for user %d: %s", req.UserId, token))
	return &pb.DeliveryTokenResp{Token: token, RefreshToken: refreshToken, Result: true}, nil
}

func (s *AuthService) VerifyToken(ctx context.Context, req *pb.VerifyTokenReq) (*pb.VerifyTokenResp, error) {
	var claims *utils.Claims
	var err error
	for range 3 {
		claims, err = utils.ParseToken(req.Token)
		if err == nil {
			break
		}
		if err.Error() == "invalid token" {
			slog.Info("Invalid", "token:", req.Token)
			return &pb.VerifyTokenResp{Result: false}, nil
		}
	}
	if claims == nil {
		slog.Info("Token claims are nil", "token:", req.Token)
		return &pb.VerifyTokenResp{Result: false}, nil
	}
	if claims.ExpiresAt.Before(time.Now().UTC()) {
		slog.Info("Token expired", "token:", req.Token)
		return &pb.VerifyTokenResp{Result: false}, nil
	}
	slog.Info("Token verified", "token:", req.Token)
	return &pb.VerifyTokenResp{Result: true, UserId: claims.UserId}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *pb.RefreshTokenReq) (*pb.RefreshTokenResp, error) {
	cli := redis.GetCLI()
	refreshToken := "refresh_token:" + req.RefreshToken
	val, err := cli.Get(ctx, refreshToken).Result()
	if err != nil && err.Error() != "redis: nil" {
		slog.Error("Redis GET error", "error", err)
		return &pb.RefreshTokenResp{Result: false}, err
	}
	if val != "" {
		slog.Info("Using cached refresh token", "token", val)
		return &pb.RefreshTokenResp{Token: val, Result: true}, nil
	}
	if claim, err := utils.ParseToken(req.RefreshToken); err == nil {
		newToken, err := utils.GenerateToken(claim.UserId, 5)
		if err != nil {
			slog.Error("Token generation error", "error", err)
			return &pb.RefreshTokenResp{Result: false}, err
		}
		if val != fmt.Sprint(claim.UserId) {
			slog.Warn("Refresh token user ID mismatch", "expected", val, "got", claim.UserId)
			return &pb.RefreshTokenResp{Result: false}, nil
		}
		return &pb.RefreshTokenResp{Token: newToken, Result: true}, nil
	} else {
		slog.Error("Refresh token parsing error", "error", err)
		return &pb.RefreshTokenResp{Result: false}, nil
	}
}

func (s *AuthService) RemoveRefreshToken(ctx context.Context, req *pb.RemoveRefreshTokenReq) (*pb.RemoveRefreshTokenResp, error) {
	cli := redis.GetCLI()
	refreshToken := "refresh_token:" + req.RefreshToken
	_, err := cli.Del(ctx, refreshToken).Result()
	if err != nil {
		slog.Error("Redis DEL error", "error", err)
		return &pb.RemoveRefreshTokenResp{Result: false}, err
	}
	slog.Info("Refresh token removed", "token", req.RefreshToken)
	return &pb.RemoveRefreshTokenResp{Result: true}, nil
}
