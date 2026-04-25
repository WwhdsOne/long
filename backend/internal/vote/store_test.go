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

	return NewStore(client, "vote:", StoreOptions{
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
		{ItemID: "cloth-armor", DropRatePercent: 25},
		{ItemID: "fire-ring", DropRatePercent: 75},
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
	if resources.BossLoot[0].ItemID != "cloth-armor" || resources.BossLoot[0].DropRatePercent != 25 {
		t.Fatalf("expected cloth-armor probability 25%%, got %+v", resources.BossLoot)
	}
	if resources.BossLoot[1].ItemID != "fire-ring" || resources.BossLoot[1].DropRatePercent != 75 {
		t.Fatalf("expected fire-ring probability 75%%, got %+v", resources.BossLoot)
	}
}

func TestBossLootDropRateIsIndependentProbability(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	boss := &Boss{
		ID:        "raid-1",
		Name:      "裂隙领主",
		Status:    bossStatusActive,
		MaxHP:     100,
		CurrentHP: 100,
		StartedAt: store.now().Unix(),
	}
	if err := store.setCurrentBoss(ctx, boss, []BossLootEntry{
		{ItemID: "mythic-core", DropRatePercent: 10},
		{ItemID: "raid-token", DropRatePercent: 10},
	}); err != nil {
		t.Fatalf("set current boss: %v", err)
	}

	resources, err := store.GetBossResources(ctx)
	if err != nil {
		t.Fatalf("get boss resources: %v", err)
	}
	if len(resources.BossLoot) != 2 {
		t.Fatalf("expected boss loot resources, got %+v", resources.BossLoot)
	}
	if resources.BossLoot[0].DropRatePercent+resources.BossLoot[1].DropRatePercent != 20 {
		t.Fatalf("expected independent drop rates to keep configured values, got %+v", resources.BossLoot)
	}
}

func TestRollLootDropsCanReturnMultipleItems(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	drops := store.rollLootDrops([]BossLootEntry{
		{ItemID: "mythic-core", DropRatePercent: 100},
		{ItemID: "raid-token", DropRatePercent: 100},
	})

	if len(drops) != 2 || drops[0].ItemID != "mythic-core" || drops[1].ItemID != "raid-token" {
		t.Fatalf("expected multiple independent drops, got %+v", drops)
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

	if result.Delta != 5 {
		t.Fatalf("expected click delta to stay 5 after removing starlight, got %+v", result)
	}
	if result.Button.Count != 5 || result.UserStats.ClickCount != 5 {
		t.Fatalf("expected single delta to apply to counts, got %+v", result)
	}
}

func TestActivateBossWithPartsUsesPartHealthAsTotalHealth(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	boss, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "part-boss",
		Name:  "分区 Boss",
		MaxHP: 9999,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 120, CurrentHP: 120, Alive: true},
			{X: 1, Y: 0, Type: PartTypeHeavy, MaxHP: 80, CurrentHP: 60, Alive: true},
			{X: 2, Y: 0, Type: PartTypeWeak, MaxHP: 50, CurrentHP: 0, Alive: false},
		},
	})
	if err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	if boss.MaxHP != 250 || boss.CurrentHP != 250 {
		t.Fatalf("expected boss health to match parts max/current sums, got max=%d current=%d parts=%+v", boss.MaxHP, boss.CurrentHP, boss.Parts)
	}
	if boss.Parts[1].CurrentHP != 80 || !boss.Parts[2].Alive {
		t.Fatalf("expected activated boss parts to be reset to full health, got %+v", boss.Parts)
	}

	stored, err := store.currentBoss(ctx)
	if err != nil {
		t.Fatalf("load current boss: %v", err)
	}
	if stored.MaxHP != 250 || stored.CurrentHP != 250 {
		t.Fatalf("expected stored boss health to match parts max/current sums, got max=%d current=%d parts=%+v", stored.MaxHP, stored.CurrentHP, stored.Parts)
	}
}

func TestBossTemplateActivationUsesPartHealthAsTotalHealth(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.SaveBossTemplate(ctx, BossTemplateUpsert{
		ID:    "template-part-boss",
		Name:  "模板分区 Boss",
		MaxHP: 9999,
		Layout: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 300, CurrentHP: 10, Alive: false},
			{X: 1, Y: 0, Type: PartTypeHeavy, MaxHP: 200, CurrentHP: 20, Alive: false},
		},
	}); err != nil {
		t.Fatalf("save boss template: %v", err)
	}

	templates, err := store.ListBossTemplates(ctx)
	if err != nil {
		t.Fatalf("list boss templates: %v", err)
	}
	if len(templates) != 1 || templates[0].MaxHP != 500 {
		t.Fatalf("expected saved template max health to match layout, got %+v", templates)
	}

	boss, err := store.activateRandomBossFromPool(ctx)
	if err != nil {
		t.Fatalf("activate boss from pool: %v", err)
	}
	if boss.MaxHP != 500 || boss.CurrentHP != 500 {
		t.Fatalf("expected activated template boss health to match layout, got max=%d current=%d parts=%+v", boss.MaxHP, boss.CurrentHP, boss.Parts)
	}
	for _, part := range boss.Parts {
		if part.CurrentHP != part.MaxHP || !part.Alive {
			t.Fatalf("expected activated template parts to be reset to full health, got %+v", boss.Parts)
		}
	}
}

func TestBossPartDisplayFieldsPersist(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.SaveBossTemplate(ctx, BossTemplateUpsert{
		ID:    "display-part-boss",
		Name:  "展示字段 Boss",
		MaxHP: 100,
		Layout: []BossPart{
			{
				X:           0,
				Y:           0,
				Type:        PartTypeWeak,
				MaxHP:       100,
				CurrentHP:   100,
				Alive:       true,
				DisplayName: "眼核",
				ImagePath:   "/assets/boss/eye.png",
			},
		},
	}); err != nil {
		t.Fatalf("save boss template: %v", err)
	}

	templates, err := store.ListBossTemplates(ctx)
	if err != nil {
		t.Fatalf("list boss templates: %v", err)
	}
	if len(templates) != 1 || len(templates[0].Layout) != 1 {
		t.Fatalf("expected one template part, got %+v", templates)
	}
	part := templates[0].Layout[0]
	if part.DisplayName != "眼核" || part.ImagePath != "/assets/boss/eye.png" {
		t.Fatalf("expected template display fields to persist, got %+v", part)
	}

	boss, err := store.activateRandomBossFromPool(ctx)
	if err != nil {
		t.Fatalf("activate boss from pool: %v", err)
	}
	if len(boss.Parts) != 1 || boss.Parts[0].DisplayName != "眼核" || boss.Parts[0].ImagePath != "/assets/boss/eye.png" {
		t.Fatalf("expected activated boss display fields to persist, got %+v", boss.Parts)
	}
}

func TestClickButtonWithBossPartsPersistsBossAndPartHealth(t *testing.T) {
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
	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "part-boss",
		Name:  "分区 Boss",
		MaxHP: 100,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 100, CurrentHP: 100, Alive: true},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	result, err := store.ClickButton(ctx, "feel", "阿明")
	if err != nil {
		t.Fatalf("click button: %v", err)
	}
	if result.Boss == nil {
		t.Fatal("expected click result to include boss state")
	}
	if result.Boss.CurrentHP != 95 || len(result.Boss.Parts) != 1 || result.Boss.Parts[0].CurrentHP != 95 {
		t.Fatalf("expected click result to reduce boss and part health, got %+v", result.Boss)
	}

	stored, err := store.currentBoss(ctx)
	if err != nil {
		t.Fatalf("load current boss: %v", err)
	}
	if stored.CurrentHP != 95 || len(stored.Parts) != 1 || stored.Parts[0].CurrentHP != 95 {
		t.Fatalf("expected stored boss and part health to be reduced, got %+v", stored)
	}
}

func TestClickButtonWithBossPartsPersistsDefeatedStatus(t *testing.T) {
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
	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "fragile-boss",
		Name:  "脆弱 Boss",
		MaxHP: 1,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 1, CurrentHP: 1, Alive: true},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	result, err := store.ClickButton(ctx, "feel", "阿明")
	if err != nil {
		t.Fatalf("click button: %v", err)
	}
	if result.Boss == nil || result.Boss.Status != bossStatusDefeated || result.Boss.CurrentHP != 0 {
		t.Fatalf("expected click result to defeat boss, got %+v", result.Boss)
	}

	stored, err := store.currentBoss(ctx)
	if err != nil {
		t.Fatalf("load current boss: %v", err)
	}
	if stored.Status != bossStatusDefeated || stored.CurrentHP != 0 || stored.DefeatedAt == 0 {
		t.Fatalf("expected stored boss to be defeated, got %+v", stored)
	}
}

func TestManualBossPartClickCountsOneButDamageUsesCombatFormula(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.critical.CriticalChancePercent = 0
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
	seedEquipmentDefinition(t, store, ctx, "strong-sword", "weapon", 7)
	if err := store.client.HSet(ctx, store.inventoryKey("阿明"), "strong-sword", "1").Err(); err != nil {
		t.Fatalf("seed inventory: %v", err)
	}
	if _, err := store.EquipItem(ctx, "阿明", "strong-sword"); err != nil {
		t.Fatalf("equip item: %v", err)
	}
	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "formula-boss",
		Name:  "公式 Boss",
		MaxHP: 100,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 100, CurrentHP: 100, Alive: true},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	result, err := store.ClickButton(ctx, "feel", "阿明")
	if err != nil {
		t.Fatalf("click button: %v", err)
	}
	if result.Delta != 1 || result.BossDamage != 12 {
		t.Fatalf("expected click delta 1 and boss damage 12, got delta=%d bossDamage=%d result=%+v", result.Delta, result.BossDamage, result)
	}
	if result.UserStats.ClickCount != 1 {
		t.Fatalf("expected manual click count to increase by 1, got %+v", result.UserStats)
	}
	if result.Boss == nil || result.Boss.CurrentHP != 88 || result.Boss.Parts[0].CurrentHP != 88 {
		t.Fatalf("expected boss health to lose 12 damage, got %+v", result.Boss)
	}
}

func TestClickBossPartWithoutButtonTargetsSelectedPart(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.critical.CriticalChancePercent = 0
	store.now = func() time.Time {
		return time.Unix(1713744000, 0)
	}

	ctx := context.Background()
	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "direct-part-boss",
		Name:  "直点 Boss",
		MaxHP: 200,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, DisplayName: "左翼", MaxHP: 100, CurrentHP: 100, Alive: true},
			{X: 1, Y: 0, Type: PartTypeWeak, DisplayName: "眼核", MaxHP: 100, CurrentHP: 100, Alive: true},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	result, err := store.ClickButton(ctx, "boss-part:1-0", "阿明")
	if err != nil {
		t.Fatalf("click boss part: %v", err)
	}
	if result.Button.Key != "boss-part:1-0" || result.Button.Label != "眼核" {
		t.Fatalf("expected pseudo part target button, got %+v", result.Button)
	}
	if result.Delta != 1 || result.UserStats.ClickCount != 1 {
		t.Fatalf("expected direct part click to count once, got %+v", result)
	}
	if result.Boss == nil || result.Boss.CurrentHP != 188 {
		t.Fatalf("expected boss health to decrease by selected part damage, got %+v", result.Boss)
	}
	if result.Boss.Parts[0].CurrentHP != 100 || result.Boss.Parts[1].CurrentHP != 88 {
		t.Fatalf("expected only selected part to lose HP, got %+v", result.Boss.Parts)
	}

	snapshot, err := store.GetSnapshot(ctx)
	if err != nil {
		t.Fatalf("get snapshot: %v", err)
	}
	if len(snapshot.Buttons) != 0 || snapshot.TotalVotes != 1 {
		t.Fatalf("expected no button records and one total click, got buttons=%+v total=%d", snapshot.Buttons, snapshot.TotalVotes)
	}
}

func TestBossAutoClickDoesNotIncreaseUserClicks(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.critical.CriticalChancePercent = 0
	store.now = func() time.Time {
		return time.Unix(1713744000, 0)
	}

	ctx := context.Background()
	seedEquipmentDefinition(t, store, ctx, "strong-sword", "weapon", 7)
	if err := store.client.HSet(ctx, store.inventoryKey("阿明"), "strong-sword", "1").Err(); err != nil {
		t.Fatalf("seed inventory: %v", err)
	}
	if _, err := store.EquipItem(ctx, "阿明", "strong-sword"); err != nil {
		t.Fatalf("equip item: %v", err)
	}
	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "auto-boss",
		Name:  "挂机 Boss",
		MaxHP: 100,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 100, CurrentHP: 100, Alive: true},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	result, err := store.AutoClickBossPart(ctx, "0-0", "阿明")
	if err != nil {
		t.Fatalf("auto click boss part: %v", err)
	}
	if result.Delta != 0 || result.BossDamage != 12 {
		t.Fatalf("expected auto click delta 0 and boss damage 12, got delta=%d bossDamage=%d result=%+v", result.Delta, result.BossDamage, result)
	}

	userStats, err := store.GetUserStats(ctx, "阿明")
	if err != nil {
		t.Fatalf("get user stats: %v", err)
	}
	if userStats.ClickCount != 0 {
		t.Fatalf("expected auto click not to increase click count, got %+v", userStats)
	}
	if result.Boss == nil || result.Boss.CurrentHP != 88 || result.Boss.Parts[0].CurrentHP != 88 {
		t.Fatalf("expected auto click to damage boss, got %+v", result.Boss)
	}
}
