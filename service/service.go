package service

import (
	"bytedancemall/order/model"
	pb "bytedancemall/order/proto"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type OrderService struct {
	LocalDb *gorm.DB
	Redis   *redis.ClusterClient
	pb.UnimplementedOrderServiceServer
}

func initOrderService() (*gorm.DB, *redis.ClusterClient) {
	// 从configs.json中读取数据库和Redis配置
	// 配置mysql和redis集群
	file, err := os.Open("configs.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var configs model.Configs
	if err = decoder.Decode(&configs); err != nil {
		panic(err)
	}

	// 生成mysql的dsn
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		configs.MysqlConfig.User,
		configs.MysqlConfig.Password,
		configs.MysqlConfig.Host,
		configs.MysqlConfig.Port,
		configs.MysqlConfig.Database,
	)

	// 连接mysql
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// 生成redis集群的地址
	var redisAddrs []string
	for _, v := range configs.RedisConfig.Configs {
		redisAddrs = append(redisAddrs, v.Host+":"+v.Port)
	}
	// 初始化redis
	redis := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    redisAddrs,
		Password: configs.RedisConfig.Password,
	})

	return db, redis
}

// 创建一个新的订单服务实例
func NewOrderService() *OrderService {
	var s OrderService
	s.LocalDb, s.Redis = initOrderService()
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
			OrderId:       orderId,
			UserId:        req.UserId,
			UserEmail:     req.Email,
			UserCurrency:  req.UserCurrency,
			StreetAddress: req.Address.StreetAddress,
			City:          req.Address.City,
			State:         req.Address.State,
			ZipCode:       req.Address.ZipCode,
			ItemId:        item.Item.ProductId,
			Quantity:      item.Item.Quantity,
			Cost:          item.Cost,
		})
	}
	s.LocalDb.Create(&orders)
	var result pb.PlaceOrderResp
	result.Order.OrderId = orderId
	return &result, nil
}

func (s *OrderService) ListOrder(ctx context.Context, req *pb.ListOrderReq) (*pb.ListOrderResp, error) {
	// 从Redis中获取订单
	val := s.Redis.Get(ctx, strconv.FormatUint(uint64(req.UserId), 10)).Val()
	if val != "" {
		// 返回订单
		var pbOrders []*pb.Order
		orders := []model.Order{}
		json.Unmarshal([]byte(val), &orders)
		for _, order := range orders {
			pbOrders = append(pbOrders, &pb.Order{
				OrderId:      order.OrderId,
				UserId:       order.UserId,
				Email:        order.UserEmail,
				UserCurrency: order.UserCurrency,
				Address: &pb.Address{
					StreetAddress: order.StreetAddress,
					City:          order.City,
					State:         order.State,
					ZipCode:       order.ZipCode,
				},
			})
		}
		return &pb.ListOrderResp{Orders: pbOrders}, nil
	}

	// redis中没有订单，从数据库中获取订单
	var orders []model.Order
	s.LocalDb.Where("user_id = ?", req.UserId).Find(&orders)
	// 返回订单
	var pbOrders []*pb.Order
	for _, order := range orders {
		pbOrders = append(pbOrders, &pb.Order{
			OrderId:      order.OrderId,
			UserId:       order.UserId,
			Email:        order.UserEmail,
			UserCurrency: order.UserCurrency,
			Address: &pb.Address{
				StreetAddress: order.StreetAddress,
				City:          order.City,
				State:         order.State,
				ZipCode:       order.ZipCode,
			},
		})
	}

	// 将订单存入Redis
	bytes, _ := json.Marshal(orders)
	s.Redis.Set(ctx, strconv.FormatUint(uint64(req.UserId), 10), bytes, 5*time.Minute)
	return &pb.ListOrderResp{Orders: pbOrders}, nil
}

// 完成订单
func (s *OrderService) MarkOrderPaid(ctx context.Context, req *pb.MarkOrderPaidReq) (*pb.MarkOrderPaidResp, error) {
	// 从数据库中获取订单
	var orders []model.Order
	s.LocalDb.Where("order_id = ?", req.OrderId).First(&orders)
	// 更新订单状态
	for _, order := range orders {
		order.Paid = true
		s.LocalDb.Save(&order)
		// 删除缓存
		s.Redis.Del(ctx, order.OrderId)
	}
	return nil, nil
}
