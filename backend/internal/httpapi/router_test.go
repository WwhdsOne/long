package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"long/internal/vote"
)

type mockStore struct {
	state       vote.State
	result      vote.ClickResult
	getStateErr error
	clickErr    error
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
