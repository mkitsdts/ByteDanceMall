package service

import (
	"context"
	"log/slog"
)

func (s *SeckillSer) ComsumeSeckill() {
	go func() {
		for {
			// 从 Kafka 中读取消息
			m, err := s.KafkaReader.ReadMessage(context.Background())
			if err != nil {
				slog.Error("Failed to read message from Kafka", "error", err)
				continue
			}
			// 向order服务下单
			// 处理消息
			slog.Info("Received message from Kafka", "message", string(m.Value))
		}
	}()
}
