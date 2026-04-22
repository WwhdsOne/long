package httpapi

import (
	"testing"

	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"long/internal/admin"
)

func TestAdminEquipmentSaveUsesNativeHertzFlowAndPersistsRarity(t *testing.T) {
	store := &mockStore{}
	handler := newNativeTestHandler(Options{
		Store: store,
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		}),
	})

	loginResponse := performJSONRequest(t, handler, consts.MethodPost, "/api/admin/login", `{"username":"admin","password":"secret"}`)
	if loginResponse.StatusCode() != consts.StatusOK {
		t.Fatalf("expected login 200, got %d", loginResponse.StatusCode())
	}
	sessionCookie := string(loginResponse.Header.Peek("Set-Cookie"))
	if sessionCookie == "" {
		t.Fatal("expected session cookie from admin login")
	}

	saveResponse := performJSONRequest(
		t,
		handler,
		consts.MethodPost,
		"/api/admin/equipment",
		`{"itemId":"fire-ring","name":"🔥 炽焰戒","slot":"accessory","rarity":"至臻","bonusClicks":2,"bonusCriticalChancePercent":6,"bonusCriticalCount":4,"enhanceCap":7}`,
		ut.Header{Key: "Cookie", Value: sessionCookie},
	)
	if saveResponse.StatusCode() != consts.StatusOK {
		t.Fatalf("expected equipment save 200, got %d", saveResponse.StatusCode())
	}
	if store.lastEquipment.Rarity != "至臻" {
		t.Fatalf("expected saved rarity 至臻, got %q", store.lastEquipment.Rarity)
	}
	if store.lastEquipment.ItemID != "fire-ring" {
		t.Fatalf("expected saved item fire-ring, got %q", store.lastEquipment.ItemID)
	}
}
