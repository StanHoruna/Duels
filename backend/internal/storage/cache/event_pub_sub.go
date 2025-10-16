package cache

import (
	"context"
	"duels-api/internal/model"
	"duels-api/pkg/apperrors"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type EventPubSub struct {
	c *redis.Client
}

func NewEventPubSub(c *redis.Client) *EventPubSub {
	return &EventPubSub{c: c}
}

func (s *EventPubSub) Subscribe(ctx context.Context, userID uuid.UUID) *redis.PubSub {
	return s.c.Subscribe(ctx, s.UserEventKey(userID))
}

func (s *EventPubSub) Publish(ctx context.Context, userID uuid.UUID, event *model.Event) error {
	err := s.c.Publish(ctx, s.UserEventKey(userID), event).Err()
	if err != nil {
		return apperrors.Internal("failed to publish event", err)
	}

	return nil
}

func (s *EventPubSub) UserEventKey(userID uuid.UUID) string {
	return fmt.Sprintf("user:%s:events", userID.String())
}
