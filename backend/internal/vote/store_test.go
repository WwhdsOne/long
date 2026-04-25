package vote

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"long/internal/nickname"
)

func newTestStore(t *testing.T) (*Store, func()) {
	t.Helper()

	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})

	return NewStore(client, "vote:button:", StoreOptions{
			CriticalChancePercent: 5,
			CriticalCount:         5,
		}, nickname.NewValidator([]string{"习近平", "xjp"})), func() {
			_ = client.Close()
			server.Close()
		}
}

func TestListButtonsFiltersDisabledAndSortsBySortThenKey(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:button:feel", map[string]any{
		"label":   "有感觉吗",
		"count":   "3",
		"sort":    "20",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed button: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:button:other", map[string]any{
		"label":   "其他",
		"count":   "5",
		"sort":    "10",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed second button: %v", err)
	}
	boss := &Boss{
		ID:        "dragon-1",
		Name:      "火龙",
		Status:    bossStatusActive,
		MaxHP:     100,
		CurrentHP: 100,
		StartedAt: store.now().Unix(),
	}
	if err := store.setCurrentBoss(ctx, boss, []BossLootEntry{
		{ItemID: "cloth-armor", Weight: 1},
		{ItemID: "fire-ring", Weight: 3},
	}); err != nil {
		t.Fatalf("set current boss: %v", err)
	}

	snapshot, err := store.GetSnapshot(ctx)
	if err != nil {
		t.Fatalf("get snapshot: %v", err)
	}

	if len(snapshot.Buttons) != 2 {
		t.Fatalf("expected full button list in snapshot, got %+v", snapshot.Buttons)
	}

	resources, err := store.GetBossResources(ctx)
	if err != nil {
		t.Fatalf("get boss resources: %v", err)
	}
	if len(resources.BossLoot) != 2 {
		t.Fatalf("expected boss loot resources, got %+v", resources.BossLoot)
	}
	if resources.BossLoot[0].DropRatePercent+resources.BossLoot[1].DropRatePercent != 100 {
		t.Fatalf("expected drop rates to add up to 100, got %+v", resources.BossLoot)
	}
	if resources.BossLoot[0].ItemID != "cloth-armor" || resources.BossLoot[0].DropRatePercent != 25 {
		t.Fatalf("expected cloth-armor probability 25%%, got %+v", resources.BossLoot)
	}
	if resources.BossLoot[1].ItemID != "fire-ring" || resources.BossLoot[1].DropRatePercent != 75 {
		t.Fatalf("expected fire-ring probability 75%%, got %+v", resources.BossLoot)
	}
}
func TestClickButtonDoesNotDoubleDeltaForFormerStarlightButton(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.now = func() time.Time {
		return time.Unix(1713744000, 0)
	}

	ctx := context.Background()
	if err := store.SaveButton(ctx, ButtonUpsert{
		Slug:    "feel",
		Label:   "有感觉吗",
		Sort:    10,
		Enabled: true,
	}); err != nil {
		t.Fatalf("save button: %v", err)
	}

	result, err := store.ClickButton(ctx, "feel", "阿明")
	if err != nil {
		t.Fatalf("click button: %v", err)
	}

	if result.Delta != 1 {
		t.Fatalf("expected click delta to stay 1 after removing starlight, got %+v", result)
	}
	if result.Button.Count != 1 || result.UserStats.ClickCount != 1 {
		t.Fatalf("expected single delta to apply to counts, got %+v", result)
	}
}
