package main

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"long/internal/config"
	"long/internal/core"
	"long/internal/nickname"
)

func TestCollectCleanupKeys包含历史Boss残留键且不误删当前Boss(t *testing.T) {
	t.Parallel()

	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer redisServer.Close()

	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer client.Close()

	store := core.NewStore(client, "vote:", core.StoreOptions{
		CriticalChancePercent: 5,
	}, nickname.NewSensitiveLexiconValidator())

	ctx := context.Background()
	historyBoss := &core.Boss{
		ID:         "history-1",
		Name:       "历史 Boss",
		Status:     "dead",
		MaxHP:      100,
		CurrentHP:  0,
		StartedAt:  100,
		DefeatedAt: 200,
	}
	if err := store.SaveBossToHistory(ctx, historyBoss); err != nil {
		t.Fatalf("save boss history: %v", err)
	}

	historyKeys := []string{
		"vote:boss:history-1:damage",
		"vote:boss:history-1:loot",
		"vote:boss:history-1:reward-lock",
	}
	for _, key := range historyKeys {
		if err := client.Set(ctx, key, "1", 0).Err(); err != nil {
			t.Fatalf("seed key %s: %v", key, err)
		}
	}
	currentKeys := []string{
		"vote:boss:current-1:damage",
		"vote:boss:current-1:loot",
		"vote:boss:current-1:reward-lock",
	}
	for _, key := range currentKeys {
		if err := client.Set(ctx, key, "1", 0).Err(); err != nil {
			t.Fatalf("seed current key %s: %v", key, err)
		}
	}

	if err := client.Set(ctx, "vote:message:1", "x", 0).Err(); err != nil {
		t.Fatalf("seed message item: %v", err)
	}

	keys, err := collectCleanupKeys(ctx, config.Config{RedisPrefix: "vote:"}, client, store)
	if err != nil {
		t.Fatalf("collect cleanup keys: %v", err)
	}

	for _, key := range historyKeys {
		if !slices.Contains(keys, key) {
			t.Fatalf("expected cleanup keys contain %s, got %v", key, keys)
		}
	}
	for _, key := range currentKeys {
		if slices.Contains(keys, key) {
			t.Fatalf("did not expect cleanup keys contain current boss key %s", key)
		}
	}
}

func TestRunAll按顺序执行并在失败时停止(t *testing.T) {
	t.Parallel()

	steps := make([]string, 0, 3)
	wantErr := errors.New("boom")

	err := runAll(
		func() error {
			steps = append(steps, "plan")
			return nil
		},
		func() error {
			steps = append(steps, "migrate")
			return wantErr
		},
		func() error {
			steps = append(steps, "verify")
			return nil
		},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}

	if !slices.Equal(steps, []string{"plan", "migrate"}) {
		t.Fatalf("unexpected step order: %v", steps)
	}
}

func TestCheckVerifyCounts区分严格模式与MongoOnly模式(t *testing.T) {
	t.Parallel()

	err := checkVerifyCounts(0, 190, 12, 12, true)
	if err == nil {
		t.Fatalf("expected strict verify to fail when redis boss history already cleaned")
	}

	err = checkVerifyCounts(0, 190, 12, 12, false)
	if err != nil {
		t.Fatalf("expected mongo-only verify to pass, got %v", err)
	}
}

func TestShouldUseMongoOnlyVerifyForAll在源数据提前清理时返回真(t *testing.T) {
	t.Parallel()

	if !shouldUseMongoOnlyVerifyForAll(0, 190, 12, 12) {
		t.Fatalf("expected auto verify to fall back to mongo-only when mongo is ahead of redis")
	}
	if shouldUseMongoOnlyVerifyForAll(190, 190, 12, 12) {
		t.Fatalf("did not expect fallback when redis and mongo counts match")
	}
	if shouldUseMongoOnlyVerifyForAll(190, 120, 12, 12) {
		t.Fatalf("did not expect fallback when mongo is behind redis")
	}
}
