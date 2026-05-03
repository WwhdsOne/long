package httpapi

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/websocket"

	"long/internal/core"
	"long/internal/events"
)

const (
	realtimeMessageTypeHello       = "hello"
	realtimeMessageTypeClick       = "click"
	realtimeMessageTypeSyncRequest = "sync_request"
	realtimeMessageTypePing        = "ping"

	realtimeMessageTypeSnapshot    = "snapshot"
	realtimeMessageTypePublicDelta = "public_delta"
	realtimeMessageTypeUserDelta   = "user_delta"
	realtimeMessageTypeOnlineCount = "online_count"
	realtimeMessageTypeClickAck    = "click_ack"
	realtimeMessageTypeError       = "error"
	realtimeMessageTypePong        = "pong"

	realtimeErrorCodeInvalidMessage = "INVALID_MESSAGE"
	realtimeErrorCodeInvalidRequest = "INVALID_REQUEST"
	realtimeErrorCodeStateFetchFail = "STATE_FETCH_FAILED"
)

type realtimeClientMessage struct {
	Type       string `json:"type"`
	RequestID  string `json:"requestId"`
	Nickname   string `json:"nickname"`
	Slug       string `json:"slug"`
	ComboCount int64  `json:"comboCount"`
}

type realtimeSnapshotMessage struct {
	Type   string        `json:"type"`
	Public core.Snapshot `json:"public"`
	User   any           `json:"user"`
}

type realtimeDeltaMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type realtimeSnapshotUser struct {
	UserStats                          *core.UserStats     `json:"userStats,omitempty"`
	MyBossStats                        *core.BossUserStats `json:"myBossStats,omitempty"`
	RoomID                             string              `json:"roomId,omitempty"`
	Loadout                            core.Loadout        `json:"loadout"`
	CombatStats                        core.CombatStats    `json:"combatStats"`
	Gold                               int64               `json:"gold"`
	Stones                             int64               `json:"stones"`
	TalentPoints                       int64               `json:"talentPoints"`
	RecentRewards                      []core.Reward       `json:"recentRewards,omitempty"`
	EquippedBattleClickSkinID          string              `json:"equippedBattleClickSkinId,omitempty"`
	EquippedBattleClickCursorImagePath string              `json:"equippedBattleClickCursorImagePath,omitempty"`
}

type realtimeClickAckPayload struct {
	Delta                int64                     `json:"delta"`
	Critical             bool                      `json:"critical"`
	BossDamage           int64                     `json:"bossDamage,omitempty"`
	MyBossDamage         int64                     `json:"myBossDamage,omitempty"`
	BossLeaderboardCount int                       `json:"bossLeaderboardCount,omitempty"`
	DamageType           string                    `json:"damageType,omitempty"`
	TalentEvents         []core.TalentTriggerEvent `json:"talentEvents,omitempty"`
	PartStateDeltas      []core.BossPartStateDelta `json:"partStateDeltas,omitempty"`
	TalentCombatState    *core.TalentCombatState   `json:"talentCombatState,omitempty"`
	UserDelta            *realtimeUserDelta        `json:"userDelta,omitempty"`
	Button               struct {
		Key string `json:"key"`
	} `json:"button"`
}

type realtimeUserDelta struct {
	Gold         *int64 `json:"gold,omitempty"`
	Stones       *int64 `json:"stones,omitempty"`
	TalentPoints *int64 `json:"talentPoints,omitempty"`
}

type realtimeClickAckMessage struct {
	Type    string                  `json:"type"`
	Payload realtimeClickAckPayload `json:"payload"`
}

type realtimeErrorMessage struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type realtimePongMessage struct {
	Type string `json:"type"`
}

type realtimeSessionOptions struct {
	stateView             StateView
	store                 ButtonStore
	hub                   RealtimeHub
	changePublisher       ChangePublisher
	clickGuard            ClickGuard
	authenticatorEnabled  bool
	authenticatedNickname string
	clientID              string
}

type realtimeSession struct {
	stateView             StateView
	store                 ButtonStore
	hub                   RealtimeHub
	changePublisher       ChangePublisher
	clickGuard            ClickGuard
	authenticatorEnabled  bool
	authenticatedNickname string
	clientID              string
	nickname              string
	updates               <-chan events.ServerEvent
	unsubscribe           func()
	lastActiveAt          time.Time
}

func newRealtimeSession(options realtimeSessionOptions) *realtimeSession {
	return &realtimeSession{
		stateView:             options.stateView,
		store:                 options.store,
		hub:                   options.hub,
		changePublisher:       options.changePublisher,
		clickGuard:            options.clickGuard,
		authenticatorEnabled:  options.authenticatorEnabled,
		authenticatedNickname: strings.TrimSpace(options.authenticatedNickname),
		clientID:              strings.TrimSpace(options.clientID),
	}
}

func newRealtimeSocketHandler(options Options) app.HandlerFunc {
	upgrader := websocket.HertzUpgrader{
		CheckOrigin:       func(_ *app.RequestContext) bool { return true },
		EnableCompression: true,
	}

	return func(ctx context.Context, c *app.RequestContext) {
		authenticatedNickname := authenticatedPlayerNickname(ctx, c, options.PlayerAuthenticator)
		session := newRealtimeSession(realtimeSessionOptions{
			stateView:             effectiveStateView(options),
			store:                 options.Store,
			hub:                   options.RealtimeHub,
			changePublisher:       options.ChangePublisher,
			clickGuard:            options.ClickGuard,
			authenticatorEnabled:  options.PlayerAuthenticator != nil,
			authenticatedNickname: authenticatedNickname,
			clientID:              clientIdentifier(c),
		})

		_ = upgrader.Upgrade(c, func(conn *websocket.Conn) {
			runRealtimeConnection(conn, session)
		})
	}
}

func effectiveStateView(options Options) StateView {
	if options.StateView != nil {
		return options.StateView
	}
	return options.Store
}

func runRealtimeConnection(conn *websocket.Conn, session *realtimeSession) {
	connectionCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer session.close()
	defer conn.Close()

	readCh := make(chan []byte)
	readDone := make(chan struct{})

	go func() {
		defer close(readDone)
		for {
			_, payload, err := conn.ReadMessage()
			if err != nil {
				return
			}
			select {
			case readCh <- append([]byte(nil), payload...):
			case <-connectionCtx.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-readDone:
			return
		case payload := <-readCh:
			if err := session.handleMessage(connectionCtx, payload, func(message any) error {
				conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				return conn.WriteJSON(message)
			}); err != nil {
				return
			}
		case update, ok := <-session.updates:
			if !ok {
				session.updates = nil
				continue
			}
			if err := session.writeUpdate(conn, update); err != nil {
				return
			}
		}
	}
}

func (s *realtimeSession) handleMessage(ctx context.Context, payload []byte, send func(any) error) error {
	s.lastActiveAt = time.Now()

	var message realtimeClientMessage
	if err := sonic.Unmarshal(payload, &message); err != nil {
		return send(s.protocolError(realtimeErrorCodeInvalidMessage, "实时消息格式不对，请刷新页面后重试。"))
	}

	switch strings.TrimSpace(message.Type) {
	case realtimeMessageTypeHello:
		s.setNickname(resolveRealtimeReadNickname(s.authenticatorEnabled, s.authenticatedNickname, message.Nickname))
		return s.sendSnapshot(ctx, send)
	case realtimeMessageTypeSyncRequest:
		return s.sendSnapshot(ctx, send)
	case realtimeMessageTypePing:
		return send(realtimePongMessage{Type: realtimeMessageTypePong})
	case realtimeMessageTypeClick:
		slug := strings.TrimSpace(message.Slug)
		if slug == "" {
			return send(s.protocolError(realtimeErrorCodeInvalidRequest, "点击消息缺少按钮标识。"))
		}

		nickname, result, apiErr := executeButtonClick(ctx, Options{
			Store:           s.store,
			ClickGuard:      s.clickGuard,
			ChangePublisher: s.changePublisher,
		}, clickRequestContext{
			Slug:                  slug,
			NicknameHint:          s.nickname,
			AuthenticatedNickname: s.authenticatedNickname,
			AuthenticatorEnabled:  s.authenticatorEnabled,
			ClientID:              s.clientID,
			ComboCount:            message.ComboCount,
		})
		if apiErr != nil {
			return send(s.protocolError(apiErr.Code, apiErr.Message))
		}

		s.setNickname(resolveRealtimeReadNickname(s.authenticatorEnabled, nickname, nickname))

		change := core.StateChange{
			Type:      core.StateChangeButtonClicked,
			Nickname:  nickname,
			RoomID:    result.RoomID,
			Timestamp: time.Now().Unix(),
		}
		if result.BroadcastUserAll {
			change.BroadcastUserAll = true
		}
		publishChange(ctx, s.changePublisher, change)

		ack := realtimeClickAckMessage{
			Type: realtimeMessageTypeClickAck,
			Payload: realtimeClickAckPayload{
				Delta:                result.Delta,
				Critical:             result.Critical,
				BossDamage:           result.BossDamage,
				MyBossDamage:         result.MyBossDamage,
				BossLeaderboardCount: result.BossLeaderboardCount,
				DamageType:           result.DamageType,
				TalentEvents:         result.TalentEvents,
				PartStateDeltas:      result.PartStateDeltas,
				TalentCombatState:    result.TalentCombatState,
				Button: struct {
					Key string `json:"key"`
				}{
					Key: slug,
				},
			},
		}
		if s.nickname != "" && s.stateView != nil {
			if userState, err := s.stateView.GetUserState(ctx, s.nickname); err == nil {
				ack.Payload.UserDelta = &realtimeUserDelta{
					Gold:         &userState.Gold,
					Stones:       &userState.Stones,
					TalentPoints: &userState.TalentPoints,
				}
			}
		}
		return send(ack)
	default:
		return send(s.protocolError(realtimeErrorCodeInvalidMessage, "不支持的实时消息类型。"))
	}
}

func (s *realtimeSession) sendSnapshot(ctx context.Context, send func(any) error) error {
	if s.stateView == nil {
		return send(s.protocolError(realtimeErrorCodeStateFetchFail, "实时状态同步失败，请稍后重试。"))
	}

	snapshot, err := s.stateView.GetSnapshot(ctx)
	if s.nickname != "" {
		if reader, ok := s.stateView.(interface {
			GetSnapshotForNickname(context.Context, string) (core.Snapshot, error)
		}); ok {
			snapshot, err = reader.GetSnapshotForNickname(ctx, s.nickname)
		}
	}
	if err != nil {
		return send(s.protocolError(realtimeErrorCodeStateFetchFail, "实时状态同步失败，请稍后重试。"))
	}

	var userState any
	if s.nickname != "" {
		state, err := s.stateView.GetUserState(ctx, s.nickname)
		if err != nil {
			return send(s.protocolError(realtimeErrorCodeStateFetchFail, "实时状态同步失败，请稍后重试。"))
		}
		payload := buildRealtimeSnapshotUser(state)
		userState = &payload
	}

	return send(realtimeSnapshotMessage{
		Type:   realtimeMessageTypeSnapshot,
		Public: snapshot,
		User:   userState,
	})
}

func buildRealtimeSnapshotUser(state core.UserState) realtimeSnapshotUser {
	return realtimeSnapshotUser{
		UserStats:                          state.UserStats,
		MyBossStats:                        state.MyBossStats,
		RoomID:                             state.RoomID,
		Loadout:                            state.Loadout,
		CombatStats:                        state.CombatStats,
		Gold:                               state.Gold,
		Stones:                             state.Stones,
		TalentPoints:                       state.TalentPoints,
		RecentRewards:                      state.RecentRewards,
		EquippedBattleClickSkinID:          state.EquippedBattleClickSkinID,
		EquippedBattleClickCursorImagePath: state.EquippedBattleClickCursorImagePath,
	}
}

func (s *realtimeSession) setNickname(nickname string) {
	normalizedNickname := strings.TrimSpace(nickname)
	if s.nickname == normalizedNickname && s.updates != nil {
		return
	}

	if s.unsubscribe != nil {
		s.unsubscribe()
		s.unsubscribe = nil
		s.updates = nil
	}
	s.nickname = normalizedNickname

	if s.hub == nil {
		return
	}

	updates, unsubscribe := s.hub.Subscribe(normalizedNickname)
	s.updates = updates
	s.unsubscribe = unsubscribe
}

func (s *realtimeSession) close() {
	if s.unsubscribe != nil {
		s.unsubscribe()
		s.unsubscribe = nil
	}
}

func (s *realtimeSession) writeUpdate(conn *websocket.Conn, update events.ServerEvent) error {
	message, ok := realtimeMessageFromEvent(update)
	if !ok {
		return nil
	}
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return conn.WriteJSON(message)
}

func (s *realtimeSession) protocolError(code string, message string) realtimeErrorMessage {
	return realtimeErrorMessage{
		Type:    realtimeMessageTypeError,
		Code:    code,
		Message: message,
	}
}

func realtimeMessageFromEvent(event events.ServerEvent) (any, bool) {
	switch event.Name {
	case events.PublicStateEventName:
		return realtimeDeltaMessage{
			Type:    realtimeMessageTypePublicDelta,
			Payload: event.Payload,
		}, true
	case events.UserStateEventName:
		return realtimeDeltaMessage{
			Type:    realtimeMessageTypeUserDelta,
			Payload: event.Payload,
		}, true
	case events.OnlineCountEventName:
		return realtimeDeltaMessage{
			Type:    realtimeMessageTypeOnlineCount,
			Payload: event.Payload,
		}, true
	default:
		return nil, false
	}
}
