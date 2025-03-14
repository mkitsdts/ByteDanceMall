package service

import (
	"bytedancemall/user/model"
	pb "bytedancemall/user/proto"
	"bytedancemall/user/utils"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Database struct {
	Master *gorm.DB
	Slaves  []*gorm.DB
}

type UserService struct {
    Db Database
	Redis *redis.ClusterClient
    pb.UnimplementedUserServiceServer
}

func initUserService() (Database, *redis.ClusterClient) {
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
	
	// 生成mysql的dsn
	masterDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		configs.MysqlConfig.Configs[0].User,
		configs.MysqlConfig.Configs[0].Password,
		configs.MysqlConfig.Configs[0].Host,
		configs.MysqlConfig.Configs[0].Port,
		configs.MysqlConfig.Configs[0].Database,
	)

	// 连接mysql
	var db Database
	db.Master, err = gorm.Open(mysql.Open(masterDsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	for i := 1; i < len(configs.MysqlConfig.Configs); i++ {
		slaveDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			configs.MysqlConfig.Configs[i].User,
			configs.MysqlConfig.Configs[i].Password,
			configs.MysqlConfig.Configs[i].Host,
			configs.MysqlConfig.Configs[i].Port,
			configs.MysqlConfig.Configs[i].Database,
		)
		slave, err := gorm.Open(mysql.Open(slaveDsn), &gorm.Config{})
		if err != nil {
			panic(err)
		}
		db.Slaves = append(db.Slaves,slave)
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

	return db, redis
}

// 创建一个新的用户服务实例
func NewUserService() *UserService {
	var s UserService
	s.Db, s.Redis = initUserService()
	return &s
}

// Register 注册新用户
func (s *UserService) Register(ctx context.Context, req *pb.RegisterReq) (*pb.RegisterResp, error) {
    // 从数据库中检查邮箱是否已存在
	var user pb.RegisterReq
	result := s.Db.Slaves[0].Where("email = ?", req.Email).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// 保存用户信息到数据库
			user := pb.RegisterReq{
				Email: req.Email,
				Password: req.Password,
			}
			s.Db.Master.Create(&user)
			id := utils.GenerateId()
			s.Redis.Set(ctx, req.Email, id, 0)
			return &pb.RegisterResp{UserId: id}, nil
		}
		return nil, status.Errorf(codes.Internal, "数据库错误")
	}
	// 
    return nil, status.Errorf(codes.AlreadyExists, "邮箱已存在")
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {
	// 检查用户是否存在
	var user pb.RegisterReq
	result := s.Db.Slaves[0].Where("email = ?", req.Email).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "用户不存在")
		}
		return nil, status.Errorf(codes.Internal, "数据库错误")
	}
	// 检查密码是否正确
	if user.Password != req.Password {
		return nil, status.Errorf(codes.InvalidArgument, "密码错误")
	}
	// 生成token
	token := time.Now().Format("2006-01-02 15:04:05")
	// 将token存入redis
	err := s.Redis.Set(ctx, req.Email, token, 0).Err()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "redis错误")
	}
	return &pb.LoginResp{}, nil
}
