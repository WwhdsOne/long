package events

import (
	"context"
	"strings"
	"testing"
	"time"

	"long/internal/core"
)

type dispatcherTestReader struct {
	snapshot          core.Snapshot
	userState         core.UserState
	realtimeUserState core.UserState
	roomList          core.RoomList
	fullCalls         int
	realtimeCalls     int
}

func (r *dispatcherTestReader) GetSnapshot(context.Context) (core.Snapshot, error) {
	return r.snapshot, nil
}

func (r *dispatcherTestReader) GetUserState(context.Context, string) (core.UserState, error) {
	r.fullCalls++
	return r.userState, nil
}

func (r *dispatcherTestReader) GetRealtimeUserState(context.Context, string) (core.UserState, error) {
	r.realtimeCalls++
	return r.realtimeUserState, nil
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
	if strings.Contains(payload, `"bossLeaderboard"`) {
		t.Fatalf("expected slim public delta to move bossLeaderboard out of public_state, got %s", payload)
	}
	assertNoEventWithin(t, client, 50*time.Millisecond, "expected click change to skip public_meta broadcast")
	assertNoEventWithin(t, client, 50*time.Millisecond, "expected click change to skip room_state broadcast")
}

func TestDispatcherHandleBossChangeBroadcastsRoomState(t *testing.T) {
	reader := &dispatcherTestReader{
		snapshot: core.Snapshot{
			RoomID: "2",
			Boss: &core.Boss{
				ID:        "boss-1",
				Name:      "木桩王",
				Status:    "active",
				MaxHP:     100,
				CurrentHP: 80,
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

	if err := dispatcher.HandleChange(context.Background(), core.StateChange{Type: core.StateChangeBossChanged}); err != nil {
		t.Fatalf("handle boss change: %v", err)
	}

	_ = readEventByName(t, client, PublicStateEventName)
	metaEvent := readEventByName(t, client, PublicMetaEventName)
	if !strings.Contains(string(metaEvent.Payload), `"bossLeaderboard":[]`) {
		t.Fatalf("expected boss change to carry public_meta payload, got %s", string(metaEvent.Payload))
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
	if strings.Contains(payload, `"leaderboard"`) {
		t.Fatalf("expected leaderboard to move out of public_state, got %s", payload)
	}

	metaEvent := readEventByName(t, client, PublicMetaEventName)
	metaPayload := string(metaEvent.Payload)
	if !strings.Contains(metaPayload, `"leaderboard":[{"rank":1,"nickname":"阿明","clickCount":12}]`) {
		t.Fatalf("expected full public_meta payload with leaderboard, got %s", metaPayload)
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
	if reader.fullCalls != 1 || reader.realtimeCalls != 0 {
		t.Fatalf("expected equipment change to use full user state path, got full=%d realtime=%d", reader.fullCalls, reader.realtimeCalls)
	}
}

func TestDispatcherHandleMessageChangeUsesRealtimeUserStateAndKeepsTalentFields(t *testing.T) {
	reader := &dispatcherTestReader{
		userState: core.UserState{
			UserStats: &core.UserStats{Nickname: "阿明", ClickCount: 9},
			Loadout: core.Loadout{
				Weapon: &core.InventoryItem{ItemID: "iron-sword", Name: "铁剑", Slot: "weapon"},
			},
			CombatStats: core.CombatStats{AttackPower: 128, EffectiveIncrement: 7},
		},
		realtimeUserState: core.UserState{
			UserStats:         &core.UserStats{Nickname: "阿明", ClickCount: 9},
			TalentEvents:      []core.TalentTriggerEvent{{TalentID: "crit_bleed", Name: "致命出血", EffectType: "bleed", Message: "出血结算"}},
			TalentCombatState: &core.TalentCombatState{},
		},
	}
	cache := NewCache(reader)
	hub := NewHub()
	dispatcher := NewDispatcher(cache, hub, 1)

	client, unsubscribe := hub.Subscribe("阿明")
	defer unsubscribe()
	_ = readEventByName(t, client, OnlineCountEventName)

	if err := dispatcher.HandleChange(context.Background(), core.StateChange{
		Type:      core.StateChangeMessageCreated,
		Nickname:  "阿明",
		Timestamp: 1,
	}); err != nil {
		t.Fatalf("handle message change: %v", err)
	}

	event := readEventByName(t, client, UserStateEventName)
	payload := string(event.Payload)
	if !strings.Contains(payload, `"talentEvents":[{"talentId":"crit_bleed"`) {
		t.Fatalf("expected realtime path to keep talent events, got %s", payload)
	}
	if !strings.Contains(payload, `"talentCombatState":{`) {
		t.Fatalf("expected realtime path to keep talent combat state, got %s", payload)
	}
	if strings.Contains(payload, `"loadout"`) || strings.Contains(payload, `"combatStats"`) {
		t.Fatalf("expected slim realtime payload without profile fields, got %s", payload)
	}
	if reader.fullCalls != 0 || reader.realtimeCalls != 1 {
		t.Fatalf("expected message change to use realtime user state path, got full=%d realtime=%d", reader.fullCalls, reader.realtimeCalls)
	}
}

func TestDispatcherThrottlePublicBroadcastWithinWindow(t *testing.T) {
	reader := &dispatcherTestReader{
		snapshot: core.Snapshot{
			TotalVotes: 12,
		},
	}
	cache := NewCache(reader)
	hub := NewHub()
	dispatcher := NewDispatcher(cache, hub, 20)

	client, unsubscribe := hub.Subscribe("")
	defer unsubscribe()
	_ = readEventByName(t, client, OnlineCountEventName)

	if err := dispatcher.HandleChange(context.Background(), core.StateChange{Type: core.StateChangeButtonClicked}); err != nil {
		t.Fatalf("first click change: %v", err)
	}
	_ = readEventByName(t, client, PublicStateEventName)

	if err := dispatcher.HandleChange(context.Background(), core.StateChange{Type: core.StateChangeButtonClicked}); err != nil {
		t.Fatalf("second click change: %v", err)
	}
	if err := dispatcher.HandleChange(context.Background(), core.StateChange{Type: core.StateChangeButtonClicked}); err != nil {
		t.Fatalf("third click change: %v", err)
	}

	assertNoEventWithin(t, client, 10*time.Millisecond, "expected throttle window to suppress immediate extra public_state")
	_ = readEventByName(t, client, PublicStateEventName)
	assertNoEventWithin(t, client, 30*time.Millisecond, "expected only one deferred public_state within throttle window")
}

func assertNoEventWithin(t *testing.T, ch <-chan ServerEvent, wait time.Duration, message string) {
	t.Helper()

	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case event := <-ch:
		t.Fatalf("%s, got %+v", message, event)
	case <-timer.C:
	}
}
