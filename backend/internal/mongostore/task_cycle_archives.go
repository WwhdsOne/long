package mongostore

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"long/internal/core"
)

const (
	taskCycleArchivesCollectionName      = "task_cycle_archives"
	taskCyclePlayerResultsCollectionName = "task_cycle_player_results"
)

type taskCycleArchiveDocument struct {
	TaskID                string                 `bson:"task_id"`
	CycleKey              string                 `bson:"cycle_key"`
	TaskType              core.TaskType          `bson:"task_type"`
	EventKind             core.TaskEventKind     `bson:"event_kind"`
	WindowKind            core.TaskWindowKind    `bson:"window_kind"`
	ConditionKind         core.TaskConditionKind `bson:"condition_kind"`
	TargetValue           int64                  `bson:"target_value"`
	StartAt               int64                  `bson:"start_at"`
	EndAt                 int64                  `bson:"end_at"`
	ParticipantsTotal     int64                  `bson:"participants_total"`
	CompletedTotal        int64                  `bson:"completed_total"`
	ClaimedTotal          int64                  `bson:"claimed_total"`
	ExpiredUnclaimedTotal int64                  `bson:"expired_unclaimed_total"`
	UnfinishedTotal       int64                  `bson:"unfinished_total"`
	NotParticipatedTotal  int64                  `bson:"not_participated_total"`
	ArchivedAt            int64                  `bson:"archived_at"`
}

// TaskCycleArchiveStore 负责任务周期归档。
type TaskCycleArchiveStore struct {
	archives      *mongo.Collection
	playerResults *mongo.Collection
	writeTimeout  time.Duration
}

func NewTaskCycleArchiveStore(db *mongo.Database, writeTimeout time.Duration) *TaskCycleArchiveStore {
	return &TaskCycleArchiveStore{
		archives:      db.Collection(taskCycleArchivesCollectionName),
		playerResults: db.Collection(taskCyclePlayerResultsCollectionName),
		writeTimeout:  writeTimeout,
	}
}

func (s *TaskCycleArchiveStore) EnsureIndexes(ctx context.Context) error {
	if s == nil || s.archives == nil || s.playerResults == nil {
		return nil
	}
	if _, err := s.archives.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "task_id", Value: 1}, {Key: "cycle_key", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("task_cycle_unique"),
		},
		{
			Keys:    bson.D{{Key: "archived_at", Value: -1}},
			Options: options.Index().SetName("archived_at_desc"),
		},
	}); err != nil {
		return err
	}
	_, err := s.playerResults.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "task_id", Value: 1}, {Key: "cycle_key", Value: 1}, {Key: "nickname", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("task_cycle_nickname_unique"),
		},
		{
			Keys:    bson.D{{Key: "task_id", Value: 1}, {Key: "cycle_key", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("task_cycle_status"),
		},
	})
	return err
}

func (s *TaskCycleArchiveStore) UpsertTaskCycleArchive(ctx context.Context, item core.TaskCycleArchive) error {
	if s == nil || s.archives == nil {
		return nil
	}
	item = core.NormalizeTaskArchiveModel(item)
	taskID := strings.TrimSpace(item.TaskID)
	cycleKey := strings.TrimSpace(item.CycleKey)
	if taskID == "" || cycleKey == "" {
		return nil
	}
	writeCtx, cancel := withTimeout(ctx, s.writeTimeout)
	defer cancel()
	_, err := s.archives.UpdateOne(
		writeCtx,
		bson.M{"task_id": taskID, "cycle_key": cycleKey},
		bson.M{"$set": bson.M{
			"task_id":                 taskID,
			"cycle_key":               cycleKey,
			"task_type":               item.TaskType,
			"event_kind":              item.EventKind,
			"window_kind":             item.WindowKind,
			"condition_kind":          item.ConditionKind,
			"target_value":            item.TargetValue,
			"start_at":                item.StartAt,
			"end_at":                  item.EndAt,
			"participants_total":      item.ParticipantsTotal,
			"completed_total":         item.CompletedTotal,
			"claimed_total":           item.ClaimedTotal,
			"expired_unclaimed_total": item.ExpiredUnclaimedTotal,
			"unfinished_total":        item.UnfinishedTotal,
			"not_participated_total":  item.NotParticipatedTotal,
			"archived_at":             item.ArchivedAt,
		}},
		options.Update().SetUpsert(true),
	)
	return err
}

func (s *TaskCycleArchiveStore) UpsertTaskCyclePlayerResults(ctx context.Context, items []core.TaskCyclePlayerResult) error {
	if s == nil || s.playerResults == nil || len(items) == 0 {
		return nil
	}
	writeCtx, cancel := withTimeout(ctx, s.writeTimeout)
	defer cancel()
	models := make([]mongo.WriteModel, 0, len(items))
	for _, item := range items {
		taskID := strings.TrimSpace(item.TaskID)
		cycleKey := strings.TrimSpace(item.CycleKey)
		nickname := strings.TrimSpace(item.Nickname)
		if taskID == "" || cycleKey == "" || nickname == "" {
			continue
		}
		models = append(models, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"task_id": taskID, "cycle_key": cycleKey, "nickname": nickname}).
			SetUpdate(bson.M{"$set": bson.M{
				"task_id":      taskID,
				"cycle_key":    cycleKey,
				"nickname":     nickname,
				"progress":     item.Progress,
				"target_value": item.TargetValue,
				"status":       item.Status,
				"completed_at": item.CompletedAt,
				"claimed_at":   item.ClaimedAt,
				"archived_at":  item.ArchivedAt,
			}}).
			SetUpsert(true))
	}
	if len(models) == 0 {
		return nil
	}
	_, err := s.playerResults.BulkWrite(writeCtx, models, options.BulkWrite().SetOrdered(false))
	return err
}

func (s *TaskCycleArchiveStore) ListTaskCycleArchives(ctx context.Context, taskID string) ([]core.TaskCycleArchive, error) {
	if s == nil || s.archives == nil {
		return []core.TaskCycleArchive{}, nil
	}
	readCtx, cancel := withTimeout(ctx, s.writeTimeout)
	defer cancel()
	cursor, err := s.archives.Find(readCtx, bson.M{
		"task_id": strings.TrimSpace(taskID),
	}, options.Find().SetSort(bson.D{{Key: "archived_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(readCtx)
	items := make([]core.TaskCycleArchive, 0)
	for cursor.Next(readCtx) {
		var doc taskCycleArchiveDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		items = append(items, core.NormalizeTaskArchiveModel(core.TaskCycleArchive{
			TaskID:                doc.TaskID,
			CycleKey:              doc.CycleKey,
			TaskType:              doc.TaskType,
			EventKind:             doc.EventKind,
			WindowKind:            doc.WindowKind,
			ConditionKind:         doc.ConditionKind,
			TargetValue:           doc.TargetValue,
			StartAt:               doc.StartAt,
			EndAt:                 doc.EndAt,
			ParticipantsTotal:     doc.ParticipantsTotal,
			CompletedTotal:        doc.CompletedTotal,
			ClaimedTotal:          doc.ClaimedTotal,
			ExpiredUnclaimedTotal: doc.ExpiredUnclaimedTotal,
			UnfinishedTotal:       doc.UnfinishedTotal,
			NotParticipatedTotal:  doc.NotParticipatedTotal,
			ArchivedAt:            doc.ArchivedAt,
		}))
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *TaskCycleArchiveStore) GetTaskCycleResults(ctx context.Context, taskID string, cycleKey string) (core.TaskCycleResultsView, error) {
	if s == nil || s.archives == nil || s.playerResults == nil {
		return core.TaskCycleResultsView{}, nil
	}
	readCtx, cancel := withTimeout(ctx, s.writeTimeout)
	defer cancel()
	var archiveDoc taskCycleArchiveDocument
	if err := s.archives.FindOne(readCtx, bson.M{
		"task_id":   strings.TrimSpace(taskID),
		"cycle_key": strings.TrimSpace(cycleKey),
	}).Decode(&archiveDoc); err != nil {
		if err == mongo.ErrNoDocuments {
			return core.TaskCycleResultsView{}, nil
		}
		return core.TaskCycleResultsView{}, err
	}
	cursor, err := s.playerResults.Find(readCtx, bson.M{
		"task_id":   strings.TrimSpace(taskID),
		"cycle_key": strings.TrimSpace(cycleKey),
	}, options.Find().SetSort(bson.D{
		{Key: "status", Value: 1},
		{Key: "nickname", Value: 1},
	}))
	if err != nil {
		return core.TaskCycleResultsView{}, err
	}
	defer cursor.Close(readCtx)
	items := make([]core.TaskCyclePlayerResult, 0)
	for cursor.Next(readCtx) {
		var doc struct {
			TaskID      string                `bson:"task_id"`
			CycleKey    string                `bson:"cycle_key"`
			Nickname    string                `bson:"nickname"`
			Progress    int64                 `bson:"progress"`
			TargetValue int64                 `bson:"target_value"`
			Status      core.TaskPlayerStatus `bson:"status"`
			CompletedAt int64                 `bson:"completed_at"`
			ClaimedAt   int64                 `bson:"claimed_at"`
			ArchivedAt  int64                 `bson:"archived_at"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return core.TaskCycleResultsView{}, err
		}
		items = append(items, core.TaskCyclePlayerResult{
			TaskID:      doc.TaskID,
			CycleKey:    doc.CycleKey,
			Nickname:    doc.Nickname,
			Progress:    doc.Progress,
			TargetValue: doc.TargetValue,
			Status:      doc.Status,
			CompletedAt: doc.CompletedAt,
			ClaimedAt:   doc.ClaimedAt,
			ArchivedAt:  doc.ArchivedAt,
		})
	}
	if err := cursor.Err(); err != nil {
		return core.TaskCycleResultsView{}, err
	}
	return core.TaskCycleResultsView{
		Archive: core.NormalizeTaskArchiveModel(core.TaskCycleArchive{
			TaskID:                archiveDoc.TaskID,
			CycleKey:              archiveDoc.CycleKey,
			TaskType:              archiveDoc.TaskType,
			EventKind:             archiveDoc.EventKind,
			WindowKind:            archiveDoc.WindowKind,
			ConditionKind:         archiveDoc.ConditionKind,
			TargetValue:           archiveDoc.TargetValue,
			StartAt:               archiveDoc.StartAt,
			EndAt:                 archiveDoc.EndAt,
			ParticipantsTotal:     archiveDoc.ParticipantsTotal,
			CompletedTotal:        archiveDoc.CompletedTotal,
			ClaimedTotal:          archiveDoc.ClaimedTotal,
			ExpiredUnclaimedTotal: archiveDoc.ExpiredUnclaimedTotal,
			UnfinishedTotal:       archiveDoc.UnfinishedTotal,
			NotParticipatedTotal:  archiveDoc.NotParticipatedTotal,
			ArchivedAt:            archiveDoc.ArchivedAt,
		}),
		Items: items,
	}, nil
}
