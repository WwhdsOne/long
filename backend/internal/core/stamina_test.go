package core

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/bytedance/sonic"
)

func TestClickButtonConsumesStaminaAndLocksDamageAfterExhausted(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	baseTime := time.Unix(1_700_000_000, 0)
	store.now = func() time.Time { return baseTime }

	seedActiveBossForStaminaTest(t, store, ctx, "1")

	for i := range 2500 {
		result, err := store.ClickButton(ctx, "boss-part:1-0", nickname, 0)
		if err != nil {
			t.Fatalf("click %d failed: %v", i+1, err)
		}
		if result.BossDamage <= 1 {
			t.Fatalf("expected normal stamina click to deal more than 1 damage before exhausted, got %d at click %d", result.BossDamage, i+1)
		}
	}

	state, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get user state after exhausting stamina: %v", err)
	}
	if state.Stamina.Current != 0 {
		t.Fatalf("expected stamina to be 0 after 2500 clicks, got %+v", state.Stamina)
	}

	result, err := store.ClickButton(ctx, "boss-part:1-0", nickname, 0)
	if err != nil {
		t.Fatalf("click after exhausting stamina failed: %v", err)
	}
	if result.BossDamage != 1 {
		t.Fatalf("expected exhausted stamina click damage to be fixed at 1, got %d", result.BossDamage)
	}
}

func TestGetUserStateRecoversStaminaOverTime(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	baseTime := time.Unix(1_700_000_000, 0)
	store.now = func() time.Time { return baseTime }

	if err := store.saveStaminaSnapshot(ctx, nickname, staminaSnapshot{
		Current:       0,
		MaxLevel:      0,
		Max:           50,
		ClickProgress: 0,
		LastRecoverAt: baseTime.Unix(),
		ZeroAt:        baseTime.Unix(),
	}); err != nil {
		t.Fatalf("seed stamina: %v", err)
	}

	store.now = func() time.Time { return baseTime.Add(staminaRecoverInterval * 3) }
	state, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get recovered user state: %v", err)
	}
	if state.Stamina.Current != 3 {
		t.Fatalf("expected recovered stamina 3, got %+v", state.Stamina)
	}
	if state.Stamina.ZeroAt != 0 {
		t.Fatalf("expected zero timestamp to clear after recovery, got %+v", state.Stamina)
	}

	store.now = func() time.Time { return baseTime.Add(staminaFullRecoverDuration) }
	state, err = store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get full recovered user state: %v", err)
	}
	if state.Stamina.Current != state.Stamina.Max || state.Stamina.Max != 50 {
		t.Fatalf("expected full stamina after full recover duration, got %+v", state.Stamina)
	}
}

func TestPurchaseStaminaFullTriggersRiskBanWithinOneSecond(t *testing.T) {
	store, _, _, cleanup := newShopTestStore(t, nil)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	baseTime := time.Unix(1_700_000_000, 0)
	store.now = func() time.Time { return baseTime }

	if err := store.client.HSet(ctx, store.resourceKey(nickname), "gold", "500000").Err(); err != nil {
		t.Fatalf("seed gold: %v", err)
	}
	if err := store.saveStaminaSnapshot(ctx, nickname, staminaSnapshot{
		Current:       0,
		MaxLevel:      0,
		Max:           50,
		ClickProgress: 0,
		LastRecoverAt: baseTime.Unix(),
		ZeroAt:        baseTime.Unix(),
	}); err != nil {
		t.Fatalf("seed stamina: %v", err)
	}

	state, err := store.PurchaseStaminaFull(ctx, nickname)
	if err != nil {
		t.Fatalf("purchase stamina full: %v", err)
	}
	if state.Stamina.Current != 50 {
		t.Fatalf("expected stamina to refill to 50, got %+v", state.Stamina)
	}
	if state.Stamina.RiskBanUntil != baseTime.Add(8*time.Hour).Unix() {
		t.Fatalf("expected 8 hour risk ban, got %+v", state.Stamina)
	}

	_, err = store.PurchaseStaminaFull(ctx, nickname)
	if err != ErrStaminaRiskBanned {
		t.Fatalf("expected banned account to reject refill, got %v", err)
	}
}

func TestPurchaseStaminaFullRiskBanEscalatesPermanently(t *testing.T) {
	store, _, logStore, cleanup := newShopTestStore(t, nil)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	baseTime := time.Unix(1_700_000_000, 0)

	for attempt, expectedDuration := range []time.Duration{
		8 * time.Hour,
		24 * time.Hour,
		72 * time.Hour,
	} {
		now := baseTime.Add(time.Duration(attempt) * 100 * time.Hour)
		store.now = func() time.Time { return now }

		if err := store.client.HSet(ctx, store.resourceKey(nickname), "gold", "999999999").Err(); err != nil {
			t.Fatalf("seed gold for attempt %d: %v", attempt+1, err)
		}
		if err := store.saveStaminaSnapshot(ctx, nickname, staminaSnapshot{
			Current:       0,
			MaxLevel:      0,
			Max:           50,
			ClickProgress: 0,
			LastRecoverAt: now.Unix(),
			ZeroAt:        now.Unix(),
		}); err != nil {
			t.Fatalf("seed stamina for attempt %d: %v", attempt+1, err)
		}
		_ = store.client.Del(ctx, store.staminaRiskBanKey(nickname)).Err()

		state, err := store.PurchaseStaminaFull(ctx, nickname)
		if err != nil {
			t.Fatalf("purchase attempt %d failed: %v", attempt+1, err)
		}
		expectedUntil := now.Add(expectedDuration).Unix()
		if state.Stamina.RiskBanUntil != expectedUntil {
			t.Fatalf("attempt %d expected ban until %d, got %+v", attempt+1, expectedUntil, state.Stamina)
		}
		strikeCount, err := store.client.Get(ctx, store.staminaRiskBanStrikeKey(nickname)).Int64()
		if err != nil {
			t.Fatalf("load strike count attempt %d: %v", attempt+1, err)
		}
		if strikeCount != int64(attempt+1) {
			t.Fatalf("attempt %d expected strike count %d, got %d", attempt+1, attempt+1, strikeCount)
		}
	}

	if len(logStore.staminaLogs) != 3 {
		t.Fatalf("expected 3 stamina logs, got %+v", logStore.staminaLogs)
	}
	if logStore.staminaLogs[2].RiskBanStrike != 3 {
		t.Fatalf("expected third stamina log strike=3, got %+v", logStore.staminaLogs[2])
	}
	if logStore.staminaLogs[2].RiskBanDurationSec != int64((72 * time.Hour).Seconds()) {
		t.Fatalf("expected third stamina log duration 72h, got %+v", logStore.staminaLogs[2])
	}
}

func TestPurchaseStaminaFullWritesAttemptLogWithSecondsSinceZero(t *testing.T) {
	store, _, logStore, cleanup := newShopTestStore(t, nil)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	baseTime := time.Unix(1_700_000_000, 0)
	store.now = func() time.Time { return baseTime }

	if err := store.client.HSet(ctx, store.resourceKey(nickname), "gold", "500000").Err(); err != nil {
		t.Fatalf("seed gold: %v", err)
	}
	if err := store.saveStaminaSnapshot(ctx, nickname, staminaSnapshot{
		Current:       0,
		MaxLevel:      0,
		Max:           50,
		ClickProgress: 0,
		LastRecoverAt: baseTime.Unix(),
		ZeroAt:        baseTime.Add(-3 * time.Second).Unix(),
	}); err != nil {
		t.Fatalf("seed stamina: %v", err)
	}

	if _, err := store.PurchaseStaminaFull(ctx, nickname); err != nil {
		t.Fatalf("purchase stamina full: %v", err)
	}
	if len(logStore.staminaLogs) != 1 {
		t.Fatalf("expected one stamina log, got %+v", logStore.staminaLogs)
	}
	entry := logStore.staminaLogs[0]
	if entry.SecondsSinceZero != 3 {
		t.Fatalf("expected secondsSinceZero=3, got %+v", entry)
	}
	if !entry.Succeeded || entry.FailureReason != "" {
		t.Fatalf("expected successful attempt log, got %+v", entry)
	}
}

func TestPurchaseStaminaFullPastWindowDoesNotTriggerRiskBan(t *testing.T) {
	store, _, _, cleanup := newShopTestStore(t, nil)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	baseTime := time.Unix(1_700_000_000, 0)
	store.now = func() time.Time { return baseTime }

	if err := store.client.HSet(ctx, store.resourceKey(nickname), "gold", "500000").Err(); err != nil {
		t.Fatalf("seed gold: %v", err)
	}
	if err := store.saveStaminaSnapshot(ctx, nickname, staminaSnapshot{
		Current:       0,
		MaxLevel:      0,
		Max:           50,
		ClickProgress: 0,
		LastRecoverAt: baseTime.Unix(),
		ZeroAt:        baseTime.Add(-2 * time.Second).Unix(),
	}); err != nil {
		t.Fatalf("seed stamina: %v", err)
	}

	state, err := store.PurchaseStaminaFull(ctx, nickname)
	if err != nil {
		t.Fatalf("purchase stamina full: %v", err)
	}
	if state.Stamina.RiskBanUntil != 0 {
		t.Fatalf("expected no risk ban after 1 second window, got %+v", state.Stamina)
	}
}

func TestUpgradeStaminaCapUsesTieredCostAndCapsAt100(t *testing.T) {
	store, _, _, cleanup := newShopTestStore(t, nil)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	baseTime := time.Unix(1_700_000_000, 0)
	store.now = func() time.Time { return baseTime }

	if err := store.client.HSet(ctx, store.resourceKey(nickname), "gold", "999999999").Err(); err != nil {
		t.Fatalf("seed gold: %v", err)
	}

	for i := range staminaMaxUpgradeLevels {
		if _, err := store.UpgradeStaminaCap(ctx, nickname); err != nil {
			t.Fatalf("upgrade %d failed: %v", i+1, err)
		}
	}

	state, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get upgraded state: %v", err)
	}
	if state.Stamina.Max != 100 || state.Stamina.MaxLevel != staminaMaxUpgradeLevels {
		t.Fatalf("expected stamina cap to reach 100, got %+v", state.Stamina)
	}

	_, err = store.UpgradeStaminaCap(ctx, nickname)
	if err != ErrStaminaMaxLevelReached {
		t.Fatalf("expected max level error, got %v", err)
	}
}

func TestPurchaseStaminaFullPriceResetsNextDay(t *testing.T) {
	store, _, _, cleanup := newShopTestStore(t, nil)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	baseTime := time.Unix(1_700_000_000, 0)
	store.now = func() time.Time { return baseTime }

	if err := store.client.HSet(ctx, store.resourceKey(nickname), "gold", "1000000000").Err(); err != nil {
		t.Fatalf("seed gold: %v", err)
	}
	if err := store.saveStaminaSnapshot(ctx, nickname, staminaSnapshot{
		Current:       1,
		MaxLevel:      0,
		Max:           50,
		ClickProgress: 0,
		LastRecoverAt: baseTime.Unix(),
	}); err != nil {
		t.Fatalf("seed stamina: %v", err)
	}

	if _, err := store.PurchaseStaminaFull(ctx, nickname); err != nil {
		t.Fatalf("first day purchase: %v", err)
	}

	store.now = func() time.Time { return baseTime.Add(24 * time.Hour) }
	state, err := store.GetUserState(ctx, nickname)
	if err != nil {
		t.Fatalf("get next day state: %v", err)
	}
	if state.Stamina.DailyFullBuyCount != 0 {
		t.Fatalf("expected next day count reset, got %+v", state.Stamina)
	}
	if state.Stamina.NextFullBuyPrice != staminaFirstFullBuyPrice {
		t.Fatalf("expected next day price reset to first price, got %+v", state.Stamina)
	}
}

func seedActiveBossForStaminaTest(t *testing.T, store *Store, ctx context.Context, roomID string) {
	t.Helper()

	boss := Boss{
		ID:          "boss-stamina",
		Name:        "体力木桩",
		Status:      bossStatusActive,
		RoomID:      roomID,
		MaxHP:       1_000_000,
		CurrentHP:   1_000_000,
		GoldOnKill:  1,
		StoneOnKill: 1,
		Parts: []BossPart{{
			X:         1,
			Y:         0,
			Type:      PartTypeSoft,
			MaxHP:     1_000_000,
			CurrentHP: 1_000_000,
			Armor:     0,
			Alive:     true,
		}},
	}
	partsRaw, err := sonic.Marshal(boss.Parts)
	if err != nil {
		t.Fatalf("marshal boss parts: %v", err)
	}
	if err := store.client.HSet(ctx, store.bossCurrentKeyForRoom(roomID), map[string]any{
		"id":                    boss.ID,
		"room_id":               boss.RoomID,
		"queue_id":              store.queueIDForRoom(roomID),
		"name":                  boss.Name,
		"status":                boss.Status,
		"max_hp":                strconv.FormatInt(boss.MaxHP, 10),
		"current_hp":            strconv.FormatInt(boss.CurrentHP, 10),
		"gold_on_kill":          strconv.FormatInt(boss.GoldOnKill, 10),
		"stone_on_kill":         strconv.FormatInt(boss.StoneOnKill, 10),
		"talent_points_on_kill": "0",
		"parts":                 string(partsRaw),
	}).Err(); err != nil {
		t.Fatalf("seed boss: %v", err)
	}
}
