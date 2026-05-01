package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"long/internal/config"
	"long/internal/mongostore"
	"long/internal/nickname"
	"long/internal/vote"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return errors.New("用法: go -C backend run ./cmd/migratecolddata <plan|migrate|verify|cleanup>")
	}

	cfg, err := config.LoadTest()
	if err != nil {
		return err
	}
	if !cfg.Mongo.Enabled {
		return errors.New("mongo.enabled=false，无法执行冷数据迁移")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	redisClient, err := newRedisClient(ctx, cfg)
	if err != nil {
		return err
	}
	defer redisClient.Close()

	mongoClient, err := mongo.Connect(ctx, options.Client().
		ApplyURI(cfg.Mongo.URI).
		SetConnectTimeout(cfg.Mongo.ConnectTimeout))
	if err != nil {
		return fmt.Errorf("connect mongo: %w", err)
	}
	defer mongoClient.Disconnect(context.Background())

	mongoDB := mongoClient.Database(cfg.Mongo.Database)
	bossStore := mongostore.NewBossHistoryStore(mongoDB, cfg.Mongo.WriteTimeout, cfg.Mongo.ReadTimeout)
	messageStore := mongostore.NewMessageStore(mongoDB, cfg.Mongo.WriteTimeout, cfg.Mongo.ReadTimeout)
	if err := bossStore.EnsureIndexes(ctx); err != nil {
		return err
	}
	if err := messageStore.EnsureIndexes(ctx); err != nil {
		return err
	}

	redisStore := vote.NewStore(redisClient, cfg.RedisPrefix, vote.StoreOptions{
		CriticalChancePercent: 5,
		CriticalCount:         0,
	}, nickname.NewSensitiveLexiconValidator())

	command := strings.ToLower(strings.TrimSpace(os.Args[1]))
	switch command {
	case "plan":
		return runPlan(ctx, cfg, redisClient, mongoDB, redisStore)
	case "migrate":
		return runMigrate(ctx, redisStore, bossStore, messageStore)
	case "verify":
		return runVerify(ctx, cfg, redisClient, mongoDB, redisStore)
	case "cleanup":
		return runCleanup(ctx, cfg, redisClient)
	default:
		return fmt.Errorf("未知命令 %q", command)
	}
}

func runPlan(ctx context.Context, cfg config.Config, redisClient redis.UniversalClient, mongoDB *mongo.Database, redisStore *vote.Store) error {
	bossHistory, err := redisStore.ListBossHistory(ctx)
	if err != nil {
		return fmt.Errorf("load redis boss history: %w", err)
	}

	messageCount, err := redisClient.ZCard(ctx, cfg.RedisPrefix+"messages").Result()
	if err != nil {
		return fmt.Errorf("count redis messages: %w", err)
	}

	mongoBossCount, err := mongoDB.Collection("boss_history").CountDocuments(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("count mongo boss history: %w", err)
	}
	mongoMessageCount, err := mongoDB.Collection("wall_messages").CountDocuments(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("count mongo messages: %w", err)
	}

	fmt.Printf("Redis Boss 历史: %d\n", len(bossHistory))
	fmt.Printf("Redis 留言历史: %d\n", messageCount)
	fmt.Printf("Mongo Boss 历史: %d\n", mongoBossCount)
	fmt.Printf("Mongo 留言历史: %d\n", mongoMessageCount)
	fmt.Printf("Cleanup 将删除 key:\n")
	fmt.Printf("- %s\n", cfg.RedisPrefix+"boss:history")
	fmt.Printf("- %s*\n", cfg.RedisPrefix+"boss:history:")
	fmt.Printf("- %s\n", cfg.RedisPrefix+"messages")
	fmt.Printf("- %s\n", cfg.RedisPrefix+"message:seq")
	fmt.Printf("- %s*\n", cfg.RedisPrefix+"message:")
	return nil
}

func runMigrate(ctx context.Context, redisStore *vote.Store, bossStore *mongostore.BossHistoryStore, messageStore *mongostore.MessageStore) error {
	bossHistory, err := redisStore.ListBossHistory(ctx)
	if err != nil {
		return fmt.Errorf("load redis boss history: %w", err)
	}
	for _, item := range bossHistory {
		if err := bossStore.SaveBossHistory(ctx, item); err != nil {
			return fmt.Errorf("save boss history %s: %w", item.ID, err)
		}
	}

	cursor := ""
	migratedMessages := 0
	for {
		page, err := redisStore.ListMessages(ctx, cursor, 200)
		if err != nil {
			return fmt.Errorf("load redis messages: %w", err)
		}
		for _, item := range page.Items {
			if err := messageStore.UpsertMessage(ctx, item); err != nil {
				return fmt.Errorf("save message %s: %w", item.ID, err)
			}
			migratedMessages++
		}
		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}

	fmt.Printf("已迁移 Boss 历史: %d\n", len(bossHistory))
	fmt.Printf("已迁移留言历史: %d\n", migratedMessages)
	return nil
}

func runVerify(ctx context.Context, cfg config.Config, redisClient redis.UniversalClient, mongoDB *mongo.Database, redisStore *vote.Store) error {
	redisBossHistory, err := redisStore.ListBossHistory(ctx)
	if err != nil {
		return fmt.Errorf("load redis boss history: %w", err)
	}
	redisMessageCount, err := redisClient.ZCard(ctx, cfg.RedisPrefix+"messages").Result()
	if err != nil {
		return fmt.Errorf("count redis messages: %w", err)
	}

	mongoBossCount, err := mongoDB.Collection("boss_history").CountDocuments(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("count mongo boss history: %w", err)
	}
	mongoMessageCount, err := mongoDB.Collection("wall_messages").CountDocuments(ctx, bson.M{"status": bson.M{"$ne": "deleted"}})
	if err != nil {
		return fmt.Errorf("count mongo messages: %w", err)
	}

	if int64(len(redisBossHistory)) != mongoBossCount {
		return fmt.Errorf("boss 历史数量不一致: redis=%d mongo=%d", len(redisBossHistory), mongoBossCount)
	}
	if redisMessageCount != mongoMessageCount {
		return fmt.Errorf("留言数量不一致: redis=%d mongo=%d", redisMessageCount, mongoMessageCount)
	}

	fmt.Printf("校验通过: boss=%d messages=%d\n", mongoBossCount, mongoMessageCount)
	return nil
}

func runCleanup(ctx context.Context, cfg config.Config, redisClient redis.UniversalClient) error {
	bossKeys, err := redisClient.Keys(ctx, cfg.RedisPrefix+"boss:history:*").Result()
	if err != nil {
		return fmt.Errorf("list boss history keys: %w", err)
	}
	messageKeys, err := redisClient.Keys(ctx, cfg.RedisPrefix+"message:*").Result()
	if err != nil {
		return fmt.Errorf("list message keys: %w", err)
	}

	keys := []string{
		cfg.RedisPrefix + "boss:history",
		cfg.RedisPrefix + "messages",
		cfg.RedisPrefix + "message:seq",
	}
	keys = append(keys, bossKeys...)
	keys = append(keys, messageKeys...)
	if len(keys) == 0 {
		fmt.Println("没有需要删除的 Redis 冷数据 key")
		return nil
	}

	deleted, err := redisClient.Del(ctx, keys...).Result()
	if err != nil {
		return fmt.Errorf("delete redis cold keys: %w", err)
	}
	fmt.Printf("已删除 Redis 冷数据 key 数量: %d\n", deleted)
	return nil
}

func newRedisClient(ctx context.Context, cfg config.Config) (*redis.Client, error) {
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
		return nil, fmt.Errorf("connect redis: %w", err)
	}
	return client, nil
}
