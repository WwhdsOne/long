package vote

import (
	"context"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestDebugWwhdsUserState(t *testing.T) {
	if os.Getenv("LONG_RUN_DEBUG_REDIS_TEST") == "" {
		t.Skip("跳过依赖外部 Redis 的调试测试；设置 LONG_RUN_DEBUG_REDIS_TEST=1 可启用")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     "47.93.83.136:6379",
		Password: "Wwh852456",
		DB:       2,
	})
	defer func() {
		_ = client.Close()
	}()

	store := NewStore(client, "vote:", StoreOptions{}, nil)
	ctx := context.Background()
	nickname := "Wwhds"

	if _, err := store.resourcesForNickname(ctx, nickname); err != nil {
		t.Fatalf("resourcesForNickname: %v", err)
	}
	if _, err := store.GetUserStats(ctx, nickname); err != nil {
		t.Fatalf("GetUserStats: %v", err)
	}
	_, equipped, err := store.loadoutForNickname(ctx, nickname)
	if err != nil {
		t.Fatalf("loadoutForNickname: %v", err)
	}
	if _, err := store.inventoryForNickname(ctx, nickname, equipped); err != nil {
		t.Fatalf("inventoryForNickname: %v", err)
	}
	if _, err := store.recentRewardsForNickname(ctx, nickname); err != nil {
		t.Fatalf("recentRewardsForNickname: %v", err)
	}
	boss, err := store.currentBoss(ctx)
	if err != nil {
		t.Fatalf("currentBoss: %v", err)
	}
	if boss != nil {
		if _, err := store.bossStatsForNickname(ctx, boss.ID, nickname); err != nil {
			t.Fatalf("bossStatsForNickname: %v", err)
		}
	}
}
