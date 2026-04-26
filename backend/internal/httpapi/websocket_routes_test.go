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

func readHubEventByName(t *testing.T, ch <-chan events.ServerEvent, name string) events.ServerEvent {
	t.Helper()

	timeout := time.After(2 * time.Second)
	for {
		select {
		case event, ok := <-ch:
			if !ok {
				t.Fatalf("hub channel closed while waiting for %s", name)
			}
			if event.Name == name {
				return event
			}
		case <-timeout:
			t.Fatalf("timed out waiting for event %s", name)
		}
	}
}

func readHubEventSet(t *testing.T, ch <-chan events.ServerEvent, expected map[string]struct{}) {
	t.Helper()

	remain := make(map[string]struct{}, len(expected))
	for name := range expected {
		remain[name] = struct{}{}
	}
	timeout := time.After(2 * time.Second)
	for len(remain) > 0 {
		select {
		case event, ok := <-ch:
			if !ok {
				t.Fatalf("hub channel closed while waiting events: %+v", remain)
			}
			delete(remain, event.Name)
		case <-timeout:
			t.Fatalf("timed out waiting events: %+v", remain)
		}
	}
}

func readHubPublicAndUserEvent(t *testing.T, ch <-chan events.ServerEvent) events.ServerEvent {
	t.Helper()

	timeout := time.After(2 * time.Second)
	gotPublic := false
	var userEvent events.ServerEvent
	gotUser := false
	for !gotPublic || !gotUser {
		select {
		case event, ok := <-ch:
			if !ok {
				t.Fatalf("hub channel closed while waiting public/user events")
			}
			if event.Name == events.PublicStateEventName {
				gotPublic = true
				continue
			}
			if event.Name == events.UserStateEventName {
				userEvent = event
				gotUser = true
			}
		case <-timeout:
			t.Fatalf("timed out waiting for public/user events")
		}
	}
	return userEvent
}

func assertNoHubEventByName(t *testing.T, ch <-chan events.ServerEvent, name string) {
	t.Helper()

	timeout := time.After(300 * time.Millisecond)
	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return
			}
			if event.Name == name {
				t.Fatalf("expected no %s event, got %+v", name, event)
			}
		case <-timeout:
			return
		}
	}
}

func TestRealtimeSessionHelloReturnsSnapshotAndUserState(t *testing.T) {
	store := &mockStore{
		state: vote.State{
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
	if len(response.Public.Leaderboard) == 0 || response.Public.Leaderboard[0].Nickname != "阿明" {
		t.Fatalf("unexpected public snapshot: %+v", response.Public)
	}
	if response.User == nil || response.User.UserStats == nil || response.User.UserStats.Nickname != "阿明" {
		t.Fatalf("expected user state for 阿明, got %+v", response.User)
	}
	if strings.Contains(string(messages[0]), "\"inventory\"") {
		t.Fatalf("expected realtime snapshot user payload to omit inventory, got %s", string(messages[0]))
	}
}

func TestRealtimeSessionHelloReturnsPublicOnlyForAnonymousUser(t *testing.T) {
	store := &mockStore{
		state: vote.State{},
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

func TestRealtimeSessionHelloReturnsSlimPublicSnapshot(t *testing.T) {
	store := &mockStore{
		snapshot: vote.Snapshot{
			Leaderboard: []vote.LeaderboardEntry{
				{Rank: 1, Nickname: "阿明", ClickCount: 3},
			},
			AnnouncementVersion: "7",
		},
		state: vote.State{
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

	response := decodeRealtimeMessage[struct {
		Type   string          `json:"type"`
		Public vote.Snapshot   `json:"public"`
		User   *vote.UserState `json:"user"`
	}](t, messages[0])
	if response.Public.AnnouncementVersion != "7" {
		t.Fatalf("expected announcement version in snapshot, got %+v", response.Public)
	}

	encoded, err := sonic.Marshal(response.Public)
	if err != nil {
		t.Fatalf("marshal public snapshot: %v", err)
	}
	if strings.Contains(string(encoded), "\"bossLoot\"") {
		t.Fatalf("expected websocket public snapshot to omit bossLoot, got %s", string(encoded))
	}
	if strings.Contains(string(encoded), "\"latestAnnouncement\"") {
		t.Fatalf("expected websocket public snapshot to omit latestAnnouncement, got %s", string(encoded))
	}
}

func TestRealtimeSessionClickReturnsAckAndPublishesDeltas(t *testing.T) {
	store := &mockStore{
		state: vote.State{
			Leaderboard: []vote.LeaderboardEntry{
				{Rank: 1, Nickname: "阿明", ClickCount: 4},
			},
			UserStats:   &vote.UserStats{Nickname: "阿明", ClickCount: 4},
			CombatStats: vote.CombatStats{EffectiveIncrement: 2},
		},
		result: vote.ClickResult{
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
			Delta    int64 `json:"delta"`
			Critical bool  `json:"critical"`
		} `json:"payload"`
	}](t, messages[0])
	if ack.Type != realtimeMessageTypeClickAck {
		t.Fatalf("expected click_ack, got %+v", ack)
	}
	if ack.Payload.Delta != 1 || ack.Payload.Critical {
		t.Fatalf("unexpected click ack payload: %+v", ack.Payload)
	}

	readHubEventSet(t, sseClient, map[string]struct{}{
		events.PublicStateEventName: {},
	})
	readHubEventSet(t, wsClient, map[string]struct{}{
		events.PublicStateEventName: {},
	})
	assertNoHubEventByName(t, sseClient, events.UserStateEventName)
	assertNoHubEventByName(t, wsClient, events.UserStateEventName)

	if len(publisher.changes) != 1 || publisher.changes[0].Type != vote.StateChangeButtonClicked {
		t.Fatalf("expected one click change, got %+v", publisher.changes)
	}
}

func TestRealtimeSessionBossPartClickPublishesBroadcastUserAll(t *testing.T) {
	store := &mockStore{
		state: vote.State{
			Boss: &vote.Boss{
				ID:        "boss-1",
				Name:      "木桩王",
				Status:    "active",
				MaxHP:     100,
				CurrentHP: 40,
			},
		},
		result: vote.ClickResult{
			Delta:    1,
			Critical: false,
			Boss: &vote.Boss{
				ID:        "boss-1",
				Name:      "木桩王",
				Status:    "active",
				MaxHP:     100,
				CurrentHP: 39,
			},
			MyBossStats: &vote.BossUserStats{
				Nickname: "阿明",
				Damage:   61,
			},
			BroadcastUserAll: true,
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
		return session.handleMessage(context.Background(), []byte(`{"type":"click","slug":"boss-part:1-2"}`), send)
	})
	if len(messages) != 1 {
		t.Fatalf("expected one realtime message, got %d", len(messages))
	}

	ack := decodeRealtimeMessage[struct {
		Type    string `json:"type"`
		Payload struct {
			Delta    int64 `json:"delta"`
			Critical bool  `json:"critical"`
		} `json:"payload"`
	}](t, messages[0])
	if ack.Type != realtimeMessageTypeClickAck {
		t.Fatalf("expected click_ack, got %+v", ack)
	}
	if ack.Payload.Delta != 1 {
		t.Fatalf("unexpected boss click ack payload: %+v", ack.Payload)
	}
	sseUserEvent := readHubPublicAndUserEvent(t, sseClient)
	wsUserEvent := readHubPublicAndUserEvent(t, wsClient)
	for _, eventPayload := range []string{string(sseUserEvent.Payload), string(wsUserEvent.Payload)} {
		if strings.Contains(eventPayload, "\"inventory\"") {
			t.Fatalf("expected broadcast user_state to omit inventory, got %s", eventPayload)
		}
		if strings.Contains(eventPayload, "\"gems\"") {
			t.Fatalf("expected broadcast user_state to omit gems, got %s", eventPayload)
		}
		if strings.Contains(eventPayload, "\"lastReward\"") {
			t.Fatalf("expected broadcast user_state to omit lastReward, got %s", eventPayload)
		}
	}

	if len(publisher.changes) != 1 {
		t.Fatalf("expected one published change, got %+v", publisher.changes)
	}
	if publisher.changes[0].Type != vote.StateChangeButtonClicked {
		t.Fatalf("expected button_clicked, got %+v", publisher.changes[0])
	}
	if !publisher.changes[0].BroadcastUserAll {
		t.Fatalf("expected BroadcastUserAll to be true, got %+v", publisher.changes[0])
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
