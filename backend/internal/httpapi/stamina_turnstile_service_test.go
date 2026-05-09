package httpapi

import (
	"context"
	"testing"
)

func TestPlayerLoginTurnstileAllowsLoopbackRequestWithoutToken(t *testing.T) {
	service := NewPlayerLoginTurnstile(PlayerLoginTurnstileConfig{
		Enabled:   true,
		SiteKey:   "site-key",
		SecretKey: "secret-key",
	})

	result, err := service.CheckPlayerLogin(context.Background(), PlayerLoginTurnstileRequest{
		Nickname: "阿明",
		RemoteIP: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("check player login: %v", err)
	}
	if result.Decision != PlayerLoginTurnstileAllow {
		t.Fatalf("expected loopback login request to bypass turnstile, got %+v", result)
	}
}

func TestPurchaseStaminaTurnstileAllowsLoopbackRequestWithoutToken(t *testing.T) {
	service := NewStaminaPurchaseTurnstile(StaminaPurchaseTurnstileConfig{
		Enabled:                   true,
		SiteKey:                   "site-key",
		SecretKey:                 "secret-key",
		PurchaseStaminaSampleRate: 1,
	})

	result, err := service.CheckPurchaseStamina(context.Background(), StaminaPurchaseTurnstileRequest{
		Nickname: "阿明",
		RemoteIP: "::1",
	})
	if err != nil {
		t.Fatalf("check purchase stamina: %v", err)
	}
	if result.Decision != StaminaPurchaseTurnstileAllow {
		t.Fatalf("expected loopback stamina request to bypass turnstile, got %+v", result)
	}
}
