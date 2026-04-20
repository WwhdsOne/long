package events

import (
	"testing"

	"long/internal/vote"
)

func TestHubBroadcastsPublicAndMatchingUserEvents(t *testing.T) {
	hub := NewHub()

	aming, unsubscribeAming := hub.Subscribe("阿明")
	defer unsubscribeAming()
	xiaohong, unsubscribeXiaohong := hub.Subscribe("小红")
	defer unsubscribeXiaohong()

	if err := hub.BroadcastPublic(vote.Snapshot{
		Buttons: []vote.Button{
			{Key: "feel", Label: "有感觉吗", Count: 3, Enabled: true},
		},
	}); err != nil {
		t.Fatalf("broadcast public snapshot: %v", err)
	}

	amingEvent := <-aming
	if amingEvent.Name != "public_state" {
		t.Fatalf("expected public_state for 阿明, got %+v", amingEvent)
	}

	xiaohongEvent := <-xiaohong
	if xiaohongEvent.Name != "public_state" {
		t.Fatalf("expected public_state for 小红, got %+v", xiaohongEvent)
	}

	if err := hub.BroadcastUser("阿明", vote.UserState{
		UserStats:   &vote.UserStats{Nickname: "阿明", ClickCount: 8},
		CombatStats: vote.CombatStats{EffectiveIncrement: 3},
	}); err != nil {
		t.Fatalf("broadcast user state: %v", err)
	}

	amingEvent = <-aming
	if amingEvent.Name != "user_state" {
		t.Fatalf("expected user_state for 阿明, got %+v", amingEvent)
	}

	select {
	case unexpected := <-xiaohong:
		t.Fatalf("expected no user event for 小红, got %+v", unexpected)
	default:
	}
}
