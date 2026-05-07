package mongostore

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"long/internal/core"
)

const staminaPurchaseLogsCollectionName = "stamina_purchase_logs"

type StaminaPurchaseLogStore struct {
	collection   *mongo.Collection
	writeTimeout time.Duration
}

func NewStaminaPurchaseLogStore(db *mongo.Database, writeTimeout time.Duration) *StaminaPurchaseLogStore {
	return &StaminaPurchaseLogStore{
		collection:   db.Collection(staminaPurchaseLogsCollectionName),
		writeTimeout: writeTimeout,
	}
}

func (s *StaminaPurchaseLogStore) EnsureIndexes(ctx context.Context) error {
	if s == nil || s.collection == nil {
		return nil
	}
	_, err := s.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "nickname", Value: 1}, {Key: "purchased_at", Value: -1}},
			Options: options.Index().SetName("nickname_purchased_at"),
		},
		{
			Keys:    bson.D{{Key: "triggered_risk_ban", Value: 1}, {Key: "purchased_at", Value: -1}},
			Options: options.Index().SetName("triggered_risk_ban_purchased_at"),
		},
	})
	return err
}

func (s *StaminaPurchaseLogStore) WriteStaminaPurchaseLog(ctx context.Context, item core.StaminaPurchaseLog) error {
	if s == nil || s.collection == nil {
		return nil
	}
	writeCtx, cancel := withTimeout(ctx, s.writeTimeout)
	defer cancel()
	_, err := s.collection.InsertOne(writeCtx, item)
	return err
}
