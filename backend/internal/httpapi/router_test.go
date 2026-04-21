package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bytedance/sonic"

	"long/internal/admin"
	ossupload "long/internal/oss"
	"long/internal/vote"
)

type mockStore struct {
	state                  vote.State
	equipState             vote.State
	adminState             vote.AdminState
	adminPlayerPage        vote.AdminPlayerPage
	adminPlayer            *vote.AdminPlayerOverview
	bossHistory            []vote.BossHistoryEntry
	announcements          []vote.Announcement
	latestAnnouncement     *vote.Announcement
	messagePage            vote.MessagePage
	result                 vote.ClickResult
	lastButton             vote.ButtonUpsert
	lastBoss               vote.BossUpsert
	lastBossTemplate       vote.BossTemplateUpsert
	lastHero               vote.HeroDefinition
	lastTemplateLootID     string
	lastTemplateLoot       []vote.BossLootEntry
	lastTemplateHeroLootID string
	lastTemplateHeroLoot   []vote.BossHeroLootEntry
	lastCycleEnabled       bool
	lastSalvageItemID      string
	lastSalvageQuantity    int64
	lastReforgeItemID      string
	lastAwakenHeroID       string
	lastPurchasedCosmetic  string
	lastCosmeticLoadout    vote.CosmeticLoadout
	getStateErr            error
	clickErr               error
	equipErr               error
	validateErr            error
	messageErr             error
	synthesizeErr          error
	salvageErr             error
	reforgeErr             error
	awakenErr              error
	purchaseErr            error
	cosmeticEquipErr       error
}

func (m *mockStore) GetState(_ context.Context, nickname string) (vote.State, error) {
	if m.getStateErr != nil {
		return vote.State{}, m.getStateErr
	}
	state := m.state
	if nickname == "" {
		state.UserStats = nil
	}
	return state, nil
}

func (m *mockStore) GetSnapshot(_ context.Context) (vote.Snapshot, error) {
	return vote.Snapshot{
		Buttons:     m.state.Buttons,
		Leaderboard: m.state.Leaderboard,
	}, nil
}

func (m *mockStore) GetUserState(_ context.Context, nickname string) (vote.UserState, error) {
	if m.getStateErr != nil {
		return vote.UserState{}, m.getStateErr
	}
	userState := vote.UserState{
		Inventory:   []vote.InventoryItem{},
		Loadout:     vote.Loadout{},
		CombatStats: vote.CombatStats{},
	}
	if nickname == "" {
		return userState, nil
	}

	userState.UserStats = m.state.UserStats
	userState.MyBossStats = m.state.MyBossStats
	userState.Inventory = m.state.Inventory
	userState.Loadout = m.state.Loadout
	userState.CombatStats = m.state.CombatStats
	userState.LastReward = m.state.LastReward
	return userState, nil
}

func (m *mockStore) ClickButton(_ context.Context, slug string, nickname string) (vote.ClickResult, error) {
	if m.clickErr != nil {
		return vote.ClickResult{}, m.clickErr
	}
	for index := range m.state.Buttons {
		if m.state.Buttons[index].Key == slug {
			if m.result.Button.Key == "" {
				m.state.Buttons[index].Count++
				if m.state.UserStats == nil && nickname != "" {
					m.state.UserStats = &vote.UserStats{Nickname: nickname}
				}
				if m.state.UserStats != nil {
					m.state.UserStats.ClickCount++
				}
				return vote.ClickResult{
					Button:   m.state.Buttons[index],
					Delta:    1,
					Critical: false,
					UserStats: vote.UserStats{
						Nickname:   nickname,
						ClickCount: 1,
					},
				}, nil
			}
			m.state.Buttons[index].Count = m.result.Button.Count
			return m.result, nil
		}
	}
	return vote.ClickResult{}, vote.ErrButtonNotFound
}

func (m *mockStore) ValidateNickname(_ context.Context, _ string) error {
	return m.validateErr
}

func (m *mockStore) EquipItem(_ context.Context, _ string, _ string) (vote.State, error) {
	if m.equipErr != nil {
		return vote.State{}, m.equipErr
	}
	if len(m.equipState.Buttons) == 0 {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) UnequipItem(_ context.Context, _ string, _ string) (vote.State, error) {
	if m.equipErr != nil {
		return vote.State{}, m.equipErr
	}
	if len(m.equipState.Buttons) == 0 {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) GetAdminState(_ context.Context) (vote.AdminState, error) {
	return m.adminState, nil
}

func (m *mockStore) ListAdminPlayers(_ context.Context, _ string, _ int64) (vote.AdminPlayerPage, error) {
	return m.adminPlayerPage, nil
}

func (m *mockStore) GetAdminPlayer(_ context.Context, _ string) (*vote.AdminPlayerOverview, error) {
	return m.adminPlayer, nil
}

func (m *mockStore) SaveButton(_ context.Context, button vote.ButtonUpsert) error {
	m.lastButton = button
	return nil
}

func (m *mockStore) SaveEquipmentDefinition(_ context.Context, _ vote.EquipmentDefinition) error {
	return nil
}

func (m *mockStore) SaveHeroDefinition(_ context.Context, hero vote.HeroDefinition) error {
	m.lastHero = hero
	return nil
}

func (m *mockStore) DeleteHeroDefinition(_ context.Context, _ string) error {
	return nil
}

func (m *mockStore) DeleteEquipmentDefinition(_ context.Context, _ string) error {
	return nil
}

func (m *mockStore) ActivateBoss(_ context.Context, boss vote.BossUpsert) (*vote.Boss, error) {
	m.lastBoss = boss
	return &vote.Boss{
		ID:        boss.ID,
		Name:      boss.Name,
		Status:    "active",
		MaxHP:     boss.MaxHP,
		CurrentHP: boss.MaxHP,
	}, nil
}

func (m *mockStore) DeactivateBoss(_ context.Context) error {
	return nil
}

func (m *mockStore) SetBossLoot(_ context.Context, _ string, _ []vote.BossLootEntry) error {
	return nil
}

func (m *mockStore) SaveBossTemplate(_ context.Context, template vote.BossTemplateUpsert) error {
	m.lastBossTemplate = template
	return nil
}

func (m *mockStore) DeleteBossTemplate(_ context.Context, _ string) error {
	return nil
}

func (m *mockStore) SetBossTemplateLoot(_ context.Context, templateID string, loot []vote.BossLootEntry) error {
	m.lastTemplateLootID = templateID
	m.lastTemplateLoot = loot
	return nil
}

func (m *mockStore) SetBossTemplateHeroLoot(_ context.Context, templateID string, loot []vote.BossHeroLootEntry) error {
	m.lastTemplateHeroLootID = templateID
	m.lastTemplateHeroLoot = loot
	return nil
}

func (m *mockStore) SetBossCycleEnabled(_ context.Context, enabled bool) (*vote.Boss, error) {
	m.lastCycleEnabled = enabled
	if !enabled {
		return nil, nil
	}
	return &vote.Boss{
		ID:         "dragon-1",
		TemplateID: "dragon",
		Name:       "火龙",
		Status:     "active",
		MaxHP:      80,
		CurrentHP:  80,
	}, nil
}

func (m *mockStore) ListBossHistory(_ context.Context) ([]vote.BossHistoryEntry, error) {
	return m.bossHistory, nil
}

func (m *mockStore) GetLatestAnnouncement(_ context.Context) (*vote.Announcement, error) {
	return m.latestAnnouncement, nil
}

func (m *mockStore) ListAnnouncements(_ context.Context, includeInactive bool) ([]vote.Announcement, error) {
	return m.announcements, nil
}

func (m *mockStore) SaveAnnouncement(_ context.Context, announcement vote.AnnouncementUpsert) (*vote.Announcement, error) {
	return &vote.Announcement{
		ID:          "1",
		Title:       announcement.Title,
		Content:     announcement.Content,
		PublishedAt: 1710000000,
		Active:      announcement.Active,
	}, nil
}

func (m *mockStore) DeleteAnnouncement(_ context.Context, _ string) error {
	return nil
}

func (m *mockStore) CreateMessage(_ context.Context, nickname string, content string) (*vote.Message, error) {
	if m.messageErr != nil {
		return nil, m.messageErr
	}
	return &vote.Message{
		ID:        "1",
		Nickname:  nickname,
		Content:   content,
		CreatedAt: 1710000000,
	}, nil
}

func (m *mockStore) ListMessages(_ context.Context, _ string, _ int64) (vote.MessagePage, error) {
	return m.messagePage, m.messageErr
}

func (m *mockStore) DeleteMessage(_ context.Context, _ string) error {
	return nil
}

func (m *mockStore) SynthesizeItem(_ context.Context, _ string, _ string) (vote.State, error) {
	if m.synthesizeErr != nil {
		return vote.State{}, m.synthesizeErr
	}
	if len(m.equipState.Buttons) == 0 && len(m.equipState.Inventory) == 0 && m.equipState.Loadout.Weapon == nil && m.equipState.Loadout.Armor == nil && m.equipState.Loadout.Accessory == nil {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) SalvageEquipment(_ context.Context, _ string, itemID string, quantity int64) (vote.State, error) {
	if m.salvageErr != nil {
		return vote.State{}, m.salvageErr
	}
	m.lastSalvageItemID = itemID
	m.lastSalvageQuantity = quantity
	if len(m.equipState.Inventory) == 0 && m.equipState.Gems == 0 && m.equipState.LastForgeResult == nil {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) ReforgeEquipment(_ context.Context, _ string, itemID string) (vote.State, error) {
	if m.reforgeErr != nil {
		return vote.State{}, m.reforgeErr
	}
	m.lastReforgeItemID = itemID
	if len(m.equipState.Inventory) == 0 && m.equipState.Gems == 0 && m.equipState.LastForgeResult == nil {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) EquipHero(_ context.Context, _ string, _ string) (vote.State, error) {
	if len(m.equipState.Buttons) == 0 && len(m.equipState.Heroes) == 0 && m.equipState.ActiveHero == nil {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) AwakenHero(_ context.Context, _ string, heroID string) (vote.State, error) {
	if m.awakenErr != nil {
		return vote.State{}, m.awakenErr
	}
	m.lastAwakenHeroID = heroID
	if len(m.equipState.Heroes) == 0 && m.equipState.Gems == 0 && m.equipState.LastForgeResult == nil {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) SalvageHero(_ context.Context, _ string, heroID string, quantity int64) (vote.State, error) {
	if m.salvageErr != nil {
		return vote.State{}, m.salvageErr
	}
	m.lastAwakenHeroID = heroID
	m.lastSalvageQuantity = quantity
	if len(m.equipState.Heroes) == 0 && m.equipState.Gems == 0 && m.equipState.LastForgeResult == nil {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) PurchaseCosmetic(_ context.Context, _ string, cosmeticID string) (vote.State, error) {
	if m.purchaseErr != nil {
		return vote.State{}, m.purchaseErr
	}
	m.lastPurchasedCosmetic = cosmeticID
	if len(m.equipState.ShopCatalog) == 0 && m.equipState.Gems == 0 {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) EquipCosmetics(_ context.Context, _ string, trailID string, impactID string) (vote.State, error) {
	if m.cosmeticEquipErr != nil {
		return vote.State{}, m.cosmeticEquipErr
	}
	m.lastCosmeticLoadout = vote.CosmeticLoadout{
		TrailID:  trailID,
		ImpactID: impactID,
	}
	if len(m.equipState.ShopCatalog) == 0 && m.equipState.EquippedCosmetics == (vote.CosmeticLoadout{}) {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) UnequipHero(_ context.Context, _ string, _ string) (vote.State, error) {
	if len(m.equipState.Buttons) == 0 && len(m.equipState.Heroes) == 0 && m.equipState.ActiveHero == nil {
		return m.state, nil
	}
	return m.equipState, nil
}

type mockOSSSigner struct {
	policy ossupload.Policy
	err    error
}

func (m *mockOSSSigner) CreatePolicy(_ context.Context) (ossupload.Policy, error) {
	return m.policy, m.err
}

type mockBroadcaster struct {
	snapshots []vote.Snapshot
}

func (m *mockBroadcaster) BroadcastSnapshot(snapshot vote.Snapshot) error {
	m.snapshots = append(m.snapshots, snapshot)
	return nil
}

type mockChangePublisher struct {
	changes []vote.StateChange
}

func (m *mockChangePublisher) PublishChange(_ context.Context, change vote.StateChange) error {
	m.changes = append(m.changes, change)
	return nil
}

func TestGetButtonsReturnsCurrentList(t *testing.T) {
	store := &mockStore{
		state: vote.State{
			Buttons: []vote.Button{
				{
					Key:      "feel",
					RedisKey: "vote:button:feel",
					Label:    "有感觉吗",
					Count:    2,
					Sort:     10,
					Enabled:  true,
				},
			},
			Leaderboard: []vote.LeaderboardEntry{
				{Rank: 1, Nickname: "阿明", ClickCount: 9},
			},
			Boss: &vote.Boss{
				ID:        "slime-king",
				Name:      "史莱姆王",
				Status:    "active",
				MaxHP:     100,
				CurrentHP: 80,
			},
			BossLoot: []vote.BossLootEntry{
				{
					ItemID:                     "cloth-armor",
					ItemName:                   "布甲",
					Slot:                       "armor",
					Weight:                     3,
					BonusClicks:                1,
					BonusCriticalChancePercent: 2,
				},
			},
		},
	}
	broadcaster := &mockBroadcaster{}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: broadcaster,
	})

	request := httptest.NewRequest(http.MethodGet, "/api/buttons", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload struct {
		Buttons     []vote.Button           `json:"buttons"`
		Leaderboard []vote.LeaderboardEntry `json:"leaderboard"`
		BossLoot    []vote.BossLootEntry    `json:"bossLoot"`
	}
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(payload.Buttons) != 1 || payload.Buttons[0].Count != 2 {
		t.Fatalf("unexpected buttons payload: %+v", payload.Buttons)
	}
	if len(payload.Leaderboard) != 1 || payload.Leaderboard[0].Nickname != "阿明" {
		t.Fatalf("unexpected leaderboard payload: %+v", payload.Leaderboard)
	}
	if len(payload.BossLoot) != 1 || payload.BossLoot[0].ItemID != "cloth-armor" || payload.BossLoot[0].BonusClicks != 1 {
		t.Fatalf("unexpected boss loot payload: %+v", payload.BossLoot)
	}

	if len(broadcaster.snapshots) != 0 {
		t.Fatalf("expected no broadcasts, got %d", len(broadcaster.snapshots))
	}
}

func TestGetBossHistoryReturnsPublicHistory(t *testing.T) {
	store := &mockStore{
		bossHistory: []vote.BossHistoryEntry{
			{
				Boss: vote.Boss{
					ID:         "slime-king",
					Name:       "史莱姆王",
					Status:     "defeated",
					MaxHP:      100,
					CurrentHP:  0,
					StartedAt:  1710000000,
					DefeatedAt: 1710000300,
				},
				Loot: []vote.BossLootEntry{
					{ItemID: "cloth-armor", ItemName: "布甲", Weight: 3},
				},
				Damage: []vote.BossLeaderboardEntry{
					{Rank: 1, Nickname: "阿明", Damage: 42},
				},
			},
		},
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodGet, "/api/boss/history", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload []vote.BossHistoryEntry
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(payload) != 1 || payload[0].Name != "史莱姆王" {
		t.Fatalf("unexpected history payload: %+v", payload)
	}
	if len(payload[0].Damage) != 1 || payload[0].Damage[0].Nickname != "阿明" {
		t.Fatalf("unexpected history damage payload: %+v", payload[0].Damage)
	}
}

func TestGetLatestAnnouncementReturnsPayload(t *testing.T) {
	store := &mockStore{
		latestAnnouncement: &vote.Announcement{
			ID:          "7",
			Title:       "更新公告",
			Content:     "留言墙已上线。",
			PublishedAt: 1710000000,
			Active:      true,
		},
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodGet, "/api/announcements/latest", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload vote.Announcement
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.ID != "7" || payload.Title != "更新公告" {
		t.Fatalf("unexpected latest announcement payload: %+v", payload)
	}
}

func TestClickButtonDoesNotUseLegacySnapshotBroadcast(t *testing.T) {
	store := &mockStore{
		state: vote.State{
			Buttons: []vote.Button{
				{
					Key:      "feel",
					RedisKey: "vote:button:feel",
					Label:    "有感觉吗",
					Count:    2,
					Sort:     10,
					Enabled:  true,
				},
			},
			Leaderboard: []vote.LeaderboardEntry{
				{Rank: 1, Nickname: "阿明", ClickCount: 3},
			},
			UserStats: &vote.UserStats{Nickname: "阿明", ClickCount: 2},
		},
	}
	broadcaster := &mockBroadcaster{}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: broadcaster,
	})

	request := httptest.NewRequest(http.MethodPost, "/api/buttons/feel/click", strings.NewReader(`{"nickname":"阿明"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload struct {
		Button      vote.Button             `json:"button"`
		Buttons     []vote.Button           `json:"buttons"`
		Delta       int64                   `json:"delta"`
		Critical    bool                    `json:"critical"`
		UserStats   vote.UserStats          `json:"userStats"`
		Leaderboard []vote.LeaderboardEntry `json:"leaderboard"`
	}
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Button.Count != 3 {
		t.Fatalf("expected count 3, got %d", payload.Button.Count)
	}
	if payload.Delta != 1 || payload.Critical {
		t.Fatalf("expected normal click payload, got delta=%d critical=%v", payload.Delta, payload.Critical)
	}
	if payload.UserStats.Nickname != "阿明" {
		t.Fatalf("expected user stats for 阿明, got %+v", payload.UserStats)
	}

	if len(broadcaster.snapshots) != 0 {
		t.Fatalf("expected no legacy snapshot broadcast, got %+v", broadcaster.snapshots)
	}
}

func TestClickButtonReturnsCriticalMetadata(t *testing.T) {
	store := &mockStore{
		state: vote.State{
			Buttons: []vote.Button{
				{
					Key:      "feel",
					RedisKey: "vote:button:feel",
					Label:    "有感觉吗",
					Count:    2,
					Sort:     10,
					Enabled:  true,
				},
			},
			Leaderboard: []vote.LeaderboardEntry{
				{Rank: 1, Nickname: "阿明", ClickCount: 7},
			},
			UserStats: &vote.UserStats{Nickname: "阿明", ClickCount: 7},
		},
		result: vote.ClickResult{
			Button: vote.Button{
				Key:      "feel",
				RedisKey: "vote:button:feel",
				Label:    "有感觉吗",
				Count:    7,
				Sort:     10,
				Enabled:  true,
			},
			Delta:    5,
			Critical: true,
			UserStats: vote.UserStats{
				Nickname:   "阿明",
				ClickCount: 7,
			},
		},
	}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/buttons/feel/click", strings.NewReader(`{"nickname":"阿明"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload struct {
		Delta    int64 `json:"delta"`
		Critical bool  `json:"critical"`
	}
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Delta != 5 || !payload.Critical {
		t.Fatalf("expected critical payload, got delta=%d critical=%v", payload.Delta, payload.Critical)
	}
}

func TestClickButtonPublishesStateChangeWithoutRefetchingState(t *testing.T) {
	store := &mockStore{
		getStateErr: context.DeadlineExceeded,
		result: vote.ClickResult{
			Button: vote.Button{
				Key:      "feel",
				RedisKey: "vote:button:feel",
				Label:    "有感觉吗",
				Count:    5,
				Sort:     10,
				Enabled:  true,
			},
			Delta:    1,
			Critical: false,
			UserStats: vote.UserStats{
				Nickname:   "阿明",
				ClickCount: 5,
			},
		},
		state: vote.State{
			Buttons: []vote.Button{
				{Key: "feel", Label: "有感觉吗", Count: 4, Sort: 10, Enabled: true},
			},
		},
	}
	changePublisher := &mockChangePublisher{}

	handler := NewHandler(Options{
		Store:           store,
		Broadcaster:     &mockBroadcaster{},
		ChangePublisher: changePublisher,
	})

	request := httptest.NewRequest(http.MethodPost, "/api/buttons/feel/click", strings.NewReader(`{"nickname":"阿明"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload struct {
		Button    vote.Button    `json:"button"`
		Delta     int64          `json:"delta"`
		Critical  bool           `json:"critical"`
		UserStats vote.UserStats `json:"userStats"`
	}
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Button.Count != 5 || payload.UserStats.ClickCount != 5 {
		t.Fatalf("unexpected click payload: %+v", payload)
	}
	if len(changePublisher.changes) != 1 {
		t.Fatalf("expected one published change, got %+v", changePublisher.changes)
	}
	if changePublisher.changes[0].Type != vote.StateChangeButtonClicked || changePublisher.changes[0].Nickname != "阿明" {
		t.Fatalf("unexpected published change: %+v", changePublisher.changes[0])
	}
}

func TestEquipItemReturnsUpdatedState(t *testing.T) {
	store := &mockStore{
		equipState: vote.State{
			Buttons: []vote.Button{
				{
					Key:      "feel",
					RedisKey: "vote:button:feel",
					Label:    "有感觉吗",
					Count:    3,
					Sort:     10,
					Enabled:  true,
				},
			},
			Loadout: vote.Loadout{
				Weapon: &vote.InventoryItem{
					ItemID:      "wood-sword",
					Name:        "木剑",
					Slot:        "weapon",
					Quantity:    1,
					BonusClicks: 2,
					Equipped:    true,
				},
			},
			CombatStats: vote.CombatStats{
				BaseIncrement:      1,
				BonusClicks:        2,
				EffectiveIncrement: 3,
				NormalDamage:       3,
				CriticalDamage:     7,
			},
		},
	}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/equipment/wood-sword/equip", strings.NewReader(`{"nickname":"阿明"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload struct {
		Loadout struct {
			Weapon *vote.InventoryItem `json:"weapon"`
		} `json:"loadout"`
		CombatStats vote.CombatStats `json:"combatStats"`
	}
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Loadout.Weapon == nil || payload.Loadout.Weapon.ItemID != "wood-sword" {
		t.Fatalf("expected equipped wood-sword, got %+v", payload.Loadout.Weapon)
	}
	if payload.CombatStats.EffectiveIncrement != 3 {
		t.Fatalf("expected effective increment 3, got %+v", payload.CombatStats)
	}
	if payload.CombatStats.NormalDamage != 3 || payload.CombatStats.CriticalDamage != 7 {
		t.Fatalf("expected actual damage 3/7, got %+v", payload.CombatStats)
	}
}

func TestSynthesizeItemReturnsUpdatedState(t *testing.T) {
	store := &mockStore{
		equipState: vote.State{
			Inventory: []vote.InventoryItem{
				{
					ItemID:      "wood-sword",
					Name:        "木剑 +1",
					Quantity:    1,
					StarLevel:   1,
					BonusClicks: 3,
				},
			},
			Loadout: vote.Loadout{
				Weapon: &vote.InventoryItem{
					ItemID:      "wood-sword",
					Name:        "木剑 +1",
					Quantity:    1,
					StarLevel:   1,
					BonusClicks: 3,
					Equipped:    true,
				},
			},
		},
	}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/equipment/wood-sword/synthesize", strings.NewReader(`{"nickname":"阿明"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload struct {
		Loadout struct {
			Weapon *vote.InventoryItem `json:"weapon"`
		} `json:"loadout"`
	}
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Loadout.Weapon == nil || payload.Loadout.Weapon.Name != "木剑 +1" || payload.Loadout.Weapon.StarLevel != 1 {
		t.Fatalf("unexpected synthesize payload: %+v", payload.Loadout.Weapon)
	}
}

func TestSalvageEquipmentReturnsUpdatedState(t *testing.T) {
	store := &mockStore{
		equipState: vote.State{
			Gems: 2,
			Inventory: []vote.InventoryItem{
				{
					ItemID:   "wood-sword",
					Name:     "木剑",
					Quantity: 1,
				},
			},
			LastForgeResult: &vote.ForgeResult{
				Kind:          "equipment_salvage",
				TargetID:      "wood-sword",
				RemainingGems: 2,
			},
		},
	}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/equipment/wood-sword/salvage", strings.NewReader(`{"nickname":"阿明","quantity":2}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if store.lastSalvageItemID != "wood-sword" || store.lastSalvageQuantity != 2 {
		t.Fatalf("expected salvage payload to be forwarded, got id=%s quantity=%d", store.lastSalvageItemID, store.lastSalvageQuantity)
	}

	var payload struct {
		Gems            int64             `json:"gems"`
		LastForgeResult *vote.ForgeResult `json:"lastForgeResult"`
	}
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Gems != 2 || payload.LastForgeResult == nil || payload.LastForgeResult.Kind != "equipment_salvage" {
		t.Fatalf("unexpected salvage payload: %+v", payload)
	}
}

func TestReforgeEquipmentReturnsUpdatedState(t *testing.T) {
	store := &mockStore{
		equipState: vote.State{
			Gems: 0,
			Inventory: []vote.InventoryItem{
				{
					ItemID:      "wood-sword",
					Name:        "木剑",
					Quantity:    1,
					BonusClicks: 3,
				},
			},
			LastForgeResult: &vote.ForgeResult{
				Kind:          "equipment_reforge",
				TargetID:      "wood-sword",
				RemainingGems: 0,
			},
		},
	}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/equipment/wood-sword/reforge", strings.NewReader(`{"nickname":"阿明"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if store.lastReforgeItemID != "wood-sword" {
		t.Fatalf("expected reforge payload to be forwarded, got %+v", store.lastReforgeItemID)
	}
}

func TestEquipHeroReturnsUpdatedState(t *testing.T) {
	store := &mockStore{
		equipState: vote.State{
			Heroes: []vote.HeroInventoryItem{
				{
					HeroID:      "spark-cat",
					Name:        "星火猫",
					Quantity:    1,
					Active:      true,
					BonusClicks: 2,
				},
			},
			ActiveHero: &vote.HeroInventoryItem{
				HeroID:      "spark-cat",
				Name:        "星火猫",
				Quantity:    1,
				Active:      true,
				BonusClicks: 2,
			},
			CombatStats: vote.CombatStats{
				BaseIncrement:      1,
				BonusClicks:        2,
				EffectiveIncrement: 3,
				NormalDamage:       3,
				CriticalDamage:     7,
			},
		},
	}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/heroes/spark-cat/equip", strings.NewReader(`{"nickname":"阿明"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload struct {
		ActiveHero *vote.HeroInventoryItem `json:"activeHero"`
	}
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.ActiveHero == nil || payload.ActiveHero.HeroID != "spark-cat" {
		t.Fatalf("unexpected hero equip payload: %+v", payload.ActiveHero)
	}
}

func TestAwakenHeroReturnsUpdatedState(t *testing.T) {
	store := &mockStore{
		equipState: vote.State{
			Gems: 0,
			Heroes: []vote.HeroInventoryItem{
				{
					HeroID:      "spark-cat",
					Name:        "星火猫",
					AwakenLevel: 1,
					BonusClicks: 3,
				},
			},
			LastForgeResult: &vote.ForgeResult{
				Kind:       "hero_awaken",
				TargetID:   "spark-cat",
				TargetName: "星火猫",
			},
		},
	}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/heroes/spark-cat/awaken", strings.NewReader(`{"nickname":"阿明"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if store.lastAwakenHeroID != "spark-cat" {
		t.Fatalf("expected awaken payload to be forwarded, got %+v", store.lastAwakenHeroID)
	}
}

func TestShopRoutesReturnCatalogAndLoadout(t *testing.T) {
	store := &mockStore{
		state: vote.State{
			Gems: 30,
			ShopCatalog: []vote.CosmeticCatalogItem{
				{
					CosmeticID: "trail-ribbon",
					Name:       "流星彩带轨迹",
					Type:       vote.CosmeticTypeTrail,
					Price:      30,
					Owned:      true,
					Equipped:   true,
				},
				{
					CosmeticID: "impact-firefly",
					Name:       "流萤追光点击特效",
					Type:       vote.CosmeticTypeImpact,
					Price:      30,
				},
			},
			EquippedCosmetics: vote.CosmeticLoadout{
				TrailID: "trail-ribbon",
			},
		},
		equipState: vote.State{
			Gems: 0,
			ShopCatalog: []vote.CosmeticCatalogItem{
				{
					CosmeticID: "trail-ribbon",
					Name:       "流星彩带轨迹",
					Type:       vote.CosmeticTypeTrail,
					Price:      30,
					Owned:      true,
					Equipped:   true,
				},
				{
					CosmeticID: "impact-firefly",
					Name:       "流萤追光点击特效",
					Type:       vote.CosmeticTypeImpact,
					Price:      30,
					Owned:      true,
					Equipped:   true,
				},
			},
			EquippedCosmetics: vote.CosmeticLoadout{
				TrailID:  "trail-ribbon",
				ImpactID: "impact-firefly",
			},
		},
	}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	getRequest := httptest.NewRequest(http.MethodGet, "/api/shop?nickname=%E9%98%BF%E6%98%8E", nil)
	getResponse := httptest.NewRecorder()
	handler.ServeHTTP(getResponse, getRequest)

	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from shop list, got %d", getResponse.Code)
	}

	purchaseRequest := httptest.NewRequest(http.MethodPost, "/api/shop/cosmetics/impact-firefly/purchase", strings.NewReader(`{"nickname":"阿明"}`))
	purchaseRequest.Header.Set("Content-Type", "application/json")
	purchaseResponse := httptest.NewRecorder()
	handler.ServeHTTP(purchaseResponse, purchaseRequest)

	if purchaseResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from cosmetic purchase, got %d", purchaseResponse.Code)
	}
	if store.lastPurchasedCosmetic != "impact-firefly" {
		t.Fatalf("expected purchase payload to be forwarded, got %+v", store.lastPurchasedCosmetic)
	}

	equipRequest := httptest.NewRequest(http.MethodPost, "/api/shop/cosmetics/equip", strings.NewReader(`{"nickname":"阿明","trailId":"trail-ribbon","impactId":"impact-firefly"}`))
	equipRequest.Header.Set("Content-Type", "application/json")
	equipResponse := httptest.NewRecorder()
	handler.ServeHTTP(equipResponse, equipRequest)

	if equipResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from cosmetic equip, got %d", equipResponse.Code)
	}
	if store.lastCosmeticLoadout.TrailID != "trail-ribbon" || store.lastCosmeticLoadout.ImpactID != "impact-firefly" {
		t.Fatalf("expected cosmetic loadout to be forwarded, got %+v", store.lastCosmeticLoadout)
	}
}

func TestPostMessageRejectsSensitiveContent(t *testing.T) {
	store := &mockStore{
		messageErr: vote.ErrSensitiveContent,
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/messages", strings.NewReader(`{"nickname":"阿明","content":"XJP后援会"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
	if body := response.Body.String(); !strings.Contains(body, "敏感词") {
		t.Fatalf("expected sensitive message content error, got %q", body)
	}
}

func TestAdminLoginCreatesSessionAndStateRequiresAuth(t *testing.T) {
	store := &mockStore{
		adminState: vote.AdminState{
			Buttons: []vote.Button{
				{
					Key:      "feel",
					RedisKey: "vote:button:feel",
					Label:    "有感觉吗",
					Count:    3,
					Sort:     10,
					Enabled:  true,
				},
			},
			Equipment: []vote.EquipmentDefinition{
				{ItemID: "wood-sword", Name: "木剑", Slot: "weapon", BonusClicks: 2},
			},
		},
	}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		}),
	})

	unauthorizedRequest := httptest.NewRequest(http.MethodGet, "/api/admin/state", nil)
	unauthorizedResponse := httptest.NewRecorder()
	handler.ServeHTTP(unauthorizedResponse, unauthorizedRequest)

	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without session, got %d", unauthorizedResponse.Code)
	}

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{"username":"admin","password":"secret"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	handler.ServeHTTP(loginResponse, loginRequest)

	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from login, got %d", loginResponse.Code)
	}

	cookies := loginResponse.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected login to set session cookie")
	}

	adminRequest := httptest.NewRequest(http.MethodGet, "/api/admin/state", nil)
	adminRequest.AddCookie(cookies[0])
	adminResponse := httptest.NewRecorder()
	handler.ServeHTTP(adminResponse, adminRequest)

	if adminResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 with session, got %d", adminResponse.Code)
	}

	var payload struct {
		Buttons   []vote.Button              `json:"buttons"`
		Equipment []vote.EquipmentDefinition `json:"equipment"`
	}
	if err := sonic.Unmarshal(adminResponse.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(payload.Buttons) != 1 || payload.Buttons[0].Key != "feel" {
		t.Fatalf("unexpected admin buttons payload: %+v", payload.Buttons)
	}
	if len(payload.Equipment) != 1 || payload.Equipment[0].ItemID != "wood-sword" {
		t.Fatalf("unexpected admin equipment payload: %+v", payload.Equipment)
	}
}

func TestAdminActivateBossAndSaveButton(t *testing.T) {
	store := &mockStore{}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		}),
	})

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{"username":"admin","password":"secret"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	handler.ServeHTTP(loginResponse, loginRequest)

	cookies := loginResponse.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie from login")
	}

	activateRequest := httptest.NewRequest(http.MethodPost, "/api/admin/boss/activate", strings.NewReader(`{"id":"slime-king","name":"史莱姆王","maxHp":50}`))
	activateRequest.Header.Set("Content-Type", "application/json")
	activateRequest.AddCookie(cookies[0])
	activateResponse := httptest.NewRecorder()
	handler.ServeHTTP(activateResponse, activateRequest)

	if activateResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from boss activate, got %d", activateResponse.Code)
	}
	if store.lastBoss.ID != "slime-king" || store.lastBoss.MaxHP != 50 {
		t.Fatalf("expected boss payload to be forwarded, got %+v", store.lastBoss)
	}

	saveButtonRequest := httptest.NewRequest(http.MethodPost, "/api/admin/buttons", strings.NewReader(`{"slug":"new-one","label":"新按钮","sort":40,"enabled":true}`))
	saveButtonRequest.Header.Set("Content-Type", "application/json")
	saveButtonRequest.AddCookie(cookies[0])
	saveButtonResponse := httptest.NewRecorder()
	handler.ServeHTTP(saveButtonResponse, saveButtonRequest)

	if saveButtonResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from save button, got %d", saveButtonResponse.Code)
	}
	if store.lastButton.Slug != "new-one" || store.lastButton.Label != "新按钮" {
		t.Fatalf("expected button payload to be forwarded, got %+v", store.lastButton)
	}
}

func TestAdminBossPoolRoutesForwardTemplateAndCyclePayloads(t *testing.T) {
	store := &mockStore{}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		}),
	})

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{"username":"admin","password":"secret"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	handler.ServeHTTP(loginResponse, loginRequest)

	cookies := loginResponse.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie from login")
	}

	saveTemplateRequest := httptest.NewRequest(http.MethodPost, "/api/admin/boss/pool", strings.NewReader(`{"id":"dragon","name":"火龙","maxHp":80}`))
	saveTemplateRequest.Header.Set("Content-Type", "application/json")
	saveTemplateRequest.AddCookie(cookies[0])
	saveTemplateResponse := httptest.NewRecorder()
	handler.ServeHTTP(saveTemplateResponse, saveTemplateRequest)

	if saveTemplateResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from boss template save, got %d", saveTemplateResponse.Code)
	}
	if store.lastBossTemplate.ID != "dragon" || store.lastBossTemplate.MaxHP != 80 {
		t.Fatalf("expected template payload to be forwarded, got %+v", store.lastBossTemplate)
	}

	saveLootRequest := httptest.NewRequest(http.MethodPut, "/api/admin/boss/pool/dragon/loot", strings.NewReader(`{"loot":[{"itemId":"fire-ring","weight":3}]}`))
	saveLootRequest.Header.Set("Content-Type", "application/json")
	saveLootRequest.AddCookie(cookies[0])
	saveLootResponse := httptest.NewRecorder()
	handler.ServeHTTP(saveLootResponse, saveLootRequest)

	if saveLootResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from boss template loot save, got %d", saveLootResponse.Code)
	}
	if store.lastTemplateLootID != "dragon" || len(store.lastTemplateLoot) != 1 || store.lastTemplateLoot[0].ItemID != "fire-ring" {
		t.Fatalf("expected template loot payload to be forwarded, got id=%s loot=%+v", store.lastTemplateLootID, store.lastTemplateLoot)
	}

	enableCycleRequest := httptest.NewRequest(http.MethodPost, "/api/admin/boss/cycle/enable", nil)
	enableCycleRequest.AddCookie(cookies[0])
	enableCycleResponse := httptest.NewRecorder()
	handler.ServeHTTP(enableCycleResponse, enableCycleRequest)

	if enableCycleResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from cycle enable, got %d", enableCycleResponse.Code)
	}
	if !store.lastCycleEnabled {
		t.Fatal("expected cycle enable to be forwarded to store")
	}
}

func TestAdminHeroRoutesForwardPayloads(t *testing.T) {
	store := &mockStore{}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		}),
	})

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{"username":"admin","password":"secret"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	handler.ServeHTTP(loginResponse, loginRequest)

	cookies := loginResponse.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie from login")
	}

	saveHeroRequest := httptest.NewRequest(http.MethodPost, "/api/admin/heroes", strings.NewReader(`{"heroId":"spark-cat","name":"星火猫","bonusClicks":2,"traitType":"final_damage_percent","traitValue":50}`))
	saveHeroRequest.Header.Set("Content-Type", "application/json")
	saveHeroRequest.AddCookie(cookies[0])
	saveHeroResponse := httptest.NewRecorder()
	handler.ServeHTTP(saveHeroResponse, saveHeroRequest)

	if saveHeroResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from hero save, got %d", saveHeroResponse.Code)
	}
	if store.lastHero.HeroID != "spark-cat" || store.lastHero.TraitValue != 50 {
		t.Fatalf("expected hero payload to be forwarded, got %+v", store.lastHero)
	}

	saveHeroLootRequest := httptest.NewRequest(http.MethodPut, "/api/admin/boss/pool/dragon/hero-loot", strings.NewReader(`{"loot":[{"heroId":"spark-cat","weight":2}]}`))
	saveHeroLootRequest.Header.Set("Content-Type", "application/json")
	saveHeroLootRequest.AddCookie(cookies[0])
	saveHeroLootResponse := httptest.NewRecorder()
	handler.ServeHTTP(saveHeroLootResponse, saveHeroLootRequest)

	if saveHeroLootResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from hero loot save, got %d", saveHeroLootResponse.Code)
	}
	if store.lastTemplateHeroLootID != "dragon" || len(store.lastTemplateHeroLoot) != 1 || store.lastTemplateHeroLoot[0].HeroID != "spark-cat" {
		t.Fatalf("expected hero loot payload to be forwarded, got id=%s loot=%+v", store.lastTemplateHeroLootID, store.lastTemplateHeroLoot)
	}
}

func TestAdminPlayersPageRequiresAuthAndReturnsPayload(t *testing.T) {
	store := &mockStore{
		adminPlayerPage: vote.AdminPlayerPage{
			Items: []vote.AdminPlayerOverview{
				{Nickname: "阿明", ClickCount: 12},
			},
			NextCursor: "1",
			Total:      3,
		},
	}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		}),
	})

	unauthorizedRequest := httptest.NewRequest(http.MethodGet, "/api/admin/players", nil)
	unauthorizedResponse := httptest.NewRecorder()
	handler.ServeHTTP(unauthorizedResponse, unauthorizedRequest)
	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without session, got %d", unauthorizedResponse.Code)
	}

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{"username":"admin","password":"secret"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	handler.ServeHTTP(loginResponse, loginRequest)

	cookies := loginResponse.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie from login")
	}

	request := httptest.NewRequest(http.MethodGet, "/api/admin/players?cursor=0&limit=1", nil)
	request.AddCookie(cookies[0])
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload vote.AdminPlayerPage
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Total != 3 || payload.NextCursor != "1" || len(payload.Items) != 1 || payload.Items[0].Nickname != "阿明" {
		t.Fatalf("unexpected admin player page payload: %+v", payload)
	}
}

func TestAdminOSSPolicyRequiresAuthAndReturnsPayload(t *testing.T) {
	store := &mockStore{}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		}),
		OSSSigner: &mockOSSSigner{
			policy: ossupload.Policy{
				AccessKeyID:   "test-ak",
				Policy:        "policy",
				Signature:     "signature",
				Host:          "https://vote-wall.oss-cn-beijing.aliyuncs.com",
				Dir:           "buttons/20260419/",
				PublicBaseURL: "https://cdn.example.com",
			},
		},
	})

	unauthorizedRequest := httptest.NewRequest(http.MethodPost, "/api/admin/oss/sts", nil)
	unauthorizedResponse := httptest.NewRecorder()
	handler.ServeHTTP(unauthorizedResponse, unauthorizedRequest)

	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without session, got %d", unauthorizedResponse.Code)
	}

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{"username":"admin","password":"secret"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	handler.ServeHTTP(loginResponse, loginRequest)

	cookies := loginResponse.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie from login")
	}

	request := httptest.NewRequest(http.MethodPost, "/api/admin/oss/sts", nil)
	request.AddCookie(cookies[0])
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload map[string]any
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["host"] != "https://vote-wall.oss-cn-beijing.aliyuncs.com" {
		t.Fatalf("unexpected oss payload: %+v", payload)
	}
}

func TestClickMissingButtonReturnsNotFound(t *testing.T) {
	store := &mockStore{
		state: vote.State{
			Buttons: []vote.Button{
				{
					Key:     "feel",
					Label:   "有感觉吗",
					Enabled: true,
				},
			},
		},
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/buttons/missing/click", strings.NewReader(`{"nickname":"阿明"}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
}

func TestClickRequiresNickname(t *testing.T) {
	store := &mockStore{
		state: vote.State{
			Buttons: []vote.Button{
				{Key: "feel", Label: "有感觉吗", Enabled: true},
			},
		},
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/buttons/feel/click", strings.NewReader(`{"nickname":"   "}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestValidateNicknameRejectsSensitiveNickname(t *testing.T) {
	store := &mockStore{
		validateErr: vote.ErrSensitiveNickname,
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/nickname/validate", strings.NewReader(`{"nickname":"我是习近平"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
	if body := response.Body.String(); !strings.Contains(body, "敏感词") {
		t.Fatalf("expected sensitive-word message, got %q", body)
	}
}

func TestGetButtonsRejectsSensitiveNickname(t *testing.T) {
	store := &mockStore{
		getStateErr: vote.ErrSensitiveNickname,
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodGet, "/api/buttons?nickname=%E4%B9%A0%E8%BF%91%E5%B9%B3", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}
