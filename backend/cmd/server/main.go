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
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"

	"long/internal/admin"
	"long/internal/archive"
	"long/internal/config"
	"long/internal/core"
	"long/internal/events"
	"long/internal/httpapi"
	"long/internal/mongostore"
	"long/internal/nickname"
	ossupload "long/internal/oss"
	"long/internal/playerauth"
	"long/internal/ratelimit"
	"long/internal/xlog"
)

func main() {
	if err := configureLocalTimezone(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
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
	listenAddr := serverAddress(cfg.Port)
	serverTLSConfig, err := resolveServerTLSConfig()
	if err != nil {
		return err
	}
	printStartupInfo(cfg, listenAddr, serverTLSConfig != nil)
	if !cfg.Mongo.Enabled {
		return errors.New("mongo.enabled 必须为 true，冷数据已固定切换到 MongoDB")
	}

	redisOptions := &redis.Options{
		Addr:         net.JoinHostPort(cfg.Redis.Host, fmt.Sprintf("%d", cfg.Redis.Port)),
		Username:     cfg.Redis.Username,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
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
	var bossHistoryStore *mongostore.BossHistoryStore
	var bossHistoryQueue *archive.BossHistoryQueue
	var mongoMessageStore *mongostore.MessageStore
	var taskDefinitionStore *mongostore.TaskDefinitionStore
	var taskClaimLogStore *mongostore.TaskClaimLogStore
	var taskCycleArchiveStore *mongostore.TaskCycleArchiveStore
	var shopItemStore *mongostore.ShopItemStore
	var shopPurchaseLogStore *mongostore.ShopPurchaseLogStore
	var staminaPurchaseLogStore *mongostore.StaminaPurchaseLogStore
	var accountRiskEventWriter accountRiskEventQueueWriter
	var equipmentDraftFailureWriter httpapi.EquipmentDraftFailureWriter
	var adminAuditWriter httpapi.AdminAuditWriter
	var domainEventWriter httpapi.DomainEventWriter
	var accessLogQueue *archive.AsyncQueue[xlog.AccessLogEntry]
	if cfg.Mongo.Enabled {
		mongoClient, err = mongo.Connect(options.Client().
			ApplyURI(cfg.Mongo.URI).
			SetConnectTimeout(cfg.Mongo.ConnectTimeout))
		if err != nil {
			return fmt.Errorf("connect mongo: %w", err)
		}
		if err := mongoClient.Ping(startupCtx, nil); err != nil {
			return fmt.Errorf("ping mongo: %w", err)
		}

		mongoDB := mongoClient.Database(cfg.Mongo.Database)
		bossHistoryStore = mongostore.NewBossHistoryStore(mongoDB, cfg.Mongo.WriteTimeout, cfg.Mongo.ReadTimeout)
		if err := bossHistoryStore.EnsureIndexes(startupCtx); err != nil {
			return fmt.Errorf("ensure mongo indexes: %w", err)
		}
		mongoMessageStore = mongostore.NewMessageStore(mongoDB, cfg.Mongo.WriteTimeout, cfg.Mongo.ReadTimeout)
		if err := mongoMessageStore.EnsureIndexes(startupCtx); err != nil {
			return fmt.Errorf("ensure mongo message indexes: %w", err)
		}
		adminAuditStore := mongostore.NewAdminAuditStore(mongoDB, cfg.Mongo.WriteTimeout)
		if err := adminAuditStore.EnsureIndexes(startupCtx); err != nil {
			return fmt.Errorf("ensure mongo admin audit indexes: %w", err)
		}
		domainEventStore := mongostore.NewDomainEventStore(mongoDB, cfg.Mongo.WriteTimeout)
		if err := domainEventStore.EnsureIndexes(startupCtx); err != nil {
			return fmt.Errorf("ensure mongo domain event indexes: %w", err)
		}
		systemLogStore := mongostore.NewSystemLogStore(mongoDB, cfg.Mongo.WriteTimeout)
		if err := systemLogStore.EnsureIndexes(startupCtx); err != nil {
			return fmt.Errorf("ensure mongo system log indexes: %w", err)
		}
		taskDefinitionStore = mongostore.NewTaskDefinitionStore(mongoDB, cfg.Mongo.WriteTimeout, cfg.Mongo.ReadTimeout)
		if err := taskDefinitionStore.EnsureIndexes(startupCtx); err != nil {
			return fmt.Errorf("ensure mongo task definition indexes: %w", err)
		}
		taskClaimLogStore = mongostore.NewTaskClaimLogStore(mongoDB, cfg.Mongo.WriteTimeout, cfg.Mongo.ReadTimeout)
		if err := taskClaimLogStore.EnsureIndexes(startupCtx); err != nil {
			return fmt.Errorf("ensure mongo task claim log indexes: %w", err)
		}
		taskCycleArchiveStore = mongostore.NewTaskCycleArchiveStore(mongoDB, cfg.Mongo.WriteTimeout)
		if err := taskCycleArchiveStore.EnsureIndexes(startupCtx); err != nil {
			return fmt.Errorf("ensure mongo task cycle archive indexes: %w", err)
		}
		shopItemStore = mongostore.NewShopItemStore(mongoDB, cfg.Mongo.WriteTimeout, cfg.Mongo.ReadTimeout)
		if err := shopItemStore.EnsureIndexes(startupCtx); err != nil {
			return fmt.Errorf("ensure mongo shop item indexes: %w", err)
		}
		shopPurchaseLogStore = mongostore.NewShopPurchaseLogStore(mongoDB, cfg.Mongo.WriteTimeout)
		if err := shopPurchaseLogStore.EnsureIndexes(startupCtx); err != nil {
			return fmt.Errorf("ensure mongo shop purchase log indexes: %w", err)
		}
		staminaPurchaseLogStore = mongostore.NewStaminaPurchaseLogStore(mongoDB, cfg.Mongo.WriteTimeout)
		if err := staminaPurchaseLogStore.EnsureIndexes(startupCtx); err != nil {
			return fmt.Errorf("ensure mongo stamina purchase log indexes: %w", err)
		}
		accountRiskEventStore := mongostore.NewAccountRiskEventStore(mongoDB, cfg.Mongo.WriteTimeout)
		if err := accountRiskEventStore.EnsureIndexes(startupCtx); err != nil {
			return fmt.Errorf("ensure mongo account risk event indexes: %w", err)
		}
		equipmentDraftFailureStore := mongostore.NewEquipmentDraftFailureStore(mongoDB, cfg.Mongo.WriteTimeout)
		if err := equipmentDraftFailureStore.EnsureIndexes(startupCtx); err != nil {
			return fmt.Errorf("ensure mongo equipment draft failure indexes: %w", err)
		}
		equipmentDraftFailureWriter = equipmentDraftFailureStore

		adminAuditQueue := archive.NewAsyncQueue[core.AdminAuditLog](archive.AsyncQueueConfig{
			Name:         "admin-audit",
			BufferSize:   256,
			WorkerCount:  2,
			WriteTimeout: cfg.Mongo.WriteTimeout,
		}, adminAuditStore.WriteAdminAuditLog)
		adminAuditQueue.Start()
		defer adminAuditQueue.Close()
		adminAuditWriter = adminAuditQueueWriter{queue: adminAuditQueue}

		domainEventQueue := archive.NewAsyncQueue[core.DomainEvent](archive.AsyncQueueConfig{
			Name:         "domain-events",
			BufferSize:   512,
			WorkerCount:  2,
			WriteTimeout: cfg.Mongo.WriteTimeout,
		}, domainEventStore.WriteDomainEvent)
		domainEventQueue.Start()
		defer domainEventQueue.Close()
		domainEventWriter = domainEventQueueWriter{queue: domainEventQueue}

		accountRiskEventQueue := archive.NewAsyncQueue[core.AccountRiskEventLog](archive.AsyncQueueConfig{
			Name:         "account-risk-events",
			BufferSize:   512,
			WorkerCount:  2,
			WriteTimeout: cfg.Mongo.WriteTimeout,
		}, accountRiskEventStore.WriteAccountRiskEvent)
		accountRiskEventQueue.Start()
		defer accountRiskEventQueue.Close()
		accountRiskEventWriter = accountRiskEventQueueWriter{queue: accountRiskEventQueue}

		systemLogQueue := archive.NewAsyncQueue[xlog.SystemLogEntry](archive.AsyncQueueConfig{
			Name:         "system-logs",
			BufferSize:   512,
			WorkerCount:  2,
			WriteTimeout: cfg.Mongo.WriteTimeout,
		}, systemLogStore.WriteSystemLog)
		systemLogQueue.Start()
		defer systemLogQueue.Close()
		xlog.SetSystemLogHook(xlog.HookFromWriter(systemLogQueueWriter{queue: systemLogQueue}))
		accessLogStore := mongostore.NewAccessLogStore(mongoDB, cfg.Mongo.WriteTimeout)
		if err := accessLogStore.EnsureIndexes(startupCtx); err != nil {
			return fmt.Errorf("ensure mongo access log indexes: %w", err)
		}
		accessLogQueue = archive.NewAsyncQueue[xlog.AccessLogEntry](archive.AsyncQueueConfig{
			Name:         "access-logs",
			BufferSize:   1024,
			WorkerCount:  4,
			WriteTimeout: cfg.Mongo.WriteTimeout,
		}, accessLogStore.WriteAccessLog)
		accessLogQueue.Start()
		defer accessLogQueue.Close()
		bossHistoryQueue = archive.NewBossHistoryQueue(archive.BossHistoryQueueConfig{
			BufferSize:   256,
			WorkerCount:  2,
			WriteTimeout: cfg.Mongo.WriteTimeout,
		}, bossHistoryStore)
		bossHistoryQueue.Start()
		defer bossHistoryQueue.Close()
		defer func() {
			if mongoClient != nil {
				_ = mongoClient.Disconnect(context.Background())
			}
		}()
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

	nicknameValidator := nickname.NewSensitiveLexiconValidator()
	store := core.NewStore(redisClient, cfg.RedisPrefix, core.StoreOptions{
		CriticalChancePercent: 5,
		AccountRisk: core.AccountRiskConfig{
			ScoreWindow:           cfg.AntiScript.ScoreWindow,
			PurchaseClickCooldown: cfg.AntiScript.PurchaseClickCooldown,
			BanThreshold8h:        cfg.AntiScript.BanThreshold8h,
			BanThreshold24h:       cfg.AntiScript.BanThreshold24h,
			BanThreshold72h:       cfg.AntiScript.BanThreshold72h,
			Points: core.AccountRiskPoints{
				ClickRateLimitHit:        cfg.AntiScript.Points.ClickRateLimitHit,
				LoginTurnstileInvalid:    cfg.AntiScript.Points.LoginTurnstileInvalid,
				StaminaTurnstileInvalid:  cfg.AntiScript.Points.StaminaTurnstileInvalid,
				PostStaminaPurchaseClick: cfg.AntiScript.Points.PostStaminaPurchaseClick,
			},
		},
		Room: core.RoomConfig{
			Enabled:        cfg.Room.Enabled,
			Count:          cfg.Room.Count,
			DefaultRoom:    cfg.Room.DefaultRoom,
			SwitchCooldown: cfg.Room.SwitchCooldown,
		},
		BossHistoryStore:         bossHistoryStore,
		BossHistoryArchiver:      bossHistoryQueue,
		MessageStore:             mongoMessageStore,
		TaskDefinitionStore:      taskDefinitionStore,
		TaskClaimLogStore:        taskClaimLogStore,
		TaskCycleArchiveStore:    taskCycleArchiveStore,
		ShopCatalogStore:         shopItemStore,
		ShopPurchaseLogStore:     shopPurchaseLogStore,
		StaminaPurchaseLogStore:  staminaPurchaseLogStore,
		AccountRiskEventLogStore: accountRiskEventWriter,
	}, nicknameValidator)
	hub := events.NewHub()
	stateCache := events.NewCache(store)
	dispatcher := events.NewDispatcher(stateCache, hub, cfg.Realtime.DebounceMs)
	changeBus := events.NewRedisChangeBus(redisClient, core.RealtimeEventChannel(cfg.RedisPrefix))
	playerAuthenticator := playerauth.NewService(redisClient, playerauth.Config{
		Namespace: cfg.RedisPrefix,
		JWTSecret: cfg.PlayerAuth.JWTSecret,
		TokenTTL:  cfg.PlayerAuth.JWTTTL,
	}, nicknameValidator)
	eventHandler := events.NewHandler(hub, stateCache, func(ctx context.Context, c *app.RequestContext) string {
		return httpapi.AuthenticatedPlayerNickname(ctx, c, playerAuthenticator)
	})
	clickLimiter := ratelimit.NewLimiter(ratelimit.Config{
		Rules: func() []ratelimit.WindowConfig {
			rules := make([]ratelimit.WindowConfig, 0, len(cfg.AntiScript.ClickRateLimit.Rules))
			for _, rule := range cfg.AntiScript.ClickRateLimit.Rules {
				rules = append(rules, ratelimit.WindowConfig{
					Limit:  rule.Limit,
					Window: rule.Window,
				})
			}
			return rules
		}(),
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
	staminaPurchaseTurnstile := httpapi.NewStaminaPurchaseTurnstile(httpapi.StaminaPurchaseTurnstileConfig{
		Enabled:                   cfg.Turnstile.Enabled,
		SiteKey:                   cfg.Turnstile.SiteKey,
		SecretKey:                 cfg.Turnstile.SecretKey,
		PurchaseStaminaSampleRate: cfg.Turnstile.PurchaseStaminaSampleRate,
		VerifyTimeoutMS:           cfg.Turnstile.VerifyTimeoutMS,
	})
	playerLoginTurnstile := httpapi.NewPlayerLoginTurnstile(httpapi.PlayerLoginTurnstileConfig{
		Enabled:         cfg.Turnstile.Enabled,
		SiteKey:         cfg.Turnstile.SiteKey,
		SecretKey:       cfg.Turnstile.SecretKey,
		VerifyTimeoutMS: cfg.Turnstile.VerifyTimeoutMS,
	})
	httpServer := httpapi.NewHertzServer(listenAddr, httpapi.Options{
		Store:                       store,
		StateView:                   stateCache,
		ChangePublisher:             changeBus,
		StaminaPurchaseTurnstile:    staminaPurchaseTurnstile,
		PlayerLoginTurnstile:        playerLoginTurnstile,
		ClickRiskDetector:           clickLimiter,
		AccountRisk:                 store,
		RateLimitNicknameWhitelist:  cfg.AntiScript.ClickRateLimit.NicknameWhitelist,
		Afk:                         afkService,
		PlayerAuthenticator:         playerAuthenticator,
		Events:                      eventHandler,
		RealtimeHub:                 hub,
		OSSSigner:                   ossSigner,
		EquipmentDraftGenerator:     equipmentDraftGenerator,
		EquipmentDraftFailureWriter: equipmentDraftFailureWriter,
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      cfg.Admin.Username,
			Password:      cfg.Admin.Password,
			SessionSecret: cfg.Admin.SessionSecret,
		}),
		AdminAuditWriter:       adminAuditWriter,
		DomainEventWriter:      domainEventWriter,
		AccessLogQueue:         accessLogQueue,
		AdminBossHistoryReader: bossHistoryStore,
	}, serverTLSConfig)

	errCh := make(chan error, 1)
	go func() {
		xlog.L().Info("🌟hai-world🌟 listening", zap.String("listen_addr", listenAddr))
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

	go func() {
		if err := archiveExpiredTaskCyclesLoop(pollCtx, store); err != nil && !errors.Is(err, context.Canceled) {
			select {
			case errCh <- fmt.Errorf("archive expired task cycles: %w", err):
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

func printStartupInfo(cfg config.Config, listenAddr string, httpsEnabled bool) {
	displayGodAnimal()
	fmt.Println(renderStartupInfo(cfg, listenAddr, httpsEnabled))
}

func renderStartupInfo(cfg config.Config, listenAddr string, httpsEnabled bool) string {
	return fmt.Sprintf(
		"启动信息\n"+
			"  监听地址: %s\n"+
			"  HTTPS: %t\n"+
			"  Redis: %s:%d/%d\n"+
			"  Redis TLS: %t\n"+
			"  Redis 前缀: %s\n"+
			"  Mongo: %t (%s)\n"+
			"  日志: level=%s format=%s\n"+
			"  OSS: %t\n"+
		"  LLM: %t",
		listenAddr,
		httpsEnabled,
		cfg.Redis.Host,
		cfg.Redis.Port,
		cfg.Redis.DB,
		cfg.Redis.TLSEnabled,
		cfg.RedisPrefix,
		cfg.Mongo.Enabled,
		cfg.Mongo.Database,
		cfg.Log.Level,
		cfg.Log.Format,
		cfg.OSS.Enabled(),
		cfg.LLM.Enabled,
	)
}

var errIncompleteTLSFiles = errors.New("LONG_TLS_CERT_FILE 和 LONG_TLS_KEY_FILE 必须同时提供")

func configureLocalTimezone() error {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return fmt.Errorf("load asia/shanghai timezone: %w", err)
	}
	time.Local = location
	if err := os.Setenv("TZ", "Asia/Shanghai"); err != nil {
		return fmt.Errorf("set TZ env: %w", err)
	}
	return nil
}

func resolveServerTLSConfig() (*tls.Config, error) {
	certFile := strings.TrimSpace(os.Getenv("LONG_TLS_CERT_FILE"))
	keyFile := strings.TrimSpace(os.Getenv("LONG_TLS_KEY_FILE"))

	if certFile == "" && keyFile == "" {
		return nil, nil
	}
	if certFile == "" || keyFile == "" {
		return nil, errIncompleteTLSFiles
	}

	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		// 当前 Hertz TLS 入口未实际提供 HTTP/2，显式限制为 HTTP/1.1，
		// 避免浏览器协商到 h2 后触发 ERR_HTTP2_PROTOCOL_ERROR。
		NextProtos: []string{"http/1.1"},
		GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
			if err != nil {
				return nil, fmt.Errorf("load tls certificate: %w", err)
			}
			return &certificate, nil
		},
	}, nil
}

func displayGodAnimal() {
	fmt.Println(godAnimalArt())
}

func godAnimalArt() string {
	return `
                           ┏━┓     ┏━┓
                          ┏┛ ┻━━━━━┛ ┻┓
                          ┃　　　　　　 ┃
                          ┃　　　━　　　┃
                          ┃　┳┛　  ┗┳　┃
                          ┃　　　　　　 ┃
                          ┃　　　┻　　　┃
                          ┃　　　　　　 ┃
                          ┗━┓　　　┏━━━┛
                            ┃　　　┃   神兽保佑
                            ┃　　　┃   代码无BUG！
                            ┃　　　┗━━━━━━━━━┓
                            ┃　　　　　　　    ┣┓
                            ┃　　　　         ┏┛
                            ┗━┓ ┓ ┏━━━┳ ┓ ┏━┛
                              ┃ ┫ ┫   ┃ ┫ ┫
                              ┗━┻━┛   ┗━┻━┛`
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

func processTalentBleedLoop(ctx context.Context, store *core.Store, changeBus *events.RedisChangeBus) error {
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

func archiveExpiredTaskCyclesLoop(ctx context.Context, store *core.Store) error {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if _, err := store.ArchiveExpiredTaskCycles(context.Background(), time.Now()); err != nil {
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

type adminAuditQueueWriter struct {
	queue *archive.AsyncQueue[core.AdminAuditLog]
}

func (w adminAuditQueueWriter) WriteAdminAuditLog(_ context.Context, item core.AdminAuditLog) error {
	if w.queue != nil {
		w.queue.Enqueue(item)
	}
	return nil
}

type domainEventQueueWriter struct {
	queue *archive.AsyncQueue[core.DomainEvent]
}

func (w domainEventQueueWriter) WriteDomainEvent(_ context.Context, item core.DomainEvent) error {
	if w.queue != nil {
		w.queue.Enqueue(item)
	}
	return nil
}

type accountRiskEventQueueWriter struct {
	queue *archive.AsyncQueue[core.AccountRiskEventLog]
}

func (w accountRiskEventQueueWriter) WriteAccountRiskEvent(_ context.Context, item core.AccountRiskEventLog) error {
	if w.queue != nil {
		w.queue.Enqueue(item)
	}
	return nil
}

type systemLogQueueWriter struct {
	queue *archive.AsyncQueue[xlog.SystemLogEntry]
}

func (w systemLogQueueWriter) WriteSystemLog(_ context.Context, item xlog.SystemLogEntry) error {
	if w.queue != nil {
		w.queue.Enqueue(item)
	}
	return nil
}
