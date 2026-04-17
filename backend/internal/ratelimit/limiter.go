package ratelimit

import (
	"errors"
	"sync"
	"time"
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
		limit = 12
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
		state.blockedUntil = now.Add(l.blacklistDuration)
		return l.blacklistDuration, ErrTooManyRequests
	}

	return 0, nil
}
