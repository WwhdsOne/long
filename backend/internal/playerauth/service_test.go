package playerauth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"long/internal/core"
)

type nicknameValidator struct {
	err error
}

func (v nicknameValidator) Validate(string) error {
	return v.err
}

func TestServiceCreatesAccountAndVerifiesJWT(t *testing.T) {
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer redisServer.Close()

	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer client.Close()

	now := time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC)
	service := NewService(client, Config{
		Namespace: "vote:",
		JWTSecret: "player-secret",
		TokenTTL:  time.Hour,
		Now: func() time.Time {
			return now
		},
	}, nicknameValidator{})

	token, _, err := service.Login(context.Background(), "阿明", "hunter2")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	nickname, err := service.Verify(context.Background(), token)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if nickname != "阿明" {
		t.Fatalf("expected nickname 阿明, got %q", nickname)
	}

	authRecord, err := client.HGetAll(context.Background(), "vote:player-auth:阿明").Result()
	if err != nil {
		t.Fatalf("read auth record: %v", err)
	}
	if authRecord["password_hash"] == "" {
		t.Fatal("expected password hash to be stored")
	}
	if authRecord["password_hash"] == "hunter2" {
		t.Fatal("expected password to be hashed instead of stored in plaintext")
	}
}

func TestServiceResetPasswordInvalidatesExistingJWT(t *testing.T) {
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer redisServer.Close()

	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer client.Close()

	now := time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC)
	service := NewService(client, Config{
		Namespace: "vote:",
		JWTSecret: "player-secret",
		TokenTTL:  time.Hour,
		Now: func() time.Time {
			return now
		},
	}, nicknameValidator{})

	token, _, err := service.Login(context.Background(), "阿明", "old-password")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	now = now.Add(2 * time.Minute)
	if err := service.ResetPassword(context.Background(), "阿明", "new-password"); err != nil {
		t.Fatalf("reset password: %v", err)
	}

	if _, err := service.Verify(context.Background(), token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected old token to be invalid after password reset, got %v", err)
	}

	if _, _, err := service.Login(context.Background(), "阿明", "old-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected old password to fail after reset, got %v", err)
	}

	newToken, _, err := service.Login(context.Background(), "阿明", "new-password")
	if err != nil {
		t.Fatalf("login with new password: %v", err)
	}

	nickname, err := service.Verify(context.Background(), newToken)
	if err != nil {
		t.Fatalf("verify new token: %v", err)
	}
	if nickname != "阿明" {
		t.Fatalf("expected nickname 阿明, got %q", nickname)
	}
}

func TestServiceRejectsInvalidNicknameAndPassword(t *testing.T) {
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer redisServer.Close()

	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer client.Close()

	service := NewService(client, Config{
		Namespace: "vote:",
		JWTSecret: "player-secret",
		TokenTTL:  time.Hour,
	}, nicknameValidator{err: core.ErrSensitiveNickname})

	if _, _, err := service.Login(context.Background(), "阿明", "   "); !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("expected blank password to be rejected, got %v", err)
	}

	if _, _, err := service.Login(context.Background(), "阿明", "password"); !errors.Is(err, core.ErrSensitiveNickname) {
		t.Fatalf("expected nickname validator error, got %v", err)
	}
}
