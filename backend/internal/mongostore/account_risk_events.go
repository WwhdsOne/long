package mongostore

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"long/internal/core"
)

const accountRiskEventsCollectionName = "account_risk_events"

// AccountRiskEventStore 负责账号风险历史事件写入。
type AccountRiskEventStore struct {
	collection   *mongo.Collection
	writeTimeout time.Duration
}

func NewAccountRiskEventStore(db *mongo.Database, writeTimeout time.Duration) *AccountRiskEventStore {
	return &AccountRiskEventStore{
		collection:   db.Collection(accountRiskEventsCollectionName),
		writeTimeout: writeTimeout,
	}
}

func (s *AccountRiskEventStore) EnsureIndexes(ctx context.Context) error {
	if s == nil || s.collection == nil {
		return nil
	}
	_, err := s.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("created_at_desc"),
		},
		{
			Keys:    bson.D{{Key: "nickname", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("nickname_created_at_desc"),
		},
		{
			Keys:    bson.D{{Key: "event_type", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("event_type_created_at_desc"),
		},
		{
			Keys:    bson.D{{Key: "ban_until_after", Value: -1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("ban_until_after_created_at_desc"),
		},
	})
	return err
}

func (s *AccountRiskEventStore) WriteAccountRiskEvent(ctx context.Context, item core.AccountRiskEventLog) error {
	if s == nil || s.collection == nil {
		return nil
	}
	writeCtx, cancel := withTimeout(ctx, s.writeTimeout)
	defer cancel()
	_, err := s.collection.InsertOne(writeCtx, bson.M{
		"nickname":        item.Nickname,
		"event_type":      item.EventType,
		"points":          item.Points,
		"score_after":     item.ScoreAfter,
		"ban_until_after": item.BanUntilAfter,
		"payload":         item.Payload,
		"created_at":      item.CreatedAt,
	})
	return err
}
