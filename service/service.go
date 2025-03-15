package service

import (
	"bytedancemall/user/model"
	pb "bytedancemall/user/proto"
	p "bytedancemall/user/proto/auth"
	"bytedancemall/user/utils"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"os"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
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
	AuthServer p.AuthServiceClient
    pb.UnimplementedUserServiceServer
}

func initUserService() (Database, *redis.ClusterClient, p.AuthServiceClient) {
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

	// 创建表
	db.Master.AutoMigrate(&model.User{})

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

	// 连接认证中心
	conn, err := grpc.NewClient(configs.AuthServerConfig.Address)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := p.NewAuthServiceClient(conn)
	return db, redis, client
}

// 创建一个新的用户服务实例
func NewUserService() *UserService {
	var s UserService
	s.Db, s.Redis, s.AuthServer = initUserService()
	return &s
}

// Register 注册新用户
func (s *UserService) Register(ctx context.Context, req *pb.RegisterReq) (*pb.RegisterResp, error) {
	// 检查密码是否一致
	if req.Password != req.ConfirmPassword {
		return nil, status.Errorf(codes.InvalidArgument, "wrong password")
	}
	// 检查redis中是否已存在用户
	_, err := s.Redis.Get(ctx, req.Email).Result()
	if err != redis.Nil {
		return nil, status.Errorf(codes.AlreadyExists, "email already exists")
	}
	// 开启事务
	tx := s.Db.Master.Begin()
	if tx.Error != nil {
		return nil, status.Errorf(codes.Internal, "database error")
	}
	var user model.User
    // 检查邮箱是否存在数据库
	if result := s.Db.Master.Where("email = ?", req.Email).First(&user); result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// 保存用户信息到数据库
			// 写入数据库上锁
			user = model.User{
				Id: utils.GenerateId(),
				Username: uuid.New().String(),
				Email: req.Email,
				Password: utils.EncryptPassword(req.Password),
			}
			
			if result := s.Db.Master.Create(&user); result.Error != nil {
				tx.Rollback()
				return nil, status.Errorf(codes.Internal, "database error")
			}
			tx.Commit()
			// 将用户id存入redis
			if result := s.Redis.Set(ctx, user.Email, user.Id, 0); result.Err() != nil {
				if result := s.Redis.Set(ctx, strconv.FormatInt(user.Id, 10), user.Password, 0); result.Err() != nil {
					return &pb.RegisterResp{UserId: user.Id}, status.Errorf(codes.Internal, "set password error")
				}
				return &pb.RegisterResp{UserId: user.Id}, status.Errorf(codes.Internal, "set email error")
			}
			return &pb.RegisterResp{UserId: user.Id}, nil
		}
		return nil, status.Errorf(codes.Internal, "database error")
	}
    return nil, status.Errorf(codes.AlreadyExists, "email already exists")
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {
	// 检查用户是否存在
	userId, err := s.Redis.Get(ctx, req.Email).Result()
	if err != nil {
		if err != redis.Nil {
			return nil, status.Errorf(codes.Internal, "redis error")
		}
	}
	password := s.Redis.Get(ctx, userId).Val()
	if password == req.Password {
		id , err:= strconv.ParseInt(userId, 10, 64)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "parse id error")
		}
		return &pb.LoginResp{UserId: id}, nil
	}

	var user model.User
	result := s.Db.Master.Where("email = ?", req.Email).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "database error")
	}
	// 检查密码是否正确
	if user.Password != req.Password {
		return nil, status.Errorf(codes.InvalidArgument, "wrong password")
	}
	// 交给认证中心
	s.AuthServer.DeliverTokenByRPC(ctx, &p.DeliverTokenReq{UserId: user.Id})
	return &pb.LoginResp{UserId: user.Id}, nil
}
