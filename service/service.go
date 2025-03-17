package service

import (
	"bytedancemall/cart/model"
	pb "bytedancemall/cart/proto"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Database struct {
	Master *gorm.DB
	Slaves  []*gorm.DB
}

type CartService struct {
    Db Database
	Redis *redis.ClusterClient
    pb.UnimplementedCartServiceServer
}

func initCartService() (Database, *redis.ClusterClient) {
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
	db.Master.AutoMigrate(&model.CartItem{})
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
		slave.AutoMigrate(&model.CartItem{})
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
func NewCartService() *CartService {
	var s CartService
	s.Db, s.Redis = initCartService()
	return &s
}

// 添加商品至购物车
func (s *CartService) AddItem(ctx context.Context, req *pb.AddItemReq) (*pb.AddItemResp, error) {
	// 从请求中获取用户ID和商品信息
	userId := req.UserId
	item := req.Item
	// 判断数据库中是否已经存在该商品
	var cartItem model.CartItem
	
	if result := s.Db.Master.Where("user_id = ? AND product_id = ?", userId, item.ProductId).First(&cartItem); result.Error != nil {
		// 不存在
		if result.Error == gorm.ErrRecordNotFound {
			s.Db.Master.Create(&model.CartItem{
				UserId: userId,
				ProductID: item.ProductId,
				Quantity: item.Quantity,
			})
			return &pb.AddItemResp{}, nil
		}else {
			return nil, result.Error
		}
	}
	// 存在则更新数量
	s.Db.Master.Model(&cartItem).Update("quantity", cartItem.Quantity + item.Quantity)
	return &pb.AddItemResp{}, nil
}

// 获取购物车中的商品
func (s *CartService) GetCart(ctx context.Context, req *pb.GetCartReq) (*pb.GetCartResp, error) {
	// 从请求中获取用户ID
	userId := req.UserId

	// 从redis中获取购物车信息
	val, err := s.Redis.Get(ctx, strconv.FormatUint(uint64(userId), 10)).Result()
	if err == nil {
		var cartItems []model.CartItem
		json.Unmarshal([]byte(val), &cartItems)
		
		// 将购物车信息转换为proto格式
		var items []*pb.CartItem
		for _, item := range cartItems {
			items = append(items, &pb.CartItem{
				ProductId: item.ProductID,
				Quantity: item.Quantity,
			})
		}
		var resp pb.GetCartResp
		resp.Cart.UserId = userId
		resp.Cart.Items = items
		return &resp, nil
	}

	// 从数据库中获取购物车信息
	var items []model.CartItem
	s.Db.Master.Where("user_id = ?", userId).Find(&items)

	// 将购物车信息转换为proto格式
	var cartItems []*pb.CartItem
	for _, item := range items {
		cartItems = append(cartItems, &pb.CartItem{
			ProductId: item.ProductID,
			Quantity: item.Quantity,
		})
	}
	var resp pb.GetCartResp
	resp.Cart.UserId = userId
	resp.Cart.Items = cartItems

	// 将购物车信息存入redis
	redisValue, _ := json.Marshal(items)
	s.Redis.Set(ctx, strconv.FormatUint(uint64(userId), 10), redisValue, 0)
	return &resp, nil
}

// 清空购物车
func (s *CartService) EmptyCart(ctx context.Context, req *pb.EmptyCartReq) (*pb.EmptyCartResp, error) {
	userId := req.UserId
	s.Db.Master.Where("user_id = ?", userId).Delete(&model.CartItem{})
	s.Redis.Del(ctx, strconv.FormatUint(uint64(userId), 10))
	return &pb.EmptyCartResp{}, nil
}
