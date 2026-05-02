package events

import (
	"context"
	"strings"
	"sync"

	"long/internal/core"
)

// Cache 在进程内缓存公共快照与个人态，避免每次推送都回源 Redis 聚合。
type Cache struct {
	reader             StateReader
	mu                 sync.RWMutex
	snapshot           core.Snapshot
	snapshotReady      bool
	bossResources      core.BossResources
	bossResourcesReady bool
	users              map[string]core.UserState
}

// NewCache 创建一个可作为只读视图使用的状态缓存。
func NewCache(reader StateReader) *Cache {
	return &Cache{
		reader: reader,
		users:  make(map[string]core.UserState),
	}
}

// GetSnapshot 返回缓存中的公共快照；未命中时回源加载。
func (c *Cache) GetSnapshot(ctx context.Context) (core.Snapshot, error) {
	c.mu.RLock()
	if c.snapshotReady {
		snapshot := c.snapshot
		c.mu.RUnlock()
		return snapshot, nil
	}
	c.mu.RUnlock()

	return c.RefreshSnapshot(ctx)
}

// RefreshSnapshot 强制回源刷新公共快照。
func (c *Cache) RefreshSnapshot(ctx context.Context) (core.Snapshot, error) {
	snapshot, err := c.reader.GetSnapshot(ctx)
	if err != nil {
		return core.Snapshot{}, err
	}

	c.mu.Lock()
	c.snapshot = snapshot
	c.snapshotReady = true
	if snapshot.Boss == nil {
		c.bossResources = core.BossResources{
			BossLoot: []core.BossLootEntry{},
		}
		c.bossResourcesReady = true
	} else if c.bossResources.BossID != snapshot.Boss.ID {
		c.bossResourcesReady = false
	}
	c.mu.Unlock()

	return snapshot, nil
}

// GetBossResources 返回当前 Boss 的低频公共资源；未命中时回源加载。
func (c *Cache) GetBossResources(ctx context.Context) (core.BossResources, error) {
	c.mu.RLock()
	if c.bossResourcesReady {
		resources := c.bossResources
		c.mu.RUnlock()
		return resources, nil
	}
	c.mu.RUnlock()

	return c.RefreshBossResources(ctx)
}

// RefreshBossResources 强制回源刷新 Boss 低频资源。
func (c *Cache) RefreshBossResources(ctx context.Context) (core.BossResources, error) {
	resources, err := c.reader.GetBossResources(ctx)
	if err != nil {
		return core.BossResources{}, err
	}

	c.mu.Lock()
	c.bossResources = resources
	c.bossResourcesReady = true
	c.mu.Unlock()

	return resources, nil
}

// GetUserState 返回指定昵称的个人态；未命中时回源加载。
func (c *Cache) GetUserState(ctx context.Context, nickname string) (core.UserState, error) {
	normalizedNickname := strings.TrimSpace(nickname)
	if normalizedNickname == "" {
		return c.reader.GetUserState(ctx, "")
	}

	c.mu.RLock()
	userState, ok := c.users[normalizedNickname]
	c.mu.RUnlock()
	if ok {
		return userState, nil
	}

	return c.RefreshUser(ctx, normalizedNickname)
}

// RefreshUser 强制回源刷新一个昵称的个人态。
func (c *Cache) RefreshUser(ctx context.Context, nickname string) (core.UserState, error) {
	normalizedNickname := strings.TrimSpace(nickname)
	userState, err := c.reader.GetUserState(ctx, normalizedNickname)
	if err != nil {
		return core.UserState{}, err
	}
	if normalizedNickname == "" {
		return userState, nil
	}

	c.mu.Lock()
	c.users[normalizedNickname] = userState
	c.mu.Unlock()

	return userState, nil
}

// RefreshUsers 批量刷新多个昵称的个人态。
func (c *Cache) RefreshUsers(ctx context.Context, nicknames []string) (map[string]core.UserState, error) {
	refreshed := make(map[string]core.UserState, len(nicknames))
	seen := make(map[string]struct{}, len(nicknames))

	for _, nickname := range nicknames {
		normalizedNickname := strings.TrimSpace(nickname)
		if normalizedNickname == "" {
			continue
		}
		if _, ok := seen[normalizedNickname]; ok {
			continue
		}
		seen[normalizedNickname] = struct{}{}

		userState, err := c.RefreshUser(ctx, normalizedNickname)
		if err != nil {
			return nil, err
		}
		refreshed[normalizedNickname] = userState
	}

	return refreshed, nil
}

// GetState 返回公共快照与个人态组合后的完整状态。
func (c *Cache) GetState(ctx context.Context, nickname string) (core.State, error) {
	snapshot, err := c.GetSnapshot(ctx)
	if err != nil {
		return core.State{}, err
	}

	userState, err := c.GetUserState(ctx, nickname)
	if err != nil {
		return core.State{}, err
	}

	return core.ComposeState(snapshot, userState), nil
}
