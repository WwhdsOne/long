package ratelimit

import (
	"errors"
	"sync"
	"time"

	"long/internal/core"
)

var ErrTooManyRequests = errors.New("too many requests")

// Config defines the burst window and blacklist duration for click abuse control.
type Config struct {
	Limit             int
	Window            time.Duration
	BlacklistDuration time.Duration
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
	limit             int
	window            time.Duration
	blacklistDuration time.Duration
	now               func() time.Time
	clients           map[string]*clientState
}

// NewLimiter creates a limiter with single-instance defaults for this project.
func NewLimiter(config Config) *Limiter {
	limit := config.Limit
	if limit <= 0 {
		limit = 42
	}

	window := config.Window
	if window <= 0 {
		window = 2 * time.Second
	}

	blacklistDuration := config.BlacklistDuration
	if blacklistDuration <= 0 {
		blacklistDuration = 10 * time.Minute
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	return &Limiter{
		limit:             limit,
		window:            window,
		blacklistDuration: blacklistDuration,
		now:               now,
		clients:           make(map[string]*clientState),
	}
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

	cutoff := now.Add(-l.window)
	filtered := state.hits[:0]
	for _, hit := range state.hits {
		if !hit.Before(cutoff) {
			filtered = append(filtered, hit)
		}
	}
	state.hits = filtered

	state.hits = append(state.hits, now)
	if len(state.hits) > l.limit {
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
