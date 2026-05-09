package ratelimit

import (
	"sync"
	"time"
)

type WindowConfig struct {
	Limit  int
	Window time.Duration
}

// Config 定义点击异常探测窗口。
type Config struct {
	Limit  int
	Window time.Duration
	Medium WindowConfig
	Long   WindowConfig
	Now    func() time.Time
}

type clientState struct {
	hits []time.Time
}

// Limiter 仅负责命中异常窗口探测，不执行封禁。
type Limiter struct {
	mu        sync.Mutex
	short     WindowConfig
	medium    WindowConfig
	long      WindowConfig
	maxWindow time.Duration
	now       func() time.Time
	clients   map[string]*clientState
}

func normalizeWindowConfig(limit int, window time.Duration, defaultLimit int, defaultWindow time.Duration) WindowConfig {
	if limit <= 0 {
		limit = defaultLimit
	}
	if window <= 0 {
		window = defaultWindow
	}
	return WindowConfig{
		Limit:  limit,
		Window: window,
	}
}

func enabledWindowConfig(cfg WindowConfig) WindowConfig {
	if cfg.Limit <= 0 || cfg.Window <= 0 {
		return WindowConfig{}
	}
	return cfg
}

func maxDuration(values ...time.Duration) time.Duration {
	var max time.Duration
	for _, value := range values {
		if value > max {
			max = value
		}
	}
	return max
}

func NewLimiter(config Config) *Limiter {
	short := normalizeWindowConfig(config.Limit, config.Window, 42, 2*time.Second)
	medium := enabledWindowConfig(config.Medium)
	long := enabledWindowConfig(config.Long)
	now := config.Now
	if now == nil {
		now = time.Now
	}
	return &Limiter{
		short:     short,
		medium:    medium,
		long:      long,
		maxWindow: maxDuration(short.Window, medium.Window, long.Window),
		now:       now,
		clients:   make(map[string]*clientState),
	}
}

func countHitsInWindow(hits []time.Time, cutoff time.Time) int {
	count := 0
	for _, hit := range hits {
		if !hit.Before(cutoff) {
			count++
		}
	}
	return count
}

func exceedsWindowLimit(hits []time.Time, now time.Time, window WindowConfig) bool {
	if window.Limit <= 0 || window.Window <= 0 {
		return false
	}
	return countHitsInWindow(hits, now.Add(-window.Window)) > window.Limit
}

func exceedsAnyWindowLimit(hits []time.Time, now time.Time, short WindowConfig, medium WindowConfig, long WindowConfig) bool {
	return exceedsWindowLimit(hits, now, short) ||
		exceedsWindowLimit(hits, now, medium) ||
		exceedsWindowLimit(hits, now, long)
}

func (l *Limiter) Detect(clientID string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	state := l.clients[clientID]
	if state == nil {
		state = &clientState{}
		l.clients[clientID] = state
	}

	cutoff := now.Add(-l.maxWindow)
	filtered := state.hits[:0]
	for _, hit := range state.hits {
		if !hit.Before(cutoff) {
			filtered = append(filtered, hit)
		}
	}
	state.hits = filtered
	prevOverflow := exceedsAnyWindowLimit(state.hits, now, l.short, l.medium, l.long)
	state.hits = append(state.hits, now)

	return !prevOverflow && exceedsAnyWindowLimit(state.hits, now, l.short, l.medium, l.long), nil
}
