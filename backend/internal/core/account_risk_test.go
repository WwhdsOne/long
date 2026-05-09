package core

import (
	"context"
	"sync"
	"testing"
	"time"
)

type stubAccountRiskEventLogStore struct {
	mu   sync.Mutex
	logs []AccountRiskEventLog
}

func (s *stubAccountRiskEventLogStore) WriteAccountRiskEvent(_ context.Context, item AccountRiskEventLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = append(s.logs, item)
	return nil
}

func (s *stubAccountRiskEventLogStore) snapshot() []AccountRiskEventLog {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]AccountRiskEventLog(nil), s.logs...)
}

func TestAccountRiskThresholdsAndBanDurations(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	baseTime := time.Unix(1_700_000_000, 0)
	store.now = func() time.Time { return baseTime }

	state, err := store.RecordAccountRiskEvent(ctx, nickname, AccountRiskEventLoginTurnstileInvalid)
	if err != nil {
		t.Fatalf("record first risk event: %v", err)
	}
	if state.Score != 3 || state.BanUntil != 0 {
		t.Fatalf("expected score 3 without ban, got %+v", state)
	}

	state, err = store.RecordAccountRiskEvent(ctx, nickname, AccountRiskEventLoginTurnstileInvalid)
	if err != nil {
		t.Fatalf("record second risk event: %v", err)
	}
	if state.Score != 6 || state.BanUntil != baseTime.Add(8*time.Hour).Unix() {
		t.Fatalf("expected 6 points with 8h ban, got %+v", state)
	}

	state, err = store.RecordAccountRiskEvent(ctx, nickname, AccountRiskEventPostStaminaPurchaseClick)
	if err != nil {
		t.Fatalf("record third risk event: %v", err)
	}
	if state.Score != 10 || state.BanUntil != baseTime.Add(24*time.Hour).Unix() {
		t.Fatalf("expected 10 points with 24h ban, got %+v", state)
	}

	state, err = store.RecordAccountRiskEvent(ctx, nickname, AccountRiskEventPostStaminaPurchaseClick)
	if err != nil {
		t.Fatalf("record fourth risk event: %v", err)
	}
	if state.Score != 14 || state.BanUntil != baseTime.Add(72*time.Hour).Unix() {
		t.Fatalf("expected 14 points with 72h ban, got %+v", state)
	}
}

func TestAccountRiskWindowExpiresOldEvents(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	baseTime := time.Unix(1_700_000_000, 0)
	store.now = func() time.Time { return baseTime }

	if _, err := store.RecordAccountRiskEvent(ctx, nickname, AccountRiskEventLoginTurnstileInvalid); err != nil {
		t.Fatalf("record first event: %v", err)
	}
	store.now = func() time.Time { return baseTime.Add(23 * time.Hour) }
	if _, err := store.RecordAccountRiskEvent(ctx, nickname, AccountRiskEventClickRateLimitHit); err != nil {
		t.Fatalf("record second event: %v", err)
	}

	store.now = func() time.Time { return baseTime.Add(25 * time.Hour) }
	state, err := store.GetAccountRiskState(ctx, nickname)
	if err != nil {
		t.Fatalf("get account risk state: %v", err)
	}
	if state.Score != 2 {
		t.Fatalf("expected only latest 2 points to remain, got %+v", state)
	}
}

func TestAccountRiskBanUsesMaxRule(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	baseTime := time.Unix(1_700_000_000, 0)
	store.now = func() time.Time { return baseTime }

	if _, err := store.RecordAccountRiskEvent(ctx, nickname, AccountRiskEventLoginTurnstileInvalid); err != nil {
		t.Fatalf("record risk event: %v", err)
	}
	state, err := store.RecordAccountRiskEvent(ctx, nickname, AccountRiskEventLoginTurnstileInvalid)
	if err != nil {
		t.Fatalf("record risk event: %v", err)
	}
	if state.BanUntil != baseTime.Add(8*time.Hour).Unix() {
		t.Fatalf("expected initial 8h ban, got %+v", state)
	}

	store.now = func() time.Time { return baseTime.Add(2 * time.Hour) }
	state, err = store.RecordAccountRiskEvent(ctx, nickname, AccountRiskEventClickRateLimitHit)
	if err != nil {
		t.Fatalf("record later risk event: %v", err)
	}
	if state.Score != 8 {
		t.Fatalf("expected score 8, got %+v", state)
	}
	if state.BanUntil != baseTime.Add(8*time.Hour).Unix() {
		t.Fatalf("expected same tier not to extend ban, got %+v", state)
	}

	state, err = store.RecordAccountRiskEvent(ctx, nickname, AccountRiskEventPostStaminaPurchaseClick)
	if err != nil {
		t.Fatalf("record tier-up risk event: %v", err)
	}
	if state.Score != 12 {
		t.Fatalf("expected score 12, got %+v", state)
	}
	if state.BanUntil != baseTime.Add(24*time.Hour).Unix() {
		t.Fatalf("expected tier-up to extend ban from original start, got %+v", state)
	}
}

func TestAccountRiskBanReadDoesNotExtend(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	baseTime := time.Unix(1_700_000_000, 0)
	store.now = func() time.Time { return baseTime }

	if _, err := store.RecordAccountRiskEvent(ctx, nickname, AccountRiskEventLoginTurnstileInvalid); err != nil {
		t.Fatalf("record first risk event: %v", err)
	}
	state, err := store.RecordAccountRiskEvent(ctx, nickname, AccountRiskEventLoginTurnstileInvalid)
	if err != nil {
		t.Fatalf("record second risk event: %v", err)
	}
	initialBanUntil := baseTime.Add(8 * time.Hour).Unix()
	if state.BanUntil != initialBanUntil {
		t.Fatalf("expected initial 8h ban, got %+v", state)
	}

	store.now = func() time.Time { return baseTime.Add(2 * time.Hour) }
	state, err = store.GetAccountRiskState(ctx, nickname)
	if err != nil {
		t.Fatalf("get account risk state: %v", err)
	}
	if state.BanUntil != initialBanUntil {
		t.Fatalf("expected read not to extend ban, got %+v", state)
	}
}

func TestListAndClearAccountRiskEntries(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	baseTime := time.Unix(1_700_000_000, 0)
	store.now = func() time.Time { return baseTime }

	if _, err := store.RecordAccountRiskEvent(ctx, "高分账号", AccountRiskEventPostStaminaPurchaseClick); err != nil {
		t.Fatalf("record high score risk event: %v", err)
	}
	if _, err := store.RecordAccountRiskEvent(ctx, "普通账号", AccountRiskEventClickRateLimitHit); err != nil {
		t.Fatalf("record normal score risk event: %v", err)
	}

	entries, err := store.ListAccountRiskEntries(ctx)
	if err != nil {
		t.Fatalf("list account risk entries: %v", err)
	}
	if len(entries) != 2 || entries[0].Nickname != "高分账号" {
		t.Fatalf("expected entries sorted by score, got %+v", entries)
	}

	if err := store.ClearAccountRiskState(ctx, "高分账号"); err != nil {
		t.Fatalf("clear account risk state: %v", err)
	}
	clearedState, err := store.GetAccountRiskState(ctx, "高分账号")
	if err != nil {
		t.Fatalf("get cleared risk state: %v", err)
	}
	if clearedState.Score != 0 || clearedState.BanUntil != 0 {
		t.Fatalf("expected cleared state, got %+v", clearedState)
	}
}

func TestManualClickAfterRecentStaminaPurchaseAddsRiskButStillSucceeds(t *testing.T) {
	store, _, _, cleanup := newShopTestStore(t, nil)
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	baseTime := time.Unix(1_700_000_000, 0)
	store.now = func() time.Time { return baseTime }
	seedActiveBossForStaminaTest(t, store, ctx, "1")

	if err := store.MarkStaminaPurchase(ctx, nickname); err != nil {
		t.Fatalf("mark stamina purchase: %v", err)
	}
	result, err := store.ClickButton(ctx, "boss-part:1-0", nickname, 0)
	if err != nil {
		t.Fatalf("click after stamina purchase: %v", err)
	}
	if result.BossDamage <= 0 {
		t.Fatalf("expected click success, got %+v", result)
	}
	state, err := store.GetAccountRiskState(ctx, nickname)
	if err != nil {
		t.Fatalf("get account risk state: %v", err)
	}
	if state.Score != store.accountRiskConfig.Points.PostStaminaPurchaseClick {
		t.Fatalf("expected post purchase click points, got %+v", state)
	}
}

func TestRecordAccountRiskEventWritesHistoryLog(t *testing.T) {
	baseStore, cleanup := newTestStore(t)
	defer cleanup()

	logStore := &stubAccountRiskEventLogStore{}
	store := NewStore(baseStore.client, "hai-world:", StoreOptions{
		CriticalChancePercent: 5,
		AccountRisk: AccountRiskConfig{
			PurchaseClickCooldown: 3 * time.Second,
		},
		AccountRiskEventLogStore: logStore,
	}, baseStore.validator)

	ctx := context.Background()
	baseTime := time.Unix(1_700_000_000, 0)
	store.now = func() time.Time { return baseTime }

	state, err := store.RecordAccountRiskEvent(ctx, "阿明", AccountRiskEventLoginTurnstileInvalid)
	if err != nil {
		t.Fatalf("record account risk event: %v", err)
	}
	logs := logStore.snapshot()
	if len(logs) != 1 {
		t.Fatalf("expected one risk history log, got %+v", logs)
	}
	if logs[0].Nickname != "阿明" || logs[0].EventType != string(AccountRiskEventLoginTurnstileInvalid) {
		t.Fatalf("unexpected risk history log: %+v", logs[0])
	}
	if logs[0].Points != store.accountRiskConfig.Points.LoginTurnstileInvalid {
		t.Fatalf("unexpected risk points: %+v", logs[0])
	}
	if logs[0].ScoreAfter != state.Score || logs[0].BanUntilAfter != state.BanUntil || logs[0].CreatedAt != baseTime.Unix() {
		t.Fatalf("unexpected risk history state: %+v, state=%+v", logs[0], state)
	}
}

func TestClearAccountRiskStateWritesHistoryLog(t *testing.T) {
	baseStore, cleanup := newTestStore(t)
	defer cleanup()

	logStore := &stubAccountRiskEventLogStore{}
	store := NewStore(baseStore.client, "hai-world:", StoreOptions{
		CriticalChancePercent: 5,
		AccountRisk: AccountRiskConfig{
			PurchaseClickCooldown: 3 * time.Second,
		},
		AccountRiskEventLogStore: logStore,
	}, baseStore.validator)

	ctx := context.Background()
	baseTime := time.Unix(1_700_000_000, 0)
	store.now = func() time.Time { return baseTime }

	if _, err := store.RecordAccountRiskEvent(ctx, "阿明", AccountRiskEventLoginTurnstileInvalid); err != nil {
		t.Fatalf("record account risk event: %v", err)
	}
	if err := store.ClearAccountRiskState(ctx, "阿明"); err != nil {
		t.Fatalf("clear account risk state: %v", err)
	}
	logs := logStore.snapshot()
	if len(logs) != 2 {
		t.Fatalf("expected event log and clear log, got %+v", logs)
	}
	clearLog := logs[1]
	if clearLog.Nickname != "阿明" || clearLog.EventType != "risk_state_cleared" {
		t.Fatalf("unexpected clear risk history log: %+v", clearLog)
	}
	if clearLog.Points != 0 || clearLog.ScoreAfter != 0 || clearLog.BanUntilAfter != 0 || clearLog.CreatedAt != baseTime.Unix() {
		t.Fatalf("unexpected clear risk history payload: %+v", clearLog)
	}
}
