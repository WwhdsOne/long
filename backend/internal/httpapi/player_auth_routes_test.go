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

	clickRequest := httptest.NewRequest(http.MethodPost, "/api/buttons/feel/click", strings.NewReader(`{"nickname":"别人"}`))
	clickRequest.Header.Set("Content-Type", "application/json")
	clickRequest.AddCookie(cookies[0])
	clickResponse := httptest.NewRecorder()

	handler.ServeHTTP(clickResponse, clickRequest)

	if clickResponse.Code != http.StatusOK {
		t.Fatalf("expected authenticated click to pass, got %d", clickResponse.Code)
	}
	if store.lastClickNickname != "阿明" {
		t.Fatalf("expected click to use authenticated nickname, got %q", store.lastClickNickname)
	}

	stateRequest := httptest.NewRequest(http.MethodGet, "/api/buttons?nickname=%E5%88%AB%E4%BA%BA", nil)
	stateRequest.AddCookie(cookies[0])
	stateResponse := httptest.NewRecorder()

	handler.ServeHTTP(stateResponse, stateRequest)

	if stateResponse.Code != http.StatusOK {
		t.Fatalf("expected authenticated state request to pass, got %d", stateResponse.Code)
	}
	if store.lastGetStateNickname != "阿明" {
		t.Fatalf("expected state fetch to use authenticated nickname, got %q", store.lastGetStateNickname)
	}
}

func TestPlayerWriteRoutesRequireAuthenticatedSession(t *testing.T) {
	handler := NewHandler(Options{
		Store:               &mockStore{state: voteStateForPlayerTests()},
		Broadcaster:         &mockBroadcaster{},
		PlayerAuthenticator: &mockPlayerAuthenticator{verifyErr: errors.New("missing")},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/buttons/feel/click", strings.NewReader(`{"nickname":"阿明"}`))
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
			CombatStats:     vote.CombatStats{EffectiveIncrement: 3, NormalDamage: 3, CriticalDamage: 6},
			Gems:            11,
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

	var payload struct {
		UserStats       *vote.UserStats          `json:"userStats"`
		Inventory       []vote.InventoryItem     `json:"inventory"`
		Loadout         vote.Loadout             `json:"loadout"`
		CombatStats     vote.CombatStats         `json:"combatStats"`
		Gems            int64                    `json:"gems"`
		RecentRewards   []vote.Reward            `json:"recentRewards"`
		LastReward      *vote.Reward             `json:"lastReward"`
		Buttons         []vote.Button            `json:"buttons"`
		Leaderboard     []vote.LeaderboardEntry  `json:"leaderboard"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode profile response: %v", err)
	}
	if payload.UserStats == nil || payload.UserStats.Nickname != "阿明" {
		t.Fatalf("expected profile user stats for 阿明, got %+v", payload.UserStats)
	}
	if len(payload.Inventory) != 1 {
		t.Fatalf("expected inventory, heroes and active hero in profile, got %+v", payload)
	}
	if len(payload.Buttons) != 0 || len(payload.Leaderboard) != 0 {
		t.Fatalf("profile endpoint should not include public battle fields, got buttons=%d leaderboard=%d", len(payload.Buttons), len(payload.Leaderboard))
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
		UserStats: &vote.UserStats{Nickname: "阿明", ClickCount: 2},
	}
}
