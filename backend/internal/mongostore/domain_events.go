package mongostore

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"long/internal/vote"
)

const domainEventsCollectionName = "domain_events"

// DomainEventStore 负责业务事件写入。
type DomainEventStore struct {
	collection   *mongo.Collection
	writeTimeout time.Duration
}

func NewDomainEventStore(db *mongo.Database, writeTimeout time.Duration) *DomainEventStore {
	return &DomainEventStore{
		collection:   db.Collection(domainEventsCollectionName),
		writeTimeout: writeTimeout,
	}
}

func (s *DomainEventStore) EnsureIndexes(ctx context.Context) error {
	if s == nil || s.collection == nil {
		return nil
	}
	_, err := s.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "created_at", Value: -1}}, Options: options.Index().SetName("created_at_desc")},
		{Keys: bson.D{{Key: "event_type", Value: 1}, {Key: "created_at", Value: -1}}, Options: options.Index().SetName("event_type_created_at_desc")},
		{Keys: bson.D{{Key: "boss_id", Value: 1}, {Key: "created_at", Value: -1}}, Options: options.Index().SetName("boss_id_created_at_desc")},
		{Keys: bson.D{{Key: "nickname", Value: 1}, {Key: "created_at", Value: -1}}, Options: options.Index().SetName("nickname_created_at_desc")},
	})
	return err
}

func (s *DomainEventStore) WriteDomainEvent(ctx context.Context, item vote.DomainEvent) error {
	if s == nil || s.collection == nil {
		return nil
	}
	writeCtx, cancel := withTimeout(ctx, s.writeTimeout)
	defer cancel()
	_, err := s.collection.InsertOne(writeCtx, bson.M{
		"event_type": item.EventType,
		"nickname":   item.Nickname,
		"boss_id":    item.BossID,
		"item_id":    item.ItemID,
		"payload":    item.Payload,
		"created_at": item.CreatedAt,
	})
	return err
}
