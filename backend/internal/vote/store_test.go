package vote

import (
	"context"
	"errors"
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
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 100, CurrentHP: 100, Alive: true},
		},
		StartedAt: store.now().Unix(),
	}
	if err := store.setCurrentBoss(ctx, boss, []BossLootEntry{
		{ItemID: "cloth-armor", DropRatePercent: 25},
		{ItemID: "fire-ring", DropRatePercent: 75},
	}); err != nil {
		t.Fatalf("set current boss: %v", err)
	}

	_, err := store.GetSnapshot(ctx)
	if err != nil {
		t.Fatalf("get snapshot: %v", err)
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

func TestGetUserStateMigratesLegacyGemKeyToResourceKey(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	legacyKey := store.legacyGemKey(nickname)
	if err := store.client.HSet(ctx, legacyKey, map[string]any{
		"gems":   "12",
		"gold":   "345",
		"stones": "67",
	}).Err(); err != nil {
		t.Fatalf("seed legacy gem key: %v", err)
	}

	state, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state: %v", err)
	}
	if state.Gems != 12 || state.Gold != 345 || state.Stones != 67 {
		t.Fatalf("expected migrated resources from legacy gem key, got gems=%d gold=%d stones=%d", state.Gems, state.Gold, state.Stones)
	}

	newKey := store.resourceKey(nickname)
	values, err := store.client.HMGet(ctx, newKey, "gems", "gold", "stones").Result()
	if err != nil {
		t.Fatalf("read new resource key: %v", err)
	}
	if int64Value(values, 0) != 12 || int64Value(values, 1) != 345 || int64Value(values, 2) != 67 {
		t.Fatalf("expected new resource key to be backfilled, got %+v", values)
	}

	exists, err := store.client.Exists(ctx, legacyKey).Result()
	if err != nil {
		t.Fatalf("check legacy key exists: %v", err)
	}
	if exists != 0 {
		t.Fatalf("expected legacy key deleted after migration, exists=%d", exists)
	}
}

func TestGetUserStateMergesLegacyGemAndResourceKey(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	legacyKey := store.legacyGemKey(nickname)
	if err := store.client.HSet(ctx, legacyKey, map[string]any{
		"gems":   "10",
		"gold":   "100",
		"stones": "40",
	}).Err(); err != nil {
		t.Fatalf("seed legacy gem key: %v", err)
	}

	newKey := store.resourceKey(nickname)
	if err := store.client.HSet(ctx, newKey, map[string]any{
		"gold": "20",
	}).Err(); err != nil {
		t.Fatalf("seed new resource key: %v", err)
	}

	state, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state: %v", err)
	}
	if state.Gems != 10 || state.Gold != 120 || state.Stones != 40 {
		t.Fatalf("expected merged resources from legacy and new key, got gems=%d gold=%d stones=%d", state.Gems, state.Gold, state.Stones)
	}

	values, err := store.client.HMGet(ctx, newKey, "gems", "gold", "stones").Result()
	if err != nil {
		t.Fatalf("read new resource key: %v", err)
	}
	if int64Value(values, 0) != 10 || int64Value(values, 1) != 120 || int64Value(values, 2) != 40 {
		t.Fatalf("expected new resource key to contain merged value, got %+v", values)
	}

	exists, err := store.client.Exists(ctx, legacyKey).Result()
	if err != nil {
		t.Fatalf("check legacy key exists: %v", err)
	}
	if exists != 0 {
		t.Fatalf("expected legacy key deleted after merge migration, exists=%d", exists)
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
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 100, CurrentHP: 100, Alive: true},
		},
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

func TestBossResourcesLootContainsEquipmentIcon(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.client.HSet(ctx, store.equipmentKey("fire-ring"), map[string]any{
		"name":       "烈焰戒指",
		"slot":       "accessory",
		"rarity":     "epic",
		"image_path": "https://cdn.example.com/items/fire-ring.png",
		"image_alt":  "烈焰戒指图标",
	}).Err(); err != nil {
		t.Fatalf("seed equipment definition: %v", err)
	}

	boss := &Boss{
		ID:        "icon-boss",
		Name:      "图标测试 Boss",
		Status:    bossStatusActive,
		MaxHP:     100,
		CurrentHP: 100,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 100, CurrentHP: 100, Alive: true},
		},
		StartedAt: store.now().Unix(),
	}
	if err := store.setCurrentBoss(ctx, boss, []BossLootEntry{
		{ItemID: "fire-ring", DropRatePercent: 30},
	}); err != nil {
		t.Fatalf("set current boss: %v", err)
	}

	resources, err := store.GetBossResources(ctx)
	if err != nil {
		t.Fatalf("get boss resources: %v", err)
	}
	if len(resources.BossLoot) != 1 {
		t.Fatalf("expected 1 loot entry, got %+v", resources.BossLoot)
	}
	if resources.BossLoot[0].ImagePath != "https://cdn.example.com/items/fire-ring.png" {
		t.Fatalf("expected loot image path to be returned, got %+v", resources.BossLoot[0])
	}
	if resources.BossLoot[0].ImageAlt != "烈焰戒指图标" {
		t.Fatalf("expected loot image alt to be returned, got %+v", resources.BossLoot[0])
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

func TestActivateBossRequiresParts(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "no-parts-boss",
		Name:  "无部位 Boss",
		MaxHP: 100,
	}); !errors.Is(err, ErrBossPartsRequired) {
		t.Fatalf("expected ErrBossPartsRequired, got %v", err)
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

func TestSaveBossTemplateRequiresLayout(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.SaveBossTemplate(ctx, BossTemplateUpsert{
		ID:    "no-layout-template",
		Name:  "无部位模板",
		MaxHP: 100,
	}); !errors.Is(err, ErrBossPartsRequired) {
		t.Fatalf("expected ErrBossPartsRequired, got %v", err)
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

func TestBossCycleQueueAdvanceAndWrapOnKill(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	mustSaveBossTemplateForCycleTest(t, store, ctx, "a", "新手木桩")
	mustSaveBossTemplateForCycleTest(t, store, ctx, "b", "史莱姆王")
	mustSaveBossTemplateForCycleTest(t, store, ctx, "c", "骷髅将军")

	if _, err := store.SetBossCycleQueue(ctx, []string{"a", "b", "c"}); err != nil {
		t.Fatalf("set boss cycle queue: %v", err)
	}

	first, err := store.SetBossCycleEnabled(ctx, true)
	if err != nil {
		t.Fatalf("enable boss cycle: %v", err)
	}
	if first == nil || first.TemplateID != "a" {
		t.Fatalf("expected first boss template a, got %+v", first)
	}

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", "阿明")
	if err != nil {
		t.Fatalf("click first boss: %v", err)
	}
	if result.Boss == nil || result.Boss.TemplateID != "b" || result.Boss.Status != bossStatusActive {
		t.Fatalf("expected next boss template b, got %+v", result.Boss)
	}

	result, err = store.ClickBossPart(ctx, "boss-part:0-0", "阿明")
	if err != nil {
		t.Fatalf("click second boss: %v", err)
	}
	if result.Boss == nil || result.Boss.TemplateID != "c" || result.Boss.Status != bossStatusActive {
		t.Fatalf("expected next boss template c, got %+v", result.Boss)
	}

	result, err = store.ClickBossPart(ctx, "boss-part:0-0", "阿明")
	if err != nil {
		t.Fatalf("click third boss: %v", err)
	}
	if result.Boss == nil || result.Boss.TemplateID != "a" || result.Boss.Status != bossStatusActive {
		t.Fatalf("expected wrapped boss template a, got %+v", result.Boss)
	}
}

func TestBossCycleQueueDynamicUpdateAppliesOnCurrentBossDefeat(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	mustSaveBossTemplateForCycleTest(t, store, ctx, "a", "新手木桩")
	mustSaveBossTemplateForCycleTest(t, store, ctx, "b", "史莱姆王")
	mustSaveBossTemplateForCycleTest(t, store, ctx, "c", "骷髅将军")

	if _, err := store.SetBossCycleQueue(ctx, []string{"a", "b", "c"}); err != nil {
		t.Fatalf("set initial boss cycle queue: %v", err)
	}
	if _, err := store.SetBossCycleEnabled(ctx, true); err != nil {
		t.Fatalf("enable boss cycle: %v", err)
	}

	if _, err := store.SetBossCycleQueue(ctx, []string{"c", "b"}); err != nil {
		t.Fatalf("set updated boss cycle queue: %v", err)
	}

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", "阿明")
	if err != nil {
		t.Fatalf("click current boss after queue update: %v", err)
	}
	if result.Boss == nil || result.Boss.TemplateID != "c" {
		t.Fatalf("expected next boss template c after queue update, got %+v", result.Boss)
	}

	result, err = store.ClickBossPart(ctx, "boss-part:0-0", "阿明")
	if err != nil {
		t.Fatalf("click updated queue boss: %v", err)
	}
	if result.Boss == nil || result.Boss.TemplateID != "b" {
		t.Fatalf("expected next boss template b after c, got %+v", result.Boss)
	}
}

func TestEnableBossCycleRequiresConfiguredQueue(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	mustSaveBossTemplateForCycleTest(t, store, ctx, "a", "新手木桩")

	if _, err := store.SetBossCycleEnabled(ctx, true); !errors.Is(err, ErrBossCycleQueueEmpty) {
		t.Fatalf("expected ErrBossCycleQueueEmpty, got %v", err)
	}
}

func TestClickButtonWithBossPartsPersistsBossAndPartHealth(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.now = func() time.Time {
		return time.Unix(1713744000, 0)
	}

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:feel", map[string]any{
		"label":   "有感觉吗",
		"count":   "0",
		"sort":    "10",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed button: %v", err)
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

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", "阿明")
	if err != nil {
		t.Fatalf("click boss part: %v", err)
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

func mustSaveBossTemplateForCycleTest(t *testing.T, store *Store, ctx context.Context, id string, name string) {
	t.Helper()
	if err := store.SaveBossTemplate(ctx, BossTemplateUpsert{
		ID:    id,
		Name:  name,
		MaxHP: 1,
		Layout: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 1, CurrentHP: 1, Alive: true},
		},
	}); err != nil {
		t.Fatalf("save boss template %s: %v", id, err)
	}
}

func TestClickButtonWithBossPartsPersistsDefeatedStatus(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.now = func() time.Time {
		return time.Unix(1713744000, 0)
	}

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:feel", map[string]any{
		"label":   "有感觉吗",
		"count":   "0",
		"sort":    "10",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed button: %v", err)
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

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", "阿明")
	if err != nil {
		t.Fatalf("click boss part: %v", err)
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
	if err := store.client.HSet(ctx, "vote:feel", map[string]any{
		"label":   "有感觉吗",
		"count":   "0",
		"sort":    "10",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed button: %v", err)
	}
	seedEquipmentDefinition(t, store, ctx, "strong-sword", "weapon", 7)
	strongSwordInst := seedOwnedInstance(t, store, ctx, "阿明", "strong-sword")
	if _, err := store.EquipItem(ctx, "阿明", strongSwordInst); err != nil {
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

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", "阿明")
	if err != nil {
		t.Fatalf("click boss part: %v", err)
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
	if snapshot.TotalVotes != 1 {
		t.Fatalf("expected total=1, got %d", snapshot.TotalVotes)
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
	strongSwordInst := seedOwnedInstance(t, store, ctx, "阿明", "strong-sword")
	if _, err := store.EquipItem(ctx, "阿明", strongSwordInst); err != nil {
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
	if result.Delta != 0 || result.BossDamage != 6 {
		t.Fatalf("expected auto click delta 0 and boss damage 6, got delta=%d bossDamage=%d result=%+v", result.Delta, result.BossDamage, result)
	}

	userStats, err := store.GetUserStats(ctx, "阿明")
	if err != nil {
		t.Fatalf("get user stats: %v", err)
	}
	if userStats.ClickCount != 0 {
		t.Fatalf("expected auto click not to increase click count, got %+v", userStats)
	}
	if result.Boss == nil || result.Boss.CurrentHP != 94 || result.Boss.Parts[0].CurrentHP != 94 {
		t.Fatalf("expected auto click to damage boss, got %+v", result.Boss)
	}
}

func TestEquipmentCritRateContributesToCriticalChance(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.critical.CriticalChancePercent = 0
	store.critical.CriticalCount = 5
	store.roll = func(limit int) int {
		return 0
	}

	ctx := context.Background()
	if err := store.SaveEquipmentDefinition(ctx, EquipmentDefinition{
		ItemID:   "crit-ring",
		Name:     "暴击戒指",
		Slot:     "accessory",
		Rarity:   "传说",
		CritRate: 0.05,
	}); err != nil {
		t.Fatalf("save equipment definition: %v", err)
	}
	critRingInst := seedOwnedInstance(t, store, ctx, "阿明", "crit-ring")
	if _, err := store.EquipItem(ctx, "阿明", critRingInst); err != nil {
		t.Fatalf("equip item: %v", err)
	}

	if _, err := store.ActivateBoss(ctx, BossUpsert{
		ID:    "crit-boss",
		Name:  "暴击测试 Boss",
		MaxHP: 100,
		Parts: []BossPart{
			{X: 0, Y: 0, Type: PartTypeSoft, MaxHP: 100, CurrentHP: 100, Alive: true},
		},
	}); err != nil {
		t.Fatalf("activate boss: %v", err)
	}

	result, err := store.ClickBossPart(ctx, "boss-part:0-0", "阿明")
	if err != nil {
		t.Fatalf("click boss part: %v", err)
	}
	if !result.Critical {
		t.Fatalf("expected critical hit from equipment critRate, got %+v", result)
	}
	if result.Delta != 1 {
		t.Fatalf("expected boss part click delta 1, got %+v", result)
	}

	userState, err := store.GetUserState(ctx, "阿明")
	if err != nil {
		t.Fatalf("get user state: %v", err)
	}
	if userState.CombatStats.CriticalChancePercent != 5 {
		t.Fatalf("expected critical chance to include equipment critRate, got %+v", userState.CombatStats)
	}
}
