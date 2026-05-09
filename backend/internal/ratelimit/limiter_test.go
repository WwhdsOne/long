package ratelimit

import (
	"testing"
	"time"
)

func TestLimiterDetectsBurstOverflow(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	limiter := NewLimiter(Config{
		Rules: []WindowConfig{{Limit: 3, Window: 2 * time.Second}},
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
		Rules: []WindowConfig{{Limit: 1, Window: time.Second}},
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

func TestLimiterUsesMultipleRules(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	limiter := NewLimiter(Config{
		Rules: []WindowConfig{
			{Limit: 3, Window: time.Second},
			{Limit: 5, Window: 10 * time.Second},
		},
		Now: func() time.Time {
			return now
		},
	})

	// Stay below both limits (limit=3 for short)
	for i := range 3 {
		hit, err := limiter.Detect("multi")
		if err != nil {
			t.Fatalf("multi detect %d: %v", i+1, err)
		}
		if hit {
			t.Fatalf("expected multi detect %d below both thresholds", i+1)
		}
	}

	// Exceed short limit on 4th call
	hit, err := limiter.Detect("multi")
	if err != nil {
		t.Fatalf("multi overflow detect: %v", err)
	}
	if !hit {
		t.Fatal("expected short window overflow")
	}
}

func TestLimiterSuppressesRepeatedOverflowSignalsUntilWindowResets(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	limiter := NewLimiter(Config{
		Rules: []WindowConfig{{Limit: 2, Window: 2 * time.Second}},
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

func TestLimiterDetectsMultiTierWindow(t *testing.T) {
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

	// Stay below short limit but trigger medium
	for i := range 5 {
		hit, err := limiter.Detect("tier")
		if err != nil {
			t.Fatalf("tier detect %d: %v", i+1, err)
		}
		if hit {
			t.Fatalf("expected tier detect %d below thresholds", i+1)
		}
		now = now.Add(2 * time.Second)
	}
	hit, err := limiter.Detect("tier")
	if err != nil {
		t.Fatalf("tier overflow detect: %v", err)
	}
	if !hit {
		t.Fatal("expected multi-tier window overflow")
	}
}

func TestLimiterDetectsLongWindowInMultiTier(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	limiter := NewLimiter(Config{
		Rules: []WindowConfig{
			{Limit: 10, Window: time.Second},
			{Limit: 3, Window: time.Hour},
		},
		Now: func() time.Time {
			return now
		},
	})

	for i := range 3 {
		hit, err := limiter.Detect("long")
		if err != nil {
			t.Fatalf("long detect %d: %v", i+1, err)
		}
		if hit {
			t.Fatalf("expected long detect %d below threshold", i+1)
		}
		now = now.Add(20 * time.Minute)
	}
	hit, err := limiter.Detect("long")
	if err != nil {
		t.Fatalf("long overflow detect: %v", err)
	}
	if !hit {
		t.Fatal("expected long window overflow")
	}
}
