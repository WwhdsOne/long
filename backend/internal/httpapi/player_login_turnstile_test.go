package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bytedance/sonic"

	playerauth "long/internal/playerauth"
)

type mockPlayerLoginTurnstile struct {
	result      PlayerLoginTurnstileResult
	lastRequest PlayerLoginTurnstileRequest
}

func (m *mockPlayerLoginTurnstile) CheckPlayerLogin(_ context.Context, req PlayerLoginTurnstileRequest) (PlayerLoginTurnstileResult, error) {
	m.lastRequest = req
	return m.result, nil
}

func TestPlayerLoginRequiresCaptchaWhenEnabled(t *testing.T) {
	handler := NewHandler(Options{
		Store:               &mockStore{state: voteStateForPlayerTests()},
		Broadcaster:         &mockBroadcaster{},
		PlayerAuthenticator: &mockPlayerAuthenticator{loginToken: "player-token", loginNickname: "阿明"},
		PlayerLoginTurnstile: &mockPlayerLoginTurnstile{
			result: PlayerLoginTurnstileResult{
				Decision: PlayerLoginTurnstileRequire,
				SiteKey:  "site-key",
			},
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/player/auth/login", strings.NewReader(`{"nickname":"阿明","password":"secret"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Forwarded-For", "198.51.100.10")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", response.Code, response.Body.String())
	}

	var payload map[string]any
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload["error"] != "CAPTCHA_REQUIRED" {
		t.Fatalf("expected CAPTCHA_REQUIRED, got %+v", payload)
	}
	if payload["siteKey"] != "site-key" {
		t.Fatalf("expected site key in response, got %+v", payload)
	}
}

func TestPlayerLoginRejectsInvalidCaptchaToken(t *testing.T) {
	authenticator := &mockPlayerAuthenticator{loginToken: "player-token", loginNickname: "阿明"}
	turnstile := &mockPlayerLoginTurnstile{
		result: PlayerLoginTurnstileResult{Decision: PlayerLoginTurnstileInvalid},
	}
	handler := NewHandler(Options{
		Store:                &mockStore{state: voteStateForPlayerTests()},
		Broadcaster:          &mockBroadcaster{},
		PlayerAuthenticator:  authenticator,
		PlayerLoginTurnstile: turnstile,
	})

	request := httptest.NewRequest(http.MethodPost, "/api/player/auth/login", strings.NewReader(`{"nickname":"阿明","password":"secret","turnstileToken":"login-token"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", response.Code, response.Body.String())
	}
	if turnstile.lastRequest.Token != "login-token" {
		t.Fatalf("expected login token to be forwarded, got %q", turnstile.lastRequest.Token)
	}
	if !strings.Contains(response.Body.String(), "CAPTCHA_INVALID") {
		t.Fatalf("expected CAPTCHA_INVALID, got %s", response.Body.String())
	}
}

func TestPlayerLoginAllowsValidCaptchaToken(t *testing.T) {
	authenticator := &mockPlayerAuthenticator{
		loginToken:     "player-token",
		loginNickname:  "阿明",
		verifyNickname: "阿明",
	}
	turnstile := &mockPlayerLoginTurnstile{
		result: PlayerLoginTurnstileResult{Decision: PlayerLoginTurnstileAllow},
	}
	handler := NewHandler(Options{
		Store:                &mockStore{state: voteStateForPlayerTests()},
		Broadcaster:          &mockBroadcaster{},
		PlayerAuthenticator:  authenticator,
		PlayerLoginTurnstile: turnstile,
	})

	request := httptest.NewRequest(http.MethodPost, "/api/player/auth/login", strings.NewReader(`{"nickname":"阿明","password":"secret","turnstileToken":"login-token"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Forwarded-For", "198.51.100.10")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", response.Code, response.Body.String())
	}
	if turnstile.lastRequest.RemoteIP != "198.51.100.10" {
		t.Fatalf("expected forwarded ip, got %q", turnstile.lastRequest.RemoteIP)
	}
}

func TestPlayerLoginReturnsInvalidCredentialsAfterCaptchaPass(t *testing.T) {
	handler := NewHandler(Options{
		Store:               &mockStore{state: voteStateForPlayerTests()},
		Broadcaster:         &mockBroadcaster{},
		PlayerAuthenticator: &mockPlayerAuthenticator{loginErr: playerauth.ErrInvalidCredentials},
		PlayerLoginTurnstile: &mockPlayerLoginTurnstile{
			result: PlayerLoginTurnstileResult{Decision: PlayerLoginTurnstileAllow},
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/player/auth/login", strings.NewReader(`{"nickname":"阿明","password":"bad","turnstileToken":"login-token"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "INVALID_CREDENTIALS") {
		t.Fatalf("expected INVALID_CREDENTIALS, got %s", response.Body.String())
	}
}
