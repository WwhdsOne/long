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
	var maxWindow time.Duration
	for _, r := range config.Rules {
		if r.Window > maxWindow {
			maxWindow = r.Window
		}
	}
	now := config.Now
	if now == nil {
		now = time.Now
	}
	return &Limiter{
		rules:     config.Rules,
		maxWindow: maxWindow,
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
	for _, r := range rules {
		if exceedsWindowLimit(hits, now, r) {
			return true
		}
	}
	return false
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
	prevOverflow := exceedsAnyWindowLimit(state.hits, now, l.rules)
	state.hits = append(state.hits, now)

	return !prevOverflow && exceedsAnyWindowLimit(state.hits, now, l.rules), nil
}
