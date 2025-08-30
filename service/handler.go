package service

import (
	pb "bytedancemall/auth/proto"
	"bytedancemall/auth/utils"
	"context"
	"fmt"
	"log/slog"
	"time"
)

func (s *AuthService) DeliverTokenByRPC(ctx context.Context, req *pb.DeliverTokenReq) (*pb.DeliveryTokenResp, error) {
	token, err := utils.GenerateToken(req.UserId)
	if err != nil {
		return nil, err
	}
	slog.Info(fmt.Sprintf("Generated token for user %d: %s", req.UserId, token))
	return &pb.DeliveryTokenResp{Token: token, Result: true}, nil
}

func (s *AuthService) VerifyTokenByRPC(ctx context.Context, req *pb.VerifyTokenReq) (*pb.VerifyTokenResp, error) {
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
