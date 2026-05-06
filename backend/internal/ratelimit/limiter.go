package ratelimit

import (
	"errors"
	"sync"
	"time"

	"long/internal/core"
)

var ErrTooManyRequests = errors.New("too many requests")

type WindowConfig struct {
	Limit  int
	Window time.Duration
}

// Config defines the burst window and blacklist duration for click abuse control.
type Config struct {
	Limit             int
	Window            time.Duration
	BlacklistDuration time.Duration
	Medium            WindowConfig
	Long              WindowConfig
	Now               func() time.Time
}

type clientState struct {
	hits         []time.Time
	blockedAt    time.Time
	blockedUntil time.Time
}

// Limiter tracks recent click bursts per client in memory.
type Limiter struct {
	mu                sync.Mutex
	short             WindowConfig
	medium            WindowConfig
	long              WindowConfig
	maxWindow         time.Duration
	blacklistDuration time.Duration
	now               func() time.Time
	clients           map[string]*clientState
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

// NewLimiter creates a limiter with single-instance defaults for this project.
func NewLimiter(config Config) *Limiter {
	short := normalizeWindowConfig(config.Limit, config.Window, 42, 2*time.Second)
	medium := enabledWindowConfig(config.Medium)
	long := enabledWindowConfig(config.Long)

	blacklistDuration := config.BlacklistDuration
	if blacklistDuration <= 0 {
		blacklistDuration = 10 * time.Minute
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	return &Limiter{
		short:             short,
		medium:            medium,
		long:              long,
		maxWindow:         maxDuration(short.Window, medium.Window, long.Window),
		blacklistDuration: blacklistDuration,
		now:               now,
		clients:           make(map[string]*clientState),
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

// Allow records one click attempt and returns a block duration when abuse is detected.
func (l *Limiter) Allow(clientID string) (time.Duration, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	state := l.clients[clientID]
	if state == nil {
		state = &clientState{}
		l.clients[clientID] = state
	}

	if state.blockedUntil.After(now) {
		return state.blockedUntil.Sub(now), ErrTooManyRequests
	}

	cutoff := now.Add(-l.maxWindow)
	filtered := state.hits[:0]
	for _, hit := range state.hits {
		if !hit.Before(cutoff) {
			filtered = append(filtered, hit)
		}
	}
	state.hits = filtered

	state.hits = append(state.hits, now)
	if exceedsWindowLimit(state.hits, now, l.short) ||
		exceedsWindowLimit(state.hits, now, l.medium) ||
		exceedsWindowLimit(state.hits, now, l.long) {
		state.hits = nil
		state.blockedAt = now
		state.blockedUntil = now.Add(l.blacklistDuration)
		return l.blacklistDuration, ErrTooManyRequests
	}

	return 0, nil
}

// ListBlacklist returns current active blocked clients.
func (l *Limiter) ListBlacklist() []core.BlacklistEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	entries := make([]core.BlacklistEntry, 0, len(l.clients))
	for clientID, state := range l.clients {
		if state == nil || !state.blockedUntil.After(now) {
			continue
		}
		nickname := clientID
		switch {
		case len(clientID) > len("nickname:") && clientID[:len("nickname:")] == "nickname:":
			nickname = clientID[len("nickname:"):]
		case len(clientID) > len("ip:") && clientID[:len("ip:")] == "ip:":
			nickname = "IP 封禁"
		}
		entries = append(entries, core.BlacklistEntry{
			ClientID:         clientID,
			Nickname:         nickname,
			BlockedAt:        state.blockedAt.Unix(),
			BlockedUntil:     state.blockedUntil.Unix(),
			RemainingSeconds: int64(state.blockedUntil.Sub(now) / time.Second),
		})
	}
	return entries
}

// Unblock removes a client from blacklist immediately.
func (l *Limiter) Unblock(clientID string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	state := l.clients[clientID]
	if state == nil {
		return false
	}
	state.hits = nil
	state.blockedAt = time.Time{}
	state.blockedUntil = time.Time{}
	return true
}
