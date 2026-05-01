package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"long/internal/admin"
	"long/internal/archive"
	"long/internal/config"
	"long/internal/events"
	"long/internal/httpapi"
	"long/internal/mongostore"
	"long/internal/nickname"
	ossupload "long/internal/oss"
	playerauth "long/internal/playerauth"
	"long/internal/ratelimit"
	"long/internal/vote"
	"long/internal/xlog"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run 启动服务：连接 Redis、注册路由、启动 SSE 广播、处理优雅关闭
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger, err := xlog.Init(xlog.Config{
		Level:         cfg.Log.Level,
		Format:        cfg.Log.Format,
		IncludeCaller: cfg.Log.IncludeCaller,
	})
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	redisOptions := &redis.Options{
		Addr:     net.JoinHostPort(cfg.Redis.Host, fmt.Sprintf("%d", cfg.Redis.Port)),
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}
	if cfg.Redis.TLSEnabled {
		redisOptions.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	redisClient := redis.NewClient(redisOptions)
	startupCtx, startupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer startupCancel()

	if err := redisClient.Ping(startupCtx).Err(); err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}

	var mongoClient *mongo.Client
	var bossHistoryArchive *archive.BossHistoryQueue
	var adminBossHistoryReader httpapi.AdminBossHistoryReader
	if cfg.Mongo.Enabled {
		mongoClient, err = mongo.Connect(startupCtx, options.Client().
			ApplyURI(cfg.Mongo.URI).
			SetConnectTimeout(cfg.Mongo.ConnectTimeout))
		if err != nil {
			return fmt.Errorf("connect mongo: %w", err)
		}
		if err := mongoClient.Ping(startupCtx, nil); err != nil {
			return fmt.Errorf("ping mongo: %w", err)
		}

		mongoDB := mongoClient.Database(cfg.Mongo.Database)
		bossHistoryStore := mongostore.NewBossHistoryStore(mongoDB, cfg.Mongo.WriteTimeout, cfg.Mongo.ReadTimeout)
		if err := bossHistoryStore.EnsureIndexes(startupCtx); err != nil {
			return fmt.Errorf("ensure mongo indexes: %w", err)
		}
		adminBossHistoryReader = bossHistoryStore

		if cfg.Archive.BossHistoryDualWrite {
			bossHistoryArchive = archive.NewBossHistoryQueue(archive.BossHistoryQueueConfig{
				BufferSize:   128,
				WorkerCount:  2,
				WriteTimeout: cfg.Mongo.WriteTimeout,
			}, bossHistoryStore)
			bossHistoryArchive.Start()
			defer bossHistoryArchive.Close()
		}
		defer func() {
			if mongoClient != nil {
				_ = mongoClient.Disconnect(context.Background())
			}
		}()
	}

	nicknameValidator := nickname.NewSensitiveLexiconValidator()
	store := vote.NewStore(redisClient, cfg.RedisPrefix, vote.StoreOptions{
		CriticalChancePercent: 5,
		CriticalCount:         0,
		BossHistoryArchiver:   bossHistoryArchive,
	}, nicknameValidator)
	hub := events.NewHub()
	stateCache := events.NewCache(store)
	dispatcher := events.NewDispatcher(stateCache, hub, cfg.Realtime.DebounceMs)
	changeBus := events.NewRedisChangeBus(redisClient, vote.RealtimeEventChannel(cfg.RedisPrefix))
	playerAuthenticator := playerauth.NewService(redisClient, playerauth.Config{
		Namespace: cfg.RedisPrefix,
		JWTSecret: cfg.PlayerAuth.JWTSecret,
		TokenTTL:  cfg.PlayerAuth.JWTTTL,
	}, nicknameValidator)
	eventHandler := events.NewHandler(hub, stateCache, func(ctx context.Context, c *app.RequestContext) string {
		return httpapi.AuthenticatedPlayerNickname(ctx, c, playerAuthenticator)
	})
	clickLimiter := ratelimit.NewLimiter(ratelimit.Config{
		Limit:             cfg.RateLimit.Limit,
		Window:            cfg.RateLimit.Window,
		BlacklistDuration: cfg.RateLimit.BlacklistDuration,
	})
	afkService := httpapi.NewAfkService(store, changeBus, redisClient, cfg.RedisPrefix)
	defer afkService.Close()
	var ossSigner *ossupload.Signer
	if cfg.OSS.Enabled() {
		ossSigner = ossupload.NewSigner(ossupload.Config{
			AccessKeyID:     cfg.OSS.AccessKeyID,
			AccessKeySecret: cfg.OSS.AccessKeySecret,
			Bucket:          cfg.OSS.Bucket,
			Region:          cfg.OSS.Region,
			PublicBaseURL:   cfg.OSS.PublicBaseURL,
			UploadDirPrefix: cfg.OSS.UploadDirPrefix,
			ExpireSeconds:   cfg.OSS.ExpireSeconds,
		})
	}
	var equipmentDraftGenerator httpapi.EquipmentDraftGenerator
	if cfg.LLM.Enabled {
		equipmentDraftGenerator = httpapi.NewOpenAIEquipmentDraftGenerator(httpapi.EquipmentDraftGeneratorConfig{
			APIKey:  cfg.LLM.APIKey,
			BaseURL: cfg.LLM.BaseURL,
			Model:   cfg.LLM.Model,
			Timeout: cfg.LLM.Timeout,
		})
	}
	httpServer := httpapi.NewHertzServer(serverAddress(cfg.Port), httpapi.Options{
		Store:                   store,
		StateView:               stateCache,
		ChangePublisher:         changeBus,
		ClickGuard:              clickLimiter,
		Afk:                     afkService,
		PlayerAuthenticator:     playerAuthenticator,
		Events:                  eventHandler,
		RealtimeHub:             hub,
		PublicDir:               cfg.PublicDir,
		OSSSigner:               ossSigner,
		EquipmentDraftGenerator: equipmentDraftGenerator,
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      cfg.Admin.Username,
			Password:      cfg.Admin.Password,
			SessionSecret: cfg.Admin.SessionSecret,
		}),
		AdminBossHistoryReader: selectAdminBossHistoryReader(cfg, store, adminBossHistoryReader),
	})

	errCh := make(chan error, 1)
	listenAddr := serverAddress(cfg.Port)
	go func() {
		xlog.L().Info("vote wall listening", zap.String("listen_addr", listenAddr))
		if err := httpServer.Run(); err != nil {
			errCh <- err
		}
	}()

	pollCtx, pollCancel := context.WithCancel(context.Background())
	defer pollCancel()

	go func() {
		if err := changeBus.Listen(pollCtx, dispatcher.HandleChange); err != nil && !errors.Is(err, context.Canceled) {
			select {
			case errCh <- fmt.Errorf("listen realtime changes: %w", err):
			default:
			}
		}
	}()

	go func() {
		if err := broadcastLeaderboardOnMinute(pollCtx, dispatcher); err != nil && !errors.Is(err, context.Canceled) {
			select {
			case errCh <- fmt.Errorf("broadcast leaderboard every minute: %w", err):
			default:
			}
		}
	}()

	go func() {
		if err := processTalentBleedLoop(pollCtx, store, changeBus); err != nil && !errors.Is(err, context.Canceled) {
			select {
			case errCh <- fmt.Errorf("process talent bleed loop: %w", err):
			default:
			}
		}
	}()

	if _, err := stateCache.RefreshSnapshot(startupCtx); err != nil {
		return fmt.Errorf("warm snapshot cache: %w", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case err := <-errCh:
		pollCancel()
		_ = redisClient.Close()
		return err
	case <-sigCh:
	}

	pollCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}

	if err := redisClient.Close(); err != nil {
		return fmt.Errorf("close redis client: %w", err)
	}

	return nil
}

func broadcastLeaderboardOnMinute(ctx context.Context, dispatcher *events.Dispatcher) error {
	for {
		next := time.Now().Truncate(time.Minute).Add(time.Minute)
		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}

		if err := dispatcher.BroadcastLeaderboard(context.Background()); err != nil {
			return err
		}
	}
}

func processTalentBleedLoop(ctx context.Context, store *vote.Store, changeBus *events.RedisChangeBus) error {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		changes, err := store.ProcessTalentBleedTicks(context.Background())
		if err != nil {
			return err
		}
		for _, change := range changes {
			if err := changeBus.PublishChange(context.Background(), change); err != nil {
				return err
			}
		}
	}
}

func serverAddress(port int) string {
	host := strings.TrimSpace(os.Getenv("LONG_LISTEN_HOST"))
	if host == "" {
		host = "127.0.0.1"
	}

	listenPort := port
	if rawPort := strings.TrimSpace(os.Getenv("LONG_LISTEN_PORT")); rawPort != "" {
		parsed, err := strconv.Atoi(rawPort)
		if err == nil && parsed > 0 {
			listenPort = parsed
		}
	}

	return net.JoinHostPort(host, fmt.Sprintf("%d", listenPort))
}

func selectAdminBossHistoryReader(cfg config.Config, store *vote.Store, mongoReader httpapi.AdminBossHistoryReader) httpapi.AdminBossHistoryReader {
	if cfg.Archive.BossHistoryReadSource == "mongo" && mongoReader != nil {
		return mongoReader
	}
	return store
}
