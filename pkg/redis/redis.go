package rds

import (
	"bytedancemall/llm/config"
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

var client *redis.Client

func GetRedisClient() *redis.Client {
	return client
}

func Init() {
	if len(config.Conf.Redis.Host) == 0 {
		panic("no redis host configured")
	}
	if len(config.Conf.Redis.Host) == 1 {
		client = redis.NewClient(&redis.Options{
			Addr:     config.Conf.Redis.Host[0] + ":" + fmt.Sprint(config.Conf.Redis.Port),
			Password: config.Conf.Redis.Password, // no password set
			DB:       config.Conf.Redis.DB,       // use default DB
			Protocol: 2,
		})
	} else {
		panic("redis cluster not supported yet")
	}
}

func CreateIndex() error {
	ctx := context.Background()

	client.FTDropIndexWithArgs(ctx,
		"smart_assistant_idx",
		&redis.FTDropIndexOptions{
			DeleteDocs: true,
		},
	)

	_, err := client.FTCreate(ctx,
		"smart_assistant_idx",
		&redis.FTCreateOptions{
			OnHash: true,
			Prefix: []any{"doc:"},
		},
		&redis.FieldSchema{
			FieldName: "content",
			FieldType: redis.SearchFieldTypeText,
		},
		&redis.FieldSchema{
			FieldName: "embedding",
			FieldType: redis.SearchFieldTypeVector,
			VectorArgs: &redis.FTVectorArgs{
				HNSWOptions: &redis.FTHNSWOptions{
					Dim:            1024,
					DistanceMetric: "L2",
					Type:           "FLOAT32",
				},
			},
		},
	).Result()

	if err != nil {
		slog.Error("failed to create index", "error", err)
		return err
	}
	return nil
}
