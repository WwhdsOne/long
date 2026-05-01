package mongostore

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"long/internal/vote"
)

const taskDefinitionsCollectionName = "task_definitions"

type taskDefinitionDocument struct {
	TaskID        string                 `bson:"task_id"`
	Title         string                 `bson:"title"`
	Description   string                 `bson:"description"`
	TaskType      vote.TaskType          `bson:"task_type"`
	EventKind     vote.TaskEventKind     `bson:"event_kind"`
	WindowKind    vote.TaskWindowKind    `bson:"window_kind"`
	Status        vote.TaskStatus        `bson:"status"`
	ConditionKind vote.TaskConditionKind `bson:"condition_kind"`
	TargetValue   int64                  `bson:"target_value"`
	Rewards       vote.TaskRewards       `bson:"rewards"`
	DisplayOrder  int64                  `bson:"display_order"`
	StartAt       int64                  `bson:"start_at"`
	EndAt         int64                  `bson:"end_at"`
	CreatedAt     int64                  `bson:"created_at"`
	UpdatedAt     int64                  `bson:"updated_at"`
}

// TaskDefinitionStore 负责任务定义的读写。
type TaskDefinitionStore struct {
	collection   *mongo.Collection
	writeTimeout time.Duration
	readTimeout  time.Duration
}

func NewTaskDefinitionStore(db *mongo.Database, writeTimeout time.Duration, readTimeout time.Duration) *TaskDefinitionStore {
	return &TaskDefinitionStore{
		collection:   db.Collection(taskDefinitionsCollectionName),
		writeTimeout: writeTimeout,
		readTimeout:  readTimeout,
	}
}

func (s *TaskDefinitionStore) EnsureIndexes(ctx context.Context) error {
	if s == nil || s.collection == nil {
		return nil
	}
	_, err := s.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "task_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("task_id_unique"),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}, {Key: "task_type", Value: 1}, {Key: "display_order", Value: 1}},
			Options: options.Index().SetName("status_task_type_display_order"),
		},
		{
			Keys:    bson.D{{Key: "start_at", Value: 1}, {Key: "end_at", Value: 1}},
			Options: options.Index().SetName("start_end"),
		},
	})
	return err
}

func (s *TaskDefinitionStore) UpsertTaskDefinition(ctx context.Context, item vote.TaskDefinition) error {
	if s == nil || s.collection == nil {
		return nil
	}
	item = vote.NormalizeTaskDefinitionModel(item)
	taskID := strings.TrimSpace(item.TaskID)
	if taskID == "" {
		return nil
	}
	writeCtx, cancel := withTimeout(ctx, s.writeTimeout)
	defer cancel()
	_, err := s.collection.UpdateOne(
		writeCtx,
		bson.M{"task_id": taskID},
		bson.M{"$set": bson.M{
			"task_id":        taskID,
			"title":          item.Title,
			"description":    item.Description,
			"task_type":      item.TaskType,
			"event_kind":     item.EventKind,
			"window_kind":    item.WindowKind,
			"status":         item.Status,
			"condition_kind": item.ConditionKind,
			"target_value":   item.TargetValue,
			"rewards":        item.Rewards,
			"display_order":  item.DisplayOrder,
			"start_at":       item.StartAt,
			"end_at":         item.EndAt,
			"created_at":     item.CreatedAt,
			"updated_at":     item.UpdatedAt,
		}},
		options.Update().SetUpsert(true),
	)
	return err
}

func (s *TaskDefinitionStore) GetTaskDefinition(ctx context.Context, taskID string) (*vote.TaskDefinition, error) {
	if s == nil || s.collection == nil {
		return nil, nil
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, nil
	}
	readCtx, cancel := withTimeout(ctx, s.readTimeout)
	defer cancel()
	var doc taskDefinitionDocument
	err := s.collection.FindOne(readCtx, bson.M{"task_id": taskID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	item := vote.NormalizeTaskDefinitionModel(vote.TaskDefinition{
		TaskID:        doc.TaskID,
		Title:         doc.Title,
		Description:   doc.Description,
		TaskType:      doc.TaskType,
		EventKind:     doc.EventKind,
		WindowKind:    doc.WindowKind,
		Status:        doc.Status,
		ConditionKind: doc.ConditionKind,
		TargetValue:   doc.TargetValue,
		Rewards:       doc.Rewards,
		DisplayOrder:  doc.DisplayOrder,
		StartAt:       doc.StartAt,
		EndAt:         doc.EndAt,
		CreatedAt:     doc.CreatedAt,
		UpdatedAt:     doc.UpdatedAt,
	})
	return &item, nil
}

func (s *TaskDefinitionStore) ListActiveTaskDefinitions(ctx context.Context, nowUnix int64) ([]vote.TaskDefinition, error) {
	if s == nil || s.collection == nil {
		return []vote.TaskDefinition{}, nil
	}
	readCtx, cancel := withTimeout(ctx, s.readTimeout)
	defer cancel()
	cursor, err := s.collection.Find(readCtx, bson.M{
		"status": vote.TaskStatusActive,
		"$and": []bson.M{
			{
				"$or": []bson.M{
					{"start_at": bson.M{"$exists": false}},
					{"start_at": 0},
					{"start_at": bson.M{"$lte": nowUnix}},
				},
			},
			{
				"$or": []bson.M{
					{"end_at": bson.M{"$exists": false}},
					{"end_at": 0},
					{"end_at": bson.M{"$gte": nowUnix}},
				},
			},
		},
	}, options.Find().SetSort(bson.D{
		{Key: "display_order", Value: 1},
		{Key: "created_at", Value: 1},
	}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(readCtx)

	items := make([]vote.TaskDefinition, 0)
	for cursor.Next(readCtx) {
		var doc taskDefinitionDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		items = append(items, vote.NormalizeTaskDefinitionModel(vote.TaskDefinition{
			TaskID:        doc.TaskID,
			Title:         doc.Title,
			Description:   doc.Description,
			TaskType:      doc.TaskType,
			EventKind:     doc.EventKind,
			WindowKind:    doc.WindowKind,
			Status:        doc.Status,
			ConditionKind: doc.ConditionKind,
			TargetValue:   doc.TargetValue,
			Rewards:       doc.Rewards,
			DisplayOrder:  doc.DisplayOrder,
			StartAt:       doc.StartAt,
			EndAt:         doc.EndAt,
			CreatedAt:     doc.CreatedAt,
			UpdatedAt:     doc.UpdatedAt,
		}))
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *TaskDefinitionStore) ListTaskDefinitions(ctx context.Context) ([]vote.TaskDefinition, error) {
	if s == nil || s.collection == nil {
		return []vote.TaskDefinition{}, nil
	}
	readCtx, cancel := withTimeout(ctx, s.readTimeout)
	defer cancel()
	cursor, err := s.collection.Find(readCtx, bson.D{}, options.Find().SetSort(bson.D{
		{Key: "display_order", Value: 1},
		{Key: "created_at", Value: 1},
	}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(readCtx)

	items := make([]vote.TaskDefinition, 0)
	for cursor.Next(readCtx) {
		var doc taskDefinitionDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		items = append(items, vote.NormalizeTaskDefinitionModel(vote.TaskDefinition{
			TaskID:        doc.TaskID,
			Title:         doc.Title,
			Description:   doc.Description,
			TaskType:      doc.TaskType,
			EventKind:     doc.EventKind,
			WindowKind:    doc.WindowKind,
			Status:        doc.Status,
			ConditionKind: doc.ConditionKind,
			TargetValue:   doc.TargetValue,
			Rewards:       doc.Rewards,
			DisplayOrder:  doc.DisplayOrder,
			StartAt:       doc.StartAt,
			EndAt:         doc.EndAt,
			CreatedAt:     doc.CreatedAt,
			UpdatedAt:     doc.UpdatedAt,
		}))
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
