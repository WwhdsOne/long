package vote

import (
	"context"
	"errors"
	"testing"
)

func TestGetSnapshotIncludesAnnouncementVersion(t *testing.T) {
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

	snapshot, err := store.GetSnapshot(ctx)
	if err != nil {
		t.Fatalf("get snapshot: %v", err)
	}

	if snapshot.AnnouncementVersion != announcement.ID {
		t.Fatalf("unexpected latest announcement version: %+v", snapshot)
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

func TestGetStateIgnoresLegacyEquipmentUpgradeFields(t *testing.T) {
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
		"wood-sword": "1",
	}).Err(); err != nil {
		t.Fatalf("seed inventory: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:user-equip-upgrade:阿明:wood-sword", map[string]any{
		"star_level":                    "9",
		"bonus_clicks":                  "99",
		"bonus_critical_chance_percent": "50",
		"bonus_critical_count":          "99",
		"reforge_pity_counter":          "30",
	}).Err(); err != nil {
		t.Fatalf("seed legacy upgrade: %v", err)
	}

	state, err := store.GetState(ctx, "阿明")
	if err != nil {
		t.Fatalf("get state: %v", err)
	}

		_ = state.Inventory[0]
	}