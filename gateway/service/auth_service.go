package service

import (
	"context"
	"errors"

	"bytedancemall/gateway/repository"
)

var ErrUnauthorized = errors.New("unauthorized")

type AuthService struct {
	repo *repository.AuthRepository
}

func NewAuthService(repo *repository.AuthRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) Verify(ctx context.Context, token, refreshToken string) error {
	if token == "" {
		return ErrUnauthorized
	}

	resp, err := s.repo.VerifyToken(ctx, token, refreshToken)
	if err != nil {
		return err
	}
	if !resp.GetResult() {
		return ErrUnauthorized
	}
	return nil
}
