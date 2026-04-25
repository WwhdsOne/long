package httpapi

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"long/internal/vote"
)

type afkPlayerState struct {
	visible        bool
	lastChangedAt  time.Time
	hiddenSince    time.Time
	afkActive      bool
	afkStartedAt   time.Time
	baselineGold   int64
	baselineStones int64
	kills          int64
}

// AfkService 负责按页面可见性托管挂机与结算。
type AfkService struct {
	mu              sync.RWMutex
	store           ButtonStore
	changePublisher ChangePublisher
	now             func() time.Time
	players         map[string]*afkPlayerState
	settlements     map[string]vote.AfkSettlement
	stopCh          chan struct{}
	doneCh          chan struct{}
	closed          bool
}

func NewAfkService(store ButtonStore, publisher ChangePublisher) *AfkService {
	s := &AfkService{
		store:           store,
		changePublisher: publisher,
		now:             time.Now,
		players:         make(map[string]*afkPlayerState),
		settlements:     make(map[string]vote.AfkSettlement),
		stopCh:          make(chan struct{}),
		doneCh:          make(chan struct{}),
	}
	go s.loop()
	return s
}

func (s *AfkService) ReportPresence(ctx context.Context, nickname string, visible bool) error {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return vote.ErrInvalidNickname
	}

	now := s.now()
	s.mu.Lock()
	state := s.players[nickname]
	if state == nil {
		state = &afkPlayerState{visible: true}
		s.players[nickname] = state
	}
	state.visible = visible
	state.lastChangedAt = now
	if visible {
		state.hiddenSince = time.Time{}
	} else {
		state.hiddenSince = now
	}
	wasActive := state.afkActive
	s.mu.Unlock()

	if visible && wasActive {
		s.stopAfk(ctx, nickname)
	}

	return nil
}

func (s *AfkService) ConsumeSettlement(nickname string) vote.AfkSettlement {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return vote.AfkSettlement{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	result := s.settlements[nickname]
	delete(s.settlements, nickname)
	return result
}

func (s *AfkService) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	close(s.stopCh)
	doneCh := s.doneCh
	s.mu.Unlock()
	<-doneCh
	return nil
}

func (s *AfkService) loop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	defer close(s.doneCh)

	for {
		select {
		case <-ticker.C:
			s.runOnce(context.Background())
		case <-s.stopCh:
			return
		}
	}
}

func (s *AfkService) runOnce(ctx context.Context) {
	now := s.now()
	nicknames := make([]string, 0)

	s.mu.Lock()
	for nickname, state := range s.players {
		if state.visible {
			continue
		}
		if !state.afkActive {
			if state.hiddenSince.IsZero() || now.Sub(state.hiddenSince) < 60*time.Second {
				continue
			}
			userState, err := s.store.GetUserState(ctx, nickname)
			if err != nil {
				continue
			}
			state.afkActive = true
			state.afkStartedAt = now
			state.baselineGold = userState.Gold
			state.baselineStones = userState.Stones
			state.kills = 0
		}
		nicknames = append(nicknames, nickname)
	}
	s.mu.Unlock()

	for _, nickname := range nicknames {
		result, err := s.store.AttackBossPartAFK(ctx, nickname)
		if err != nil {
			if errors.Is(err, vote.ErrInvalidNickname) || errors.Is(err, vote.ErrSensitiveNickname) {
				continue
			}
			log.Printf("afk attack failed nickname=%s err=%v", nickname, err)
			continue
		}
		if result.BroadcastUserAll {
			s.mu.Lock()
			if state := s.players[nickname]; state != nil && state.afkActive {
				state.kills++
			}
			s.mu.Unlock()
			publishChange(ctx, s.changePublisher, vote.StateChange{
				Type:             vote.StateChangeBossChanged,
				BroadcastUserAll: true,
				Timestamp:        now.Unix(),
			})
		} else if result.Boss != nil {
			publishChange(ctx, s.changePublisher, vote.StateChange{
				Type:      vote.StateChangeBossChanged,
				Timestamp: now.Unix(),
			})
		}
	}
}

func (s *AfkService) stopAfk(ctx context.Context, nickname string) {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return
	}
	now := s.now()

	s.mu.Lock()
	state := s.players[nickname]
	if state == nil || !state.afkActive {
		s.mu.Unlock()
		return
	}
	startedAt := state.afkStartedAt
	baselineGold := state.baselineGold
	baselineStones := state.baselineStones
	kills := state.kills
	state.afkActive = false
	state.afkStartedAt = time.Time{}
	s.mu.Unlock()

	userState, err := s.store.GetUserState(ctx, nickname)
	if err != nil {
		return
	}
	goldDelta := userState.Gold - baselineGold
	stoneDelta := userState.Stones - baselineStones
	if goldDelta < 0 {
		goldDelta = 0
	}
	if stoneDelta < 0 {
		stoneDelta = 0
	}
	if kills <= 0 && goldDelta <= 0 && stoneDelta <= 0 {
		return
	}

	s.mu.Lock()
	existing := s.settlements[nickname]
	if existing.StartedAt == 0 || (startedAt.Unix() > 0 && startedAt.Unix() < existing.StartedAt) {
		existing.StartedAt = startedAt.Unix()
	}
	existing.EndedAt = now.Unix()
	existing.Kills += kills
	existing.GoldTotal += goldDelta
	existing.StoneTotal += stoneDelta
	s.settlements[nickname] = existing
	s.mu.Unlock()
}
