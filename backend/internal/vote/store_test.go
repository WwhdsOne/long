package vote

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/bytedance/sonic"
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

func TestListButtonsPrefersExplicitIndexWhenPresent(t *testing.T) {
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
	if err := store.client.HSet(ctx, "vote:button:orphan", map[string]any{
		"label":   "孤儿按钮",
		"count":   "9",
		"sort":    "1",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed orphan: %v", err)
	}
	if err := store.client.ZAdd(ctx, "vote:buttons:index", redis.Z{
		Score:  20,
		Member: "feel",
	}).Err(); err != nil {
		t.Fatalf("seed button index: %v", err)
	}

	buttons, err := store.ListButtons(ctx)
	if err != nil {
		t.Fatalf("list buttons: %v", err)
	}

	if len(buttons) != 1 || buttons[0].Key != "feel" {
		t.Fatalf("expected only indexed button, got %+v", buttons)
	}
}

func TestGetUserStatsFallsBackToLegacyStringValue(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.client.Set(ctx, store.userPrefix+"阿明", "100", 0).Err(); err != nil {
		t.Fatalf("seed legacy user stat: %v", err)
	}

	stats, err := store.GetUserStats(ctx, "阿明")
	if err != nil {
		t.Fatalf("get user stats: %v", err)
	}

	if stats.Nickname != "阿明" {
		t.Fatalf("expected nickname 阿明, got %q", stats.Nickname)
	}
	if stats.ClickCount != 100 {
		t.Fatalf("expected click count 100, got %d", stats.ClickCount)
	}
}

func TestSyncButtonIndexFindsButtonsAddedOutsideSupportedWritePath(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.EnsureDefaults(ctx, config.DefaultButtons); err != nil {
		t.Fatalf("seed defaults: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:button:custom", map[string]any{
		"label":   "自定义",
		"count":   "0",
		"sort":    "40",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed custom button directly: %v", err)
	}

	changed, err := store.SyncButtonIndex(ctx)
	if err != nil {
		t.Fatalf("sync button index: %v", err)
	}
	if !changed {
		t.Fatal("expected sync to detect new custom button")
	}

	buttons, err := store.ListButtons(ctx)
	if err != nil {
		t.Fatalf("list buttons after sync: %v", err)
	}
	if len(buttons) != 4 || buttons[3].Key != "custom" {
		t.Fatalf("expected custom button to appear after sync, got %+v", buttons)
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

func TestClickButtonDoesNotReturnHistoricalRewardsOnNormalClick(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.roll = func(int) int { return 99 }

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:button:feel", map[string]any{
		"label":   "有感觉吗",
		"count":   "2",
		"sort":    "10",
		"enabled": "1",
	}).Err(); err != nil {
		t.Fatalf("seed button: %v", err)
	}
	if err := store.client.HSet(ctx, store.lastRewardKey("阿明"), rewardRecordValues([]Reward{
		{
			BossID:    "slime-king",
			BossName:  "史莱姆王",
			ItemID:    "cloth-armor",
			ItemName:  "布甲",
			GrantedAt: 1713744300,
		},
	})).Err(); err != nil {
		t.Fatalf("seed last reward: %v", err)
	}

	updated, err := store.ClickButton(ctx, "feel", "阿明")
	if err != nil {
		t.Fatalf("click button: %v", err)
	}

	if updated.LastReward != nil {
		t.Fatalf("expected normal click to omit historical last reward, got %+v", updated.LastReward)
	}
	if len(updated.RecentRewards) != 0 {
		t.Fatalf("expected normal click to omit historical recent rewards, got %+v", updated.RecentRewards)
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

func TestClickButtonAppliesEquippedCriticalChanceBonus(t *testing.T) {
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
	if err := store.client.HSet(ctx, "vote:equip:def:lucky-ring", map[string]any{
		"name":                          "幸运戒指",
		"slot":                          "accessory",
		"bonus_critical_chance_percent": "30",
	}).Err(); err != nil {
		t.Fatalf("seed equipment definition: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:user-inventory:阿明", map[string]any{
		"lucky-ring": "1",
	}).Err(); err != nil {
		t.Fatalf("seed inventory: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:user-loadout:阿明", map[string]any{
		"accessory": "lucky-ring",
	}).Err(); err != nil {
		t.Fatalf("seed loadout: %v", err)
	}

	updated, err := store.ClickButton(ctx, "feel", "阿明")
	if err != nil {
		t.Fatalf("click button: %v", err)
	}

	if updated.Delta != 5 || !updated.Critical {
		t.Fatalf("expected equipped critical chance bonus to trigger crit, got delta=%d critical=%v", updated.Delta, updated.Critical)
	}
}

func TestClickButtonUsesGlobalCriticalChanceWhenNoEquipmentBonus(t *testing.T) {
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

func TestClickButtonAppliesEquippedCriticalCountBonus(t *testing.T) {
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
	if err := store.client.HSet(ctx, "vote:equip:def:berserk-ring", map[string]any{
		"name":                 "狂暴戒指",
		"slot":                 "accessory",
		"bonus_critical_count": "4",
	}).Err(); err != nil {
		t.Fatalf("seed equipment definition: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:user-inventory:阿明", map[string]any{
		"berserk-ring": "1",
	}).Err(); err != nil {
		t.Fatalf("seed inventory: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:user-loadout:阿明", map[string]any{
		"accessory": "berserk-ring",
	}).Err(); err != nil {
		t.Fatalf("seed loadout: %v", err)
	}

	updated, err := store.ClickButton(ctx, "feel", "阿明")
	if err != nil {
		t.Fatalf("click button: %v", err)
	}

	if updated.Delta != 9 || !updated.Critical {
		t.Fatalf("expected equipped critical count bonus to raise crit damage, got delta=%d critical=%v", updated.Delta, updated.Critical)
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

func TestClickButtonAddsPlayerToExplicitIndex(t *testing.T) {
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
		t.Fatalf("seed button: %v", err)
	}

	if _, err := store.ClickButton(ctx, "feel", "阿明"); err != nil {
		t.Fatalf("click button: %v", err)
	}

	members, err := store.client.ZRange(ctx, "vote:players:index", 0, -1).Result()
	if err != nil {
		t.Fatalf("load players index: %v", err)
	}
	if len(members) != 1 || members[0] != "阿明" {
		t.Fatalf("expected 阿明 in players index, got %+v", members)
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

	if err := store.SaveButton(ctx, ButtonUpsert{
		Slug:    "custom",
		Label:   "自定义",
		Sort:    40,
		Enabled: true,
	}); err != nil {
		t.Fatalf("save custom button: %v", err)
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
	if state.CombatStats.NormalDamage != 3 || state.CombatStats.CriticalDamage != 7 {
		t.Fatalf("expected actual damage 3/7 after equip, got %+v", state.CombatStats)
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
	if result.LastReward.BossName != "史莱姆王" {
		t.Fatalf("expected reward boss name 史莱姆王, got %+v", result.LastReward)
	}

	state, err := store.GetState(ctx, "阿明")
	if err != nil {
		t.Fatalf("get state after boss kill: %v", err)
	}

	if state.Boss == nil || state.Boss.Status != "defeated" {
		t.Fatalf("expected defeated boss in state, got %+v", state.Boss)
	}
	if state.LastReward == nil || state.LastReward.BossName != "史莱姆王" {
		t.Fatalf("expected persisted reward boss name 史莱姆王, got %+v", state.LastReward)
	}
	if len(state.BossLeaderboard) != 1 || state.BossLeaderboard[0].Damage != 3 {
		t.Fatalf("expected boss leaderboard damage 3, got %+v", state.BossLeaderboard)
	}
	if state.Inventory[0].ItemID == "" {
		t.Fatalf("expected inventory entries, got %+v", state.Inventory)
	}

	resources, err := store.GetBossResources(ctx)
	if err != nil {
		t.Fatalf("get boss resources after boss kill: %v", err)
	}
	if len(resources.BossLoot) != 1 || resources.BossLoot[0].ItemID != "cloth-armor" {
		t.Fatalf("expected current boss loot to be returned, got %+v", resources.BossLoot)
	}
	if resources.BossLoot[0].ItemName != "布甲" || resources.BossLoot[0].BonusClicks != 1 {
		t.Fatalf("expected boss loot attributes to include equipment stats, got %+v", resources.BossLoot[0])
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

func TestSetBossCycleEnabledSpawnsBossFromPoolImmediately(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.SaveBossTemplate(ctx, BossTemplateUpsert{
		ID:    "slime-king",
		Name:  "史莱姆王",
		MaxHP: 30,
	}); err != nil {
		t.Fatalf("save slime template: %v", err)
	}
	if err := store.SaveBossTemplate(ctx, BossTemplateUpsert{
		ID:    "dragon",
		Name:  "火龙",
		MaxHP: 80,
	}); err != nil {
		t.Fatalf("save dragon template: %v", err)
	}
	if err := store.SetBossTemplateLoot(ctx, "dragon", []BossLootEntry{
		{ItemID: "cloth-armor", Weight: 1},
	}); err != nil {
		t.Fatalf("set dragon loot: %v", err)
	}

	store.roll = func(limit int) int {
		if limit <= 1 {
			return 0
		}
		return 1
	}

	boss, err := store.SetBossCycleEnabled(ctx, true)
	if err != nil {
		t.Fatalf("enable boss cycle: %v", err)
	}

	if boss == nil || boss.Status != bossStatusActive {
		t.Fatalf("expected active boss after enabling cycle, got %+v", boss)
	}
	if boss.TemplateID != "dragon" || boss.Name != "火龙" {
		t.Fatalf("expected dragon template to be activated, got %+v", boss)
	}
	if boss.ID == boss.TemplateID {
		t.Fatalf("expected unique boss instance id, got %+v", boss)
	}

	adminState, err := store.GetAdminState(ctx)
	if err != nil {
		t.Fatalf("get admin state: %v", err)
	}

	if !adminState.BossCycleEnabled {
		t.Fatal("expected admin state to report enabled cycle")
	}
	if len(adminState.BossPool) != 2 {
		t.Fatalf("expected 2 boss templates, got %+v", adminState.BossPool)
	}
	if len(adminState.Loot) != 1 || adminState.Loot[0].ItemID != "cloth-armor" {
		t.Fatalf("expected current boss loot copied from template, got %+v", adminState.Loot)
	}
}

func TestClickButtonDefeatAutoSpawnsNextBossWhenCycleEnabled(t *testing.T) {
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
	if err := store.client.HSet(ctx, "vote:equip:def:cloth-armor", map[string]any{
		"name":         "布甲",
		"slot":         "armor",
		"bonus_clicks": "1",
	}).Err(); err != nil {
		t.Fatalf("seed cloth-armor definition: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:equip:def:fire-ring", map[string]any{
		"name":                          "火戒",
		"slot":                          "accessory",
		"bonus_critical_chance_percent": "1",
	}).Err(); err != nil {
		t.Fatalf("seed fire-ring definition: %v", err)
	}
	if err := store.SaveBossTemplate(ctx, BossTemplateUpsert{
		ID:    "slime-king",
		Name:  "史莱姆王",
		MaxHP: 1,
	}); err != nil {
		t.Fatalf("save slime template: %v", err)
	}
	if err := store.SaveBossTemplate(ctx, BossTemplateUpsert{
		ID:    "dragon",
		Name:  "火龙",
		MaxHP: 5,
	}); err != nil {
		t.Fatalf("save dragon template: %v", err)
	}
	if err := store.SetBossTemplateLoot(ctx, "slime-king", []BossLootEntry{
		{ItemID: "cloth-armor", Weight: 1},
	}); err != nil {
		t.Fatalf("set slime loot: %v", err)
	}
	if err := store.SetBossTemplateLoot(ctx, "dragon", []BossLootEntry{
		{ItemID: "fire-ring", Weight: 1},
	}); err != nil {
		t.Fatalf("set dragon loot: %v", err)
	}

	rolls := []int{0, 0, 1}
	store.roll = func(limit int) int {
		if limit <= 1 {
			return 0
		}
		if len(rolls) == 0 {
			return 0
		}
		next := rolls[0]
		rolls = rolls[1:]
		if next >= limit {
			return limit - 1
		}
		return next
	}

	firstBoss, err := store.SetBossCycleEnabled(ctx, true)
	if err != nil {
		t.Fatalf("enable boss cycle: %v", err)
	}
	if firstBoss == nil || firstBoss.TemplateID != "slime-king" {
		t.Fatalf("expected slime boss first, got %+v", firstBoss)
	}

	result, err := store.ClickButton(ctx, "feel", "阿明")
	if err != nil {
		t.Fatalf("click feel: %v", err)
	}

	if !result.BroadcastUserAll {
		t.Fatal("expected boss kill to trigger user refresh for all participants")
	}
	if result.Boss == nil || result.Boss.Status != bossStatusActive {
		t.Fatalf("expected next active boss after kill, got %+v", result.Boss)
	}
	if result.Boss.TemplateID != "dragon" || result.Boss.Name != "火龙" {
		t.Fatalf("expected dragon to replace defeated boss, got %+v", result.Boss)
	}
	if result.LastReward == nil || result.LastReward.BossName != "史莱姆王" {
		t.Fatalf("expected reward from defeated slime boss, got %+v", result.LastReward)
	}

	state, err := store.GetState(ctx, "阿明")
	if err != nil {
		t.Fatalf("get state after auto rotate: %v", err)
	}
	if state.Boss == nil || state.Boss.Status != bossStatusActive || state.Boss.TemplateID != "dragon" {
		t.Fatalf("expected current boss to be dragon, got %+v", state.Boss)
	}

	resources, err := store.GetBossResources(ctx)
	if err != nil {
		t.Fatalf("get boss resources after auto rotate: %v", err)
	}
	if len(resources.BossLoot) != 1 || resources.BossLoot[0].ItemID != "fire-ring" {
		t.Fatalf("expected current boss loot to switch to dragon loot, got %+v", resources.BossLoot)
	}

	history, err := store.ListBossHistory(ctx)
	if err != nil {
		t.Fatalf("list boss history: %v", err)
	}
	if len(history) != 1 || history[0].TemplateID != "slime-king" || history[0].Status != bossStatusDefeated {
		t.Fatalf("expected defeated slime boss in history, got %+v", history)
	}
}

func TestUpdatingBossTemplateLootDoesNotRewriteCurrentBossLoot(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:equip:def:cloth-armor", map[string]any{
		"name":         "布甲",
		"slot":         "armor",
		"bonus_clicks": "1",
	}).Err(); err != nil {
		t.Fatalf("seed cloth-armor definition: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:equip:def:fire-ring", map[string]any{
		"name":                          "火戒",
		"slot":                          "accessory",
		"bonus_critical_chance_percent": "1",
	}).Err(); err != nil {
		t.Fatalf("seed fire-ring definition: %v", err)
	}
	if err := store.SaveBossTemplate(ctx, BossTemplateUpsert{
		ID:    "slime-king",
		Name:  "史莱姆王",
		MaxHP: 10,
	}); err != nil {
		t.Fatalf("save slime template: %v", err)
	}
	if err := store.SetBossTemplateLoot(ctx, "slime-king", []BossLootEntry{
		{ItemID: "cloth-armor", Weight: 1},
	}); err != nil {
		t.Fatalf("set initial template loot: %v", err)
	}

	store.roll = func(int) int { return 0 }
	if _, err := store.SetBossCycleEnabled(ctx, true); err != nil {
		t.Fatalf("enable boss cycle: %v", err)
	}

	if err := store.SetBossTemplateLoot(ctx, "slime-king", []BossLootEntry{
		{ItemID: "fire-ring", Weight: 1},
	}); err != nil {
		t.Fatalf("update template loot: %v", err)
	}

	adminState, err := store.GetAdminState(ctx)
	if err != nil {
		t.Fatalf("get admin state: %v", err)
	}

	if len(adminState.Loot) != 1 || adminState.Loot[0].ItemID != "cloth-armor" {
		t.Fatalf("expected current boss loot snapshot to remain cloth-armor, got %+v", adminState.Loot)
	}
	if len(adminState.BossPool) != 1 || len(adminState.BossPool[0].Loot) != 1 || adminState.BossPool[0].Loot[0].ItemID != "fire-ring" {
		t.Fatalf("expected template loot to update independently, got %+v", adminState.BossPool)
	}
}

func TestGetAdminStateReturnsEmptyCollectionsWithoutBoss(t *testing.T) {
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

	state, err := store.GetAdminState(ctx)
	if err != nil {
		t.Fatalf("get admin state: %v", err)
	}

	if state.BossLeaderboard == nil {
		t.Fatalf("expected empty boss leaderboard slice, got nil")
	}
	if state.Loot == nil {
		t.Fatalf("expected empty loot slice, got nil")
	}
}

func TestListEquipmentDefinitionsPrefersExplicitIndexWhenPresent(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:equip:def:wood-sword", map[string]any{
		"name":         "木剑",
		"slot":         "weapon",
		"bonus_clicks": "2",
	}).Err(); err != nil {
		t.Fatalf("seed indexed equipment: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:equip:def:orphan", map[string]any{
		"name":         "孤儿装备",
		"slot":         "weapon",
		"bonus_clicks": "9",
	}).Err(); err != nil {
		t.Fatalf("seed orphan equipment: %v", err)
	}
	if err := store.client.SAdd(ctx, "vote:equipment:index", "wood-sword").Err(); err != nil {
		t.Fatalf("seed equipment index: %v", err)
	}

	items, err := store.ListEquipmentDefinitions(ctx)
	if err != nil {
		t.Fatalf("list equipment definitions: %v", err)
	}

	if len(items) != 1 || items[0].ItemID != "wood-sword" {
		t.Fatalf("expected only indexed equipment, got %+v", items)
	}
}

func TestListPlayerOverviewsPrefersExplicitIndexWhenPresent(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:user:阿明", map[string]any{
		"nickname":    "阿明",
		"click_count": "5",
	}).Err(); err != nil {
		t.Fatalf("seed indexed player: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:user:小红", map[string]any{
		"nickname":    "小红",
		"click_count": "9",
	}).Err(); err != nil {
		t.Fatalf("seed orphan player: %v", err)
	}
	if err := store.client.ZAdd(ctx, "vote:players:index", redis.Z{
		Score:  1710000000,
		Member: "阿明",
	}).Err(); err != nil {
		t.Fatalf("seed players index: %v", err)
	}

	players, err := store.ListPlayerOverviews(ctx)
	if err != nil {
		t.Fatalf("list player overviews: %v", err)
	}

	if len(players) != 1 || players[0].Nickname != "阿明" {
		t.Fatalf("expected only indexed player, got %+v", players)
	}
}

func TestSaveButtonPersistsTags(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.SaveButton(ctx, ButtonUpsert{
		Slug:    "feel",
		Label:   "有感觉吗",
		Sort:    10,
		Enabled: true,
		Tags:    []string{"日常", "互动"},
	}); err != nil {
		t.Fatalf("save button: %v", err)
	}

	buttons, err := store.ListButtons(ctx)
	if err != nil {
		t.Fatalf("list buttons: %v", err)
	}

	if len(buttons) != 1 {
		t.Fatalf("expected 1 button, got %+v", buttons)
	}
	if len(buttons[0].Tags) != 2 || buttons[0].Tags[0] != "日常" || buttons[0].Tags[1] != "互动" {
		t.Fatalf("expected tags to round-trip, got %+v", buttons[0])
	}
}

func TestGetSnapshotIncludesAllButtonsAndDropRates(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()
	store.roll = func(int) int { return 99 }

	store.now = func() time.Time {
		return time.Unix(1713744000, 0)
	}

	ctx := context.Background()
	if err := store.SaveButton(ctx, ButtonUpsert{
		Slug:    "feel",
		Label:   "有感觉吗",
		Sort:    20,
		Enabled: true,
	}); err != nil {
		t.Fatalf("save feel button: %v", err)
	}
	if err := store.SaveButton(ctx, ButtonUpsert{
		Slug:    "understand",
		Label:   "有没有懂的",
		Sort:    30,
		Enabled: true,
	}); err != nil {
		t.Fatalf("save understand button: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:equip:def:cloth-armor", map[string]any{
		"name":         "布甲",
		"slot":         "armor",
		"bonus_clicks": "1",
	}).Err(); err != nil {
		t.Fatalf("seed cloth-armor definition: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:equip:def:fire-ring", map[string]any{
		"name":                          "火戒",
		"slot":                          "accessory",
		"bonus_critical_chance_percent": "1",
	}).Err(); err != nil {
		t.Fatalf("seed fire-ring definition: %v", err)
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
	}, nil); err != nil {
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

func TestGetSnapshotReturnsSlimPayloadAndAnnouncementVersion(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.now = func() time.Time {
		return time.Unix(1713744000, 0)
	}

	ctx := context.Background()
	if err := store.SaveButton(ctx, ButtonUpsert{
		Slug:    "feel",
		Label:   "有感觉吗",
		Sort:    20,
		Enabled: true,
	}); err != nil {
		t.Fatalf("save button: %v", err)
	}
	if err := store.client.HSet(ctx, "vote:equip:def:cloth-armor", map[string]any{
		"name":         "布甲",
		"slot":         "armor",
		"bonus_clicks": "1",
	}).Err(); err != nil {
		t.Fatalf("seed equipment definition: %v", err)
	}
	if err := store.setCurrentBoss(ctx, &Boss{
		ID:        "dragon-1",
		Name:      "火龙",
		Status:    bossStatusActive,
		MaxHP:     100,
		CurrentHP: 88,
		StartedAt: store.now().Unix(),
	}, []BossLootEntry{
		{ItemID: "cloth-armor", Weight: 1},
	}, nil); err != nil {
		t.Fatalf("set current boss: %v", err)
	}

	announcement, err := store.SaveAnnouncement(ctx, AnnouncementUpsert{
		Title:   "更新公告",
		Content: "公告正文",
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
		t.Fatalf("expected announcement version %q, got %+v", announcement.ID, snapshot)
	}

	encoded, err := sonic.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	jsonText := string(encoded)
	if strings.Contains(jsonText, "\"bossLoot\"") {
		t.Fatalf("expected slim snapshot without bossLoot, got %s", jsonText)
	}
	if strings.Contains(jsonText, "\"bossHeroLoot\"") {
		t.Fatalf("expected slim snapshot without bossHeroLoot, got %s", jsonText)
	}
	if strings.Contains(jsonText, "\"latestAnnouncement\"") {
		t.Fatalf("expected slim snapshot without latestAnnouncement, got %s", jsonText)
	}
}

func TestGetBossResourcesReturnsCurrentBossLootAndHeroLoot(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:equip:def:cloth-armor", map[string]any{
		"name":         "布甲",
		"slot":         "armor",
		"bonus_clicks": "1",
	}).Err(); err != nil {
		t.Fatalf("seed equipment definition: %v", err)
	}
	if err := store.SaveHeroDefinition(ctx, HeroDefinition{
		HeroID:      "slime",
		Name:        "史莱姆勇者",
		AwakenCap:   5,
		ImagePath:   "/hero/slime.png",
		ImageAlt:    "史莱姆勇者",
		BonusClicks: 2,
	}); err != nil {
		t.Fatalf("save hero definition: %v", err)
	}
	if err := store.setCurrentBoss(ctx, &Boss{
		ID:         "dragon-2",
		TemplateID: "dragon",
		Name:       "火龙",
		Status:     bossStatusActive,
		MaxHP:      200,
		CurrentHP:  120,
		StartedAt:  1713744000,
	}, []BossLootEntry{
		{ItemID: "cloth-armor", Weight: 3},
	}, []BossHeroLootEntry{
		{HeroID: "slime", Weight: 2},
	}); err != nil {
		t.Fatalf("set current boss: %v", err)
	}

	resources, err := store.GetBossResources(ctx)
	if err != nil {
		t.Fatalf("get boss resources: %v", err)
	}
	if resources.BossID != "dragon-2" || resources.TemplateID != "dragon" || resources.Status != bossStatusActive {
		t.Fatalf("unexpected boss resource meta: %+v", resources)
	}
	if len(resources.BossLoot) != 1 || resources.BossLoot[0].ItemID != "cloth-armor" {
		t.Fatalf("unexpected boss loot resources: %+v", resources.BossLoot)
	}
	if len(resources.BossHeroLoot) != 1 || resources.BossHeroLoot[0].HeroID != "slime" {
		t.Fatalf("unexpected boss hero loot resources: %+v", resources.BossHeroLoot)
	}
}

func TestFinalizeBossKillRewardsOnlyQualifiedParticipants(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.roll = func(int) int { return 0 }

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:equip:def:cloth-armor", map[string]any{
		"name":         "布甲",
		"slot":         "armor",
		"bonus_clicks": "1",
	}).Err(); err != nil {
		t.Fatalf("seed cloth-armor definition: %v", err)
	}

	boss := &Boss{
		ID:         "dragon-1",
		TemplateID: "dragon",
		Name:       "火龙",
		Status:     bossStatusDefeated,
		MaxHP:      1000,
		CurrentHP:  0,
		StartedAt:  1713744000,
		DefeatedAt: 1713744300,
	}
	if err := store.setCurrentBoss(ctx, boss, []BossLootEntry{
		{ItemID: "cloth-armor", Weight: 1},
	}, nil); err != nil {
		t.Fatalf("set current boss: %v", err)
	}
	if err := store.client.ZAdd(ctx, store.bossDamageKey(boss.ID),
		redis.Z{Score: 9, Member: "小红"},
		redis.Z{Score: 10, Member: "阿明"},
	).Err(); err != nil {
		t.Fatalf("seed boss damage: %v", err)
	}

	if _, err := store.finalizeBossKill(ctx, boss); err != nil {
		t.Fatalf("finalize boss kill: %v", err)
	}

	amingReward, err := store.lastRewardForNickname(ctx, "阿明")
	if err != nil {
		t.Fatalf("load 阿明 reward: %v", err)
	}
	if amingReward == nil || amingReward.ItemID != "cloth-armor" {
		t.Fatalf("expected 阿明 to receive reward, got %+v", amingReward)
	}

	xiaohongReward, err := store.lastRewardForNickname(ctx, "小红")
	if err != nil {
		t.Fatalf("load 小红 reward: %v", err)
	}
	if xiaohongReward != nil {
		t.Fatalf("expected 小红 to miss reward below 1%% threshold, got %+v", xiaohongReward)
	}
}

func TestGetUserStateIncludesAllRewardsFromBossSettlement(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.roll = func(int) int { return 0 }

	ctx := context.Background()
	if err := store.client.HSet(ctx, "vote:equip:def:cloth-armor", map[string]any{
		"name":         "布甲",
		"slot":         "armor",
		"bonus_clicks": "1",
	}).Err(); err != nil {
		t.Fatalf("seed cloth-armor definition: %v", err)
	}
	if err := store.SaveHeroDefinition(ctx, HeroDefinition{
		HeroID: "spark-cat",
		Name:   "星火猫",
	}); err != nil {
		t.Fatalf("save hero definition: %v", err)
	}

	boss := &Boss{
		ID:         "dragon-2",
		TemplateID: "dragon",
		Name:       "火龙",
		Status:     bossStatusDefeated,
		MaxHP:      100,
		CurrentHP:  0,
		StartedAt:  1713744000,
		DefeatedAt: 1713744300,
	}
	if err := store.setCurrentBoss(ctx, boss, []BossLootEntry{
		{ItemID: "cloth-armor", Weight: 1},
	}, []BossHeroLootEntry{
		{HeroID: "spark-cat", Weight: 1},
	}); err != nil {
		t.Fatalf("set current boss: %v", err)
	}
	if err := store.client.ZAdd(ctx, store.bossDamageKey(boss.ID),
		redis.Z{Score: 10, Member: "阿明"},
	).Err(); err != nil {
		t.Fatalf("seed boss damage: %v", err)
	}

	if _, err := store.finalizeBossKill(ctx, boss); err != nil {
		t.Fatalf("finalize boss kill: %v", err)
	}

	state, err := store.GetUserState(ctx, "阿明")
	if err != nil {
		t.Fatalf("get user state: %v", err)
	}

	if len(state.Inventory) != 1 || state.Inventory[0].ItemID != "cloth-armor" {
		t.Fatalf("expected equipment reward in inventory, got %+v", state.Inventory)
	}
	if len(state.Heroes) != 1 || state.Heroes[0].HeroID != "spark-cat" {
		t.Fatalf("expected hero reward in inventory, got %+v", state.Heroes)
	}
	if len(state.RecentRewards) != 2 {
		t.Fatalf("expected two visible rewards, got %+v", state.RecentRewards)
	}
	if state.RecentRewards[0].ItemID != "cloth-armor" || state.RecentRewards[1].ItemID != "spark-cat" {
		t.Fatalf("expected equipment and hero rewards in order, got %+v", state.RecentRewards)
	}
}

func TestGetStateIncludesActiveHeroBonuses(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.SaveHeroDefinition(ctx, HeroDefinition{
		HeroID:                     "spark-cat",
		Name:                       "星火猫",
		BonusClicks:                2,
		BonusCriticalChancePercent: 3,
		TraitType:                  HeroTraitFinalDamagePercent,
		TraitValue:                 50,
	}); err != nil {
		t.Fatalf("save hero definition: %v", err)
	}
	if err := store.client.HSet(ctx, store.heroInventoryKey("阿明"), map[string]any{
		"spark-cat": "1",
	}).Err(); err != nil {
		t.Fatalf("seed hero inventory: %v", err)
	}
	if err := store.client.Set(ctx, store.activeHeroKey("阿明"), "spark-cat", 0).Err(); err != nil {
		t.Fatalf("seed active hero: %v", err)
	}

	state, err := store.GetState(ctx, "阿明")
	if err != nil {
		t.Fatalf("get state: %v", err)
	}

	if len(state.Heroes) != 1 {
		t.Fatalf("expected 1 hero in inventory, got %+v", state.Heroes)
	}
	if state.ActiveHero == nil || state.ActiveHero.HeroID != "spark-cat" {
		t.Fatalf("expected spark-cat to be active, got %+v", state.ActiveHero)
	}
	if state.CombatStats.NormalDamage != 5 {
		t.Fatalf("expected hero bonuses to raise normal damage to 5, got %+v", state.CombatStats)
	}
	if state.CombatStats.CriticalChancePercent != 8 {
		t.Fatalf("expected hero bonuses to raise critical chance to 8, got %+v", state.CombatStats)
	}
}

func TestFinalizeBossKillAwardsQualifiedHeroLoot(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.roll = func(int) int { return 0 }

	ctx := context.Background()
	if err := store.SaveHeroDefinition(ctx, HeroDefinition{
		HeroID: "spark-cat",
		Name:   "星火猫",
	}); err != nil {
		t.Fatalf("save hero definition: %v", err)
	}

	boss := &Boss{
		ID:         "dragon-2",
		TemplateID: "dragon",
		Name:       "火龙",
		Status:     bossStatusDefeated,
		MaxHP:      100,
		CurrentHP:  0,
		StartedAt:  1713744000,
		DefeatedAt: 1713744300,
	}
	if err := store.setCurrentBoss(ctx, boss, []BossLootEntry{}, []BossHeroLootEntry{
		{HeroID: "spark-cat", Weight: 1},
	}); err != nil {
		t.Fatalf("set current boss: %v", err)
	}
	if err := store.client.ZAdd(ctx, store.bossDamageKey(boss.ID),
		redis.Z{Score: 1, Member: "阿明"},
		redis.Z{Score: 0.5, Member: "小红"},
	).Err(); err != nil {
		t.Fatalf("seed boss damage: %v", err)
	}

	if _, err := store.finalizeBossKill(ctx, boss); err != nil {
		t.Fatalf("finalize boss kill: %v", err)
	}

	quantities, err := store.heroInventoryQuantities(ctx, "阿明")
	if err != nil {
		t.Fatalf("load 阿明 hero inventory: %v", err)
	}
	if quantities["spark-cat"] != 1 {
		t.Fatalf("expected 阿明 to receive hero reward, got %+v", quantities)
	}

	xiaohongQuantities, err := store.heroInventoryQuantities(ctx, "小红")
	if err != nil {
		t.Fatalf("load 小红 hero inventory: %v", err)
	}
	if len(xiaohongQuantities) != 0 {
		t.Fatalf("expected 小红 to miss hero reward below threshold, got %+v", xiaohongQuantities)
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
