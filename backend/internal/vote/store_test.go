package vote

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"long/internal/config"
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
		}), func() {
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
		t.Fatalf("seed feel: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:button:understand", map[string]any{
		"label":   "有没有懂的",
		"count":   "5",
		"sort":    "10",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed understand: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:button:hidden", map[string]any{
		"label":   "隐藏按钮",
		"count":   "99",
		"sort":    "1",
		"enabled": "0",
	}).Err(); err != nil {
		t.Fatalf("seed hidden: %v", err)
	}

	buttons, err := store.ListButtons(ctx)
	if err != nil {
		t.Fatalf("list buttons: %v", err)
	}

	if len(buttons) != 2 {
		t.Fatalf("expected 2 buttons, got %d", len(buttons))
	}
	if buttons[0].Key != "understand" || buttons[1].Key != "feel" {
		t.Fatalf("unexpected order: %+v", buttons)
	}
}

func TestClickButtonUsesNormalCountAndAppliesFallbackImage(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.roll = func(int) int { return 99 }

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:button:wechat-pity", map[string]any{
		"label":   "微信[可怜]表情",
		"count":   "12",
		"sort":    "30",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed button: %v", err)
	}

	updated, err := store.ClickButton(ctx, "wechat-pity")
	if err != nil {
		t.Fatalf("click button: %v", err)
	}

	if updated.Button.Count != 13 {
		t.Fatalf("expected count 13, got %d", updated.Button.Count)
	}
	if updated.Delta != 1 || updated.Critical {
		t.Fatalf("expected normal click, got delta=%d critical=%v", updated.Delta, updated.Critical)
	}
	if updated.Button.ImagePath != "/images/emojipedia-wechat-whimper.png" {
		t.Fatalf("expected fallback image path, got %q", updated.Button.ImagePath)
	}
}

func TestClickButtonAppliesCriticalHitWhenRollMatches(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.roll = func(int) int { return 0 }

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:button:feel", map[string]any{
		"label":   "有感觉吗",
		"count":   "2",
		"sort":    "10",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed button: %v", err)
	}

	updated, err := store.ClickButton(ctx, "feel")
	if err != nil {
		t.Fatalf("click button: %v", err)
	}

	if updated.Button.Count != 7 {
		t.Fatalf("expected crit count 7, got %d", updated.Button.Count)
	}
	if updated.Delta != 5 || !updated.Critical {
		t.Fatalf("expected critical click, got delta=%d critical=%v", updated.Delta, updated.Critical)
	}
}

func TestEnsureDefaultsSeedsOnlyWhenNoButtonsExist(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.EnsureDefaults(ctx, config.DefaultButtons); err != nil {
		t.Fatalf("seed defaults: %v", err)
	}

	buttons, err := store.ListButtons(ctx)
	if err != nil {
		t.Fatalf("list buttons after defaults: %v", err)
	}
	if len(buttons) != 3 {
		t.Fatalf("expected 3 buttons, got %d", len(buttons))
	}

	if err := store.client.HSet(ctx, "vote:button:custom", map[string]any{
		"label":   "自定义",
		"count":   "0",
		"sort":    "40",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed custom button: %v", err)
	}

	if err := store.EnsureDefaults(ctx, []config.ButtonSeed{
		{Slug: "new-default", Label: "不会补进来", Sort: 50},
	}); err != nil {
		t.Fatalf("re-run ensure defaults: %v", err)
	}

	buttons, err = store.ListButtons(ctx)
	if err != nil {
		t.Fatalf("list buttons after second seed: %v", err)
	}
	if len(buttons) != 4 {
		t.Fatalf("expected 4 buttons, got %d", len(buttons))
	}
}
