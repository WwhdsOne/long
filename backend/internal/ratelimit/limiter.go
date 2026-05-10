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
	Rules []WindowConfig
	Now   func() time.Time
}

type clientState struct {
	hits []time.Time
}

// Limiter 仅负责命中异常窗口探测，不执行封禁。
type Limiter struct {
	mu        sync.Mutex
	rules     []WindowConfig
	maxWindow time.Duration
	now       func() time.Time
	clients   map[string]*clientState
}

func normalizeWindowConfig(cfg WindowConfig, defaultLimit int, defaultWindow time.Duration) WindowConfig {
	limit := cfg.Limit
	window := cfg.Window
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

func normalizeRules(rules []WindowConfig) []WindowConfig {
	if len(rules) == 0 {
		return []WindowConfig{
			normalizeWindowConfig(WindowConfig{}, 42, 2*time.Second),
		}
	}
	normalized := make([]WindowConfig, 0, len(rules))
	for _, rule := range rules {
		normalized = append(normalized, normalizeWindowConfig(rule, 42, 2*time.Second))
	}
	return normalized
}

func maxDuration(values []time.Duration) time.Duration {
	var max time.Duration
	for _, value := range values {
		if value > max {
			max = value
		}
	}
	return max
}

func NewLimiter(config Config) *Limiter {
	rules := normalizeRules(config.Rules)
	windows := make([]time.Duration, 0, len(rules))
	for _, rule := range rules {
		windows = append(windows, rule.Window)
	}
	now := config.Now
	if now == nil {
		now = time.Now
	}
	return &Limiter{
		rules:     rules,
		maxWindow: maxDuration(windows),
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

func exceedsAnyWindowLimit(hits []time.Time, now time.Time, rules []WindowConfig) bool {
	for _, rule := range rules {
		if exceedsWindowLimit(hits, now, rule) {
			return true
		}
	}
	return false
}

func countNewlyExceededWindows(prevHits []time.Time, currentHits []time.Time, now time.Time, rules []WindowConfig) int {
	count := 0
	for _, rule := range rules {
		if exceedsWindowLimit(prevHits, now, rule) {
			continue
		}
		if exceedsWindowLimit(currentHits, now, rule) {
			count++
		}
	}
	return count
}

func (l *Limiter) DetectCount(clientID string) (int, error) {
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
	prevHits := append([]time.Time(nil), state.hits...)
	state.hits = append(state.hits, now)

	return countNewlyExceededWindows(prevHits, state.hits, now, l.rules), nil
}

func (l *Limiter) Detect(clientID string) (bool, error) {
	hitCount, err := l.DetectCount(clientID)
	if err != nil {
		return false, err
	}
	return hitCount > 0, nil
}
