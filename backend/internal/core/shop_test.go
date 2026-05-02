package core

import (
	"context"
	"testing"

	"long/internal/nickname"
)

type stubShopCatalogStore struct {
	items []ShopItem
}

func (s *stubShopCatalogStore) ListActiveShopItems(_ context.Context) ([]ShopItem, error) {
	result := make([]ShopItem, 0, len(s.items))
	for _, item := range s.items {
		if item.Active {
			result = append(result, item)
		}
	}
	return result, nil
}

func (s *stubShopCatalogStore) ListShopItems(_ context.Context) ([]ShopItem, error) {
	return append([]ShopItem(nil), s.items...), nil
}

func (s *stubShopCatalogStore) GetShopItem(_ context.Context, itemID string) (*ShopItem, error) {
	for _, item := range s.items {
		if item.ItemID == itemID {
			copyItem := item
			return &copyItem, nil
		}
	}
	return nil, nil
}

func (s *stubShopCatalogStore) UpsertShopItem(_ context.Context, item ShopItem) error {
	for index := range s.items {
		if s.items[index].ItemID == item.ItemID {
			s.items[index] = item
			return nil
		}
	}
	s.items = append(s.items, item)
	return nil
}

func (s *stubShopCatalogStore) DeleteShopItem(_ context.Context, itemID string) error {
	next := make([]ShopItem, 0, len(s.items))
	for _, item := range s.items {
		if item.ItemID != itemID {
			next = append(next, item)
		}
	}
	s.items = next
	return nil
}

type stubShopPurchaseLogStore struct {
	logs []ShopPurchaseLog
}

func (s *stubShopPurchaseLogStore) WriteShopPurchaseLog(_ context.Context, item ShopPurchaseLog) error {
	s.logs = append(s.logs, item)
	return nil
}

func newShopTestStore(t *testing.T, items []ShopItem) (*Store, *stubShopCatalogStore, *stubShopPurchaseLogStore, func()) {
	t.Helper()

	baseStore, cleanup := newTestStore(t)
	catalogStore := &stubShopCatalogStore{items: items}
	logStore := &stubShopPurchaseLogStore{}
	store := NewStore(baseStore.client, "vote:", StoreOptions{
		CriticalChancePercent: 5,
		ShopCatalogStore:      catalogStore,
		ShopPurchaseLogStore:  logStore,
	}, nickname.NewValidator([]string{"习近平", "xjp"}))
	return store, catalogStore, logStore, cleanup
}

func TestListShopCatalogItemsForPlayerMarksOwnedAndEquipped(t *testing.T) {
	store, _, _, cleanup := newShopTestStore(t, []ShopItem{
		{
			ItemID:                     "skin-basic",
			Title:                      "基础剑光",
			ItemType:                   ShopItemTypeBattleClickSkin,
			PriceGold:                  50,
			BattleClickCursorImagePath: "https://example.com/basic.png",
			PreviewImagePath:           "https://example.com/basic-preview.png",
			Active:                     true,
		},
		{
			ItemID:                     "skin-offline",
			Title:                      "下架皮肤",
			ItemType:                   ShopItemTypeBattleClickSkin,
			PriceGold:                  50,
			BattleClickCursorImagePath: "https://example.com/offline.png",
			Active:                     false,
		},
	})
	defer cleanup()

	ctx := context.Background()
	if err := store.client.SAdd(ctx, store.ownedBattleClickSkinsKey("阿明"), "skin-basic").Err(); err != nil {
		t.Fatalf("seed owned skins: %v", err)
	}
	if err := store.client.Set(ctx, store.equippedBattleClickSkinKey("阿明"), "skin-basic", 0).Err(); err != nil {
		t.Fatalf("seed equipped skin: %v", err)
	}

	items, err := store.ListShopCatalogItemsForPlayer(ctx, "阿明")
	if err != nil {
		t.Fatalf("list shop catalog items: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one active item, got %+v", items)
	}
	if !items[0].Owned || !items[0].Equipped {
		t.Fatalf("expected owned and equipped item, got %+v", items[0])
	}
}

func TestPurchaseShopItemDeductsGoldAutoEquipsAndLogs(t *testing.T) {
	store, _, logStore, cleanup := newShopTestStore(t, []ShopItem{{
		ItemID:                     "skin-basic",
		Title:                      "基础剑光",
		ItemType:                   ShopItemTypeBattleClickSkin,
		PriceGold:                  120,
		BattleClickCursorImagePath: "https://example.com/basic.png",
		Active:                     true,
		AutoEquipOnPurchase:        true,
	}})
	defer cleanup()

	ctx := context.Background()
	if err := store.client.HSet(ctx, store.resourceKey("阿明"), "gold", "200").Err(); err != nil {
		t.Fatalf("seed gold: %v", err)
	}

	state, err := store.PurchaseShopItem(ctx, "阿明", "skin-basic")
	if err != nil {
		t.Fatalf("purchase shop item: %v", err)
	}
	if state.Gold != 80 {
		t.Fatalf("expected gold 80 after purchase, got %d", state.Gold)
	}
	if state.EquippedBattleClickSkinID != "skin-basic" {
		t.Fatalf("expected equipped skin-basic, got %+v", state)
	}
	if state.EquippedBattleClickCursorImagePath != "https://example.com/basic.png" {
		t.Fatalf("expected cursor image path to be returned, got %+v", state)
	}
	owned, err := store.client.SIsMember(ctx, store.ownedBattleClickSkinsKey("阿明"), "skin-basic").Result()
	if err != nil {
		t.Fatalf("check owned skin: %v", err)
	}
	if !owned {
		t.Fatal("expected purchased skin to be owned")
	}
	if len(logStore.logs) != 1 || logStore.logs[0].ItemID != "skin-basic" {
		t.Fatalf("expected one purchase log, got %+v", logStore.logs)
	}
}

func TestPurchaseShopItemRejectsInsufficientGold(t *testing.T) {
	store, _, _, cleanup := newShopTestStore(t, []ShopItem{{
		ItemID:    "skin-basic",
		Title:     "基础剑光",
		ItemType:  ShopItemTypeBattleClickSkin,
		PriceGold: 120,
		Active:    true,
	}})
	defer cleanup()

	ctx := context.Background()
	if err := store.client.HSet(ctx, store.resourceKey("阿明"), "gold", "30").Err(); err != nil {
		t.Fatalf("seed gold: %v", err)
	}

	_, err := store.PurchaseShopItem(ctx, "阿明", "skin-basic")
	if err != ErrShopInsufficientGold {
		t.Fatalf("expected ErrShopInsufficientGold, got %v", err)
	}
}

func TestEquipBattleClickSkinRejectsUnownedItem(t *testing.T) {
	store, _, _, cleanup := newShopTestStore(t, []ShopItem{{
		ItemID:    "skin-basic",
		Title:     "基础剑光",
		ItemType:  ShopItemTypeBattleClickSkin,
		PriceGold: 120,
		Active:    true,
	}})
	defer cleanup()

	_, err := store.EquipShopItem(context.Background(), "阿明", "skin-basic")
	if err != ErrShopItemNotOwned {
		t.Fatalf("expected ErrShopItemNotOwned, got %v", err)
	}
}

func TestGetUserStateIncludesEquippedBattleClickSkinFields(t *testing.T) {
	store, _, _, cleanup := newShopTestStore(t, []ShopItem{{
		ItemID:                     "skin-basic",
		Title:                      "基础剑光",
		ItemType:                   ShopItemTypeBattleClickSkin,
		PriceGold:                  120,
		BattleClickCursorImagePath: "https://example.com/basic.png",
		Active:                     true,
	}})
	defer cleanup()

	ctx := context.Background()
	if err := store.client.Set(ctx, store.equippedBattleClickSkinKey("阿明"), "skin-basic", 0).Err(); err != nil {
		t.Fatalf("seed equipped skin: %v", err)
	}

	state, err := store.GetUserState(ctx, "阿明")
	if err != nil {
		t.Fatalf("get user state: %v", err)
	}
	if state.EquippedBattleClickSkinID != "skin-basic" {
		t.Fatalf("expected equipped skin id, got %+v", state)
	}
	if state.EquippedBattleClickCursorImagePath != "https://example.com/basic.png" {
		t.Fatalf("expected equipped cursor image path, got %+v", state)
	}
}
