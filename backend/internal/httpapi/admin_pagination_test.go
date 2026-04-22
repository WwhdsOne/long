package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bytedance/sonic"

	"long/internal/admin"
	"long/internal/vote"
)

func TestAdminStateReturnsLightweightSummary(t *testing.T) {
	store := &mockStore{
		adminState: vote.AdminState{
			PlayerCount:       12,
			RecentPlayerCount: 4,
		},
	}

	handler := NewHandler(Options{
		Store: store,
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		}),
	})

	cookie := loginAdminForTest(t, handler)
	request := httptest.NewRequest(http.MethodGet, "/api/admin/state", nil)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload map[string]any
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if _, ok := payload["buttons"]; ok {
		t.Fatalf("expected admin summary to omit buttons, got payload=%v", payload)
	}
	if _, ok := payload["equipment"]; ok {
		t.Fatalf("expected admin summary to omit equipment, got payload=%v", payload)
	}
	if got := int64(payload["playerCount"].(float64)); got != 12 {
		t.Fatalf("expected playerCount 12, got %d", got)
	}
}

func TestAdminCatalogPagesRequireAuthAndReturnPagePayloads(t *testing.T) {
	store := &mockStore{
		adminButtonPage: vote.AdminButtonPage{
			Items: []vote.Button{
				{Key: "feel", Label: "有感觉吗", Sort: 10, Enabled: true},
			},
			Page:       2,
			PageSize:   1,
			Total:      3,
			TotalPages: 3,
		},
		adminEquipmentPage: vote.AdminEquipmentPage{
			Items: []vote.EquipmentDefinition{
				{ItemID: "wood-sword", Name: "木剑", Slot: "weapon", BonusClicks: 2},
			},
			Page:       2,
			PageSize:   1,
			Total:      3,
			TotalPages: 3,
		},
		adminBossHistoryPage: vote.AdminBossHistoryPage{
			Items: []vote.BossHistoryEntry{
				{Boss: vote.Boss{ID: "boss-2", Name: "史莱姆王", Status: "defeated", StartedAt: 2}},
			},
			Page:       2,
			PageSize:   1,
			Total:      4,
			TotalPages: 4,
		},
	}

	handler := NewHandler(Options{
		Store: store,
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		}),
	})

	unauthorizedRequest := httptest.NewRequest(http.MethodGet, "/api/admin/buttons?page=2&pageSize=1", nil)
	unauthorizedResponse := httptest.NewRecorder()
	handler.ServeHTTP(unauthorizedResponse, unauthorizedRequest)
	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without session, got %d", unauthorizedResponse.Code)
	}

	cookie := loginAdminForTest(t, handler)

	buttonRequest := httptest.NewRequest(http.MethodGet, "/api/admin/buttons?page=2&pageSize=1", nil)
	buttonRequest.AddCookie(cookie)
	buttonResponse := httptest.NewRecorder()
	handler.ServeHTTP(buttonResponse, buttonRequest)
	if buttonResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from admin buttons page, got %d", buttonResponse.Code)
	}

	var buttonPayload vote.AdminButtonPage
	if err := sonic.Unmarshal(buttonResponse.Body.Bytes(), &buttonPayload); err != nil {
		t.Fatalf("decode button page: %v", err)
	}
	if buttonPayload.Page != 2 || buttonPayload.PageSize != 1 || buttonPayload.Total != 3 || len(buttonPayload.Items) != 1 {
		t.Fatalf("unexpected button page payload: %+v", buttonPayload)
	}

	equipmentRequest := httptest.NewRequest(http.MethodGet, "/api/admin/equipment?page=2&pageSize=1", nil)
	equipmentRequest.AddCookie(cookie)
	equipmentResponse := httptest.NewRecorder()
	handler.ServeHTTP(equipmentResponse, equipmentRequest)
	if equipmentResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from admin equipment page, got %d", equipmentResponse.Code)
	}

	var equipmentPayload vote.AdminEquipmentPage
	if err := sonic.Unmarshal(equipmentResponse.Body.Bytes(), &equipmentPayload); err != nil {
		t.Fatalf("decode equipment page: %v", err)
	}
	if equipmentPayload.Page != 2 || equipmentPayload.PageSize != 1 || equipmentPayload.Total != 3 || len(equipmentPayload.Items) != 1 {
		t.Fatalf("unexpected equipment page payload: %+v", equipmentPayload)
	}

	historyRequest := httptest.NewRequest(http.MethodGet, "/api/admin/boss/history?page=2&pageSize=1", nil)
	historyRequest.AddCookie(cookie)
	historyResponse := httptest.NewRecorder()
	handler.ServeHTTP(historyResponse, historyRequest)
	if historyResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from admin boss history page, got %d", historyResponse.Code)
	}

	var historyPayload vote.AdminBossHistoryPage
	if err := sonic.Unmarshal(historyResponse.Body.Bytes(), &historyPayload); err != nil {
		t.Fatalf("decode boss history page: %v", err)
	}
	if historyPayload.Page != 2 || historyPayload.PageSize != 1 || historyPayload.Total != 4 || len(historyPayload.Items) != 1 {
		t.Fatalf("unexpected boss history page payload: %+v", historyPayload)
	}
}

func loginAdminForTest(t *testing.T, handler http.Handler) *http.Cookie {
	t.Helper()

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{"username":"admin","password":"secret"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	handler.ServeHTTP(loginResponse, loginRequest)

	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from login, got %d", loginResponse.Code)
	}

	cookies := loginResponse.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected login to set session cookie")
	}
	return cookies[0]
}
