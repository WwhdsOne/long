package report

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// QueryBossStats 查询 Boss 战况统计数据。
func QueryBossStats(ctx context.Context, db *mongo.Database, from, to int64) (BossStats, error) {
	var stats BossStats

	spawnCount, err := db.Collection("boss_history").CountDocuments(ctx, timeFilter("started_at", from, to))
	if err != nil {
		return stats, fmt.Errorf("查询Boss生成次数: %w", err)
	}
	stats.SpawnCount = spawnCount

	killCount, err := db.Collection("boss_history").CountDocuments(ctx, bson.M{
		"status":      "defeated",
		"defeated_at": bson.M{"$gte": from, "$lt": to},
	})
	if err != nil {
		return stats, fmt.Errorf("查询Boss击杀次数: %w", err)
	}
	stats.KillCount = killCount

	if spawnCount > 0 {
		stats.KillRate = float64(killCount) / float64(spawnCount) * 100
	}

	totalDamage, err := queryTotalDamage(ctx, db, from, to)
	if err != nil {
		return stats, fmt.Errorf("查询总伤害: %w", err)
	}
	stats.TotalDamage = totalDamage

	avgSurvival, err := queryAvgSurvival(ctx, db, from, to)
	if err != nil {
		return stats, fmt.Errorf("查询平均存活: %w", err)
	}
	stats.AvgSurvivalSecs = avgSurvival

	topDamagers, err := queryTopDamagers(ctx, db, from, to, 10)
	if err != nil {
		return stats, fmt.Errorf("查询伤害榜: %w", err)
	}
	stats.TopDamagers = topDamagers

	topLoot, err := queryTopLoot(ctx, db, from, to, 5)
	if err != nil {
		return stats, fmt.Errorf("查询掉落榜: %w", err)
	}
	stats.TopLoot = topLoot

	return stats, nil
}

func queryTotalDamage(ctx context.Context, db *mongo.Database, from, to int64) (int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: timeFilter("started_at", from, to)}},
		{{Key: "$unwind", Value: "$damage"}},
		{{Key: "$group", Value: bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$damage.damage"},
		}}},
	}

	cursor, err := db.Collection("boss_history").Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return 0, err
		}
		return int64FromDoc(doc, "total"), nil
	}
	return 0, nil
}

func queryAvgSurvival(ctx context.Context, db *mongo.Database, from, to int64) (float64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"status":      "defeated",
			"defeated_at": bson.M{"$gte": from, "$lt": to},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": nil,
			"avg": bson.M{"$avg": bson.M{"$subtract": bson.A{"$defeated_at", "$started_at"}}},
		}}},
	}

	cursor, err := db.Collection("boss_history").Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return 0, err
		}
		if v, ok := doc["avg"].(float64); ok {
			return v, nil
		}
	}
	return 0, nil
}

func queryTopDamagers(ctx context.Context, db *mongo.Database, from, to int64, limit int) ([]DamageRankItem, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: timeFilter("started_at", from, to)}},
		{{Key: "$unwind", Value: "$damage"}},
		{{Key: "$group", Value: bson.M{
			"_id":         "$damage.nickname",
			"totalDamage": bson.M{"$sum": "$damage.damage"},
		}}},
		{{Key: "$sort", Value: bson.M{"totalDamage": -1}}},
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := db.Collection("boss_history").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []DamageRankItem
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		nickname, _ := doc["_id"].(string)
		damage := int64FromDoc(doc, "totalDamage")
		if nickname == "" {
			continue
		}
		items = append(items, DamageRankItem{Nickname: nickname, TotalDamage: damage})
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func queryTopLoot(ctx context.Context, db *mongo.Database, from, to int64, limit int) ([]LootRankItem, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: timeFilter("started_at", from, to)}},
		{{Key: "$unwind", Value: "$loot"}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$loot.itemName",
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.M{"count": -1}}},
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := db.Collection("boss_history").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []LootRankItem
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		itemName, _ := doc["_id"].(string)
		count := int64FromDoc(doc, "count")
		if itemName == "" {
			continue
		}
		items = append(items, LootRankItem{ItemName: itemName, Count: count})
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
