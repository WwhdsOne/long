package events

import (
	"context"
	"strings"
	"sync"

	"long/internal/vote"
)

// Cache 在进程内缓存公共快照与个人态，避免每次推送都回源 Redis 聚合。
type Cache struct {
	reader        StateReader
	mu            sync.RWMutex
	snapshot      vote.Snapshot
	snapshotReady bool
	users         map[string]vote.UserState
}

// NewCache 创建一个可作为只读视图使用的状态缓存。
func NewCache(reader StateReader) *Cache {
	return &Cache{
		reader: reader,
		users:  make(map[string]vote.UserState),
	}
}

// GetSnapshot 返回缓存中的公共快照；未命中时回源加载。
func (c *Cache) GetSnapshot(ctx context.Context) (vote.Snapshot, error) {
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
func (c *Cache) RefreshSnapshot(ctx context.Context) (vote.Snapshot, error) {
	snapshot, err := c.reader.GetSnapshot(ctx)
	if err != nil {
		return vote.Snapshot{}, err
	}

	c.mu.Lock()
	c.snapshot = snapshot
	c.snapshotReady = true
	c.mu.Unlock()

	return snapshot, nil
}

// GetUserState 返回指定昵称的个人态；未命中时回源加载。
func (c *Cache) GetUserState(ctx context.Context, nickname string) (vote.UserState, error) {
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
func (c *Cache) RefreshUser(ctx context.Context, nickname string) (vote.UserState, error) {
	normalizedNickname := strings.TrimSpace(nickname)
	userState, err := c.reader.GetUserState(ctx, normalizedNickname)
	if err != nil {
		return vote.UserState{}, err
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
func (c *Cache) RefreshUsers(ctx context.Context, nicknames []string) (map[string]vote.UserState, error) {
	refreshed := make(map[string]vote.UserState, len(nicknames))
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
func (c *Cache) GetState(ctx context.Context, nickname string) (vote.State, error) {
	snapshot, err := c.GetSnapshot(ctx)
	if err != nil {
		return vote.State{}, err
	}

	userState, err := c.GetUserState(ctx, nickname)
	if err != nil {
		return vote.State{}, err
	}

	return vote.ComposeState(snapshot, userState), nil
}
