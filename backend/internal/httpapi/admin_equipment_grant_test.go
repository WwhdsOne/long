package httpapi

import (
	"testing"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"long/internal/admin"
	"long/internal/core"
)

func TestAdminGrantEquipmentToPlayerPublishesEquipmentChange(t *testing.T) {
	store := &mockStore{
		grantEquipmentState: core.UserState{
			Inventory: []core.InventoryItem{
				{ItemID: "wood-sword", InstanceID: "inst-1", Name: "木剑"},
				{ItemID: "wood-sword", InstanceID: "inst-2", Name: "木剑"},
			},
		},
	}
	publisher := &mockChangePublisher{}
	handler := newNativeTestHandler(Options{
		Store:           store,
		ChangePublisher: publisher,
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

	response := performJSONRequest(
		t,
		handler,
		consts.MethodPost,
		"/api/admin/players/阿明/equipment",
		`{"itemId":"wood-sword","quantity":2}`,
		ut.Header{Key: "Cookie", Value: sessionCookie},
	)
	if response.StatusCode() != consts.StatusOK {
		t.Fatalf("expected grant 200, got %d body=%s", response.StatusCode(), response.Body())
	}
	if store.lastGrantNickname != "阿明" || store.lastGrantItemID != "wood-sword" || store.lastGrantQuantity != 2 {
		t.Fatalf("unexpected grant args: nickname=%q item=%q quantity=%d", store.lastGrantNickname, store.lastGrantItemID, store.lastGrantQuantity)
	}
	if len(publisher.changes) != 1 {
		t.Fatalf("expected one published change, got %+v", publisher.changes)
	}
	if publisher.changes[0].Type != core.StateChangeEquipmentChanged || publisher.changes[0].Nickname != "阿明" {
		t.Fatalf("unexpected change payload: %+v", publisher.changes[0])
	}

	var payload struct {
		OK    bool           `json:"ok"`
		State core.UserState `json:"state"`
	}
	if err := sonic.Unmarshal(response.Body(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !payload.OK || len(payload.State.Inventory) != 2 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestAdminGrantEquipmentToPlayerRejectsInvalidQuantity(t *testing.T) {
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

	response := performJSONRequest(
		t,
		handler,
		consts.MethodPost,
		"/api/admin/players/阿明/equipment",
		`{"itemId":"wood-sword","quantity":0}`,
		ut.Header{Key: "Cookie", Value: sessionCookie},
	)
	if response.StatusCode() != consts.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.StatusCode())
	}
	if store.lastGrantNickname != "" || store.lastGrantItemID != "" || store.lastGrantQuantity != 0 {
		t.Fatalf("expected grant not called, got nickname=%q item=%q quantity=%d", store.lastGrantNickname, store.lastGrantItemID, store.lastGrantQuantity)
	}
}
