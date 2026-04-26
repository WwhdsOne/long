package events

import (
	"context"
	"sync"
	"time"

	"long/internal/vote"
)

// Dispatcher 将业务变更转换成公共态和个人态推送。
type Dispatcher struct {
	cache      *Cache
	hub        *Hub
	debounceMs int

	mu          sync.Mutex
	publicDirty bool
	publicTimer *time.Timer
}

// NewDispatcher 创建一个实时分发器。
func NewDispatcher(cache *Cache, hub *Hub, debounceMs ...int) *Dispatcher {
	ms := 50
	if len(debounceMs) > 0 && debounceMs[0] > 0 {
		ms = debounceMs[0]
	}
	return &Dispatcher{
		cache:      cache,
		hub:        hub,
		debounceMs: ms,
	}
}

// HandleChange 刷新受影响的缓存并推送到对应订阅者。
func (d *Dispatcher) HandleChange(ctx context.Context, change vote.StateChange) error {
	if d == nil || d.cache == nil || d.hub == nil {
		return nil
	}

	if affectsPublicState(change.Type) {
		d.mu.Lock()
		d.publicDirty = true
		if d.publicTimer != nil {
			d.publicTimer.Stop()
		}
		d.publicTimer = time.AfterFunc(time.Duration(d.debounceMs)*time.Millisecond, d.flushPublic)
		d.mu.Unlock()
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

func (d *Dispatcher) flushPublic() {
	d.mu.Lock()
	if !d.publicDirty {
		d.mu.Unlock()
		return
	}
	d.publicDirty = false
	d.publicTimer = nil
	d.mu.Unlock()

	ctx := context.Background()
	snapshot, err := d.cache.RefreshSnapshot(ctx)
	if err != nil {
		return
	}
	if snapshot.Boss != nil {
		_, _ = d.cache.RefreshBossResources(ctx)
	}
	_ = d.hub.BroadcastPublic(snapshot)
}

func affectsPublicState(changeType vote.StateChangeType) bool {
	switch changeType {
	case vote.StateChangeButtonClicked,
		vote.StateChangeBossChanged,
		vote.StateChangeAnnouncementChanged,
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
	if change.Type == vote.StateChangeButtonClicked {
		return nil
	}
	if change.Nickname == "" {
		return nil
	}
	return []string{change.Nickname}
}
