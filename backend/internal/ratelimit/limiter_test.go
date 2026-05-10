package ratelimit

import (
	"testing"
	"time"
)

func TestLimiterDetectsBurstOverflow(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	limiter := NewLimiter(Config{
		Rules: []WindowConfig{
			{Limit: 3, Window: 2 * time.Second},
		},
		Now: func() time.Time {
			return now
		},
	})

	for range 3 {
		hit, err := limiter.Detect("203.0.113.10")
		if err != nil {
			t.Fatalf("detect hit: %v", err)
		}
		if hit {
			t.Fatal("expected hit to stay false before threshold")
		}
	}

	hit, err := limiter.Detect("203.0.113.10")
	if err != nil {
		t.Fatalf("detect overflow: %v", err)
	}
	if !hit {
		t.Fatal("expected fourth request to trigger detector")
	}

	hit, err = limiter.Detect("203.0.113.10")
	if err != nil {
		t.Fatalf("detect repeated overflow: %v", err)
	}
	if hit {
		t.Fatal("expected repeated overflow to stay quiet after the first trigger")
	}
}

func TestLimiterKeepsDifferentKeysIndependent(t *testing.T) {
	limiter := NewLimiter(Config{
		Rules: []WindowConfig{
			{Limit: 1, Window: time.Second},
		},
	})

	if hit, err := limiter.Detect("198.51.100.10"); err != nil || hit {
		t.Fatalf("expected first key first hit to be normal, hit=%v err=%v", hit, err)
	}
	if hit, err := limiter.Detect("198.51.100.10"); err != nil || !hit {
		t.Fatalf("expected first key second hit to overflow, hit=%v err=%v", hit, err)
	}
	if hit, err := limiter.Detect("198.51.100.11"); err != nil || hit {
		t.Fatalf("expected second key to stay independent, hit=%v err=%v", hit, err)
	}
}

func TestLimiterUsesWindowDefaults(t *testing.T) {
	limiter := NewLimiter(Config{
		Rules: []WindowConfig{
			{Window: 2 * time.Second},
		},
	})

	for i := range 42 {
		hit, err := limiter.Detect("203.0.113.20")
		if err != nil {
			t.Fatalf("detect %d: %v", i+1, err)
		}
		if hit {
			t.Fatalf("expected request %d to stay below default threshold", i+1)
		}
	}

	hit, err := limiter.Detect("203.0.113.20")
	if err != nil {
		t.Fatalf("detect 43: %v", err)
	}
	if !hit {
		t.Fatal("expected request 43 to overflow default threshold")
	}
}

func TestLimiterSuppressesRepeatedOverflowSignalsUntilWindowResets(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	limiter := NewLimiter(Config{
		Rules: []WindowConfig{
			{Limit: 2, Window: 2 * time.Second},
		},
		Now: func() time.Time {
			return now
		},
	})

	for i := range 2 {
		hit, err := limiter.Detect("burst")
		if err != nil {
			t.Fatalf("detect warmup %d: %v", i+1, err)
		}
		if hit {
			t.Fatalf("expected warmup request %d to stay below threshold", i+1)
		}
	}

	hit, err := limiter.Detect("burst")
	if err != nil {
		t.Fatalf("detect first overflow: %v", err)
	}
	if !hit {
		t.Fatal("expected first overflow to trigger detector")
	}

	hit, err = limiter.Detect("burst")
	if err != nil {
		t.Fatalf("detect repeated overflow: %v", err)
	}
	if hit {
		t.Fatal("expected repeated overflow to stay suppressed")
	}

	now = now.Add(3 * time.Second)
	hit, err = limiter.Detect("burst")
	if err != nil {
		t.Fatalf("detect after window reset: %v", err)
	}
	if hit {
		t.Fatal("expected post-reset request to stay below threshold")
	}
}

func TestLimiterDetectsSecondRuleWindow(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	limiter := NewLimiter(Config{
		Rules: []WindowConfig{
			{Limit: 10, Window: time.Second},
			{Limit: 5, Window: 10 * time.Second},
		},
		Now: func() time.Time {
			return now
		},
	})

	for i := range 5 {
		hit, err := limiter.Detect("medium")
		if err != nil {
			t.Fatalf("medium detect %d: %v", i+1, err)
		}
		if hit {
			t.Fatalf("expected second-rule detect %d below threshold", i+1)
		}
		now = now.Add(2 * time.Second)
	}
	hit, err := limiter.Detect("medium")
	if err != nil {
		t.Fatalf("second-rule overflow detect: %v", err)
	}
	if !hit {
		t.Fatal("expected second rule window overflow")
	}
}

func TestLimiterDetectsThirdRuleWindow(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	limiter := NewLimiter(Config{
		Rules: []WindowConfig{
			{Limit: 10, Window: time.Second},
			{Limit: 20, Window: 10 * time.Second},
			{Limit: 3, Window: time.Hour},
		},
		Now: func() time.Time {
			return now
		},
	})

	now = time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	var hit bool
	var err error
	for i := range 3 {
		hit, err = limiter.Detect("long")
		if err != nil {
			t.Fatalf("long detect %d: %v", i+1, err)
		}
		if hit {
			t.Fatalf("expected long detect %d below threshold", i+1)
		}
		now = now.Add(20 * time.Minute)
	}
	hit, err = limiter.Detect("long")
	if err != nil {
		t.Fatalf("long overflow detect: %v", err)
	}
	if !hit {
		t.Fatal("expected long window overflow")
	}
}

func TestLimiterDetectCountReportsEveryNewlyTriggeredWindow(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	limiter := NewLimiter(Config{
		Rules: []WindowConfig{
			{Limit: 2, Window: time.Second},
			{Limit: 2, Window: 2 * time.Second},
		},
		Now: func() time.Time {
			return now
		},
	})

	for i := range 2 {
		hitCount, err := limiter.DetectCount("multi")
		if err != nil {
			t.Fatalf("detect warmup %d: %v", i+1, err)
		}
		if hitCount != 0 {
			t.Fatalf("expected warmup request %d not to trigger, got %d", i+1, hitCount)
		}
	}

	hitCount, err := limiter.DetectCount("multi")
	if err != nil {
		t.Fatalf("detect first overflow: %v", err)
	}
	if hitCount != 2 {
		t.Fatalf("expected two windows to trigger together, got %d", hitCount)
	}

	hitCount, err = limiter.DetectCount("multi")
	if err != nil {
		t.Fatalf("detect repeated overflow: %v", err)
	}
	if hitCount != 0 {
		t.Fatalf("expected repeated overflow to stay suppressed, got %d", hitCount)
	}
}

func TestLimiterDetectCountCanReportLaterWindowAfterEarlierWindowAlreadyOverflowed(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	limiter := NewLimiter(Config{
		Rules: []WindowConfig{
			{Limit: 2, Window: time.Second},
			{Limit: 3, Window: 10 * time.Second},
		},
		Now: func() time.Time {
			return now
		},
	})

	for i := range 2 {
		hitCount, err := limiter.DetectCount("staggered")
		if err != nil {
			t.Fatalf("detect warmup %d: %v", i+1, err)
		}
		if hitCount != 0 {
			t.Fatalf("expected warmup request %d not to trigger, got %d", i+1, hitCount)
		}
	}

	hitCount, err := limiter.DetectCount("staggered")
	if err != nil {
		t.Fatalf("detect short-window overflow: %v", err)
	}
	if hitCount != 1 {
		t.Fatalf("expected only short window to trigger first, got %d", hitCount)
	}

	now = now.Add(2 * time.Second)
	hitCount, err = limiter.DetectCount("staggered")
	if err != nil {
		t.Fatalf("detect long-window overflow: %v", err)
	}
	if hitCount != 1 {
		t.Fatalf("expected later long window to trigger independently, got %d", hitCount)
	}
}
