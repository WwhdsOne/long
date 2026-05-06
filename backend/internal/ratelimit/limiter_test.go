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

	for i := range 42 {
		if _, err := limiter.Allow("203.0.113.20"); err != nil {
			t.Fatalf("expected request %d to pass before default threshold, got %v", i+1, err)
		}
	}

	if _, err := limiter.Allow("203.0.113.20"); err == nil {
		t.Fatal("expected request 43 to be rate limited")
	}
}

func TestLimiterListsAndUnblocksBlacklistEntries(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	limiter := NewLimiter(Config{
		Limit:             1,
		Window:            time.Second,
		BlacklistDuration: 10 * time.Minute,
		Now: func() time.Time {
			return now
		},
	})

	if _, err := limiter.Allow("203.0.113.99"); err != nil {
		t.Fatalf("expected first request to pass, got %v", err)
	}
	if _, err := limiter.Allow("203.0.113.99"); err == nil {
		t.Fatal("expected client to be blocked")
	}

	entries := limiter.ListBlacklist()
	if len(entries) != 1 {
		t.Fatalf("expected 1 blacklist entry, got %d", len(entries))
	}
	if entries[0].ClientID != "203.0.113.99" {
		t.Fatalf("expected client id 203.0.113.99, got %s", entries[0].ClientID)
	}
	if entries[0].BlockedAt == 0 {
		t.Fatal("expected blocked at to be recorded")
	}

	if !limiter.Unblock("203.0.113.99") {
		t.Fatal("expected unblock to succeed")
	}
	if got := limiter.ListBlacklist(); len(got) != 0 {
		t.Fatalf("expected blacklist to be empty after unblock, got %d", len(got))
	}
	if _, err := limiter.Allow("203.0.113.99"); err != nil {
		t.Fatalf("expected client to pass after unblock, got %v", err)
	}
}

func TestLimiterBlocksClientWhenMediumWindowExceedsLimit(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	limiter := NewLimiter(Config{
		Limit:             10,
		Window:            time.Second,
		BlacklistDuration: 10 * time.Minute,
		Medium: WindowConfig{
			Limit:  5,
			Window: 10 * time.Second,
		},
		Now: func() time.Time {
			return now
		},
	})

	for i := range 5 {
		if _, err := limiter.Allow("203.0.113.30"); err != nil {
			t.Fatalf("expected request %d to pass before medium threshold, got %v", i+1, err)
		}
		now = now.Add(2 * time.Second)
	}

	retryAfter, err := limiter.Allow("203.0.113.30")
	if err == nil {
		t.Fatal("expected medium window overflow to be rate limited")
	}
	if retryAfter != 10*time.Minute {
		t.Fatalf("expected retryAfter 10m, got %s", retryAfter)
	}
}

func TestLimiterBlocksClientWhenLongWindowExceedsLimit(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	limiter := NewLimiter(Config{
		Limit:             10,
		Window:            time.Second,
		BlacklistDuration: 10 * time.Minute,
		Long: WindowConfig{
			Limit:  3,
			Window: time.Hour,
		},
		Now: func() time.Time {
			return now
		},
	})

	for i := range 3 {
		if _, err := limiter.Allow("203.0.113.40"); err != nil {
			t.Fatalf("expected request %d to pass before long threshold, got %v", i+1, err)
		}
		now = now.Add(20 * time.Minute)
	}

	retryAfter, err := limiter.Allow("203.0.113.40")
	if err == nil {
		t.Fatal("expected long window overflow to be rate limited")
	}
	if retryAfter != 10*time.Minute {
		t.Fatalf("expected retryAfter 10m, got %s", retryAfter)
	}
}
