package events

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/protocol/sse"

	"long/internal/core"
)

const (
	PublicStateEventName = "public_state"
	PublicMetaEventName  = "public_meta"
	UserStateEventName   = "user_state"
	OnlineCountEventName = "online_count"
	RoomStateEventName   = "room_state"
)

type onlineCountPayload struct {
	Count int `json:"count"`
}

type realtimeUserStatePayload struct {
	UserStats                          *core.UserStats           `json:"userStats,omitempty"`
	MyBossStats                        *core.BossUserStats       `json:"myBossStats,omitempty"`
	MyBossKills                        int64                     `json:"myBossKills"`
	TotalBossKills                     int64                     `json:"totalBossKills"`
	RoomID                             string                    `json:"roomId,omitempty"`
	Loadout                            *core.Loadout             `json:"loadout,omitempty"`
	CombatStats                        *core.CombatStats         `json:"combatStats,omitempty"`
	Gold                               int64                     `json:"gold"`
	Stones                             int64                     `json:"stones"`
	TalentPoints                       int64                     `json:"talentPoints"`
	RecentRewards                      []core.Reward             `json:"recentRewards,omitempty"`
	TalentEvents                       []core.TalentTriggerEvent `json:"talentEvents,omitempty"`
	TalentCombatState                  *core.TalentCombatState   `json:"talentCombatState,omitempty"`
	EquippedBattleClickSkinID          string                    `json:"equippedBattleClickSkinId,omitempty"`
	EquippedBattleClickCursorImagePath string                    `json:"equippedBattleClickCursorImagePath,omitempty"`
}

type publicStatePayload struct {
	TotalVotes  int64               `json:"totalVotes"`
	RoomID      string              `json:"roomId,omitempty"`
	Boss        *core.Boss          `json:"boss,omitempty"`
	BossID      string              `json:"bossId,omitempty"`
	BossVersion int64               `json:"bossVersion,omitempty"`
	BossStatic  *bossStaticPayload  `json:"bossStatic,omitempty"`
	BossRuntime *bossRuntimePayload `json:"bossRuntime,omitempty"`
}

type publicMetaPayload struct {
	Leaderboard         *[]core.LeaderboardEntry    `json:"leaderboard,omitempty"`
	BossLeaderboard     []core.BossLeaderboardEntry `json:"bossLeaderboard"`
	AnnouncementVersion string                      `json:"announcementVersion,omitempty"`
}

type roomStatePayload = core.RoomList

// StateReader 提供 SSE 初始状态所需的公共态与个人态读取能力。
type StateReader interface {
	GetSnapshot(context.Context) (core.Snapshot, error)
	GetUserState(context.Context, string) (core.UserState, error)
	GetBossResources(context.Context) (core.BossResources, error)
	ListRooms(context.Context, string) (core.RoomList, error)
}

// ServerEvent 是发往浏览器的一条 SSE 事件。
type ServerEvent struct {
	Name    string
	Payload []byte
}

type subscriber struct {
	nickname string
	ch       chan ServerEvent
}

type bossStaticPayload struct {
	ID                 string                  `json:"id,omitempty"`
	TemplateID         string                  `json:"templateId,omitempty"`
	RoomID             string                  `json:"roomId,omitempty"`
	QueueID            string                  `json:"queueId,omitempty"`
	Name               string                  `json:"name,omitempty"`
	MaxHP              int64                   `json:"maxHp,omitempty"`
	GoldOnKill         int64                   `json:"goldOnKill,omitempty"`
	StoneOnKill        int64                   `json:"stoneOnKill,omitempty"`
	TalentPointsOnKill int64                   `json:"talentPointsOnKill,omitempty"`
	Parts              []bossPartStaticPayload `json:"parts,omitempty"`
	StartedAt          int64                   `json:"startedAt,omitempty"`
}

type bossPartStaticPayload struct {
	X           int    `json:"x"`
	Y           int    `json:"y"`
	Type        string `json:"type,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	ImagePath   string `json:"imagePath,omitempty"`
	MaxHP       int64  `json:"maxHp,omitempty"`
	Armor       int64  `json:"armor,omitempty"`
}

type bossRuntimePayload struct {
	Status     string                   `json:"status,omitempty"`
	CurrentHP  int64                    `json:"currentHp,omitempty"`
	Parts      []bossPartRuntimePayload `json:"parts,omitempty"`
	DefeatedAt int64                    `json:"defeatedAt,omitempty"`
}

type bossPartRuntimePayload struct {
	X         int   `json:"x"`
	Y         int   `json:"y"`
	CurrentHP int64 `json:"currentHp,omitempty"`
	Alive     bool  `json:"alive"`
}

type bossVersionState struct {
	signature string
	version   int64
}

// Hub 按事件类型向浏览器广播公共态和个人态。
type Hub struct {
	mu           sync.RWMutex
	clients      map[*subscriber]struct{}
	bossVersionM sync.RWMutex
	bossVersions map[string]bossVersionState
}

func NewHub() *Hub {
	return &Hub{
		clients:      make(map[*subscriber]struct{}),
		bossVersions: make(map[string]bossVersionState),
	}
}

func (h *Hub) Subscribe(nickname string) (<-chan ServerEvent, func()) {
	client := &subscriber{
		nickname: strings.TrimSpace(nickname),
		ch:       make(chan ServerEvent, 4),
	}

	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.broadcastOnlineCountLocked()
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		if _, ok := h.clients[client]; ok {
			delete(h.clients, client)
			close(client.ch)
			h.broadcastOnlineCountLocked()
		}
		h.mu.Unlock()
	}

	return client.ch, unsubscribe
}

func (h *Hub) BroadcastPublic(snapshot core.Snapshot) error {
	payload, err := sonic.Marshal(h.buildPublicStatePayload(snapshot))
	if err != nil {
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.nickname != "" {
			continue
		}
		deliverEvent(client.ch, ServerEvent{Name: PublicStateEventName, Payload: payload})
	}

	return nil
}

func (h *Hub) BroadcastPublicTo(nickname string, snapshot core.Snapshot) error {
	normalizedNickname := strings.TrimSpace(nickname)
	if normalizedNickname == "" {
		return nil
	}
	payload, err := sonic.Marshal(h.buildPublicStatePayload(snapshot))
	if err != nil {
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.nickname != normalizedNickname {
			continue
		}
		deliverEvent(client.ch, ServerEvent{Name: PublicStateEventName, Payload: payload})
	}

	return nil
}

func (h *Hub) BroadcastPublicMeta(snapshot core.Snapshot, includeLeaderboard bool) error {
	payload, err := sonic.Marshal(buildPublicMetaPayload(snapshot, includeLeaderboard))
	if err != nil {
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.nickname != "" {
			continue
		}
		deliverEvent(client.ch, ServerEvent{Name: PublicMetaEventName, Payload: payload})
	}

	return nil
}

func (h *Hub) BroadcastPublicMetaTo(nickname string, snapshot core.Snapshot, includeLeaderboard bool) error {
	normalizedNickname := strings.TrimSpace(nickname)
	if normalizedNickname == "" {
		return nil
	}
	payload, err := sonic.Marshal(buildPublicMetaPayload(snapshot, includeLeaderboard))
	if err != nil {
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.nickname != normalizedNickname {
			continue
		}
		deliverEvent(client.ch, ServerEvent{Name: PublicMetaEventName, Payload: payload})
	}

	return nil
}

func (h *Hub) buildPublicStatePayload(snapshot core.Snapshot) publicStatePayload {
	payload := publicStatePayload{
		TotalVotes: snapshot.TotalVotes,
		RoomID:     snapshot.RoomID,
		Boss:       snapshot.Boss,
	}
	if snapshot.Boss == nil {
		return payload
	}
	payload.BossID = snapshot.Boss.ID
	payload.BossVersion = h.nextBossVersion(snapshot.TotalVotes, snapshot.Boss)
	payload.BossStatic = buildBossStaticPayload(snapshot.Boss)
	payload.BossRuntime = buildBossRuntimePayload(snapshot.Boss)
	return payload
}

func buildBossStaticPayload(boss *core.Boss) *bossStaticPayload {
	if boss == nil {
		return nil
	}
	parts := make([]bossPartStaticPayload, 0, len(boss.Parts))
	for _, part := range boss.Parts {
		parts = append(parts, bossPartStaticPayload{
			X:           part.X,
			Y:           part.Y,
			Type:        string(part.Type),
			DisplayName: part.DisplayName,
			ImagePath:   part.ImagePath,
			MaxHP:       part.MaxHP,
			Armor:       part.Armor,
		})
	}
	return &bossStaticPayload{
		ID:                 boss.ID,
		TemplateID:         boss.TemplateID,
		RoomID:             boss.RoomID,
		QueueID:            boss.QueueID,
		Name:               boss.Name,
		MaxHP:              boss.MaxHP,
		GoldOnKill:         boss.GoldOnKill,
		StoneOnKill:        boss.StoneOnKill,
		TalentPointsOnKill: boss.TalentPointsOnKill,
		Parts:              parts,
		StartedAt:          boss.StartedAt,
	}
}

func buildBossRuntimePayload(boss *core.Boss) *bossRuntimePayload {
	if boss == nil {
		return nil
	}
	parts := make([]bossPartRuntimePayload, 0, len(boss.Parts))
	for _, part := range boss.Parts {
		parts = append(parts, bossPartRuntimePayload{
			X:         part.X,
			Y:         part.Y,
			CurrentHP: part.CurrentHP,
			Alive:     part.Alive,
		})
	}
	return &bossRuntimePayload{
		Status:     boss.Status,
		CurrentHP:  boss.CurrentHP,
		Parts:      parts,
		DefeatedAt: boss.DefeatedAt,
	}
}

func bossVersionSignature(totalVotes int64, boss *core.Boss) string {
	if boss == nil {
		return ""
	}
	signatureBody, _ := sonic.Marshal(struct {
		TotalVotes int64               `json:"totalVotes"`
		Runtime    *bossRuntimePayload `json:"runtime"`
	}{
		TotalVotes: totalVotes,
		Runtime:    buildBossRuntimePayload(boss),
	})
	return string(signatureBody)
}

func (h *Hub) nextBossVersion(totalVotes int64, boss *core.Boss) int64 {
	if h == nil || boss == nil {
		return 0
	}
	signature := bossVersionSignature(totalVotes, boss)
	h.bossVersionM.Lock()
	defer h.bossVersionM.Unlock()

	current := h.bossVersions[boss.ID]
	if current.signature == signature {
		return current.version
	}
	if current.version <= 0 {
		current.version = 1
	} else {
		current.version++
	}
	current.signature = signature
	h.bossVersions[boss.ID] = current
	return current.version
}

func (h *Hub) CurrentBossVersion(bossID string) int64 {
	if h == nil || strings.TrimSpace(bossID) == "" {
		return 0
	}
	h.bossVersionM.RLock()
	defer h.bossVersionM.RUnlock()
	return h.bossVersions[bossID].version
}

func buildPublicMetaPayload(snapshot core.Snapshot, includeLeaderboard bool) publicMetaPayload {
	payload := publicMetaPayload{
		BossLeaderboard:     snapshot.BossLeaderboard,
		AnnouncementVersion: snapshot.AnnouncementVersion,
	}
	if payload.BossLeaderboard == nil {
		payload.BossLeaderboard = []core.BossLeaderboardEntry{}
	}
	if includeLeaderboard {
		leaderboard := snapshot.Leaderboard
		if leaderboard == nil {
			leaderboard = []core.LeaderboardEntry{}
		}
		payload.Leaderboard = &leaderboard
	}
	return payload
}

func (h *Hub) BroadcastUser(nickname string, state core.UserState, includeProfile bool) error {
	normalizedNickname := strings.TrimSpace(nickname)
	if normalizedNickname == "" {
		return nil
	}

	payload, err := sonic.Marshal(buildRealtimeUserStatePayload(state, includeProfile))
	if err != nil {
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.nickname != normalizedNickname {
			continue
		}
		deliverEvent(client.ch, ServerEvent{Name: UserStateEventName, Payload: payload})
	}

	return nil
}

func (h *Hub) BroadcastRoomState(nickname string, rooms core.RoomList) error {
	normalizedNickname := strings.TrimSpace(nickname)
	payload, err := sonic.Marshal(roomStatePayload(rooms))
	if err != nil {
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.nickname != normalizedNickname {
			continue
		}
		deliverEvent(client.ch, ServerEvent{Name: RoomStateEventName, Payload: payload})
	}

	return nil
}

func buildRealtimeUserStatePayload(state core.UserState, includeProfile bool) realtimeUserStatePayload {
	payload := realtimeUserStatePayload{
		UserStats:                          state.UserStats,
		MyBossStats:                        state.MyBossStats,
		MyBossKills:                        state.MyBossKills,
		TotalBossKills:                     state.TotalBossKills,
		RoomID:                             state.RoomID,
		Gold:                               state.Gold,
		Stones:                             state.Stones,
		TalentPoints:                       state.TalentPoints,
		RecentRewards:                      state.RecentRewards,
		TalentEvents:                       state.TalentEvents,
		TalentCombatState:                  state.TalentCombatState,
		EquippedBattleClickSkinID:          state.EquippedBattleClickSkinID,
		EquippedBattleClickCursorImagePath: state.EquippedBattleClickCursorImagePath,
	}
	if includeProfile {
		loadout := state.Loadout
		combatStats := state.CombatStats
		payload.Loadout = &loadout
		payload.CombatStats = &combatStats
	}
	return payload
}

func (h *Hub) SubscriberCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) ActiveNicknames() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	seen := make(map[string]struct{}, len(h.clients))
	for client := range h.clients {
		if client.nickname == "" {
			continue
		}
		seen[client.nickname] = struct{}{}
	}

	nicknames := make([]string, 0, len(seen))
	for nickname := range seen {
		nicknames = append(nicknames, nickname)
	}
	return nicknames
}

func (h *Hub) broadcastOnlineCountLocked() {
	payload, err := sonic.Marshal(onlineCountPayload{Count: len(h.clients)})
	if err != nil {
		return
	}
	for client := range h.clients {
		deliverEvent(client.ch, ServerEvent{Name: OnlineCountEventName, Payload: payload})
	}
}

// NewHandler 暴露浏览器 EventSource 使用的 Hertz 原生 SSE 入口。
func NewHandler(hub *Hub, reader StateReader, resolveNickname func(context.Context, *app.RequestContext) string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		nickname := strings.TrimSpace(c.Query("nickname"))
		if resolveNickname != nil {
			nickname = strings.TrimSpace(resolveNickname(ctx, c))
		}

		snapshot, err := reader.GetSnapshot(ctx)
		if nickname != "" {
			if roomReader, ok := reader.(interface {
				GetSnapshotForNickname(context.Context, string) (core.Snapshot, error)
			}); ok {
				snapshot, err = roomReader.GetSnapshotForNickname(ctx, nickname)
			}
		}
		if err != nil {
			c.JSON(consts.StatusInternalServerError, map[string]string{"error": "STATE_FETCH_FAILED"})
			return
		}

		writer := sse.NewWriter(c)
		defer writer.Close()

		if err := writeEvent(writer, PublicStateEventName, hub.buildPublicStatePayload(snapshot)); err != nil {
			return
		}
		if err := writeEvent(writer, PublicMetaEventName, buildPublicMetaPayload(snapshot, true)); err != nil {
			return
		}

		if roomState, err := reader.ListRooms(ctx, nickname); err == nil {
			if err := writeEvent(writer, RoomStateEventName, roomState); err != nil {
				return
			}
		}

		if nickname != "" {
			userState, err := reader.GetUserState(ctx, nickname)
			if err != nil {
				c.JSON(consts.StatusInternalServerError, map[string]string{"error": "STATE_FETCH_FAILED"})
				return
			}
			if err := writeEvent(writer, UserStateEventName, buildRealtimeUserStatePayload(userState, true)); err != nil {
				return
			}
		}

		client, unsubscribe := hub.Subscribe(nickname)
		defer unsubscribe()

		heartbeat := time.NewTicker(25 * time.Second)
		defer heartbeat.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-client:
				if !ok {
					return
				}
				if err := writer.WriteEvent("", event.Name, event.Payload); err != nil {
					return
				}
			case <-heartbeat.C:
				if err := writer.WriteComment("ping"); err != nil {
					return
				}
			}
		}
	}
}

func deliverEvent(client chan ServerEvent, event ServerEvent) {
	select {
	case client <- event:
	default:
		select {
		case <-client:
		default:
		}

		select {
		case client <- event:
		default:
		}
	}
}

func writeEvent(writer *sse.Writer, name string, payload any) error {
	encoded, err := sonic.Marshal(payload)
	if err != nil {
		return err
	}
	return writer.WriteEvent("", name, encoded)
}
