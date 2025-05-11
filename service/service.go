package service

import (
	pb "bytedancemall/user/proto"
	p "bytedancemall/user/proto/auth"
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

type UserService struct {
	Db         Database
	Redis      *redis.ClusterClient
	AuthServer p.AuthServiceClient
	pb.UnimplementedUserServiceServer
}

func (s *UserService) InitUserService() {
	// 从configs.json中读取数据库和Redis配置
	// 配置mysql和redis集群
	file, err := os.Open("configs.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	configs := Configs{}
	err = decoder.Decode(&configs)
	if err != nil {
		panic(err)
	}
	s.Db.ch = make(chan bool)
	s.Db.InitDatabase(configs.MysqlConfig.Configs)

	// 连接redis
	{
		// 生成redis集群的地址
		var redisAddrs []string
		for _, v := range configs.RedisConfig.Configs {
			redisAddrs = append(redisAddrs, v.Host+":"+v.Port)
		}
		// 初始化redis
		s.Redis = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        redisAddrs,
			Password:     configs.RedisConfig.Password,
			PoolSize:     50,
			MinIdleConns: 10,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		})
		// 测试redis连接
		_, err = s.Redis.Ping(context.Background()).Result()
		if err != nil {
			panic(err)
		}
	}

	// 连接认证中心
	{
		conn, err := grpc.NewClient(configs.AuthServerConfig.Address)
		if err != nil {
			panic(err)
		}
		s.AuthServer = p.NewAuthServiceClient(conn)
	}
}
