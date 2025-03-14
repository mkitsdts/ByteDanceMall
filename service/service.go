package service

import (
	pb "bytedancemall/auth/proto"
	"bytedancemall/auth/model"
	"bytedancemall/auth/utils"
	"context"
	"github.com/redis/go-redis/v9"
	"encoding/json"
	"os"
	"strconv"
)

type AuthService struct {
	Redis *redis.ClusterClient
	pb.UnimplementedAuthServiceServer
}

func initUserService() (*redis.ClusterClient) {
	// 从configs.json中读取数据库和Redis配置
	// 配置mysql和redis集群
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
	err = s.Redis.Set(ctx, token, req.UserId, 0).Err()
	if err != nil {
		return nil, err
	}
	return &pb.DeliveryResp{}, nil
}

func (s *AuthService) VerifyTokenByRPC(ctx context.Context, req *pb.VerifyTokenReq) (*pb.VerifyResp, error) {
	// 从redis中获取token
	userId, err := s.Redis.Get(ctx, req.Token).Result()
	if err != nil {
		return &pb.VerifyResp{Res: false}, err
	}
	// 字符串转换成claims
	claims, err := utils.ParseToken(req.Token)
	if err != nil {
		return &pb.VerifyResp{Res: false}, err
	}
	// 字符串转换成int64
	userIdInt, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		return &pb.VerifyResp{Res: false}, err
	}
	// 验证token
	if claims.UserId != userIdInt {
		return &pb.VerifyResp{Res: false}, err
	}
	return &pb.VerifyResp{Res: true}, nil
}