package ratelimit

import (
	"testing"
	"time"
)

func TestLimiterBlocksBurstingClientForTenMinutes(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	limiter := NewLimiter(Config{
		Limit:             3,
		Window:            2 * time.Second,
		BlacklistDuration: 10 * time.Minute,
		Now: func() time.Time {
			return now
		},
	})

	for range 3 {
		if _, err := limiter.Allow("203.0.113.10"); err != nil {
			t.Fatalf("expected request to pass before threshold, got %v", err)
		}
	}

	retryAfter, err := limiter.Allow("203.0.113.10")
	if err == nil {
		t.Fatal("expected fourth request to be rate limited")
	}
	if retryAfter != 10*time.Minute {
		t.Fatalf("expected retryAfter 10m, got %s", retryAfter)
	}

	now = now.Add(5 * time.Minute)
	retryAfter, err = limiter.Allow("203.0.113.10")
	if err == nil {
		t.Fatal("expected client to stay blocked during blacklist window")
	}
	if retryAfter != 5*time.Minute {
		t.Fatalf("expected remaining block time 5m, got %s", retryAfter)
	}

	now = now.Add(5*time.Minute + time.Second)
	if _, err := limiter.Allow("203.0.113.10"); err != nil {
		t.Fatalf("expected client to recover after blacklist expires, got %v", err)
	}
}

func TestLimiterKeepsDifferentIPsIndependent(t *testing.T) {
	limiter := NewLimiter(Config{
		Limit:             1,
		Window:            2 * time.Second,
		BlacklistDuration: 10 * time.Minute,
	})

	if _, err := limiter.Allow("198.51.100.10"); err != nil {
		t.Fatalf("expected first ip to pass, got %v", err)
	}
	if _, err := limiter.Allow("198.51.100.10"); err == nil {
		t.Fatal("expected first ip to be blocked on second hit")
	}

	if _, err := limiter.Allow("198.51.100.11"); err != nil {
		t.Fatalf("expected second ip to stay independent, got %v", err)
	}
}

func TestLimiterUsesFortyTwoAsDefaultBurstLimit(t *testing.T) {
	limiter := NewLimiter(Config{
		Window:            2 * time.Second,
		BlacklistDuration: 10 * time.Minute,
	})

	for i := 0; i < 42; i++ {
		if _, err := limiter.Allow("203.0.113.20"); err != nil {
			t.Fatalf("expected request %d to pass before default threshold, got %v", i+1, err)
		}
	}

	if _, err := limiter.Allow("203.0.113.20"); err == nil {
		t.Fatal("expected request 43 to be rate limited")
	}
}
