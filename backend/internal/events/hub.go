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

	"long/internal/vote"
)

const (
	PublicStateEventName = "public_state"
	UserStateEventName   = "user_state"
	OnlineCountEventName = "online_count"
)

type onlineCountPayload struct {
	Count int `json:"count"`
}

type realtimeUserStatePayload struct {
	UserStats     *vote.UserStats     `json:"userStats,omitempty"`
	MyBossStats   *vote.BossUserStats `json:"myBossStats,omitempty"`
	Loadout       vote.Loadout        `json:"loadout"`
	CombatStats   vote.CombatStats    `json:"combatStats"`
	Gold          int64               `json:"gold"`
	Stones        int64               `json:"stones"`
	TalentPoints  int64               `json:"talentPoints"`
	RecentRewards []vote.Reward       `json:"recentRewards,omitempty"`
}

type publicStatePayload struct {
	TotalVotes          int64                       `json:"totalVotes"`
	Leaderboard         *[]vote.LeaderboardEntry    `json:"leaderboard,omitempty"`
	Boss                *vote.Boss                  `json:"boss,omitempty"`
	BossLeaderboard     []vote.BossLeaderboardEntry `json:"bossLeaderboard"`
	AnnouncementVersion string                      `json:"announcementVersion,omitempty"`
}

// StateReader 提供 SSE 初始状态所需的公共态与个人态读取能力。
type StateReader interface {
	GetSnapshot(context.Context) (vote.Snapshot, error)
	GetUserState(context.Context, string) (vote.UserState, error)
	GetBossResources(context.Context) (vote.BossResources, error)
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

// Hub 按事件类型向浏览器广播公共态和个人态。
type Hub struct {
	mu      sync.RWMutex
	clients map[*subscriber]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*subscriber]struct{})}
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

func (h *Hub) BroadcastPublic(snapshot vote.Snapshot, includeLeaderboard bool) error {
	payload, err := sonic.Marshal(buildPublicStatePayload(snapshot, includeLeaderboard))
	if err != nil {
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		deliverEvent(client.ch, ServerEvent{Name: PublicStateEventName, Payload: payload})
	}

	return nil
}

func buildPublicStatePayload(snapshot vote.Snapshot, includeLeaderboard bool) publicStatePayload {
	payload := publicStatePayload{
		TotalVotes:          snapshot.TotalVotes,
		Boss:                snapshot.Boss,
		BossLeaderboard:     snapshot.BossLeaderboard,
		AnnouncementVersion: snapshot.AnnouncementVersion,
	}
	if payload.BossLeaderboard == nil {
		payload.BossLeaderboard = []vote.BossLeaderboardEntry{}
	}
	if includeLeaderboard {
		leaderboard := snapshot.Leaderboard
		if leaderboard == nil {
			leaderboard = []vote.LeaderboardEntry{}
		}
		payload.Leaderboard = &leaderboard
	}
	return payload
}

func (h *Hub) BroadcastUser(nickname string, state vote.UserState) error {
	normalizedNickname := strings.TrimSpace(nickname)
	if normalizedNickname == "" {
		return nil
	}

	payload, err := sonic.Marshal(buildRealtimeUserStatePayload(state))
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

func buildRealtimeUserStatePayload(state vote.UserState) realtimeUserStatePayload {
	return realtimeUserStatePayload{
		UserStats:     state.UserStats,
		MyBossStats:   state.MyBossStats,
		Loadout:       state.Loadout,
		CombatStats:   state.CombatStats,
		Gold:          state.Gold,
		Stones:        state.Stones,
		TalentPoints:  state.TalentPoints,
		RecentRewards: state.RecentRewards,
	}
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
		if err != nil {
			c.JSON(consts.StatusInternalServerError, map[string]string{"error": "STATE_FETCH_FAILED"})
			return
		}

		writer := sse.NewWriter(c)
		defer writer.Close()

		if err := writeEvent(writer, PublicStateEventName, snapshot); err != nil {
			return
		}

		if nickname != "" {
			userState, err := reader.GetUserState(ctx, nickname)
			if err != nil {
				c.JSON(consts.StatusInternalServerError, map[string]string{"error": "STATE_FETCH_FAILED"})
				return
			}
			if err := writeEvent(writer, UserStateEventName, buildRealtimeUserStatePayload(userState)); err != nil {
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
