package mongostore

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"long/internal/xlog"
)

const systemLogsCollectionName = "system_logs"

// SystemLogStore 负责系统日志写入。
type SystemLogStore struct {
	collection   *mongo.Collection
	writeTimeout time.Duration
}

func NewSystemLogStore(db *mongo.Database, writeTimeout time.Duration) *SystemLogStore {
	return &SystemLogStore{
		collection:   db.Collection(systemLogsCollectionName),
		writeTimeout: writeTimeout,
	}
}

func (s *SystemLogStore) EnsureIndexes(ctx context.Context) error {
	if s == nil || s.collection == nil {
		return nil
	}
	_, err := s.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "created_at", Value: -1}}, Options: options.Index().SetName("created_at_desc")},
		{Keys: bson.D{{Key: "level", Value: 1}, {Key: "created_at", Value: -1}}, Options: options.Index().SetName("level_created_at_desc")},
		{Keys: bson.D{{Key: "module", Value: 1}, {Key: "created_at", Value: -1}}, Options: options.Index().SetName("module_created_at_desc")},
	})
	return err
}

func (s *SystemLogStore) WriteSystemLog(ctx context.Context, item xlog.SystemLogEntry) error {
	if s == nil || s.collection == nil {
		return nil
	}
	writeCtx, cancel := withTimeout(ctx, s.writeTimeout)
	defer cancel()
	_, err := s.collection.InsertOne(writeCtx, bson.M{
		"level":      item.Level,
		"module":     item.Module,
		"message":    item.Message,
		"request_id": item.RequestID,
		"created_at": item.CreatedAt,
	})
	return err
}
