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

const shopItemsCollectionName = "shop_items"

type ShopItemStore struct {
	collection   *mongo.Collection
	writeTimeout time.Duration
	readTimeout  time.Duration
}

func NewShopItemStore(db *mongo.Database, writeTimeout time.Duration, readTimeout time.Duration) *ShopItemStore {
	return &ShopItemStore{
		collection:   db.Collection(shopItemsCollectionName),
		writeTimeout: writeTimeout,
		readTimeout:  readTimeout,
	}
}

func (s *ShopItemStore) EnsureIndexes(ctx context.Context) error {
	if s == nil || s.collection == nil {
		return nil
	}
	_, err := s.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "item_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("item_id_unique"),
		},
		{
			Keys:    bson.D{{Key: "active", Value: 1}, {Key: "sort_order", Value: 1}, {Key: "created_at", Value: 1}},
			Options: options.Index().SetName("active_sort_created"),
		},
	})
	return err
}

func (s *ShopItemStore) ListActiveShopItems(ctx context.Context) ([]core.ShopItem, error) {
	return s.listByFilter(ctx, bson.M{"active": true})
}

func (s *ShopItemStore) ListShopItems(ctx context.Context) ([]core.ShopItem, error) {
	return s.listByFilter(ctx, bson.D{})
}

func (s *ShopItemStore) listByFilter(ctx context.Context, filter any) ([]core.ShopItem, error) {
	if s == nil || s.collection == nil {
		return []core.ShopItem{}, nil
	}
	readCtx, cancel := withTimeout(ctx, s.readTimeout)
	defer cancel()
	cursor, err := s.collection.Find(readCtx, filter, options.Find().SetSort(bson.D{
		{Key: "sort_order", Value: 1},
		{Key: "created_at", Value: 1},
	}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(readCtx)

	items := make([]core.ShopItem, 0)
	for cursor.Next(readCtx) {
		var item core.ShopItem
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		items = append(items, core.NormalizeShopItemModel(item))
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *ShopItemStore) GetShopItem(ctx context.Context, itemID string) (*core.ShopItem, error) {
	if s == nil || s.collection == nil {
		return nil, nil
	}
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return nil, nil
	}
	readCtx, cancel := withTimeout(ctx, s.readTimeout)
	defer cancel()
	var item core.ShopItem
	err := s.collection.FindOne(readCtx, bson.M{"item_id": itemID}).Decode(&item)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	item = core.NormalizeShopItemModel(item)
	return &item, nil
}

func (s *ShopItemStore) UpsertShopItem(ctx context.Context, item core.ShopItem) error {
	if s == nil || s.collection == nil {
		return nil
	}
	item = core.NormalizeShopItemModel(item)
	if item.ItemID == "" {
		return nil
	}
	writeCtx, cancel := withTimeout(ctx, s.writeTimeout)
	defer cancel()
	_, err := s.collection.UpdateOne(
		writeCtx,
		bson.M{"item_id": item.ItemID},
		bson.M{"$set": item},
		options.Update().SetUpsert(true),
	)
	return err
}

func (s *ShopItemStore) DeleteShopItem(ctx context.Context, itemID string) error {
	if s == nil || s.collection == nil {
		return nil
	}
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return nil
	}
	writeCtx, cancel := withTimeout(ctx, s.writeTimeout)
	defer cancel()
	_, err := s.collection.DeleteOne(writeCtx, bson.M{"item_id": itemID})
	return err
}
