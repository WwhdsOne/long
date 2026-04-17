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
	buttons []vote.Button
	result  vote.ClickResult
}

func (m *mockStore) ListButtons(context.Context) ([]vote.Button, error) {
	return m.buttons, nil
}

func (m *mockStore) ClickButton(_ context.Context, slug string) (vote.ClickResult, error) {
	for index := range m.buttons {
		if m.buttons[index].Key == slug {
			if m.result.Button.Key == "" {
				m.buttons[index].Count++
				return vote.ClickResult{
					Button:   m.buttons[index],
					Delta:    1,
					Critical: false,
				}, nil
			}
			m.buttons[index].Count = m.result.Button.Count
			return m.result, nil
		}
	}
	return vote.ClickResult{}, vote.ErrButtonNotFound
}

type mockBroadcaster struct {
	snapshots [][]vote.Button
}

func (m *mockBroadcaster) BroadcastSnapshot(buttons []vote.Button) error {
	copied := append([]vote.Button(nil), buttons...)
	m.snapshots = append(m.snapshots, copied)
	return nil
}

func TestGetButtonsReturnsCurrentList(t *testing.T) {
	store := &mockStore{
		buttons: []vote.Button{
			{
				Key:      "feel",
				RedisKey: "vote:button:feel",
				Label:    "有感觉吗",
				Count:    2,
				Sort:     10,
				Enabled:  true,
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
		Buttons []vote.Button `json:"buttons"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(payload.Buttons) != 1 || payload.Buttons[0].Count != 2 {
		t.Fatalf("unexpected buttons payload: %+v", payload.Buttons)
	}

	if len(broadcaster.snapshots) != 0 {
		t.Fatalf("expected no broadcasts, got %d", len(broadcaster.snapshots))
	}
}

func TestClickButtonBroadcastsLatestSnapshot(t *testing.T) {
	store := &mockStore{
		buttons: []vote.Button{
			{
				Key:      "feel",
				RedisKey: "vote:button:feel",
				Label:    "有感觉吗",
				Count:    2,
				Sort:     10,
				Enabled:  true,
			},
		},
	}
	broadcaster := &mockBroadcaster{}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: broadcaster,
	})

	request := httptest.NewRequest(http.MethodPost, "/api/buttons/feel/click", strings.NewReader(""))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload struct {
		Button   vote.Button   `json:"button"`
		Buttons  []vote.Button `json:"buttons"`
		Delta    int64         `json:"delta"`
		Critical bool          `json:"critical"`
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

	if len(broadcaster.snapshots) != 1 || broadcaster.snapshots[0][0].Count != 3 {
		t.Fatalf("unexpected broadcast payload: %+v", broadcaster.snapshots)
	}
}

func TestClickButtonReturnsCriticalMetadata(t *testing.T) {
	store := &mockStore{
		buttons: []vote.Button{
			{
				Key:      "feel",
				RedisKey: "vote:button:feel",
				Label:    "有感觉吗",
				Count:    2,
				Sort:     10,
				Enabled:  true,
			},
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
		},
	}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/buttons/feel/click", strings.NewReader(""))
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
		buttons: []vote.Button{
			{
				Key:     "feel",
				Label:   "有感觉吗",
				Enabled: true,
			},
		},
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/buttons/missing/click", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
}
