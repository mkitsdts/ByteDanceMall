package repository

import (
	"context"

	"bytedancemall/gateway/config"
	authpb "bytedancemall/gateway/proto/auth"
)

type AuthRepository struct {
	client *grpcClient
}

func NewAuthRepository(cfg config.AuthService) *AuthRepository {
	return &AuthRepository{client: newGRPCClient(cfg.Host, cfg.Port)}
}

func (r *AuthRepository) VerifyToken(ctx context.Context, token, refreshToken string) (*authpb.VerifyTokenResp, error) {
	conn, err := r.client.Conn()
	if err != nil {
		return nil, err
	}
	return authpb.NewAuthServiceClient(conn).VerifyToken(ctx, &authpb.VerifyTokenReq{
		Token:        token,
		RefreshToken: refreshToken,
	})
}
