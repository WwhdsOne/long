package core

import (
	"context"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	staminaBaseMax             int64         = 50
	staminaClicksPerPoint      int64         = 20
	staminaMaxUpgradeLevels    int64         = 50
	staminaRecoverInterval     time.Duration = 5 * time.Minute
	staminaFullRecoverDuration time.Duration = staminaRecoverInterval * time.Duration(staminaBaseMax)
	staminaFirstFullBuyPrice   int64         = 200_000
)

type staminaSnapshot struct {
	Current       int64
	MaxLevel      int64
	Max           int64
	ClickProgress int64
	LastRecoverAt int64
}

type StaminaState struct {
	Current            int64 `json:"current"`
	MaxLevel           int64 `json:"maxLevel"`
	Max                int64 `json:"max"`
	ClickProgress      int64 `json:"clickProgress"`
	NextRecoverAt      int64 `json:"nextRecoverAt"`
	DailyFullBuyCount  int64 `json:"dailyFullBuyCount"`
	NextFullBuyPrice   int64 `json:"nextFullBuyPrice"`
	NextCapUpgradeCost int64 `json:"nextCapUpgradeCost"`
	RiskBanUntil       int64 `json:"riskBanUntil"`
}

func (s *Store) GetStaminaState(ctx context.Context, nickname string) (StaminaState, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return StaminaState{}, err
	}
	snapshot, err := s.loadAndRefreshStaminaSnapshot(ctx, normalizedNickname)
	if err != nil {
		return StaminaState{}, err
	}
	return s.composeStaminaState(ctx, normalizedNickname, snapshot)
}

func (s *Store) PurchaseStaminaFull(ctx context.Context, nickname string) (UserState, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return UserState{}, err
	}

	snapshot, err := s.loadAndRefreshStaminaSnapshot(ctx, normalizedNickname)
	if err != nil {
		return UserState{}, err
	}

	price, count, err := s.staminaFullBuyPrice(ctx, normalizedNickname)
	if err != nil {
		return UserState{}, err
	}
	currentStaminaBefore := snapshot.Current
	maxStamina := snapshot.Max
	nowUnix := s.now().Unix()
	_, banned, err := s.GetAccountRiskBanStatus(ctx, normalizedNickname)
	if err != nil {
		return UserState{}, err
	}
	if banned {
		s.writeStaminaPurchaseLog(ctx, StaminaPurchaseLog{
			Nickname:            normalizedNickname,
			PriceGold:           price,
			PurchasedAt:         nowUnix,
			Succeeded:           false,
			FailureReason:       "account_risk_banned",
			CurrentStamina:      currentStaminaBefore,
			MaxStamina:          maxStamina,
			DailyBuyCountBefore: count,
			DailyBuyCountAfter:  count,
		})
		return UserState{}, ErrAccountRiskBanned
	}
	resources, err := s.resourcesForNickname(ctx, normalizedNickname)
	if err != nil {
		return UserState{}, err
	}
	if resources.Gold < price {
		s.writeStaminaPurchaseLog(ctx, StaminaPurchaseLog{
			Nickname:            normalizedNickname,
			PriceGold:           price,
			PurchasedAt:         nowUnix,
			Succeeded:           false,
			FailureReason:       "shop_insufficient_gold",
			CurrentStamina:      currentStaminaBefore,
			MaxStamina:          maxStamina,
			DailyBuyCountBefore: count,
			DailyBuyCountAfter:  count,
		})
		return UserState{}, ErrShopInsufficientGold
	}

	snapshot.Current = snapshot.Max
	snapshot.ClickProgress = 0
	snapshot.LastRecoverAt = nowUnix

	pipe := s.client.TxPipeline()
	pipe.HIncrBy(ctx, s.resourceKey(normalizedNickname), "gold", -price)
	saveStaminaSnapshotToPipe(ctx, pipe, s.staminaKey(normalizedNickname), snapshot)
	pipe.Incr(ctx, s.staminaDailyBuyKey(normalizedNickname, s.now()))
	pipe.ExpireAt(ctx, s.staminaDailyBuyKey(normalizedNickname, s.now()), endOfDay(s.now()))
	if _, err := pipe.Exec(ctx); err != nil {
		return UserState{}, err
	}
	if err := s.MarkStaminaPurchase(ctx, normalizedNickname); err != nil {
		return UserState{}, err
	}
	s.writeStaminaPurchaseLog(ctx, StaminaPurchaseLog{
		Nickname:            normalizedNickname,
		PriceGold:           price,
		PurchasedAt:         nowUnix,
		Succeeded:           true,
		CurrentStamina:      currentStaminaBefore,
		MaxStamina:          maxStamina,
		DailyBuyCountBefore: count,
		DailyBuyCountAfter:  count + 1,
	})
	return s.GetUserState(ctx, normalizedNickname)
}

func (s *Store) UpgradeStaminaCap(ctx context.Context, nickname string) (UserState, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return UserState{}, err
	}
	if _, banned, err := s.GetAccountRiskBanStatus(ctx, normalizedNickname); err != nil {
		return UserState{}, err
	} else if banned {
		return UserState{}, ErrAccountRiskBanned
	}

	snapshot, err := s.loadAndRefreshStaminaSnapshot(ctx, normalizedNickname)
	if err != nil {
		return UserState{}, err
	}
	if snapshot.MaxLevel >= staminaMaxUpgradeLevels {
		return UserState{}, ErrStaminaMaxLevelReached
	}

	cost := staminaUpgradeCost(snapshot.MaxLevel + 1)
	resources, err := s.resourcesForNickname(ctx, normalizedNickname)
	if err != nil {
		return UserState{}, err
	}
	if resources.Gold < cost {
		return UserState{}, ErrShopInsufficientGold
	}

	nowUnix := s.now().Unix()
	snapshot.MaxLevel++
	snapshot.Max = staminaBaseMax + snapshot.MaxLevel
	if snapshot.Current == snapshot.Max-1 {
		snapshot.LastRecoverAt = nowUnix
	}

	pipe := s.client.TxPipeline()
	pipe.HIncrBy(ctx, s.resourceKey(normalizedNickname), "gold", -cost)
	saveStaminaSnapshotToPipe(ctx, pipe, s.staminaKey(normalizedNickname), snapshot)
	if _, err := pipe.Exec(ctx); err != nil {
		return UserState{}, err
	}
	return s.GetUserState(ctx, normalizedNickname)
}

func (s *Store) loadAndRefreshStaminaSnapshot(ctx context.Context, nickname string) (staminaSnapshot, error) {
	snapshot, err := s.loadStaminaSnapshot(ctx, nickname)
	if err != nil {
		return staminaSnapshot{}, err
	}
	next := snapshot
	changed := applyStaminaRecovery(&next, s.now())
	if changed {
		if err := s.saveStaminaSnapshot(ctx, nickname, next); err != nil {
			return staminaSnapshot{}, err
		}
		return next, nil
	}
	return snapshot, nil
}

func (s *Store) loadStaminaSnapshot(ctx context.Context, nickname string) (staminaSnapshot, error) {
	values, err := s.client.HMGet(ctx, s.staminaKey(nickname),
		"current",
		"max_level",
		"max",
		"click_progress",
		"last_recover_at",
	).Result()
	if err != nil {
		return staminaSnapshot{}, err
	}
	maxLevel := int64Value(values, 1)
	maxValue := int64Value(values, 2)
	if maxValue <= 0 {
		maxValue = staminaBaseMax + maxLevel
	}
	if maxLevel < 0 {
		maxLevel = 0
	}
	if maxLevel > staminaMaxUpgradeLevels {
		maxLevel = staminaMaxUpgradeLevels
	}
	current := int64Value(values, 0)
	if current <= 0 && values[0] == nil {
		current = maxValue
	}
	if current < 0 {
		current = 0
	}
	if current > maxValue {
		current = maxValue
	}
	lastRecoverAt := int64Value(values, 4)
	if lastRecoverAt <= 0 {
		lastRecoverAt = s.now().Unix()
	}
	return staminaSnapshot{
		Current:       current,
		MaxLevel:      maxLevel,
		Max:           maxValue,
		ClickProgress: maxInt64(0, int64Value(values, 3)),
		LastRecoverAt: lastRecoverAt,
	}, nil
}

func (s *Store) saveStaminaSnapshot(ctx context.Context, nickname string, snapshot staminaSnapshot) error {
	return s.client.HSet(ctx, s.staminaKey(nickname), staminaSnapshotValues(snapshot)).Err()
}

func saveStaminaSnapshotToPipe(ctx context.Context, pipe redis.Pipeliner, key string, snapshot staminaSnapshot) {
	pipe.HSet(ctx, key, staminaSnapshotValues(snapshot))
}

func staminaSnapshotValues(snapshot staminaSnapshot) map[string]any {
	return map[string]any{
		"current":         strconv.FormatInt(snapshot.Current, 10),
		"max_level":       strconv.FormatInt(snapshot.MaxLevel, 10),
		"max":             strconv.FormatInt(snapshot.Max, 10),
		"click_progress":  strconv.FormatInt(snapshot.ClickProgress, 10),
		"last_recover_at": strconv.FormatInt(snapshot.LastRecoverAt, 10),
	}
}

func applyStaminaRecovery(snapshot *staminaSnapshot, now time.Time) bool {
	changed := false
	if snapshot.Max <= 0 {
		snapshot.Max = staminaBaseMax
		changed = true
	}
	if snapshot.Current >= snapshot.Max {
		snapshot.Current = snapshot.Max
		if snapshot.LastRecoverAt != now.Unix() {
			snapshot.LastRecoverAt = now.Unix()
			changed = true
		}
		return changed
	}
	if snapshot.LastRecoverAt <= 0 {
		snapshot.LastRecoverAt = now.Unix()
		return true
	}
	elapsed := now.Unix() - snapshot.LastRecoverAt
	if elapsed < int64(staminaRecoverInterval.Seconds()) {
		return changed
	}
	recovered := elapsed / int64(staminaRecoverInterval.Seconds())
	if recovered <= 0 {
		return changed
	}
	snapshot.Current = minInt64(snapshot.Max, snapshot.Current+recovered)
	snapshot.LastRecoverAt += recovered * int64(staminaRecoverInterval.Seconds())
	return true
}

func (s *Store) consumeManualStamina(ctx context.Context, nickname string) (StaminaState, bool, error) {
	snapshot, err := s.loadAndRefreshStaminaSnapshot(ctx, nickname)
	if err != nil {
		return StaminaState{}, false, err
	}
	hadStamina := snapshot.Current > 0
	if hadStamina {
		snapshot.ClickProgress++
		if snapshot.ClickProgress >= staminaClicksPerPoint {
			spent := snapshot.ClickProgress / staminaClicksPerPoint
			snapshot.ClickProgress %= staminaClicksPerPoint
			snapshot.Current = maxInt64(0, snapshot.Current-spent)
		}
		if snapshot.Current < snapshot.Max && snapshot.LastRecoverAt <= 0 {
			snapshot.LastRecoverAt = s.now().Unix()
		}
		if err := s.saveStaminaSnapshot(ctx, nickname, snapshot); err != nil {
			return StaminaState{}, false, err
		}
	}
	state, err := s.composeStaminaState(ctx, nickname, snapshot)
	return state, hadStamina, err
}

func (s *Store) composeStaminaState(ctx context.Context, nickname string, snapshot staminaSnapshot) (StaminaState, error) {
	buyCount, err := s.staminaDailyBuyCount(ctx, nickname)
	if err != nil {
		return StaminaState{}, err
	}
	banUntil, _, err := s.GetAccountRiskBanStatus(ctx, nickname)
	if err != nil {
		return StaminaState{}, err
	}
	state := StaminaState{
		Current:            snapshot.Current,
		MaxLevel:           snapshot.MaxLevel,
		Max:                snapshot.Max,
		ClickProgress:      snapshot.ClickProgress,
		DailyFullBuyCount:  buyCount,
		NextFullBuyPrice:   staminaFullBuyPriceForCount(buyCount),
		NextCapUpgradeCost: staminaUpgradeCost(snapshot.MaxLevel + 1),
		RiskBanUntil:       banUntil,
	}
	if snapshot.Current < snapshot.Max {
		state.NextRecoverAt = snapshot.LastRecoverAt + int64(staminaRecoverInterval.Seconds())
	}
	if snapshot.MaxLevel >= staminaMaxUpgradeLevels {
		state.NextCapUpgradeCost = 0
	}
	return state, nil
}

func (s *Store) staminaDailyBuyCount(ctx context.Context, nickname string) (int64, error) {
	value, err := s.client.Get(ctx, s.staminaDailyBuyKey(nickname, s.now())).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	return int64FromString(value), nil
}

func (s *Store) staminaFullBuyPrice(ctx context.Context, nickname string) (int64, int64, error) {
	count, err := s.staminaDailyBuyCount(ctx, nickname)
	if err != nil {
		return 0, 0, err
	}
	return staminaFullBuyPriceForCount(count), count, nil
}

func staminaFullBuyPriceForCount(count int64) int64 {
	if count <= 0 {
		return staminaFirstFullBuyPrice
	}
	if count >= 62 {
		return math.MaxInt64
	}
	return staminaFirstFullBuyPrice << count
}

func staminaUpgradeCost(nextLevel int64) int64 {
	switch {
	case nextLevel <= 0:
		return 1000
	case nextLevel <= 10:
		return 1000
	case nextLevel <= 20:
		return 10000
	case nextLevel <= 30:
		return 100000
	case nextLevel <= 40:
		return 1000000
	case nextLevel <= staminaMaxUpgradeLevels:
		return 2000000
	default:
		return 0
	}
}

func endOfDay(now time.Time) time.Time {
	year, month, day := now.Date()
	location := now.Location()
	return time.Date(year, month, day, 23, 59, 59, 0, location)
}

func (s *Store) staminaKey(nickname string) string {
	return s.staminaPrefix + strings.TrimSpace(nickname)
}

func (s *Store) staminaDailyBuyKey(nickname string, now time.Time) string {
	return s.staminaDailyBuyPrefix + strings.TrimSpace(nickname) + ":" + now.Format("20060102")
}

func (s *Store) writeStaminaPurchaseLog(ctx context.Context, item StaminaPurchaseLog) {
	if s.staminaPurchaseLogStore == nil {
		return
	}
	_ = s.staminaPurchaseLogStore.WriteStaminaPurchaseLog(ctx, item)
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
