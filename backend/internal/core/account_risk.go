package core

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type AccountRiskEvent string

const (
	AccountRiskEventClickRateLimitHit        AccountRiskEvent = "click_rate_limit_hit"
	AccountRiskEventLoginTurnstileInvalid    AccountRiskEvent = "login_turnstile_invalid"
	AccountRiskEventStaminaTurnstileInvalid  AccountRiskEvent = "stamina_turnstile_invalid"
	AccountRiskEventPostStaminaPurchaseClick AccountRiskEvent = "post_stamina_purchase_click"
)

type AccountRiskPoints struct {
	ClickRateLimitHit        int64
	LoginTurnstileInvalid    int64
	StaminaTurnstileInvalid  int64
	PostStaminaPurchaseClick int64
}

type AccountRiskConfig struct {
	ScoreWindow           time.Duration
	PurchaseClickCooldown time.Duration
	BanThreshold8h        int64
	BanThreshold24h       int64
	BanThreshold72h       int64
	Points                AccountRiskPoints
}

type AccountRiskState struct {
	Nickname string `json:"nickname,omitempty"`
	Score    int64  `json:"score"`
	BanUntil int64  `json:"banUntil,omitempty"`
}

func normalizeAccountRiskConfig(cfg AccountRiskConfig) AccountRiskConfig {
	if cfg.ScoreWindow <= 0 {
		cfg.ScoreWindow = 24 * time.Hour
	}
	if cfg.PurchaseClickCooldown < 0 {
		cfg.PurchaseClickCooldown = 0
	}
	if cfg.BanThreshold8h <= 0 {
		cfg.BanThreshold8h = 6
	}
	if cfg.BanThreshold24h <= cfg.BanThreshold8h {
		cfg.BanThreshold24h = 10
	}
	if cfg.BanThreshold72h <= cfg.BanThreshold24h {
		cfg.BanThreshold72h = 14
	}
	if cfg.Points.ClickRateLimitHit <= 0 {
		cfg.Points.ClickRateLimitHit = 2
	}
	if cfg.Points.LoginTurnstileInvalid <= 0 {
		cfg.Points.LoginTurnstileInvalid = 3
	}
	if cfg.Points.StaminaTurnstileInvalid <= 0 {
		cfg.Points.StaminaTurnstileInvalid = 3
	}
	if cfg.Points.PostStaminaPurchaseClick <= 0 {
		cfg.Points.PostStaminaPurchaseClick = 4
	}
	return cfg
}

func (s *Store) accountRiskEventPoint(event AccountRiskEvent) int64 {
	switch event {
	case AccountRiskEventClickRateLimitHit:
		return s.accountRiskConfig.Points.ClickRateLimitHit
	case AccountRiskEventLoginTurnstileInvalid:
		return s.accountRiskConfig.Points.LoginTurnstileInvalid
	case AccountRiskEventStaminaTurnstileInvalid:
		return s.accountRiskConfig.Points.StaminaTurnstileInvalid
	case AccountRiskEventPostStaminaPurchaseClick:
		return s.accountRiskConfig.Points.PostStaminaPurchaseClick
	default:
		return 0
	}
}

func (s *Store) accountRiskBanDuration(score int64) time.Duration {
	switch {
	case score >= s.accountRiskConfig.BanThreshold72h:
		return 72 * time.Hour
	case score >= s.accountRiskConfig.BanThreshold24h:
		return 24 * time.Hour
	case score >= s.accountRiskConfig.BanThreshold8h:
		return 8 * time.Hour
	default:
		return 0
	}
}

func (s *Store) RecordAccountRiskEvent(ctx context.Context, nickname string, event AccountRiskEvent) (AccountRiskState, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return AccountRiskState{}, err
	}
	points := s.accountRiskEventPoint(event)
	if points <= 0 {
		return s.GetAccountRiskState(ctx, normalizedNickname)
	}

	now := s.now()
	nowUnix := now.Unix()
	seq, err := s.client.Incr(ctx, s.accountRiskEventSeqKey(normalizedNickname)).Result()
	if err != nil {
		return AccountRiskState{}, err
	}
	member := fmt.Sprintf("%d:%d:%s", now.UnixNano(), seq, event)
	eventKey := s.accountRiskEventKey(normalizedNickname)
	indexKey := s.accountRiskIndexKey

	pipe := s.client.TxPipeline()
	pipe.ZAdd(ctx, eventKey, redis.Z{
		Score:  float64(nowUnix),
		Member: member,
	})
	pipe.Expire(ctx, eventKey, s.accountRiskConfig.ScoreWindow+time.Hour)
	pipe.SAdd(ctx, indexKey, normalizedNickname)
	if _, err := pipe.Exec(ctx); err != nil {
		return AccountRiskState{}, err
	}

	state, err := s.recalculateAccountRiskState(ctx, normalizedNickname)
	if err != nil {
		return AccountRiskState{}, err
	}

	s.writeAccountRiskEventLog(ctx, AccountRiskEventLog{
		Nickname:      normalizedNickname,
		EventType:     string(event),
		Points:        points,
		ScoreAfter:    state.Score,
		BanUntilAfter: state.BanUntil,
		CreatedAt:     nowUnix,
	})
	return state, nil
}

func (s *Store) GetAccountRiskState(ctx context.Context, nickname string) (AccountRiskState, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return AccountRiskState{}, err
	}
	return s.recalculateAccountRiskState(ctx, normalizedNickname)
}

func (s *Store) GetAccountRiskBanStatus(ctx context.Context, nickname string) (int64, bool, error) {
	state, err := s.GetAccountRiskState(ctx, nickname)
	if err != nil {
		return 0, false, err
	}
	return state.BanUntil, state.BanUntil > s.now().Unix(), nil
}

func (s *Store) ClearAccountRiskState(ctx context.Context, nickname string) error {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return err
	}
	_, err = s.client.Del(ctx,
		s.accountRiskEventKey(normalizedNickname),
		s.accountRiskEventSeqKey(normalizedNickname),
		s.accountRiskBanKey(normalizedNickname),
		s.accountRiskBanStartKey(normalizedNickname),
		s.accountRiskLastStaminaPurchaseKey(normalizedNickname),
	).Result()
	if err != nil {
		return err
	}
	if err := s.client.SRem(ctx, s.accountRiskIndexKey, normalizedNickname).Err(); err != nil {
		return err
	}
	s.writeAccountRiskEventLog(ctx, AccountRiskEventLog{
		Nickname:  normalizedNickname,
		EventType: "risk_state_cleared",
		CreatedAt: s.now().Unix(),
	})
	return nil
}

func (s *Store) ListAccountRiskEntries(ctx context.Context) ([]AccountRiskState, error) {
	nicknames, err := s.client.SMembers(ctx, s.accountRiskIndexKey).Result()
	if err != nil {
		if err == redis.Nil {
			return []AccountRiskState{}, nil
		}
		return nil, err
	}
	entries := make([]AccountRiskState, 0, len(nicknames))
	for _, nickname := range nicknames {
		trimmed := strings.TrimSpace(nickname)
		if trimmed == "" {
			continue
		}
		state, err := s.recalculateAccountRiskState(ctx, trimmed)
		if err != nil {
			return nil, err
		}
		if state.Score <= 0 {
			continue
		}
		entries = append(entries, state)
	}
	slices.SortFunc(entries, func(a, b AccountRiskState) int {
		if a.Score != b.Score {
			if a.Score > b.Score {
				return -1
			}
			return 1
		}
		if a.BanUntil != b.BanUntil {
			if a.BanUntil > b.BanUntil {
				return -1
			}
			return 1
		}
		return strings.Compare(a.Nickname, b.Nickname)
	})
	return entries, nil
}

func (s *Store) MarkStaminaPurchase(ctx context.Context, nickname string) error {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, s.accountRiskLastStaminaPurchaseKey(normalizedNickname), strconv.FormatInt(s.now().Unix(), 10), s.accountRiskConfig.ScoreWindow).Err()
}

func (s *Store) shouldRecordPostStaminaPurchaseClick(ctx context.Context, nickname string) (bool, error) {
	if s.accountRiskConfig.PurchaseClickCooldown <= 0 {
		return false, nil
	}
	raw, err := s.client.Get(ctx, s.accountRiskLastStaminaPurchaseKey(nickname)).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	lastPurchaseAt, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return false, nil
	}
	if lastPurchaseAt <= 0 {
		return false, nil
	}
	return s.now().Unix()-lastPurchaseAt <= int64(s.accountRiskConfig.PurchaseClickCooldown.Seconds()), nil
}

func (s *Store) recalculateAccountRiskState(ctx context.Context, nickname string) (AccountRiskState, error) {
	nowUnix := s.now().Unix()
	windowStart := nowUnix - int64(s.accountRiskConfig.ScoreWindow.Seconds())
	eventKey := s.accountRiskEventKey(nickname)

	if err := s.client.ZRemRangeByScore(ctx, eventKey, "-inf", strconv.FormatInt(windowStart-1, 10)).Err(); err != nil {
		return AccountRiskState{}, err
	}
	members, err := s.client.ZRangeByScore(ctx, eventKey, &redis.ZRangeBy{
		Min: strconv.FormatInt(windowStart, 10),
		Max: strconv.FormatInt(nowUnix, 10),
	}).Result()
	if err != nil && err != redis.Nil {
		return AccountRiskState{}, err
	}

	score := int64(0)
	for _, member := range members {
		score += s.accountRiskEventPoint(parseAccountRiskEventMember(member))
	}

	banKey := s.accountRiskBanKey(nickname)
	banStartKey := s.accountRiskBanStartKey(nickname)
	currentBanUntil := int64(0)
	rawBan, err := s.client.Get(ctx, banKey).Result()
	if err != nil && err != redis.Nil {
		return AccountRiskState{}, err
	}
	if err == nil {
		currentBanUntil, _ = strconv.ParseInt(strings.TrimSpace(rawBan), 10, 64)
	}
	if currentBanUntil > 0 && currentBanUntil <= nowUnix {
		currentBanUntil = 0
		_ = s.client.Del(ctx, banKey, banStartKey).Err()
	}

	nextBanUntil := currentBanUntil
	if duration := s.accountRiskBanDuration(score); duration > 0 {
		banStartUnix := int64(0)
		rawBanStart, err := s.client.Get(ctx, banStartKey).Result()
		if err != nil && err != redis.Nil {
			return AccountRiskState{}, err
		}
		if err == nil {
			banStartUnix, _ = strconv.ParseInt(strings.TrimSpace(rawBanStart), 10, 64)
		}
		if banStartUnix <= 0 {
			banStartUnix = nowUnix
			if err := s.client.Set(ctx, banStartKey, strconv.FormatInt(banStartUnix, 10), s.accountRiskConfig.ScoreWindow+72*time.Hour).Err(); err != nil {
				return AccountRiskState{}, err
			}
		}

		candidate := banStartUnix + int64(duration.Seconds())
		if candidate > nextBanUntil {
			nextBanUntil = candidate
		}
		ttl := time.Unix(nextBanUntil, 0).Sub(s.now())
		if ttl > 0 {
			if err := s.client.Set(ctx, banKey, strconv.FormatInt(nextBanUntil, 10), ttl).Err(); err != nil {
				return AccountRiskState{}, err
			}
		}
	}
	if score <= 0 {
		_ = s.client.SRem(ctx, s.accountRiskIndexKey, nickname).Err()
	} else {
		_ = s.client.SAdd(ctx, s.accountRiskIndexKey, nickname).Err()
	}

	return AccountRiskState{
		Nickname: nickname,
		Score:    score,
		BanUntil: nextBanUntil,
	}, nil
}

func parseAccountRiskEventMember(member string) AccountRiskEvent {
	parts := strings.Split(strings.TrimSpace(member), ":")
	if len(parts) < 3 {
		return ""
	}
	return AccountRiskEvent(parts[len(parts)-1])
}

func (s *Store) accountRiskEventKey(nickname string) string {
	return s.accountRiskEventPrefix + strings.TrimSpace(nickname)
}

func (s *Store) accountRiskBanKey(nickname string) string {
	return s.accountRiskBanPrefix + strings.TrimSpace(nickname)
}

func (s *Store) accountRiskBanStartKey(nickname string) string {
	return s.accountRiskBanStartPrefix + strings.TrimSpace(nickname)
}

func (s *Store) accountRiskEventSeqKey(nickname string) string {
	return s.accountRiskEventSeqPrefix + strings.TrimSpace(nickname)
}

func (s *Store) accountRiskLastStaminaPurchaseKey(nickname string) string {
	return s.accountRiskLastPurchasePrefix + strings.TrimSpace(nickname)
}

func (s *Store) writeAccountRiskEventLog(ctx context.Context, item AccountRiskEventLog) {
	if s.accountRiskEventLogStore == nil {
		return
	}
	_ = s.accountRiskEventLogStore.WriteAccountRiskEvent(ctx, item)
}
