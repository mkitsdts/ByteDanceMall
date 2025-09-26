package service

import (
	"bytedancemall/user/model"
	"bytedancemall/user/pkg"
	pb "bytedancemall/user/proto"
	"bytedancemall/user/utils"
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *UserService) Register(ctx context.Context, req *pb.RegisterReq) (*pb.RegisterResp, error) {
	// 参数验证
	if !utils.IsValidEmail(req.Email) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid email")
	}

	// 开启主库事务
	tx := pkg.DB().Begin()
	if tx.Error != nil {
		return nil, status.Errorf(codes.Internal, "database error: %v", tx.Error)
	}

	// 确保事务最终会被回滚或提交
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	// 创建新用户
	userId := utils.GenerateId()
	register := model.Register{
		Id:       userId,
		Username: uuid.New().String(),
		Email:    req.Email,
		Password: utils.EncryptPassword(req.Password),
	}

	if result := tx.Create(&register); result.Error != nil {
		return nil, status.Errorf(codes.Internal, "database error: %v", result.Error)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}
	committed = true

	return &pb.RegisterResp{UserId: register.Id}, nil
}
