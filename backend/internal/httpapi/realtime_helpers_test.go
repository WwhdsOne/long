package httpapi

import (
	"context"
	"testing"

	"long/internal/core"
)

func TestExecuteButtonClickSkipsRateLimitForWhitelistedAuthenticatedNickname(t *testing.T) {
	store := &mockStore{}
	detector := &mockClickRiskDetector{hit: true}

	nickname, _, apiErr := executeButtonClick(context.Background(), Options{
		Store:                      store,
		ClickRiskDetector:          detector,
		AccountRisk:                &mockAccountRiskManager{},
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
	if len(detector.calls) != 0 {
		t.Fatalf("expected click detector to be skipped, got calls %v", detector.calls)
	}
	if store.lastClickNickname != "压测账号" {
		t.Fatalf("expected click to use authenticated nickname, got %q", store.lastClickNickname)
	}
}

func TestExecuteButtonClickStillEnforcesRateLimitForNonWhitelistedNickname(t *testing.T) {
	store := &mockStore{}
	detector := &mockClickRiskDetector{hit: true}
	accountRisk := &mockAccountRiskManager{}

	nickname, _, apiErr := executeButtonClick(context.Background(), Options{
		Store:                      store,
		ClickRiskDetector:          detector,
		AccountRisk:                accountRisk,
		RateLimitNicknameWhitelist: []string{"压测账号"},
	}, clickRequestContext{
		Slug:                  "feel",
		AuthenticatedNickname: "普通账号",
		AuthenticatorEnabled:  true,
		ClientID:              "127.0.0.1",
	})
	if apiErr != nil {
		t.Fatalf("expected click to continue after risk detect, got %+v", apiErr)
	}
	if nickname != "普通账号" {
		t.Fatalf("expected nickname 普通账号, got %q", nickname)
	}
	if len(detector.calls) == 0 {
		t.Fatal("expected click detector to be called for non-whitelisted nickname")
	}
	if len(accountRisk.recorded) != 1 {
		t.Fatalf("expected one risk record, got %+v", accountRisk.recorded)
	}
	if accountRisk.recorded[0].event != core.AccountRiskEventClickRateLimitHit {
		t.Fatalf("expected click rate limit event, got %+v", accountRisk.recorded[0])
	}
}

func TestExecuteButtonClickReturnsReadableRiskBanError(t *testing.T) {
	store := &mockStore{clickErr: core.ErrAccountRiskBanned}

	_, _, apiErr := executeButtonClick(context.Background(), Options{
		Store: store,
	}, clickRequestContext{
		Slug:                  "boss-part:1-0",
		AuthenticatedNickname: "普通账号",
		AuthenticatorEnabled:  true,
		ClientID:              "127.0.0.1",
	})
	if apiErr == nil {
		t.Fatal("expected risk ban error")
	}
	if apiErr.Code != "ACCOUNT_RISK_BANNED" {
		t.Fatalf("expected ACCOUNT_RISK_BANNED, got %+v", apiErr)
	}
}
