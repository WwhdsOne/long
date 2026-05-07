package mongostore

import (
	"context"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"long/internal/core"
)

const bossHistoryCollectionName = "boss_history"

type bossHistoryDocument struct {
	BossID             string                      `bson:"boss_id"`
	TemplateID         string                      `bson:"template_id,omitempty"`
	RoomID             string                      `bson:"room_id,omitempty"`
	QueueID            string                      `bson:"queue_id,omitempty"`
	Name               string                      `bson:"name"`
	Status             string                      `bson:"status"`
	MaxHP              int64                       `bson:"max_hp"`
	CurrentHP          int64                       `bson:"current_hp"`
	GoldOnKill         int64                       `bson:"gold_on_kill"`
	StoneOnKill        int64                       `bson:"stone_on_kill"`
	TalentPointsOnKill int64                       `bson:"talent_points_on_kill"`
	Parts              []core.BossPart             `bson:"parts,omitempty"`
	StartedAt          int64                       `bson:"started_at"`
	DefeatedAt         int64                       `bson:"defeated_at,omitempty"`
	Loot               []core.BossLootEntry        `bson:"loot,omitempty"`
	Damage             []core.BossLeaderboardEntry `bson:"damage,omitempty"`
	ArchivedAt         int64                       `bson:"archived_at"`
}

// BossHistoryStore 负责 MongoDB 中的 Boss 历史归档与分页查询。
type BossHistoryStore struct {
	collection   *mongo.Collection
	writeTimeout time.Duration
	readTimeout  time.Duration
}

// NewBossHistoryStore 创建 Boss 历史 Mongo 仓储。
func NewBossHistoryStore(db *mongo.Database, writeTimeout time.Duration, readTimeout time.Duration) *BossHistoryStore {
	return &BossHistoryStore{
		collection:   db.Collection(bossHistoryCollectionName),
		writeTimeout: writeTimeout,
		readTimeout:  readTimeout,
	}
}

// EnsureIndexes 初始化 Boss 历史索引。
func (s *BossHistoryStore) EnsureIndexes(ctx context.Context) error {
	if s == nil || s.collection == nil {
		return nil
	}

	_, err := s.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "boss_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("boss_id_unique"),
		},
		{
			Keys:    bson.D{{Key: "started_at", Value: -1}},
			Options: options.Index().SetName("started_at_desc"),
		},
		{
			Keys:    bson.D{{Key: "template_id", Value: 1}, {Key: "started_at", Value: -1}},
			Options: options.Index().SetName("template_started_at_desc"),
		},
		{
			Keys:    bson.D{{Key: "room_id", Value: 1}, {Key: "started_at", Value: -1}},
			Options: options.Index().SetName("room_started_at_desc"),
		},
		{
			Keys:    bson.D{{Key: "queue_id", Value: 1}, {Key: "started_at", Value: -1}},
			Options: options.Index().SetName("queue_started_at_desc"),
		},
		{
			Keys:    bson.D{{Key: "room_id", Value: 1}, {Key: "queue_id", Value: 1}, {Key: "started_at", Value: -1}},
			Options: options.Index().SetName("room_queue_started_at_desc"),
		},
	})
	return err
}

// SaveBossHistory 写入或覆盖一条 Boss 历史。
func (s *BossHistoryStore) SaveBossHistory(ctx context.Context, entry core.BossHistoryEntry) error {
	if s == nil || s.collection == nil {
		return nil
	}
	if entry.ID == "" {
		return nil
	}

	writeCtx, cancel := withTimeout(ctx, s.writeTimeout)
	defer cancel()

	doc := bossHistoryDocument{
		BossID:             entry.ID,
		TemplateID:         entry.TemplateID,
		RoomID:             entry.RoomID,
		QueueID:            entry.QueueID,
		Name:               entry.Name,
		Status:             entry.Status,
		MaxHP:              entry.MaxHP,
		CurrentHP:          entry.CurrentHP,
		GoldOnKill:         entry.GoldOnKill,
		StoneOnKill:        entry.StoneOnKill,
		TalentPointsOnKill: entry.TalentPointsOnKill,
		Parts:              entry.Parts,
		StartedAt:          entry.StartedAt,
		DefeatedAt:         entry.DefeatedAt,
		Loot:               entry.Loot,
		Damage:             entry.Damage,
		ArchivedAt:         time.Now().Unix(),
	}

	_, err := s.collection.UpdateOne(
		writeCtx,
		bson.M{"boss_id": entry.ID},
		bson.M{"$set": doc},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

// ListAdminBossHistoryPage 返回后台 Boss 历史分页。
func (s *BossHistoryStore) ListAdminBossHistoryPage(ctx context.Context, page int64, pageSize int64) (core.AdminBossHistoryPage, error) {
	if s == nil || s.collection == nil {
		return core.AdminBossHistoryPage{}, nil
	}

	page, pageSize = normalizePage(page, pageSize)
	readCtx, cancel := withTimeout(ctx, s.readTimeout)
	defer cancel()

	total, err := s.collection.CountDocuments(readCtx, bson.D{})
	if err != nil {
		return core.AdminBossHistoryPage{}, err
	}

	findOptions := options.Find().
		SetSort(bson.D{{Key: "started_at", Value: -1}}).
		SetSkip((page - 1) * pageSize).
		SetLimit(pageSize)

	cursor, err := s.collection.Find(readCtx, bson.D{}, findOptions)
	if err != nil {
		return core.AdminBossHistoryPage{}, err
	}
	defer cursor.Close(readCtx)

	items := make([]core.BossHistoryEntry, 0)
	for cursor.Next(readCtx) {
		var doc bossHistoryDocument
		if err := cursor.Decode(&doc); err != nil {
			return core.AdminBossHistoryPage{}, err
		}
		items = append(items, core.BossHistoryEntry{
			Boss: core.Boss{
				ID:                 doc.BossID,
				TemplateID:         doc.TemplateID,
				RoomID:             doc.RoomID,
				QueueID:            doc.QueueID,
				Name:               doc.Name,
				Status:             doc.Status,
				MaxHP:              doc.MaxHP,
				CurrentHP:          doc.CurrentHP,
				GoldOnKill:         doc.GoldOnKill,
				StoneOnKill:        doc.StoneOnKill,
				TalentPointsOnKill: doc.TalentPointsOnKill,
				Parts:              doc.Parts,
				StartedAt:          doc.StartedAt,
				DefeatedAt:         doc.DefeatedAt,
			},
			Loot:   doc.Loot,
			Damage: doc.Damage,
		})
	}
	if err := cursor.Err(); err != nil {
		return core.AdminBossHistoryPage{}, err
	}

	totalPages := int64(0)
	if total > 0 {
		totalPages = int64(math.Ceil(float64(total) / float64(pageSize)))
	}

	return core.AdminBossHistoryPage{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// ListBossHistory 返回全部 Boss 历史倒序列表。
func (s *BossHistoryStore) ListBossHistory(ctx context.Context) ([]core.BossHistoryEntry, error) {
	return collectAllBossHistoryPages(func(page int64, pageSize int64) (core.AdminBossHistoryPage, error) {
		return s.ListAdminBossHistoryPage(ctx, page, pageSize)
	}, 100)
}

func normalizePage(page int64, pageSize int64) (int64, int64) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func collectAllBossHistoryPages(fetch func(page int64, pageSize int64) (core.AdminBossHistoryPage, error), pageSize int64) ([]core.BossHistoryEntry, error) {
	if pageSize <= 0 {
		pageSize = 100
	}

	items := make([]core.BossHistoryEntry, 0)
	for page := int64(1); ; page++ {
		result, err := fetch(page, pageSize)
		if err != nil {
			return nil, err
		}
		items = append(items, result.Items...)
		if result.TotalPages > 0 && page >= result.TotalPages {
			break
		}
		if len(result.Items) == 0 {
			break
		}
	}
	return items, nil
}

func withTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, timeout)
}
