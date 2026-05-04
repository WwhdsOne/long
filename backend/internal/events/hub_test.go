package events

import (
	"strings"
	"testing"
	"time"

	"long/internal/core"
)

func TestHubBroadcastsPublicAndMatchingUserEvents(t *testing.T) {
	hub := NewHub()

	anonymous, unsubscribeAnonymous := hub.Subscribe("")
	defer unsubscribeAnonymous()
	aming, unsubscribeAming := hub.Subscribe("阿明")
	defer unsubscribeAming()
	xiaohong, unsubscribeXiaohong := hub.Subscribe("小红")
	defer unsubscribeXiaohong()

	_ = readEventByName(t, anonymous, OnlineCountEventName)
	_ = readEventByName(t, aming, OnlineCountEventName)
	_ = readEventByName(t, xiaohong, OnlineCountEventName)
	drainEvents(anonymous)
	drainEvents(aming)
	drainEvents(xiaohong)

	if err := hub.BroadcastPublic(core.Snapshot{}, true); err != nil {
		t.Fatalf("broadcast public snapshot: %v", err)
	}

	anonymousEvent := readEventByName(t, anonymous, PublicStateEventName)
	if anonymousEvent.Name != PublicStateEventName {
		t.Fatalf("expected public_state for anonymous client, got %+v", anonymousEvent)
	}
	assertNoEvent(t, aming, "阿明")
	assertNoEvent(t, xiaohong, "小红")

	if err := hub.BroadcastPublicTo("阿明", core.Snapshot{RoomID: "2"}, true); err != nil {
		t.Fatalf("broadcast public snapshot to 阿明: %v", err)
	}

	amingEvent := readEventByName(t, aming, PublicStateEventName)
	if !strings.Contains(string(amingEvent.Payload), `"roomId":"2"`) {
		t.Fatalf("expected room 2 public state for 阿明, got %s", string(amingEvent.Payload))
	}
	assertNoEvent(t, xiaohong, "小红")

	if err := hub.BroadcastUser("阿明", core.UserState{
		UserStats:                          &core.UserStats{Nickname: "阿明", ClickCount: 8},
		CombatStats:                        core.CombatStats{EffectiveIncrement: 3},
		EquippedBattleClickSkinID:          "skin-basic",
		EquippedBattleClickCursorImagePath: "https://example.com/basic.png",
	}, false); err != nil {
		t.Fatalf("broadcast user state: %v", err)
	}

	amingEvent = readEventByName(t, aming, UserStateEventName)
	if amingEvent.Name != UserStateEventName {
		t.Fatalf("expected user_state for 阿明, got %+v", amingEvent)
	}
	payload := string(amingEvent.Payload)
	if !strings.Contains(payload, `"userStats":{"nickname":"阿明","clickCount":8}`) {
		t.Fatalf("expected user stats in slim payload, got %s", payload)
	}
	if !strings.Contains(payload, `"equippedBattleClickSkinId":"skin-basic"`) {
		t.Fatalf("expected equipped battle click skin id in slim payload, got %s", payload)
	}
	if !strings.Contains(payload, `"equippedBattleClickCursorImagePath":"https://example.com/basic.png"`) {
		t.Fatalf("expected equipped battle click cursor path in slim payload, got %s", payload)
	}
	if strings.Contains(payload, `"inventory"`) {
		t.Fatalf("expected slim user payload to omit inventory, got %s", payload)
	}
	if strings.Contains(payload, `"combatStats"`) {
		t.Fatalf("expected slim user payload to omit combatStats, got %s", payload)
	}

	select {
	case unexpected := <-xiaohong:
		t.Fatalf("expected no user event for 小红, got %+v", unexpected)
	default:
	}

	if err := hub.BroadcastRoomState("阿明", core.RoomList{
		CurrentRoomID:                  "2",
		SwitchCooldownRemainingSeconds: 5,
		Rooms: []core.RoomInfo{
			{ID: "2", DisplayName: "二线", Current: true, OnlineCount: 9},
		},
	}); err != nil {
		t.Fatalf("broadcast room state: %v", err)
	}

	roomEvent := readEventByName(t, aming, RoomStateEventName)
	if !strings.Contains(string(roomEvent.Payload), `"currentRoomId":"2"`) {
		t.Fatalf("expected room_state for 阿明, got %s", string(roomEvent.Payload))
	}
	assertNoEvent(t, anonymous, "匿名")
	assertNoEvent(t, xiaohong, "小红")
}

func assertNoEvent(t *testing.T, ch <-chan ServerEvent, label string) {
	t.Helper()

	select {
	case unexpected := <-ch:
		t.Fatalf("expected no event for %s, got %+v", label, unexpected)
	default:
	}
}

func drainEvents(ch <-chan ServerEvent) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}

func TestHubSubscribeAndUnsubscribeBroadcastOnlineCount(t *testing.T) {
	hub := NewHub()

	aming, unsubscribeAming := hub.Subscribe("阿明")
	online := readEventByName(t, aming, OnlineCountEventName)
	if string(online.Payload) != `{"count":1}` {
		t.Fatalf("expected online count 1, got %s", string(online.Payload))
	}

	xiaohong, unsubscribeXiaohong := hub.Subscribe("小红")
	online = readEventByName(t, aming, OnlineCountEventName)
	if string(online.Payload) != `{"count":2}` {
		t.Fatalf("expected online count 2 for 阿明, got %s", string(online.Payload))
	}
	online = readEventByName(t, xiaohong, OnlineCountEventName)
	if string(online.Payload) != `{"count":2}` {
		t.Fatalf("expected online count 2 for 小红, got %s", string(online.Payload))
	}

	unsubscribeXiaohong()
	online = readEventByName(t, aming, OnlineCountEventName)
	if string(online.Payload) != `{"count":1}` {
		t.Fatalf("expected online count 1 after unsubscribe, got %s", string(online.Payload))
	}

	unsubscribeAming()
}

func readEventByName(t *testing.T, ch <-chan ServerEvent, name string) ServerEvent {
	t.Helper()

	timeout := time.After(2 * time.Second)
	for {
		select {
		case event, ok := <-ch:
			if !ok {
				t.Fatalf("event channel closed while waiting for %s", name)
			}
			if event.Name == name {
				return event
			}
		case <-timeout:
			t.Fatalf("timed out waiting for event %s", name)
		}
	}
}
