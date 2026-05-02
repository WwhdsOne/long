package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"long/internal/core"
	"long/internal/xlog"
)

const afkStartDelay = 60 * time.Second
const afkMaxDuration = 8 * time.Hour

type afkPlayerState struct {
	LastSeenAt     int64
	HiddenSince    int64
	AfkActive      bool
	AfkStartedAt   int64
	BaselineGold   int64
	BaselineStones int64
	Kills          int64
	Rewards        []core.Reward
}

type afkSettlementState struct {
	Kills      int64
	GoldTotal  int64
	StoneTotal int64
	StartedAt  int64
	EndedAt    int64
	Rewards    []core.Reward
}

// AfkService 负责按页面可见性托管挂机与结算。
type AfkService struct {
	store           ButtonStore
	changePublisher ChangePublisher
	redis           redis.UniversalClient
	now             func() time.Time
	keyPrefix       string
	stopCh          chan struct{}
	doneCh          chan struct{}

	mu     sync.Mutex
	closed bool
}

func NewAfkService(store ButtonStore, publisher ChangePublisher, redisClient redis.UniversalClient, namespace string) *AfkService {
	s := &AfkService{
		store:           store,
		changePublisher: publisher,
		redis:           redisClient,
		now:             time.Now,
		keyPrefix:       strings.TrimSpace(namespace) + "afk:",
		stopCh:          make(chan struct{}),
		doneCh:          make(chan struct{}),
	}
	go s.loop()
	return s
}

func (s *AfkService) ReportPresence(ctx context.Context, nickname string, visible bool) error {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return core.ErrInvalidNickname
	}
	if s.redis == nil {
		return nil
	}

	nowUnix := s.now().Unix()
	state, exists, err := s.loadPlayerState(ctx, nickname)
	if err != nil {
		return err
	}
	if !exists {
		state = afkPlayerState{}
	}

	if visible {
		state.LastSeenAt = nowUnix
		state.HiddenSince = 0
	} else {
		if state.HiddenSince == 0 {
			state.HiddenSince = nowUnix
		}
	}

	if err := s.savePlayerState(ctx, nickname, state); err != nil {
		return err
	}
	if visible && state.AfkActive {
		s.stopAfk(ctx, nickname)
	}
	return nil
}

func (s *AfkService) ConsumeSettlement(nickname string) core.AfkSettlement {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" || s.redis == nil {
		return core.AfkSettlement{}
	}

	ctx := context.Background()
	state, exists, err := s.loadSettlement(ctx, nickname)
	if err != nil || !exists {
		return core.AfkSettlement{}
	}
	if err := s.redis.Del(ctx, s.settlementKey(nickname)).Err(); err != nil {
		xlog.L().Error("afk consume settlement delete failed", xlog.Err(err))
	}
	return core.AfkSettlement{
		Kills:      state.Kills,
		GoldTotal:  state.GoldTotal,
		StoneTotal: state.StoneTotal,
		StartedAt:  state.StartedAt,
		EndedAt:    state.EndedAt,
		Rewards:    state.Rewards,
	}
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
	if s.redis == nil {
		return
	}

	nicknames, err := s.redis.ZRange(ctx, s.playersIndexKey(), 0, -1).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		xlog.L().Error("afk load players failed", xlog.Err(err))
		return
	}

	now := s.now()
	nowUnix := now.Unix()
	for _, nickname := range nicknames {
		nickname = strings.TrimSpace(nickname)
		if nickname == "" {
			continue
		}
		s.runPlayerOnce(ctx, nickname, nowUnix)
	}
}

func (s *AfkService) runPlayerOnce(ctx context.Context, nickname string, nowUnix int64) {
	state, exists, err := s.loadPlayerState(ctx, nickname)
	if err != nil {
		xlog.L().Error("afk load player failed", xlog.Err(err))
		return
	}
	if !exists {
		return
	}

	inactiveFor := int64(0)
	if state.LastSeenAt > 0 {
		inactiveFor = nowUnix - state.LastSeenAt
	}

	if inactiveFor < int64(afkStartDelay.Seconds()) {
		if state.AfkActive {
			s.stopAfk(ctx, nickname)
		}
		return
	}

	if !state.AfkActive {
		if err := s.startAfk(ctx, nickname, &state, nowUnix); err != nil {
			xlog.L().Error("afk start failed", xlog.Err(err))
			return
		}
	}

	if state.AfkStartedAt > 0 && nowUnix-state.AfkStartedAt >= int64(afkMaxDuration.Seconds()) {
		return
	}

	result, err := s.store.AttackBossPartAFK(ctx, nickname)
	if err != nil {
		if errors.Is(err, core.ErrInvalidNickname) || errors.Is(err, core.ErrSensitiveNickname) {
			return
		}
		xlog.L().Error("afk attack failed", xlog.Err(err))
		return
	}

	if result.BroadcastUserAll {
		state.Kills++
		if len(result.RecentRewards) > 0 {
			state.Rewards = append(state.Rewards, result.RecentRewards...)
		}
		if err := s.savePlayerState(ctx, nickname, state); err != nil {
			xlog.L().Error("afk save player failed", xlog.Err(err))
		}
		publishChange(ctx, s.changePublisher, core.StateChange{
			Type:             core.StateChangeBossChanged,
			BroadcastUserAll: true,
			Timestamp:        nowUnix,
		})
		return
	}

	if result.Boss != nil {
		publishChange(ctx, s.changePublisher, core.StateChange{
			Type:      core.StateChangeBossChanged,
			Timestamp: nowUnix,
		})
	}
}

func (s *AfkService) startAfk(ctx context.Context, nickname string, state *afkPlayerState, nowUnix int64) error {
	userState, err := s.store.GetUserState(ctx, nickname)
	if err != nil {
		return err
	}
	state.AfkActive = true
	state.AfkStartedAt = nowUnix
	state.BaselineGold = userState.Gold
	state.BaselineStones = userState.Stones
	state.Kills = 0
	state.Rewards = nil
	if state.HiddenSince == 0 {
		state.HiddenSince = nowUnix - int64(afkStartDelay.Seconds())
	}
	return s.savePlayerState(ctx, nickname, *state)
}

func (s *AfkService) stopAfk(ctx context.Context, nickname string) {
	state, exists, err := s.loadPlayerState(ctx, nickname)
	if err != nil || !exists || !state.AfkActive {
		return
	}

	userState, err := s.store.GetUserState(ctx, nickname)
	if err != nil {
		return
	}

	goldDelta := userState.Gold - state.BaselineGold
	stoneDelta := userState.Stones - state.BaselineStones
	if goldDelta < 0 {
		goldDelta = 0
	}
	if stoneDelta < 0 {
		stoneDelta = 0
	}

	if state.Kills > 0 || goldDelta > 0 || stoneDelta > 0 || len(state.Rewards) > 0 {
		if err := s.mergeSettlement(ctx, nickname, afkSettlementState{
			Kills:      state.Kills,
			GoldTotal:  goldDelta,
			StoneTotal: stoneDelta,
			StartedAt:  state.AfkStartedAt,
			EndedAt:    s.now().Unix(),
			Rewards:    append([]core.Reward(nil), state.Rewards...),
		}); err != nil {
			xlog.L().Error("afk merge settlement failed", xlog.Err(err))
		}
	}

	state.AfkActive = false
	state.AfkStartedAt = 0
	state.BaselineGold = 0
	state.BaselineStones = 0
	state.Kills = 0
	state.Rewards = nil
	state.HiddenSince = 0
	_ = s.savePlayerState(ctx, nickname, state)
}

func (s *AfkService) loadPlayerState(ctx context.Context, nickname string) (afkPlayerState, bool, error) {
	values, err := s.redis.HGetAll(ctx, s.playerStateKey(nickname)).Result()
	if err != nil {
		return afkPlayerState{}, false, err
	}
	if len(values) == 0 {
		return afkPlayerState{}, false, nil
	}

	rewards, _ := parseRewardJSON(values["rewards_json"])
	return afkPlayerState{
		LastSeenAt:     parseInt64(values["last_seen_at"]),
		HiddenSince:    parseInt64(values["hidden_since"]),
		AfkActive:      parseBool(values["afk_active"]),
		AfkStartedAt:   parseInt64(values["afk_started_at"]),
		BaselineGold:   parseInt64(values["baseline_gold"]),
		BaselineStones: parseInt64(values["baseline_stones"]),
		Kills:          parseInt64(values["kills"]),
		Rewards:        rewards,
	}, true, nil
}

func (s *AfkService) savePlayerState(ctx context.Context, nickname string, state afkPlayerState) error {
	rewardsRaw, err := json.Marshal(state.Rewards)
	if err != nil {
		return err
	}
	if err := s.redis.HSet(ctx, s.playerStateKey(nickname), map[string]any{
		"last_seen_at":    strconv.FormatInt(state.LastSeenAt, 10),
		"hidden_since":    strconv.FormatInt(state.HiddenSince, 10),
		"afk_active":      boolToInt(state.AfkActive),
		"afk_started_at":  strconv.FormatInt(state.AfkStartedAt, 10),
		"baseline_gold":   strconv.FormatInt(state.BaselineGold, 10),
		"baseline_stones": strconv.FormatInt(state.BaselineStones, 10),
		"kills":           strconv.FormatInt(state.Kills, 10),
		"rewards_json":    string(rewardsRaw),
	}).Err(); err != nil {
		return err
	}
	score := float64(state.LastSeenAt)
	if score <= 0 {
		score = float64(s.now().Unix())
	}
	return s.redis.ZAdd(ctx, s.playersIndexKey(), redis.Z{
		Score:  score,
		Member: nickname,
	}).Err()
}

func (s *AfkService) loadSettlement(ctx context.Context, nickname string) (afkSettlementState, bool, error) {
	values, err := s.redis.HGetAll(ctx, s.settlementKey(nickname)).Result()
	if err != nil {
		return afkSettlementState{}, false, err
	}
	if len(values) == 0 {
		return afkSettlementState{}, false, nil
	}
	rewards, _ := parseRewardJSON(values["rewards_json"])
	return afkSettlementState{
		Kills:      parseInt64(values["kills"]),
		GoldTotal:  parseInt64(values["gold_total"]),
		StoneTotal: parseInt64(values["stone_total"]),
		StartedAt:  parseInt64(values["started_at"]),
		EndedAt:    parseInt64(values["ended_at"]),
		Rewards:    rewards,
	}, true, nil
}

func (s *AfkService) mergeSettlement(ctx context.Context, nickname string, delta afkSettlementState) error {
	existing, exists, err := s.loadSettlement(ctx, nickname)
	if err != nil {
		return err
	}
	if !exists {
		existing = afkSettlementState{}
	}
	if existing.StartedAt == 0 || (delta.StartedAt > 0 && delta.StartedAt < existing.StartedAt) {
		existing.StartedAt = delta.StartedAt
	}
	if delta.EndedAt > existing.EndedAt {
		existing.EndedAt = delta.EndedAt
	}
	existing.Kills += delta.Kills
	existing.GoldTotal += delta.GoldTotal
	existing.StoneTotal += delta.StoneTotal
	if len(delta.Rewards) > 0 {
		existing.Rewards = append(existing.Rewards, delta.Rewards...)
	}

	rewardsRaw, err := json.Marshal(existing.Rewards)
	if err != nil {
		return err
	}
	return s.redis.HSet(ctx, s.settlementKey(nickname), map[string]any{
		"kills":        strconv.FormatInt(existing.Kills, 10),
		"gold_total":   strconv.FormatInt(existing.GoldTotal, 10),
		"stone_total":  strconv.FormatInt(existing.StoneTotal, 10),
		"started_at":   strconv.FormatInt(existing.StartedAt, 10),
		"ended_at":     strconv.FormatInt(existing.EndedAt, 10),
		"rewards_json": string(rewardsRaw),
	}).Err()
}

func (s *AfkService) playersIndexKey() string {
	return s.keyPrefix + "players"
}

func (s *AfkService) playerStateKey(nickname string) string {
	return s.keyPrefix + "player:" + nickname
}

func (s *AfkService) settlementKey(nickname string) string {
	return s.keyPrefix + "settlement:" + nickname
}

func parseRewardJSON(raw string) ([]core.Reward, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var rewards []core.Reward
	if err := json.Unmarshal([]byte(raw), &rewards); err != nil {
		return nil, err
	}
	return rewards, nil
}

func parseInt64(raw string) int64 {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0
	}
	return value
}

func parseBool(raw string) bool {
	raw = strings.TrimSpace(strings.ToLower(raw))
	return raw == "1" || raw == "true"
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
