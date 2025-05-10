package service

import (
	pb "bytedancemall/user/proto"
	"bytedancemall/user/utils"
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxRedisRetries int = 5

func (s *UserService) Register(ctx context.Context, req *pb.RegisterReq) (*pb.RegisterResp, error) {
	// 参数验证
	if !utils.IsValidEmail(req.Email) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid email")
	}

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 检查Redis缓存
	_, err := s.Redis.Get(ctx, req.Email).Result()
	if err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "email already exists")
	} else if err != redis.Nil {
		// 仅在非"key不存在"错误时重试
		err = utils.RetryRedisOperation(ctx, func() error {
			_, e := s.Redis.Get(ctx, req.Email).Result()
			if e == nil {
				return status.Errorf(codes.AlreadyExists, "email already exists")
			}
			return e
		})

		if err != nil && err != redis.Nil {
			return nil, status.Errorf(codes.Internal, "redis error: %v", err)
		}
	}

	// 开启主库事务
	tx := s.Db.Cluster[s.Db.masterIdx].Begin()
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
	user := User{
		Id:       userId,
		Username: uuid.New().String(),
		Email:    req.Email,
		Password: utils.EncryptPassword(req.Password),
	}

	if result := tx.Create(&user); result.Error != nil {
		return nil, status.Errorf(codes.Internal, "database error: %v", result.Error)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}
	committed = true

	// 更新Redis缓存并确保写入成功
	if err := s.ensureRedisUpdate(user.Email, user.Id); err != nil {
		// Redis操作失败是严重问题，记录日志，但仍向客户端返回结果
		// 也可以选择启动一个异步任务继续尝试写入Redis
		log.Printf("Critical: Failed to update Redis after %d retries: %v", maxRedisRetries, err)
		go s.asyncEnsureRedisUpdate(user.Email, user.Id)
	}

	return &pb.RegisterResp{UserId: user.Id}, nil
}

// 确保数据写入Redis成功
func (s *UserService) ensureRedisUpdate(email string, userId int64) error {
	var lastErr error

	// 创建一个更长的超时上下文，确保有足够时间重试
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for i := range maxRedisRetries {
		// 使用管道执行多个操作
		pipe := s.Redis.Pipeline()
		pipe.Set(ctxTimeout, email, userId, 0) // 永不过期，或根据业务需求设置TTL

		// 可以添加其他必要的Redis操作
		// pipe.HSet(ctxTimeout, "users", userId, email)

		_, err := pipe.Exec(ctxTimeout)
		if err == nil {
			return nil // 成功写入
		}

		lastErr = err
		backoffTime := time.Duration(math.Pow(2, float64(i))) * 100 * time.Millisecond

		select {
		case <-ctxTimeout.Done():
			return fmt.Errorf("context timeout during Redis update: %w", ctxTimeout.Err())
		case <-time.After(backoffTime):
			// 等待后继续重试
			log.Printf("Redis update retry %d after %v: %v", i+1, backoffTime, err)
		}
	}

	return fmt.Errorf("failed to update Redis after %d retries: %w", maxRedisRetries, lastErr)
}

// 异步确保数据写入Redis
func (s *UserService) asyncEnsureRedisUpdate(email string, userId int64) {

	const maxAttempts int = 100 // 设置一个上限，避免无限重试

	for attempt := range maxAttempts {
		// 指数退避，但有上限
		backoffSeconds := math.Min(math.Pow(2, float64(attempt)), 300) // 最多等待5分钟
		time.Sleep(time.Duration(backoffSeconds) * time.Second)

		err := s.ensureRedisUpdate(email, userId)
		if err == nil {
			log.Printf("Async Redis update succeeded after %d attempts", attempt+1)
			return
		}

		log.Printf("Async Redis update attempt %d failed: %v", attempt+1, err)
	}

	log.Printf("CRITICAL: Failed to update Redis after %d async attempts", maxAttempts)
}
