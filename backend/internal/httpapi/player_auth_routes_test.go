package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"long/internal/admin"
	"long/internal/vote"
)

type mockPlayerAuthenticator struct {
	loginToken        string
	loginNickname     string
	verifyNickname    string
	loginErr          error
	verifyErr         error
	resetErr          error
	lastResetNickname string
	lastResetPassword string
}

func (m *mockPlayerAuthenticator) Login(_ context.Context, nickname string, _ string) (string, string, error) {
	if m.loginErr != nil {
		return "", "", m.loginErr
	}
	if m.loginNickname == "" {
		m.loginNickname = nickname
	}
	return m.loginToken, m.loginNickname, nil
}

func (m *mockPlayerAuthenticator) Verify(_ context.Context, _ string) (string, error) {
	if m.verifyErr != nil {
		return "", m.verifyErr
	}
	return m.verifyNickname, nil
}

func (m *mockPlayerAuthenticator) ResetPassword(_ context.Context, nickname string, password string) error {
	m.lastResetNickname = nickname
	m.lastResetPassword = password
	return m.resetErr
}

func TestPlayerLoginCreatesCookieAndSessionUsesAuthenticatedNickname(t *testing.T) {
	store := &mockStore{
		state: voteStateForPlayerTests(),
	}
	authenticator := &mockPlayerAuthenticator{
		loginToken:     "player-token",
		loginNickname:  "阿明",
		verifyNickname: "阿明",
	}
	handler := NewHandler(Options{
		Store:               store,
		Broadcaster:         &mockBroadcaster{},
		PlayerAuthenticator: authenticator,
	})

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/player/auth/login", strings.NewReader(`{"nickname":"阿明","password":"secret"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()

	handler.ServeHTTP(loginResponse, loginRequest)

	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from player login, got %d", loginResponse.Code)
	}

	cookies := loginResponse.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected player login to set session cookie")
	}

	sessionRequest := httptest.NewRequest(http.MethodGet, "/api/player/auth/session", nil)
	sessionRequest.AddCookie(cookies[0])
	sessionResponse := httptest.NewRecorder()

	handler.ServeHTTP(sessionResponse, sessionRequest)

	if sessionResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from player session, got %d", sessionResponse.Code)
	}

	clickRequest := httptest.NewRequest(http.MethodPost, "/api/equipment/instance-wood-sword/equip", strings.NewReader(`{"nickname":"别人"}`))
	clickRequest.Header.Set("Content-Type", "application/json")
	clickRequest.AddCookie(cookies[0])
	clickResponse := httptest.NewRecorder()

	handler.ServeHTTP(clickResponse, clickRequest)

	if clickResponse.Code != http.StatusOK {
		t.Fatalf("expected authenticated equip to pass, got %d", clickResponse.Code)
	}
	if store.lastClickNickname != "阿明" {
		t.Fatalf("expected click to use authenticated nickname, got %q", store.lastClickNickname)
	}

	stateRequest := httptest.NewRequest(http.MethodGet, "/api/battle/state?nickname=%E5%88%AB%E4%BA%BA", nil)
	stateRequest.AddCookie(cookies[0])
	stateResponse := httptest.NewRecorder()

	handler.ServeHTTP(stateResponse, stateRequest)

	if stateResponse.Code != http.StatusOK {
		t.Fatalf("expected authenticated battle state request to pass, got %d", stateResponse.Code)
	}
	if store.lastGetStateNickname != "阿明" {
		t.Fatalf("expected battle state fetch to use authenticated nickname, got %q", store.lastGetStateNickname)
	}
}

func TestPlayerWriteRoutesRequireAuthenticatedSession(t *testing.T) {
	handler := NewHandler(Options{
		Store:               &mockStore{state: voteStateForPlayerTests()},
		Broadcaster:         &mockBroadcaster{},
		PlayerAuthenticator: &mockPlayerAuthenticator{verifyErr: errors.New("missing")},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/equipment/instance-wood-sword/equip", strings.NewReader(`{"nickname":"阿明"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without player session, got %d", response.Code)
	}
}

func TestPlayerProfileRequiresSessionAndReturnsProfileDataset(t *testing.T) {
	store := &mockStore{
		state: vote.State{
			UserStats: &vote.UserStats{Nickname: "阿明", ClickCount: 7},
			Inventory: []vote.InventoryItem{
				{ItemID: "sword-1", Name: "短剑", Slot: "weapon"},
			},
			Loadout: vote.Loadout{
				Weapon: &vote.InventoryItem{ItemID: "sword-1", Name: "短剑", Slot: "weapon"},
			},
			CombatStats: vote.CombatStats{EffectiveIncrement: 3, NormalDamage: 3, CriticalDamage: 6},
		},
	}
	authenticator := &mockPlayerAuthenticator{verifyNickname: "阿明"}
	handler := NewHandler(Options{
		Store:               store,
		Broadcaster:         &mockBroadcaster{},
		PlayerAuthenticator: authenticator,
	})

	unauthorizedRequest := httptest.NewRequest(http.MethodGet, "/api/player/profile", nil)
	unauthorizedResponse := httptest.NewRecorder()
	handler.ServeHTTP(unauthorizedResponse, unauthorizedRequest)
	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without player session, got %d", unauthorizedResponse.Code)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/player/profile", nil)
	request.AddCookie(&http.Cookie{Name: playerSessionCookieName, Value: "player-token"})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200 from profile endpoint, got %d", response.Code)
	}

	var rawPayload map[string]any
	if err := json.NewDecoder(strings.NewReader(response.Body.String())).Decode(&rawPayload); err != nil {
		t.Fatalf("decode raw profile response: %v", err)
	}
	if _, ok := rawPayload["gems"]; ok {
		t.Fatalf("profile endpoint should not include gems, got %+v", rawPayload)
	}
	if _, ok := rawPayload["lastReward"]; ok {
		t.Fatalf("profile endpoint should not include lastReward, got %+v", rawPayload)
	}

	var payload struct {
		UserStats     *vote.UserStats         `json:"userStats"`
		Inventory     []vote.InventoryItem    `json:"inventory"`
		Loadout       vote.Loadout            `json:"loadout"`
		CombatStats   vote.CombatStats        `json:"combatStats"`
		RecentRewards []vote.Reward           `json:"recentRewards"`
		Leaderboard   []vote.LeaderboardEntry `json:"leaderboard"`
	}
	if err := json.NewDecoder(strings.NewReader(response.Body.String())).Decode(&payload); err != nil {
		t.Fatalf("decode profile response: %v", err)
	}
	if payload.UserStats == nil || payload.UserStats.Nickname != "阿明" {
		t.Fatalf("expected profile user stats for 阿明, got %+v", payload.UserStats)
	}
	if len(payload.Inventory) != 1 {
		t.Fatalf("expected inventory, heroes and active hero in profile, got %+v", payload)
	}
	if len(payload.Leaderboard) != 0 {
		t.Fatalf("profile endpoint should not include public battle fields, got leaderboard=%d", len(payload.Leaderboard))
	}
}

func TestBattleStateOmitsInventoryAndLegacyRewardFields(t *testing.T) {
	store := &mockStore{
		state: vote.State{
			UserStats: &vote.UserStats{Nickname: "阿明", ClickCount: 7},
			Inventory: []vote.InventoryItem{
				{ItemID: "sword-1", Name: "短剑", Slot: "weapon"},
			},
			Loadout:       vote.Loadout{Weapon: &vote.InventoryItem{ItemID: "sword-1", Name: "短剑", Slot: "weapon"}},
			CombatStats:   vote.CombatStats{EffectiveIncrement: 3, NormalDamage: 3, CriticalDamage: 6},
			RecentRewards: []vote.Reward{{ItemID: "sword-1", ItemName: "短剑"}},
			Gold:          12,
			Stones:        34,
			TalentPoints:  56,
		},
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodGet, "/api/battle/state?nickname=%E9%98%BF%E6%98%8E", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200 from battle state endpoint, got %d", response.Code)
	}

	var payload map[string]any
	if err := json.NewDecoder(strings.NewReader(response.Body.String())).Decode(&payload); err != nil {
		t.Fatalf("decode battle state response: %v", err)
	}
	if _, ok := payload["inventory"]; ok {
		t.Fatalf("battle state should not include inventory, got %+v", payload)
	}
	if _, ok := payload["gems"]; ok {
		t.Fatalf("battle state should not include gems, got %+v", payload)
	}
	if _, ok := payload["lastReward"]; ok {
		t.Fatalf("battle state should not include lastReward, got %+v", payload)
	}
	if _, ok := payload["recentRewards"]; !ok {
		t.Fatalf("battle state should include recentRewards, got %+v", payload)
	}
}

func TestTasksEndpointRequiresSessionAndReturnsPlayerTasks(t *testing.T) {
	store := &mockStore{
		tasks: []vote.PlayerTask{
			{
				TaskID:      "daily-click-1",
				Title:       "今日点击",
				Progress:    3,
				TargetValue: 10,
				Status:      vote.TaskPlayerStatusInProgress,
			},
		},
	}
	authenticator := &mockPlayerAuthenticator{verifyNickname: "阿明"}
	handler := NewHandler(Options{
		Store:               store,
		Broadcaster:         &mockBroadcaster{},
		PlayerAuthenticator: authenticator,
	})

	unauthorizedRequest := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	unauthorizedResponse := httptest.NewRecorder()
	handler.ServeHTTP(unauthorizedResponse, unauthorizedRequest)
	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without player session, got %d", unauthorizedResponse.Code)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	request.AddCookie(&http.Cookie{Name: playerSessionCookieName, Value: "player-token"})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200 from tasks endpoint, got %d", response.Code)
	}

	var payload []vote.PlayerTask
	if err := json.NewDecoder(strings.NewReader(response.Body.String())).Decode(&payload); err != nil {
		t.Fatalf("decode tasks response: %v", err)
	}
	if len(payload) != 1 || payload[0].TaskID != "daily-click-1" {
		t.Fatalf("unexpected tasks payload: %+v", payload)
	}
}

func TestClaimTaskEndpointUsesAuthenticatedPlayer(t *testing.T) {
	store := &mockStore{}
	authenticator := &mockPlayerAuthenticator{verifyNickname: "阿明"}
	handler := NewHandler(Options{
		Store:               store,
		Broadcaster:         &mockBroadcaster{},
		PlayerAuthenticator: authenticator,
	})

	request := httptest.NewRequest(http.MethodPost, "/api/tasks/daily-click-1/claim", nil)
	request.AddCookie(&http.Cookie{Name: playerSessionCookieName, Value: "player-token"})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200 from claim task endpoint, got %d", response.Code)
	}
	if store.lastClaimTaskID != "daily-click-1" {
		t.Fatalf("expected claim endpoint to use task id, got %q", store.lastClaimTaskID)
	}
}

func TestAdminCanResetPlayerPassword(t *testing.T) {
	playerAuthenticator := &mockPlayerAuthenticator{}
	handler := NewHandler(Options{
		Store:               &mockStore{state: voteStateForPlayerTests()},
		Broadcaster:         &mockBroadcaster{},
		PlayerAuthenticator: playerAuthenticator,
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
		t.Fatal("expected admin login to set session cookie")
	}

	resetRequest := httptest.NewRequest(http.MethodPost, "/api/admin/players/%E9%98%BF%E6%98%8E/password/reset", strings.NewReader(`{"password":"new-secret"}`))
	resetRequest.Header.Set("Content-Type", "application/json")
	resetRequest.AddCookie(cookies[0])
	resetResponse := httptest.NewRecorder()

	handler.ServeHTTP(resetResponse, resetRequest)

	if resetResponse.Code != http.StatusOK {
		t.Fatalf("expected admin password reset to pass, got %d", resetResponse.Code)
	}
	if playerAuthenticator.lastResetNickname != "阿明" {
		t.Fatalf("expected reset nickname 阿明, got %q", playerAuthenticator.lastResetNickname)
	}
	if playerAuthenticator.lastResetPassword != "new-secret" {
		t.Fatalf("expected reset password to be forwarded, got %q", playerAuthenticator.lastResetPassword)
	}
}

func voteStateForPlayerTests() vote.State {
	return vote.State{
		UserStats: &vote.UserStats{Nickname: "阿明", ClickCount: 2},
	}
}
