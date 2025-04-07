package service

import (
	"bytedancemall/auth/model"
	pb "bytedancemall/auth/proto"
	"bytedancemall/auth/utils"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type AuthService struct {
	Redis *redis.ClusterClient
	pb.UnimplementedAuthServiceServer
}

func initUserService() (*redis.ClusterClient) {
	// 从configs.json中读取数据库和Redis配置
	// 配置redis集群
	file, err := os.Open("configs.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	configs := model.Configs{}
	err = decoder.Decode(&configs)
	if err != nil {
		panic(err)
	}

	// 生成redis集群的地址
	var redisAddrs []string
	for _, v := range configs.RedisConfig.Configs {
		redisAddrs = append(redisAddrs, v.Host + ":" + v.Port)
	}
	// 初始化redis
	redis := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: redisAddrs,
		Password: configs.RedisConfig.Password,
	})
	return redis
}

func NewAuthService() *AuthService {
	var s AuthService
	s.Redis = initUserService()
	return &s;
}

func (s *AuthService) DeliverTokenByRPC(ctx context.Context, req *pb.DeliverTokenReq) (*pb.DeliveryResp, error) {
	// 生成token
	token, err := utils.GenerateToken(req.UserId)
	if err != nil {
		return nil, err
	}
	// 将token存入redis
	err = s.Redis.Set(ctx, token, req.UserId, 24 * time.Hour).Err()
	if err != nil {
		// 错误处理
		return &pb.DeliveryResp{Token: ""}, err
	}
	return &pb.DeliveryResp{Token: token}, nil
}

func (s *AuthService) VerifyTokenByRPC(ctx context.Context, req *pb.VerifyTokenReq) (*pb.VerifyResp, error) {
	// 从redis中获取token
	userId, err := s.Redis.Get(ctx, req.Token).Result()
	if err != nil {
		return &pb.VerifyResp{Result: false, UserId: 0}, err
	}
	// 字符串转换成claims
	claims, err := utils.ParseToken(req.Token)
	if err != nil {
		return &pb.VerifyResp{Result: false, UserId: 0}, err
	}
	// 字符串转换成uing32
	id, err := strconv.ParseUint(userId, 10, 32)
	if err != nil {
		// 系统错误
		return &pb.VerifyResp{Result: false, UserId: 0}, err
	}
	// 验证token
	if claims.UserId != uint32(id) {
		return &pb.VerifyResp{Result: false, UserId: 0}, err
	}
	return &pb.VerifyResp{Result: true,UserId: uint32(id)}, nil
}

func (s *AuthService) ProlongTokenByRPC(ctx context.Context, req *pb.ProlongTokenReq) (*pb.ProlongTokenResp, error) {
	token, _ := utils.GenerateToken(req.UserId)
	// 验证用户提供的token是否存在
    userId, err := s.Redis.Get(ctx, token).Result()
    if err != nil {
        if err == redis.Nil {
			return &pb.ProlongTokenResp{Result: false} , errors.New("token不存在")
		}
        return &pb.ProlongTokenResp{Result: false}, nil
    }
	reqId := strconv.FormatUint(uint64(req.UserId), 10)
	if userId != reqId {
		return &pb.ProlongTokenResp{Result: false}, nil
	}

	// 更新redis中的token
	result, err:= s.Redis.Expire(ctx, token, 24 * time.Hour).Result()
	if err != nil {
		return nil, err
	}
	if !result {
		return &pb.ProlongTokenResp{Result: false}, err
	}
	return &pb.ProlongTokenResp{Result: true}, nil
}