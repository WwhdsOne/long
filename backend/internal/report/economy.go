package report

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// QueryEconomyStats 查询经济系统统计数据。
func QueryEconomyStats(ctx context.Context, db *mongo.Database, from, to int64) (EconomyStats, error) {
	var stats EconomyStats

	shopGold, shopCnt, err := queryShopSales(ctx, db, from, to)
	if err != nil {
		return stats, fmt.Errorf("查询商店销售: %w", err)
	}
	stats.ShopTotalGold = shopGold
	stats.ShopPurchaseCnt = shopCnt

	topShop, err := queryTopShopItems(ctx, db, from, to, 10)
	if err != nil {
		return stats, fmt.Errorf("查询热销榜: %w", err)
	}
	stats.TopShopItems = topShop

	taskClaimCnt, err := db.Collection("task_claim_logs").CountDocuments(ctx, timeFilter("claimed_at", from, to))
	if err != nil {
		return stats, fmt.Errorf("查询任务完成数: %w", err)
	}
	stats.TaskClaimCnt = taskClaimCnt

	taskRewardGold, taskRewardStones, taskRewardTP, err := queryTaskRewardSums(ctx, db, from, to)
	if err != nil {
		return stats, fmt.Errorf("查询任务奖励汇总: %w", err)
	}
	stats.TaskRewardGold = taskRewardGold
	stats.TaskRewardStones = taskRewardStones
	stats.TaskRewardTP = taskRewardTP

	participants, err := countDistinct(ctx, db.Collection("task_claim_logs"), "nickname", timeFilter("claimed_at", from, to))
	if err != nil {
		return stats, fmt.Errorf("查询任务参与人数: %w", err)
	}
	stats.TaskParticipants = participants

	return stats, nil
}

func queryShopSales(ctx context.Context, db *mongo.Database, from, to int64) (int64, int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: timeFilter("purchased_at", from, to)}},
		{{Key: "$group", Value: bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$price_gold"},
			"count": bson.M{"$sum": 1},
		}}},
	}

	cursor, err := db.Collection("shop_purchase_logs").Aggregate(ctx, pipeline)
	if err != nil {
		return 0, 0, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return 0, 0, err
		}
		return int64FromDoc(doc, "total"), int64FromDoc(doc, "count"), nil
	}
	return 0, 0, nil
}

func queryTopShopItems(ctx context.Context, db *mongo.Database, from, to int64, limit int) ([]ShopRankItem, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: timeFilter("purchased_at", from, to)}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$item_id",
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.M{"count": -1}}},
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := db.Collection("shop_purchase_logs").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []ShopRankItem
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		itemID, _ := doc["_id"].(string)
		count := int64FromDoc(doc, "count")
		if itemID == "" {
			continue
		}
		items = append(items, ShopRankItem{ItemID: itemID, Count: count})
	}
	return items, nil
}

func queryTaskRewardSums(ctx context.Context, db *mongo.Database, from, to int64) (int64, int64, int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: timeFilter("claimed_at", from, to)}},
		{{Key: "$group", Value: bson.M{
			"_id":    nil,
			"gold":   bson.M{"$sum": "$rewards.gold"},
			"stones": bson.M{"$sum": "$rewards.stones"},
			"tp":     bson.M{"$sum": "$rewards.talent_points"},
		}}},
	}

	cursor, err := db.Collection("task_claim_logs").Aggregate(ctx, pipeline)
	if err != nil {
		return 0, 0, 0, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return 0, 0, 0, err
		}
		return int64FromDoc(doc, "gold"), int64FromDoc(doc, "stones"), int64FromDoc(doc, "tp"), nil
	}
	return 0, 0, 0, nil
}
