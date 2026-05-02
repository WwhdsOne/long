package mongostore

import (
	"context"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"long/internal/core"
)

const (
	messageCollectionName = "wall_messages"
	counterCollectionName = "counters"
	messageCounterID      = "wall_messages"
)

type messageDocument struct {
	MessageID string `bson:"message_id"`
	Seq       int64  `bson:"seq"`
	Nickname  string `bson:"nickname"`
	Content   string `bson:"content"`
	CreatedAt int64  `bson:"created_at"`
	Status    string `bson:"status"`
}

type counterDocument struct {
	ID    string `bson:"_id"`
	Value int64  `bson:"value"`
}

// MessageStore 负责 MongoDB 下的留言墙主存。
type MessageStore struct {
	messages     *mongo.Collection
	counters     *mongo.Collection
	writeTimeout time.Duration
	readTimeout  time.Duration
}

// NewMessageStore 创建留言墙 Mongo 仓储。
func NewMessageStore(db *mongo.Database, writeTimeout time.Duration, readTimeout time.Duration) *MessageStore {
	return &MessageStore{
		messages:     db.Collection(messageCollectionName),
		counters:     db.Collection(counterCollectionName),
		writeTimeout: writeTimeout,
		readTimeout:  readTimeout,
	}
}

// EnsureIndexes 初始化索引和计数器。
func (s *MessageStore) EnsureIndexes(ctx context.Context) error {
	if s == nil || s.messages == nil {
		return nil
	}

	if _, err := s.messages.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "message_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("message_id_unique"),
		},
		{
			Keys:    bson.D{{Key: "seq", Value: -1}},
			Options: options.Index().SetName("seq_desc"),
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("created_at_desc"),
		},
	}); err != nil {
		return err
	}

	return s.ensureCounterFloor(ctx)
}

// CreateMessage 创建一条留言。
func (s *MessageStore) CreateMessage(ctx context.Context, nickname string, content string) (*core.Message, error) {
	writeCtx, cancel := withTimeout(ctx, s.writeTimeout)
	defer cancel()

	seq, err := s.nextSequence(writeCtx)
	if err != nil {
		return nil, err
	}

	item := &core.Message{
		ID:        strconv.FormatInt(seq, 10),
		Nickname:  strings.TrimSpace(nickname),
		Content:   strings.TrimSpace(content),
		CreatedAt: time.Now().Unix(),
	}
	doc := messageDocument{
		MessageID: item.ID,
		Seq:       seq,
		Nickname:  item.Nickname,
		Content:   item.Content,
		CreatedAt: item.CreatedAt,
		Status:    "active",
	}

	if _, err := s.messages.InsertOne(writeCtx, doc); err != nil {
		return nil, err
	}
	return item, nil
}

// ListMessages 返回留言分页。
func (s *MessageStore) ListMessages(ctx context.Context, cursor string, limit int64) (core.MessagePage, error) {
	if limit <= 0 {
		limit = 50
	}

	filter := bson.M{"status": bson.M{"$ne": "deleted"}}
	if trimmed := strings.TrimSpace(cursor); trimmed != "" {
		seq, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return core.MessagePage{}, err
		}
		filter["seq"] = bson.M{"$lt": seq}
	}

	readCtx, cancel := withTimeout(ctx, s.readTimeout)
	defer cancel()

	findOptions := options.Find().
		SetSort(bson.D{{Key: "seq", Value: -1}}).
		SetLimit(limit + 1)

	cursorResult, err := s.messages.Find(readCtx, filter, findOptions)
	if err != nil {
		return core.MessagePage{}, err
	}
	defer cursorResult.Close(readCtx)

	page := core.MessagePage{
		Items: make([]core.Message, 0, limit),
	}
	for cursorResult.Next(readCtx) {
		var doc messageDocument
		if err := cursorResult.Decode(&doc); err != nil {
			return core.MessagePage{}, err
		}

		if int64(len(page.Items)) >= limit {
			page.NextCursor = page.Items[len(page.Items)-1].ID
			break
		}

		page.Items = append(page.Items, core.Message{
			ID:        doc.MessageID,
			Nickname:  doc.Nickname,
			Content:   doc.Content,
			CreatedAt: doc.CreatedAt,
		})
	}
	if err := cursorResult.Err(); err != nil {
		return core.MessagePage{}, err
	}
	return page, nil
}

// DeleteMessage 删除一条留言。
func (s *MessageStore) DeleteMessage(ctx context.Context, id string) error {
	writeCtx, cancel := withTimeout(ctx, s.writeTimeout)
	defer cancel()

	_, err := s.messages.UpdateOne(
		writeCtx,
		bson.M{"message_id": strings.TrimSpace(id)},
		bson.M{"$set": bson.M{"status": "deleted"}},
	)
	return err
}

// UpsertMessage 用于迁移和补数据。
func (s *MessageStore) UpsertMessage(ctx context.Context, item core.Message) error {
	writeCtx, cancel := withTimeout(ctx, s.writeTimeout)
	defer cancel()

	seq, err := strconv.ParseInt(strings.TrimSpace(item.ID), 10, 64)
	if err != nil {
		return err
	}

	_, err = s.messages.UpdateOne(
		writeCtx,
		bson.M{"message_id": item.ID},
		bson.M{"$set": messageDocument{
			MessageID: item.ID,
			Seq:       seq,
			Nickname:  item.Nickname,
			Content:   item.Content,
			CreatedAt: item.CreatedAt,
			Status:    "active",
		}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return err
	}

	_, err = s.counters.UpdateOne(
		writeCtx,
		bson.M{"_id": messageCounterID},
		bson.M{"$max": bson.M{"value": seq}},
		options.Update().SetUpsert(true),
	)
	return err
}

func (s *MessageStore) nextSequence(ctx context.Context) (int64, error) {
	var result counterDocument
	err := s.counters.FindOneAndUpdate(
		ctx,
		bson.M{"_id": messageCounterID},
		bson.M{"$inc": bson.M{"value": 1}},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&result)
	return result.Value, err
}

func (s *MessageStore) ensureCounterFloor(ctx context.Context) error {
	var doc messageDocument
	err := s.messages.FindOne(ctx, bson.D{}, options.FindOne().SetSort(bson.D{{Key: "seq", Value: -1}})).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		_, err = s.counters.UpdateOne(
			ctx,
			bson.M{"_id": messageCounterID},
			bson.M{"$setOnInsert": bson.M{"value": int64(0)}},
			options.Update().SetUpsert(true),
		)
		return err
	}
	if err != nil {
		return err
	}

	_, err = s.counters.UpdateOne(
		ctx,
		bson.M{"_id": messageCounterID},
		bson.M{"$max": bson.M{"value": doc.Seq}},
		options.Update().SetUpsert(true),
	)
	return err
}
