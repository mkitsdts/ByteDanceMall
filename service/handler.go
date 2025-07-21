package service

import (
	pb "bytedancemall/auth/proto"
	"bytedancemall/auth/utils"
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func (s AuthService) DeliverTokenByRPC(ctx context.Context, req *pb.DeliverTokenReq) (*pb.DeliveryTokenResp, error) {
	// 生成token
	token, err := utils.GenerateToken(req.UserId)
	if err != nil {
		return nil, err
	}
	// 将token存入redis
	maxRetries := 3
	for i := range maxRetries {
		if err = s.Redis.Set(ctx, token, req.UserId, 5*24*time.Hour).Err(); err != nil {
			// 操作成功
			continue
		}
		if err = s.Redis.Set(ctx, fmt.Sprint(req.UserId), req.DeviceId, 5*24*time.Hour).Err(); err == nil {
			// 操作成功
			continue
		}
		// 判断是否为临时性错误
		if i < maxRetries-1 {
			// 指数退避重试
			backoffTime := time.Duration(math.Pow(2, float64(i))) * time.Millisecond * 50
			select {
			case <-ctx.Done():
				return &pb.DeliveryTokenResp{Token: "", Result: false}, ctx.Err()
			case <-time.After(backoffTime):
				continue
			}
		}
	}
	return &pb.DeliveryTokenResp{Token: token, Result: true}, nil
}

func (s AuthService) VerifyTokenByRPC(ctx context.Context, req *pb.VerifyTokenReq) (*pb.VerifyTokenResp, error) {
	// 从redis中获取token
	maxRetries := 3
	var userId string
	for i := range maxRetries {
		err := s.Redis.Get(ctx, req.Token).Scan(&userId)
		if err == redis.Nil {
			// token不存在
			return &pb.VerifyTokenResp{Result: false, UserId: 0}, errors.New("token不存在")
		} else if err == nil {
			break
		}
		// 判断是否为临时性错误
		if i < maxRetries-1 {
			// 指数退避重试
			backoffTime := time.Duration(math.Pow(2, float64(i))) * time.Millisecond * 50
			select {
			case <-ctx.Done():
				return &pb.VerifyTokenResp{Result: false, UserId: 0}, ctx.Err()
			case <-time.After(backoffTime):
				continue
			}
		} else {
			// 不可重试错误或重试次数已用完
			return &pb.VerifyTokenResp{Result: false, UserId: 0}, err
		}
	}
	// 字符串转换成claims
	claims, err := utils.ParseToken(req.Token)
	if err != nil {
		return &pb.VerifyTokenResp{Result: false, UserId: 0}, err
	}
	// 字符串转换成uing64
	id, err := strconv.ParseUint(userId, 10, 64)
	if err != nil {
		// 系统错误
		return &pb.VerifyTokenResp{Result: false, UserId: 0}, err
	}
	// 验证token
	if claims.UserId != id {
		return &pb.VerifyTokenResp{Result: false, UserId: 0}, err
	}
	return &pb.VerifyTokenResp{Result: true, UserId: id}, nil
}

func (s AuthService) ProlongTokenByRPC(ctx context.Context, req *pb.ProlongTokenReq) (*pb.ProlongTokenResp, error) {
	token, _ := utils.GenerateToken(req.UserId)
	// 验证用户提供的token是否存在
	var userId string
	maxRetries := 3
	for i := range maxRetries {
		err := s.Redis.Get(ctx, token).Scan(&userId)
		if err == redis.Nil {
			// token不存在
			return &pb.ProlongTokenResp{Result: false}, errors.New("token不存在")
		} else if err == nil {
			// 操作成功
			break
		}
		// 判断是否为临时性错误
		if i < maxRetries-1 {
			// 指数退避重试
			backoffTime := time.Duration(math.Pow(2, float64(i))) * time.Millisecond * 50
			select {
			case <-ctx.Done():
				return &pb.ProlongTokenResp{Result: false}, ctx.Err()
			case <-time.After(backoffTime):
				continue
			}
		} else {
			// 不可重试错误或重试次数已用完
			return &pb.ProlongTokenResp{Result: false}, err
		}
	}
	reqId := strconv.FormatUint(uint64(req.UserId), 10)
	if userId != reqId {
		return &pb.ProlongTokenResp{Result: false}, nil
	}

	// 更新redis中的token
	for i := range maxRetries {
		result, err := s.Redis.Expire(ctx, token, 24*time.Hour).Result()
		if err == nil {
			// 操作成功
			return &pb.ProlongTokenResp{Result: result}, nil
		}
		// 判断是否为临时性错误
		if i < maxRetries-1 {
			// 指数退避重试
			backoffTime := time.Duration(math.Pow(2, float64(i))) * time.Millisecond * 50
			select {
			case <-ctx.Done():
				return &pb.ProlongTokenResp{Result: false}, ctx.Err()
			case <-time.After(backoffTime):
				continue
			}
		} else {
			// 不可重试错误或重试次数已用完
			return &pb.ProlongTokenResp{Result: false}, err
		}
	}
	return &pb.ProlongTokenResp{Result: false}, nil
}
