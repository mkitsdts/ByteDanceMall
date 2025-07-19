package service

import (
	"bytedancemall/user/model"
	pb "bytedancemall/user/proto"
	"bytedancemall/user/utils"
	"context"
	"log"
	"log/slog"
	"math"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {
	// 参数校验
	if !utils.IsValidEmail(req.Email) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid email")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 先尝试从Redis缓存获取用户ID
	var userId uint64
	var password string

	userIdStr, err := s.Redis.Get(ctx, req.Email).Result()
	if err != nil && err != redis.Nil {
		slog.Error("Redis error when getting user ID: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get user ID from Redis")
	} else if err == redis.Nil {
		slog.Info("User email %s not found in Redis, querying database", req.Email)
	} else {
		// 成功从Redis获取到用户ID
		userId, err = strconv.ParseUint(userIdStr, 10, 64)
		if err != nil {
			slog.Error("Failed to parse user ID from Redis: %v", err)
			return nil, status.Errorf(codes.Internal, "failed to parse user ID from Redis")
		} else {
			slog.Info("Found user ID %d in Redis for email %s", userId, req.Email)
			if password, err = s.Redis.Get(ctx, userIdStr).Result(); err != nil {
				slog.Error("Failed to get password from Redis for user ID %d: %v", userId, err)
			}
			if password != "" {
				if utils.EncryptPassword(req.Password) == password {
					slog.Info("Login successful for user ID %d", userId)
					return &pb.LoginResp{
						Result: true,
					}, nil
				} else {
					slog.Warn("Incorrect password for user ID %d", userId)
					return nil, status.Errorf(codes.InvalidArgument, "incorrect password")
				}
			}
		}
	}

	var user model.User

	// 如果Redis查询失败或数据不完整，尝试从数据库获取
	for i := range 3 { // 最多重试3次
		result := s.Db.GetReader().Where("email = ?", req.Email).First(&user)
		if result.Error == nil {
			// 查询成功
			if userId == user.Id && password == user.Password {
				// 异步更新Redis缓存
				go func() {
					pipe := s.Redis.Pipeline()
					pipe.Set(context.Background(), req.Email, strconv.FormatUint(user.Id, 10), 24*time.Hour)
					pipe.Set(context.Background(), strconv.FormatUint(user.Id, 10), user.Password, 24*time.Hour)
					_, err := pipe.Exec(context.Background())
					if err != nil {
						log.Printf("Failed to update Redis after DB query: %v", err)
					}
				}()
				return &pb.LoginResp{
					Result: true,
				}, nil
			}
		} else if result.Error == gorm.ErrRecordNotFound {
			// 用户确实不存在
			return &pb.LoginResp{
				Result: false,
			}, status.Errorf(codes.NotFound, "user not found")
		} else {
			// 其他数据库错误，带退避重试
			backoff := time.Duration(math.Pow(2, float64(i))) * 100 * time.Millisecond
			time.Sleep(backoff)
		}
	}

	// 返回登录成功响应，包含用户ID和token
	return &pb.LoginResp{
		Result: false,
	}, status.Errorf(codes.Internal, "failed to login after multiple attempts")
}
