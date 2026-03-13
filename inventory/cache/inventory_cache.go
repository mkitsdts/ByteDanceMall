package cache

import (
	"context"
	"fmt"
	"hash/crc32"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	BucketCount      = 4
	CommittingMarker = int64(-1)
)

type InventoryCache struct {
	client       *redis.Client
	deductScript *redis.Script
}

func New(client *redis.Client) *InventoryCache {
	const deductLua = `
local product_key = KEYS[1]
local amount = tonumber(ARGV[1])
local current_stock = tonumber(redis.call("GET", product_key) or "-1")
if current_stock == -1 then
    return -1
end
if current_stock < amount then
    return -2
end
redis.call("DECRBY", product_key, amount)
return current_stock - amount
`

	return &InventoryCache{
		client:       client,
		deductScript: redis.NewScript(deductLua),
	}
}

func ProductBucketKey(productID uint64, bucket int32) string {
	return fmt.Sprintf("product:%d:bucket:%d", productID, bucket)
}

func ProductAvailableStockKey(productID uint64) string {
	return fmt.Sprintf("product:%d:available_stock", productID)
}

func (c *InventoryCache) Reserve(ctx context.Context, productID uint64, recordID string, amount uint64) (int32, int64, error) {
	startBucket := int32(crc32.ChecksumIEEE([]byte(recordID)) % BucketCount)
	var insufficientCount int32

	for offset := range BucketCount {
		bucket := (startBucket + int32(offset)) % BucketCount
		result, err := c.deductScript.Run(ctx, c.client, []string{ProductBucketKey(productID, bucket)}, amount).Int64()
		if err != nil {
			continue
		}
		switch result {
		case CommittingMarker:
			continue
		case -2:
			insufficientCount++
			continue
		default:
			if result >= 0 {
				return bucket, result, nil
			}
		}
	}

	if insufficientCount == BucketCount {
		return 0, 0, ErrInsufficientCacheStock
	}
	return 0, 0, ErrInventoryCacheBusy
}

func (c *InventoryCache) RestoreReservation(ctx context.Context, productID uint64, bucket int32, amount uint64) error {
	return c.client.IncrBy(ctx, ProductBucketKey(productID, bucket), int64(amount)).Err()
}

func (c *InventoryCache) MarkBucketCommitting(ctx context.Context, productID uint64, bucket int32) error {
	return c.client.Set(ctx, ProductBucketKey(productID, bucket), CommittingMarker, 30*time.Second).Err()
}

func (c *InventoryCache) DeleteBucket(ctx context.Context, productID uint64, bucket int32) error {
	return c.client.Del(ctx, ProductBucketKey(productID, bucket)).Err()
}

func (c *InventoryCache) DeleteProductBuckets(ctx context.Context, productID uint64) error {
	keys := make([]string, 0, BucketCount)
	for bucket := int32(0); bucket < BucketCount; bucket++ {
		keys = append(keys, ProductBucketKey(productID, bucket))
	}
	return c.client.Del(ctx, keys...).Err()
}

func (c *InventoryCache) GetAvailableStock(ctx context.Context, productID uint64) (uint64, bool, error) {
	value, err := c.client.Get(ctx, ProductAvailableStockKey(productID)).Result()
	if err == redis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}

	stock, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, false, err
	}
	return stock, true, nil
}

func (c *InventoryCache) SetAvailableStock(ctx context.Context, productID, stock uint64) error {
	return c.client.Set(ctx, ProductAvailableStockKey(productID), stock, 0).Err()
}

func (c *InventoryCache) DeleteAvailableStock(ctx context.Context, productID uint64) error {
	return c.client.Del(ctx, ProductAvailableStockKey(productID)).Err()
}

func (c *InventoryCache) SeedProductBuckets(ctx context.Context, productID, totalStock uint64) error {
	for bucket := int32(0); bucket < BucketCount; bucket++ {
		stock := totalStock / BucketCount
		if bucket < int32(totalStock%BucketCount) {
			stock++
		}
		if err := c.client.Set(ctx, ProductBucketKey(productID, bucket), stock, 0).Err(); err != nil {
			return err
		}
	}
	return nil
}
