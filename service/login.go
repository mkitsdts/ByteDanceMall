package service

import (
	pb "bytedancemall/user/proto"
	p "bytedancemall/user/proto/auth"
	"bytedancemall/user/utils"
	"context"
	"log"
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

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 先尝试从Redis缓存获取用户ID
	var userId int64
	var userIdStr string
	var err error

	// 带退避策略的Redis重试
	err = utils.RetryRedisOperation(ctx, func() error {
		userIdStr, err = s.Redis.Get(ctx, req.Email).Result()
		if err == nil {
			// 转换用户ID
			userId, err = strconv.ParseInt(userIdStr, 10, 64)
			return err
		}
		return err // 返回Redis错误以便外部处理
	})

	var user User
	var passwordFromStore string

	if err == nil {
		// Redis中找到了用户ID，尝试获取密码哈希
		err = utils.RetryRedisOperation(ctx, func() error {
			passwordFromStore, err = s.Redis.Get(ctx, userIdStr).Result()
			return err
		})

		if err != nil && err != redis.Nil {
			// Redis错误但非"不存在"错误，记录日志
			log.Printf("Redis error when getting password: %v", err)
			// 继续从数据库查询，不要直接返回错误
		}
	}

	// 如果Redis查询失败或数据不完整，尝试从数据库获取
	if err == redis.Nil || passwordFromStore == "" {
		// 从数据库查询用户
		for i := 0; i < 3; i++ { // 最多重试3次
			result := s.Db.GetReader().Where("email = ?", req.Email).First(&user)
			if result.Error == nil {
				// 查询成功
				userId = user.Id
				passwordFromStore = user.Password

				// 异步更新Redis缓存
				go func() {
					pipe := s.Redis.Pipeline()
					pipe.Set(context.Background(), req.Email, strconv.FormatInt(user.Id, 10), 24*time.Hour)
					pipe.Set(context.Background(), strconv.FormatInt(user.Id, 10), user.Password, 24*time.Hour)
					_, err := pipe.Exec(context.Background())
					if err != nil {
						log.Printf("Failed to update Redis after DB query: %v", err)
					}
				}()

				break
			} else if result.Error == gorm.ErrRecordNotFound {
				// 用户确实不存在
				return nil, status.Errorf(codes.NotFound, "user not found")
			} else {
				// 其他数据库错误，带退避重试
				backoff := time.Duration(math.Pow(2, float64(i))) * 100 * time.Millisecond
				time.Sleep(backoff)
			}
		}
	}

	// 此时应该已从Redis或数据库获取到用户信息
	if passwordFromStore == "" {
		return nil, status.Errorf(codes.Internal, "failed to retrieve user data")
	}

	// 验证密码 (应该使用安全的密码验证)
	if utils.EncryptPassword(req.Password) != passwordFromStore {
		return nil, status.Errorf(codes.InvalidArgument, "incorrect password")
	}

	// 生成认证令牌
	tokenResp, err := s.AuthServer.DeliverTokenByRPC(ctx, &p.DeliverTokenReq{UserId: userId})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "authentication service error: %v", err)
	}

	// 返回登录成功响应，包含用户ID和token
	return &pb.LoginResp{
		Result: true,
		Token:  tokenResp.Token,
	}, nil
}
