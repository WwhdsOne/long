package vote

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestDebugWwhdsUserState(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "47.93.83.136:6379",
		Password: "Wwh852456",
		DB:       2,
	})
	defer func() {
		_ = client.Close()
	}()

	store := NewStore(client, "vote:button:", StoreOptions{}, nil)
	ctx := context.Background()
	nickname := "Wwhds"

	if _, err := store.gemsForNickname(ctx, nickname); err != nil {
		t.Fatalf("gemsForNickname: %v", err)
	}
	if _, _, err := store.ownedCosmeticsForNickname(ctx, nickname); err != nil {
		t.Fatalf("ownedCosmeticsForNickname: %v", err)
	}
	if _, err := store.cosmeticLoadoutForNickname(ctx, nickname); err != nil {
		t.Fatalf("cosmeticLoadoutForNickname: %v", err)
	}
	if _, err := store.lastForgeResultForNickname(ctx, nickname); err != nil {
		t.Fatalf("lastForgeResultForNickname: %v", err)
	}
	if _, err := store.GetUserStats(ctx, nickname); err != nil {
		t.Fatalf("GetUserStats: %v", err)
	}
	quantities, err := store.inventoryQuantities(ctx, nickname)
	if err != nil {
		t.Fatalf("inventoryQuantities: %v", err)
	}
	loadout, equipped, err := store.loadoutForNickname(ctx, nickname, quantities)
	if err != nil {
		t.Fatalf("loadoutForNickname: %v", err)
	}
	if _, err := store.inventoryForNickname(ctx, nickname, quantities, equipped); err != nil {
		t.Fatalf("inventoryForNickname: %v", err)
	}
	heroQuantities, err := store.heroInventoryQuantities(ctx, nickname)
	if err != nil {
		t.Fatalf("heroInventoryQuantities: %v", err)
	}
	activeHero, err := store.activeHeroForNickname(ctx, nickname, heroQuantities)
	if err != nil {
		t.Fatalf("activeHeroForNickname: %v", err)
	}
	if _, err := store.heroInventoryForNickname(ctx, nickname, heroQuantities, activeHero); err != nil {
		t.Fatalf("heroInventoryForNickname: %v", err)
	}
	if _, err := store.combatStatsForNickname(ctx, nickname, loadout, activeHero); err != nil {
		t.Fatalf("combatStatsForNickname: %v", err)
	}
	if _, err := store.recentRewardsForNickname(ctx, nickname); err != nil {
		t.Fatalf("recentRewardsForNickname: %v", err)
	}
	boss, err := store.currentBoss(ctx)
	if err != nil {
		t.Fatalf("currentBoss: %v", err)
	}
	if boss != nil {
		if _, err := store.bossStatsForNickname(ctx, boss.ID, nickname); err != nil {
			t.Fatalf("bossStatsForNickname: %v", err)
		}
	}
}
