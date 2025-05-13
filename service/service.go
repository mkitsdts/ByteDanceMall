package service

import (
	pb "bytedancemall/product/proto"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Database struct {
	Master *gorm.DB
	Slaves []*gorm.DB
}

type ProductService struct {
	Db    Database
	Redis *redis.ClusterClient
	pb.UnimplementedProductCatalogServiceServer
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
	configs := Configs{}
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
	db.Master.AutoMigrate(&Product{})
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
		slave.AutoMigrate(&Product{})
		db.Slaves = append(db.Slaves, slave)
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

// 创建一个新的用户服务实例
func NewProductService() *ProductService {
	var s ProductService
	s.Db, s.Redis = initUserService()
	return &s
}

// 根据页数获取商品
func (s *ProductService) ListProducts(ctx context.Context, req *pb.ListProductsReq) (*pb.ListProductsResp, error) {
	// 计算偏移量
	offset := ((int64(req.Page)) - 1) * req.PageSize
	// 从数据库中获取商品
	var products []Product
	result := s.Db.Master.Limit(int(req.PageSize)).Offset(int(offset)).Find(&products)
	if result.Error != nil {
		return nil, result.Error
	}
	// 将商品转换为proto格式
	var respProducts []*pb.Product
	for _, v := range products {
		respProducts = append(respProducts, &pb.Product{
			Id:          v.ID,
			Name:        v.Name,
			Description: v.Description,
			Price:       v.Price,
		})
	}
	// 返回商品
	return &pb.ListProductsResp{
		Products: respProducts,
	}, nil
}

// 获取单个产品
func (s *ProductService) GetProduct(ctx context.Context, req *pb.GetProductReq) (*pb.GetProductResp, error) {
	var product Product
	result := s.Db.Master.First(&product, req.Id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &pb.GetProductResp{
		Product: &pb.Product{
			Id:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
		},
	}, nil
}

func (s *ProductService) ListProSearchProductsducts(ctx context.Context, req *pb.SearchProductsReq) (*pb.SearchProductsResp, error) {
	// 参数验证
	if req.Query == "" {
		return &pb.SearchProductsResp{Results: []*pb.Product{}}, nil
	}

	// 尝试从缓存获取结果
	cacheKey := fmt.Sprintf("product:search:%s", req.Query)
	cachedData, err := s.Redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		// 缓存命中
		var products []*pb.Product
		if err := json.Unmarshal(cachedData, &products); err == nil {
			return &pb.SearchProductsResp{Results: products}, nil
		}
	}

	// 缓存未命中，从数据库查询
	var products []Product

	// 限制返回的最大记录数，避免返回过多数据
	const maxResults = 100

	result := s.Db.Master.Where("name LIKE ?", "%"+req.Query+"%").Limit(maxResults).Find(&products)

	if result.Error != nil {
		return nil, result.Error
	}

	// 将商品转换为proto格式
	var respProducts []*pb.Product
	for _, v := range products {
		respProducts = append(respProducts, &pb.Product{
			Id:          v.ID,
			Name:        v.Name,
			Description: v.Description,
			Price:       v.Price,
		})
	}

	// 将结果缓存到Redis，设置过期时间为5分钟
	if len(respProducts) > 0 {
		cacheData, err := json.Marshal(respProducts)
		if err == nil {
			s.Redis.Set(ctx, cacheKey, cacheData, 5*time.Minute)
		}
	}

	return &pb.SearchProductsResp{
		Results: respProducts,
	}, nil
}
