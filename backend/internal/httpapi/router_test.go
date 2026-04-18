package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"long/internal/admin"
	"long/internal/vote"
)

type mockStore struct {
	state       vote.State
	equipState  vote.State
	adminState  vote.AdminState
	result      vote.ClickResult
	lastButton  vote.ButtonUpsert
	lastBoss    vote.BossUpsert
	getStateErr error
	clickErr    error
	equipErr    error
	validateErr error
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

func (m *mockStore) SaveButton(_ context.Context, button vote.ButtonUpsert) error {
	m.lastButton = button
	return nil
}

func (m *mockStore) SaveEquipmentDefinition(_ context.Context, _ vote.EquipmentDefinition) error {
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

func (m *mockStore) ListBossHistory(_ context.Context) ([]vote.BossHistoryEntry, error) {
	return []vote.BossHistoryEntry{}, nil
}

type mockBroadcaster struct {
	snapshots []vote.Snapshot
}

func (m *mockBroadcaster) BroadcastSnapshot(snapshot vote.Snapshot) error {
	m.snapshots = append(m.snapshots, snapshot)
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
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(payload.Buttons) != 1 || payload.Buttons[0].Count != 2 {
		t.Fatalf("unexpected buttons payload: %+v", payload.Buttons)
	}
	if len(payload.Leaderboard) != 1 || payload.Leaderboard[0].Nickname != "阿明" {
		t.Fatalf("unexpected leaderboard payload: %+v", payload.Leaderboard)
	}

	if len(broadcaster.snapshots) != 0 {
		t.Fatalf("expected no broadcasts, got %d", len(broadcaster.snapshots))
	}
}

func TestClickButtonBroadcastsLatestSnapshot(t *testing.T) {
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
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
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

	if len(broadcaster.snapshots) != 1 || broadcaster.snapshots[0].Buttons[0].Count != 3 {
		t.Fatalf("unexpected broadcast payload: %+v", broadcaster.snapshots)
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
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Delta != 5 || !payload.Critical {
		t.Fatalf("expected critical payload, got delta=%d critical=%v", payload.Delta, payload.Critical)
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
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Loadout.Weapon == nil || payload.Loadout.Weapon.ItemID != "wood-sword" {
		t.Fatalf("expected equipped wood-sword, got %+v", payload.Loadout.Weapon)
	}
	if payload.CombatStats.EffectiveIncrement != 3 {
		t.Fatalf("expected effective increment 3, got %+v", payload.CombatStats)
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
	if err := json.Unmarshal(adminResponse.Body.Bytes(), &payload); err != nil {
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
