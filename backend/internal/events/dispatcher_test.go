package events

import (
	"context"
	"strings"
	"testing"

	"long/internal/core"
)

type dispatcherTestReader struct {
	snapshot  core.Snapshot
	userState core.UserState
	roomList  core.RoomList
}

func (r *dispatcherTestReader) GetSnapshot(context.Context) (core.Snapshot, error) {
	return r.snapshot, nil
}

func (r *dispatcherTestReader) GetUserState(context.Context, string) (core.UserState, error) {
	return r.userState, nil
}

func (r *dispatcherTestReader) GetBossResources(context.Context) (core.BossResources, error) {
	return core.BossResources{}, nil
}

func (r *dispatcherTestReader) ListRooms(context.Context, string) (core.RoomList, error) {
	return r.roomList, nil
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
		roomList: core.RoomList{
			CurrentRoomID:                  "2",
			SwitchCooldownRemainingSeconds: 7,
			Rooms: []core.RoomInfo{
				{ID: "2", DisplayName: "二线", Current: true, OnlineCount: 3},
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

	roomEvent := readEventByName(t, client, RoomStateEventName)
	roomPayload := string(roomEvent.Payload)
	if !strings.Contains(roomPayload, `"currentRoomId":"2"`) || !strings.Contains(roomPayload, `"switchCooldownRemainingSeconds":7`) {
		t.Fatalf("expected room state payload, got %s", roomPayload)
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

func TestDispatcherHandleEquipmentChangeBroadcastsProfileFields(t *testing.T) {
	reader := &dispatcherTestReader{
		userState: core.UserState{
			UserStats: &core.UserStats{Nickname: "阿明", ClickCount: 9},
			Loadout: core.Loadout{
				Weapon: &core.InventoryItem{ItemID: "iron-sword", Name: "铁剑", Slot: "weapon"},
			},
			CombatStats: core.CombatStats{AttackPower: 128, EffectiveIncrement: 7},
		},
	}
	cache := NewCache(reader)
	hub := NewHub()
	dispatcher := NewDispatcher(cache, hub, 1)

	client, unsubscribe := hub.Subscribe("阿明")
	defer unsubscribe()
	_ = readEventByName(t, client, OnlineCountEventName)

	if err := dispatcher.HandleChange(context.Background(), core.StateChange{
		Type:      core.StateChangeEquipmentChanged,
		Nickname:  "阿明",
		Timestamp: 1,
	}); err != nil {
		t.Fatalf("handle equipment change: %v", err)
	}

	event := readEventByName(t, client, UserStateEventName)
	payload := string(event.Payload)
	if !strings.Contains(payload, `"loadout":{"weapon":{"itemId":"iron-sword"`) {
		t.Fatalf("expected equipment change to keep loadout, got %s", payload)
	}
	if !strings.Contains(payload, `"attackPower":128`) || !strings.Contains(payload, `"effectiveIncrement":7`) {
		t.Fatalf("expected equipment change to keep combatStats, got %s", payload)
	}
}
