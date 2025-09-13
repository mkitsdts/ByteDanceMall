package service

import (
	"bytedancemall/cart/model"
	"bytedancemall/cart/pkg"
	pb "bytedancemall/cart/proto"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 添加商品至购物车
func (s *CartService) AddItem(ctx context.Context, req *pb.AddItemReq) (*pb.AddItemResp, error) {
	// TODO:
	// 1. 从 MySQL 中获取商品价格
	// 2. 将商品添加到 MySQL 中
	// 3. 删除旧页码

	// 1. 从 MySQL 中获取商品价格
	tx := pkg.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	maxRetries := 3
	item := &model.ProductItem{}
	for i := range maxRetries {
		if err := tx.Model(item).Where("product_id = ?", req.ProductId).First(item).Error; err == nil {
			break
		} else if err == gorm.ErrRecordNotFound {
			tx.Rollback()
			return &pb.AddItemResp{
				Result: false,
			}, fmt.Errorf("product not found")
		}
		time.Sleep(10 << i * time.Millisecond)
		if i == maxRetries-1 {
			tx.Rollback()
			return &pb.AddItemResp{
				Result: false,
			}, fmt.Errorf("failed to query cart item")
		}
	}

	now := uint64(time.Now().Unix())

	// 2. 将商品添加到 MySQL 中
	cartItem := model.CartItem{
		UserID:      req.UserId,
		ProductID:   req.ProductId,
		OriginPrice: item.Price,
		Quantity:    1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Upsert 更新用户购物车商品数量
	for i := range maxRetries {
		res := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "product_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"quantity":   gorm.Expr("quantity + 1"),
				"updated_at": now,
			}),
		}).Create(&cartItem)
		if res.Error != nil {
			return &pb.AddItemResp{Result: false}, fmt.Errorf("upsert cart item failed: %w", res.Error)
		}
		time.Sleep(10 << i * time.Millisecond)
		if i == maxRetries-1 {
			return &pb.AddItemResp{
				Result: false,
			}, fmt.Errorf("failed to get cart item count from redis")
		}
	}
	for i := range maxRetries {
		if err := tx.Commit().Error; err == nil {
			go deleteItemFromCache(fmt.Sprintf("cart_%d", req.UserId))
			return &pb.AddItemResp{
				Result: true,
			}, nil
		}
		time.Sleep(10 << i * time.Millisecond)
		if i == maxRetries-1 {
			tx.Rollback()
			return &pb.AddItemResp{
				Result: false,
			}, fmt.Errorf("failed to commit transaction")
		}
	}
	return &pb.AddItemResp{
		Result: false,
	}, nil
}

// 获取购物车中的商品
func (s *CartService) GetCart(ctx context.Context, req *pb.GetCartReq) (*pb.GetCartResp, error) {
	// TODO:
	// 1.读取缓存中的购物车数据
	// 2.缓存未命中，从 MySQL 中读取购物车数据
	// 3.将购物车数据写入缓存

	cartKey := fmt.Sprintf("cart_%d", req.UserId)
	cart := model.Cart{}
	maxRetries := 3
	// 1.读取缓存中的购物车数据
	for i := range maxRetries {
		res, err := pkg.GetRedisCli().Get(ctx, cartKey).Result()
		if err == nil {
			_ = json.Unmarshal([]byte(res), &cart)
			return &pb.GetCartResp{
				Result: true,
				Items: func() []*pb.CartItem {
					items := make([]*pb.CartItem, 0, len(cart.Items))
					for _, item := range cart.Items {
						items = append(items, &pb.CartItem{
							ProductId:   item.ProductID,
							OriginPrice: item.OriginPrice,
							Quantity:    item.Quantity,
						})
					}
					return items
				}(),
			}, nil
		} else if err == redis.Nil {
			break
		}
		time.Sleep(10 << i * time.Millisecond)
		if i == maxRetries-1 {
			return &pb.GetCartResp{
				Result: false,
			}, fmt.Errorf("failed to get cart from cache")
		}
	}

	// 2.缓存未命中，从 MySQL 中读取购物车数据
	tx := pkg.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	for i := range maxRetries {
		if err := tx.Model(&model.CartItem{}).Where("user_id = ?", req.UserId).Find(&cart.Items).Error; err == nil {
			break
		} else if err == gorm.ErrRecordNotFound {
			tx.Rollback()
			return &pb.GetCartResp{
				Result: true,
				Items:  []*pb.CartItem{},
			}, nil
		}
		time.Sleep(10 << i * time.Millisecond)
		if i == maxRetries-1 {
			tx.Rollback()
			return &pb.GetCartResp{
				Result: false,
			}, fmt.Errorf("failed to get cart items from db")
		}
	}
	for i := range maxRetries {
		if err := tx.Commit().Error; err == nil {
			break
		}
		time.Sleep(10 << i * time.Millisecond)
		if i == maxRetries-1 {
			tx.Rollback()
			return &pb.GetCartResp{
				Result: false,
			}, fmt.Errorf("failed to commit transaction")
		}
	}

	// 3.将购物车数据写入缓存
	data, _ := json.Marshal(cart)
	for i := range maxRetries {
		if err := pkg.GetRedisCli().Set(ctx, cartKey, data, 0).Err(); err == nil {
			break
		}
		time.Sleep(10 << i * time.Millisecond)
		if i == maxRetries-1 {
			return &pb.GetCartResp{
				Result: false,
			}, fmt.Errorf("failed to set cart to redis")
		}
	}

	return &pb.GetCartResp{
		Result: true,
		Items: func() []*pb.CartItem {
			items := make([]*pb.CartItem, 0, len(cart.Items))
			for _, item := range cart.Items {
				items = append(items, &pb.CartItem{
					ProductId:   item.ProductID,
					OriginPrice: item.OriginPrice,
					Quantity:    item.Quantity,
				})
			}
			return items
		}(),
	}, nil
}

func (s *CartService) RemoveItem(ctx context.Context, req *pb.RemoveItemReq) (*pb.RemoveItemResp, error) {
	// TODO:
	// 1.从 MySQL 中删除购物车商品
	// 2.删除缓存
	tx := pkg.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	maxRetries := 3
	for i := range maxRetries {
		if err := tx.Where("user_id = ? AND product_id = ?", req.UserId, req.ProductId).Delete(&model.CartItem{}).Error; err == nil {
			break
		}
		time.Sleep(10 << i * time.Millisecond)
		if i == maxRetries-1 {
			tx.Rollback()
			return &pb.RemoveItemResp{
				Result: false,
			}, fmt.Errorf("failed to delete cart item from db")
		}
	}
	for i := range maxRetries {
		if err := tx.Commit().Error; err == nil {
			go deleteItemFromCache(fmt.Sprintf("cart_%d", req.UserId))
			return &pb.RemoveItemResp{
				Result: true,
			}, nil
		}
		time.Sleep(10 << i * time.Millisecond)
		if i == maxRetries-1 {
			tx.Rollback()
			return &pb.RemoveItemResp{
				Result: false,
			}, fmt.Errorf("failed to commit transaction")
		}
	}
	return &pb.RemoveItemResp{
		Result: false,
	}, nil
}
