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

	"long/cmd/internal/toolconfig"
	"long/internal/config"
	"long/internal/mongostore"
)

const defaultRoomID = "1"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return errors.New("用法: go -C backend run ./cmd/migrateroommodel <plan|migrate|verify>")
	}
	cfg, err := toolconfig.Load(toolconfig.Options{NeedMongo: true})
	if err != nil {
		return err
	}
	if !cfg.Mongo.Enabled {
		return errors.New("mongo.enabled=false，无法执行房间模型迁移")
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

	db := mongoClient.Database(cfg.Mongo.Database)
	command := strings.ToLower(strings.TrimSpace(os.Args[1]))
	switch command {
	case "plan":
		return runPlan(ctx, cfg, redisClient, db)
	case "migrate":
		return runMigrate(ctx, cfg, redisClient, db)
	case "verify":
		return runVerify(ctx, cfg, redisClient, db)
	default:
		return fmt.Errorf("未知命令 %q", command)
	}
}

func runPlan(ctx context.Context, cfg config.Config, redisClient redis.UniversalClient, db *mongo.Database) error {
	oldCurrent, _ := redisClient.Exists(ctx, cfg.RedisPrefix+"boss:current").Result()
	oldCycle, _ := redisClient.Exists(ctx, cfg.RedisPrefix+"boss:cycle").Result()
	fmt.Printf("Redis 待迁移 boss:current=%d boss:cycle=%d\n", oldCurrent, oldCycle)
	for _, name := range []string{"boss_history", "domain_events", "admin_audit_logs"} {
		count, err := db.Collection(name).CountDocuments(ctx, bson.M{"room_id": bson.M{"$exists": false}})
		if err != nil {
			return fmt.Errorf("count %s: %w", name, err)
		}
		fmt.Printf("Mongo 待回填 %s=%d\n", name, count)
	}
	return nil
}

func runMigrate(ctx context.Context, cfg config.Config, redisClient redis.UniversalClient, db *mongo.Database) error {
	if err := moveRedisKey(ctx, redisClient, cfg.RedisPrefix+"boss:current", cfg.RedisPrefix+"boss:current:"+defaultRoomID); err != nil {
		return err
	}
	if err := moveRedisKey(ctx, redisClient, cfg.RedisPrefix+"boss:cycle", cfg.RedisPrefix+"boss:cycle:"+defaultRoomID); err != nil {
		return err
	}
	if err := redisClient.HSet(ctx, cfg.RedisPrefix+"boss:current:"+defaultRoomID, "room_id", defaultRoomID, "queue_id", defaultRoomID).Err(); err != nil {
		return fmt.Errorf("set boss current room fields: %w", err)
	}
	if err := redisClient.HSet(ctx, cfg.RedisPrefix+"boss:cycle:"+defaultRoomID, "queue_id", defaultRoomID).Err(); err != nil {
		return fmt.Errorf("set boss cycle queue_id: %w", err)
	}

	for _, name := range []string{"boss_history", "domain_events", "admin_audit_logs"} {
		if _, err := db.Collection(name).UpdateMany(ctx,
			bson.M{"$or": []bson.M{
				{"room_id": bson.M{"$exists": false}},
				{"room_id": ""},
				{"queue_id": bson.M{"$exists": false}},
				{"queue_id": ""},
			}},
			bson.M{"$set": bson.M{"room_id": defaultRoomID, "queue_id": defaultRoomID}},
		); err != nil {
			return fmt.Errorf("backfill %s: %w", name, err)
		}
	}

	if err := ensureMongoIndexes(ctx, cfg, db); err != nil {
		return err
	}
	return runVerify(ctx, cfg, redisClient, db)
}

func runVerify(ctx context.Context, cfg config.Config, redisClient redis.UniversalClient, db *mongo.Database) error {
	if exists, err := redisClient.Exists(ctx, cfg.RedisPrefix+"boss:current").Result(); err != nil {
		return err
	} else if exists != 0 {
		return errors.New("旧 boss:current 仍存在")
	}
	if exists, err := redisClient.Exists(ctx, cfg.RedisPrefix+"boss:cycle").Result(); err != nil {
		return err
	} else if exists != 0 {
		return errors.New("旧 boss:cycle 仍存在")
	}
	if roomID, err := redisClient.HGet(ctx, cfg.RedisPrefix+"boss:current:"+defaultRoomID, "room_id").Result(); err != nil && !errors.Is(err, redis.Nil) {
		return err
	} else if roomID != "" && roomID != defaultRoomID {
		return fmt.Errorf("boss:current:%s room_id=%q", defaultRoomID, roomID)
	}
	for _, name := range []string{"boss_history", "domain_events", "admin_audit_logs"} {
		count, err := db.Collection(name).CountDocuments(ctx, bson.M{"$or": []bson.M{
			{"room_id": bson.M{"$exists": false}},
			{"queue_id": bson.M{"$exists": false}},
		}})
		if err != nil {
			return fmt.Errorf("verify %s: %w", name, err)
		}
		if count != 0 {
			return fmt.Errorf("%s 仍有 %d 条缺少 room_id/queue_id", name, count)
		}
	}
	fmt.Println("房间模型迁移校验通过")
	return nil
}

func moveRedisKey(ctx context.Context, client redis.UniversalClient, oldKey string, newKey string) error {
	exists, err := client.Exists(ctx, oldKey).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return nil
	}
	destExists, err := client.Exists(ctx, newKey).Result()
	if err != nil {
		return err
	}
	if destExists != 0 {
		return fmt.Errorf("目标 Redis key 已存在，拒绝覆盖: %s", newKey)
	}
	if err := client.Rename(ctx, oldKey, newKey).Err(); err != nil {
		return fmt.Errorf("rename %s -> %s: %w", oldKey, newKey, err)
	}
	return nil
}

func ensureMongoIndexes(ctx context.Context, cfg config.Config, db *mongo.Database) error {
	if err := mongostore.NewBossHistoryStore(db, cfg.Mongo.WriteTimeout, cfg.Mongo.ReadTimeout).EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("ensure boss history indexes: %w", err)
	}
	if err := mongostore.NewDomainEventStore(db, cfg.Mongo.WriteTimeout).EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("ensure domain event indexes: %w", err)
	}
	if err := mongostore.NewAdminAuditStore(db, cfg.Mongo.WriteTimeout).EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("ensure admin audit indexes: %w", err)
	}
	return nil
}

func newRedisClient(ctx context.Context, cfg config.Config) (*redis.Client, error) {
	options := &redis.Options{
		Addr:     net.JoinHostPort(cfg.Redis.Host, fmt.Sprintf("%d", cfg.Redis.Port)),
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}
	if cfg.Redis.TLSEnabled {
		options.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	client := redis.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("connect redis: %w", err)
	}
	return client, nil
}
