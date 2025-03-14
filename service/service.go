package service

import (
	"bytedancemall/order/model"
	pb "bytedancemall/order/proto"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"github.com/google/uuid"
)

type Database struct {
	Master *gorm.DB
	Slaves  []*gorm.DB
}

type OrderService struct {
    Db Database
	Redis *redis.ClusterClient
    pb.UnimplementedOrderServiceServer
}

func initOrderService() (Database, *redis.ClusterClient) {
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

// 创建一个新的订单服务实例
func NewOrderService() *OrderService {
	var s OrderService
	s.Db, s.Redis = initOrderService()
	return &s
}

// 下单
func (s *OrderService) PlaceOrder(ctx context.Context, req *pb.PlaceOrderReq) (*pb.PlaceOrderResp, error) {
    // 为请求的每一个物品生成一个订单
	var orders []model.Order
	orderId := uuid.New().String()
	for _, item := range req.OrderItems {
		// 生成订单
		orders = append(orders, model.Order{
			OrderId: orderId,
			UserId: req.UserId,
			UserEmail: req.Email,
			UserCurrency: req.UserCurrency,
			StreetAddress: req.Address.StreetAddress,
			City: req.Address.City,
			State: req.Address.State,
			ZipCode: req.Address.ZipCode,
			ItemId: item.Item.ProductId,
			Quantity: item.Item.Quantity,
			Cost: item.Cost,
		})
	}
	s.Db.Master.Create(&orders)
	var result pb.PlaceOrderResp
	result.Order.OrderId = orderId
	return &result, nil
}

func (s *OrderService) ListOrder(ctx context.Context, req *pb.ListOrderReq) (*pb.ListOrderResp, error) {
	// 从数据库中获取订单
	var orders []model.Order
	s.Db.Master.Where("user_id = ?", req.UserId).Find(&orders)
	// 返回订单
	var pbOrders []*pb.Order
	for _, order := range orders {
		pbOrders = append(pbOrders, &pb.Order{
			OrderId: order.OrderId,
			UserId: order.UserId,
			Email: order.UserEmail,
			UserCurrency: order.UserCurrency,
			Address: &pb.Address{
				StreetAddress: order.StreetAddress,
				City: order.City,
				State: order.State,
				ZipCode: order.ZipCode,
			},
		})
	}
	return &pb.ListOrderResp{Orders: pbOrders}, nil
}

// 完成订单
func (s *OrderService) MarkOrderPaid(ctx context.Context, req *pb.MarkOrderPaidReq) (*pb.MarkOrderPaidResp, error) {
	// 从数据库中获取订单
	var orders []model.Order
	s.Db.Master.Where("order_id = ?", req.OrderId).First(&orders)
	// 更新订单状态
	for _, order := range orders {
		order.Paid = true
		s.Db.Master.Save(&order)
	}
	return nil, nil
}
