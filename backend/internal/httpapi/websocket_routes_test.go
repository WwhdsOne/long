package httpapi

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/bytedance/sonic"

	"long/internal/events"
	"long/internal/ratelimit"
	"long/internal/vote"
)

type dispatchingChangePublisher struct {
	dispatcher *events.Dispatcher
	changes    []vote.StateChange
}

func (p *dispatchingChangePublisher) PublishChange(ctx context.Context, change vote.StateChange) error {
	p.changes = append(p.changes, change)
	if p.dispatcher == nil {
		return nil
	}
	return p.dispatcher.HandleChange(ctx, change)
}

type mockClickGuard struct {
	err   error
	calls []string
}

func (m *mockClickGuard) Allow(key string) (time.Duration, error) {
	m.calls = append(m.calls, key)
	if m.err != nil {
		if m.err == ratelimit.ErrTooManyRequests {
			return 10 * time.Minute, m.err
		}
		return 0, m.err
	}
	return 0, nil
}

func captureRealtimeMessages(t *testing.T, fn func(func(any) error) error) [][]byte {
	t.Helper()

	var messages [][]byte
	err := fn(func(payload any) error {
		encoded, err := sonic.Marshal(payload)
		if err != nil {
			return err
		}
		messages = append(messages, encoded)
		return nil
	})
	if err != nil {
		t.Fatalf("capture realtime messages: %v", err)
	}

	return messages
}

func decodeRealtimeMessage[T any](t *testing.T, payload []byte) T {
	t.Helper()

	var message T
	if err := sonic.Unmarshal(payload, &message); err != nil {
		t.Fatalf("decode realtime message: %v", err)
	}
	return message
}

func TestRealtimeSessionHelloReturnsSnapshotAndUserState(t *testing.T) {
	store := &mockStore{
		state: vote.State{
			Buttons: []vote.Button{
				{Key: "feel", Label: "有感觉吗", Count: 3, Enabled: true},
			},
			Leaderboard: []vote.LeaderboardEntry{
				{Rank: 1, Nickname: "阿明", ClickCount: 3},
			},
			UserStats:   &vote.UserStats{Nickname: "阿明", ClickCount: 3},
			Inventory:   []vote.InventoryItem{{ItemID: "wood-sword", Name: "木剑", Quantity: 1}},
			CombatStats: vote.CombatStats{EffectiveIncrement: 2},
		},
	}
	session := newRealtimeSession(realtimeSessionOptions{
		stateView:            store,
		store:                store,
		hub:                  events.NewHub(),
		authenticatorEnabled: false,
		clientID:             "127.0.0.1",
	})

	messages := captureRealtimeMessages(t, func(send func(any) error) error {
		return session.handleMessage(context.Background(), []byte(`{"type":"hello","nickname":"阿明"}`), send)
	})
	if len(messages) != 1 {
		t.Fatalf("expected one realtime message, got %d", len(messages))
	}

	var response struct {
		Type   string          `json:"type"`
		Public vote.Snapshot   `json:"public"`
		User   *vote.UserState `json:"user"`
	}
	decoded := decodeRealtimeMessage[struct {
		Type   string          `json:"type"`
		Public vote.Snapshot   `json:"public"`
		User   *vote.UserState `json:"user"`
	}](t, messages[0])
	response = decoded

	if response.Type != realtimeMessageTypeSnapshot {
		t.Fatalf("expected snapshot response, got %+v", response)
	}
	if len(response.Public.Buttons) != 1 || response.Public.Buttons[0].Key != "feel" {
		t.Fatalf("unexpected public snapshot: %+v", response.Public)
	}
	if response.User == nil || response.User.UserStats == nil || response.User.UserStats.Nickname != "阿明" {
		t.Fatalf("expected user state for 阿明, got %+v", response.User)
	}
}

func TestRealtimeSessionHelloReturnsPublicOnlyForAnonymousUser(t *testing.T) {
	store := &mockStore{
		state: vote.State{
			Buttons: []vote.Button{
				{Key: "feel", Label: "有感觉吗", Count: 3, Enabled: true},
			},
		},
	}
	session := newRealtimeSession(realtimeSessionOptions{
		stateView:             store,
		store:                 store,
		hub:                   events.NewHub(),
		authenticatorEnabled:  true,
		authenticatedNickname: "",
		clientID:              "127.0.0.1",
	})

	messages := captureRealtimeMessages(t, func(send func(any) error) error {
		return session.handleMessage(context.Background(), []byte(`{"type":"hello","nickname":"阿明"}`), send)
	})
	if len(messages) != 1 {
		t.Fatalf("expected one realtime message, got %d", len(messages))
	}

	response := decodeRealtimeMessage[struct {
		Type string          `json:"type"`
		User *vote.UserState `json:"user"`
	}](t, messages[0])
	if response.Type != realtimeMessageTypeSnapshot {
		t.Fatalf("expected snapshot response, got %+v", response)
	}
	if response.User != nil {
		t.Fatalf("expected anonymous snapshot to omit user state, got %+v", response.User)
	}
	if session.nickname != "" {
		t.Fatalf("expected anonymous session nickname, got %q", session.nickname)
	}
}

func TestRealtimeSessionClickReturnsAckAndPublishesDeltas(t *testing.T) {
	store := &mockStore{
		state: vote.State{
			Buttons: []vote.Button{
				{Key: "feel", Label: "有感觉吗", Count: 4, Enabled: true},
			},
			Leaderboard: []vote.LeaderboardEntry{
				{Rank: 1, Nickname: "阿明", ClickCount: 4},
			},
			UserStats:   &vote.UserStats{Nickname: "阿明", ClickCount: 4},
			CombatStats: vote.CombatStats{EffectiveIncrement: 2},
		},
		result: vote.ClickResult{
			Button: vote.Button{
				Key:     "feel",
				Label:   "有感觉吗",
				Count:   5,
				Enabled: true,
			},
			Delta:    1,
			Critical: false,
			UserStats: vote.UserStats{
				Nickname:   "阿明",
				ClickCount: 5,
			},
		},
	}
	hub := events.NewHub()
	cache := events.NewCache(store)
	dispatcher := events.NewDispatcher(cache, hub)
	publisher := &dispatchingChangePublisher{dispatcher: dispatcher}
	session := newRealtimeSession(realtimeSessionOptions{
		stateView:            store,
		store:                store,
		hub:                  hub,
		changePublisher:      publisher,
		authenticatorEnabled: false,
		clientID:             "127.0.0.1",
	})

	sseClient, unsubscribeSSE := hub.Subscribe("阿明")
	defer unsubscribeSSE()
	wsClient, unsubscribeWS := hub.Subscribe("阿明")
	defer unsubscribeWS()

	_ = captureRealtimeMessages(t, func(send func(any) error) error {
		return session.handleMessage(context.Background(), []byte(`{"type":"hello","nickname":"阿明"}`), send)
	})

	messages := captureRealtimeMessages(t, func(send func(any) error) error {
		return session.handleMessage(context.Background(), []byte(`{"type":"click","slug":"feel"}`), send)
	})
	if len(messages) != 1 {
		t.Fatalf("expected one realtime message, got %d", len(messages))
	}

	ack := decodeRealtimeMessage[struct {
		Type    string `json:"type"`
		Payload struct {
			Button   vote.Button `json:"button"`
			Delta    int64       `json:"delta"`
			Critical bool        `json:"critical"`
		} `json:"payload"`
	}](t, messages[0])
	if ack.Type != realtimeMessageTypeClickAck {
		t.Fatalf("expected click_ack, got %+v", ack)
	}
	if ack.Payload.Button.Key != "feel" || ack.Payload.Delta != 1 || ack.Payload.Critical {
		t.Fatalf("unexpected click ack payload: %+v", ack.Payload)
	}

	publicEvent := <-sseClient
	if publicEvent.Name != events.PublicStateEventName {
		t.Fatalf("expected public state event for SSE subscriber, got %+v", publicEvent)
	}
	userEvent := <-sseClient
	if userEvent.Name != events.UserStateEventName {
		t.Fatalf("expected user state event for SSE subscriber, got %+v", userEvent)
	}

	publicEvent = <-wsClient
	if publicEvent.Name != events.PublicStateEventName {
		t.Fatalf("expected public state event for realtime subscriber, got %+v", publicEvent)
	}
	userEvent = <-wsClient
	if userEvent.Name != events.UserStateEventName {
		t.Fatalf("expected user state event for realtime subscriber, got %+v", userEvent)
	}

	if len(publisher.changes) != 1 || publisher.changes[0].Type != vote.StateChangeButtonClicked {
		t.Fatalf("expected one click change, got %+v", publisher.changes)
	}
}

func TestRealtimeSessionReturnsProtocolErrors(t *testing.T) {
	session := newRealtimeSession(realtimeSessionOptions{
		stateView:            &mockStore{},
		store:                &mockStore{},
		hub:                  events.NewHub(),
		authenticatorEnabled: false,
		clientID:             "127.0.0.1",
	})

	testCases := []struct {
		name string
		raw  string
		code string
	}{
		{
			name: "invalid json",
			raw:  `{`,
			code: realtimeErrorCodeInvalidMessage,
		},
		{
			name: "unknown type",
			raw:  `{"type":"mystery"}`,
			code: realtimeErrorCodeInvalidMessage,
		},
		{
			name: "missing slug",
			raw:  `{"type":"click","slug":"   "}`,
			code: realtimeErrorCodeInvalidRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			messages := captureRealtimeMessages(t, func(send func(any) error) error {
				return session.handleMessage(context.Background(), []byte(tc.raw), send)
			})
			if len(messages) != 1 {
				t.Fatalf("expected one realtime error, got %d", len(messages))
			}

			response := decodeRealtimeMessage[struct {
				Type    string `json:"type"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}](t, messages[0])
			if response.Type != realtimeMessageTypeError || response.Code != tc.code {
				t.Fatalf("unexpected realtime error response: %+v", response)
			}
			if strings.TrimSpace(response.Message) == "" {
				t.Fatalf("expected error message, got %+v", response)
			}
		})
	}
}

func TestRealtimeSessionClickRequiresAuthenticatedNicknameWhenConfigured(t *testing.T) {
	session := newRealtimeSession(realtimeSessionOptions{
		stateView:             &mockStore{},
		store:                 &mockStore{},
		hub:                   events.NewHub(),
		authenticatorEnabled:  true,
		authenticatedNickname: "",
		clientID:              "127.0.0.1",
	})

	messages := captureRealtimeMessages(t, func(send func(any) error) error {
		return session.handleMessage(context.Background(), []byte(`{"type":"click","slug":"feel"}`), send)
	})
	if len(messages) != 1 {
		t.Fatalf("expected one realtime error, got %d", len(messages))
	}

	response := decodeRealtimeMessage[struct {
		Type string `json:"type"`
		Code string `json:"code"`
	}](t, messages[0])
	if response.Type != realtimeMessageTypeError || response.Code != "UNAUTHORIZED" {
		t.Fatalf("unexpected unauthorized response: %+v", response)
	}
}

func TestRealtimeSessionClickReturnsRateLimitError(t *testing.T) {
	session := newRealtimeSession(realtimeSessionOptions{
		stateView:            &mockStore{},
		store:                &mockStore{},
		hub:                  events.NewHub(),
		clickGuard:           &mockClickGuard{err: ratelimit.ErrTooManyRequests},
		authenticatorEnabled: false,
		clientID:             "127.0.0.1",
	})
	_ = captureRealtimeMessages(t, func(send func(any) error) error {
		return session.handleMessage(context.Background(), []byte(`{"type":"hello","nickname":"阿明"}`), send)
	})

	messages := captureRealtimeMessages(t, func(send func(any) error) error {
		return session.handleMessage(context.Background(), []byte(`{"type":"click","slug":"feel"}`), send)
	})
	if len(messages) != 1 {
		t.Fatalf("expected one realtime error, got %d", len(messages))
	}

	response := decodeRealtimeMessage[struct {
		Type string `json:"type"`
		Code string `json:"code"`
	}](t, messages[0])
	if response.Type != realtimeMessageTypeError || response.Code != "TOO_MANY_REQUESTS" {
		t.Fatalf("unexpected rate limit response: %+v", response)
	}
}
