package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bytedance/sonic"

	"long/internal/admin"
	ossupload "long/internal/oss"
	"long/internal/vote"
)

type mockStore struct {
	state                 vote.State
	snapshot              vote.Snapshot
	equipState            vote.State
	adminState            vote.AdminState
	bossResources         vote.BossResources
	adminEquipmentPage    vote.AdminEquipmentPage
	adminBossHistoryPage  vote.AdminBossHistoryPage
	adminPlayerPage       vote.AdminPlayerPage
	adminPlayer           *vote.AdminPlayerOverview
	bossHistory           []vote.BossHistoryEntry
	announcements         []vote.Announcement
	latestAnnouncement    *vote.Announcement
	messagePage           vote.MessagePage
	result                vote.ClickResult
	lastBoss              vote.BossUpsert
	lastBossTemplate      vote.BossTemplateUpsert
	lastEquipment         vote.EquipmentDefinition
	lastTemplateLootID    string
	lastTemplateLoot      []vote.BossLootEntry
	lastCycleQueue        []string
	lastCycleEnabled      bool
	lastSalvageItemID     string
	lastSalvageQuantity   int64
	lastLockItemID        string
	lastLockState         bool
	lastClickNickname     string
	lastAutoClickNickname string
	lastGetStateNickname  string
	getStateErr           error
	clickErr              error
	equipErr              error
	enhanceErr            error
	validateErr           error
	messageErr            error
	salvageErr            error
	activateBossErr       error
	saveBossTemplateErr   error
	setBossCycleErr       error
}

func (m *mockStore) GetState(_ context.Context, nickname string) (vote.State, error) {
	m.lastGetStateNickname = nickname
	if m.getStateErr != nil {
		return vote.State{}, m.getStateErr
	}
	if len(m.snapshot.Leaderboard) > 0 || m.snapshot.Boss != nil || m.snapshot.AnnouncementVersion != "" {
		return vote.ComposeState(m.snapshot, m.userStateForNickname(nickname)), nil
	}
	state := m.state
	if nickname == "" {
		state.UserStats = nil
	}
	return state, nil
}

func (m *mockStore) GetSnapshot(_ context.Context) (vote.Snapshot, error) {
	if len(m.snapshot.Leaderboard) > 0 || m.snapshot.Boss != nil || m.snapshot.AnnouncementVersion != "" {
		return m.snapshot, nil
	}
	return vote.Snapshot{
		Leaderboard: m.state.Leaderboard,
	}, nil
}

func (m *mockStore) GetBossResources(_ context.Context) (vote.BossResources, error) {
	return m.bossResources, nil
}

func (m *mockStore) GetUserState(_ context.Context, nickname string) (vote.UserState, error) {
	if m.getStateErr != nil {
		return vote.UserState{}, m.getStateErr
	}
	return m.userStateForNickname(nickname), nil
}

func (m *mockStore) userStateForNickname(nickname string) vote.UserState {
	userState := vote.UserState{
		Inventory:   []vote.InventoryItem{},
		Loadout:     vote.Loadout{},
		CombatStats: vote.CombatStats{},
	}
	if nickname == "" {
		return userState
	}

	userState.UserStats = m.state.UserStats
	userState.MyBossStats = m.state.MyBossStats
	userState.Inventory = m.state.Inventory
	userState.Loadout = m.state.Loadout
	userState.CombatStats = m.state.CombatStats
	userState.Gold = m.state.Gold
	userState.Stones = m.state.Stones
	userState.TalentPoints = m.state.TalentPoints
	userState.RecentRewards = m.state.RecentRewards
	return userState
}

func (m *mockStore) ClickButton(_ context.Context, slug string, nickname string, comboCount int64) (vote.ClickResult, error) {
	m.lastClickNickname = nickname
	if m.clickErr != nil {
		return vote.ClickResult{}, m.clickErr
	}
	if m.result.Delta == 0 && m.result.UserStats.Nickname == "" {
		m.result.Delta = 1
		m.result.UserStats = vote.UserStats{Nickname: nickname, ClickCount: 1}
	}
	return m.result, nil
}

func (m *mockStore) AutoClickBossPart(_ context.Context, slug string, nickname string) (vote.ClickResult, error) {
	m.lastAutoClickNickname = nickname
	return m.ClickButton(context.Background(), slug, nickname, 0)
}

func (m *mockStore) ClickBossPart(_ context.Context, slug string, nickname string) (vote.ClickResult, error) {
	return m.ClickButton(context.Background(), slug, nickname, 0)
}

func (m *mockStore) AttackBossPartAFK(_ context.Context, nickname string) (vote.ClickResult, error) {
	m.lastAutoClickNickname = nickname
	return vote.ClickResult{
		Boss: &vote.Boss{
			ID:        "boss-1",
			Name:      "测试 Boss",
			Status:    "active",
			MaxHP:     100,
			CurrentHP: 90,
		},
	}, nil
}

func (m *mockStore) ValidateNickname(_ context.Context, _ string) error {
	return m.validateErr
}

func (m *mockStore) EquipItem(_ context.Context, nickname string, _ string) (vote.State, error) {
	m.lastClickNickname = nickname
	if m.equipErr != nil {
		return vote.State{}, m.equipErr
	}
	if m.equipState.Loadout.Weapon == nil {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) UnequipItem(_ context.Context, _ string, _ string) (vote.State, error) {
	if m.equipErr != nil {
		return vote.State{}, m.equipErr
	}
	if m.equipState.Loadout.Weapon == nil {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) EnhanceItem(_ context.Context, _ string, _ string) (vote.State, error) {
	if m.enhanceErr != nil {
		return vote.State{}, m.enhanceErr
	}
	if m.equipState.Loadout.Weapon == nil {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) SalvageItem(_ context.Context, _ string, itemID string) (vote.SalvageResult, error) {
	m.lastSalvageItemID = itemID
	if m.salvageErr != nil {
		return vote.SalvageResult{}, m.salvageErr
	}
	return vote.SalvageResult{
		ItemID:         itemID,
		GoldReward:     500,
		StoneReward:    1,
		RefundedStones: 12,
		Gold:           66,
		Stones:         34,
	}, nil
}

func (m *mockStore) BulkSalvageUnequipped(_ context.Context, _ string) (vote.BulkSalvageResult, error) {
	if m.salvageErr != nil {
		return vote.BulkSalvageResult{}, m.salvageErr
	}
	return vote.BulkSalvageResult{
		SalvagedCount:       3,
		SalvagedByRarity:    map[string]int{"普通": 1, "稀有": 2},
		GoldReward:          1200,
		StoneReward:         2,
		RefundedStones:      4,
		Gold:                2000,
		Stones:              66,
		HasEnhancedSalvaged: true,
	}, nil
}

func (m *mockStore) LockItem(_ context.Context, _ string, itemID string) (vote.State, error) {
	m.lastLockItemID = itemID
	m.lastLockState = true
	if m.equipErr != nil {
		return vote.State{}, m.equipErr
	}
	if m.equipState.Loadout.Weapon == nil {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) UnlockItem(_ context.Context, _ string, itemID string) (vote.State, error) {
	m.lastLockItemID = itemID
	m.lastLockState = false
	if m.equipErr != nil {
		return vote.State{}, m.equipErr
	}
	if m.equipState.Loadout.Weapon == nil {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) GetAdminState(_ context.Context) (vote.AdminState, error) {
	return m.adminState, nil
}

func (m *mockStore) ListAdminEquipmentPage(_ context.Context, _ int64, _ int64) (vote.AdminEquipmentPage, error) {
	return m.adminEquipmentPage, nil
}

func (m *mockStore) ListAdminBossHistoryPage(_ context.Context, _ int64, _ int64) (vote.AdminBossHistoryPage, error) {
	return m.adminBossHistoryPage, nil
}

func (m *mockStore) ListAdminPlayers(_ context.Context, _ string, _ int64) (vote.AdminPlayerPage, error) {
	return m.adminPlayerPage, nil
}

func (m *mockStore) GetAdminPlayer(_ context.Context, _ string) (*vote.AdminPlayerOverview, error) {
	return m.adminPlayer, nil
}

func (m *mockStore) SaveEquipmentDefinition(_ context.Context, definition vote.EquipmentDefinition) error {
	m.lastEquipment = definition
	return nil
}

func (m *mockStore) DeleteEquipmentDefinition(_ context.Context, _ string) error {
	return nil
}

func (m *mockStore) ActivateBoss(_ context.Context, boss vote.BossUpsert) (*vote.Boss, error) {
	if m.activateBossErr != nil {
		return nil, m.activateBossErr
	}
	m.lastBoss = boss
	return &vote.Boss{
		ID:        boss.ID,
		Name:      boss.Name,
		Status:    "active",
		MaxHP:     boss.MaxHP,
		CurrentHP: boss.MaxHP,
	}, nil
}

func (m *mockStore) DeactivateBoss(_ context.Context) error {
	return nil
}

func (m *mockStore) SetBossLoot(_ context.Context, _ string, _ []vote.BossLootEntry) error {
	return nil
}

func (m *mockStore) SaveBossTemplate(_ context.Context, template vote.BossTemplateUpsert) error {
	if m.saveBossTemplateErr != nil {
		return m.saveBossTemplateErr
	}
	m.lastBossTemplate = template
	return nil
}

func (m *mockStore) DeleteBossTemplate(_ context.Context, _ string) error {
	return nil
}

func (m *mockStore) SetBossTemplateLoot(_ context.Context, templateID string, loot []vote.BossLootEntry) error {
	m.lastTemplateLootID = templateID
	m.lastTemplateLoot = loot
	return nil
}

func (m *mockStore) SetBossCycleQueue(_ context.Context, templateIDs []string) ([]string, error) {
	m.lastCycleQueue = append([]string(nil), templateIDs...)
	return append([]string(nil), templateIDs...), nil
}

func (m *mockStore) SetBossCycleEnabled(_ context.Context, enabled bool) (*vote.Boss, error) {
	if m.setBossCycleErr != nil {
		return nil, m.setBossCycleErr
	}
	m.lastCycleEnabled = enabled
	if !enabled {
		return nil, nil
	}
	return &vote.Boss{
		ID:         "dragon-1",
		TemplateID: "dragon",
		Name:       "火龙",
		Status:     "active",
		MaxHP:      80,
		CurrentHP:  80,
	}, nil
}

func (m *mockStore) ListBossHistory(_ context.Context) ([]vote.BossHistoryEntry, error) {
	return m.bossHistory, nil
}

func (m *mockStore) GetLatestAnnouncement(_ context.Context) (*vote.Announcement, error) {
	return m.latestAnnouncement, nil
}

func (m *mockStore) ListAnnouncements(_ context.Context, includeInactive bool) ([]vote.Announcement, error) {
	return m.announcements, nil
}

func (m *mockStore) SaveAnnouncement(_ context.Context, announcement vote.AnnouncementUpsert) (*vote.Announcement, error) {
	return &vote.Announcement{
		ID:          "1",
		Title:       announcement.Title,
		Content:     announcement.Content,
		PublishedAt: 1710000000,
		Active:      announcement.Active,
	}, nil
}

func (m *mockStore) DeleteAnnouncement(_ context.Context, _ string) error {
	return nil
}

func (m *mockStore) CreateMessage(_ context.Context, nickname string, content string) (*vote.Message, error) {
	if m.messageErr != nil {
		return nil, m.messageErr
	}
	return &vote.Message{
		ID:        "1",
		Nickname:  nickname,
		Content:   content,
		CreatedAt: 1710000000,
	}, nil
}

func (m *mockStore) ListMessages(_ context.Context, _ string, _ int64) (vote.MessagePage, error) {
	return m.messagePage, m.messageErr
}

func (m *mockStore) DeleteMessage(_ context.Context, _ string) error {
	return nil
}


func (m *mockStore) GetTalentState(_ context.Context, _ string) (*vote.TalentState, error) {
	return nil, nil
}

func (m *mockStore) LearnTalent(_ context.Context, _ string, _ string) error {
	return nil
}

func (m *mockStore) ResetTalents(_ context.Context, _ string) error {
	return nil
}

func (m *mockStore) ComputeTalentModifiers(_ context.Context, _ string) (*vote.TalentModifiers, error) {
	return nil, nil
}

type mockOSSSigner struct {
	policy ossupload.Policy
	err    error
}

func (m *mockOSSSigner) CreatePolicy(_ context.Context) (ossupload.Policy, error) {
	return m.policy, m.err
}

type mockAutoClickController struct {
	status             AutoClickStatus
	startErr           error
	lastStartNickname  string
	lastStartSlug      string
	lastStopNickname   string
	lastStatusNickname string
}

func (m *mockAutoClickController) Start(_ context.Context, nickname string, slug string) (AutoClickStatus, error) {
	m.lastStartNickname = nickname
	m.lastStartSlug = slug
	if m.startErr != nil {
		return AutoClickStatus{}, m.startErr
	}
	status := m.status
	status.Active = true
	status.ButtonKey = slug
	return status, nil
}

func (m *mockAutoClickController) Stop(nickname string) AutoClickStatus {
	m.lastStopNickname = nickname
	status := m.status
	status.Active = false
	status.ButtonKey = ""
	return status
}

func (m *mockAutoClickController) Status(nickname string) AutoClickStatus {
	m.lastStatusNickname = nickname
	return m.status
}

func (m *mockAutoClickController) Close() error {
	return nil
}

type mockBroadcaster struct {
	snapshots []vote.Snapshot
}

func (m *mockBroadcaster) BroadcastSnapshot(snapshot vote.Snapshot) error {
	m.snapshots = append(m.snapshots, snapshot)
	return nil
}

type mockChangePublisher struct {
	changes []vote.StateChange
}

func (m *mockChangePublisher) PublishChange(_ context.Context, change vote.StateChange) error {
	m.changes = append(m.changes, change)
	return nil
}
func TestGetBossHistoryReturnsPublicHistory(t *testing.T) {
	store := &mockStore{
		bossHistory: []vote.BossHistoryEntry{
			{
				Boss: vote.Boss{
					ID:         "slime-king",
					Name:       "史莱姆王",
					Status:     "defeated",
					MaxHP:      100,
					CurrentHP:  0,
					StartedAt:  1710000000,
					DefeatedAt: 1710000300,
				},
				Loot: []vote.BossLootEntry{
					{ItemID: "cloth-armor", ItemName: "布甲", Weight: 3},
				},
				Damage: []vote.BossLeaderboardEntry{
					{Rank: 1, Nickname: "阿明", Damage: 42},
				},
			},
		},
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodGet, "/api/boss/history", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload []vote.BossHistoryEntry
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(payload) != 1 || payload[0].Name != "史莱姆王" {
		t.Fatalf("unexpected history payload: %+v", payload)
	}
	if len(payload[0].Damage) != 1 || payload[0].Damage[0].Nickname != "阿明" {
		t.Fatalf("unexpected history damage payload: %+v", payload[0].Damage)
	}
}

func TestGetLatestAnnouncementReturnsPayload(t *testing.T) {
	store := &mockStore{
		latestAnnouncement: &vote.Announcement{
			ID:          "7",
			Title:       "更新公告",
			Content:     "留言墙已上线。",
			PublishedAt: 1710000000,
			Active:      true,
		},
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodGet, "/api/announcements/latest", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload vote.Announcement
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.ID != "7" || payload.Title != "更新公告" {
		t.Fatalf("unexpected latest announcement payload: %+v", payload)
	}
}
func TestEquipItemReturnsUpdatedState(t *testing.T) {
	store := &mockStore{
		equipState: vote.State{

			Loadout: vote.Loadout{
				Weapon: &vote.InventoryItem{
					ItemID:   "wood-sword",
					Name:     "木剑",
					Slot:     "weapon",
					Quantity: 1,
					Equipped: true,
				},
			},
			CombatStats: vote.CombatStats{
				EffectiveIncrement: 3,
				NormalDamage:       3,
				CriticalDamage:     7,
			},
		},
	}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/equipment/instance-wood-sword/equip", strings.NewReader(`{"nickname":"阿明"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload struct {
		Loadout struct {
			Weapon *vote.InventoryItem `json:"weapon"`
		} `json:"loadout"`
		CombatStats vote.CombatStats `json:"combatStats"`
	}
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Loadout.Weapon == nil || payload.Loadout.Weapon.ItemID != "wood-sword" {
		t.Fatalf("expected equipped wood-sword, got %+v", payload.Loadout.Weapon)
	}
	if payload.CombatStats.EffectiveIncrement != 3 {
		t.Fatalf("expected effective increment 3, got %+v", payload.CombatStats)
	}
	if payload.CombatStats.NormalDamage != 3 || payload.CombatStats.CriticalDamage != 7 {
		t.Fatalf("expected actual damage 3/7, got %+v", payload.CombatStats)
	}
}

func TestLockItemForwardsInstanceID(t *testing.T) {
	store := &mockStore{}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/equipment/instance-wood-sword/lock", strings.NewReader(`{"nickname":"阿明"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if store.lastLockItemID != "instance-wood-sword" || !store.lastLockState {
		t.Fatalf("expected lock item forwarded, got item=%q locked=%v", store.lastLockItemID, store.lastLockState)
	}
}

func TestUnlockItemForwardsInstanceID(t *testing.T) {
	store := &mockStore{}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/equipment/instance-wood-sword/unlock", strings.NewReader(`{"nickname":"阿明"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if store.lastLockItemID != "instance-wood-sword" || store.lastLockState {
		t.Fatalf("expected unlock item forwarded, got item=%q locked=%v", store.lastLockItemID, store.lastLockState)
	}
}

func TestBulkSalvageUnequippedReturnsSummary(t *testing.T) {
	store := &mockStore{}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/equipment/salvage/unequipped", strings.NewReader(`{"nickname":"阿明"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload vote.BulkSalvageResult
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.SalvagedCount != 3 || payload.GoldReward != 1200 || payload.Stones != 66 {
		t.Fatalf("unexpected bulk salvage payload: %+v", payload)
	}
}

func TestSynthesizeItemReturnsDeprecatedError(t *testing.T) {
	handler := NewHandler(Options{
		Store:       &mockStore{},
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/equipment/instance-wood-sword/synthesize", strings.NewReader(`{"nickname":"阿明"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusGone {
		t.Fatalf("expected 410, got %d", response.Code)
	}

	var payload map[string]string
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["error"] == "" {
		t.Fatalf("expected deprecated error payload, got %+v", payload)
	}
}

func TestPostMessageRejectsSensitiveContent(t *testing.T) {
	store := &mockStore{
		messageErr: vote.ErrSensitiveContent,
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/messages", strings.NewReader(`{"nickname":"阿明","content":"XJP后援会"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
	if body := response.Body.String(); !strings.Contains(body, "敏感词") {
		t.Fatalf("expected sensitive message content error, got %q", body)
	}
}

func TestAdminLoginCreatesSessionAndStateRequiresAuth(t *testing.T) {
	store := &mockStore{
		adminState: vote.AdminState{
			PlayerCount:       8,
			RecentPlayerCount: 3,
		},
	}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		}),
	})

	unauthorizedRequest := httptest.NewRequest(http.MethodGet, "/api/admin/state", nil)
	unauthorizedResponse := httptest.NewRecorder()
	handler.ServeHTTP(unauthorizedResponse, unauthorizedRequest)

	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without session, got %d", unauthorizedResponse.Code)
	}

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

	adminRequest := httptest.NewRequest(http.MethodGet, "/api/admin/state", nil)
	adminRequest.AddCookie(cookies[0])
	adminResponse := httptest.NewRecorder()
	handler.ServeHTTP(adminResponse, adminRequest)

	if adminResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 with session, got %d", adminResponse.Code)
	}

	var payload map[string]any
	if err := sonic.Unmarshal(adminResponse.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if _, ok := payload["buttons"]; ok {
		t.Fatalf("expected admin state summary to omit buttons, got %+v", payload)
	}
	if _, ok := payload["equipment"]; ok {
		t.Fatalf("expected admin state summary to omit equipment, got %+v", payload)
	}
	if got := int64(payload["playerCount"].(float64)); got != 8 {
		t.Fatalf("expected playerCount 8, got %d", got)
	}
}

func TestAdminBossPoolRoutesForwardTemplateAndCyclePayloads(t *testing.T) {
	store := &mockStore{}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		}),
	})

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{"username":"admin","password":"secret"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	handler.ServeHTTP(loginResponse, loginRequest)

	cookies := loginResponse.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie from login")
	}

	saveTemplateRequest := httptest.NewRequest(http.MethodPost, "/api/admin/boss/pool", strings.NewReader(`{"id":"dragon","name":"火龙","maxHp":80,"layout":[{"x":0,"y":0,"type":"soft","maxHp":80}]}`))
	saveTemplateRequest.Header.Set("Content-Type", "application/json")
	saveTemplateRequest.AddCookie(cookies[0])
	saveTemplateResponse := httptest.NewRecorder()
	handler.ServeHTTP(saveTemplateResponse, saveTemplateRequest)

	if saveTemplateResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from boss template save, got %d", saveTemplateResponse.Code)
	}
	if store.lastBossTemplate.ID != "dragon" || store.lastBossTemplate.MaxHP != 80 {
		t.Fatalf("expected template payload to be forwarded, got %+v", store.lastBossTemplate)
	}

	saveLootRequest := httptest.NewRequest(http.MethodPut, "/api/admin/boss/pool/dragon/loot", strings.NewReader(`{"loot":[{"itemId":"fire-ring","dropRatePercent":35}]}`))
	saveLootRequest.Header.Set("Content-Type", "application/json")
	saveLootRequest.AddCookie(cookies[0])
	saveLootResponse := httptest.NewRecorder()
	handler.ServeHTTP(saveLootResponse, saveLootRequest)

	if saveLootResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from boss template loot save, got %d", saveLootResponse.Code)
	}
	if store.lastTemplateLootID != "dragon" || len(store.lastTemplateLoot) != 1 || store.lastTemplateLoot[0].ItemID != "fire-ring" || store.lastTemplateLoot[0].DropRatePercent != 35 {
		t.Fatalf("expected template loot payload to be forwarded, got id=%s loot=%+v", store.lastTemplateLootID, store.lastTemplateLoot)
	}

	saveQueueRequest := httptest.NewRequest(http.MethodPut, "/api/admin/boss/cycle/queue", strings.NewReader(`{"templateIds":["dragon","slime-king"]}`))
	saveQueueRequest.Header.Set("Content-Type", "application/json")
	saveQueueRequest.AddCookie(cookies[0])
	saveQueueResponse := httptest.NewRecorder()
	handler.ServeHTTP(saveQueueResponse, saveQueueRequest)

	if saveQueueResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from cycle queue save, got %d", saveQueueResponse.Code)
	}
	if len(store.lastCycleQueue) != 2 || store.lastCycleQueue[0] != "dragon" || store.lastCycleQueue[1] != "slime-king" {
		t.Fatalf("expected cycle queue payload to be forwarded, got %+v", store.lastCycleQueue)
	}

	enableCycleRequest := httptest.NewRequest(http.MethodPost, "/api/admin/boss/cycle/enable", nil)
	enableCycleRequest.AddCookie(cookies[0])
	enableCycleResponse := httptest.NewRecorder()
	handler.ServeHTTP(enableCycleResponse, enableCycleRequest)

	if enableCycleResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from cycle enable, got %d", enableCycleResponse.Code)
	}
	if !store.lastCycleEnabled {
		t.Fatal("expected cycle enable to be forwarded to store")
	}
}

func TestAdminBossPartsRequiredReturnsBadRequest(t *testing.T) {
	store := &mockStore{
		activateBossErr:     vote.ErrBossPartsRequired,
		saveBossTemplateErr: vote.ErrBossPartsRequired,
	}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		}),
	})

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{"username":"admin","password":"secret"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	handler.ServeHTTP(loginResponse, loginRequest)

	cookies := loginResponse.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie from login")
	}

	activateRequest := httptest.NewRequest(http.MethodPost, "/api/admin/boss/activate", strings.NewReader(`{"id":"slime-king","name":"史莱姆王","maxHp":50}`))
	activateRequest.Header.Set("Content-Type", "application/json")
	activateRequest.AddCookie(cookies[0])
	activateResponse := httptest.NewRecorder()
	handler.ServeHTTP(activateResponse, activateRequest)
	if activateResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 from boss activate with no parts, got %d", activateResponse.Code)
	}

	saveTemplateRequest := httptest.NewRequest(http.MethodPost, "/api/admin/boss/pool", strings.NewReader(`{"id":"dragon","name":"火龙","maxHp":80}`))
	saveTemplateRequest.Header.Set("Content-Type", "application/json")
	saveTemplateRequest.AddCookie(cookies[0])
	saveTemplateResponse := httptest.NewRecorder()
	handler.ServeHTTP(saveTemplateResponse, saveTemplateRequest)
	if saveTemplateResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 from boss pool save with no layout, got %d", saveTemplateResponse.Code)
	}
}

func TestAdminOSSPolicyRequiresAuthAndReturnsPayload(t *testing.T) {
	store := &mockStore{}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		}),
		OSSSigner: &mockOSSSigner{
			policy: ossupload.Policy{
				AccessKeyID:   "test-ak",
				Policy:        "policy",
				Signature:     "signature",
				Host:          "https://vote-wall.oss-cn-beijing.aliyuncs.com",
				Dir:           "buttons/20260419/",
				PublicBaseURL: "https://cdn.example.com",
			},
		},
	})

	unauthorizedRequest := httptest.NewRequest(http.MethodPost, "/api/admin/oss/sts", nil)
	unauthorizedResponse := httptest.NewRecorder()
	handler.ServeHTTP(unauthorizedResponse, unauthorizedRequest)

	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without session, got %d", unauthorizedResponse.Code)
	}

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{"username":"admin","password":"secret"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	handler.ServeHTTP(loginResponse, loginRequest)

	cookies := loginResponse.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie from login")
	}

	request := httptest.NewRequest(http.MethodPost, "/api/admin/oss/sts", nil)
	request.AddCookie(cookies[0])
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload map[string]any
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["host"] != "https://vote-wall.oss-cn-beijing.aliyuncs.com" {
		t.Fatalf("unexpected oss payload: %+v", payload)
	}
}
func TestValidateNicknameRejectsSensitiveNickname(t *testing.T) {
	store := &mockStore{
		validateErr: vote.ErrSensitiveNickname,
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/nickname/validate", strings.NewReader(`{"nickname":"我是习近平"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
	if body := response.Body.String(); !strings.Contains(body, "敏感词") {
		t.Fatalf("expected sensitive-word message, got %q", body)
	}
}
