package mongostore

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"long/internal/core"
)

const taskClaimLogsCollectionName = "task_claim_logs"

// TaskClaimLogStore 负责任务领取日志写入。
type TaskClaimLogStore struct {
	collection   *mongo.Collection
	writeTimeout time.Duration
	readTimeout  time.Duration
}

func NewTaskClaimLogStore(db *mongo.Database, writeTimeout time.Duration, readTimeout time.Duration) *TaskClaimLogStore {
	return &TaskClaimLogStore{
		collection:   db.Collection(taskClaimLogsCollectionName),
		writeTimeout: writeTimeout,
		readTimeout:  readTimeout,
	}
}

func (s *TaskClaimLogStore) EnsureIndexes(ctx context.Context) error {
	if s == nil || s.collection == nil {
		return nil
	}
	_, err := s.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "nickname", Value: 1}, {Key: "task_id", Value: 1}, {Key: "cycle_key", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("nickname_task_cycle_unique"),
		},
		{
			Keys:    bson.D{{Key: "claimed_at", Value: -1}},
			Options: options.Index().SetName("claimed_at_desc"),
		},
	})
	return err
}

func (s *TaskClaimLogStore) WriteTaskClaimLog(ctx context.Context, item core.TaskClaimLog) error {
	if s == nil || s.collection == nil {
		return nil
	}
	taskID := strings.TrimSpace(item.TaskID)
	cycleKey := strings.TrimSpace(item.CycleKey)
	nickname := strings.TrimSpace(item.Nickname)
	if taskID == "" || cycleKey == "" || nickname == "" {
		return nil
	}
	writeCtx, cancel := withTimeout(ctx, s.writeTimeout)
	defer cancel()
	_, err := s.collection.UpdateOne(
		writeCtx,
		bson.M{"task_id": taskID, "cycle_key": cycleKey, "nickname": nickname},
		bson.M{"$set": bson.M{
			"task_id":     taskID,
			"cycle_key":   cycleKey,
			"nickname":    nickname,
			"rewards":     item.Rewards,
			"claimed_at":  item.ClaimedAt,
			"archived_at": item.ArchivedAt,
		}},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func (s *TaskClaimLogStore) HasTaskClaimed(ctx context.Context, taskID string, cycleKey string, nickname string) (bool, error) {
	if s == nil || s.collection == nil {
		return false, nil
	}
	readCtx, cancel := withTimeout(ctx, s.readTimeout)
	defer cancel()
	count, err := s.collection.CountDocuments(readCtx, bson.M{
		"task_id":   strings.TrimSpace(taskID),
		"cycle_key": strings.TrimSpace(cycleKey),
		"nickname":  strings.TrimSpace(nickname),
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *TaskClaimLogStore) ListTaskClaimLogs(ctx context.Context, taskID string, cycleKey string) ([]core.TaskClaimLog, error) {
	if s == nil || s.collection == nil {
		return []core.TaskClaimLog{}, nil
	}
	readCtx, cancel := withTimeout(ctx, s.readTimeout)
	defer cancel()
	cursor, err := s.collection.Find(readCtx, bson.M{
		"task_id":   strings.TrimSpace(taskID),
		"cycle_key": strings.TrimSpace(cycleKey),
	}, options.Find().SetSort(bson.D{{Key: "claimed_at", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(readCtx)
	items := make([]core.TaskClaimLog, 0)
	for cursor.Next(readCtx) {
		var doc struct {
			TaskID     string           `bson:"task_id"`
			CycleKey   string           `bson:"cycle_key"`
			Nickname   string           `bson:"nickname"`
			Rewards    core.TaskRewards `bson:"rewards"`
			ClaimedAt  int64            `bson:"claimed_at"`
			ArchivedAt int64            `bson:"archived_at"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		items = append(items, core.TaskClaimLog{
			TaskID:     doc.TaskID,
			CycleKey:   doc.CycleKey,
			Nickname:   doc.Nickname,
			Rewards:    doc.Rewards,
			ClaimedAt:  doc.ClaimedAt,
			ArchivedAt: doc.ArchivedAt,
		})
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *TaskClaimLogStore) HasTaskClaimLog(ctx context.Context, taskID string, cycleKey string) (bool, error) {
	if s == nil || s.collection == nil {
		return false, nil
	}
	readCtx, cancel := withTimeout(ctx, s.readTimeout)
	defer cancel()
	count, err := s.collection.CountDocuments(readCtx, bson.M{
		"task_id":   strings.TrimSpace(taskID),
		"cycle_key": strings.TrimSpace(cycleKey),
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
