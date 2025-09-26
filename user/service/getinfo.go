package service

import (
	"bytedancemall/user/model"
	"bytedancemall/user/pkg"
	pb "bytedancemall/user/proto"
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func (s *UserService) GetUserInfo(ctx context.Context, req *pb.GetUserInfoReq) (*pb.GetUserInfoResp, error) {
	type UserInfo struct {
		Email string
		Name  string
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	maxRetriesTime := 10
	for i := 0; i < maxRetriesTime; i++ {
		val, err := pkg.GetCLI().Get(ctx, fmt.Sprintf("user:%d", req.UserId)).Result()
		if err == nil {
			var userInfo model.User
			if err := json.Unmarshal([]byte(val), &userInfo); err != nil {
				return nil, err
			}
			return &pb.GetUserInfoResp{
				Exists: true,
				Email:  userInfo.Email,
			}, nil
		} else if err == redis.Nil {
			tx := pkg.DB().Begin()
			// Fetch user info from the database
			var email, name string
			err := pkg.DB().Model(&model.User{}).Select("email", "name").Where("id = ?", req.UserId).Row().Scan(&email, &name)
			if err == gorm.ErrRecordNotFound {
				return &pb.GetUserInfoResp{
					Exists: false,
				}, nil
			} else if err != nil {
				tx.Rollback()
				return &pb.GetUserInfoResp{
					Exists: false,
				}, nil
			}

			userInfo := UserInfo{
				Email: email,
				Name:  name,
			}
			body, _ := json.Marshal(userInfo)
			if err := pkg.GetCLI().Set(ctx, fmt.Sprintf("user:%d", req.UserId), string(body), 0).Err(); err != nil {
				tx.Rollback()
				return nil, err
			}
			tx.Commit()
		} else {
			return &pb.GetUserInfoResp{
				Exists: false,
			}, nil
		}
	}
	return &pb.GetUserInfoResp{
		Exists: false,
	}, nil
}
