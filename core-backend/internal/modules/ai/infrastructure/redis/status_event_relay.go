package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/port"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	StatusEventChannelPrefix = "ingestion:status:"
	StatusEventListKey       = "ingestion:status:events"
	MaxReplayBufferSize      = 1000
)

type StatusEventRelay struct {
	client *redis.Client
	logger core.Logger
}

var _ port.StatusEventRelay = (*StatusEventRelay)(nil)

func NewStatusEventRelay(client *redis.Client, logger core.Logger) *StatusEventRelay {
	return &StatusEventRelay{
		client: client,
		logger: logger,
	}
}

func (r *StatusEventRelay) PublishStatusUpdate(ctx context.Context, projection *entity.IngestionStatusProjection) error {
	event := map[string]any{
		"documentId":           projection.DocumentID,
		"accountId":            projection.AccountID,
		"userId":               projection.UserID,
		"currentStage":         string(projection.CurrentStage),
		"isTerminal":           projection.IsTerminal,
		"startedAt":            projection.StartedAt,
		"updatedAt":            projection.UpdatedAt,
		"completedAt":          projection.CompletedAt,
		"lastError":            projection.LastError,
		"chunksProcessedCount": projection.ChunksProcessedCount,
		"chunksFailedCount":    projection.ChunksFailedCount,
		"eventSequence":        projection.LastEventSequence,
		"timestamp":            time.Now().UTC().UnixMilli(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		r.logger.Error("Failed to marshal status event", core.Error(err))
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	channel := StatusEventChannelPrefix + projection.DocumentID.String()
	if err := r.client.Publish(ctx, channel, data).Err(); err != nil {
		r.logger.Error("Failed to publish status event", core.Error(err))
		return fmt.Errorf("failed to publish: %w", err)
	}

	if err := r.appendToReplayBuffer(ctx, projection); err != nil {
		r.logger.Warn("Failed to append to replay buffer", core.Error(err))
	}

	return nil
}

func (r *StatusEventRelay) appendToReplayBuffer(ctx context.Context, projection *entity.IngestionStatusProjection) error {
	event := map[string]any{
		"documentId":    projection.DocumentID.String(),
		"accountId":     projection.AccountID.String(),
		"eventSequence": projection.LastEventSequence,
		"currentStage":  string(projection.CurrentStage),
		"isTerminal":    projection.IsTerminal,
		"timestamp":     time.Now().UTC().UnixMilli(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s:%s", StatusEventListKey, projection.DocumentID.String())
	if err := r.client.RPush(ctx, key, data).Err(); err != nil {
		return err
	}

	if err := r.client.LTrim(ctx, key, -MaxReplayBufferSize, -1).Err(); err != nil {
		return err
	}

	if err := r.client.Expire(ctx, key, 24*time.Hour).Err(); err != nil {
		return err
	}

	return nil
}

func (r *StatusEventRelay) SubscribeToDocument(ctx context.Context, documentID uuid.UUID) *redis.PubSub {
	channel := StatusEventChannelPrefix + documentID.String()
	return r.client.Subscribe(ctx, channel)
}

func (r *StatusEventRelay) GetReplayEvents(ctx context.Context, documentID uuid.UUID, sinceSequence int64) ([]map[string]any, error) {
	key := fmt.Sprintf("%s:%s", StatusEventListKey, documentID.String())

	events, err := r.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var result []map[string]any
	for _, e := range events {
		var event map[string]any
		if err := json.Unmarshal([]byte(e), &event); err != nil {
			continue
		}

		seq, ok := event["eventSequence"].(float64)
		if !ok {
			continue
		}

		if int64(seq) > sinceSequence {
			result = append(result, event)
		}
	}

	return result, nil
}

func (r *StatusEventRelay) PublishAccountUpdate(ctx context.Context, accountID uuid.UUID, projections []*entity.IngestionStatusProjection) error {
	event := map[string]any{
		"accountId":   accountID.String(),
		"projections": projections,
		"timestamp":   time.Now().UTC().UnixMilli(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		r.logger.Error("Failed to marshal account event", core.Error(err))
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	channel := fmt.Sprintf("%saccount:%s", StatusEventChannelPrefix, accountID.String())
	if err := r.client.Publish(ctx, channel, data).Err(); err != nil {
		r.logger.Error("Failed to publish account event", core.Error(err))
		return fmt.Errorf("failed to publish: %w", err)
	}

	return nil
}
