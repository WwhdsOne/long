package vote

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

const (
	equipmentSalvageGemValue = 1
	heroSalvageGemValue      = 1
	equipmentEnhanceCost     = 10
	heroAwakenCost           = 15
	criticalChanceGrowthStep = 0.2
)

type ForgeResult struct {
	Kind          string `json:"kind"`
	TargetID      string `json:"targetId"`
	TargetName    string `json:"targetName"`
	RewardSummary string `json:"rewardSummary"`
	GemsDelta     int64  `json:"gemsDelta"`
	RemainingGems int64  `json:"remainingGems"`
	Jackpot       bool   `json:"jackpot"`
}

type heroUpgrade struct {
	AwakenLevel                int
	BonusClicks                int64
	BonusCriticalChancePercent float64
	BonusCriticalCount         int64
	TraitValue                 int64
	PityCounter                int
}

func (s *Store) gemsKey(nickname string) string {
	return s.namespace + "user-gems:" + nickname
}

func (s *Store) heroUpgradeKey(nickname string, heroID string) string {
	return s.namespace + "user-hero-upgrade:" + nickname + ":" + heroID
}

func (s *Store) lastForgeResultKey(nickname string) string {
	return s.namespace + "user-last-forge-result:" + nickname
}

func (s *Store) setGems(ctx context.Context, nickname string, gems int64) error {
	if strings.TrimSpace(nickname) == "" {
		return nil
	}
	if gems < 0 {
		gems = 0
	}
	return s.client.Set(ctx, s.gemsKey(strings.TrimSpace(nickname)), strconv.FormatInt(gems, 10), 0).Err()
}

func (s *Store) gemsForNickname(ctx context.Context, nickname string) (int64, error) {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return 0, nil
	}

	value, err := s.client.Get(ctx, s.gemsKey(nickname)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, err
	}

	return int64FromString(value), nil
}

func (s *Store) getHeroUpgrade(ctx context.Context, nickname string, heroID string) (heroUpgrade, error) {
	if strings.TrimSpace(nickname) == "" || strings.TrimSpace(heroID) == "" {
		return heroUpgrade{}, nil
	}

	values, err := s.client.HGetAll(ctx, s.heroUpgradeKey(nickname, heroID)).Result()
	if err != nil {
		return heroUpgrade{}, err
	}
	if len(values) == 0 {
		return heroUpgrade{}, nil
	}

	return heroUpgrade{
		AwakenLevel:                int(int64FromString(values["awaken_level"])),
		BonusClicks:                int64FromString(values["clicks_delta"]),
		BonusCriticalChancePercent: float64FromString(values["critical_chance_delta"]),
		BonusCriticalCount:         int64FromString(values["critical_count_delta"]),
	}, nil
}

func (s *Store) lastForgeResultForNickname(ctx context.Context, nickname string) (*ForgeResult, error) {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return nil, nil
	}

	value, err := s.client.Get(ctx, s.lastForgeResultKey(nickname)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	result := &ForgeResult{}
	if err := sonic.Unmarshal([]byte(value), result); err != nil {
		return nil, nil
	}
	return result, nil
}

func setLastForgeResultOnPipeline(ctx context.Context, pipe redis.Pipeliner, key string, result *ForgeResult) {
	if result == nil || strings.TrimSpace(key) == "" {
		return
	}
	encoded, err := sonic.Marshal(result)
	if err != nil {
		return
	}
	pipe.Set(ctx, key, string(encoded), 0)
}

func statGrowthBase(clicks int64, critCount int64, critChance float64) int64 {
	total := float64(clicks+critCount) + critChance
	if total <= 0 {
		return 1
	}
	return maxInt64(int64(ceilFloat(total/4)), 1)
}

func formatFloatForRedis(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func formatPercentage(value float64) string {
	return strconv.FormatFloat(roundToDecimals(value, 2), 'f', 2, 64)
}

func (s *Store) SalvageEquipment(ctx context.Context, nickname string, itemID string, quantity int64) (State, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return State{}, err
	}
	if quantity <= 0 {
		return State{}, ErrInvalidQuantity
	}

	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return State{}, ErrEquipmentNotFound
	}

	definition, err := s.getEquipmentDefinition(ctx, itemID)
	if err != nil {
		return State{}, err
	}

	ownedQuantity, err := s.client.HGet(ctx, s.inventoryKey(normalizedNickname), itemID).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return State{}, ErrEquipmentNotOwned
		}
		return State{}, err
	}
	if ownedQuantity <= 0 {
		return State{}, ErrEquipmentNotOwned
	}

	protectedCount := int64(0)
	equippedItemID, err := s.client.HGet(ctx, s.loadoutKey(normalizedNickname), definition.Slot).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return State{}, err
	}
	if strings.TrimSpace(equippedItemID) == itemID {
		protectedCount = 1
	}

	if quantity > ownedQuantity-protectedCount {
		return State{}, ErrEquipmentNotEnough
	}

	currentGems, err := s.gemsForNickname(ctx, normalizedNickname)
	if err != nil {
		return State{}, err
	}
	remainingGems := currentGems + quantity*equipmentSalvageGemValue
	forgeResult := &ForgeResult{
		Kind:          "equipment_salvage",
		TargetID:      itemID,
		TargetName:    definition.Name,
		RewardSummary: "分解装备，获得原石",
		GemsDelta:     quantity * equipmentSalvageGemValue,
		RemainingGems: remainingGems,
	}

	now := s.now().Unix()
	pipe := s.client.TxPipeline()
	pipe.HIncrBy(ctx, s.inventoryKey(normalizedNickname), itemID, -quantity)
	pipe.Set(ctx, s.gemsKey(normalizedNickname), strconv.FormatInt(remainingGems, 10), 0)
	setLastForgeResultOnPipeline(ctx, pipe, s.lastForgeResultKey(normalizedNickname), forgeResult)
	pipe.ZAdd(ctx, s.playerIndexKey, redis.Z{
		Score:  float64(now),
		Member: normalizedNickname,
	})
	if _, err := pipe.Exec(ctx); err != nil {
		return State{}, err
	}

	return s.GetState(ctx, normalizedNickname)
}

func (s *Store) SalvageHero(ctx context.Context, nickname string, heroID string, quantity int64) (State, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return State{}, err
	}
	if quantity <= 0 {
		return State{}, ErrInvalidQuantity
	}

	heroID = strings.TrimSpace(heroID)
	if heroID == "" {
		return State{}, ErrHeroNotFound
	}

	definition, err := s.getHeroDefinition(ctx, heroID)
	if err != nil {
		return State{}, err
	}

	ownedQuantity, err := s.client.HGet(ctx, s.heroInventoryKey(normalizedNickname), heroID).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return State{}, ErrHeroNotOwned
		}
		return State{}, err
	}
	if ownedQuantity <= 0 {
		return State{}, ErrHeroNotOwned
	}

	protectedCount := int64(0)
	activeHeroID, err := s.client.Get(ctx, s.activeHeroKey(normalizedNickname)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return State{}, err
	}
	if strings.TrimSpace(activeHeroID) == heroID {
		protectedCount = 1
	}

	if quantity > ownedQuantity-protectedCount {
		return State{}, ErrHeroNotEnough
	}

	currentGems, err := s.gemsForNickname(ctx, normalizedNickname)
	if err != nil {
		return State{}, err
	}
	remainingGems := currentGems + quantity*heroSalvageGemValue
	forgeResult := &ForgeResult{
		Kind:          "hero_salvage",
		TargetID:      heroID,
		TargetName:    definition.Name,
		RewardSummary: "分解重复英雄，获得原石",
		GemsDelta:     quantity * heroSalvageGemValue,
		RemainingGems: remainingGems,
	}

	now := s.now().Unix()
	pipe := s.client.TxPipeline()
	pipe.HIncrBy(ctx, s.heroInventoryKey(normalizedNickname), heroID, -quantity)
	pipe.Set(ctx, s.gemsKey(normalizedNickname), strconv.FormatInt(remainingGems, 10), 0)
	setLastForgeResultOnPipeline(ctx, pipe, s.lastForgeResultKey(normalizedNickname), forgeResult)
	pipe.ZAdd(ctx, s.playerIndexKey, redis.Z{
		Score:  float64(now),
		Member: normalizedNickname,
	})
	if _, err := pipe.Exec(ctx); err != nil {
		return State{}, err
	}

	return s.GetState(ctx, normalizedNickname)
}

func (s *Store) EnhanceEquipment(ctx context.Context, nickname string, itemID string) (State, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return State{}, err
	}

	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return State{}, ErrEquipmentNotFound
	}
	definition, err := s.getEquipmentDefinition(ctx, itemID)
	if err != nil {
		return State{}, err
	}

	ownedQuantity, err := s.client.HGet(ctx, s.inventoryKey(normalizedNickname), itemID).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return State{}, ErrEquipmentNotOwned
		}
		return State{}, err
	}
	if ownedQuantity <= 0 {
		return State{}, ErrEquipmentNotOwned
	}

	currentGems, err := s.gemsForNickname(ctx, normalizedNickname)
	if err != nil {
		return State{}, err
	}
	if currentGems < equipmentEnhanceCost {
		return State{}, ErrGemsNotEnough
	}

	upgrade, err := s.getEquipmentUpgrade(ctx, normalizedNickname, itemID)
	if err != nil {
		return State{}, err
	}
	if definition.EnhanceCap > 0 && upgrade.EnhanceLevel >= definition.EnhanceCap {
		return State{}, ErrEquipmentMaxEnhance
	}

	rewardSummary := applyEquipmentEnhance(&upgrade, definition, s.roll)
	remainingGems := currentGems - equipmentEnhanceCost
	forgeResult := &ForgeResult{
		Kind:          "equipment_enhance",
		TargetID:      itemID,
		TargetName:    definition.Name,
		RewardSummary: rewardSummary,
		GemsDelta:     -equipmentEnhanceCost,
		RemainingGems: remainingGems,
	}

	now := s.now().Unix()
	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, s.upgradeKey(normalizedNickname, itemID), map[string]any{
		"enhance_level":         strconv.Itoa(upgrade.EnhanceLevel),
		"clicks_delta":          strconv.FormatInt(upgrade.BonusClicks, 10),
		"critical_chance_delta": formatFloatForRedis(upgrade.BonusCriticalChancePercent),
		"critical_count_delta":  strconv.FormatInt(upgrade.BonusCriticalCount, 10),
	})
	pipe.Set(ctx, s.gemsKey(normalizedNickname), strconv.FormatInt(remainingGems, 10), 0)
	setLastForgeResultOnPipeline(ctx, pipe, s.lastForgeResultKey(normalizedNickname), forgeResult)
	pipe.ZAdd(ctx, s.playerIndexKey, redis.Z{
		Score:  float64(now),
		Member: normalizedNickname,
	})
	if _, err := pipe.Exec(ctx); err != nil {
		return State{}, err
	}

	return s.GetState(ctx, normalizedNickname)
}

func applyEquipmentEnhance(upgrade *equipmentUpgrade, definition EquipmentDefinition, roll func(int) int) string {
	growth := statGrowthBase(definition.BonusClicks+upgrade.BonusClicks, definition.BonusCriticalCount+upgrade.BonusCriticalCount, definition.BonusCriticalChancePercent+upgrade.BonusCriticalChancePercent)
	upgrade.EnhanceLevel++
	upgrade.StarLevel = upgrade.EnhanceLevel

	switch roll(3) {
	case 0:
		upgrade.BonusClicks += growth
		return "点击 +" + strconv.FormatInt(growth, 10)
	case 1:
		upgrade.BonusCriticalCount += growth
		return "暴击 +" + strconv.FormatInt(growth, 10)
	default:
		upgrade.BonusCriticalChancePercent += criticalChanceGrowthStep
		return "暴击率 +" + formatPercentage(criticalChanceGrowthStep) + "%"
	}
}

func (s *Store) AwakenHero(ctx context.Context, nickname string, heroID string) (State, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return State{}, err
	}

	heroID = strings.TrimSpace(heroID)
	if heroID == "" {
		return State{}, ErrHeroNotFound
	}
	definition, err := s.getHeroDefinition(ctx, heroID)
	if err != nil {
		return State{}, err
	}

	ownedQuantity, err := s.client.HGet(ctx, s.heroInventoryKey(normalizedNickname), heroID).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return State{}, ErrHeroNotOwned
		}
		return State{}, err
	}
	if ownedQuantity <= 0 {
		return State{}, ErrHeroNotOwned
	}

	currentGems, err := s.gemsForNickname(ctx, normalizedNickname)
	if err != nil {
		return State{}, err
	}
	if currentGems < heroAwakenCost {
		return State{}, ErrGemsNotEnough
	}

	upgrade, err := s.getHeroUpgrade(ctx, normalizedNickname, heroID)
	if err != nil {
		return State{}, err
	}
	if definition.AwakenCap > 0 && upgrade.AwakenLevel >= definition.AwakenCap {
		return State{}, ErrHeroMaxAwaken
	}

	rewardSummary := applyHeroAwaken(&upgrade, definition, s.roll)
	remainingGems := currentGems - heroAwakenCost
	forgeResult := &ForgeResult{
		Kind:          "hero_awaken",
		TargetID:      heroID,
		TargetName:    definition.Name,
		RewardSummary: rewardSummary,
		GemsDelta:     -heroAwakenCost,
		RemainingGems: remainingGems,
	}

	now := s.now().Unix()
	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, s.heroUpgradeKey(normalizedNickname, heroID), map[string]any{
		"awaken_level":          strconv.Itoa(upgrade.AwakenLevel),
		"clicks_delta":          strconv.FormatInt(upgrade.BonusClicks, 10),
		"critical_chance_delta": formatFloatForRedis(upgrade.BonusCriticalChancePercent),
		"critical_count_delta":  strconv.FormatInt(upgrade.BonusCriticalCount, 10),
	})
	pipe.Set(ctx, s.gemsKey(normalizedNickname), strconv.FormatInt(remainingGems, 10), 0)
	setLastForgeResultOnPipeline(ctx, pipe, s.lastForgeResultKey(normalizedNickname), forgeResult)
	pipe.ZAdd(ctx, s.playerIndexKey, redis.Z{
		Score:  float64(now),
		Member: normalizedNickname,
	})
	if _, err := pipe.Exec(ctx); err != nil {
		return State{}, err
	}

	return s.GetState(ctx, normalizedNickname)
}

func applyHeroAwaken(upgrade *heroUpgrade, definition HeroDefinition, roll func(int) int) string {
	growth := statGrowthBase(definition.BonusClicks+upgrade.BonusClicks, definition.BonusCriticalCount+upgrade.BonusCriticalCount, definition.BonusCriticalChancePercent+upgrade.BonusCriticalChancePercent)
	upgrade.AwakenLevel++

	switch roll(3) {
	case 0:
		upgrade.BonusClicks += growth
		return "点击 +" + strconv.FormatInt(growth, 10)
	case 1:
		upgrade.BonusCriticalCount += growth
		return "暴击 +" + strconv.FormatInt(growth, 10)
	default:
		upgrade.BonusCriticalChancePercent += criticalChanceGrowthStep
		return "暴击率 +" + formatPercentage(criticalChanceGrowthStep) + "%"
	}
}
