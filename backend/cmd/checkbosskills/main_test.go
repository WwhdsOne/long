package main

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedisClient(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()

	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	return server, client
}

func TestCollectPlayerBossKillsReadsResourceKeys(t *testing.T) {
	server, client := newTestRedisClient(t)
	defer server.Close()
	defer client.Close()

	ctx := context.Background()
	prefix := "vote:"
	if err := client.HSet(ctx, playerResourceKey(prefix, "阿明"), "boss_kills", "7").Err(); err != nil {
		t.Fatalf("seed player resource: %v", err)
	}
	if err := client.HSet(ctx, playerResourceKey(prefix, "小红"), "boss_kills", "3").Err(); err != nil {
		t.Fatalf("seed second player resource: %v", err)
	}
	if err := client.HSet(ctx, prefix+"user:旧数据", "boss_kills", "99").Err(); err != nil {
		t.Fatalf("seed legacy user key: %v", err)
	}

	rows, err := collectPlayerBossKills(ctx, client, prefix, 10)
	if err != nil {
		t.Fatalf("collect player boss kills: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0].Nickname != "阿明" || rows[0].BossKills != 7 {
		t.Fatalf("expected top row 阿明=7, got %+v", rows[0])
	}
	if rows[1].Nickname != "小红" || rows[1].BossKills != 3 {
		t.Fatalf("expected second row 小红=3, got %+v", rows[1])
	}
}

func TestPlayerResourceKeyUsesResourceNamespace(t *testing.T) {
	if got := playerResourceKey("hai-world:", " 阿明 "); got != "hai-world:resource:阿明" {
		t.Fatalf("unexpected player resource key: %s", got)
	}
}
