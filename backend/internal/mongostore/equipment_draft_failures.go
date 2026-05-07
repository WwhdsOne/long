package mongostore

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"long/internal/core"
)

const equipmentDraftFailuresCollectionName = "equipment_draft_failures"

// EquipmentDraftFailureStore 负责记录装备草稿生成失败上下文。
type EquipmentDraftFailureStore struct {
	collection   *mongo.Collection
	writeTimeout time.Duration
}

func NewEquipmentDraftFailureStore(db *mongo.Database, writeTimeout time.Duration) *EquipmentDraftFailureStore {
	return &EquipmentDraftFailureStore{
		collection:   db.Collection(equipmentDraftFailuresCollectionName),
		writeTimeout: writeTimeout,
	}
}

func (s *EquipmentDraftFailureStore) EnsureIndexes(ctx context.Context) error {
	if s == nil || s.collection == nil {
		return nil
	}
	_, err := s.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "created_at", Value: -1}}, Options: options.Index().SetName("created_at_desc")},
		{Keys: bson.D{{Key: "error_message", Value: 1}, {Key: "created_at", Value: -1}}, Options: options.Index().SetName("error_message_created_at_desc")},
	})
	return err
}

func (s *EquipmentDraftFailureStore) WriteEquipmentDraftFailure(ctx context.Context, item core.EquipmentDraftFailureLog) error {
	if s == nil || s.collection == nil {
		return nil
	}
	writeCtx, cancel := withTimeout(ctx, s.writeTimeout)
	defer cancel()
	_, err := s.collection.InsertOne(writeCtx, bson.M{
		"prompt":        item.Prompt,
		"draft":         item.Draft,
		"error_message": item.ErrorMessage,
		"raw_response":  item.RawResponse,
		"created_at":    item.CreatedAt,
	})
	return err
}
