package usecase

import (
	"bytedancemall/user/model"
	pb "bytedancemall/user/proto"
	"bytedancemall/user/repository"
	"bytedancemall/user/utils"
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrInvalidRequest = errors.New("invalid request")

type UserUsecase struct {
	repos *repository.Repositories
}

func New(repos *repository.Repositories) *UserUsecase {
	return &UserUsecase{repos: repos}
}

func (u *UserUsecase) Register(ctx context.Context, req *pb.RegisterReq) (*pb.RegisterResp, error) {
	if !utils.IsValidEmail(req.GetEmail()) || req.GetPassword() == "" {
		return nil, ErrInvalidRequest
	}

	userID := utils.GenerateId()
	salt, err := utils.GenerateSalt()
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Id:           userID,
		Username:     uuid.New().String(),
		Email:        req.GetEmail(),
		Password:     utils.EncryptPassword(req.GetPassword(), salt),
		PasswordSalt: salt,
	}

	if err := u.repos.User.Create(ctx, user); err != nil {
		return nil, err
	}

	return &pb.RegisterResp{Result: true, UserId: user.Id}, nil
}

func (u *UserUsecase) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {
	if req.GetPassword() == "" || (req.GetEmail() == "" && req.GetUserId() == 0) {
		return nil, ErrInvalidRequest
	}

	var (
		user *model.User
		err  error
	)
	if req.GetEmail() != "" {
		if !utils.IsValidEmail(req.GetEmail()) {
			return nil, ErrInvalidRequest
		}
		user, err = u.repos.User.FindByEmail(ctx, req.GetEmail())
	} else {
		user, err = u.repos.User.FindByID(ctx, req.GetUserId())
	}
	if err != nil {
		return nil, err
	}

	if !utils.VerifyPassword(req.GetPassword(), user.PasswordSalt, user.Password) {
		return &pb.LoginResp{Result: false}, gorm.ErrInvalidData
	}

	if err := u.repos.Login.CreateRecord(ctx, &model.LoginRecord{UserId: user.Id}); err != nil {
		return nil, err
	}

	loginCount, err := u.repos.UserCache.IncrementLoginCount(ctx, user.Id)
	if err != nil {
		slog.Warn("failed to increment login count", "user_id", user.Id, "error", err)
	} else {
		slog.Info("user login succeeded", "user_id", user.Id, "login_count", loginCount)
	}

	return &pb.LoginResp{Result: true}, nil
}

func (u *UserUsecase) GetUserInfo(ctx context.Context, req *pb.GetUserInfoReq) (*pb.GetUserInfoResp, error) {
	if req.GetUserId() == 0 {
		return &pb.GetUserInfoResp{Exists: false}, nil
	}

	cached, hit, err := u.repos.UserCache.GetUserInfo(ctx, req.GetUserId())
	if err == nil && hit {
		return &pb.GetUserInfoResp{
			Exists: true,
			Email:  cached.Email,
			Name:   cached.Username,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	user, err := u.repos.User.FindByID(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &pb.GetUserInfoResp{Exists: false}, nil
		}
		return nil, err
	}

	_ = u.repos.UserCache.SetUserInfo(ctx, user.Id, &repository.UserInfoCache{
		Email:    user.Email,
		Username: user.Username,
	}, time.Hour)

	return &pb.GetUserInfoResp{
		Exists: true,
		Email:  user.Email,
		Name:   user.Username,
	}, nil
}
