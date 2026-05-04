package mongostore

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"long/internal/xlog"
)

const accessLogsCollectionName = "access_logs"

// AccessLogStore 负责请求日志写入。
type AccessLogStore struct {
	collection   *mongo.Collection
	writeTimeout time.Duration
}

// NewAccessLogStore 创建请求日志存储。
func NewAccessLogStore(db *mongo.Database, writeTimeout time.Duration) *AccessLogStore {
	return &AccessLogStore{
		collection:   db.Collection(accessLogsCollectionName),
		writeTimeout: writeTimeout,
	}
}

// EnsureIndexes 创建请求日志索引。
func (s *AccessLogStore) EnsureIndexes(ctx context.Context) error {
	if s == nil || s.collection == nil {
		return nil
	}
	_, err := s.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "created_at", Value: -1}}, Options: options.Index().SetName("created_at_desc")},
		{Keys: bson.D{{Key: "status_code", Value: 1}, {Key: "created_at", Value: -1}}, Options: options.Index().SetName("status_created_at_desc")},
		{Keys: bson.D{{Key: "path", Value: 1}, {Key: "created_at", Value: -1}}, Options: options.Index().SetName("path_created_at_desc")},
	})
	return err
}

// WriteAccessLog 写入单条请求日志。
func (s *AccessLogStore) WriteAccessLog(ctx context.Context, item xlog.AccessLogEntry) error {
	if s == nil || s.collection == nil {
		return nil
	}
	writeCtx, cancel := withTimeout(ctx, s.writeTimeout)
	defer cancel()
	_, err := s.collection.InsertOne(writeCtx, bson.M{
		"method":      item.Method,
		"path":        item.Path,
		"nickname":    item.Nickname,
		"body":        item.Body,
		"status_code": item.StatusCode,
		"latency_ms":  item.LatencyMs,
		"client_ip":   item.ClientIP,
		"user_agent":  item.UserAgent,
		"created_at":  item.CreatedAt,
	})
	return err
}
