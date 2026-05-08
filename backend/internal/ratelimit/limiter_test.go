package ratelimit

import (
	"testing"
	"time"
)

func TestLimiterDetectsBurstOverflow(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	limiter := NewLimiter(Config{
		Limit:  3,
		Window: 2 * time.Second,
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
}

func TestLimiterKeepsDifferentKeysIndependent(t *testing.T) {
	limiter := NewLimiter(Config{
		Limit:  1,
		Window: time.Second,
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
		Window: 2 * time.Second,
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

func TestLimiterDetectsMediumWindow(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	limiter := NewLimiter(Config{
		Limit:  10,
		Window: time.Second,
		Medium: WindowConfig{
			Limit:  5,
			Window: 10 * time.Second,
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
			t.Fatalf("expected medium detect %d below threshold", i+1)
		}
		now = now.Add(2 * time.Second)
	}
	hit, err := limiter.Detect("medium")
	if err != nil {
		t.Fatalf("medium overflow detect: %v", err)
	}
	if !hit {
		t.Fatal("expected medium window overflow")
	}
}

func TestLimiterDetectsLongWindow(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	limiter := NewLimiter(Config{
		Limit:  10,
		Window: time.Second,
		Long: WindowConfig{
			Limit:  3,
			Window: time.Hour,
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
