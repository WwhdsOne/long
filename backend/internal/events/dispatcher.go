package events

import (
	"context"
	"sync"
	"time"

	"long/internal/core"
)

// Dispatcher 将业务变更转换成公共态和个人态推送。
type Dispatcher struct {
	cache      *Cache
	hub        *Hub
	debounceMs int

	mu                sync.Mutex
	publicDirty       bool
	publicMeta        bool
	publicRoomState   bool
	publicTimer       *time.Timer
	lastPublicFlushAt time.Time
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
func (d *Dispatcher) HandleChange(ctx context.Context, change core.StateChange) error {
	if d == nil || d.cache == nil || d.hub == nil {
		return nil
	}

	if affectsPublicState(change.Type) {
		d.schedulePublicFlush(change.Type)
	}

	targetNicknames := userTargetsForChange(change, d.hub.ActiveNicknames())
	if len(targetNicknames) == 0 {
		return nil
	}

	userStates, err := d.cache.RefreshUsers(ctx, targetNicknames)
	if err != nil {
		return err
	}
	includeProfile := shouldBroadcastUserProfile(change.Type)
	for nickname, userState := range userStates {
		if err := d.hub.BroadcastUser(nickname, userState, includeProfile); err != nil {
			return err
		}
	}

	return nil
}

func (d *Dispatcher) schedulePublicFlush(changeType core.StateChangeType) {
	d.mu.Lock()
	d.publicDirty = true
	d.publicMeta = d.publicMeta || shouldBroadcastPublicMeta(changeType)
	d.publicRoomState = d.publicRoomState || shouldBroadcastRoomState(changeType)

	window := time.Duration(d.debounceMs) * time.Millisecond
	now := time.Now()
	if d.lastPublicFlushAt.IsZero() || now.Sub(d.lastPublicFlushAt) >= window {
		d.lastPublicFlushAt = now
		d.mu.Unlock()
		go d.flushPublic()
		return
	}

	if d.publicTimer == nil {
		delay := window - now.Sub(d.lastPublicFlushAt)
		d.publicTimer = time.AfterFunc(delay, d.flushPublic)
	}
	d.mu.Unlock()
}

func (d *Dispatcher) flushPublic() {
	d.mu.Lock()
	if !d.publicDirty {
		d.publicTimer = nil
		d.mu.Unlock()
		return
	}
	d.publicDirty = false
	includeMeta := d.publicMeta
	d.publicMeta = false
	includeRoomState := d.publicRoomState
	d.publicRoomState = false
	d.publicTimer = nil
	d.lastPublicFlushAt = time.Now()
	d.mu.Unlock()

	ctx := context.Background()
	_ = d.broadcastPublic(ctx, includeMeta, includeRoomState, false)
}

// BroadcastLeaderboard 主动广播一份带点击总榜的公共态。
func (d *Dispatcher) BroadcastLeaderboard(ctx context.Context) error {
	if d == nil || d.cache == nil || d.hub == nil {
		return nil
	}

	return d.broadcastPublic(ctx, true, true, true)
}

func (d *Dispatcher) broadcastPublic(ctx context.Context, includeMeta bool, includeRoomState bool, includeLeaderboard bool) error {
	snapshot, err := d.cache.RefreshSnapshot(ctx)
	if err != nil {
		return err
	}
	if snapshot.Boss != nil {
		_, _ = d.cache.RefreshBossResources(ctx)
	}
	if err := d.hub.BroadcastPublic(snapshot); err != nil {
		return err
	}
	if includeMeta {
		if err := d.hub.BroadcastPublicMeta(snapshot, includeLeaderboard); err != nil {
			return err
		}
	}

	for _, nickname := range d.hub.ActiveNicknames() {
		snapshot, err := d.cache.GetSnapshotForNickname(ctx, nickname)
		if err != nil {
			return err
		}
		if err := d.hub.BroadcastPublicTo(nickname, snapshot); err != nil {
			return err
		}
		if includeMeta {
			if err := d.hub.BroadcastPublicMetaTo(nickname, snapshot, includeLeaderboard); err != nil {
				return err
			}
		}
		if includeRoomState {
			rooms, err := d.cache.ListRooms(ctx, nickname)
			if err != nil {
				return err
			}
			if err := d.hub.BroadcastRoomState(nickname, rooms); err != nil {
				return err
			}
		}
	}
	return nil
}

func affectsPublicState(changeType core.StateChangeType) bool {
	switch changeType {
	case core.StateChangeButtonClicked,
		core.StateChangeBossChanged,
		core.StateChangeAnnouncementChanged,
		core.StateChangeEquipmentMetaChanged:
		return true
	default:
		return false
	}
}

func userTargetsForChange(change core.StateChange, activeNicknames []string) []string {
	if change.BroadcastUserAll {
		return activeNicknames
	}
	if change.Type == core.StateChangeButtonClicked {
		return nil
	}
	if change.Nickname == "" {
		return nil
	}
	return []string{change.Nickname}
}

func shouldBroadcastUserProfile(changeType core.StateChangeType) bool {
	switch changeType {
	case core.StateChangeEquipmentChanged, core.StateChangeEquipmentMetaChanged:
		return true
	default:
		return false
	}
}

func shouldBroadcastPublicMeta(changeType core.StateChangeType) bool {
	switch changeType {
	case core.StateChangeBossChanged, core.StateChangeAnnouncementChanged:
		return true
	default:
		return false
	}
}

func shouldBroadcastRoomState(changeType core.StateChangeType) bool {
	switch changeType {
	case core.StateChangeButtonClicked:
		return false
	default:
		return true
	}
}
