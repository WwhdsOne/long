package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bytedance/sonic"

	"long/internal/core"
)

type mockStaminaPurchaseTurnstile struct {
	result      StaminaPurchaseTurnstileResult
	lastRequest StaminaPurchaseTurnstileRequest
}

func (m *mockStaminaPurchaseTurnstile) CheckPurchaseStamina(ctx context.Context, req StaminaPurchaseTurnstileRequest) (StaminaPurchaseTurnstileResult, error) {
	m.lastRequest = req
	return m.result, nil
}

func TestPurchaseStaminaFullAllowsRequestWhenCaptchaNotRequired(t *testing.T) {
	store := &mockStore{
		state: core.State{
			UserStats: &core.UserStats{Nickname: "阿明"},
			Stamina: core.StaminaState{
				Current:          12,
				Max:              50,
				NextFullBuyPrice: 200000,
			},
		},
	}
	turnstile := &mockStaminaPurchaseTurnstile{
		result: StaminaPurchaseTurnstileResult{Decision: StaminaPurchaseTurnstileAllow},
	}
	handler := NewHandler(Options{
		Store:                    store,
		Broadcaster:              &mockBroadcaster{},
		StaminaPurchaseTurnstile: turnstile,
		PlayerAuthenticator: &mockPlayerAuthenticator{
			loginToken:     "player-token",
			loginNickname:  "阿明",
			verifyNickname: "阿明",
		},
	})

	cookie := loginPlayerCookie(t, handler)
	request := httptest.NewRequest(http.MethodPost, "/api/shop/stamina/full/purchase", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Forwarded-For", "203.0.113.8")
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", response.Code, response.Body.String())
	}
	if store.lastClickNickname != "阿明" {
		t.Fatalf("expected purchase nickname 阿明, got %q", store.lastClickNickname)
	}
	if turnstile.lastRequest.RemoteIP != "203.0.113.8" {
		t.Fatalf("expected forwarded ip, got %q", turnstile.lastRequest.RemoteIP)
	}
}

func TestPurchaseStaminaFullRequiresCaptchaToken(t *testing.T) {
	store := &mockStore{}
	turnstile := &mockStaminaPurchaseTurnstile{
		result: StaminaPurchaseTurnstileResult{
			Decision: StaminaPurchaseTurnstileRequire,
			SiteKey:  "site-key",
		},
	}
	handler := NewHandler(Options{
		Store:                    store,
		Broadcaster:              &mockBroadcaster{},
		StaminaPurchaseTurnstile: turnstile,
		PlayerAuthenticator: &mockPlayerAuthenticator{
			loginToken:     "player-token",
			loginNickname:  "阿明",
			verifyNickname: "阿明",
		},
	})

	cookie := loginPlayerCookie(t, handler)
	request := httptest.NewRequest(http.MethodPost, "/api/shop/stamina/full/purchase", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", response.Code, response.Body.String())
	}
	if store.lastClickNickname != "" {
		t.Fatalf("expected purchase to be blocked, got nickname %q", store.lastClickNickname)
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

func TestPurchaseStaminaFullRejectsInvalidCaptchaToken(t *testing.T) {
	store := &mockStore{}
	turnstile := &mockStaminaPurchaseTurnstile{
		result: StaminaPurchaseTurnstileResult{Decision: StaminaPurchaseTurnstileInvalid},
	}
	handler := NewHandler(Options{
		Store:                    store,
		Broadcaster:              &mockBroadcaster{},
		StaminaPurchaseTurnstile: turnstile,
		PlayerAuthenticator: &mockPlayerAuthenticator{
			loginToken:     "player-token",
			loginNickname:  "阿明",
			verifyNickname: "阿明",
		},
	})

	cookie := loginPlayerCookie(t, handler)
	request := httptest.NewRequest(http.MethodPost, "/api/shop/stamina/full/purchase", strings.NewReader(`{"turnstileToken":"token-1"}`))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", response.Code, response.Body.String())
	}
	if store.lastClickNickname != "" {
		t.Fatalf("expected purchase to be blocked, got nickname %q", store.lastClickNickname)
	}
	if turnstile.lastRequest.Token != "token-1" {
		t.Fatalf("expected token forwarded, got %q", turnstile.lastRequest.Token)
	}
	if !strings.Contains(response.Body.String(), "CAPTCHA_INVALID") {
		t.Fatalf("expected CAPTCHA_INVALID, got %s", response.Body.String())
	}
}

func TestPurchaseStaminaFullReturnsUnavailableWhenCaptchaVerifyFails(t *testing.T) {
	store := &mockStore{}
	turnstile := &mockStaminaPurchaseTurnstile{
		result: StaminaPurchaseTurnstileResult{Decision: StaminaPurchaseTurnstileUnavailable},
	}
	handler := NewHandler(Options{
		Store:                    store,
		Broadcaster:              &mockBroadcaster{},
		StaminaPurchaseTurnstile: turnstile,
		PlayerAuthenticator: &mockPlayerAuthenticator{
			loginToken:     "player-token",
			loginNickname:  "阿明",
			verifyNickname: "阿明",
		},
	})

	cookie := loginPlayerCookie(t, handler)
	request := httptest.NewRequest(http.MethodPost, "/api/shop/stamina/full/purchase", strings.NewReader(`{"turnstileToken":"token-2"}`))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d body=%s", response.Code, response.Body.String())
	}
	if store.lastClickNickname != "" {
		t.Fatalf("expected purchase to be blocked, got nickname %q", store.lastClickNickname)
	}
	if !strings.Contains(response.Body.String(), "CAPTCHA_VERIFY_UNAVAILABLE") {
		t.Fatalf("expected CAPTCHA_VERIFY_UNAVAILABLE, got %s", response.Body.String())
	}
}

func TestPurchaseStaminaFullAllowsValidCaptchaToken(t *testing.T) {
	store := &mockStore{
		state: core.State{
			UserStats: &core.UserStats{Nickname: "阿明"},
			Stamina: core.StaminaState{
				Current:          50,
				Max:              50,
				NextFullBuyPrice: 200000,
			},
		},
	}
	turnstile := &mockStaminaPurchaseTurnstile{
		result: StaminaPurchaseTurnstileResult{Decision: StaminaPurchaseTurnstileAllow},
	}
	handler := NewHandler(Options{
		Store:                    store,
		Broadcaster:              &mockBroadcaster{},
		StaminaPurchaseTurnstile: turnstile,
		PlayerAuthenticator: &mockPlayerAuthenticator{
			loginToken:     "player-token",
			loginNickname:  "阿明",
			verifyNickname: "阿明",
		},
	})

	cookie := loginPlayerCookie(t, handler)
	request := httptest.NewRequest(http.MethodPost, "/api/shop/stamina/full/purchase", strings.NewReader(`{"turnstileToken":"token-3"}`))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", response.Code, response.Body.String())
	}
	if store.lastClickNickname != "阿明" {
		t.Fatalf("expected purchase nickname 阿明, got %q", store.lastClickNickname)
	}
	if turnstile.lastRequest.Token != "token-3" {
		t.Fatalf("expected token forwarded, got %q", turnstile.lastRequest.Token)
	}
}

func loginPlayerCookie(t *testing.T, handler *testHandler) *http.Cookie {
	t.Helper()

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/player/auth/login", strings.NewReader(`{"nickname":"阿明","password":"secret"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	handler.ServeHTTP(loginResponse, loginRequest)
	cookies := loginResponse.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected player login to set cookie")
	}
	return cookies[0]
}
