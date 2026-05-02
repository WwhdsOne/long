package mongostore

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"long/internal/core"
)

const shopPurchaseLogsCollectionName = "shop_purchase_logs"

type ShopPurchaseLogStore struct {
	collection   *mongo.Collection
	writeTimeout time.Duration
}

func NewShopPurchaseLogStore(db *mongo.Database, writeTimeout time.Duration) *ShopPurchaseLogStore {
	return &ShopPurchaseLogStore{
		collection:   db.Collection(shopPurchaseLogsCollectionName),
		writeTimeout: writeTimeout,
	}
}

func (s *ShopPurchaseLogStore) EnsureIndexes(ctx context.Context) error {
	if s == nil || s.collection == nil {
		return nil
	}
	_, err := s.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "nickname", Value: 1}, {Key: "purchased_at", Value: -1}},
			Options: options.Index().SetName("nickname_purchased_at"),
		},
		{
			Keys:    bson.D{{Key: "item_id", Value: 1}, {Key: "purchased_at", Value: -1}},
			Options: options.Index().SetName("item_id_purchased_at"),
		},
	})
	return err
}

func (s *ShopPurchaseLogStore) WriteShopPurchaseLog(ctx context.Context, item core.ShopPurchaseLog) error {
	if s == nil || s.collection == nil {
		return nil
	}
	writeCtx, cancel := withTimeout(ctx, s.writeTimeout)
	defer cancel()
	_, err := s.collection.InsertOne(writeCtx, item)
	return err
}
