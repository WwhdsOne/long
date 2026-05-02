package core

import (
	"context"
	"errors"
	"testing"
)

type recordingMessageStore struct {
	created *Message
	page    MessagePage
	deleted string
}

func (m *recordingMessageStore) CreateMessage(_ context.Context, nickname string, content string) (*Message, error) {
	if m.created != nil {
		return m.created, nil
	}
	return &Message{ID: "101", Nickname: nickname, Content: content, CreatedAt: 123}, nil
}

func (m *recordingMessageStore) ListMessages(_ context.Context, _ string, _ int64) (MessagePage, error) {
	return m.page, nil
}

func (m *recordingMessageStore) DeleteMessage(_ context.Context, id string) error {
	m.deleted = id
	return nil
}

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

func TestCreateMessageDelegatesToExternalStoreAfterValidation(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()
	store.messageStore = &recordingMessageStore{}

	ctx := context.Background()
	message, err := store.CreateMessage(ctx, "阿明", "你好呀")
	if err != nil {
		t.Fatalf("create message with external store: %v", err)
	}
	if message.ID != "101" || message.Nickname != "阿明" {
		t.Fatalf("unexpected delegated message: %+v", message)
	}

	if _, err := store.CreateMessage(ctx, "阿明", "我是XJP后援会"); !errors.Is(err, ErrSensitiveContent) {
		t.Fatalf("expected sensitive content validation before delegation, got %v", err)
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
	_ = seedOwnedInstance(t, store, ctx, "阿明", "wood-sword")
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
	if len(state.Inventory) != 1 {
		t.Fatalf("expected one instance inventory item, got %+v", state.Inventory)
	}
	if state.Inventory[0].EnhanceLevel != 0 {
		t.Fatalf("expected legacy upgrade hash to be ignored, got %+v", state.Inventory[0])
	}
}
