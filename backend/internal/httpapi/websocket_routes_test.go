package httpapi

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/hertz-contrib/websocket"

	"long/internal/core"
	"long/internal/events"
	"long/internal/ratelimit"
	"long/internal/realtimepb"
)

type dispatchingChangePublisher struct {
	dispatcher *events.Dispatcher
	changes    []core.StateChange
}

func (p *dispatchingChangePublisher) PublishChange(ctx context.Context, change core.StateChange) error {
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

type capturedRealtimeFrame struct {
	messageType int
	payload     []byte
}

func captureRealtimeFrames(t *testing.T, fn func(func(realtimeOutboundFrame) error) error) []capturedRealtimeFrame {
	t.Helper()

	var frames []capturedRealtimeFrame
	err := fn(func(frame realtimeOutboundFrame) error {
		frames = append(frames, capturedRealtimeFrame{
			messageType: frame.messageType,
			payload:     append([]byte(nil), frame.payload...),
		})
		return nil
	})
	if err != nil {
		t.Fatalf("capture realtime messages: %v", err)
	}

	return frames
}

func captureRealtimeMessages(t *testing.T, fn func(func(realtimeOutboundFrame) error) error) [][]byte {
	t.Helper()

	frames := captureRealtimeFrames(t, fn)
	messages := make([][]byte, 0, len(frames))
	for _, frame := range frames {
		messages = append(messages, frame.payload)
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

func encodeRealtimeBinaryClickRequestForTest(t *testing.T, slug string, comboCount int64) []byte {
	t.Helper()

	payload, err := packRealtimeBinaryMessage(realtimeBinaryTypeClickRequest, &realtimepb.ClickRequest{
		Slug:       slug,
		ComboCount: comboCount,
	})
	if err != nil {
		t.Fatalf("encode realtime binary click request: %v", err)
	}
	return payload
}

func decodeRealtimeBinaryClickAckForTest(t *testing.T, payload []byte) *realtimepb.ClickAck {
	t.Helper()

	message := &realtimepb.ClickAck{}
	if err := unpackRealtimeBinaryMessage(payload, realtimeBinaryTypeClickAck, message); err != nil {
		t.Fatalf("decode realtime binary click ack: %v", err)
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
		state: core.State{
			Leaderboard: []core.LeaderboardEntry{
				{Rank: 1, Nickname: "阿明", ClickCount: 3},
			},
			UserStats:   &core.UserStats{Nickname: "阿明", ClickCount: 3},
			Inventory:   []core.InventoryItem{{ItemID: "wood-sword", Name: "木剑", Quantity: 1}},
			CombatStats: core.CombatStats{EffectiveIncrement: 2},
		},
	}
	session := newRealtimeSession(realtimeSessionOptions{
		stateView:            store,
		store:                store,
		hub:                  events.NewHub(),
		authenticatorEnabled: false,
		clientID:             "127.0.0.1",
	})

	messages := captureRealtimeMessages(t, func(send func(realtimeOutboundFrame) error) error {
		return session.handleMessage(context.Background(), websocket.TextMessage, []byte(`{"type":"hello","nickname":"阿明"}`), send)
	})
	if len(messages) != 1 {
		t.Fatalf("expected one realtime message, got %d", len(messages))
	}

	var response struct {
		Type   string          `json:"type"`
		Public core.Snapshot   `json:"public"`
		User   *core.UserState `json:"user"`
	}
	decoded := decodeRealtimeMessage[struct {
		Type   string          `json:"type"`
		Public core.Snapshot   `json:"public"`
		User   *core.UserState `json:"user"`
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
		state: core.State{},
	}
	session := newRealtimeSession(realtimeSessionOptions{
		stateView:             store,
		store:                 store,
		hub:                   events.NewHub(),
		authenticatorEnabled:  true,
		authenticatedNickname: "",
		clientID:              "127.0.0.1",
	})

	messages := captureRealtimeMessages(t, func(send func(realtimeOutboundFrame) error) error {
		return session.handleMessage(context.Background(), websocket.TextMessage, []byte(`{"type":"hello","nickname":"阿明"}`), send)
	})
	if len(messages) != 1 {
		t.Fatalf("expected one realtime message, got %d", len(messages))
	}

	response := decodeRealtimeMessage[struct {
		Type string          `json:"type"`
		User *core.UserState `json:"user"`
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
		snapshot: core.Snapshot{
			Leaderboard: []core.LeaderboardEntry{
				{Rank: 1, Nickname: "阿明", ClickCount: 3},
			},
			AnnouncementVersion: "7",
		},
		state: core.State{
			UserStats:   &core.UserStats{Nickname: "阿明", ClickCount: 3},
			Inventory:   []core.InventoryItem{{ItemID: "wood-sword", Name: "木剑", Quantity: 1}},
			CombatStats: core.CombatStats{EffectiveIncrement: 2},
		},
	}
	session := newRealtimeSession(realtimeSessionOptions{
		stateView:            store,
		store:                store,
		hub:                  events.NewHub(),
		authenticatorEnabled: false,
		clientID:             "127.0.0.1",
	})

	messages := captureRealtimeMessages(t, func(send func(realtimeOutboundFrame) error) error {
		return session.handleMessage(context.Background(), websocket.TextMessage, []byte(`{"type":"hello","nickname":"阿明"}`), send)
	})
	if len(messages) != 1 {
		t.Fatalf("expected one realtime message, got %d", len(messages))
	}

	response := decodeRealtimeMessage[struct {
		Type   string          `json:"type"`
		Public core.Snapshot   `json:"public"`
		User   *core.UserState `json:"user"`
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
		state: core.State{
			Leaderboard: []core.LeaderboardEntry{
				{Rank: 1, Nickname: "阿明", ClickCount: 4},
			},
			UserStats:   &core.UserStats{Nickname: "阿明", ClickCount: 4},
			CombatStats: core.CombatStats{EffectiveIncrement: 2},
		},
		result: core.ClickResult{
			Delta:                1,
			Critical:             false,
			MyBossDamage:         61,
			BossLeaderboardCount: 2,
			UserStats: core.UserStats{
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

	_ = captureRealtimeMessages(t, func(send func(realtimeOutboundFrame) error) error {
		return session.handleMessage(context.Background(), websocket.TextMessage, []byte(`{"type":"hello","nickname":"阿明"}`), send)
	})

	frames := captureRealtimeFrames(t, func(send func(realtimeOutboundFrame) error) error {
		return session.handleMessage(context.Background(), websocket.BinaryMessage, encodeRealtimeBinaryClickRequestForTest(t, "feel", 0), send)
	})
	if len(frames) != 1 {
		t.Fatalf("expected one realtime message, got %d", len(frames))
	}

	if frames[0].messageType != websocket.BinaryMessage {
		t.Fatalf("expected binary click ack, got %d", frames[0].messageType)
	}
	ack := decodeRealtimeBinaryClickAckForTest(t, frames[0].payload)
	if ack.GetDelta() != 1 || ack.GetCritical() {
		t.Fatalf("unexpected click ack payload: %+v", ack)
	}
	if ack.GetMyBossDamage() != 61 || ack.GetBossLeaderboardCount() != 2 {
		t.Fatalf("expected realtime ack to carry boss summary fields, got %+v", ack)
	}

	readHubEventSet(t, sseClient, map[string]struct{}{
		events.PublicStateEventName: {},
	})
	readHubEventSet(t, wsClient, map[string]struct{}{
		events.PublicStateEventName: {},
	})
	assertNoHubEventByName(t, sseClient, events.UserStateEventName)
	assertNoHubEventByName(t, wsClient, events.UserStateEventName)

	if len(publisher.changes) != 1 || publisher.changes[0].Type != core.StateChangeButtonClicked {
		t.Fatalf("expected one click change, got %+v", publisher.changes)
	}
}

func TestRealtimeSessionBossPartClickPublishesBroadcastUserAll(t *testing.T) {
	store := &mockStore{
		state: core.State{
			Boss: &core.Boss{
				ID:        "boss-1",
				Name:      "木桩王",
				Status:    "active",
				MaxHP:     100,
				CurrentHP: 40,
			},
		},
		result: core.ClickResult{
			Delta:                1,
			Critical:             false,
			MyBossDamage:         61,
			BossLeaderboardCount: 1,
			Boss: &core.Boss{
				ID:        "boss-1",
				Name:      "木桩王",
				Status:    "active",
				MaxHP:     100,
				CurrentHP: 39,
			},
			MyBossStats: &core.BossUserStats{
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

	_ = captureRealtimeMessages(t, func(send func(realtimeOutboundFrame) error) error {
		return session.handleMessage(context.Background(), websocket.TextMessage, []byte(`{"type":"hello","nickname":"阿明"}`), send)
	})
	frames := captureRealtimeFrames(t, func(send func(realtimeOutboundFrame) error) error {
		return session.handleMessage(context.Background(), websocket.BinaryMessage, encodeRealtimeBinaryClickRequestForTest(t, "boss-part:1-2", 0), send)
	})
	if len(frames) != 1 {
		t.Fatalf("expected one realtime message, got %d", len(frames))
	}

	if frames[0].messageType != websocket.BinaryMessage {
		t.Fatalf("expected binary click ack, got %d", frames[0].messageType)
	}
	ack := decodeRealtimeBinaryClickAckForTest(t, frames[0].payload)
	if ack.GetDelta() != 1 {
		t.Fatalf("unexpected boss click ack payload: %+v", ack)
	}
	if ack.GetMyBossDamage() != 61 || ack.GetBossLeaderboardCount() != 1 {
		t.Fatalf("expected boss summary in boss click ack, got %+v", ack)
	}
	sseUserEvent := readHubPublicAndUserEvent(t, sseClient)
	wsUserEvent := readHubPublicAndUserEvent(t, wsClient)
	for _, eventPayload := range []string{string(sseUserEvent.Payload), string(wsUserEvent.Payload)} {
		if strings.Contains(eventPayload, "\"inventory\"") {
			t.Fatalf("expected broadcast user_state to omit inventory, got %s", eventPayload)
		}
		if strings.Contains(eventPayload, "\"loadout\"") {
			t.Fatalf("expected boss click user_state to omit loadout, got %s", eventPayload)
		}
		if strings.Contains(eventPayload, "\"combatStats\"") {
			t.Fatalf("expected boss click user_state to omit combatStats, got %s", eventPayload)
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
	if publisher.changes[0].Type != core.StateChangeButtonClicked {
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
			messages := captureRealtimeMessages(t, func(send func(realtimeOutboundFrame) error) error {
				return session.handleMessage(context.Background(), websocket.TextMessage, []byte(tc.raw), send)
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

	messages := captureRealtimeMessages(t, func(send func(realtimeOutboundFrame) error) error {
		return session.handleMessage(context.Background(), websocket.TextMessage, []byte(`{"type":"click","slug":"feel"}`), send)
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
	_ = captureRealtimeMessages(t, func(send func(realtimeOutboundFrame) error) error {
		return session.handleMessage(context.Background(), websocket.TextMessage, []byte(`{"type":"hello","nickname":"阿明"}`), send)
	})

	messages := captureRealtimeMessages(t, func(send func(realtimeOutboundFrame) error) error {
		return session.handleMessage(context.Background(), websocket.TextMessage, []byte(`{"type":"click","slug":"feel"}`), send)
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
