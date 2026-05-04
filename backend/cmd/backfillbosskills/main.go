package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"long/cmd/internal/toolconfig"
	"long/internal/config"
	"long/internal/core"
	"long/internal/mongostore"
	"long/internal/nickname"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := toolconfig.Load(toolconfig.Options{
		IncludeRoom: true,
	})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	redisClient, err := newRedisClient(ctx, cfg)
	if err != nil {
		return err
	}
	defer redisClient.Close()

	storeOptions := core.StoreOptions{
		CriticalChancePercent: 5,
		Room:                  core.RoomConfig{Enabled: cfg.Room.Enabled, Count: cfg.Room.Count, DefaultRoom: cfg.Room.DefaultRoom, SwitchCooldown: cfg.Room.SwitchCooldown},
	}

	var mongoClient *mongo.Client
	if cfg.Mongo.Enabled {
		mongoClient, err = mongo.Connect(ctx, options.Client().
			ApplyURI(cfg.Mongo.URI).
			SetConnectTimeout(cfg.Mongo.ConnectTimeout))
		if err != nil {
			return fmt.Errorf("connect mongo: %w", err)
		}
		defer mongoClient.Disconnect(context.Background())

		bossStore := mongostore.NewBossHistoryStore(mongoClient.Database(cfg.Mongo.Database), cfg.Mongo.WriteTimeout, cfg.Mongo.ReadTimeout)
		storeOptions.BossHistoryStore = bossStore
	}

	store := core.NewStore(redisClient, cfg.RedisPrefix, storeOptions, nickname.NewSensitiveLexiconValidator())
	stats, err := store.RebuildBossKillCounters(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Boss 击杀回填完成: boss=%d players=%d\n", stats.HistoryBosses, stats.PlayerCount)
	return nil
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
