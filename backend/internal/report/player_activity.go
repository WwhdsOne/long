package report

import (
	"context"
	"fmt"
	"sort"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// QueryPlayerActivity 查询玩家活跃度统计数据。
func QueryPlayerActivity(ctx context.Context, db *mongo.Database, from, to int64) (PlayerActivityStats, error) {
	var stats PlayerActivityStats

	uniqueIPs, err := countDistinct(ctx, db.Collection("access_logs"), "client_ip", timeFilter("created_at", from, to))
	if err != nil {
		return stats, fmt.Errorf("查询独立IP: %w", err)
	}
	stats.UniqueIPs = uniqueIPs

	totalReqs, err := db.Collection("access_logs").CountDocuments(ctx, timeFilter("created_at", from, to))
	if err != nil {
		return stats, fmt.Errorf("查询总请求量: %w", err)
	}
	stats.TotalRequests = totalReqs

	p95, err := queryP95Latency(ctx, db.Collection("access_logs"), "latency_ms", timeFilter("created_at", from, to))
	if err != nil {
		return stats, fmt.Errorf("查询P95延迟: %w", err)
	}
	stats.P95LatencyMs = p95

	activePlayers, err := countDistinct(ctx, db.Collection("domain_events"), "nickname", timeFilter("created_at", from, to))
	if err != nil {
		return stats, fmt.Errorf("查询活跃玩家: %w", err)
	}
	stats.ActivePlayers = activePlayers

	newPlayers, err := countNewPlayers(ctx, db.Collection("domain_events"), from, to)
	if err != nil {
		return stats, fmt.Errorf("查询新增玩家: %w", err)
	}
	stats.NewPlayers = newPlayers

	topPlayers, err := queryTopPlayers(ctx, db.Collection("domain_events"), from, to, 10)
	if err != nil {
		return stats, fmt.Errorf("查询活跃榜: %w", err)
	}
	stats.TopPlayers = topPlayers

	return stats, nil
}

// countDistinct 返回某字段的去重计数。
func countDistinct(ctx context.Context, coll *mongo.Collection, field string, filter bson.M) (int64, error) {
	values, err := coll.Distinct(ctx, field, filter)
	if err != nil {
		return 0, err
	}
	return int64(len(values)), nil
}

// countNewPlayers 统计区间内首次出现的 nickname 数。
func countNewPlayers(ctx context.Context, coll *mongo.Collection, from, to int64) (int64, error) {
	inRange, err := coll.Distinct(ctx, "nickname", timeFilter("created_at", from, to))
	if err != nil {
		return 0, err
	}
	if len(inRange) == 0 {
		return 0, nil
	}

	beforeRange, err := coll.Distinct(ctx, "nickname", bson.M{"created_at": bson.M{"$lt": from}})
	if err != nil {
		return 0, err
	}

	beforeSet := make(map[string]struct{}, len(beforeRange))
	for _, n := range beforeRange {
		if s, ok := n.(string); ok {
			beforeSet[s] = struct{}{}
		}
	}

	var count int64
	for _, n := range inRange {
		if s, ok := n.(string); ok {
			if _, exists := beforeSet[s]; !exists {
				count++
			}
		}
	}
	return count, nil
}

// queryP95Latency 计算指定字段的 P95 值。
// 因无法保证 MongoDB 版本支持 $percentile，采用采样排序近似计算。
func queryP95Latency(ctx context.Context, coll *mongo.Collection, field string, filter bson.M) (float64, error) {
	cursor, err := coll.Find(ctx, filter, nil)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var latencies []float64
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		if v, ok := getNestedFloat64(doc, field); ok {
			latencies = append(latencies, v)
		}
	}
	if err := cursor.Err(); err != nil {
		return 0, err
	}

	if len(latencies) == 0 {
		return 0, nil
	}

	sort.Float64s(latencies)
	idx := int(float64(len(latencies)) * 0.95)
	if idx >= len(latencies) {
		idx = len(latencies) - 1
	}
	return latencies[idx], nil
}

// queryTopPlayers 查询事件数最多的 N 个玩家。
func queryTopPlayers(ctx context.Context, coll *mongo.Collection, from, to int64, limit int) ([]PlayerActivity, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: timeFilter("created_at", from, to)}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$nickname",
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.M{"count": -1}}},
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []PlayerActivity
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		nickname, _ := doc["_id"].(string)
		count := int64FromDoc(doc, "count")
		if nickname == "" {
			continue
		}
		items = append(items, PlayerActivity{Nickname: nickname, EventCount: count})
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// getNestedFloat64 从 bson.M 中按点号分隔的路径提取 float64。
// 例如 field="rewards.gold" 会先取 doc["rewards"]，再取其 ["gold"]。
func getNestedFloat64(doc bson.M, field string) (float64, bool) {
	val, ok := getNestedValue(doc, field)
	if !ok {
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		return v, true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

// getNestedValue 从 bson.M 中按点号分隔的路径提取值。
func getNestedValue(doc bson.M, field string) (any, bool) {
	// 简单字段，直接取
	if v, ok := doc[field]; ok {
		return v, true
	}
	// 嵌套字段尝试
	current := any(doc)
	for {
		dot := -1
		for i := 0; i < len(field); i++ {
			if field[i] == '.' {
				dot = i
				break
			}
		}
		if dot < 0 {
			if m, ok := current.(bson.M); ok {
				v, exists := m[field]
				return v, exists
			}
			return nil, false
		}
		key := field[:dot]
		field = field[dot+1:]
		if m, ok := current.(bson.M); ok {
			current = m[key]
		} else {
			return nil, false
		}
	}
}

// int64FromDoc 安全地从 bson.M 中提取 int64。
func int64FromDoc(doc bson.M, key string) int64 {
	switch v := doc[key].(type) {
	case int64:
		return v
	case int32:
		return int64(v)
	case float64:
		return int64(v)
	default:
		return 0
	}
}
