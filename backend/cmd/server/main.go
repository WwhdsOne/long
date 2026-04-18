package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"long/internal/admin"
	"long/internal/config"
	"long/internal/events"
	"long/internal/httpapi"
	"long/internal/nickname"
	"long/internal/ratelimit"
	"long/internal/vote"
)

// snapshotPublisher 快照发布器，避免 Redis 轮询时重复广播相同状态
type snapshotPublisher struct {
	hub  *events.Hub
	mu   sync.Mutex
	last string
}

func newSnapshotPublisher(hub *events.Hub) *snapshotPublisher {
	return &snapshotPublisher{hub: hub}
}

func (p *snapshotPublisher) BroadcastSnapshot(snapshot vote.Snapshot) error {
	_, err := p.broadcast(snapshot, false)
	return err
}

func (p *snapshotPublisher) BroadcastSnapshotIfChanged(snapshot vote.Snapshot) (bool, error) {
	return p.broadcast(snapshot, false)
}

func (p *snapshotPublisher) ForceSnapshot(snapshot vote.Snapshot) error {
	_, err := p.broadcast(snapshot, true)
	return err
}

func (p *snapshotPublisher) broadcast(snapshot vote.Snapshot, force bool) (bool, error) {
	signatureBytes, err := json.Marshal(snapshot)
	if err != nil {
		return false, err
	}

	signature := string(signatureBytes)

	p.mu.Lock()
	if !force && signature == p.last {
		p.mu.Unlock()
		return false, nil
	}
	p.last = signature
	p.mu.Unlock()

	return true, p.hub.BroadcastSnapshot()
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run 启动服务：连接 Redis、注册路由、启动 SSE 广播、处理优雅关闭
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

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

	nicknameValidator := nickname.NewSensitiveLexiconValidator()
	store := vote.NewStore(redisClient, cfg.RedisPrefix, vote.StoreOptions{
		CriticalChancePercent: cfg.CriticalHit.ChancePercent,
		CriticalCount:         cfg.CriticalHit.Count,
	}, nicknameValidator)
	if err := store.EnsureDefaults(startupCtx, config.DefaultButtons); err != nil {
		return fmt.Errorf("seed default buttons: %w", err)
	}

	hub := events.NewHub()
	publisher := newSnapshotPublisher(hub)
	eventHandler := events.NewHandler(hub, store.GetState)
	clickLimiter := ratelimit.NewLimiter(ratelimit.Config{
		Limit:             cfg.RateLimit.Limit,
		Window:            cfg.RateLimit.Window,
		BlacklistDuration: cfg.RateLimit.BlacklistDuration,
	})
	handler := httpapi.NewHandler(httpapi.Options{
		Store:       store,
		Broadcaster: publisher,
		ClickGuard:  clickLimiter,
		Events:      eventHandler,
		PublicDir:   cfg.PublicDir,
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      cfg.Admin.Username,
			Password:      cfg.Admin.Password,
			SessionSecret: cfg.Admin.SessionSecret,
		}),
	})

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("Vote wall listening on port %d", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	pollCtx, pollCancel := context.WithCancel(context.Background())
	defer pollCancel()

	go func() {
		// 定期轮询 Redis，确保动态添加的按钮即使没有点击也能显示
		ticker := time.NewTicker(cfg.ButtonPollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-pollCtx.Done():
				return
			case <-ticker.C:
				snapshot, err := store.GetSnapshot(pollCtx)
				if err != nil {
					log.Printf("Failed to sync state from Redis: %v", err)
					continue
				}
				if _, err := publisher.BroadcastSnapshotIfChanged(snapshot); err != nil {
					log.Printf("Failed to publish state: %v", err)
				}
			}
		}
	}()

	initialSnapshot, err := store.GetSnapshot(startupCtx)
	if err != nil {
		return fmt.Errorf("load initial state: %w", err)
	}
	if err := publisher.ForceSnapshot(initialSnapshot); err != nil {
		return fmt.Errorf("broadcast initial state: %w", err)
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

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}

	if err := redisClient.Close(); err != nil {
		return fmt.Errorf("close redis client: %w", err)
	}

	return nil
}
