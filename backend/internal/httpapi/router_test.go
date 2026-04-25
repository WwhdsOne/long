package httpapi

import (
	"context"
	"errors"
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
	adminButtonPage       vote.AdminButtonPage
	adminEquipmentPage    vote.AdminEquipmentPage
	adminBossHistoryPage  vote.AdminBossHistoryPage
	adminPlayerPage       vote.AdminPlayerPage
	adminPlayer           *vote.AdminPlayerOverview
	bossHistory           []vote.BossHistoryEntry
	announcements         []vote.Announcement
	latestAnnouncement    *vote.Announcement
	messagePage           vote.MessagePage
	result                vote.ClickResult
	lastButton            vote.ButtonUpsert
	lastBoss              vote.BossUpsert
	lastBossTemplate      vote.BossTemplateUpsert
	lastEquipment         vote.EquipmentDefinition
	lastTemplateLootID    string
	lastTemplateLoot      []vote.BossLootEntry
	lastCycleEnabled      bool
	lastSalvageItemID     string
	lastSalvageQuantity   int64
	lastClickNickname     string
	lastAutoClickNickname string
	lastGetStateNickname  string
	getStateErr           error
	clickErr              error
	equipErr              error
	validateErr           error
	messageErr            error
	salvageErr            error
}

func (m *mockStore) GetState(_ context.Context, nickname string) (vote.State, error) {
	m.lastGetStateNickname = nickname
	if m.getStateErr != nil {
		return vote.State{}, m.getStateErr
	}
	if len(m.snapshot.Buttons) > 0 || len(m.snapshot.Leaderboard) > 0 || m.snapshot.Boss != nil || m.snapshot.AnnouncementVersion != "" {
		return vote.ComposeState(m.snapshot, m.userStateForNickname(nickname)), nil
	}
	state := m.state
	if nickname == "" {
		state.UserStats = nil
	}
	return state, nil
}

func (m *mockStore) GetSnapshot(_ context.Context) (vote.Snapshot, error) {
	if len(m.snapshot.Buttons) > 0 || len(m.snapshot.Leaderboard) > 0 || m.snapshot.Boss != nil || m.snapshot.AnnouncementVersion != "" {
		return m.snapshot, nil
	}
	return vote.Snapshot{
		Buttons:     m.state.Buttons,
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
	userState.Gems = m.state.Gems
	userState.RecentRewards = m.state.RecentRewards
	userState.LastReward = m.state.LastReward
	return userState
}

func (m *mockStore) ClickButton(_ context.Context, slug string, nickname string) (vote.ClickResult, error) {
	m.lastClickNickname = nickname
	if m.clickErr != nil {
		return vote.ClickResult{}, m.clickErr
	}
	for index := range m.state.Buttons {
		if m.state.Buttons[index].Key == slug {
			if m.result.Button.Key == "" {
				m.state.Buttons[index].Count++
				if m.state.UserStats == nil && nickname != "" {
					m.state.UserStats = &vote.UserStats{Nickname: nickname}
				}
				if m.state.UserStats != nil {
					m.state.UserStats.ClickCount++
				}
				return vote.ClickResult{
					Button:   m.state.Buttons[index],
					Delta:    1,
					Critical: false,
					UserStats: vote.UserStats{
						Nickname:   nickname,
						ClickCount: 1,
					},
				}, nil
			}
			m.state.Buttons[index].Count = m.result.Button.Count
			return m.result, nil
		}
	}
	return vote.ClickResult{}, vote.ErrButtonNotFound
}

func (m *mockStore) AutoClickBossPart(_ context.Context, slug string, nickname string) (vote.ClickResult, error) {
	m.lastAutoClickNickname = nickname
	return m.ClickButton(context.Background(), slug, nickname)
}

func (m *mockStore) ValidateNickname(_ context.Context, _ string) error {
	return m.validateErr
}

func (m *mockStore) EquipItem(_ context.Context, _ string, _ string) (vote.State, error) {
	if m.equipErr != nil {
		return vote.State{}, m.equipErr
	}
	if len(m.equipState.Buttons) == 0 {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) UnequipItem(_ context.Context, _ string, _ string) (vote.State, error) {
	if m.equipErr != nil {
		return vote.State{}, m.equipErr
	}
	if len(m.equipState.Buttons) == 0 {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) GetAdminState(_ context.Context) (vote.AdminState, error) {
	return m.adminState, nil
}

func (m *mockStore) ListAdminButtonsPage(_ context.Context, _ int64, _ int64) (vote.AdminButtonPage, error) {
	return m.adminButtonPage, nil
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

func (m *mockStore) SaveButton(_ context.Context, button vote.ButtonUpsert) error {
	m.lastButton = button
	return nil
}

func (m *mockStore) SaveEquipmentDefinition(_ context.Context, definition vote.EquipmentDefinition) error {
	m.lastEquipment = definition
	return nil
}

func (m *mockStore) DeleteEquipmentDefinition(_ context.Context, _ string) error {
	return nil
}

func (m *mockStore) ActivateBoss(_ context.Context, boss vote.BossUpsert) (*vote.Boss, error) {
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

func (m *mockStore) SetBossCycleEnabled(_ context.Context, enabled bool) (*vote.Boss, error) {
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

func (m *mockStore) SelectTalentTree(_ context.Context, _ string, _ vote.TalentTree, _ vote.TalentTree) error {
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

type mockManualClickController struct {
	ticket       ClickTicket
	clickResult  vote.ClickResult
	issueErr     error
	clickErr     error
	lastIssueReq TicketIssueRequest
	lastClickReq ManualClickRequest
}

func (m *mockManualClickController) IssueTicket(_ context.Context, request TicketIssueRequest) (ClickTicket, error) {
	m.lastIssueReq = request
	if m.issueErr != nil {
		return ClickTicket{}, m.issueErr
	}
	return m.ticket, nil
}

func (m *mockManualClickController) Click(_ context.Context, request ManualClickRequest) (vote.ClickResult, error) {
	m.lastClickReq = request
	if m.clickErr != nil {
		return vote.ClickResult{}, m.clickErr
	}
	return m.clickResult, nil
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

func TestGetButtonsReturnsCurrentList(t *testing.T) {
	store := &mockStore{
		state: vote.State{
			Buttons: []vote.Button{
				{
					Key:      "feel",
					RedisKey: "vote:button:feel",
					Label:    "有感觉吗",
					Count:    2,
					Sort:     10,
					Enabled:  true,
				},
			},
			Leaderboard: []vote.LeaderboardEntry{
				{Rank: 1, Nickname: "阿明", ClickCount: 9},
			},
			Boss: &vote.Boss{
				ID:        "slime-king",
				Name:      "史莱姆王",
				Status:    "active",
				MaxHP:     100,
				CurrentHP: 80,
			},
			BossLoot: []vote.BossLootEntry{
				{
					ItemID:   "cloth-armor",
					ItemName: "布甲",
					Slot:     "armor",
					Weight:   3,
				},
			},
			LatestAnnouncement: &vote.Announcement{
				ID:          "7",
				Title:       "更新公告",
				Content:     "公告正文",
				PublishedAt: 1710000000,
				Active:      true,
			},
		},
		snapshot: vote.Snapshot{
			Buttons: []vote.Button{
				{
					Key:      "feel",
					RedisKey: "vote:button:feel",
					Label:    "有感觉吗",
					Count:    2,
					Sort:     10,
					Enabled:  true,
				},
			},
			Leaderboard: []vote.LeaderboardEntry{
				{Rank: 1, Nickname: "阿明", ClickCount: 9},
			},
			Boss: &vote.Boss{
				ID:        "slime-king",
				Name:      "史莱姆王",
				Status:    "active",
				MaxHP:     100,
				CurrentHP: 80,
			},
			AnnouncementVersion: "7",
		},
	}
	broadcaster := &mockBroadcaster{}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: broadcaster,
	})

	request := httptest.NewRequest(http.MethodGet, "/api/buttons", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload map[string]any
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	buttons, ok := payload["buttons"].([]any)
	if !ok || len(buttons) != 1 {
		t.Fatalf("unexpected buttons payload: %+v", payload["buttons"])
	}
	leaderboard, ok := payload["leaderboard"].([]any)
	if !ok || len(leaderboard) != 1 {
		t.Fatalf("unexpected leaderboard payload: %+v", payload["leaderboard"])
	}
	if payload["announcementVersion"] != "7" {
		t.Fatalf("unexpected announcement version payload: %+v", payload)
	}
	if _, exists := payload["bossLoot"]; exists {
		t.Fatalf("expected public buttons payload to omit bossLoot, got %+v", payload)
	}
	if _, exists := payload["bossHeroLoot"]; exists {
		t.Fatalf("expected public buttons payload to omit bossHeroLoot, got %+v", payload)
	}
	if _, exists := payload["latestAnnouncement"]; exists {
		t.Fatalf("expected public buttons payload to omit latestAnnouncement, got %+v", payload)
	}

	if len(broadcaster.snapshots) != 0 {
		t.Fatalf("expected no broadcasts, got %d", len(broadcaster.snapshots))
	}
}

func TestButtonPagesRouteIsRemoved(t *testing.T) {
	store := &mockStore{}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodGet, "/api/buttons/pages?page=2&pageSize=9", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected removed route to return 404, got %d", response.Code)
	}
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
func TestClickButtonDoesNotUseLegacySnapshotBroadcast(t *testing.T) {
	store := &mockStore{
		state: vote.State{
			Buttons: []vote.Button{
				{
					Key:      "feel",
					RedisKey: "vote:button:feel",
					Label:    "有感觉吗",
					Count:    2,
					Sort:     10,
					Enabled:  true,
				},
			},
			Leaderboard: []vote.LeaderboardEntry{
				{Rank: 1, Nickname: "阿明", ClickCount: 3},
			},
			UserStats: &vote.UserStats{Nickname: "阿明", ClickCount: 2},
		},
	}
	broadcaster := &mockBroadcaster{}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: broadcaster,
	})

	request := httptest.NewRequest(http.MethodPost, "/api/buttons/feel/click", strings.NewReader(`{"nickname":"阿明"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload struct {
		Button      vote.Button             `json:"button"`
		Buttons     []vote.Button           `json:"buttons"`
		Delta       int64                   `json:"delta"`
		Critical    bool                    `json:"critical"`
		UserStats   vote.UserStats          `json:"userStats"`
		Leaderboard []vote.LeaderboardEntry `json:"leaderboard"`
	}
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Button.Count != 3 {
		t.Fatalf("expected count 3, got %d", payload.Button.Count)
	}
	if payload.Delta != 1 || payload.Critical {
		t.Fatalf("expected normal click payload, got delta=%d critical=%v", payload.Delta, payload.Critical)
	}
	if payload.UserStats.Nickname != "阿明" {
		t.Fatalf("expected user stats for 阿明, got %+v", payload.UserStats)
	}

	if len(broadcaster.snapshots) != 0 {
		t.Fatalf("expected no legacy snapshot broadcast, got %+v", broadcaster.snapshots)
	}
}

func TestClickButtonReturnsCriticalMetadata(t *testing.T) {
	store := &mockStore{
		state: vote.State{
			Buttons: []vote.Button{
				{
					Key:      "feel",
					RedisKey: "vote:button:feel",
					Label:    "有感觉吗",
					Count:    2,
					Sort:     10,
					Enabled:  true,
				},
			},
			Leaderboard: []vote.LeaderboardEntry{
				{Rank: 1, Nickname: "阿明", ClickCount: 7},
			},
			UserStats: &vote.UserStats{Nickname: "阿明", ClickCount: 7},
		},
		result: vote.ClickResult{
			Button: vote.Button{
				Key:      "feel",
				RedisKey: "vote:button:feel",
				Label:    "有感觉吗",
				Count:    7,
				Sort:     10,
				Enabled:  true,
			},
			Delta:    5,
			Critical: true,
			UserStats: vote.UserStats{
				Nickname:   "阿明",
				ClickCount: 7,
			},
		},
	}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/buttons/feel/click", strings.NewReader(`{"nickname":"阿明"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload struct {
		Delta    int64 `json:"delta"`
		Critical bool  `json:"critical"`
	}
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Delta != 5 || !payload.Critical {
		t.Fatalf("expected critical payload, got delta=%d critical=%v", payload.Delta, payload.Critical)
	}
}

func TestClickButtonReturnsMinimalResponseForRealtimeClients(t *testing.T) {
	store := &mockStore{
		result: vote.ClickResult{
			Button: vote.Button{
				Key:      "feel",
				RedisKey: "vote:button:feel",
				Label:    "有感觉吗",
				Count:    7,
				Sort:     10,
				Enabled:  true,
			},
			Delta:     5,
			Critical:  true,
			UserStats: vote.UserStats{Nickname: "阿明", ClickCount: 7},
			Boss: &vote.Boss{
				ID:        "boss-1",
				Name:      "木桩王",
				Status:    "active",
				MaxHP:     100,
				CurrentHP: 40,
			},
			BossLeaderboard: []vote.BossLeaderboardEntry{
				{Rank: 1, Nickname: "阿明", Damage: 60},
			},
			MyBossStats: &vote.BossUserStats{Nickname: "阿明", Damage: 60},
			RecentRewards: []vote.Reward{
				{BossID: "boss-1", BossName: "木桩王", ItemID: "club", ItemName: "木棒", GrantedAt: 123},
			},
			LastReward: &vote.Reward{BossID: "boss-1", BossName: "木桩王", ItemID: "club", ItemName: "木棒", GrantedAt: 123},
		},
		state: vote.State{
			Buttons: []vote.Button{
				{Key: "feel", Label: "有感觉吗", Count: 6, Sort: 10, Enabled: true},
			},
		},
	}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/buttons/feel/click", strings.NewReader(`{"nickname":"阿明","realtimeConnected":true}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload map[string]any
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	for _, key := range []string{"button", "delta", "critical"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected key %q in payload: %+v", key, payload)
		}
	}
	for _, key := range []string{"userStats", "boss", "bossLeaderboard", "myBossStats", "recentRewards", "lastReward"} {
		if _, ok := payload[key]; ok {
			t.Fatalf("expected realtime payload to omit %q: %+v", key, payload)
		}
	}
}

func TestClickButtonReturnsFallbackStateWhenRealtimeDisconnected(t *testing.T) {
	store := &mockStore{
		result: vote.ClickResult{
			Button: vote.Button{
				Key:      "feel",
				RedisKey: "vote:button:feel",
				Label:    "有感觉吗",
				Count:    7,
				Sort:     10,
				Enabled:  true,
			},
			Delta:     5,
			Critical:  true,
			UserStats: vote.UserStats{Nickname: "阿明", ClickCount: 7},
			Boss: &vote.Boss{
				ID:        "boss-1",
				Name:      "木桩王",
				Status:    "active",
				MaxHP:     100,
				CurrentHP: 40,
			},
			BossLeaderboard: []vote.BossLeaderboardEntry{
				{Rank: 1, Nickname: "阿明", Damage: 60},
			},
			MyBossStats: &vote.BossUserStats{Nickname: "阿明", Damage: 60},
			RecentRewards: []vote.Reward{
				{BossID: "boss-1", BossName: "木桩王", ItemID: "club", ItemName: "木棒", GrantedAt: 123},
			},
			LastReward: &vote.Reward{BossID: "boss-1", BossName: "木桩王", ItemID: "club", ItemName: "木棒", GrantedAt: 123},
		},
		state: vote.State{
			Buttons: []vote.Button{
				{Key: "feel", Label: "有感觉吗", Count: 6, Sort: 10, Enabled: true},
			},
		},
	}

	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/buttons/feel/click", strings.NewReader(`{"nickname":"阿明","realtimeConnected":false}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload map[string]any
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	for _, key := range []string{"button", "delta", "critical", "userStats", "boss", "bossLeaderboard", "myBossStats", "recentRewards", "lastReward"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected key %q in payload: %+v", key, payload)
		}
	}
}

func TestClickButtonPublishesStateChangeWithoutRefetchingState(t *testing.T) {
	store := &mockStore{
		getStateErr: context.DeadlineExceeded,
		result: vote.ClickResult{
			Button: vote.Button{
				Key:      "feel",
				RedisKey: "vote:button:feel",
				Label:    "有感觉吗",
				Count:    5,
				Sort:     10,
				Enabled:  true,
			},
			Delta:    1,
			Critical: false,
			UserStats: vote.UserStats{
				Nickname:   "阿明",
				ClickCount: 5,
			},
			BroadcastUserAll: true,
		},
		state: vote.State{
			Buttons: []vote.Button{
				{Key: "feel", Label: "有感觉吗", Count: 4, Sort: 10, Enabled: true},
			},
		},
	}
	changePublisher := &mockChangePublisher{}

	handler := NewHandler(Options{
		Store:           store,
		Broadcaster:     &mockBroadcaster{},
		ChangePublisher: changePublisher,
	})

	request := httptest.NewRequest(http.MethodPost, "/api/buttons/feel/click", strings.NewReader(`{"nickname":"阿明"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload struct {
		Button    vote.Button    `json:"button"`
		Delta     int64          `json:"delta"`
		Critical  bool           `json:"critical"`
		UserStats vote.UserStats `json:"userStats"`
	}
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Button.Count != 5 || payload.UserStats.ClickCount != 5 {
		t.Fatalf("unexpected click payload: %+v", payload)
	}
	if len(changePublisher.changes) != 1 {
		t.Fatalf("expected one published change, got %+v", changePublisher.changes)
	}
	if changePublisher.changes[0].Type != vote.StateChangeBossChanged || changePublisher.changes[0].Nickname != "阿明" {
		t.Fatalf("unexpected published change: %+v", changePublisher.changes[0])
	}
	if !changePublisher.changes[0].BroadcastUserAll {
		t.Fatalf("expected BroadcastUserAll to be preserved, got %+v", changePublisher.changes[0])
	}
}

func TestEquipItemReturnsUpdatedState(t *testing.T) {
	store := &mockStore{
		equipState: vote.State{
			Buttons: []vote.Button{
				{
					Key:      "feel",
					RedisKey: "vote:button:feel",
					Label:    "有感觉吗",
					Count:    3,
					Sort:     10,
					Enabled:  true,
				},
			},
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

	request := httptest.NewRequest(http.MethodPost, "/api/equipment/wood-sword/equip", strings.NewReader(`{"nickname":"阿明"}`))
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

func TestSynthesizeItemReturnsDeprecatedError(t *testing.T) {
	handler := NewHandler(Options{
		Store:       &mockStore{},
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/equipment/wood-sword/synthesize", strings.NewReader(`{"nickname":"阿明"}`))
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

func TestClickButtonUsesManualClickControllerWhenConfigured(t *testing.T) {
	controller := &mockManualClickController{
		clickResult: vote.ClickResult{
			Button: vote.Button{
				Key:     "feel",
				Label:   "有感觉吗",
				Count:   5,
				Enabled: true,
			},
			Delta: 1,
			UserStats: vote.UserStats{
				Nickname:   "阿明",
				ClickCount: 5,
			},
		},
	}
	handler := NewHandler(Options{
		Store:               &mockStore{state: voteStateForPlayerTests()},
		Broadcaster:         &mockBroadcaster{},
		PlayerAuthenticator: &mockPlayerAuthenticator{verifyNickname: "阿明"},
		ManualClick:         controller,
	})

	request := httptest.NewRequest(http.MethodPost, "/api/buttons/feel/click", strings.NewReader(`{"ticket":"ticket-1","realtimeConnected":true}`))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(&http.Cookie{Name: playerSessionCookieName, Value: "player-token"})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200 from click route, got %d", response.Code)
	}
	if controller.lastClickReq.Nickname != "阿明" || controller.lastClickReq.Ticket != "ticket-1" || controller.lastClickReq.EntryType != clickEntryHTTP {
		t.Fatalf("expected click controller to receive ticket protocol, got %+v", controller.lastClickReq)
	}
}

func TestAutoClickRoutesUseController(t *testing.T) {
	controller := &mockAutoClickController{
		status: AutoClickStatus{
			Active:        true,
			ButtonKey:     "feel",
			IntervalMs:    333,
			RatePerSecond: 3,
		},
	}
	handler := NewHandler(Options{
		Store:               &mockStore{state: voteStateForPlayerTests()},
		Broadcaster:         &mockBroadcaster{},
		PlayerAuthenticator: &mockPlayerAuthenticator{verifyNickname: "阿明"},
		AutoClick:           controller,
	})

	statusRequest := httptest.NewRequest(http.MethodGet, "/api/auto-click", nil)
	statusRequest.AddCookie(&http.Cookie{Name: playerSessionCookieName, Value: "player-token"})
	statusResponse := httptest.NewRecorder()
	handler.ServeHTTP(statusResponse, statusRequest)
	if statusResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from auto-click status, got %d", statusResponse.Code)
	}

	startRequest := httptest.NewRequest(http.MethodPost, "/api/auto-click/start", strings.NewReader(`{"slug":"understand"}`))
	startRequest.Header.Set("Content-Type", "application/json")
	startRequest.AddCookie(&http.Cookie{Name: playerSessionCookieName, Value: "player-token"})
	startResponse := httptest.NewRecorder()
	handler.ServeHTTP(startResponse, startRequest)
	if startResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from auto-click start, got %d", startResponse.Code)
	}
	if controller.lastStartNickname != "阿明" || controller.lastStartSlug != "understand" {
		t.Fatalf("expected auto-click start to forward nickname and slug, got nickname=%q slug=%q", controller.lastStartNickname, controller.lastStartSlug)
	}

	stopRequest := httptest.NewRequest(http.MethodPost, "/api/auto-click/stop", nil)
	stopRequest.AddCookie(&http.Cookie{Name: playerSessionCookieName, Value: "player-token"})
	stopResponse := httptest.NewRecorder()
	handler.ServeHTTP(stopResponse, stopRequest)
	if stopResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from auto-click stop, got %d", stopResponse.Code)
	}
	if controller.lastStopNickname != "阿明" {
		t.Fatalf("expected auto-click stop to use 阿明, got %q", controller.lastStopNickname)
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

func TestAdminActivateBossAndSaveButton(t *testing.T) {
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

	activateRequest := httptest.NewRequest(http.MethodPost, "/api/admin/boss/activate", strings.NewReader(`{"id":"slime-king","name":"史莱姆王","maxHp":50}`))
	activateRequest.Header.Set("Content-Type", "application/json")
	activateRequest.AddCookie(cookies[0])
	activateResponse := httptest.NewRecorder()
	handler.ServeHTTP(activateResponse, activateRequest)

	if activateResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from boss activate, got %d", activateResponse.Code)
	}
	if store.lastBoss.ID != "slime-king" || store.lastBoss.MaxHP != 50 {
		t.Fatalf("expected boss payload to be forwarded, got %+v", store.lastBoss)
	}

	saveButtonRequest := httptest.NewRequest(http.MethodPost, "/api/admin/buttons", strings.NewReader(`{"slug":"new-one","label":"新按钮","sort":40,"enabled":true}`))
	saveButtonRequest.Header.Set("Content-Type", "application/json")
	saveButtonRequest.AddCookie(cookies[0])
	saveButtonResponse := httptest.NewRecorder()
	handler.ServeHTTP(saveButtonResponse, saveButtonRequest)

	if saveButtonResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from save button, got %d", saveButtonResponse.Code)
	}
	if store.lastButton.Slug != "new-one" || store.lastButton.Label != "新按钮" {
		t.Fatalf("expected button payload to be forwarded, got %+v", store.lastButton)
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

	saveTemplateRequest := httptest.NewRequest(http.MethodPost, "/api/admin/boss/pool", strings.NewReader(`{"id":"dragon","name":"火龙","maxHp":80}`))
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

	saveLootRequest := httptest.NewRequest(http.MethodPut, "/api/admin/boss/pool/dragon/loot", strings.NewReader(`{"loot":[{"itemId":"fire-ring","weight":3}]}`))
	saveLootRequest.Header.Set("Content-Type", "application/json")
	saveLootRequest.AddCookie(cookies[0])
	saveLootResponse := httptest.NewRecorder()
	handler.ServeHTTP(saveLootResponse, saveLootRequest)

	if saveLootResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from boss template loot save, got %d", saveLootResponse.Code)
	}
	if store.lastTemplateLootID != "dragon" || len(store.lastTemplateLoot) != 1 || store.lastTemplateLoot[0].ItemID != "fire-ring" {
		t.Fatalf("expected template loot payload to be forwarded, got id=%s loot=%+v", store.lastTemplateLootID, store.lastTemplateLoot)
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

func TestClickMissingButtonReturnsNotFound(t *testing.T) {
	store := &mockStore{
		state: vote.State{
			Buttons: []vote.Button{
				{
					Key:     "feel",
					Label:   "有感觉吗",
					Enabled: true,
				},
			},
		},
	}
	handler := NewHandler(Options{
		Store:               store,
		Broadcaster:         &mockBroadcaster{},
		PlayerAuthenticator: &mockPlayerAuthenticator{verifyNickname: "阿明"},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/buttons/missing/click", strings.NewReader(`{"nickname":"阿明"}`))
	request.AddCookie(&http.Cookie{Name: playerSessionCookieName, Value: "player-token"})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
}

func TestClickRequiresNickname(t *testing.T) {
	store := &mockStore{
		state: vote.State{
			Buttons: []vote.Button{
				{Key: "feel", Label: "有感觉吗", Enabled: true},
			},
		},
	}
	handler := NewHandler(Options{
		Store:               store,
		Broadcaster:         &mockBroadcaster{},
		PlayerAuthenticator: &mockPlayerAuthenticator{verifyErr: errors.New("missing")},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/buttons/feel/click", strings.NewReader(`{"nickname":"   "}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
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

func TestGetButtonsRejectsSensitiveNickname(t *testing.T) {
	store := &mockStore{
		getStateErr: vote.ErrSensitiveNickname,
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodGet, "/api/buttons?nickname=%E4%B9%A0%E8%BF%91%E5%B9%B3", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}
