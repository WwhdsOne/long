package events

import (
	"strings"
	"testing"
	"time"

	"long/internal/core"
)

func TestHubBroadcastsPublicAndMatchingUserEvents(t *testing.T) {
	hub := NewHub()

	aming, unsubscribeAming := hub.Subscribe("阿明")
	defer unsubscribeAming()
	xiaohong, unsubscribeXiaohong := hub.Subscribe("小红")
	defer unsubscribeXiaohong()

	if err := hub.BroadcastPublic(core.Snapshot{}, true); err != nil {
		t.Fatalf("broadcast public snapshot: %v", err)
	}

	amingEvent := readEventByName(t, aming, PublicStateEventName)
	if amingEvent.Name != PublicStateEventName {
		t.Fatalf("expected public_state for 阿明, got %+v", amingEvent)
	}

	xiaohongEvent := readEventByName(t, xiaohong, PublicStateEventName)
	if xiaohongEvent.Name != PublicStateEventName {
		t.Fatalf("expected public_state for 小红, got %+v", xiaohongEvent)
	}

	if err := hub.BroadcastUser("阿明", core.UserState{
		UserStats:                          &core.UserStats{Nickname: "阿明", ClickCount: 8},
		CombatStats:                        core.CombatStats{EffectiveIncrement: 3},
		EquippedBattleClickSkinID:          "skin-basic",
		EquippedBattleClickCursorImagePath: "https://example.com/basic.png",
	}); err != nil {
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

	select {
	case unexpected := <-xiaohong:
		t.Fatalf("expected no user event for 小红, got %+v", unexpected)
	default:
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
