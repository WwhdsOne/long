package events

import (
	"context"

	"long/internal/vote"
)

// Dispatcher 将业务变更转换成公共态和个人态推送。
type Dispatcher struct {
	cache *Cache
	hub   *Hub
}

// NewDispatcher 创建一个实时分发器。
func NewDispatcher(cache *Cache, hub *Hub) *Dispatcher {
	return &Dispatcher{
		cache: cache,
		hub:   hub,
	}
}

// HandleChange 刷新受影响的缓存并推送到对应订阅者。
func (d *Dispatcher) HandleChange(ctx context.Context, change vote.StateChange) error {
	if d == nil || d.cache == nil || d.hub == nil {
		return nil
	}

	if affectsPublicState(change.Type) {
		snapshot, err := d.cache.RefreshSnapshot(ctx)
		if err != nil {
			return err
		}
		if err := d.hub.BroadcastPublic(snapshot); err != nil {
			return err
		}
	}

	targetNicknames := userTargetsForChange(change, d.hub.ActiveNicknames())
	if len(targetNicknames) == 0 {
		return nil
	}

	userStates, err := d.cache.RefreshUsers(ctx, targetNicknames)
	if err != nil {
		return err
	}
	for nickname, userState := range userStates {
		if err := d.hub.BroadcastUser(nickname, userState); err != nil {
			return err
		}
	}

	return nil
}

func affectsPublicState(changeType vote.StateChangeType) bool {
	switch changeType {
	case vote.StateChangeButtonClicked,
		vote.StateChangeBossChanged,
		vote.StateChangeAnnouncementChanged,
		vote.StateChangeButtonMetaChanged,
		vote.StateChangeEquipmentMetaChanged:
		return true
	default:
		return false
	}
}

func userTargetsForChange(change vote.StateChange, activeNicknames []string) []string {
	if change.BroadcastUserAll {
		return activeNicknames
	}
	if change.Nickname == "" {
		return nil
	}
	return []string{change.Nickname}
}
