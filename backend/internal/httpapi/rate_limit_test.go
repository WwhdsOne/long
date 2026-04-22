package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"long/internal/ratelimit"
	"long/internal/vote"
)

type blockingLimiter struct {
	retryAfter time.Duration
	err        error
}

func (b blockingLimiter) Allow(string) (time.Duration, error) {
	return b.retryAfter, b.err
}

type selectiveLimiter struct {
	blockedKey string
}

func (s selectiveLimiter) Allow(key string) (time.Duration, error) {
	if key == s.blockedKey {
		return 10 * time.Minute, ratelimit.ErrTooManyRequests
	}
	return 0, nil
}

func TestClickButtonReturnsTooManyRequestsWhenClientIsBlocked(t *testing.T) {
	store := &mockStore{
		state: vote.State{
			Buttons: []vote.Button{
				{
					Key:      "feel",
					RedisKey: "vote:button:feel",
					Label:    "有感觉吗",
					Count:    2,
					Sort:     10,
					Enabled:  true,
				},
			},
		},
	}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
		ClickGuard: blockingLimiter{
			retryAfter: 10 * time.Minute,
			err:        ratelimit.ErrTooManyRequests,
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/buttons/feel/click", strings.NewReader(`{"nickname":"阿明"}`))
	request.RemoteAddr = "203.0.113.30:4567"
	request.Header.Set("X-Forwarded-For", "198.51.100.24, 203.0.113.30")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", response.Code)
	}
	if retryAfter := response.Header().Get("Retry-After"); retryAfter != "600" {
		t.Fatalf("expected Retry-After 600, got %q", retryAfter)
	}
	if body := response.Body.String(); body == "" {
		t.Fatal("expected response body for rate limit error")
	}
}

func TestClickButtonBlocksSameNicknameAcrossDifferentIPs(t *testing.T) {
	store := &mockStore{
		state: vote.State{
			Buttons: []vote.Button{
				{
					Key:      "feel",
					RedisKey: "vote:button:feel",
					Label:    "有感觉吗",
					Count:    2,
					Sort:     10,
					Enabled:  true,
				},
			},
		},
	}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
		ClickGuard: selectiveLimiter{
			blockedKey: "nickname:阿明",
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/buttons/feel/click", strings.NewReader(`{"nickname":"阿明"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("expected click to be blocked by nickname limit, got %d", response.Code)
	}
}
