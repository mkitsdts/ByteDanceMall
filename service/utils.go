package service

import (
	"bytedancemall/cart/model"
	"bytedancemall/cart/pkg"
	"context"
	"encoding/json"
	"log/slog"
	"time"
)

func deleteItemFromCache(key string) {
	for i := range 100 {
		if err := pkg.GetRedisCli().Del(context.Background(), key).Err(); err == nil {
			break
		}
		time.Sleep((10 << i) * time.Millisecond)
		if i == 99 {
			// 这里是非常危险的事情
			slog.Error("deleteItemFromPageCache failed", "key", key)
		}
	}
}

func loopUpdateProductInfo() {
	slog.Info("start loopUpdateProductInfo")
	type ProductInfoMsg struct {
		ProductID   uint64
		OriginPrice float64
	}
	maxRetries := 5
	// TODO: 监听 Kafka 的 product_info_update 主题，更新商品信息
	for {
		msg, err := pkg.GetKafkaReader("product_info_update").FetchMessage(context.Background())
		if err != nil {
			time.Sleep(1 * time.Second)
		}
		slog.Info("Received message from Kafka", "topic", "product_info_update", "key", string(msg.Key), "value", string(msg.Value))
		info := &ProductInfoMsg{}
		if err := json.Unmarshal(msg.Value, info); err != nil {
			slog.Error("failed to unmarshal product info", "error", err)
			continue
		}
		slog.Info("Received product info update", "product_id", info.ProductID, "origin_price", info.OriginPrice)
		// 更新 MySQL
		tx := pkg.DB().Begin()
		for i := range maxRetries {
			if err := tx.Model(&model.CartItem{}).Where("product_id = ?", info.ProductID).Updates(map[string]any{
				"origin_price": info.OriginPrice,
				"updated_at":   uint64(time.Now().Unix()),
			}).Error; err == nil {
				break
			}
			time.Sleep(10 << i * time.Millisecond)
			if i == maxRetries-1 {
				slog.Error("failed to update product info in cart items", "product_id", info.ProductID, "error", err)
				tx.Rollback()
				continue
			}
		}
		tx.Commit()
	}
}
