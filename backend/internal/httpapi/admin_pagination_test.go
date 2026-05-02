package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bytedance/sonic"

	"long/internal/admin"
	"long/internal/core"
)

type mockBossHistoryReader struct {
	page core.AdminBossHistoryPage
	err  error
}

func (m *mockBossHistoryReader) ListAdminBossHistoryPage(_ context.Context, _ int64, _ int64) (core.AdminBossHistoryPage, error) {
	return m.page, m.err
}

type mockMessageStore struct {
	page          core.MessagePage
	created       *core.Message
	deletedID     string
	listCursor    string
	listLimit     int64
	createNick    string
	createContent string
	err           error
}

func (m *mockMessageStore) CreateMessage(_ context.Context, nickname string, content string) (*core.Message, error) {
	m.createNick = nickname
	m.createContent = content
	if m.err != nil {
		return nil, m.err
	}
	if m.created != nil {
		return m.created, nil
	}
	return &core.Message{ID: "1", Nickname: nickname, Content: content, CreatedAt: 1}, nil
}

func (m *mockMessageStore) ListMessages(_ context.Context, cursor string, limit int64) (core.MessagePage, error) {
	m.listCursor = cursor
	m.listLimit = limit
	return m.page, m.err
}

func (m *mockMessageStore) DeleteMessage(_ context.Context, id string) error {
	m.deletedID = id
	return m.err
}

func TestAdminStateReturnsLightweightSummary(t *testing.T) {
	store := &mockStore{
		adminState: core.AdminState{
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

func TestAdminBossHistoryPagePrefersOptionalReader(t *testing.T) {
	store := &mockStore{
		adminBossHistoryPage: core.AdminBossHistoryPage{
			Items: []core.BossHistoryEntry{
				{Boss: core.Boss{ID: "redis-boss", Name: "Redis Boss", Status: "defeated", StartedAt: 1}},
			},
			Page:       1,
			PageSize:   20,
			Total:      1,
			TotalPages: 1,
		},
	}
	reader := &mockBossHistoryReader{
		page: core.AdminBossHistoryPage{
			Items: []core.BossHistoryEntry{
				{Boss: core.Boss{ID: "mongo-boss", Name: "Mongo Boss", Status: "defeated", StartedAt: 2}},
			},
			Page:       1,
			PageSize:   20,
			Total:      1,
			TotalPages: 1,
		},
	}

	handler := NewHandler(Options{
		Store:                  store,
		AdminBossHistoryReader: reader,
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		}),
	})

	cookie := loginAdminForTest(t, handler)
	request := httptest.NewRequest(http.MethodGet, "/api/admin/boss/history?page=1&pageSize=20", nil)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200 from admin boss history page, got %d", response.Code)
	}

	var payload core.AdminBossHistoryPage
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode boss history page: %v", err)
	}
	if len(payload.Items) != 1 || payload.Items[0].ID != "mongo-boss" {
		t.Fatalf("expected optional reader payload, got %+v", payload)
	}
}

func TestAdminMessagesPreferOptionalMessageStore(t *testing.T) {
	store := &mockStore{
		messagePage: core.MessagePage{
			Items: []core.Message{{ID: "redis-1", Nickname: "阿明", Content: "redis"}},
		},
	}
	messageStore := &mockMessageStore{
		page: core.MessagePage{
			Items: []core.Message{{ID: "mongo-1", Nickname: "小红", Content: "mongo"}},
		},
	}

	handler := NewHandler(Options{
		Store:        store,
		MessageStore: messageStore,
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		}),
	})

	cookie := loginAdminForTest(t, handler)
	request := httptest.NewRequest(http.MethodGet, "/api/admin/messages?cursor=10", nil)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200 from admin messages, got %d", response.Code)
	}

	var payload core.MessagePage
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode message page: %v", err)
	}
	if len(payload.Items) != 1 || payload.Items[0].ID != "mongo-1" {
		t.Fatalf("expected optional message store payload, got %+v", payload)
	}
	if messageStore.listCursor != "10" || messageStore.listLimit != 50 {
		t.Fatalf("unexpected message store list params: cursor=%q limit=%d", messageStore.listCursor, messageStore.listLimit)
	}
}

func TestAdminCatalogPagesRequireAuthAndReturnPagePayloads(t *testing.T) {
	store := &mockStore{
		adminEquipmentPage: core.AdminEquipmentPage{
			Items: []core.EquipmentDefinition{
				{ItemID: "wood-sword", Name: "木剑", Slot: "weapon"},
			},
			Page:       2,
			PageSize:   1,
			Total:      3,
			TotalPages: 3,
		},
		adminBossHistoryPage: core.AdminBossHistoryPage{
			Items: []core.BossHistoryEntry{
				{Boss: core.Boss{ID: "boss-2", Name: "史莱姆王", Status: "defeated", StartedAt: 2}},
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

	cookie := loginAdminForTest(t, handler)

	equipmentRequest := httptest.NewRequest(http.MethodGet, "/api/admin/equipment?page=2&pageSize=1", nil)
	equipmentRequest.AddCookie(cookie)
	equipmentResponse := httptest.NewRecorder()
	handler.ServeHTTP(equipmentResponse, equipmentRequest)
	if equipmentResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from admin equipment page, got %d", equipmentResponse.Code)
	}

	var equipmentPayload core.AdminEquipmentPage
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

	var historyPayload core.AdminBossHistoryPage
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
