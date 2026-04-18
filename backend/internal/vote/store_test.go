package vote

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"long/internal/config"
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

	updated, err := store.ClickButton(ctx, "wechat-pity", "阿明")
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
	if updated.UserStats.Nickname != "阿明" || updated.UserStats.ClickCount != 1 {
		t.Fatalf("unexpected user stats: %+v", updated.UserStats)
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

	updated, err := store.ClickButton(ctx, "feel", "阿明")
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

func TestClickButtonUsesUserAbilityCriticalChanceOverride(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.roll = func(int) int { return 20 }

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:button:feel", map[string]any{
		"label":   "有感觉吗",
		"count":   "2",
		"sort":    "10",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed button: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:user-ability:阿明", map[string]any{
		"critical_chance_percent": "30",
	}).Err(); err != nil {
		t.Fatalf("seed user ability: %v", err)
	}

	updated, err := store.ClickButton(ctx, "feel", "阿明")
	if err != nil {
		t.Fatalf("click button: %v", err)
	}

	if updated.Delta != 5 || !updated.Critical {
		t.Fatalf("expected critical click from user ability override, got delta=%d critical=%v", updated.Delta, updated.Critical)
	}
}

func TestClickButtonFallsBackToDefaultCriticalChanceWhenUserAbilityMissing(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.roll = func(int) int { return 20 }

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:button:feel", map[string]any{
		"label":   "有感觉吗",
		"count":   "2",
		"sort":    "10",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed button: %v", err)
	}

	updated, err := store.ClickButton(ctx, "feel", "阿明")
	if err != nil {
		t.Fatalf("click button: %v", err)
	}

	if updated.Delta != 1 || updated.Critical {
		t.Fatalf("expected default non-critical click without user ability override, got delta=%d critical=%v", updated.Delta, updated.Critical)
	}
}

func TestClickButtonIgnoresInvalidUserAbilityCriticalChance(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.roll = func(int) int { return 20 }

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:button:feel", map[string]any{
		"label":   "有感觉吗",
		"count":   "2",
		"sort":    "10",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed button: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:user-ability:阿明", map[string]any{
		"critical_chance_percent": "200",
	}).Err(); err != nil {
		t.Fatalf("seed invalid user ability: %v", err)
	}

	updated, err := store.ClickButton(ctx, "feel", "阿明")
	if err != nil {
		t.Fatalf("click button: %v", err)
	}

	if updated.Delta != 1 || updated.Critical {
		t.Fatalf("expected invalid user ability to fall back to default non-critical click, got delta=%d critical=%v", updated.Delta, updated.Critical)
	}
}

func TestClickButtonAllowsUserAbilityToDisableCriticalChance(t *testing.T) {
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
	if err := store.client.HSet(ctx, "vote:user-ability:阿明", map[string]any{
		"critical_chance_percent": "0",
	}).Err(); err != nil {
		t.Fatalf("seed zero user ability: %v", err)
	}

	updated, err := store.ClickButton(ctx, "feel", "阿明")
	if err != nil {
		t.Fatalf("click button: %v", err)
	}

	if updated.Delta != 1 || updated.Critical {
		t.Fatalf("expected zero critical chance override to disable critical, got delta=%d critical=%v", updated.Delta, updated.Critical)
	}
}

func TestClickButtonUsesUserAbilityCriticalCountOverride(t *testing.T) {
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
	if err := store.client.HSet(ctx, "vote:user-ability:阿明", map[string]any{
		"critical_count": "9",
	}).Err(); err != nil {
		t.Fatalf("seed user ability critical count: %v", err)
	}

	updated, err := store.ClickButton(ctx, "feel", "阿明")
	if err != nil {
		t.Fatalf("click button: %v", err)
	}

	if updated.Delta != 9 || !updated.Critical {
		t.Fatalf("expected critical count override to apply, got delta=%d critical=%v", updated.Delta, updated.Critical)
	}
}

func TestClickButtonIgnoresInvalidUserAbilityCriticalCount(t *testing.T) {
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
	if err := store.client.HSet(ctx, "vote:user-ability:阿明", map[string]any{
		"critical_count": "1",
	}).Err(); err != nil {
		t.Fatalf("seed invalid user ability critical count: %v", err)
	}

	updated, err := store.ClickButton(ctx, "feel", "阿明")
	if err != nil {
		t.Fatalf("click button: %v", err)
	}

	if updated.Delta != 5 || !updated.Critical {
		t.Fatalf("expected invalid critical count override to fall back to default, got delta=%d critical=%v", updated.Delta, updated.Critical)
	}
}

func TestGetStateReturnsLeaderboardAndUserStats(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.roll = func(int) int { return 99 }

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:button:feel", map[string]any{
		"label":   "有感觉吗",
		"count":   "0",
		"sort":    "10",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed feel: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:button:understand", map[string]any{
		"label":   "有没有懂的",
		"count":   "0",
		"sort":    "20",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed understand: %v", err)
	}

	if _, err := store.ClickButton(ctx, "feel", "阿明"); err != nil {
		t.Fatalf("click feel by 阿明: %v", err)
	}
	if _, err := store.ClickButton(ctx, "understand", "小红"); err != nil {
		t.Fatalf("click understand by 小红: %v", err)
	}
	if _, err := store.ClickButton(ctx, "understand", "小红"); err != nil {
		t.Fatalf("second click understand by 小红: %v", err)
	}

	state, err := store.GetState(ctx, "阿明")
	if err != nil {
		t.Fatalf("get state: %v", err)
	}

	if len(state.Leaderboard) != 2 {
		t.Fatalf("expected 2 leaderboard entries, got %d", len(state.Leaderboard))
	}
	if state.Leaderboard[0].Nickname != "小红" || state.Leaderboard[0].ClickCount != 2 {
		t.Fatalf("unexpected leaderboard winner: %+v", state.Leaderboard[0])
	}
	if state.UserStats == nil || state.UserStats.Nickname != "阿明" || state.UserStats.ClickCount != 1 {
		t.Fatalf("unexpected user stats: %+v", state.UserStats)
	}
}

func TestClickButtonRejectsEmptyNickname(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:button:feel", map[string]any{
		"label":   "有感觉吗",
		"count":   "0",
		"sort":    "10",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed feel: %v", err)
	}

	if _, err := store.ClickButton(ctx, "feel", "   "); !errors.Is(err, ErrInvalidNickname) {
		t.Fatalf("expected invalid nickname error, got %v", err)
	}
}

func TestClickButtonRejectsSensitiveNickname(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:button:feel", map[string]any{
		"label":   "有感觉吗",
		"count":   "0",
		"sort":    "10",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed feel: %v", err)
	}

	if _, err := store.ClickButton(ctx, "feel", "XJP后援会"); !errors.Is(err, ErrSensitiveNickname) {
		t.Fatalf("expected sensitive nickname error, got %v", err)
	}
}

func TestGetStateRejectsSensitiveNickname(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:button:feel", map[string]any{
		"label":   "有感觉吗",
		"count":   "0",
		"sort":    "10",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed feel: %v", err)
	}

	if _, err := store.GetState(ctx, "我是习近平"); !errors.Is(err, ErrSensitiveNickname) {
		t.Fatalf("expected sensitive nickname error, got %v", err)
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

func TestEquipItemAndClickButtonApplyEquippedBonusWithoutBoss(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.roll = func(int) int { return 99 }

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:button:feel", map[string]any{
		"label":   "有感觉吗",
		"count":   "0",
		"sort":    "10",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed feel: %v", err)
	}
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

	state, err := store.EquipItem(ctx, "阿明", "wood-sword")
	if err != nil {
		t.Fatalf("equip item: %v", err)
	}

	if state.Loadout.Weapon == nil || state.Loadout.Weapon.ItemID != "wood-sword" {
		t.Fatalf("expected weapon slot to equip wood-sword, got %+v", state.Loadout.Weapon)
	}
	if state.CombatStats.EffectiveIncrement != 3 {
		t.Fatalf("expected effective increment 3 after equip, got %+v", state.CombatStats)
	}

	result, err := store.ClickButton(ctx, "feel", "阿明")
	if err != nil {
		t.Fatalf("click feel: %v", err)
	}

	if result.Delta != 3 || result.Critical {
		t.Fatalf("expected non-critical delta 3, got delta=%d critical=%v", result.Delta, result.Critical)
	}
	if result.Button.Count != 3 {
		t.Fatalf("expected button count 3, got %d", result.Button.Count)
	}
	if result.UserStats.ClickCount != 3 {
		t.Fatalf("expected user click count 3, got %+v", result.UserStats)
	}
}

func TestClickButtonDefeatsActiveBossAndAwardsLootOnce(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.roll = func(int) int { return 99 }

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:button:feel", map[string]any{
		"label":   "有感觉吗",
		"count":   "0",
		"sort":    "10",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed feel: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:equip:def:wood-sword", map[string]any{
		"name":         "木剑",
		"slot":         "weapon",
		"bonus_clicks": "2",
	}).Err(); err != nil {
		t.Fatalf("seed wood-sword definition: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:equip:def:cloth-armor", map[string]any{
		"name":         "布甲",
		"slot":         "armor",
		"bonus_clicks": "1",
	}).Err(); err != nil {
		t.Fatalf("seed cloth-armor definition: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:user-inventory:阿明", map[string]any{
		"wood-sword": "1",
	}).Err(); err != nil {
		t.Fatalf("seed inventory: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:user-loadout:阿明", map[string]any{
		"weapon": "wood-sword",
	}).Err(); err != nil {
		t.Fatalf("seed loadout: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:boss:current", map[string]any{
		"id":         "slime-king",
		"name":       "史莱姆王",
		"status":     "active",
		"max_hp":     "3",
		"current_hp": "3",
	}).Err(); err != nil {
		t.Fatalf("seed boss: %v", err)
	}
	if err := store.client.ZAdd(ctx, "vote:boss:slime-king:loot", redis.Z{
		Score:  1,
		Member: "cloth-armor",
	}).Err(); err != nil {
		t.Fatalf("seed boss loot: %v", err)
	}

	result, err := store.ClickButton(ctx, "feel", "阿明")
	if err != nil {
		t.Fatalf("click feel: %v", err)
	}

	if result.Delta != 3 {
		t.Fatalf("expected delta 3 from equipped clicks, got %d", result.Delta)
	}
	if result.Boss == nil || result.Boss.Status != "defeated" || result.Boss.CurrentHP != 0 {
		t.Fatalf("expected defeated boss payload, got %+v", result.Boss)
	}
	if result.LastReward == nil || result.LastReward.ItemID != "cloth-armor" {
		t.Fatalf("expected cloth-armor reward, got %+v", result.LastReward)
	}

	state, err := store.GetState(ctx, "阿明")
	if err != nil {
		t.Fatalf("get state after boss kill: %v", err)
	}

	if state.Boss == nil || state.Boss.Status != "defeated" {
		t.Fatalf("expected defeated boss in state, got %+v", state.Boss)
	}
	if len(state.BossLeaderboard) != 1 || state.BossLeaderboard[0].Damage != 3 {
		t.Fatalf("expected boss leaderboard damage 3, got %+v", state.BossLeaderboard)
	}
	if state.Inventory[0].ItemID == "" {
		t.Fatalf("expected inventory entries, got %+v", state.Inventory)
	}

	var foundReward bool
	for _, item := range state.Inventory {
		if item.ItemID == "cloth-armor" && item.Quantity == 1 {
			foundReward = true
		}
	}
	if !foundReward {
		t.Fatalf("expected rewarded cloth-armor in inventory, got %+v", state.Inventory)
	}
}
