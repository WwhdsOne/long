package events

import (
	"context"
	"strings"
	"testing"

	"long/internal/core"
)

type dispatcherTestReader struct {
	snapshot core.Snapshot
}

func (r *dispatcherTestReader) GetSnapshot(context.Context) (core.Snapshot, error) {
	return r.snapshot, nil
}

func (r *dispatcherTestReader) GetUserState(context.Context, string) (core.UserState, error) {
	return core.UserState{}, nil
}

func (r *dispatcherTestReader) GetBossResources(context.Context) (core.BossResources, error) {
	return core.BossResources{}, nil
}

func TestDispatcherHandleChangeBroadcastsSlimPublicDelta(t *testing.T) {
	reader := &dispatcherTestReader{
		snapshot: core.Snapshot{
			TotalVotes: 12,
			Leaderboard: []core.LeaderboardEntry{
				{Rank: 1, Nickname: "阿明", ClickCount: 12},
			},
			BossLeaderboard: []core.BossLeaderboardEntry{
				{Rank: 1, Nickname: "阿明", Damage: 88},
			},
		},
	}
	cache := NewCache(reader)
	hub := NewHub()
	dispatcher := NewDispatcher(cache, hub, 1)

	client, unsubscribe := hub.Subscribe("阿明")
	defer unsubscribe()
	_ = readEventByName(t, client, OnlineCountEventName)

	if err := dispatcher.HandleChange(context.Background(), core.StateChange{Type: core.StateChangeButtonClicked}); err != nil {
		t.Fatalf("handle change: %v", err)
	}

	event := readEventByName(t, client, PublicStateEventName)
	payload := string(event.Payload)
	if strings.Contains(payload, `"leaderboard"`) {
		t.Fatalf("expected slim public delta without leaderboard, got %s", payload)
	}
	if !strings.Contains(payload, `"bossLeaderboard"`) {
		t.Fatalf("expected slim public delta to keep bossLeaderboard, got %s", payload)
	}
}

func TestDispatcherBroadcastLeaderboardIncludesLeaderboard(t *testing.T) {
	reader := &dispatcherTestReader{
		snapshot: core.Snapshot{
			TotalVotes: 12,
			Leaderboard: []core.LeaderboardEntry{
				{Rank: 1, Nickname: "阿明", ClickCount: 12},
			},
		},
	}
	cache := NewCache(reader)
	hub := NewHub()
	dispatcher := NewDispatcher(cache, hub)

	client, unsubscribe := hub.Subscribe("阿明")
	defer unsubscribe()
	_ = readEventByName(t, client, OnlineCountEventName)

	if err := dispatcher.BroadcastLeaderboard(context.Background()); err != nil {
		t.Fatalf("broadcast leaderboard: %v", err)
	}

	event := readEventByName(t, client, PublicStateEventName)
	payload := string(event.Payload)
	if !strings.Contains(payload, `"leaderboard":[{"rank":1,"nickname":"阿明","clickCount":12}]`) {
		t.Fatalf("expected full public payload with leaderboard, got %s", payload)
	}
}
