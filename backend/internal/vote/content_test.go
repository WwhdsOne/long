package vote

import (
	"context"
	"errors"
	"testing"
)

func TestGetStateIncludesLatestAnnouncement(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	announcement, err := store.SaveAnnouncement(ctx, AnnouncementUpsert{
		Title:   "新版本上线",
		Content: "开放留言墙和装备升星。",
		Active:  true,
	})
	if err != nil {
		t.Fatalf("save announcement: %v", err)
	}

	state, err := store.GetState(ctx, "")
	if err != nil {
		t.Fatalf("get state: %v", err)
	}

	if state.LatestAnnouncement == nil {
		t.Fatal("expected latest announcement in state, got nil")
	}
	if state.LatestAnnouncement.ID != announcement.ID || state.LatestAnnouncement.Title != "新版本上线" {
		t.Fatalf("unexpected latest announcement: %+v", state.LatestAnnouncement)
	}
}

func TestCreateMessageRejectsSensitiveContent(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if _, err := store.CreateMessage(ctx, "阿明", "我是XJP后援会"); !errors.Is(err, ErrSensitiveContent) {
		t.Fatalf("expected sensitive content error, got %v", err)
	}
}

func TestCreateMessageAndListMessages(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	first, err := store.CreateMessage(ctx, "阿明", "第一条留言")
	if err != nil {
		t.Fatalf("create first message: %v", err)
	}
	second, err := store.CreateMessage(ctx, "小红", "第二条留言")
	if err != nil {
		t.Fatalf("create second message: %v", err)
	}

	page, err := store.ListMessages(ctx, "", 10)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}

	if len(page.Items) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(page.Items))
	}
	if page.Items[0].ID != second.ID || page.Items[0].Nickname != "小红" {
		t.Fatalf("expected latest message first, got %+v", page.Items[0])
	}
	if page.Items[1].ID != first.ID || page.Items[1].Nickname != "阿明" {
		t.Fatalf("expected first message second, got %+v", page.Items[1])
	}
}

func TestSynthesizeItemConsumesThreeCopiesAndImprovesStats(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.roll = func(int) int { return 0 }

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:equip:def:wood-sword", map[string]any{
		"name":         "木剑",
		"slot":         "weapon",
		"bonus_clicks": "2",
	}).Err(); err != nil {
		t.Fatalf("seed equipment definition: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:user-inventory:阿明", map[string]any{
		"wood-sword": "3",
	}).Err(); err != nil {
		t.Fatalf("seed inventory: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:user-loadout:阿明", map[string]any{
		"weapon": "wood-sword",
	}).Err(); err != nil {
		t.Fatalf("seed loadout: %v", err)
	}

	state, err := store.SynthesizeItem(ctx, "阿明", "wood-sword")
	if err != nil {
		t.Fatalf("synthesize item: %v", err)
	}

	if state.Loadout.Weapon == nil {
		t.Fatal("expected weapon to remain equipped after synthesize")
	}
	if state.Loadout.Weapon.Name != "木剑 +1" {
		t.Fatalf("expected display name 木剑 +1, got %+v", state.Loadout.Weapon)
	}
	if state.Loadout.Weapon.StarLevel != 1 {
		t.Fatalf("expected star level 1, got %+v", state.Loadout.Weapon)
	}
	if state.Loadout.Weapon.Quantity != 1 {
		t.Fatalf("expected quantity 1 after consuming 3 copies, got %+v", state.Loadout.Weapon)
	}
	if state.Loadout.Weapon.BonusClicks != 3 {
		t.Fatalf("expected click bonus 3 after synthesize, got %+v", state.Loadout.Weapon)
	}
	if state.CombatStats.NormalDamage != 4 {
		t.Fatalf("expected normal damage 4 after synthesize, got %+v", state.CombatStats)
	}
}

func TestSynthesizeItemRejectsWhenAtMaxStar(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:equip:def:wood-sword", map[string]any{
		"name":         "木剑",
		"slot":         "weapon",
		"bonus_clicks": "2",
	}).Err(); err != nil {
		t.Fatalf("seed equipment definition: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:user-inventory:阿明", map[string]any{
		"wood-sword": "9",
	}).Err(); err != nil {
		t.Fatalf("seed inventory: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:user-equip-upgrade:阿明:wood-sword", map[string]any{
		"star_level": "5",
	}).Err(); err != nil {
		t.Fatalf("seed upgrade: %v", err)
	}

	if _, err := store.SynthesizeItem(ctx, "阿明", "wood-sword"); !errors.Is(err, ErrEquipmentMaxStar) {
		t.Fatalf("expected max star error, got %v", err)
	}
}
