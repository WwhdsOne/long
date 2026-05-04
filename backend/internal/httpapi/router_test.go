package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bytedance/sonic"

	"long/internal/admin"
	"long/internal/core"
	ossupload "long/internal/oss"
)

type mockStore struct {
	state                     core.State
	talentState               *core.TalentState
	snapshot                  core.Snapshot
	equipState                core.State
	adminState                core.AdminState
	bossResources             core.BossResources
	adminEquipmentPage        core.AdminEquipmentPage
	adminBossHistoryPage      core.AdminBossHistoryPage
	adminPlayerPage           core.AdminPlayerPage
	adminPlayer               *core.AdminPlayerOverview
	bossHistory               []core.BossHistoryEntry
	announcements             []core.Announcement
	latestAnnouncement        *core.Announcement
	messagePage               core.MessagePage
	bossResourcesByNickname   map[string]core.BossResources
	roomList                  core.RoomList
	roomSwitchResult          core.RoomSwitchResult
	roomDisplayNames          map[string]string
	tasks                     []core.PlayerTask
	shopItems                 []core.ShopCatalogItemView
	result                    core.ClickResult
	lastBoss                  core.BossUpsert
	lastBossRoomID            string
	lastBossTemplate          core.BossTemplateUpsert
	lastEquipment             core.EquipmentDefinition
	lastShopItemID            string
	lastClaimTaskID           string
	lastTaskDefinition        core.TaskDefinition
	lastArchiveNow            time.Time
	lastTemplateLootID        string
	lastTemplateLoot          []core.BossLootEntry
	lastCycleQueue            []string
	lastCycleEnabled          bool
	lastCycleRoomID           string
	lastDeactivateRoomID      string
	lastBossResourcesNickname string
	lastListRoomsNickname     string
	lastSwitchRoomNickname    string
	lastSwitchRoomID          string
	lastSalvageItemID         string
	lastSalvageQuantity       int64
	lastLockItemID            string
	lastLockState             bool
	lastClickNickname         string
	lastAutoClickNickname     string
	lastGetStateNickname      string
	saveTaskErr               error
	activateTaskErr           error
	deactivateTaskErr         error
	duplicateTaskErr          error
	getStateErr               error
	clickErr                  error
	equipErr                  error
	enhanceErr                error
	validateErr               error
	messageErr                error
	salvageErr                error
	activateBossErr           error
	saveBossTemplateErr       error
	setBossCycleErr           error
	purchaseShopErr           error
	equipShopErr              error
	roomSwitchErr             error
}

func (m *mockStore) GetState(_ context.Context, nickname string) (core.State, error) {
	m.lastGetStateNickname = nickname
	if m.getStateErr != nil {
		return core.State{}, m.getStateErr
	}
	if len(m.snapshot.Leaderboard) > 0 || m.snapshot.Boss != nil || m.snapshot.AnnouncementVersion != "" {
		return core.ComposeState(m.snapshot, m.userStateForNickname(nickname)), nil
	}
	state := m.state
	if nickname == "" {
		state.UserStats = nil
	}
	return state, nil
}

func (m *mockStore) GetSnapshot(_ context.Context) (core.Snapshot, error) {
	if len(m.snapshot.Leaderboard) > 0 || m.snapshot.Boss != nil || m.snapshot.AnnouncementVersion != "" {
		return m.snapshot, nil
	}
	return core.Snapshot{
		Leaderboard: m.state.Leaderboard,
	}, nil
}

func (m *mockStore) GetBossResources(_ context.Context) (core.BossResources, error) {
	return m.bossResources, nil
}

func (m *mockStore) GetBossResourcesForNickname(_ context.Context, nickname string) (core.BossResources, error) {
	m.lastBossResourcesNickname = nickname
	if m.bossResourcesByNickname != nil {
		if resources, ok := m.bossResourcesByNickname[nickname]; ok {
			return resources, nil
		}
	}
	return m.bossResources, nil
}

func (m *mockStore) ListRooms(_ context.Context, nickname string) (core.RoomList, error) {
	m.lastListRoomsNickname = nickname
	return m.roomList, nil
}

func (m *mockStore) SetRoomDisplayName(_ context.Context, roomID string, displayName string) error {
	if !m.hasRoom(roomID) {
		return core.ErrRoomNotFound
	}
	if m.roomDisplayNames == nil {
		m.roomDisplayNames = map[string]string{}
	}
	if strings.TrimSpace(displayName) == "" {
		delete(m.roomDisplayNames, roomID)
		for index, room := range m.roomList.Rooms {
			if room.ID == roomID {
				m.roomList.Rooms[index].DisplayName = "房间 " + roomID
				break
			}
		}
		return nil
	}
	m.roomDisplayNames[roomID] = displayName
	for index, room := range m.roomList.Rooms {
		if room.ID == roomID {
			m.roomList.Rooms[index].DisplayName = displayName
			break
		}
	}
	return nil
}

func (m *mockStore) hasRoom(roomID string) bool {
	for _, room := range m.roomList.Rooms {
		if room.ID == roomID {
			return true
		}
	}
	return false
}

func (m *mockStore) SwitchPlayerRoom(_ context.Context, nickname string, roomID string) (core.RoomSwitchResult, error) {
	m.lastSwitchRoomNickname = nickname
	m.lastSwitchRoomID = roomID
	if m.roomSwitchErr != nil {
		return core.RoomSwitchResult{}, m.roomSwitchErr
	}
	return m.roomSwitchResult, nil
}

func (m *mockStore) GetUserState(_ context.Context, nickname string) (core.UserState, error) {
	if m.getStateErr != nil {
		return core.UserState{}, m.getStateErr
	}
	return m.userStateForNickname(nickname), nil
}

func (m *mockStore) userStateForNickname(nickname string) core.UserState {
	userState := core.UserState{
		Inventory:   []core.InventoryItem{},
		Loadout:     core.Loadout{},
		CombatStats: core.CombatStats{},
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
	userState.Tasks = m.tasks
	userState.EquippedBattleClickSkinID = m.state.EquippedBattleClickSkinID
	userState.EquippedBattleClickCursorImagePath = m.state.EquippedBattleClickCursorImagePath
	return userState
}

func (m *mockStore) ListShopCatalogItemsForPlayer(_ context.Context, _ string) ([]core.ShopCatalogItemView, error) {
	if m.shopItems == nil {
		return []core.ShopCatalogItemView{}, nil
	}
	return append([]core.ShopCatalogItemView(nil), m.shopItems...), nil
}

func (m *mockStore) PurchaseShopItem(_ context.Context, _ string, itemID string) (core.UserState, error) {
	m.lastShopItemID = itemID
	if m.purchaseShopErr != nil {
		return core.UserState{}, m.purchaseShopErr
	}
	return m.userStateForNickname("阿明"), nil
}

func (m *mockStore) EquipShopItem(_ context.Context, _ string, itemID string) (core.UserState, error) {
	m.lastShopItemID = itemID
	if m.equipShopErr != nil {
		return core.UserState{}, m.equipShopErr
	}
	return m.userStateForNickname("阿明"), nil
}

func (m *mockStore) UnequipShopItem(_ context.Context, _ string) (core.UserState, error) {
	m.state.EquippedBattleClickCursorImagePath = ""
	m.state.EquippedBattleClickSkinID = ""
	return m.userStateForNickname("阿明"), nil
}

func (m *mockStore) ListShopItems(_ context.Context) ([]core.ShopItem, error) {
	result := make([]core.ShopItem, 0, len(m.shopItems))
	for _, item := range m.shopItems {
		result = append(result, core.ShopItem{
			ItemID:                     item.ItemID,
			Title:                      item.Title,
			ItemType:                   item.ItemType,
			PriceGold:                  item.PriceGold,
			ImagePath:                  item.ImagePath,
			ImageAlt:                   item.ImageAlt,
			PreviewImagePath:           item.PreviewImagePath,
			BattleClickCursorImagePath: item.BattleClickCursorImagePath,
			Description:                item.Description,
			Active:                     item.Active,
			SortOrder:                  item.SortOrder,
			AutoEquipOnPurchase:        item.AutoEquipOnPurchase,
		})
	}
	return result, nil
}

func (m *mockStore) SaveShopItem(_ context.Context, item core.ShopItem) error {
	m.lastShopItemID = item.ItemID
	return nil
}

func (m *mockStore) DeleteShopItem(_ context.Context, itemID string) error {
	m.lastShopItemID = itemID
	return nil
}

func (m *mockStore) ListTasksForPlayer(_ context.Context, _ string) ([]core.PlayerTask, error) {
	return m.tasks, nil
}

func (m *mockStore) ClaimTaskReward(_ context.Context, _ string, taskID string) (core.UserState, error) {
	m.lastClaimTaskID = taskID
	return m.userStateForNickname("阿明"), nil
}

func (m *mockStore) ListTaskDefinitions(_ context.Context) ([]core.TaskDefinition, error) {
	if m.tasks == nil {
		return []core.TaskDefinition{}, nil
	}
	items := make([]core.TaskDefinition, 0, len(m.tasks))
	for _, item := range m.tasks {
		items = append(items, core.TaskDefinition{
			TaskID:        item.TaskID,
			Title:         item.Title,
			Description:   item.Description,
			TaskType:      item.TaskType,
			ConditionKind: item.ConditionKind,
			TargetValue:   item.TargetValue,
			Rewards:       item.Rewards,
			DisplayOrder:  item.DisplayOrder,
			StartAt:       item.StartAt,
			EndAt:         item.EndAt,
		})
	}
	return items, nil
}

func (m *mockStore) SaveTaskDefinition(_ context.Context, item core.TaskDefinition) error {
	m.lastTaskDefinition = item
	if m.saveTaskErr != nil {
		return m.saveTaskErr
	}
	return nil
}

func (m *mockStore) ActivateTaskDefinition(_ context.Context, taskID string) error {
	if m.activateTaskErr != nil {
		return m.activateTaskErr
	}
	m.lastTaskDefinition.TaskID = taskID
	m.lastTaskDefinition.Status = core.TaskStatusActive
	return nil
}

func (m *mockStore) DeactivateTaskDefinition(_ context.Context, taskID string) error {
	if m.deactivateTaskErr != nil {
		return m.deactivateTaskErr
	}
	m.lastTaskDefinition.TaskID = taskID
	m.lastTaskDefinition.Status = core.TaskStatusInactive
	return nil
}

func (m *mockStore) DuplicateTaskDefinition(_ context.Context, taskID string, newTaskID string) (*core.TaskDefinition, error) {
	if m.duplicateTaskErr != nil {
		return nil, m.duplicateTaskErr
	}
	item := &core.TaskDefinition{TaskID: newTaskID}
	if item.TaskID == "" {
		item.TaskID = taskID + "-copy"
	}
	return item, nil
}

func TestAdminTaskRoutesReturnReadableMessagesOnBusinessErrors(t *testing.T) {
	store := &mockStore{
		saveTaskErr:      core.ErrTaskNotClaimable,
		activateTaskErr:  core.ErrTaskNotClaimable,
		duplicateTaskErr: core.ErrTaskImmutable,
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "task-secret",
		}),
	})

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{"username":"admin","password":"secret"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	handler.ServeHTTP(loginResponse, loginRequest)
	cookies := loginResponse.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected admin login to set cookie")
	}

	saveRequest := httptest.NewRequest(http.MethodPost, "/api/admin/tasks", strings.NewReader(`{"taskId":"limited-click","title":"限时点击","taskType":"limited","conditionKind":"daily_clicks","targetValue":3}`))
	saveRequest.Header.Set("Content-Type", "application/json")
	saveRequest.AddCookie(cookies[0])
	saveResponse := httptest.NewRecorder()
	handler.ServeHTTP(saveResponse, saveRequest)
	if saveResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 from task save, got %d", saveResponse.Code)
	}
	if !strings.Contains(saveResponse.Body.String(), "任务定义不合法") {
		t.Fatalf("expected readable task save message, got %s", saveResponse.Body.String())
	}

	activateRequest := httptest.NewRequest(http.MethodPost, "/api/admin/tasks/limited-click/activate", nil)
	activateRequest.AddCookie(cookies[0])
	activateResponse := httptest.NewRecorder()
	handler.ServeHTTP(activateResponse, activateRequest)
	if activateResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 from task activate, got %d", activateResponse.Code)
	}
	if !strings.Contains(activateResponse.Body.String(), "任务定义不合法") {
		t.Fatalf("expected readable task activate message, got %s", activateResponse.Body.String())
	}

	duplicateRequest := httptest.NewRequest(http.MethodPost, "/api/admin/tasks/limited-click/duplicate", strings.NewReader(`{"taskId":"existing-copy"}`))
	duplicateRequest.Header.Set("Content-Type", "application/json")
	duplicateRequest.AddCookie(cookies[0])
	duplicateResponse := httptest.NewRecorder()
	handler.ServeHTTP(duplicateResponse, duplicateRequest)
	if duplicateResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 from task duplicate, got %d", duplicateResponse.Code)
	}
	if !strings.Contains(duplicateResponse.Body.String(), "生效中的任务") {
		t.Fatalf("expected readable task duplicate message, got %s", duplicateResponse.Body.String())
	}
}

func TestShopRoutesReturnCatalogAndSupportAuthenticatedPurchase(t *testing.T) {
	store := &mockStore{
		state: core.State{
			UserStats:                          &core.UserStats{Nickname: "阿明", ClickCount: 8},
			Gold:                               80,
			EquippedBattleClickSkinID:          "skin-basic",
			EquippedBattleClickCursorImagePath: "https://example.com/basic.png",
		},
		shopItems: []core.ShopCatalogItemView{{
			ShopItem: core.ShopItem{
				ItemID:                     "skin-basic",
				Title:                      "基础剑光",
				ItemType:                   core.ShopItemTypeBattleClickSkin,
				PriceGold:                  120,
				PreviewImagePath:           "https://example.com/preview.png",
				BattleClickCursorImagePath: "https://example.com/basic.png",
				Active:                     true,
			},
			Owned:    true,
			Equipped: true,
		}},
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "task-secret",
		}),
		PlayerAuthenticator: &mockPlayerAuthenticator{
			loginToken:     "player-token",
			loginNickname:  "阿明",
			verifyNickname: "阿明",
		},
	})

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/player/auth/login", strings.NewReader(`{"nickname":"阿明","password":"secret"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	handler.ServeHTTP(loginResponse, loginRequest)
	cookies := loginResponse.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected player login to set cookie")
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/api/shop/items", nil)
	listRequest.AddCookie(cookies[0])
	listResponse := httptest.NewRecorder()
	handler.ServeHTTP(listResponse, listRequest)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from shop list, got %d", listResponse.Code)
	}
	if !strings.Contains(listResponse.Body.String(), `"owned":true`) || !strings.Contains(listResponse.Body.String(), `"equipped":true`) {
		t.Fatalf("expected owned and equipped flags in shop list, got %s", listResponse.Body.String())
	}

	purchaseRequest := httptest.NewRequest(http.MethodPost, "/api/shop/items/skin-basic/purchase", nil)
	purchaseRequest.AddCookie(cookies[0])
	purchaseResponse := httptest.NewRecorder()
	handler.ServeHTTP(purchaseResponse, purchaseRequest)
	if purchaseResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from shop purchase, got %d", purchaseResponse.Code)
	}
	if store.lastShopItemID != "skin-basic" {
		t.Fatalf("expected purchase to target skin-basic, got %q", store.lastShopItemID)
	}
	if !strings.Contains(purchaseResponse.Body.String(), `"equippedBattleClickSkinId":"skin-basic"`) {
		t.Fatalf("expected purchase response to include equipped skin, got %s", purchaseResponse.Body.String())
	}
}

func TestShopPurchaseRouteReturnsReadableBusinessErrors(t *testing.T) {
	store := &mockStore{
		purchaseShopErr: core.ErrShopInsufficientGold,
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
		PlayerAuthenticator: &mockPlayerAuthenticator{
			loginToken:     "player-token",
			loginNickname:  "阿明",
			verifyNickname: "阿明",
		},
	})

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/player/auth/login", strings.NewReader(`{"nickname":"阿明","password":"secret"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	handler.ServeHTTP(loginResponse, loginRequest)
	cookies := loginResponse.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected player login to set cookie")
	}

	purchaseRequest := httptest.NewRequest(http.MethodPost, "/api/shop/items/skin-basic/purchase", nil)
	purchaseRequest.AddCookie(cookies[0])
	purchaseResponse := httptest.NewRecorder()
	handler.ServeHTTP(purchaseResponse, purchaseRequest)
	if purchaseResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 from shop purchase, got %d", purchaseResponse.Code)
	}
	if !strings.Contains(purchaseResponse.Body.String(), "金币不足") {
		t.Fatalf("expected readable insufficient gold message, got %s", purchaseResponse.Body.String())
	}
}

func (m *mockStore) ArchiveExpiredTaskCycles(_ context.Context, now time.Time) ([]core.TaskCycleArchive, error) {
	m.lastArchiveNow = now
	return []core.TaskCycleArchive{{TaskID: "daily-click-1", CycleKey: "2026-05-01"}}, nil
}

func (m *mockStore) ListTaskCycleArchives(_ context.Context, taskID string) ([]core.TaskCycleArchive, error) {
	return []core.TaskCycleArchive{{TaskID: taskID, CycleKey: "2026-05-01"}}, nil
}

func (m *mockStore) GetTaskCycleResults(_ context.Context, taskID string, cycleKey string) (core.TaskCycleResultsView, error) {
	return core.TaskCycleResultsView{
		Archive: core.TaskCycleArchive{TaskID: taskID, CycleKey: cycleKey},
		Items:   []core.TaskCyclePlayerResult{{TaskID: taskID, CycleKey: cycleKey, Nickname: "阿明"}},
	}, nil
}

func TestAdminTaskRoutesListAndSave(t *testing.T) {
	store := &mockStore{
		tasks: []core.PlayerTask{
			{TaskID: "daily-click-1", Title: "今日点击", TargetValue: 10},
		},
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "task-secret",
		}),
	})

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{"username":"admin","password":"secret"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	handler.ServeHTTP(loginResponse, loginRequest)
	cookies := loginResponse.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected admin login to set cookie")
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/api/admin/tasks", nil)
	listRequest.AddCookie(cookies[0])
	listResponse := httptest.NewRecorder()
	handler.ServeHTTP(listResponse, listRequest)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from admin task list, got %d", listResponse.Code)
	}

	saveRequest := httptest.NewRequest(http.MethodPost, "/api/admin/tasks", strings.NewReader(`{"taskId":"daily-click-2","title":"今日强化","taskType":"daily","conditionKind":"enhance_count","targetValue":3}`))
	saveRequest.Header.Set("Content-Type", "application/json")
	saveRequest.AddCookie(cookies[0])
	saveResponse := httptest.NewRecorder()
	handler.ServeHTTP(saveResponse, saveRequest)
	if saveResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from admin task save, got %d", saveResponse.Code)
	}
	if store.lastTaskDefinition.TaskID != "daily-click-2" {
		t.Fatalf("expected saved task id to be captured, got %+v", store.lastTaskDefinition)
	}

	activateRequest := httptest.NewRequest(http.MethodPost, "/api/admin/tasks/daily-click-2/activate", nil)
	activateRequest.AddCookie(cookies[0])
	activateResponse := httptest.NewRecorder()
	handler.ServeHTTP(activateResponse, activateRequest)
	if activateResponse.Code != http.StatusOK || store.lastTaskDefinition.Status != core.TaskStatusActive {
		t.Fatalf("expected activate to succeed, code=%d task=%+v", activateResponse.Code, store.lastTaskDefinition)
	}

	duplicateRequest := httptest.NewRequest(http.MethodPost, "/api/admin/tasks/daily-click-2/duplicate", strings.NewReader(`{"taskId":"daily-click-3"}`))
	duplicateRequest.Header.Set("Content-Type", "application/json")
	duplicateRequest.AddCookie(cookies[0])
	duplicateResponse := httptest.NewRecorder()
	handler.ServeHTTP(duplicateResponse, duplicateRequest)
	if duplicateResponse.Code != http.StatusOK {
		t.Fatalf("expected duplicate to succeed, got %d", duplicateResponse.Code)
	}

	archiveRequest := httptest.NewRequest(http.MethodPost, "/api/admin/tasks/archive-expired", nil)
	archiveRequest.AddCookie(cookies[0])
	archiveResponse := httptest.NewRecorder()
	handler.ServeHTTP(archiveResponse, archiveRequest)
	if archiveResponse.Code != http.StatusOK || store.lastArchiveNow.IsZero() {
		t.Fatalf("expected archive-expired to succeed, code=%d archivedAt=%v", archiveResponse.Code, store.lastArchiveNow)
	}

	cyclesRequest := httptest.NewRequest(http.MethodGet, "/api/admin/tasks/daily-click-1/cycles", nil)
	cyclesRequest.AddCookie(cookies[0])
	cyclesResponse := httptest.NewRecorder()
	handler.ServeHTTP(cyclesResponse, cyclesRequest)
	if cyclesResponse.Code != http.StatusOK {
		t.Fatalf("expected cycles list to succeed, got %d", cyclesResponse.Code)
	}

	resultsRequest := httptest.NewRequest(http.MethodGet, "/api/admin/tasks/daily-click-1/cycles/2026-05-01/results", nil)
	resultsRequest.AddCookie(cookies[0])
	resultsResponse := httptest.NewRecorder()
	handler.ServeHTTP(resultsResponse, resultsRequest)
	if resultsResponse.Code != http.StatusOK {
		t.Fatalf("expected cycle results to succeed, got %d", resultsResponse.Code)
	}
}

func TestAdminRoomsReturnsDisplayNames(t *testing.T) {
	store := &mockStore{
		roomList: core.RoomList{
			Rooms: []core.RoomInfo{
				{ID: "1", DisplayName: "房间 1", CurrentBossName: "一线 Boss", CycleEnabled: true},
				{ID: "2", DisplayName: "高压线", CurrentBossName: "二线 Boss", CycleEnabled: false},
			},
		},
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "room-secret",
		}),
	})

	unauthorizedRequest := httptest.NewRequest(http.MethodGet, "/api/admin/rooms", nil)
	unauthorizedResponse := httptest.NewRecorder()
	handler.ServeHTTP(unauthorizedResponse, unauthorizedRequest)
	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without admin cookie, got %d", unauthorizedResponse.Code)
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/api/admin/rooms", nil)
	listRequest.AddCookie(mustAdminLoginCookie(t, handler))
	listResponse := httptest.NewRecorder()
	handler.ServeHTTP(listResponse, listRequest)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from admin room list, got %d", listResponse.Code)
	}
	if store.lastListRoomsNickname != "" {
		t.Fatalf("expected admin room list to use empty nickname, got %q", store.lastListRoomsNickname)
	}
	if !strings.Contains(listResponse.Body.String(), `"id":"1"`) ||
		!strings.Contains(listResponse.Body.String(), `"displayName":"高压线"`) ||
		!strings.Contains(listResponse.Body.String(), `"currentBossName":"一线 Boss"`) ||
		!strings.Contains(listResponse.Body.String(), `"cycleEnabled":true`) {
		t.Fatalf("expected admin room list to expose room runtime fields, got %s", listResponse.Body.String())
	}
}

func TestAdminRoomUpdateSavesDisplayName(t *testing.T) {
	store := &mockStore{
		roomList: core.RoomList{
			Rooms: []core.RoomInfo{
				{ID: "1", DisplayName: "房间 1"},
				{ID: "2", DisplayName: "房间 2"},
			},
		},
	}
	publisher := &mockChangePublisher{}
	handler := NewHandler(Options{
		Store:           store,
		Broadcaster:     &mockBroadcaster{},
		ChangePublisher: publisher,
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "room-secret",
		}),
	})

	updateRequest := httptest.NewRequest(http.MethodPut, "/api/admin/rooms/2", strings.NewReader(`{"displayName":"高压线"}`))
	updateRequest.Header.Set("Content-Type", "application/json")
	updateRequest.AddCookie(mustAdminLoginCookie(t, handler))
	updateResponse := httptest.NewRecorder()
	handler.ServeHTTP(updateResponse, updateRequest)
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from admin room update, got %d", updateResponse.Code)
	}
	if got := store.roomDisplayNames["2"]; got != "高压线" {
		t.Fatalf("expected room 2 display name to be saved, got %q", got)
	}
	if !strings.Contains(updateResponse.Body.String(), `"displayName":"高压线"`) {
		t.Fatalf("expected update response to include saved displayName, got %s", updateResponse.Body.String())
	}
	if len(publisher.changes) == 0 {
		t.Fatal("expected room rename to publish change")
	}
}

func TestAdminRoomUpdateClearsDisplayNameWhenEmpty(t *testing.T) {
	store := &mockStore{
		roomList: core.RoomList{
			Rooms: []core.RoomInfo{
				{ID: "1", DisplayName: "房间 1"},
				{ID: "2", DisplayName: "高压线"},
			},
		},
		roomDisplayNames: map[string]string{"2": "高压线"},
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "room-secret",
		}),
	})

	updateRequest := httptest.NewRequest(http.MethodPut, "/api/admin/rooms/2", strings.NewReader(`{"displayName":""}`))
	updateRequest.Header.Set("Content-Type", "application/json")
	updateRequest.AddCookie(mustAdminLoginCookie(t, handler))
	updateResponse := httptest.NewRecorder()
	handler.ServeHTTP(updateResponse, updateRequest)
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from admin room clear, got %d", updateResponse.Code)
	}
	if _, ok := store.roomDisplayNames["2"]; ok {
		t.Fatal("expected custom display name to be cleared")
	}
	if !strings.Contains(updateResponse.Body.String(), `"displayName":"房间 2"`) {
		t.Fatalf("expected cleared response to fall back to default display name, got %s", updateResponse.Body.String())
	}
}

func TestAdminRoomUpdateRejectsUnknownRoom(t *testing.T) {
	store := &mockStore{
		roomList: core.RoomList{
			Rooms: []core.RoomInfo{
				{ID: "1", DisplayName: "房间 1"},
				{ID: "2", DisplayName: "房间 2"},
			},
		},
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "room-secret",
		}),
	})

	updateRequest := httptest.NewRequest(http.MethodPut, "/api/admin/rooms/99", strings.NewReader(`{"displayName":"未知房间"}`))
	updateRequest.Header.Set("Content-Type", "application/json")
	updateRequest.AddCookie(mustAdminLoginCookie(t, handler))
	updateResponse := httptest.NewRecorder()
	handler.ServeHTTP(updateResponse, updateRequest)
	if updateResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 from unknown room update, got %d", updateResponse.Code)
	}
	if !strings.Contains(updateResponse.Body.String(), "ROOM_NOT_FOUND") {
		t.Fatalf("expected room not found error, got %s", updateResponse.Body.String())
	}
}

func mustAdminLoginCookie(t *testing.T, handler http.Handler) *http.Cookie {
	t.Helper()

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{"username":"admin","password":"secret"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	handler.ServeHTTP(loginResponse, loginRequest)
	cookies := loginResponse.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected admin login to set cookie")
	}
	return cookies[0]
}

func (m *mockStore) ClickButton(_ context.Context, slug string, nickname string, comboCount int64) (core.ClickResult, error) {
	m.lastClickNickname = nickname
	if m.clickErr != nil {
		return core.ClickResult{}, m.clickErr
	}
	if m.result.Delta == 0 && m.result.UserStats.Nickname == "" {
		m.result.Delta = 1
		m.result.UserStats = core.UserStats{Nickname: nickname, ClickCount: 1}
	}
	return m.result, nil
}

func (m *mockStore) AutoClickBossPart(_ context.Context, slug string, nickname string) (core.ClickResult, error) {
	m.lastAutoClickNickname = nickname
	return m.ClickButton(context.Background(), slug, nickname, 0)
}

func (m *mockStore) ClickBossPart(_ context.Context, slug string, nickname string) (core.ClickResult, error) {
	return m.ClickButton(context.Background(), slug, nickname, 0)
}

func (m *mockStore) AttackBossPartAFK(_ context.Context, nickname string) (core.ClickResult, error) {
	m.lastAutoClickNickname = nickname
	return core.ClickResult{
		Boss: &core.Boss{
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

func (m *mockStore) EquipItem(_ context.Context, nickname string, _ string) (core.State, error) {
	m.lastClickNickname = nickname
	if m.equipErr != nil {
		return core.State{}, m.equipErr
	}
	if m.equipState.Loadout.Weapon == nil {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) UnequipItem(_ context.Context, _ string, _ string) (core.State, error) {
	if m.equipErr != nil {
		return core.State{}, m.equipErr
	}
	if m.equipState.Loadout.Weapon == nil {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) EnhanceItem(_ context.Context, _ string, _ string) (core.State, error) {
	if m.enhanceErr != nil {
		return core.State{}, m.enhanceErr
	}
	if m.equipState.Loadout.Weapon == nil {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) SalvageItem(_ context.Context, _ string, itemID string) (core.SalvageResult, error) {
	m.lastSalvageItemID = itemID
	if m.salvageErr != nil {
		return core.SalvageResult{}, m.salvageErr
	}
	return core.SalvageResult{
		ItemID:         itemID,
		GoldReward:     500,
		StoneReward:    1,
		RefundedStones: 12,
		Gold:           66,
		Stones:         34,
	}, nil
}

func (m *mockStore) BulkSalvageUnequipped(_ context.Context, _ string) (core.BulkSalvageResult, error) {
	if m.salvageErr != nil {
		return core.BulkSalvageResult{}, m.salvageErr
	}
	return core.BulkSalvageResult{
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

func (m *mockStore) LockItem(_ context.Context, _ string, itemID string) (core.State, error) {
	m.lastLockItemID = itemID
	m.lastLockState = true
	if m.equipErr != nil {
		return core.State{}, m.equipErr
	}
	if m.equipState.Loadout.Weapon == nil {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) UnlockItem(_ context.Context, _ string, itemID string) (core.State, error) {
	m.lastLockItemID = itemID
	m.lastLockState = false
	if m.equipErr != nil {
		return core.State{}, m.equipErr
	}
	if m.equipState.Loadout.Weapon == nil {
		return m.state, nil
	}
	return m.equipState, nil
}

func (m *mockStore) GetAdminState(_ context.Context) (core.AdminState, error) {
	return m.adminState, nil
}

func (m *mockStore) ListAdminEquipmentPage(_ context.Context, _ int64, _ int64) (core.AdminEquipmentPage, error) {
	return m.adminEquipmentPage, nil
}

func (m *mockStore) ListAdminBossHistoryPage(_ context.Context, _ int64, _ int64) (core.AdminBossHistoryPage, error) {
	return m.adminBossHistoryPage, nil
}

func (m *mockStore) ListAdminPlayers(_ context.Context, _ string, _ int64) (core.AdminPlayerPage, error) {
	return m.adminPlayerPage, nil
}

func (m *mockStore) GetAdminPlayer(_ context.Context, _ string) (*core.AdminPlayerOverview, error) {
	return m.adminPlayer, nil
}

func (m *mockStore) SaveEquipmentDefinition(_ context.Context, definition core.EquipmentDefinition) error {
	m.lastEquipment = definition
	return nil
}

func (m *mockStore) DeleteEquipmentDefinition(_ context.Context, _ string) error {
	return nil
}

func (m *mockStore) ActivateBoss(_ context.Context, boss core.BossUpsert) (*core.Boss, error) {
	if m.activateBossErr != nil {
		return nil, m.activateBossErr
	}
	m.lastBoss = boss
	return &core.Boss{
		ID:        boss.ID,
		Name:      boss.Name,
		Status:    "active",
		MaxHP:     boss.MaxHP,
		CurrentHP: boss.MaxHP,
	}, nil
}

func (m *mockStore) ActivateBossInRoom(ctx context.Context, roomID string, boss core.BossUpsert) (*core.Boss, error) {
	m.lastBossRoomID = roomID
	boss.RoomID = roomID
	return m.ActivateBoss(ctx, boss)
}

func (m *mockStore) DeactivateBoss(_ context.Context) error {
	return nil
}

func (m *mockStore) DeactivateBossInRoom(_ context.Context, roomID string) error {
	m.lastDeactivateRoomID = roomID
	return nil
}

func (m *mockStore) SetBossLoot(_ context.Context, _ string, _ []core.BossLootEntry) error {
	return nil
}

func (m *mockStore) SaveBossTemplate(_ context.Context, template core.BossTemplateUpsert) error {
	if m.saveBossTemplateErr != nil {
		return m.saveBossTemplateErr
	}
	m.lastBossTemplate = template
	return nil
}

func (m *mockStore) DeleteBossTemplate(_ context.Context, _ string) error {
	return nil
}

func (m *mockStore) SetBossTemplateLoot(_ context.Context, templateID string, loot []core.BossLootEntry) error {
	m.lastTemplateLootID = templateID
	m.lastTemplateLoot = loot
	return nil
}

func (m *mockStore) SetBossCycleQueue(_ context.Context, templateIDs []string) ([]string, error) {
	m.lastCycleQueue = append([]string(nil), templateIDs...)
	return append([]string(nil), templateIDs...), nil
}

func (m *mockStore) SetBossCycleQueueForRoom(ctx context.Context, roomID string, templateIDs []string) ([]string, error) {
	m.lastCycleRoomID = roomID
	return m.SetBossCycleQueue(ctx, templateIDs)
}

func (m *mockStore) SetBossCycleEnabled(_ context.Context, enabled bool) (*core.Boss, error) {
	if m.setBossCycleErr != nil {
		return nil, m.setBossCycleErr
	}
	m.lastCycleEnabled = enabled
	if !enabled {
		return nil, nil
	}
	return &core.Boss{
		ID:         "dragon-1",
		TemplateID: "dragon",
		Name:       "火龙",
		Status:     "active",
		MaxHP:      80,
		CurrentHP:  80,
	}, nil
}

func (m *mockStore) SetBossCycleEnabledForRoom(ctx context.Context, roomID string, enabled bool) (*core.Boss, error) {
	m.lastCycleRoomID = roomID
	boss, err := m.SetBossCycleEnabled(ctx, enabled)
	if boss != nil {
		boss.RoomID = roomID
	}
	return boss, err
}

func (m *mockStore) ListBossHistory(_ context.Context) ([]core.BossHistoryEntry, error) {
	return m.bossHistory, nil
}

func (m *mockStore) GetLatestAnnouncement(_ context.Context) (*core.Announcement, error) {
	return m.latestAnnouncement, nil
}

func (m *mockStore) ListAnnouncements(_ context.Context, includeInactive bool) ([]core.Announcement, error) {
	return m.announcements, nil
}

func (m *mockStore) SaveAnnouncement(_ context.Context, announcement core.AnnouncementUpsert) (*core.Announcement, error) {
	return &core.Announcement{
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

func (m *mockStore) CreateMessage(_ context.Context, nickname string, content string) (*core.Message, error) {
	if m.messageErr != nil {
		return nil, m.messageErr
	}
	return &core.Message{
		ID:        "1",
		Nickname:  nickname,
		Content:   content,
		CreatedAt: 1710000000,
	}, nil
}

func (m *mockStore) ListMessages(_ context.Context, _ string, _ int64) (core.MessagePage, error) {
	return m.messagePage, m.messageErr
}

func (m *mockStore) DeleteMessage(_ context.Context, _ string) error {
	return nil
}

func (m *mockStore) GetTalentState(_ context.Context, _ string) (*core.TalentState, error) {
	if m.talentState != nil {
		return m.talentState, nil
	}
	return &core.TalentState{Talents: map[string]int{}}, nil
}

func (m *mockStore) UpgradeTalent(_ context.Context, _ string, _ string, _ int) error {
	if m.talentState == nil {
		m.talentState = &core.TalentState{Talents: map[string]int{}}
	}
	if m.talentState.Talents == nil {
		m.talentState.Talents = map[string]int{}
	}
	m.talentState.Talents["normal_core"] = 2
	return nil
}

func (m *mockStore) ResetTalents(_ context.Context, _ string) error {
	return nil
}

func (m *mockStore) ComputeTalentModifiers(_ context.Context, _ string) (*core.TalentModifiers, error) {
	return nil, nil
}

type mockOSSSigner struct {
	policy ossupload.Policy
	err    error
}

func (m *mockOSSSigner) CreatePolicy(_ context.Context, _ string) (ossupload.Policy, error) {
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
	snapshots []core.Snapshot
}

func (m *mockBroadcaster) BroadcastSnapshot(snapshot core.Snapshot) error {
	m.snapshots = append(m.snapshots, snapshot)
	return nil
}

type mockChangePublisher struct {
	changes []core.StateChange
}

func (m *mockChangePublisher) PublishChange(_ context.Context, change core.StateChange) error {
	m.changes = append(m.changes, change)
	return nil
}

func TestGetBossHistoryReturnsPublicHistory(t *testing.T) {
	store := &mockStore{
		bossHistory: []core.BossHistoryEntry{
			{
				Boss: core.Boss{
					ID:         "slime-king",
					Name:       "史莱姆王",
					Status:     "defeated",
					MaxHP:      100,
					CurrentHP:  0,
					StartedAt:  1710000000,
					DefeatedAt: 1710000300,
				},
				Loot: []core.BossLootEntry{
					{ItemID: "cloth-armor", ItemName: "布甲", Weight: 3},
				},
				Damage: []core.BossLeaderboardEntry{
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

	var payload []core.BossHistoryEntry
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

func TestGetBossResourcesUsesReadNicknameRoom(t *testing.T) {
	store := &mockStore{
		bossResources: core.BossResources{
			BossID: "room-1-boss",
			RoomID: "1",
		},
		bossResourcesByNickname: map[string]core.BossResources{
			"阿明": {
				BossID: "room-2-boss",
				RoomID: "2",
			},
		},
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodGet, "/api/boss/resources?nickname=%E9%98%BF%E6%98%8E", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if store.lastBossResourcesNickname != "阿明" {
		t.Fatalf("expected resources to use read nickname, got %q", store.lastBossResourcesNickname)
	}

	var payload core.BossResources
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.RoomID != "2" || payload.BossID != "room-2-boss" {
		t.Fatalf("expected room 2 resources, got %+v", payload)
	}
}

func TestJoinRoomRejectsNotJoinableRoom(t *testing.T) {
	store := &mockStore{
		roomSwitchErr: core.ErrRoomNotJoinable,
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/rooms/join", strings.NewReader(`{"nickname":"阿明","roomId":"2"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
	if store.lastSwitchRoomNickname != "阿明" || store.lastSwitchRoomID != "2" {
		t.Fatalf("expected join request to reach store, got nickname=%q room=%q", store.lastSwitchRoomNickname, store.lastSwitchRoomID)
	}
	if !strings.Contains(response.Body.String(), "ROOM_NOT_JOINABLE") {
		t.Fatalf("expected ROOM_NOT_JOINABLE response, got %s", response.Body.String())
	}
}

func TestGetRoomsReturnsDisplayName(t *testing.T) {
	store := &mockStore{
		roomList: core.RoomList{
			CurrentRoomID: "1",
			Rooms: []core.RoomInfo{
				{ID: "1", DisplayName: "高压线", Joinable: true},
			},
		},
	}
	handler := NewHandler(Options{
		Store:       store,
		Broadcaster: &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodGet, "/api/rooms?nickname=%E9%98%BF%E6%98%8E", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"displayName":"高压线"`) {
		t.Fatalf("expected room list to include displayName, got %s", response.Body.String())
	}
}

func TestTalentStateReturnsBackendEffectLines(t *testing.T) {
	store := &mockStore{
		state: core.State{
			TalentPoints: 345,
		},
		talentState: &core.TalentState{
			Talents: map[string]int{
				"normal_core": 2,
			},
		},
	}
	handler := NewHandler(Options{
		Store:               store,
		Broadcaster:         &mockBroadcaster{},
		PlayerAuthenticator: &mockPlayerAuthenticator{verifyNickname: "阿明"},
	})

	request := httptest.NewRequest(http.MethodGet, "/api/talents/state", nil)
	request.AddCookie(&http.Cookie{Name: playerSessionCookieName, Value: "player-token"})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload map[string]any
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	effectLines, ok := payload["effectLines"].(map[string]any)
	if !ok {
		t.Fatalf("expected effectLines map in payload, got %+v", payload["effectLines"])
	}
	normalCore, ok := effectLines["normal_core"].([]any)
	if !ok || len(normalCore) != 3 {
		t.Fatalf("expected normal_core effect lines, got %+v", effectLines["normal_core"])
	}
	firstLine, ok := normalCore[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first line object, got %+v", normalCore[0])
	}
	if firstLine["label"] != "触发次数" || firstLine["text"] != "55 → 50" {
		t.Fatalf("expected trigger count preview from backend, got %+v", firstLine)
	}

	effectDescriptions, ok := payload["effectDescriptions"].(map[string]any)
	if !ok {
		t.Fatalf("expected effectDescriptions map in payload, got %+v", payload["effectDescriptions"])
	}
	if effectDescriptions["normal_core"] != "每 55 次点击触发追击爆发，造成 基础伤害 x 68% x 18 段总伤。可无限触发。" {
		t.Fatalf("expected dynamic normal_core description, got %+v", effectDescriptions["normal_core"])
	}
}

func TestTalentUpgradeReturnsUpdatedEffectLines(t *testing.T) {
	store := &mockStore{
		state: core.State{
			TalentPoints: 338,
		},
		talentState: &core.TalentState{
			Talents: map[string]int{
				"normal_core": 1,
			},
		},
	}
	handler := NewHandler(Options{
		Store:               store,
		Broadcaster:         &mockBroadcaster{},
		PlayerAuthenticator: &mockPlayerAuthenticator{verifyNickname: "阿明"},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/talents/upgrade", strings.NewReader(`{"talentId":"normal_core","targetLevel":2}`))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(&http.Cookie{Name: playerSessionCookieName, Value: "player-token"})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload map[string]any
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	effectLines, ok := payload["effectLines"].(map[string]any)
	if !ok {
		t.Fatalf("expected effectLines map in payload, got %+v", payload["effectLines"])
	}
	normalCore, ok := effectLines["normal_core"].([]any)
	if !ok || len(normalCore) != 3 {
		t.Fatalf("expected normal_core effect lines, got %+v", effectLines["normal_core"])
	}
	secondLine, ok := normalCore[1].(map[string]any)
	if !ok {
		t.Fatalf("expected second line object, got %+v", normalCore[1])
	}
	if secondLine["label"] != "追击段数" || secondLine["text"] != "18 → 22" {
		t.Fatalf("expected updated preview from backend, got %+v", secondLine)
	}
	firstLine, ok := normalCore[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first line object, got %+v", normalCore[0])
	}
	if firstLine["label"] != "触发次数" || firstLine["text"] != "55 → 50" {
		t.Fatalf("expected updated trigger count preview from backend, got %+v", firstLine)
	}

	effectDescriptions, ok := payload["effectDescriptions"].(map[string]any)
	if !ok {
		t.Fatalf("expected effectDescriptions map in payload, got %+v", payload["effectDescriptions"])
	}
	if effectDescriptions["normal_core"] != "每 55 次点击触发追击爆发，造成 基础伤害 x 68% x 18 段总伤。可无限触发。" {
		t.Fatalf("expected upgraded dynamic normal_core description, got %+v", effectDescriptions["normal_core"])
	}
}

func TestGetLatestAnnouncementReturnsPayload(t *testing.T) {
	store := &mockStore{
		latestAnnouncement: &core.Announcement{
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

	var payload core.Announcement
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.ID != "7" || payload.Title != "更新公告" {
		t.Fatalf("unexpected latest announcement payload: %+v", payload)
	}
}
func TestEquipItemReturnsUpdatedState(t *testing.T) {
	store := &mockStore{
		equipState: core.State{

			Loadout: core.Loadout{
				Weapon: &core.InventoryItem{
					ItemID:   "wood-sword",
					Name:     "木剑",
					Slot:     "weapon",
					Quantity: 1,
					Equipped: true,
				},
			},
			CombatStats: core.CombatStats{
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
			Weapon *core.InventoryItem `json:"weapon"`
		} `json:"loadout"`
		CombatStats core.CombatStats `json:"combatStats"`
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

	var payload core.BulkSalvageResult
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
		messageErr: core.ErrSensitiveContent,
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

func TestPublicMessagesPreferOptionalMessageStore(t *testing.T) {
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
		Broadcaster:  &mockBroadcaster{},
	})

	request := httptest.NewRequest(http.MethodGet, "/api/messages?cursor=99", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload core.MessagePage
	if err := sonic.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Items) != 1 || payload.Items[0].ID != "mongo-1" {
		t.Fatalf("expected optional message store payload, got %+v", payload)
	}
}

func TestAdminLoginCreatesSessionAndStateRequiresAuth(t *testing.T) {
	store := &mockStore{
		adminState: core.AdminState{
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

	saveRoomQueueRequest := httptest.NewRequest(http.MethodPut, "/api/admin/boss/cycle/queue?roomId=2", strings.NewReader(`{"roomId":"2","templateIds":["dragon"]}`))
	saveRoomQueueRequest.Header.Set("Content-Type", "application/json")
	saveRoomQueueRequest.AddCookie(cookies[0])
	saveRoomQueueResponse := httptest.NewRecorder()
	handler.ServeHTTP(saveRoomQueueResponse, saveRoomQueueRequest)
	if saveRoomQueueResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from room cycle queue save, got %d", saveRoomQueueResponse.Code)
	}
	if store.lastCycleRoomID != "2" || len(store.lastCycleQueue) != 1 || store.lastCycleQueue[0] != "dragon" {
		t.Fatalf("expected room cycle queue to be forwarded, room=%q queue=%+v", store.lastCycleRoomID, store.lastCycleQueue)
	}

	activateRoomBossRequest := httptest.NewRequest(http.MethodPost, "/api/admin/boss/activate?roomId=2", strings.NewReader(`{"id":"room-boss","name":"二线 Boss","maxHp":80,"parts":[{"x":0,"y":0,"type":"soft","maxHp":80}]}`))
	activateRoomBossRequest.Header.Set("Content-Type", "application/json")
	activateRoomBossRequest.AddCookie(cookies[0])
	activateRoomBossResponse := httptest.NewRecorder()
	handler.ServeHTTP(activateRoomBossResponse, activateRoomBossRequest)
	if activateRoomBossResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from room boss activate, got %d", activateRoomBossResponse.Code)
	}
	if store.lastBossRoomID != "2" || store.lastBoss.RoomID != "2" || store.lastBoss.ID != "room-boss" {
		t.Fatalf("expected room boss activate to be forwarded, room=%q boss=%+v", store.lastBossRoomID, store.lastBoss)
	}
}

func TestAdminBossPoolRoutesAcceptStringInt64HP(t *testing.T) {
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

	request := httptest.NewRequest(http.MethodPost, "/api/admin/boss/pool", strings.NewReader(`{"id":"dragon","name":"火龙","maxHp":"9223372036854775800","layout":[{"x":0,"y":0,"type":"soft","maxHp":"9223372036854775800","currentHp":"9223372036854775800","armor":"0"}]}`))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(cookies[0])
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200 from boss template save with string int64 hp, got %d body=%s", response.Code, response.Body.String())
	}
	if store.lastBossTemplate.MaxHP != 9223372036854775800 {
		t.Fatalf("expected template maxHp preserved, got %+v", store.lastBossTemplate)
	}
	if len(store.lastBossTemplate.Layout) != 1 || store.lastBossTemplate.Layout[0].MaxHP != 9223372036854775800 {
		t.Fatalf("expected layout maxHp preserved, got %+v", store.lastBossTemplate.Layout)
	}
}

func TestAdminBossPartsRequiredReturnsBadRequest(t *testing.T) {
	store := &mockStore{
		activateBossErr:     core.ErrBossPartsRequired,
		saveBossTemplateErr: core.ErrBossPartsRequired,
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
		validateErr: core.ErrSensitiveNickname,
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
