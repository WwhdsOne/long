package events

import (
	"context"
	"strings"
	"testing"

	"long/internal/vote"
)

type dispatcherTestReader struct {
	snapshot vote.Snapshot
}

func (r *dispatcherTestReader) GetSnapshot(context.Context) (vote.Snapshot, error) {
	return r.snapshot, nil
}

func (r *dispatcherTestReader) GetUserState(context.Context, string) (vote.UserState, error) {
	return vote.UserState{}, nil
}

func (r *dispatcherTestReader) GetBossResources(context.Context) (vote.BossResources, error) {
	return vote.BossResources{}, nil
}

func TestDispatcherHandleChangeBroadcastsSlimPublicDelta(t *testing.T) {
	reader := &dispatcherTestReader{
		snapshot: vote.Snapshot{
			TotalVotes: 12,
			Leaderboard: []vote.LeaderboardEntry{
				{Rank: 1, Nickname: "阿明", ClickCount: 12},
			},
			BossLeaderboard: []vote.BossLeaderboardEntry{
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

	if err := dispatcher.HandleChange(context.Background(), vote.StateChange{Type: vote.StateChangeButtonClicked}); err != nil {
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
		snapshot: vote.Snapshot{
			TotalVotes: 12,
			Leaderboard: []vote.LeaderboardEntry{
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
