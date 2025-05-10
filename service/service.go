package service

import (
	pb "bytedancemall/user/proto"
	p "bytedancemall/user/proto/auth"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
	s.Db.Configs = configs.MysqlConfig.Configs
	s.Db.healthy = make([]bool, len(configs.MysqlConfig.Configs))
	s.Db.ch = make(chan bool)
	// 连接 mysql
	{
		for i := range len(configs.MysqlConfig.Configs) {
			dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
				configs.MysqlConfig.Configs[i].User,
				configs.MysqlConfig.Configs[i].Password,
				configs.MysqlConfig.Configs[i].Host,
				configs.MysqlConfig.Configs[i].Port,
				configs.MysqlConfig.Configs[i].Database,
			)
			slave, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
			if err != nil {
				panic(err)
			}
			s.Db.healthy[i] = true
			s.Db.Cluster = append(s.Db.Cluster, slave)
		}
		s.Db.masterIdx = 0
		// 创建表
		if err := s.Db.Cluster[s.Db.masterIdx].AutoMigrate(&User{}); err != nil {
			panic(err)
		}
		// 启动健康检查
		s.startHealthCheck()
	}

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

func (s *UserService) startHealthCheck() {
	go func() {
		for {
			select {
			case <-s.Db.ch:
				return
			default:
				s.checkMysqlHealth()
				time.Sleep(5 * time.Second)
			}
		}
	}()
}

func (s *UserService) checkMysqlHealth() {
	s.Db.mu.Lock()
	defer s.Db.mu.Unlock()

	for i, db := range s.Db.Cluster {
		if err := db.Exec("SELECT 1").Error; err != nil {
			s.Db.healthy[i] = false
		} else {
			s.Db.healthy[i] = true
		}
	}

	// 检查主库是否健康
	index := s.Db.masterIdx
	if !s.Db.healthy[index] {
		// 如果主库不健康，选择一个健康的从库
		for i := range len(s.Db.healthy) {
			if s.Db.healthy[i] {
				s.Db.masterIdx = i
				break
			}
		}
	}
	// 如果没有健康的从库，返回错误
	if s.Db.masterIdx == index {
		fmt.Println("主库不可用，所有数据库都不可用")
		return
	}
}
