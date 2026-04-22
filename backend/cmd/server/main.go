package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"long/internal/admin"
	"long/internal/config"
	"long/internal/events"
	"long/internal/httpapi"
	"long/internal/nickname"
	ossupload "long/internal/oss"
	"long/internal/ratelimit"
	"long/internal/vote"
)

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
	stateCache := events.NewCache(store)
	dispatcher := events.NewDispatcher(stateCache, hub)
	changeBus := events.NewRedisChangeBus(redisClient, vote.RealtimeEventChannel(cfg.RedisPrefix))
	eventHandler := events.NewHandler(hub, stateCache)
	clickLimiter := ratelimit.NewLimiter(ratelimit.Config{
		Limit:             cfg.RateLimit.Limit,
		Window:            cfg.RateLimit.Window,
		BlacklistDuration: cfg.RateLimit.BlacklistDuration,
	})
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
	httpServer := httpapi.NewHertzServer(serverAddress(cfg.Port), httpapi.Options{
		Store:           store,
		StateView:       stateCache,
		ChangePublisher: changeBus,
		ClickGuard:      clickLimiter,
		Events:          eventHandler,
		PublicDir:       cfg.PublicDir,
		OSSSigner:       ossSigner,
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      cfg.Admin.Username,
			Password:      cfg.Admin.Password,
			SessionSecret: cfg.Admin.SessionSecret,
		}),
	})

	errCh := make(chan error, 1)
	go func() {
		log.Printf("Vote wall listening on port %d", cfg.Port)
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
		// 低频扫描 Redis，兜底同步手工新增但未走支持写路径的按钮。
		ticker := time.NewTicker(cfg.ButtonPollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-pollCtx.Done():
				return
			case <-ticker.C:
				changed, err := store.SyncButtonIndex(pollCtx)
				if err != nil {
					log.Printf("Failed to sync button index from Redis: %v", err)
					continue
				}
				if !changed {
					continue
				}
				if err := changeBus.PublishChange(pollCtx, vote.StateChange{
					Type:      vote.StateChangeButtonMetaChanged,
					Timestamp: time.Now().Unix(),
				}); err != nil {
					log.Printf("Failed to publish button metadata change: %v", err)
				}
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

func serverAddress(port int) string {
	return net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port))
}
