package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

func (s *InventoryService) startConsumers() {
	go s.consumeCommitTopic()
	go s.consumeReleaseTopic()
}

func (s *InventoryService) consumeCommitTopic() {
	s.consumeInventoryTopic(inventoryCommitTopic, func(ctx context.Context, payload inventoryEventMessage) error {
		if payload.RecordID == "" || payload.OrderID == 0 {
			return fmt.Errorf("invalid commit message")
		}
		return s.usecase.CommitByRecordID(ctx, payload.RecordID, payload.OrderID)
	})
}

func (s *InventoryService) consumeReleaseTopic() {
	s.consumeInventoryTopic(inventoryRecoveryTopic, func(ctx context.Context, payload inventoryEventMessage) error {
		if payload.RecordID == "" {
			return fmt.Errorf("invalid release message")
		}
		return s.usecase.ReleaseByRecordID(ctx, payload.RecordID)
	})
}

func (s *InventoryService) consumeInventoryTopic(topic string, handler func(context.Context, inventoryEventMessage) error) {
	reader, ok := s.reader[topic]
	if !ok {
		slog.Warn("inventory kafka reader not configured", "topic", topic)
		return
	}

	for {
		msg, err := reader.FetchMessage(context.Background())
		if err != nil {
			slog.Error("failed to fetch inventory kafka message", "topic", topic, "error", err)
			time.Sleep(200 * time.Millisecond)
			continue
		}

		payload, err := parseInventoryEventMessage(msg.Value)
		if err == nil {
			err = handler(context.Background(), payload)
		}
		if err != nil {
			slog.Error("failed to handle inventory kafka message", "topic", topic, "error", err)
			time.Sleep(200 * time.Millisecond)
			continue
		}

		if err := reader.CommitMessages(context.Background(), msg); err != nil {
			slog.Error("failed to commit inventory kafka message", "topic", topic, "error", err)
		}
	}
}

type inventoryEventMessage struct {
	RecordID string `json:"record_id"`
	OrderID  uint64 `json:"order_id"`
}

func parseInventoryEventMessage(value []byte) (inventoryEventMessage, error) {
	var payload inventoryEventMessage
	if err := json.Unmarshal(value, &payload); err == nil && payload.RecordID != "" {
		return payload, nil
	}
	if len(value) == 0 {
		return inventoryEventMessage{}, fmt.Errorf("empty inventory event message")
	}
	return inventoryEventMessage{RecordID: string(value)}, nil
}
