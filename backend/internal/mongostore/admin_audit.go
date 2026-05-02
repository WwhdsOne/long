package mongostore

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"long/internal/core"
)

const adminAuditCollectionName = "admin_audit_logs"

// AdminAuditStore 负责后台审计日志写入。
type AdminAuditStore struct {
	collection   *mongo.Collection
	writeTimeout time.Duration
}

func NewAdminAuditStore(db *mongo.Database, writeTimeout time.Duration) *AdminAuditStore {
	return &AdminAuditStore{
		collection:   db.Collection(adminAuditCollectionName),
		writeTimeout: writeTimeout,
	}
}

func (s *AdminAuditStore) EnsureIndexes(ctx context.Context) error {
	if s == nil || s.collection == nil {
		return nil
	}
	_, err := s.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "created_at", Value: -1}}, Options: options.Index().SetName("created_at_desc")},
		{Keys: bson.D{{Key: "operator", Value: 1}, {Key: "created_at", Value: -1}}, Options: options.Index().SetName("operator_created_at_desc")},
		{Keys: bson.D{{Key: "action", Value: 1}, {Key: "created_at", Value: -1}}, Options: options.Index().SetName("action_created_at_desc")},
	})
	return err
}

func (s *AdminAuditStore) WriteAdminAuditLog(ctx context.Context, item core.AdminAuditLog) error {
	if s == nil || s.collection == nil {
		return nil
	}
	writeCtx, cancel := withTimeout(ctx, s.writeTimeout)
	defer cancel()
	_, err := s.collection.InsertOne(writeCtx, bson.M{
		"operator":        item.Operator,
		"action":          item.Action,
		"target_type":     item.TargetType,
		"target_id":       item.TargetID,
		"request_path":    item.RequestPath,
		"request_ip":      item.RequestIP,
		"payload_summary": item.PayloadSummary,
		"result":          item.Result,
		"error_code":      item.ErrorCode,
		"created_at":      item.CreatedAt,
	})
	return err
}
