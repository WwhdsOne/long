package vote

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"long/internal/config"
	nicknamefilter "long/internal/nickname"
)

// 错误定义
var ErrButtonNotFound = errors.New("button not found")
var ErrInvalidNickname = errors.New("invalid nickname")
var ErrSensitiveNickname = errors.New("sensitive nickname")
var ErrEquipmentNotFound = errors.New("equipment not found")
var ErrEquipmentNotOwned = errors.New("equipment not owned")

const (
	bossStatusActive   = "active"
	bossStatusDefeated = "defeated"
)

var loadoutSlots = []string{"weapon", "armor", "accessory"}

// Button 按钮数据结构，返回给前端和 SSE 客户端
type Button struct {
	Key       string `json:"key"`
	RedisKey  string `json:"redisKey"`
	Label     string `json:"label"`
	Count     int64  `json:"count"`
	Sort      int    `json:"sort"`
	Enabled   bool   `json:"enabled"`
	ImagePath string `json:"imagePath,omitempty"`
	ImageAlt  string `json:"imageAlt,omitempty"`
}

// UserStats 用户统计信息
type UserStats struct {
	Nickname   string `json:"nickname"`
	ClickCount int64  `json:"clickCount"`
}

// LeaderboardEntry 排行榜条目
type LeaderboardEntry struct {
	Rank       int    `json:"rank"`
	Nickname   string `json:"nickname"`
	ClickCount int64  `json:"clickCount"`
}

// Boss 世界 Boss 状态
type Boss struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	MaxHP      int64  `json:"maxHp"`
	CurrentHP  int64  `json:"currentHp"`
	StartedAt  int64  `json:"startedAt,omitempty"`
	DefeatedAt int64  `json:"defeatedAt,omitempty"`
}

// BossLeaderboardEntry Boss 伤害榜
type BossLeaderboardEntry struct {
	Rank     int    `json:"rank"`
	Nickname string `json:"nickname"`
	Damage   int64  `json:"damage"`
}

// BossUserStats 用户对当前 Boss 的伤害
type BossUserStats struct {
	Nickname string `json:"nickname"`
	Damage   int64  `json:"damage"`
}

// EquipmentDefinition 装备模板
type EquipmentDefinition struct {
	ItemID                     string `json:"itemId"`
	Name                       string `json:"name"`
	Slot                       string `json:"slot"`
	BonusClicks                int64  `json:"bonusClicks"`
	BonusCriticalChancePercent int    `json:"bonusCriticalChancePercent"`
	BonusCriticalCount         int64  `json:"bonusCriticalCount"`
}

// InventoryItem 背包道具
type InventoryItem struct {
	ItemID                     string `json:"itemId"`
	Name                       string `json:"name"`
	Slot                       string `json:"slot"`
	Quantity                   int64  `json:"quantity"`
	BonusClicks                int64  `json:"bonusClicks"`
	BonusCriticalChancePercent int    `json:"bonusCriticalChancePercent"`
	BonusCriticalCount         int64  `json:"bonusCriticalCount"`
	Equipped                   bool   `json:"equipped"`
}

// Loadout 已穿戴装备
type Loadout struct {
	Weapon    *InventoryItem `json:"weapon,omitempty"`
	Armor     *InventoryItem `json:"armor,omitempty"`
	Accessory *InventoryItem `json:"accessory,omitempty"`
}

// CombatStats 当前生效的点击战斗属性
type CombatStats struct {
	BaseIncrement         int64 `json:"baseIncrement"`
	BonusClicks           int64 `json:"bonusClicks"`
	EffectiveIncrement    int64 `json:"effectiveIncrement"`
	NormalDamage          int64 `json:"normalDamage"`
	CriticalDamage        int64 `json:"criticalDamage"`
	CriticalChancePercent int   `json:"criticalChancePercent"`
	CriticalCount         int64 `json:"criticalCount"`
}

// Reward 最近一次掉落
type Reward struct {
	BossID    string `json:"bossId"`
	BossName  string `json:"bossName"`
	ItemID    string `json:"itemId"`
	ItemName  string `json:"itemName"`
	GrantedAt int64  `json:"grantedAt"`
}

// BossLootEntry Boss 掉落池条目
type BossLootEntry struct {
	ItemID                     string `json:"itemId"`
	ItemName                   string `json:"itemName"`
	Slot                       string `json:"slot"`
	Weight                     int64  `json:"weight"`
	BonusClicks                int64  `json:"bonusClicks"`
	BonusCriticalChancePercent int    `json:"bonusCriticalChancePercent"`
	BonusCriticalCount         int64  `json:"bonusCriticalCount"`
}

// Snapshot 公共实时状态，广播给所有连接的客户端
type Snapshot struct {
	Buttons         []Button               `json:"buttons"`
	Leaderboard     []LeaderboardEntry     `json:"leaderboard"`
	Boss            *Boss                  `json:"boss,omitempty"`
	BossLeaderboard []BossLeaderboardEntry `json:"bossLeaderboard"`
	BossLoot        []BossLootEntry        `json:"bossLoot"`
}

// State 完整状态，包含个人统计与玩法状态
type State struct {
	Buttons         []Button               `json:"buttons"`
	Leaderboard     []LeaderboardEntry     `json:"leaderboard"`
	UserStats       *UserStats             `json:"userStats,omitempty"`
	Boss            *Boss                  `json:"boss,omitempty"`
	BossLeaderboard []BossLeaderboardEntry `json:"bossLeaderboard"`
	BossLoot        []BossLootEntry        `json:"bossLoot"`
	MyBossStats     *BossUserStats         `json:"myBossStats,omitempty"`
	Inventory       []InventoryItem        `json:"inventory"`
	Loadout         Loadout                `json:"loadout"`
	CombatStats     CombatStats            `json:"combatStats"`
	LastReward      *Reward                `json:"lastReward,omitempty"`
}

// ClickResult 点击结果，包含更新后的增量与状态摘要
type ClickResult struct {
	Button          Button                 `json:"button"`
	Delta           int64                  `json:"delta"`
	Critical        bool                   `json:"critical"`
	UserStats       UserStats              `json:"userStats"`
	Boss            *Boss                  `json:"boss,omitempty"`
	BossLeaderboard []BossLeaderboardEntry `json:"bossLeaderboard,omitempty"`
	MyBossStats     *BossUserStats         `json:"myBossStats,omitempty"`
	LastReward      *Reward                `json:"lastReward,omitempty"`
}

// StoreOptions 暴击机制配置
type StoreOptions struct {
	CriticalChancePercent int
	CriticalCount         int64
}

// buttonFallback 按钮回退数据（用于图片等元数据）
type buttonFallback struct {
	Label     string
	ImagePath string
	ImageAlt  string
}

// Store Redis 投票存储，管理按钮列表、点击计数、Boss 与装备状态
type Store struct {
	client             redis.UniversalClient
	prefix             string
	namespace          string
	userPrefix         string
	leaderboardKey     string
	bossCurrentKey     string
	bossHistoryKey     string
	bossHistoryPrefix  string
	equipmentDefPrefix string
	inventoryPrefix    string
	loadoutPrefix      string
	lastRewardPrefix   string
	fallbacks          map[string]buttonFallback
	critical           StoreOptions
	roll               func(int) int
	validator          interface{ Validate(string) error }
}

// hashFields Redis Hash 中存储的字段列表
var hashFields = []string{
	"label",
	"count",
	"sort",
	"enabled",
	"image_path",
	"image_alt",
}

// NewStore 创建 Redis 投票存储实例
func NewStore(client redis.UniversalClient, prefix string, options StoreOptions, validator interface{ Validate(string) error }) *Store {
	namespace := deriveNamespace(prefix)

	return &Store{
		client:             client,
		prefix:             prefix,
		namespace:          namespace,
		userPrefix:         namespace + "user:",
		leaderboardKey:     namespace + "leaderboard",
		bossCurrentKey:     namespace + "boss:current",
		bossHistoryKey:     namespace + "boss:history",
		bossHistoryPrefix:  namespace + "boss:history:",
		equipmentDefPrefix: namespace + "equip:def:",
		inventoryPrefix:    namespace + "user-inventory:",
		loadoutPrefix:      namespace + "user-loadout:",
		lastRewardPrefix:   namespace + "user-last-reward:",
		fallbacks: map[string]buttonFallback{
			"wechat-pity": {
				ImagePath: "/images/emojipedia-wechat-whimper.png",
				ImageAlt:  "微信可怜表情",
			},
		},
		critical: options,
		roll: func(limit int) int {
			return rand.IntN(limit)
		},
		validator: validator,
	}
}

// ValidateNickname checks whether the provided nickname is usable.
func (s *Store) ValidateNickname(_ context.Context, nickname string) error {
	_, err := s.validatedNickname(nickname)
	return err
}

// GetSnapshot 获取公共快照（按钮列表 + 公共排行榜 + Boss 状态）
func (s *Store) GetSnapshot(ctx context.Context) (Snapshot, error) {
	buttons, err := s.ListButtons(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	leaderboard, err := s.ListLeaderboard(ctx, 10)
	if err != nil {
		return Snapshot{}, err
	}

	boss, err := s.currentBoss(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	bossLeaderboard := []BossLeaderboardEntry{}
	bossLoot := []BossLootEntry{}
	if boss != nil {
		bossLeaderboard, err = s.ListBossLeaderboard(ctx, boss.ID, 10)
		if err != nil {
			return Snapshot{}, err
		}
		bossLoot, err = s.loadBossLoot(ctx, boss.ID)
		if err != nil {
			return Snapshot{}, err
		}
	}

	return Snapshot{
		Buttons:         buttons,
		Leaderboard:     leaderboard,
		Boss:            boss,
		BossLeaderboard: bossLeaderboard,
		BossLoot:        bossLoot,
	}, nil
}

// GetState 获取完整状态（公共快照 + 个人统计）
func (s *Store) GetState(ctx context.Context, nickname string) (State, error) {
	snapshot, err := s.GetSnapshot(ctx)
	if err != nil {
		return State{}, err
	}

	state := State{
		Buttons:         snapshot.Buttons,
		Leaderboard:     snapshot.Leaderboard,
		Boss:            snapshot.Boss,
		BossLeaderboard: snapshot.BossLeaderboard,
		BossLoot:        snapshot.BossLoot,
		Inventory:       []InventoryItem{},
		Loadout:         Loadout{},
		CombatStats:     s.baseCombatStats(),
	}

	trimmedNickname, hasNickname := normalizeNickname(nickname)
	if !hasNickname {
		return state, nil
	}

	normalizedNickname, err := s.validatedNickname(trimmedNickname)
	if err != nil {
		return State{}, err
	}

	userStats, err := s.GetUserStats(ctx, normalizedNickname)
	if err != nil {
		return State{}, err
	}
	state.UserStats = &userStats

	quantities, err := s.inventoryQuantities(ctx, normalizedNickname)
	if err != nil {
		return State{}, err
	}

	loadout, equipped, err := s.loadoutForNickname(ctx, normalizedNickname, quantities)
	if err != nil {
		return State{}, err
	}
	state.Loadout = loadout

	inventory, err := s.inventoryForNickname(ctx, quantities, equipped)
	if err != nil {
		return State{}, err
	}
	state.Inventory = inventory

	combatStats, err := s.combatStatsForNickname(ctx, normalizedNickname, loadout)
	if err != nil {
		return State{}, err
	}
	state.CombatStats = combatStats

	lastReward, err := s.lastRewardForNickname(ctx, normalizedNickname)
	if err != nil {
		return State{}, err
	}
	state.LastReward = lastReward

	if state.Boss != nil {
		myBossStats, err := s.bossStatsForNickname(ctx, state.Boss.ID, normalizedNickname)
		if err != nil {
			return State{}, err
		}
		state.MyBossStats = myBossStats
	}

	return state, nil
}

// ListButtons 扫描 Redis，过滤禁用按钮，按排序权重返回
func (s *Store) ListButtons(ctx context.Context) ([]Button, error) {
	keys, err := s.scanKeys(ctx)
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return []Button{}, nil
	}

	pipe := s.client.Pipeline()
	cmds := make([]*redis.SliceCmd, len(keys))
	for index, redisKey := range keys {
		cmds[index] = pipe.HMGet(ctx, redisKey, hashFields...)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	buttons := make([]Button, 0, len(keys))
	for index, redisKey := range keys {
		button := s.normalizeButton(redisKey, cmds[index].Val())
		if button.Enabled {
			buttons = append(buttons, button)
		}
	}

	slices.SortFunc(buttons, func(left, right Button) int {
		if left.Sort == right.Sort {
			return strings.Compare(left.Key, right.Key)
		}
		if left.Sort < right.Sort {
			return -1
		}
		return 1
	})

	return buttons, nil
}

// ClickButton 处理按钮点击。无活动 Boss 时只更新投票；有活动 Boss 时附加结算伤害与掉落。
func (s *Store) ClickButton(ctx context.Context, slug string, nickname string) (ClickResult, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return ClickResult{}, err
	}

	redisKey := s.prefix + slug
	currentValues, err := s.client.HMGet(ctx, redisKey, hashFields...).Result()
	if err != nil {
		return ClickResult{}, err
	}

	current := s.normalizeButton(redisKey, currentValues)
	if current.Key == slug && current.Label == slug && current.Count == 0 && current.Sort == 0 && !current.Enabled {
		exists, existsErr := s.client.Exists(ctx, redisKey).Result()
		if existsErr != nil {
			return ClickResult{}, existsErr
		}
		if exists == 0 {
			return ClickResult{}, ErrButtonNotFound
		}
	}
	if !current.Enabled {
		return ClickResult{}, ErrButtonNotFound
	}

	delta, critical, err := s.nextIncrement(ctx, normalizedNickname)
	if err != nil {
		return ClickResult{}, err
	}

	boss, err := s.currentBoss(ctx)
	if err != nil {
		return ClickResult{}, err
	}

	var result ClickResult
	if boss == nil || boss.Status != bossStatusActive {
		result, err = s.applyVoteOnlyClick(ctx, redisKey, normalizedNickname, delta, critical)
		if err != nil {
			return ClickResult{}, err
		}
	} else {
		result, err = s.applyBossClick(ctx, redisKey, normalizedNickname, delta, critical)
		if err != nil {
			return ClickResult{}, err
		}
	}

	state, err := s.GetState(ctx, normalizedNickname)
	if err != nil {
		return ClickResult{}, err
	}

	result.Boss = state.Boss
	result.BossLeaderboard = state.BossLeaderboard
	result.MyBossStats = state.MyBossStats
	result.LastReward = state.LastReward

	return result, nil
}

// EquipItem 穿戴一件装备。装备效果会影响平时点击与 Boss 伤害。
func (s *Store) EquipItem(ctx context.Context, nickname string, itemID string) (State, error) {
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

	quantity, err := s.client.HGet(ctx, s.inventoryKey(normalizedNickname), itemID).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return State{}, ErrEquipmentNotOwned
		}
		return State{}, err
	}
	if quantity <= 0 {
		return State{}, ErrEquipmentNotOwned
	}

	if err := s.client.HSet(ctx, s.loadoutKey(normalizedNickname), definition.Slot, itemID).Err(); err != nil {
		return State{}, err
	}

	return s.GetState(ctx, normalizedNickname)
}

// UnequipItem 卸下一件当前已穿戴的装备。
func (s *Store) UnequipItem(ctx context.Context, nickname string, itemID string) (State, error) {
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

	if err := s.client.HDel(ctx, s.loadoutKey(normalizedNickname), definition.Slot).Err(); err != nil {
		return State{}, err
	}

	return s.GetState(ctx, normalizedNickname)
}

// GetCurrentBoss 返回当前世界 Boss。
func (s *Store) GetCurrentBoss(ctx context.Context) (*Boss, error) {
	return s.currentBoss(ctx)
}

// ListBossLeaderboard 获取指定 Boss 的伤害榜。
func (s *Store) ListBossLeaderboard(ctx context.Context, bossID string, limit int64) ([]BossLeaderboardEntry, error) {
	if strings.TrimSpace(bossID) == "" {
		return []BossLeaderboardEntry{}, nil
	}
	if limit <= 0 {
		limit = 10
	}

	scores, err := s.client.ZRevRangeWithScores(ctx, s.bossDamageKey(bossID), 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	leaderboard := make([]BossLeaderboardEntry, 0, len(scores))
	for index, score := range scores {
		nickname, ok := score.Member.(string)
		if !ok || nickname == "" {
			continue
		}

		leaderboard = append(leaderboard, BossLeaderboardEntry{
			Rank:     index + 1,
			Nickname: nickname,
			Damage:   int64(score.Score),
		})
	}

	return leaderboard, nil
}

// EnsureDefaults 在 Redis 为空时初始化默认按钮
func (s *Store) EnsureDefaults(ctx context.Context, buttons []config.ButtonSeed) error {
	keys, err := s.scanKeys(ctx)
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return nil
	}

	pipe := s.client.Pipeline()
	for _, button := range buttons {
		values := map[string]any{
			"label":   button.Label,
			"count":   "0",
			"sort":    strconv.Itoa(button.Sort),
			"enabled": "1",
		}
		if button.ImagePath != "" {
			values["image_path"] = button.ImagePath
		}
		if button.ImageAlt != "" {
			values["image_alt"] = button.ImageAlt
		}
		pipe.HSet(ctx, s.prefix+button.Slug, values)
	}

	_, err = pipe.Exec(ctx)
	return err
}

// ListLeaderboard 获取排行榜前 N 名
func (s *Store) ListLeaderboard(ctx context.Context, limit int64) ([]LeaderboardEntry, error) {
	if limit <= 0 {
		limit = 10
	}

	scores, err := s.client.ZRevRangeWithScores(ctx, s.leaderboardKey, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	leaderboard := make([]LeaderboardEntry, 0, len(scores))
	for index, score := range scores {
		nickname, ok := score.Member.(string)
		if !ok || nickname == "" {
			continue
		}

		leaderboard = append(leaderboard, LeaderboardEntry{
			Rank:       index + 1,
			Nickname:   nickname,
			ClickCount: int64(score.Score),
		})
	}

	return leaderboard, nil
}

// GetUserStats 获取指定用户的统计信息
func (s *Store) GetUserStats(ctx context.Context, nickname string) (UserStats, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return UserStats{}, err
	}

	values, err := s.client.HMGet(ctx, s.userPrefix+normalizedNickname, "nickname", "click_count").Result()
	if err != nil {
		return UserStats{}, err
	}

	return UserStats{
		Nickname:   normalizedNickname,
		ClickCount: int64Value(values, 1),
	}, nil
}

func (s *Store) applyVoteOnlyClick(ctx context.Context, redisKey string, nickname string, delta int64, critical bool) (ClickResult, error) {
	pipe := s.client.TxPipeline()
	pipe.HIncrBy(ctx, redisKey, "count", delta)
	userCountCmd := pipe.HIncrBy(ctx, s.userPrefix+nickname, "click_count", delta)
	pipe.HSet(ctx, s.userPrefix+nickname, map[string]any{
		"nickname":   nickname,
		"updated_at": strconv.FormatInt(time.Now().Unix(), 10),
	})
	pipe.ZIncrBy(ctx, s.leaderboardKey, float64(delta), nickname)

	if _, err := pipe.Exec(ctx); err != nil {
		return ClickResult{}, err
	}

	updatedValues, err := s.client.HMGet(ctx, redisKey, hashFields...).Result()
	if err != nil {
		return ClickResult{}, err
	}

	return ClickResult{
		Button:   s.normalizeButton(redisKey, updatedValues),
		Delta:    delta,
		Critical: critical,
		UserStats: UserStats{
			Nickname:   nickname,
			ClickCount: userCountCmd.Val(),
		},
	}, nil
}

func (s *Store) applyBossClick(ctx context.Context, redisKey string, nickname string, delta int64, critical bool) (ClickResult, error) {
	for range 6 {
		var result ClickResult
		err := s.client.Watch(ctx, func(tx *redis.Tx) error {
			boss, err := s.currentBossFromCmdable(ctx, tx)
			if err != nil {
				return err
			}
			if boss == nil || boss.Status != bossStatusActive {
				baseResult, baseErr := s.applyVoteOnlyClick(ctx, redisKey, nickname, delta, critical)
				if baseErr != nil {
					return baseErr
				}
				result = baseResult
				return nil
			}

			updatedBoss := *boss
			updatedBoss.CurrentHP -= delta
			if updatedBoss.CurrentHP < 0 {
				updatedBoss.CurrentHP = 0
			}
			if updatedBoss.CurrentHP == 0 {
				updatedBoss.Status = bossStatusDefeated
				updatedBoss.DefeatedAt = time.Now().Unix()
			}

			var userCountCmd *redis.IntCmd
			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.HIncrBy(ctx, redisKey, "count", delta)
				userCountCmd = pipe.HIncrBy(ctx, s.userPrefix+nickname, "click_count", delta)
				pipe.HSet(ctx, s.userPrefix+nickname, map[string]any{
					"nickname":   nickname,
					"updated_at": strconv.FormatInt(time.Now().Unix(), 10),
				})
				pipe.ZIncrBy(ctx, s.leaderboardKey, float64(delta), nickname)
				pipe.ZIncrBy(ctx, s.bossDamageKey(boss.ID), float64(delta), nickname)
				bossFields := map[string]any{
					"id":         updatedBoss.ID,
					"name":       updatedBoss.Name,
					"status":     updatedBoss.Status,
					"max_hp":     strconv.FormatInt(updatedBoss.MaxHP, 10),
					"current_hp": strconv.FormatInt(updatedBoss.CurrentHP, 10),
				}
				if boss.StartedAt != 0 {
					bossFields["started_at"] = strconv.FormatInt(boss.StartedAt, 10)
				}
				if updatedBoss.DefeatedAt != 0 {
					bossFields["defeated_at"] = strconv.FormatInt(updatedBoss.DefeatedAt, 10)
				}
				pipe.HSet(ctx, s.bossCurrentKey, bossFields)
				return nil
			})
			if err != nil {
				return err
			}

			updatedValues, err := s.client.HMGet(ctx, redisKey, hashFields...).Result()
			if err != nil {
				return err
			}

			result = ClickResult{
				Button:   s.normalizeButton(redisKey, updatedValues),
				Delta:    delta,
				Critical: critical,
				UserStats: UserStats{
					Nickname:   nickname,
					ClickCount: userCountCmd.Val(),
				},
				Boss: &updatedBoss,
			}
			return nil
		}, s.bossCurrentKey)
		if err == nil {
			if result.Boss != nil && result.Boss.Status == bossStatusDefeated {
				if finalizeErr := s.finalizeBossKill(ctx, result.Boss); finalizeErr != nil {
					return ClickResult{}, finalizeErr
				}
			}
			return result, nil
		}
		if errors.Is(err, redis.TxFailedErr) {
			continue
		}
		return ClickResult{}, err
	}

	return ClickResult{}, redis.TxFailedErr
}

func (s *Store) finalizeBossKill(ctx context.Context, boss *Boss) error {
	if boss == nil || strings.TrimSpace(boss.ID) == "" {
		return nil
	}
	bossID := strings.TrimSpace(boss.ID)
	bossName := strings.TrimSpace(boss.Name)

	acquired, err := s.client.SetNX(ctx, s.bossRewardLockKey(bossID), "1", 0).Result()
	if err != nil || !acquired {
		return err
	}

	lootEntries, err := s.loadBossLoot(ctx, bossID)
	if err != nil {
		return err
	}
	if len(lootEntries) == 0 {
		return nil
	}

	participants, err := s.client.ZRevRangeWithScores(ctx, s.bossDamageKey(bossID), 0, -1).Result()
	if err != nil {
		return err
	}
	if len(participants) == 0 {
		return nil
	}

	pipe := s.client.Pipeline()
	now := time.Now().Unix()
	for _, participant := range participants {
		nickname, ok := participant.Member.(string)
		if !ok || nickname == "" || participant.Score <= 0 {
			continue
		}

		reward := s.chooseLoot(lootEntries)
		if reward == nil {
			continue
		}

		pipe.HIncrBy(ctx, s.inventoryKey(nickname), reward.ItemID, 1)
		pipe.HSet(ctx, s.lastRewardKey(nickname), map[string]any{
			"boss_id":    bossID,
			"boss_name":  bossName,
			"item_id":    reward.ItemID,
			"item_name":  reward.ItemName,
			"granted_at": strconv.FormatInt(now, 10),
		})
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (s *Store) chooseLoot(entries []BossLootEntry) *BossLootEntry {
	if len(entries) == 0 {
		return nil
	}

	totalWeight := 0
	for _, entry := range entries {
		if entry.Weight > 0 {
			totalWeight += int(entry.Weight)
		}
	}
	if totalWeight <= 0 {
		return nil
	}

	cursor := s.roll(totalWeight)
	running := 0
	for _, entry := range entries {
		if entry.Weight <= 0 {
			continue
		}
		running += int(entry.Weight)
		if cursor < running {
			selected := entry
			return &selected
		}
	}

	selected := entries[len(entries)-1]
	return &selected
}

func (s *Store) nextIncrement(ctx context.Context, nickname string) (int64, bool, error) {
	quantities, err := s.inventoryQuantities(ctx, nickname)
	if err != nil {
		return 0, false, err
	}
	loadout, _, err := s.loadoutForNickname(ctx, nickname, quantities)
	if err != nil {
		return 0, false, err
	}
	combatStats, err := s.combatStatsForNickname(ctx, nickname, loadout)
	if err != nil {
		return 0, false, err
	}

	delta := combatStats.NormalDamage
	if delta <= 0 {
		delta = 1
	}

	if combatStats.CriticalChancePercent <= 0 || combatStats.CriticalCount <= 1 {
		return delta, false, nil
	}

	if s.roll(100) < combatStats.CriticalChancePercent {
		return combatStats.CriticalDamage, true, nil
	}

	return delta, false, nil
}

func (s *Store) combatStatsForNickname(_ context.Context, _ string, loadout Loadout) (CombatStats, error) {
	stats := s.baseCombatStats()

	bonusClicks, bonusChance, bonusCount := loadoutBonuses(loadout)
	stats.BonusClicks = bonusClicks
	stats.CriticalChancePercent = clampInt(stats.CriticalChancePercent+bonusChance, 0, 100)
	stats.CriticalCount += bonusCount

	return deriveCombatStats(stats), nil
}

func (s *Store) baseCombatStats() CombatStats {
	return deriveCombatStats(CombatStats{
		BaseIncrement:         1,
		BonusClicks:           0,
		CriticalChancePercent: clampInt(s.critical.CriticalChancePercent, 0, 100),
		CriticalCount:         s.critical.CriticalCount,
	})
}

func loadoutBonuses(loadout Loadout) (int64, int, int64) {
	items := []*InventoryItem{loadout.Weapon, loadout.Armor, loadout.Accessory}
	var bonusClicks int64
	var bonusChance int
	var bonusCount int64
	for _, item := range items {
		if item == nil {
			continue
		}
		bonusClicks += item.BonusClicks
		bonusChance += item.BonusCriticalChancePercent
		bonusCount += item.BonusCriticalCount
	}
	return bonusClicks, bonusChance, bonusCount
}

func deriveCombatStats(stats CombatStats) CombatStats {
	if stats.BaseIncrement <= 0 {
		stats.BaseIncrement = 1
	}

	stats.EffectiveIncrement = stats.BaseIncrement + stats.BonusClicks
	if stats.EffectiveIncrement <= 0 {
		stats.EffectiveIncrement = 1
	}

	stats.NormalDamage = stats.EffectiveIncrement
	if stats.CriticalCount <= 1 {
		stats.CriticalCount = 1
	}

	stats.CriticalDamage = max(stats.NormalDamage+stats.CriticalCount-1, stats.NormalDamage)

	return stats
}

func (s *Store) currentBoss(ctx context.Context) (*Boss, error) {
	return s.currentBossFromCmdable(ctx, s.client)
}

func (s *Store) currentBossFromCmdable(ctx context.Context, client redis.Cmdable) (*Boss, error) {
	values, err := client.HGetAll(ctx, s.bossCurrentKey).Result()
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}

	return normalizeBoss(values), nil
}

func normalizeBoss(values map[string]string) *Boss {
	id := strings.TrimSpace(values["id"])
	name := strings.TrimSpace(values["name"])
	if id == "" && name == "" {
		return nil
	}

	return &Boss{
		ID:         id,
		Name:       name,
		Status:     strings.TrimSpace(values["status"]),
		MaxHP:      int64FromString(values["max_hp"]),
		CurrentHP:  int64FromString(values["current_hp"]),
		StartedAt:  int64FromString(values["started_at"]),
		DefeatedAt: int64FromString(values["defeated_at"]),
	}
}

func (s *Store) bossStatsForNickname(ctx context.Context, bossID string, nickname string) (*BossUserStats, error) {
	if strings.TrimSpace(bossID) == "" || strings.TrimSpace(nickname) == "" {
		return nil, nil
	}

	score, err := s.client.ZScore(ctx, s.bossDamageKey(bossID), nickname).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	return &BossUserStats{
		Nickname: nickname,
		Damage:   int64(score),
	}, nil
}

func (s *Store) getEquipmentDefinition(ctx context.Context, itemID string) (EquipmentDefinition, error) {
	values, err := s.client.HGetAll(ctx, s.equipmentKey(itemID)).Result()
	if err != nil {
		return EquipmentDefinition{}, err
	}
	if len(values) == 0 {
		return EquipmentDefinition{}, ErrEquipmentNotFound
	}

	return EquipmentDefinition{
		ItemID:                     itemID,
		Name:                       firstNonEmpty(strings.TrimSpace(values["name"]), itemID),
		Slot:                       strings.TrimSpace(values["slot"]),
		BonusClicks:                int64FromString(values["bonus_clicks"]),
		BonusCriticalChancePercent: int(int64FromString(values["bonus_critical_chance_percent"])),
		BonusCriticalCount:         int64FromString(values["bonus_critical_count"]),
	}, nil
}

func (s *Store) inventoryQuantities(ctx context.Context, nickname string) (map[string]int64, error) {
	values, err := s.client.HGetAll(ctx, s.inventoryKey(nickname)).Result()
	if err != nil {
		return nil, err
	}

	quantities := make(map[string]int64, len(values))
	for itemID, rawQuantity := range values {
		quantity := int64FromString(rawQuantity)
		if quantity <= 0 {
			continue
		}
		quantities[itemID] = quantity
	}

	return quantities, nil
}

func (s *Store) loadoutForNickname(ctx context.Context, nickname string, quantities map[string]int64) (Loadout, map[string]string, error) {
	values, err := s.client.HGetAll(ctx, s.loadoutKey(nickname)).Result()
	if err != nil {
		return Loadout{}, nil, err
	}

	loadout := Loadout{}
	equipped := make(map[string]string, len(values))
	for slot, itemID := range values {
		itemID = strings.TrimSpace(itemID)
		if itemID == "" {
			continue
		}

		definition, defErr := s.getEquipmentDefinition(ctx, itemID)
		if defErr != nil {
			continue
		}

		item := &InventoryItem{
			ItemID:                     definition.ItemID,
			Name:                       definition.Name,
			Slot:                       definition.Slot,
			Quantity:                   quantities[itemID],
			BonusClicks:                definition.BonusClicks,
			BonusCriticalChancePercent: definition.BonusCriticalChancePercent,
			BonusCriticalCount:         definition.BonusCriticalCount,
			Equipped:                   true,
		}

		equipped[itemID] = slot
		switch slot {
		case "weapon":
			loadout.Weapon = item
		case "armor":
			loadout.Armor = item
		case "accessory":
			loadout.Accessory = item
		}
	}

	return loadout, equipped, nil
}

func (s *Store) inventoryForNickname(ctx context.Context, quantities map[string]int64, equipped map[string]string) ([]InventoryItem, error) {
	if len(quantities) == 0 {
		return []InventoryItem{}, nil
	}

	items := make([]InventoryItem, 0, len(quantities))
	for itemID, quantity := range quantities {
		definition, err := s.getEquipmentDefinition(ctx, itemID)
		if err != nil {
			items = append(items, InventoryItem{
				ItemID:   itemID,
				Name:     itemID,
				Quantity: quantity,
				Equipped: equipped[itemID] != "",
			})
			continue
		}

		items = append(items, InventoryItem{
			ItemID:                     definition.ItemID,
			Name:                       definition.Name,
			Slot:                       definition.Slot,
			Quantity:                   quantity,
			BonusClicks:                definition.BonusClicks,
			BonusCriticalChancePercent: definition.BonusCriticalChancePercent,
			BonusCriticalCount:         definition.BonusCriticalCount,
			Equipped:                   equipped[itemID] != "",
		})
	}

	slices.SortFunc(items, func(left, right InventoryItem) int {
		if left.Slot == right.Slot {
			return strings.Compare(left.Name, right.Name)
		}
		return strings.Compare(left.Slot, right.Slot)
	})

	return items, nil
}

func (s *Store) lastRewardForNickname(ctx context.Context, nickname string) (*Reward, error) {
	values, err := s.client.HGetAll(ctx, s.lastRewardKey(nickname)).Result()
	if err != nil {
		return nil, err
	}
	if len(values) == 0 || strings.TrimSpace(values["item_id"]) == "" {
		return nil, nil
	}

	return &Reward{
		BossID:    strings.TrimSpace(values["boss_id"]),
		BossName:  strings.TrimSpace(values["boss_name"]),
		ItemID:    strings.TrimSpace(values["item_id"]),
		ItemName:  strings.TrimSpace(values["item_name"]),
		GrantedAt: int64FromString(values["granted_at"]),
	}, nil
}

func (s *Store) loadBossLoot(ctx context.Context, bossID string) ([]BossLootEntry, error) {
	if strings.TrimSpace(bossID) == "" {
		return []BossLootEntry{}, nil
	}

	entries, err := s.client.ZRangeWithScores(ctx, s.bossLootKey(bossID), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	loot := make([]BossLootEntry, 0, len(entries))
	for _, entry := range entries {
		itemID, ok := entry.Member.(string)
		if !ok || strings.TrimSpace(itemID) == "" {
			continue
		}

		definition, defErr := s.getEquipmentDefinition(ctx, itemID)
		if defErr != nil {
			loot = append(loot, BossLootEntry{
				ItemID: itemID,
				Weight: int64(entry.Score),
			})
			continue
		}

		loot = append(loot, BossLootEntry{
			ItemID:                     itemID,
			ItemName:                   definition.Name,
			Slot:                       definition.Slot,
			Weight:                     int64(entry.Score),
			BonusClicks:                definition.BonusClicks,
			BonusCriticalChancePercent: definition.BonusCriticalChancePercent,
			BonusCriticalCount:         definition.BonusCriticalCount,
		})
	}

	return loot, nil
}

// normalizeButton 将 Redis 数据转换为 Button 结构
func (s *Store) normalizeButton(redisKey string, values []any) Button {
	slug := strings.TrimPrefix(redisKey, s.prefix)
	fallback := s.fallbacks[slug]

	label := stringValue(values, 0)
	if label == "" {
		if fallback.Label != "" {
			label = fallback.Label
		} else {
			label = slug
		}
	}

	imagePath := stringValue(values, 4)
	if imagePath == "" {
		imagePath = fallback.ImagePath
	}

	imageAlt := stringValue(values, 5)
	if imageAlt == "" {
		imageAlt = fallback.ImageAlt
	}

	return Button{
		Key:       slug,
		RedisKey:  redisKey,
		Label:     label,
		Count:     int64Value(values, 1),
		Sort:      int(int64Value(values, 2)),
		Enabled:   stringValue(values, 3) != "0",
		ImagePath: imagePath,
		ImageAlt:  imageAlt,
	}
}

// scanKeys 扫描 Redis 中匹配前缀的所有键
func (s *Store) scanKeys(ctx context.Context) ([]string, error) {
	var (
		cursor uint64
		keys   []string
	)

	for {
		foundKeys, nextCursor, err := s.client.Scan(ctx, cursor, s.prefix+"*", 100).Result()
		if err != nil {
			return nil, err
		}

		keys = append(keys, foundKeys...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

func (s *Store) validatedNickname(nickname string) (string, error) {
	normalizedNickname, ok := normalizeNickname(nickname)
	if !ok {
		return "", ErrInvalidNickname
	}

	if s.validator != nil {
		if err := s.validator.Validate(normalizedNickname); err != nil {
			if errors.Is(err, nicknamefilter.ErrSensitiveNickname) {
				return "", ErrSensitiveNickname
			}
			return "", err
		}
	}

	return normalizedNickname, nil
}

func (s *Store) inventoryKey(nickname string) string {
	return s.inventoryPrefix + nickname
}

func (s *Store) loadoutKey(nickname string) string {
	return s.loadoutPrefix + nickname
}

func (s *Store) lastRewardKey(nickname string) string {
	return s.lastRewardPrefix + nickname
}

func (s *Store) equipmentKey(itemID string) string {
	return s.equipmentDefPrefix + itemID
}

func (s *Store) bossDamageKey(bossID string) string {
	return s.namespace + "boss:" + bossID + ":damage"
}

func (s *Store) bossLootKey(bossID string) string {
	return s.namespace + "boss:" + bossID + ":loot"
}

func (s *Store) bossRewardLockKey(bossID string) string {
	return s.namespace + "boss:" + bossID + ":reward-lock"
}

// deriveNamespace 从按钮前缀推导命名空间
func deriveNamespace(prefix string) string {
	if before, ok := strings.CutSuffix(prefix, "button:"); ok {
		return before
	}

	return prefix
}

// normalizeNickname 规范化昵称（去除首尾空格）
func normalizeNickname(nickname string) (string, bool) {
	trimmed := strings.TrimSpace(nickname)
	if trimmed == "" {
		return "", false
	}

	return trimmed, true
}

func stringValue(values []any, index int) string {
	if index >= len(values) || values[index] == nil {
		return ""
	}

	switch value := values[index].(type) {
	case string:
		return value
	case []byte:
		return string(value)
	default:
		return fmt.Sprint(value)
	}
}

func int64Value(values []any, index int) int64 {
	raw := stringValue(values, index)
	if raw == "" {
		return 0
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0
	}

	return value
}

func int64FromString(raw string) int64 {
	if strings.TrimSpace(raw) == "" {
		return 0
	}

	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0
	}

	return value
}

func clampInt(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
