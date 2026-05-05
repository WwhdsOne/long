package httpapi

import (
	"context"
	"testing"

	"long/internal/ratelimit"
)

func TestExecuteButtonClickSkipsRateLimitForWhitelistedAuthenticatedNickname(t *testing.T) {
	store := &mockStore{}
	guard := &mockClickGuard{err: ratelimit.ErrTooManyRequests}

	nickname, _, apiErr := executeButtonClick(context.Background(), Options{
		Store:                      store,
		ClickGuard:                 guard,
		RateLimitNicknameWhitelist: []string{"压测账号"},
	}, clickRequestContext{
		Slug:                  "feel",
		AuthenticatedNickname: "压测账号",
		AuthenticatorEnabled:  true,
		ClientID:              "127.0.0.1",
	})
	if apiErr != nil {
		t.Fatalf("expected whitelisted nickname to skip rate limit, got %+v", apiErr)
	}
	if nickname != "压测账号" {
		t.Fatalf("expected nickname 压测账号, got %q", nickname)
	}
	if len(guard.calls) != 0 {
		t.Fatalf("expected click guard to be skipped, got calls %v", guard.calls)
	}
	if store.lastClickNickname != "压测账号" {
		t.Fatalf("expected click to use authenticated nickname, got %q", store.lastClickNickname)
	}
}

func TestExecuteButtonClickStillEnforcesRateLimitForNonWhitelistedNickname(t *testing.T) {
	store := &mockStore{}
	guard := &mockClickGuard{err: ratelimit.ErrTooManyRequests}

	_, _, apiErr := executeButtonClick(context.Background(), Options{
		Store:                      store,
		ClickGuard:                 guard,
		RateLimitNicknameWhitelist: []string{"压测账号"},
	}, clickRequestContext{
		Slug:                  "feel",
		AuthenticatedNickname: "普通账号",
		AuthenticatorEnabled:  true,
		ClientID:              "127.0.0.1",
	})
	if apiErr == nil {
		t.Fatal("expected non-whitelisted nickname to still hit rate limit")
	}
	if apiErr.Code != "TOO_MANY_REQUESTS" {
		t.Fatalf("expected TOO_MANY_REQUESTS, got %+v", apiErr)
	}
	if len(guard.calls) == 0 {
		t.Fatal("expected click guard to be called for non-whitelisted nickname")
	}
}
