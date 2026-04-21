package vote

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
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
var ErrSensitiveContent = errors.New("sensitive content")
var ErrEquipmentNotFound = errors.New("equipment not found")
var ErrEquipmentNotOwned = errors.New("equipment not owned")
var ErrEquipmentNotEnough = errors.New("equipment not enough")
var ErrEquipmentMaxStar = errors.New("equipment max star")
var ErrHeroNotFound = errors.New("hero not found")
var ErrHeroNotOwned = errors.New("hero not owned")
var ErrMessageEmpty = errors.New("message empty")
var ErrMessageTooLong = errors.New("message too long")
var ErrBossTemplateNotFound = errors.New("boss template not found")
var ErrBossPoolEmpty = errors.New("boss pool empty")

const (
	bossStatusActive   = "active"
	bossStatusDefeated = "defeated"
)

var loadoutSlots = []string{"weapon", "armor", "accessory"}

// Button 按钮数据结构，返回给前端和 SSE 客户端
type Button struct {
	Key               string   `json:"key"`
	RedisKey          string   `json:"redisKey"`
	Label             string   `json:"label"`
	Count             int64    `json:"count"`
	Sort              int      `json:"sort"`
	Enabled           bool     `json:"enabled"`
	Tags              []string `json:"tags,omitempty"`
	StarlightEligible bool     `json:"starlightEligible,omitempty"`
	ImagePath         string   `json:"imagePath,omitempty"`
	ImageAlt          string   `json:"imageAlt,omitempty"`
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
	TemplateID string `json:"templateId,omitempty"`
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

type HeroTraitType string

const (
	HeroTraitBonusClicks           HeroTraitType = "bonus_clicks"
	HeroTraitCriticalChancePercent HeroTraitType = "critical_chance_percent"
	HeroTraitCriticalCountBonus    HeroTraitType = "critical_count_bonus"
	HeroTraitFinalDamagePercent    HeroTraitType = "final_damage_percent"
)

// HeroDefinition 小小英雄模板。
type HeroDefinition struct {
	HeroID                     string        `json:"heroId"`
	Name                       string        `json:"name"`
	ImagePath                  string        `json:"imagePath,omitempty"`
	ImageAlt                   string        `json:"imageAlt,omitempty"`
	BonusClicks                int64         `json:"bonusClicks"`
	BonusCriticalChancePercent int           `json:"bonusCriticalChancePercent"`
	BonusCriticalCount         int64         `json:"bonusCriticalCount"`
	TraitType                  HeroTraitType `json:"traitType"`
	TraitValue                 int64         `json:"traitValue"`
}

// HeroInventoryItem 玩家持有的小小英雄。
type HeroInventoryItem struct {
	HeroID                     string        `json:"heroId"`
	Name                       string        `json:"name"`
	ImagePath                  string        `json:"imagePath,omitempty"`
	ImageAlt                   string        `json:"imageAlt,omitempty"`
	Quantity                   int64         `json:"quantity"`
	Active                     bool          `json:"active"`
	BonusClicks                int64         `json:"bonusClicks"`
	BonusCriticalChancePercent int           `json:"bonusCriticalChancePercent"`
	BonusCriticalCount         int64         `json:"bonusCriticalCount"`
	TraitType                  HeroTraitType `json:"traitType"`
	TraitValue                 int64         `json:"traitValue"`
}

// Announcement 更新公告
type Announcement struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	PublishedAt int64  `json:"publishedAt"`
	Active      bool   `json:"active"`
}

// AnnouncementUpsert 后台公告保存载荷
type AnnouncementUpsert struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Active  bool   `json:"active"`
}

// Message 公共留言
type Message struct {
	ID        string `json:"id"`
	Nickname  string `json:"nickname"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"createdAt"`
}

// MessagePage 留言分页结果
type MessagePage struct {
	Items      []Message `json:"items"`
	NextCursor string    `json:"nextCursor,omitempty"`
}

// InventoryItem 背包道具
type InventoryItem struct {
	ItemID                     string `json:"itemId"`
	Name                       string `json:"name"`
	Slot                       string `json:"slot"`
	Quantity                   int64  `json:"quantity"`
	StarLevel                  int    `json:"starLevel"`
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
	ItemID                     string  `json:"itemId"`
	ItemName                   string  `json:"itemName"`
	Slot                       string  `json:"slot"`
	Weight                     int64   `json:"weight"`
	DropRatePercent            float64 `json:"dropRatePercent"`
	BonusClicks                int64   `json:"bonusClicks"`
	BonusCriticalChancePercent int     `json:"bonusCriticalChancePercent"`
	BonusCriticalCount         int64   `json:"bonusCriticalCount"`
}

// BossHeroLootEntry Boss 英雄掉落池条目。
type BossHeroLootEntry struct {
	HeroID                     string        `json:"heroId"`
	HeroName                   string        `json:"heroName"`
	ImagePath                  string        `json:"imagePath,omitempty"`
	ImageAlt                   string        `json:"imageAlt,omitempty"`
	Weight                     int64         `json:"weight"`
	DropRatePercent            float64       `json:"dropRatePercent"`
	BonusClicks                int64         `json:"bonusClicks"`
	BonusCriticalChancePercent int           `json:"bonusCriticalChancePercent"`
	BonusCriticalCount         int64         `json:"bonusCriticalCount"`
	TraitType                  HeroTraitType `json:"traitType"`
	TraitValue                 int64         `json:"traitValue"`
}

// StarlightState 描述当前生效的星光卡片窗口。
type StarlightState struct {
	ActiveKeys []string `json:"activeKeys"`
	StartedAt  int64    `json:"startedAt"`
	EndsAt     int64    `json:"endsAt"`
}

// Snapshot 公共实时状态，广播给所有连接的客户端
type Snapshot struct {
	Buttons            []Button               `json:"buttons"`
	Leaderboard        []LeaderboardEntry     `json:"leaderboard"`
	Boss               *Boss                  `json:"boss,omitempty"`
	BossLeaderboard    []BossLeaderboardEntry `json:"bossLeaderboard"`
	BossLoot           []BossLootEntry        `json:"bossLoot"`
	BossHeroLoot       []BossHeroLootEntry    `json:"bossHeroLoot"`
	Starlight          StarlightState         `json:"starlight"`
	LatestAnnouncement *Announcement          `json:"latestAnnouncement,omitempty"`
}

// UserState 个人实时状态，只推送给对应昵称的连接
type UserState struct {
	UserStats   *UserStats          `json:"userStats,omitempty"`
	MyBossStats *BossUserStats      `json:"myBossStats,omitempty"`
	Inventory   []InventoryItem     `json:"inventory"`
	Heroes      []HeroInventoryItem `json:"heroes"`
	ActiveHero  *HeroInventoryItem  `json:"activeHero,omitempty"`
	Loadout     Loadout             `json:"loadout"`
	CombatStats CombatStats         `json:"combatStats"`
	LastReward  *Reward             `json:"lastReward,omitempty"`
}

// State 完整状态，包含个人统计与玩法状态
type State struct {
	Buttons            []Button               `json:"buttons"`
	Leaderboard        []LeaderboardEntry     `json:"leaderboard"`
	UserStats          *UserStats             `json:"userStats,omitempty"`
	Boss               *Boss                  `json:"boss,omitempty"`
	BossLeaderboard    []BossLeaderboardEntry `json:"bossLeaderboard"`
	BossLoot           []BossLootEntry        `json:"bossLoot"`
	BossHeroLoot       []BossHeroLootEntry    `json:"bossHeroLoot"`
	Starlight          StarlightState         `json:"starlight"`
	LatestAnnouncement *Announcement          `json:"latestAnnouncement,omitempty"`
	MyBossStats        *BossUserStats         `json:"myBossStats,omitempty"`
	Inventory          []InventoryItem        `json:"inventory"`
	Heroes             []HeroInventoryItem    `json:"heroes"`
	ActiveHero         *HeroInventoryItem     `json:"activeHero,omitempty"`
	Loadout            Loadout                `json:"loadout"`
	CombatStats        CombatStats            `json:"combatStats"`
	LastReward         *Reward                `json:"lastReward,omitempty"`
}

// ClickResult 点击结果，包含更新后的增量与状态摘要
type ClickResult struct {
	Button           Button                 `json:"button"`
	Delta            int64                  `json:"delta"`
	Critical         bool                   `json:"critical"`
	UserStats        UserStats              `json:"userStats"`
	Boss             *Boss                  `json:"boss,omitempty"`
	BossLeaderboard  []BossLeaderboardEntry `json:"bossLeaderboard,omitempty"`
	MyBossStats      *BossUserStats         `json:"myBossStats,omitempty"`
	LastReward       *Reward                `json:"lastReward,omitempty"`
	BroadcastUserAll bool                   `json:"-"`
}

// StateChangeType 实时状态变更类型
type StateChangeType string

const (
	StateChangeButtonClicked        StateChangeType = "button_clicked"
	StateChangeEquipmentChanged     StateChangeType = "equipment_changed"
	StateChangeBossChanged          StateChangeType = "boss_changed"
	StateChangeAnnouncementChanged  StateChangeType = "announcement_changed"
	StateChangeMessageCreated       StateChangeType = "message_created"
	StateChangeMessageDeleted       StateChangeType = "message_deleted"
	StateChangeButtonMetaChanged    StateChangeType = "button_meta_changed"
	StateChangeEquipmentMetaChanged StateChangeType = "equipment_meta_changed"
)

// StateChange 描述一次需要推送到实时层的状态变化
type StateChange struct {
	Type             StateChangeType `json:"type"`
	Nickname         string          `json:"nickname,omitempty"`
	BroadcastUserAll bool            `json:"broadcastUserAll,omitempty"`
	Timestamp        int64           `json:"timestamp"`
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
	client               redis.UniversalClient
	prefix               string
	namespace            string
	buttonIndexKey       string
	buttonStarlightKey   string
	equipmentIndexKey    string
	heroIndexKey         string
	playerIndexKey       string
	userPrefix           string
	leaderboardKey       string
	bossCurrentKey       string
	bossHistoryKey       string
	bossHistoryPrefix    string
	bossTemplateIndexKey string
	bossTemplatePrefix   string
	bossCycleKey         string
	bossInstanceSeqKey   string
	announcementSeqKey   string
	announcementKey      string
	announcementPrefix   string
	messageSeqKey        string
	messageKey           string
	messagePrefix        string
	equipmentDefPrefix   string
	heroDefPrefix        string
	inventoryPrefix      string
	heroInventoryPrefix  string
	loadoutPrefix        string
	activeHeroPrefix     string
	lastRewardPrefix     string
	upgradePrefix        string
	fallbacks            map[string]buttonFallback
	critical             StoreOptions
	luaRunner            luaScriptRunner
	bossClickScript      *cachedLuaScript
	roll                 func(int) int
	now                  func() time.Time
	validator            interface{ Validate(string) error }
}

// hashFields Redis Hash 中存储的字段列表
var hashFields = []string{
	"label",
	"count",
	"sort",
	"enabled",
	"tags",
	"starlight_eligible",
	"image_path",
	"image_alt",
}

const (
	starlightWindow = 5 * time.Minute
	starlightLimit  = 6
)

// NewStore 创建 Redis 投票存储实例
func NewStore(client redis.UniversalClient, prefix string, options StoreOptions, validator interface{ Validate(string) error }) *Store {
	namespace := deriveNamespace(prefix)
	luaCache := newLuaScriptCache()

	return &Store{
		client:               client,
		prefix:               prefix,
		namespace:            namespace,
		buttonIndexKey:       namespace + "buttons:index",
		buttonStarlightKey:   namespace + "buttons:starlight",
		equipmentIndexKey:    namespace + "equipment:index",
		heroIndexKey:         namespace + "heroes:index",
		playerIndexKey:       namespace + "players:index",
		userPrefix:           namespace + "user:",
		leaderboardKey:       namespace + "leaderboard",
		bossCurrentKey:       namespace + "boss:current",
		bossHistoryKey:       namespace + "boss:history",
		bossHistoryPrefix:    namespace + "boss:history:",
		bossTemplateIndexKey: namespace + "boss:pool:index",
		bossTemplatePrefix:   namespace + "boss:pool:",
		bossCycleKey:         namespace + "boss:cycle",
		bossInstanceSeqKey:   namespace + "boss:instance:seq",
		announcementSeqKey:   namespace + "announcement:seq",
		announcementKey:      namespace + "announcements",
		announcementPrefix:   namespace + "announcement:",
		messageSeqKey:        namespace + "message:seq",
		messageKey:           namespace + "messages",
		messagePrefix:        namespace + "message:",
		equipmentDefPrefix:   namespace + "equip:def:",
		heroDefPrefix:        namespace + "hero:def:",
		inventoryPrefix:      namespace + "user-inventory:",
		heroInventoryPrefix:  namespace + "user-hero-inventory:",
		loadoutPrefix:        namespace + "user-loadout:",
		activeHeroPrefix:     namespace + "user-active-hero:",
		lastRewardPrefix:     namespace + "user-last-reward:",
		upgradePrefix:        namespace + "user-equip-upgrade:",
		fallbacks: map[string]buttonFallback{
			"wechat-pity": {
				ImagePath: "/images/emojipedia-wechat-whimper.png",
				ImageAlt:  "微信可怜表情",
			},
		},
		critical: options,
		luaRunner: redisLuaRunner{
			client: client,
		},
		bossClickScript: newCachedLuaScript("boss-click", bossClickLuaSource, luaCache),
		roll: func(limit int) int {
			return rand.IntN(limit)
		},
		now:       time.Now,
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
	bossHeroLoot := []BossHeroLootEntry{}
	if boss != nil {
		bossLeaderboard, err = s.ListBossLeaderboard(ctx, boss.ID, 10)
		if err != nil {
			return Snapshot{}, err
		}
		bossLoot, err = s.loadBossLoot(ctx, boss.ID)
		if err != nil {
			return Snapshot{}, err
		}
		bossHeroLoot, err = s.loadBossHeroLoot(ctx, boss.ID)
		if err != nil {
			return Snapshot{}, err
		}
	}

	latestAnnouncement, err := s.GetLatestAnnouncement(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	return Snapshot{
		Buttons:            buttons,
		Leaderboard:        leaderboard,
		Boss:               boss,
		BossLeaderboard:    bossLeaderboard,
		BossLoot:           bossLoot,
		BossHeroLoot:       bossHeroLoot,
		Starlight:          starlightStateForButtons(buttons, s.now()),
		LatestAnnouncement: latestAnnouncement,
	}, nil
}

// GetState 获取完整状态（公共快照 + 个人统计）
func (s *Store) GetState(ctx context.Context, nickname string) (State, error) {
	snapshot, err := s.GetSnapshot(ctx)
	if err != nil {
		return State{}, err
	}

	userState, err := s.GetUserState(ctx, nickname)
	if err != nil {
		return State{}, err
	}

	return ComposeState(snapshot, userState), nil
}

// GetUserState 获取仅与指定用户相关的状态。
func (s *Store) GetUserState(ctx context.Context, nickname string) (UserState, error) {
	userState := UserState{
		Inventory:   []InventoryItem{},
		Heroes:      []HeroInventoryItem{},
		Loadout:     Loadout{},
		CombatStats: s.baseCombatStats(),
	}

	trimmedNickname, hasNickname := normalizeNickname(nickname)
	if !hasNickname {
		return userState, nil
	}

	normalizedNickname, err := s.validatedNickname(trimmedNickname)
	if err != nil {
		return UserState{}, err
	}

	userStats, err := s.GetUserStats(ctx, normalizedNickname)
	if err != nil {
		return UserState{}, err
	}
	userState.UserStats = &userStats

	quantities, err := s.inventoryQuantities(ctx, normalizedNickname)
	if err != nil {
		return UserState{}, err
	}

	loadout, equipped, err := s.loadoutForNickname(ctx, normalizedNickname, quantities)
	if err != nil {
		return UserState{}, err
	}
	userState.Loadout = loadout

	inventory, err := s.inventoryForNickname(ctx, normalizedNickname, quantities, equipped)
	if err != nil {
		return UserState{}, err
	}
	userState.Inventory = inventory

	heroQuantities, err := s.heroInventoryQuantities(ctx, normalizedNickname)
	if err != nil {
		return UserState{}, err
	}
	activeHero, err := s.activeHeroForNickname(ctx, normalizedNickname, heroQuantities)
	if err != nil {
		return UserState{}, err
	}
	heroes, err := s.heroInventoryForNickname(ctx, normalizedNickname, heroQuantities, activeHero)
	if err != nil {
		return UserState{}, err
	}
	userState.Heroes = heroes
	userState.ActiveHero = activeHero

	combatStats, err := s.combatStatsForNickname(ctx, normalizedNickname, loadout, activeHero)
	if err != nil {
		return UserState{}, err
	}
	userState.CombatStats = combatStats

	lastReward, err := s.lastRewardForNickname(ctx, normalizedNickname)
	if err != nil {
		return UserState{}, err
	}
	userState.LastReward = lastReward

	boss, err := s.currentBoss(ctx)
	if err != nil {
		return UserState{}, err
	}
	if boss != nil {
		myBossStats, err := s.bossStatsForNickname(ctx, boss.ID, normalizedNickname)
		if err != nil {
			return UserState{}, err
		}
		userState.MyBossStats = myBossStats
	}

	return userState, nil
}

// ListButtons 扫描 Redis，过滤禁用按钮，按排序权重返回
func (s *Store) ListButtons(ctx context.Context) ([]Button, error) {
	keys, err := s.listButtonKeys(ctx)
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
	if active, err := s.buttonStarlightActive(ctx, slug, s.now()); err != nil {
		return ClickResult{}, err
	} else if active {
		delta *= 2
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
		result, err = s.applyBossClick(ctx, current, boss, normalizedNickname, delta, critical)
		if err != nil {
			return ClickResult{}, err
		}
	}

	if result.Boss != nil {
		leaderboard, err := s.ListBossLeaderboard(ctx, result.Boss.ID, 10)
		if err != nil {
			return ClickResult{}, err
		}
		result.BossLeaderboard = leaderboard

		myBossStats, err := s.bossStatsForNickname(ctx, result.Boss.ID, normalizedNickname)
		if err != nil {
			return ClickResult{}, err
		}
		result.MyBossStats = myBossStats
	}

	lastReward, err := s.lastRewardForNickname(ctx, normalizedNickname)
	if err != nil {
		return ClickResult{}, err
	}
	result.LastReward = lastReward

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

	now := time.Now().Unix()
	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, s.loadoutKey(normalizedNickname), definition.Slot, itemID)
	pipe.ZAdd(ctx, s.playerIndexKey, redis.Z{
		Score:  float64(now),
		Member: normalizedNickname,
	})
	if _, err := pipe.Exec(ctx); err != nil {
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

	now := time.Now().Unix()
	pipe := s.client.TxPipeline()
	pipe.HDel(ctx, s.loadoutKey(normalizedNickname), definition.Slot)
	pipe.ZAdd(ctx, s.playerIndexKey, redis.Z{
		Score:  float64(now),
		Member: normalizedNickname,
	})
	if _, err := pipe.Exec(ctx); err != nil {
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
	keys, err := s.listButtonKeys(ctx)
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
		pipe.ZAdd(ctx, s.buttonIndexKey, redis.Z{
			Score:  float64(button.Sort),
			Member: button.Slug,
		})
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

// ComposeState 将公共快照与个人态组合成完整状态。
func ComposeState(snapshot Snapshot, userState UserState) State {
	return State{
		Buttons:            snapshot.Buttons,
		Leaderboard:        snapshot.Leaderboard,
		UserStats:          userState.UserStats,
		Boss:               snapshot.Boss,
		BossLeaderboard:    snapshot.BossLeaderboard,
		BossLoot:           snapshot.BossLoot,
		BossHeroLoot:       snapshot.BossHeroLoot,
		Starlight:          snapshot.Starlight,
		LatestAnnouncement: snapshot.LatestAnnouncement,
		MyBossStats:        userState.MyBossStats,
		Inventory:          userState.Inventory,
		Heroes:             userState.Heroes,
		ActiveHero:         userState.ActiveHero,
		Loadout:            userState.Loadout,
		CombatStats:        userState.CombatStats,
		LastReward:         userState.LastReward,
	}
}

func (s *Store) applyVoteOnlyClick(ctx context.Context, redisKey string, nickname string, delta int64, critical bool) (ClickResult, error) {
	now := time.Now().Unix()
	pipe := s.client.TxPipeline()
	pipe.HIncrBy(ctx, redisKey, "count", delta)
	userCountCmd := pipe.HIncrBy(ctx, s.userPrefix+nickname, "click_count", delta)
	pipe.HSet(ctx, s.userPrefix+nickname, map[string]any{
		"nickname":   nickname,
		"updated_at": strconv.FormatInt(now, 10),
	})
	pipe.ZIncrBy(ctx, s.leaderboardKey, float64(delta), nickname)
	pipe.ZAdd(ctx, s.playerIndexKey, redis.Z{
		Score:  float64(now),
		Member: nickname,
	})

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

func (s *Store) applyBossClick(ctx context.Context, current Button, boss *Boss, nickname string, delta int64, critical bool) (ClickResult, error) {
	if boss == nil || boss.Status != bossStatusActive {
		return s.applyVoteOnlyClick(ctx, current.RedisKey, nickname, delta, critical)
	}

	now := time.Now().Unix()
	scriptResult, err := s.bossClickScript.Run(ctx, s.luaRunner, []string{
		current.RedisKey,
		s.userPrefix + nickname,
		s.leaderboardKey,
		s.playerIndexKey,
		s.bossCurrentKey,
		s.bossDamageKey(boss.ID),
	}, delta, nickname, now, boss.ID, now)
	if err != nil {
		return ClickResult{}, err
	}

	values, ok := scriptResult.([]any)
	if !ok || len(values) < 3 {
		return ClickResult{}, fmt.Errorf("unexpected boss click script result: %T", scriptResult)
	}

	button := current
	button.Count = int64Value(values, 1)

	result := ClickResult{
		Button:   button,
		Delta:    delta,
		Critical: critical,
		UserStats: UserStats{
			Nickname:   nickname,
			ClickCount: int64Value(values, 2),
		},
	}

	if int64Value(values, 0) == 0 {
		return result, nil
	}

	result.Boss = &Boss{
		ID:         stringValue(values, 3),
		TemplateID: stringValue(values, 4),
		Name:       stringValue(values, 5),
		Status:     stringValue(values, 6),
		MaxHP:      int64Value(values, 7),
		CurrentHP:  int64Value(values, 8),
		StartedAt:  int64FromString(stringValue(values, 9)),
		DefeatedAt: int64FromString(stringValue(values, 10)),
	}

	if result.Boss.Status == bossStatusDefeated {
		result.BroadcastUserAll = true
		nextBoss, finalizeErr := s.finalizeBossKill(ctx, result.Boss)
		if finalizeErr != nil {
			return ClickResult{}, finalizeErr
		}
		if nextBoss != nil {
			result.Boss = nextBoss
		}
	}

	return result, nil
}

func (s *Store) finalizeBossKill(ctx context.Context, boss *Boss) (*Boss, error) {
	if boss == nil || strings.TrimSpace(boss.ID) == "" {
		return nil, nil
	}
	bossID := strings.TrimSpace(boss.ID)
	bossName := strings.TrimSpace(boss.Name)

	acquired, err := s.client.SetNX(ctx, s.bossRewardLockKey(bossID), "1", 0).Result()
	if err != nil {
		return nil, err
	}
	if !acquired {
		return s.currentBoss(ctx)
	}

	lootEntries, err := s.loadBossLoot(ctx, bossID)
	if err != nil {
		return nil, err
	}
	heroLootEntries, err := s.loadBossHeroLoot(ctx, bossID)
	if err != nil {
		return nil, err
	}

	if len(lootEntries) > 0 || len(heroLootEntries) > 0 {
		participants, err := s.client.ZRevRangeWithScores(ctx, s.bossDamageKey(bossID), 0, -1).Result()
		if err != nil {
			return nil, err
		}

		pipe := s.client.Pipeline()
		now := s.now().Unix()
		minDamage := (maxInt64(1, boss.MaxHP) + 99) / 100
		for _, participant := range participants {
			nickname, ok := participant.Member.(string)
			if !ok || nickname == "" || participant.Score < float64(minDamage) {
				continue
			}

			if reward := s.chooseLoot(lootEntries); reward != nil {
				pipe.HIncrBy(ctx, s.inventoryKey(nickname), reward.ItemID, 1)
				pipe.HSet(ctx, s.lastRewardKey(nickname), map[string]any{
					"boss_id":    bossID,
					"boss_name":  bossName,
					"item_id":    reward.ItemID,
					"item_name":  reward.ItemName,
					"granted_at": strconv.FormatInt(now, 10),
				})
			}
			if heroReward := s.chooseHeroLoot(heroLootEntries); heroReward != nil {
				pipe.HIncrBy(ctx, s.heroInventoryKey(nickname), heroReward.HeroID, 1)
				pipe.HSet(ctx, s.lastRewardKey(nickname), map[string]any{
					"boss_id":    bossID,
					"boss_name":  bossName,
					"item_id":    heroReward.HeroID,
					"item_name":  heroReward.HeroName,
					"granted_at": strconv.FormatInt(now, 10),
				})
			}
		}

		if _, err = pipe.Exec(ctx); err != nil {
			return nil, err
		}
	}

	if err := s.SaveBossToHistory(ctx, boss); err != nil {
		return nil, err
	}

	enabled, err := s.bossCycleEnabled(ctx)
	if err != nil {
		return nil, err
	}
	if enabled {
		nextBoss, err := s.activateRandomBossFromPool(ctx)
		if err != nil && !errors.Is(err, ErrBossPoolEmpty) {
			return nil, err
		}
		if nextBoss != nil {
			return nextBoss, nil
		}
	}

	return s.currentBoss(ctx)
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
	heroQuantities, err := s.heroInventoryQuantities(ctx, nickname)
	if err != nil {
		return 0, false, err
	}
	activeHero, err := s.activeHeroForNickname(ctx, nickname, heroQuantities)
	if err != nil {
		return 0, false, err
	}
	combatStats, err := s.combatStatsForNickname(ctx, nickname, loadout, activeHero)
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

func (s *Store) combatStatsForNickname(_ context.Context, _ string, loadout Loadout, activeHero *HeroInventoryItem) (CombatStats, error) {
	stats := s.baseCombatStats()

	bonusClicks, bonusChance, bonusCount := loadoutBonuses(loadout)
	stats.BonusClicks = bonusClicks
	stats.CriticalChancePercent = clampInt(stats.CriticalChancePercent+bonusChance, 0, 100)
	stats.CriticalCount += bonusCount

	heroBonusClicks, heroBonusChance, heroBonusCount, heroFinalDamagePercent := heroBonuses(activeHero)
	stats.BonusClicks += heroBonusClicks
	stats.CriticalChancePercent = clampInt(stats.CriticalChancePercent+heroBonusChance, 0, 100)
	stats.CriticalCount += heroBonusCount

	return applyFinalDamagePercent(deriveCombatStats(stats), heroFinalDamagePercent), nil
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
		TemplateID: strings.TrimSpace(values["template_id"]),
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
		upgrade, upgradeErr := s.getEquipmentUpgrade(ctx, nickname, itemID)
		if upgradeErr != nil {
			return Loadout{}, nil, upgradeErr
		}

		item := s.buildInventoryItem(definition, upgrade, quantities[itemID], true)

		equipped[itemID] = slot
		switch slot {
		case "weapon":
			loadout.Weapon = &item
		case "armor":
			loadout.Armor = &item
		case "accessory":
			loadout.Accessory = &item
		}
	}

	return loadout, equipped, nil
}

func (s *Store) inventoryForNickname(ctx context.Context, nickname string, quantities map[string]int64, equipped map[string]string) ([]InventoryItem, error) {
	if len(quantities) == 0 {
		return []InventoryItem{}, nil
	}

	items := make([]InventoryItem, 0, len(quantities))
	for itemID, quantity := range quantities {
		upgrade, upgradeErr := s.getEquipmentUpgrade(ctx, nickname, itemID)
		if upgradeErr != nil {
			return nil, upgradeErr
		}
		definition, err := s.getEquipmentDefinition(ctx, itemID)
		if err != nil {
			items = append(items, unknownInventoryItem(itemID, upgrade, quantity, equipped[itemID] != ""))
			continue
		}

		items = append(items, s.buildInventoryItem(definition, upgrade, quantity, equipped[itemID] != ""))
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
	totalWeight := int64(0)
	for _, entry := range entries {
		if entry.Score > 0 {
			totalWeight += int64(entry.Score)
		}
	}
	for _, entry := range entries {
		itemID, ok := entry.Member.(string)
		if !ok || strings.TrimSpace(itemID) == "" {
			continue
		}

		dropRatePercent := percentageFromWeight(int64(entry.Score), totalWeight)

		definition, defErr := s.getEquipmentDefinition(ctx, itemID)
		if defErr != nil {
			loot = append(loot, BossLootEntry{
				ItemID:          itemID,
				Weight:          int64(entry.Score),
				DropRatePercent: dropRatePercent,
			})
			continue
		}

		loot = append(loot, BossLootEntry{
			ItemID:                     itemID,
			ItemName:                   definition.Name,
			Slot:                       definition.Slot,
			Weight:                     int64(entry.Score),
			DropRatePercent:            dropRatePercent,
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

	imagePath := stringValue(values, 6)
	if imagePath == "" {
		imagePath = fallback.ImagePath
	}

	imageAlt := stringValue(values, 7)
	if imageAlt == "" {
		imageAlt = fallback.ImageAlt
	}

	return Button{
		Key:               slug,
		RedisKey:          redisKey,
		Label:             label,
		Count:             int64Value(values, 1),
		Sort:              int(int64Value(values, 2)),
		Enabled:           stringValue(values, 3) != "0",
		Tags:              decodeStringList(stringValue(values, 4)),
		StarlightEligible: stringValue(values, 5) == "1",
		ImagePath:         imagePath,
		ImageAlt:          imageAlt,
	}
}

func decodeStringList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}

	normalized := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		normalized = append(normalized, item)
	}

	return normalized
}

func encodeStringList(items []string) string {
	normalized := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		normalized = append(normalized, item)
	}

	if len(normalized) == 0 {
		return "[]"
	}

	encoded, err := json.Marshal(normalized)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

func starlightStateForButtons(buttons []Button, now time.Time) StarlightState {
	startedAt := now.Unix() - (now.Unix() % int64(starlightWindow/time.Second))
	eligible := make([]string, 0, len(buttons))
	for _, button := range buttons {
		if button.StarlightEligible {
			eligible = append(eligible, button.Key)
		}
	}

	return StarlightState{
		ActiveKeys: activeStarlightKeys(eligible, startedAt),
		StartedAt:  startedAt,
		EndsAt:     startedAt + int64(starlightWindow/time.Second),
	}
}

func activeStarlightKeys(keys []string, startedAt int64) []string {
	normalized := make([]string, 0, len(keys))
	seen := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, key)
	}

	if len(normalized) <= starlightLimit {
		slices.Sort(normalized)
		return normalized
	}

	type scoredKey struct {
		key   string
		score uint64
	}

	scored := make([]scoredKey, 0, len(normalized))
	for _, key := range normalized {
		hasher := fnv.New64a()
		_, _ = hasher.Write([]byte(strconv.FormatInt(startedAt, 10)))
		_, _ = hasher.Write([]byte(":"))
		_, _ = hasher.Write([]byte(key))
		scored = append(scored, scoredKey{
			key:   key,
			score: hasher.Sum64(),
		})
	}

	slices.SortFunc(scored, func(left, right scoredKey) int {
		if left.score == right.score {
			return strings.Compare(left.key, right.key)
		}
		if left.score < right.score {
			return -1
		}
		return 1
	})

	active := make([]string, 0, starlightLimit)
	for _, item := range scored[:starlightLimit] {
		active = append(active, item.key)
	}
	slices.Sort(active)
	return active
}

func (s *Store) buttonStarlightActive(ctx context.Context, slug string, now time.Time) (bool, error) {
	keys, err := s.client.SMembers(ctx, s.buttonStarlightKey).Result()
	if err != nil {
		return false, err
	}

	startedAt := now.Unix() - (now.Unix() % int64(starlightWindow/time.Second))
	for _, key := range activeStarlightKeys(keys, startedAt) {
		if key == slug {
			return true, nil
		}
	}
	return false, nil
}

func percentageFromWeight(weight int64, totalWeight int64) float64 {
	if weight <= 0 || totalWeight <= 0 {
		return 0
	}

	return float64(weight) * 100 / float64(totalWeight)
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

func (s *Store) listButtonKeys(ctx context.Context) ([]string, error) {
	slugs, err := s.client.ZRange(ctx, s.buttonIndexKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	if len(slugs) > 0 {
		keys := make([]string, 0, len(slugs))
		for _, slug := range slugs {
			slug = strings.TrimSpace(slug)
			if slug == "" {
				continue
			}
			keys = append(keys, s.prefix+slug)
		}
		return keys, nil
	}

	keys, err := s.scanKeys(ctx)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return []string{}, nil
	}
	if err := s.rebuildButtonIndex(ctx, keys); err != nil {
		return nil, err
	}
	return keys, nil
}

func (s *Store) rebuildButtonIndex(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	pipe := s.client.Pipeline()
	cmds := make([]*redis.SliceCmd, len(keys))
	for index, key := range keys {
		cmds[index] = pipe.HMGet(ctx, key, "sort")
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	entries := make([]redis.Z, 0, len(keys))
	for index, key := range keys {
		slug := strings.TrimSpace(strings.TrimPrefix(key, s.prefix))
		if slug == "" {
			continue
		}
		entries = append(entries, redis.Z{
			Score:  float64(int64Value(cmds[index].Val(), 0)),
			Member: slug,
		})
	}
	if len(entries) == 0 {
		return nil
	}

	return s.client.ZAdd(ctx, s.buttonIndexKey, entries...).Err()
}

// SyncButtonIndex 将直接写入 Redis 的按钮补进显式索引，供低频兜底扫描使用。
func (s *Store) SyncButtonIndex(ctx context.Context) (bool, error) {
	keys, err := s.scanKeys(ctx)
	if err != nil {
		return false, err
	}
	if len(keys) == 0 {
		return false, nil
	}

	indexedSlugs, err := s.client.ZRange(ctx, s.buttonIndexKey, 0, -1).Result()
	if err != nil {
		return false, err
	}
	indexed := make(map[string]struct{}, len(indexedSlugs))
	for _, slug := range indexedSlugs {
		indexed[strings.TrimSpace(slug)] = struct{}{}
	}

	missingKeys := make([]string, 0)
	for _, key := range keys {
		slug := strings.TrimSpace(strings.TrimPrefix(key, s.prefix))
		if slug == "" {
			continue
		}
		if _, ok := indexed[slug]; ok {
			continue
		}
		missingKeys = append(missingKeys, key)
	}
	if len(missingKeys) == 0 {
		return false, nil
	}

	if err := s.rebuildButtonIndex(ctx, missingKeys); err != nil {
		return false, err
	}

	return true, nil
}

func (s *Store) listEquipmentIDs(ctx context.Context) ([]string, error) {
	itemIDs, err := s.client.SMembers(ctx, s.equipmentIndexKey).Result()
	if err != nil {
		return nil, err
	}
	if len(itemIDs) > 0 {
		return itemIDs, nil
	}

	keys, err := s.scanByPrefix(ctx, s.equipmentDefPrefix)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return []string{}, nil
	}

	itemIDs = make([]string, 0, len(keys))
	for _, key := range keys {
		itemID := strings.TrimSpace(strings.TrimPrefix(key, s.equipmentDefPrefix))
		if itemID == "" {
			continue
		}
		itemIDs = append(itemIDs, itemID)
	}
	if len(itemIDs) == 0 {
		return []string{}, nil
	}
	if err := s.client.SAdd(ctx, s.equipmentIndexKey, toAnySlice(itemIDs)...).Err(); err != nil {
		return nil, err
	}

	return itemIDs, nil
}

func (s *Store) listPlayerNicknames(ctx context.Context) ([]string, error) {
	nicknames, err := s.client.ZRevRange(ctx, s.playerIndexKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	if len(nicknames) > 0 {
		return nicknames, nil
	}

	keys, err := s.scanByPrefix(ctx, s.userPrefix)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return []string{}, nil
	}

	type playerEntry struct {
		nickname  string
		updatedAt int64
	}

	entries := make([]playerEntry, 0, len(keys))
	for _, key := range keys {
		nickname := strings.TrimSpace(strings.TrimPrefix(key, s.userPrefix))
		if nickname == "" {
			continue
		}
		values, err := s.client.HMGet(ctx, key, "updated_at").Result()
		if err != nil {
			return nil, err
		}
		entries = append(entries, playerEntry{
			nickname:  nickname,
			updatedAt: int64Value(values, 0),
		})
	}
	if len(entries) == 0 {
		return []string{}, nil
	}

	zEntries := make([]redis.Z, 0, len(entries))
	nicknames = make([]string, 0, len(entries))
	for _, entry := range entries {
		nicknames = append(nicknames, entry.nickname)
		zEntries = append(zEntries, redis.Z{
			Score:  float64(entry.updatedAt),
			Member: entry.nickname,
		})
	}
	if len(zEntries) > 0 {
		if err := s.client.ZAdd(ctx, s.playerIndexKey, zEntries...).Err(); err != nil {
			return nil, err
		}
	}

	return nicknames, nil
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

func (s *Store) heroInventoryKey(nickname string) string {
	return s.heroInventoryPrefix + nickname
}

func (s *Store) loadoutKey(nickname string) string {
	return s.loadoutPrefix + nickname
}

func (s *Store) activeHeroKey(nickname string) string {
	return s.activeHeroPrefix + nickname
}

func (s *Store) lastRewardKey(nickname string) string {
	return s.lastRewardPrefix + nickname
}

func (s *Store) announcementItemKey(id string) string {
	return s.announcementPrefix + strings.TrimSpace(id)
}

func (s *Store) messageItemKey(id string) string {
	return s.messagePrefix + strings.TrimSpace(id)
}

func (s *Store) upgradeKey(nickname string, itemID string) string {
	return s.upgradePrefix + nickname + ":" + itemID
}

func (s *Store) equipmentKey(itemID string) string {
	return s.equipmentDefPrefix + itemID
}

func (s *Store) heroKey(heroID string) string {
	return s.heroDefPrefix + heroID
}

func (s *Store) bossDamageKey(bossID string) string {
	return s.namespace + "boss:" + bossID + ":damage"
}

func (s *Store) bossLootKey(bossID string) string {
	return s.namespace + "boss:" + bossID + ":loot"
}

func (s *Store) bossTemplateKey(templateID string) string {
	return s.bossTemplatePrefix + strings.TrimSpace(templateID)
}

func (s *Store) bossTemplateLootKey(templateID string) string {
	return s.bossTemplateKey(templateID) + ":loot"
}

func (s *Store) bossTemplateHeroLootKey(templateID string) string {
	return s.bossTemplateKey(templateID) + ":hero-loot"
}

func (s *Store) bossHeroLootKey(bossID string) string {
	return s.namespace + "boss:" + bossID + ":hero-loot"
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

// RealtimeEventChannel 返回当前命名空间对应的 Redis 实时事件通道名。
func RealtimeEventChannel(prefix string) string {
	return deriveNamespace(prefix) + "events"
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

func toAnySlice(values []string) []any {
	items := make([]any, 0, len(values))
	for _, value := range values {
		items = append(items, value)
	}
	return items
}
