package service

import (
	pb "bytedancemall/user/proto"
	"bytedancemall/user/usecase"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type UserService struct {
	usecase *usecase.UserUsecase
	pb.UnimplementedUserServiceServer
}

func New(uc *usecase.UserUsecase) *UserService {
	return &UserService{usecase: uc}
}

func (s *UserService) Register(ctx context.Context, req *pb.RegisterReq) (*pb.RegisterResp, error) {
	resp, err := s.usecase.Register(ctx, req)
	if err != nil {
		if err == usecase.ErrInvalidRequest {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (s *UserService) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {
	resp, err := s.usecase.Login(ctx, req)
	if err != nil {
		switch {
		case err == usecase.ErrInvalidRequest:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case err == gorm.ErrInvalidData:
			return &pb.LoginResp{Result: false}, status.Error(codes.InvalidArgument, "incorrect password")
		case err == gorm.ErrRecordNotFound:
			return &pb.LoginResp{Result: false}, status.Error(codes.NotFound, "user not found")
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return resp, nil
}

func (s *UserService) GetUserInfo(ctx context.Context, req *pb.GetUserInfoReq) (*pb.GetUserInfoResp, error) {
	resp, err := s.usecase.GetUserInfo(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}
