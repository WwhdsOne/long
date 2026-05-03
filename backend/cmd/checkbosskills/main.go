package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"long/internal/config"
)

type playerBossKill struct {
	Nickname  string
	BossKills int64
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var nickname string
	var limit int

	fs := flag.NewFlagSet("checkbosskills", flag.ContinueOnError)
	fs.StringVar(&nickname, "nickname", "", "只检查指定昵称")
	fs.IntVar(&limit, "limit", 20, "最多展示多少个玩家")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}

	cfg, err := config.LoadTest()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	redisClient, err := newRedisClient(ctx, cfg)
	if err != nil {
		return err
	}
	defer redisClient.Close()

	prefix := cfg.RedisPrefix
	totalKey := prefix + "total:boss:kills"
	totalBossKills, err := redisClient.Get(ctx, totalKey).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("read total boss kills: %w", err)
	}

	fmt.Printf("Redis: %s db=%d tls=%t\n", net.JoinHostPort(cfg.Redis.Host, fmt.Sprintf("%d", cfg.Redis.Port)), cfg.Redis.DB, cfg.Redis.TLSEnabled)
	fmt.Printf("RedisPrefix: %s\n", prefix)
	fmt.Printf("TotalBossKillsKey: %s\n", totalKey)
	if err == redis.Nil {
		fmt.Printf("TotalBossKills: <missing>\n")
	} else {
		fmt.Printf("TotalBossKills: %s\n", totalBossKills)
	}

	trimmedNickname := strings.TrimSpace(nickname)
	if trimmedNickname != "" {
		key := playerResourceKey(prefix, trimmedNickname)
		value, readErr := redisClient.HGet(ctx, key, "boss_kills").Result()
		if readErr != nil && readErr != redis.Nil {
			return fmt.Errorf("read player boss_kills: %w", readErr)
		}
		if readErr == redis.Nil {
			fmt.Printf("Player %s boss_kills: <missing>\n", trimmedNickname)
		} else {
			fmt.Printf("Player %s boss_kills: %s\n", trimmedNickname, value)
		}
		return nil
	}

	rows, err := collectPlayerBossKills(ctx, redisClient, prefix, limit)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		fmt.Println("PlayersWithBossKills: <none>")
		return nil
	}

	fmt.Printf("PlayersWithBossKills: %d\n", len(rows))
	for _, row := range rows {
		fmt.Printf("- %s: %d\n", row.Nickname, row.BossKills)
	}
	return nil
}

func collectPlayerBossKills(ctx context.Context, client redis.UniversalClient, prefix string, limit int) ([]playerBossKill, error) {
	if limit <= 0 {
		limit = 20
	}
	var cursor uint64
	pattern := playerResourceKey(prefix, "*")
	rows := make([]playerBossKill, 0, limit)

	for {
		keys, nextCursor, err := client.Scan(ctx, cursor, pattern, 200).Result()
		if err != nil {
			return nil, fmt.Errorf("scan player keys: %w", err)
		}
		for _, key := range keys {
			values, err := client.HMGet(ctx, key, "boss_kills").Result()
			if err != nil {
				return nil, fmt.Errorf("read boss_kills from %s: %w", key, err)
			}
			if len(values) == 0 || values[0] == nil {
				continue
			}
			raw := strings.TrimSpace(fmt.Sprint(values[0]))
			if raw == "" {
				continue
			}
			count, err := parseInt64(raw)
			if err != nil || count <= 0 {
				continue
			}
			rows = append(rows, playerBossKill{
				Nickname:  strings.TrimPrefix(key, playerResourceKey(prefix, "")),
				BossKills: count,
			})
		}
		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].BossKills == rows[j].BossKills {
			return rows[i].Nickname < rows[j].Nickname
		}
		return rows[i].BossKills > rows[j].BossKills
	})
	if len(rows) > limit {
		rows = rows[:limit]
	}
	return rows, nil
}

func playerResourceKey(prefix string, nickname string) string {
	return prefix + "resource:" + strings.TrimSpace(nickname)
}

func parseInt64(value string) (int64, error) {
	var result int64
	_, err := fmt.Sscan(value, &result)
	return result, err
}

func newRedisClient(ctx context.Context, cfg config.Config) (redis.UniversalClient, error) {
	redisOptions := &redis.Options{
		Addr:     net.JoinHostPort(cfg.Redis.Host, fmt.Sprintf("%d", cfg.Redis.Port)),
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}
	if cfg.Redis.TLSEnabled {
		redisOptions.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	client := redis.NewClient(redisOptions)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("connect redis: %w", err)
	}
	return client, nil
}
