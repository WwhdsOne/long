package core

import (
	"context"
	"strings"

	"github.com/redis/go-redis/v9"
)

// ListShopItems 返回后台商店商品列表。
func (s *Store) ListShopItems(ctx context.Context) ([]ShopItem, error) {
	if s.shopCatalogStore == nil {
		return []ShopItem{}, nil
	}
	items, err := s.shopCatalogStore.ListShopItems(ctx)
	if err != nil {
		return nil, err
	}
	for index := range items {
		items[index] = NormalizeShopItemModel(items[index])
	}
	return items, nil
}

// SaveShopItem 创建或更新一个商店商品。
func (s *Store) SaveShopItem(ctx context.Context, item ShopItem) error {
	if s.shopCatalogStore == nil {
		return ErrShopItemNotFound
	}
	item = NormalizeShopItemModel(item)
	if item.ItemID == "" || item.Title == "" {
		return ErrShopItemNotPurchasable
	}
	if item.ItemType != ShopItemTypeBattleClickSkin {
		return ErrShopUnsupportedItemType
	}
	nowUnix := s.now().Unix()
	existing, err := s.shopCatalogStore.GetShopItem(ctx, item.ItemID)
	if err != nil {
		return err
	}
	if existing != nil {
		item.CreatedAt = existing.CreatedAt
	} else if item.CreatedAt == 0 {
		item.CreatedAt = nowUnix
	}
	item.UpdatedAt = nowUnix
	return s.shopCatalogStore.UpsertShopItem(ctx, item)
}

// DeleteShopItem 删除一个商店商品。
func (s *Store) DeleteShopItem(ctx context.Context, itemID string) error {
	if s.shopCatalogStore == nil {
		return nil
	}
	return s.shopCatalogStore.DeleteShopItem(ctx, strings.TrimSpace(itemID))
}

// ListShopCatalogItemsForPlayer 返回前台商店目录。
func (s *Store) ListShopCatalogItemsForPlayer(ctx context.Context, nickname string) ([]ShopCatalogItemView, error) {
	if s.shopCatalogStore == nil {
		return []ShopCatalogItemView{}, nil
	}
	items, err := s.shopCatalogStore.ListActiveShopItems(ctx)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return []ShopCatalogItemView{}, nil
	}

	ownedSet := map[string]struct{}{}
	equippedID := ""
	if normalizedNickname, ok := normalizeNickname(nickname); ok {
		if normalizedNickname, err = s.validatedNickname(normalizedNickname); err == nil {
			ownedIDs, ownedErr := s.client.SMembers(ctx, s.ownedBattleClickSkinsKey(normalizedNickname)).Result()
			if ownedErr != nil {
				return nil, ownedErr
			}
			for _, itemID := range ownedIDs {
				ownedSet[itemID] = struct{}{}
			}
			equippedID, err = s.client.Get(ctx, s.equippedBattleClickSkinKey(normalizedNickname)).Result()
			if err != nil && err != redis.Nil {
				return nil, err
			}
		}
	}

	result := make([]ShopCatalogItemView, 0, len(items))
	for _, item := range items {
		item = NormalizeShopItemModel(item)
		_, owned := ownedSet[item.ItemID]
		result = append(result, ShopCatalogItemView{
			ShopItem: item,
			Owned:    owned,
			Equipped: item.ItemID == equippedID,
		})
	}
	return result, nil
}

// PurchaseShopItem 购买一个商店商品并返回最新个人态。
func (s *Store) PurchaseShopItem(ctx context.Context, nickname string, itemID string) (UserState, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return UserState{}, err
	}
	item, err := s.getPurchasableShopItem(ctx, itemID)
	if err != nil {
		return UserState{}, err
	}

	owned, err := s.client.SIsMember(ctx, s.ownedBattleClickSkinsKey(normalizedNickname), item.ItemID).Result()
	if err != nil {
		return UserState{}, err
	}
	if owned {
		return UserState{}, ErrShopItemAlreadyOwned
	}

	resources, err := s.resourcesForNickname(ctx, normalizedNickname)
	if err != nil {
		return UserState{}, err
	}
	if resources.Gold < item.PriceGold {
		return UserState{}, ErrShopInsufficientGold
	}

	nowUnix := s.now().Unix()
	pipe := s.client.TxPipeline()
	if item.PriceGold > 0 {
		pipe.HIncrBy(ctx, s.resourceKey(normalizedNickname), "gold", -item.PriceGold)
	}
	pipe.SAdd(ctx, s.ownedBattleClickSkinsKey(normalizedNickname), item.ItemID)
	if item.AutoEquipOnPurchase {
		pipe.Set(ctx, s.equippedBattleClickSkinKey(normalizedNickname), item.ItemID, 0)
	}
	pipe.ZAdd(ctx, s.playerIndexKey, redis.Z{
		Score:  float64(nowUnix),
		Member: normalizedNickname,
	})
	if _, err := pipe.Exec(ctx); err != nil {
		return UserState{}, err
	}

	if s.shopPurchaseLogStore != nil {
		if err := s.shopPurchaseLogStore.WriteShopPurchaseLog(ctx, ShopPurchaseLog{
			ItemID:      item.ItemID,
			Nickname:    normalizedNickname,
			ItemType:    item.ItemType,
			PriceGold:   item.PriceGold,
			PurchasedAt: nowUnix,
			Equipped:    item.AutoEquipOnPurchase,
		}); err != nil {
			return UserState{}, err
		}
	}

	return s.GetUserState(ctx, normalizedNickname)
}

// UnequipShopItem 卸下当前点击图标，恢复默认外观。
func (s *Store) UnequipShopItem(ctx context.Context, nickname string) (UserState, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return UserState{}, err
	}

	nowUnix := s.now().Unix()
	pipe := s.client.TxPipeline()
	pipe.Del(ctx, s.equippedBattleClickSkinKey(normalizedNickname))
	pipe.ZAdd(ctx, s.playerIndexKey, redis.Z{
		Score:  float64(nowUnix),
		Member: normalizedNickname,
	})
	if _, err := pipe.Exec(ctx); err != nil {
		return UserState{}, err
	}
	return s.GetUserState(ctx, normalizedNickname)
}

// EquipShopItem 切换当前点击图标外观。
func (s *Store) EquipShopItem(ctx context.Context, nickname string, itemID string) (UserState, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return UserState{}, err
	}
	item, err := s.getPurchasableShopItem(ctx, itemID)
	if err != nil {
		return UserState{}, err
	}

	owned, err := s.client.SIsMember(ctx, s.ownedBattleClickSkinsKey(normalizedNickname), item.ItemID).Result()
	if err != nil {
		return UserState{}, err
	}
	if !owned {
		return UserState{}, ErrShopItemNotOwned
	}

	nowUnix := s.now().Unix()
	pipe := s.client.TxPipeline()
	pipe.Set(ctx, s.equippedBattleClickSkinKey(normalizedNickname), item.ItemID, 0)
	pipe.ZAdd(ctx, s.playerIndexKey, redis.Z{
		Score:  float64(nowUnix),
		Member: normalizedNickname,
	})
	if _, err := pipe.Exec(ctx); err != nil {
		return UserState{}, err
	}
	return s.GetUserState(ctx, normalizedNickname)
}

func (s *Store) getPurchasableShopItem(ctx context.Context, itemID string) (ShopItem, error) {
	if s.shopCatalogStore == nil {
		return ShopItem{}, ErrShopItemNotFound
	}
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return ShopItem{}, ErrShopItemNotFound
	}
	item, err := s.shopCatalogStore.GetShopItem(ctx, itemID)
	if err != nil {
		return ShopItem{}, err
	}
	if item == nil {
		return ShopItem{}, ErrShopItemNotFound
	}
	*item = NormalizeShopItemModel(*item)
	if !item.Active {
		return ShopItem{}, ErrShopItemNotPurchasable
	}
	if item.ItemType != ShopItemTypeBattleClickSkin {
		return ShopItem{}, ErrShopUnsupportedItemType
	}
	return *item, nil
}

func (s *Store) equippedBattleClickSkinKey(nickname string) string {
	return s.equippedBattleClickSkinPrefix + nickname
}

func (s *Store) ownedBattleClickSkinsKey(nickname string) string {
	return s.ownedBattleClickSkinsPrefix + nickname
}

func (s *Store) equippedBattleClickSkinState(ctx context.Context, nickname string) (string, string, error) {
	if s.shopCatalogStore == nil || strings.TrimSpace(nickname) == "" {
		return "", "", nil
	}
	itemID, err := s.client.Get(ctx, s.equippedBattleClickSkinKey(nickname)).Result()
	if err != nil {
		if err == redis.Nil {
			return "", "", nil
		}
		return "", "", err
	}
	item, err := s.shopCatalogStore.GetShopItem(ctx, itemID)
	if err != nil {
		return "", "", err
	}
	if item == nil {
		return itemID, "", nil
	}
	*item = NormalizeShopItemModel(*item)
	if !item.Active || item.ItemType != ShopItemTypeBattleClickSkin {
		return itemID, "", nil
	}
	return itemID, item.BattleClickCursorImagePath, nil
}
