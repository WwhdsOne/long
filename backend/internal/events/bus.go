package events

import (
	"context"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"

	"long/internal/vote"
)

// RedisChangeBus 通过 Redis Pub/Sub 传播实时状态变化。
type RedisChangeBus struct {
	client  redis.UniversalClient
	channel string
}

// NewRedisChangeBus 创建一个 Redis 事件总线。
func NewRedisChangeBus(client redis.UniversalClient, channel string) *RedisChangeBus {
	return &RedisChangeBus{
		client:  client,
		channel: strings.TrimSpace(channel),
	}
}

// PublishChange 发布一条状态变更。
func (b *RedisChangeBus) PublishChange(ctx context.Context, change vote.StateChange) error {
	if b == nil || b.client == nil || b.channel == "" {
		return nil
	}
	if change.Timestamp == 0 {
		change.Timestamp = time.Now().Unix()
	}

	payload, err := sonic.Marshal(change)
	if err != nil {
		return err
	}

	return b.client.Publish(ctx, b.channel, payload).Err()
}

// Listen 持续消费 Pub/Sub 里的状态变更。
func (b *RedisChangeBus) Listen(ctx context.Context, handler func(context.Context, vote.StateChange) error) error {
	if b == nil || b.client == nil || b.channel == "" {
		return nil
	}

	pubsub := b.client.Subscribe(ctx, b.channel)
	defer pubsub.Close()

	if _, err := pubsub.Receive(ctx); err != nil {
		return err
	}

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case message, ok := <-ch:
			if !ok {
				return nil
			}

			var change vote.StateChange
			if err := sonic.Unmarshal([]byte(message.Payload), &change); err != nil {
				continue
			}
			if handler == nil {
				continue
			}
			if err := handler(ctx, change); err != nil {
				return err
			}
		}
	}
}
