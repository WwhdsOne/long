package vote

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
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
var ErrEquipmentMaxEnhance = errors.New("equipment max enhance")
var ErrGemsNotEnough = errors.New("gems not enough")
var ErrInvalidQuantity = errors.New("invalid quantity")
var ErrMessageEmpty = errors.New("message empty")
var ErrMessageTooLong = errors.New("message too long")
var ErrBossTemplateNotFound = errors.New("boss template not found")
var ErrBossPoolEmpty = errors.New("boss pool empty")
var ErrBossPartNotFound = errors.New("boss part not found")
var ErrBossPartAlreadyDead = errors.New("boss part already dead")
var ErrTalentTreeNotSet = errors.New("talent tree not set")
var ErrTalentAlreadyLearned = errors.New("talent already learned")
var ErrTalentPrerequisite = errors.New("talent prerequisite not met")
var ErrTalentNotFound = errors.New("talent not found")
var ErrTalentMaxLevel = errors.New("talent max level reached")
var ErrInvalidPartType = errors.New("invalid part type")
var ErrInvalidTalentTree = errors.New("invalid talent tree")

const (
	bossStatusActive   = "active"
	bossStatusDefeated = "defeated"
)

// Button 按钮数据结构，返回给前端和 SSE 客户端
type Button struct {
	Key       string   `json:"key"`
	RedisKey  string   `json:"redisKey"`
	Label     string   `json:"label"`
	Count     int64    `json:"count"`
	Sort      int      `json:"sort"`
	Enabled   bool     `json:"enabled"`
	Tags      []string `json:"tags,omitempty"`
	ImagePath string   `json:"imagePath,omitempty"`
	ImageAlt  string   `json:"imageAlt,omitempty"`
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

// PartType 部位类型
type PartType string

const (
	PartTypeSoft  PartType = "soft"  // 软组织
	PartTypeHeavy PartType = "heavy" // 重甲
	PartTypeWeak  PartType = "weak"  // 弱点
)

// PartDamageCoefficient 返回部位类型的伤害系数
func (p PartType) DamageCoefficient() float64 {
	switch p {
	case PartTypeSoft:
		return 1.0
	case PartTypeHeavy:
		return 0.4
	case PartTypeWeak:
		return 2.5
	default:
		return 1.0
	}
}

// BossPart Boss 的战斗部位
type BossPart struct {
	X         int      `json:"x"`
	Y         int      `json:"y"`
	Type      PartType `json:"type"`
	MaxHP     int64    `json:"maxHp"`
	CurrentHP int64    `json:"currentHp"`
	Armor     int64    `json:"armor"`
	Alive     bool     `json:"alive"`
}

// Boss 世界 Boss 状态
type Boss struct {
	ID         string     `json:"id"`
	TemplateID string     `json:"templateId,omitempty"`
	Name       string     `json:"name"`
	Status     string     `json:"status"`
	MaxHP      int64      `json:"maxHp"`
	CurrentHP  int64      `json:"currentHp"`
	Parts      []BossPart `json:"parts,omitempty"`
	StartedAt  int64      `json:"startedAt,omitempty"`
	DefeatedAt int64      `json:"defeatedAt,omitempty"`
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

// EquipmentDefinition 武器模板
type EquipmentDefinition struct {
	ItemID              string  `json:"itemId"`
	Name                string  `json:"name"`
	Slot                string  `json:"slot"`
	Rarity              string  `json:"rarity"`
	ImagePath           string  `json:"imagePath,omitempty"`
	ImageAlt            string  `json:"imageAlt,omitempty"`
	AttackPower         int64   `json:"attackPower,omitempty"`
	ArmorPenPercent     float64 `json:"armorPenPercent,omitempty"`
	CritDamageMultiplier float64 `json:"critDamageMultiplier,omitempty"`
	BossDamagePercent   float64 `json:"bossDamagePercent,omitempty"`
	PartTypeDamageSoft  float64 `json:"partTypeDamageSoft,omitempty"`  // 软组织增伤
	PartTypeDamageHeavy float64 `json:"partTypeDamageHeavy,omitempty"` // 重甲增伤
	PartTypeDamageWeak  float64 `json:"partTypeDamageWeak,omitempty"`  // 弱点增伤
	TalentAffinity      string  `json:"talentAffinity,omitempty"`      // 天赋系绑定
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
// InventoryItem 背包道具（武器/装备）
type InventoryItem struct {
	ItemID              string  `json:"itemId"`
	Name                string  `json:"name"`
	Slot                string  `json:"slot"`
	Rarity              string  `json:"rarity"`
	ImagePath           string  `json:"imagePath,omitempty"`
	ImageAlt            string  `json:"imageAlt,omitempty"`
	Quantity            int64   `json:"quantity"`
	Equipped            bool    `json:"equipped"`
	AttackPower         int64   `json:"attackPower,omitempty"`
	ArmorPenPercent     float64 `json:"armorPenPercent,omitempty"`
	CritDamageMultiplier float64 `json:"critDamageMultiplier,omitempty"`
	BossDamagePercent   float64 `json:"bossDamagePercent,omitempty"`
	PartTypeDamageSoft  float64 `json:"partTypeDamageSoft,omitempty"`
	PartTypeDamageHeavy float64 `json:"partTypeDamageHeavy,omitempty"`
	PartTypeDamageWeak  float64 `json:"partTypeDamageWeak,omitempty"`
}

// Loadout 已穿戴装备
type Loadout struct {
	Weapon    *InventoryItem `json:"weapon,omitempty"`
	Armor     *InventoryItem `json:"armor,omitempty"`
	Accessory *InventoryItem `json:"accessory,omitempty"`
}

// CombatStats 当前生效的点击战斗属性
type CombatStats struct {
	EffectiveIncrement     int64   `json:"effectiveIncrement"`
	NormalDamage           int64   `json:"normalDamage"`
	CriticalChancePercent  float64 `json:"criticalChancePercent"`
	CriticalCount          int64   `json:"criticalCount"`
	CriticalDamage         int64   `json:"criticalDamage"`
	AttackPower           int64   `json:"attackPower"`
	ArmorPenPercent       float64 `json:"armorPenPercent"`
	CritDamageMultiplier  float64 `json:"critDamageMultiplier"`
	BossDamagePercent     float64 `json:"bossDamagePercent"`
	AllDamageAmplify      float64 `json:"allDamageAmplify"`
	PartTypeDamageSoft   float64 `json:"partTypeDamageSoft,omitempty"`
	PartTypeDamageHeavy  float64 `json:"partTypeDamageHeavy,omitempty"`
	PartTypeDamageWeak   float64 `json:"partTypeDamageWeak,omitempty"`
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
	ItemID              string  `json:"itemId"`
	ItemName            string  `json:"itemName"`
	Slot                string  `json:"slot"`
	Rarity              string  `json:"rarity"`
	Weight              int64   `json:"weight"`
	DropRatePercent     float64 `json:"dropRatePercent"`
	AttackPower         int64   `json:"attackPower,omitempty"`
	ArmorPenPercent     float64 `json:"armorPenPercent,omitempty"`
	CritDamageMultiplier float64 `json:"critDamageMultiplier,omitempty"`
	BossDamagePercent   float64 `json:"bossDamagePercent,omitempty"`
	PartTypeDamageSoft  float64 `json:"partTypeDamageSoft,omitempty"`
	PartTypeDamageHeavy float64 `json:"partTypeDamageHeavy,omitempty"`
	PartTypeDamageWeak  float64 `json:"partTypeDamageWeak,omitempty"`
	TalentAffinity      string  `json:"talentAffinity,omitempty"`
}

// BossResources 描述当前 Boss 的低频公共资源。
type BossResources struct {
	BossID       string              `json:"bossId,omitempty"`
	TemplateID   string              `json:"templateId,omitempty"`
	Status       string              `json:"status,omitempty"`
	BossLoot     []BossLootEntry     `json:"bossLoot"`
}

// Snapshot 公共实时状态，广播给所有连接的客户端
type Snapshot struct {
	Buttons             []Button               `json:"buttons"`
	TotalVotes          int64                  `json:"totalVotes"`
	Leaderboard         []LeaderboardEntry     `json:"leaderboard"`
	Boss                *Boss                  `json:"boss,omitempty"`
	BossLeaderboard     []BossLeaderboardEntry `json:"bossLeaderboard"`
	AnnouncementVersion string                 `json:"announcementVersion,omitempty"`
}

// UserState 个人实时状态，只推送给对应昵称的连接
type UserState struct {
	UserStats       *UserStats          `json:"userStats,omitempty"`
	MyBossStats     *BossUserStats      `json:"myBossStats,omitempty"`
	Inventory       []InventoryItem     `json:"inventory"`
	Loadout         Loadout             `json:"loadout"`
	CombatStats     CombatStats         `json:"combatStats"`
	Gems            int64               `json:"gems"`
	RecentRewards   []Reward            `json:"recentRewards,omitempty"`
	LastReward      *Reward             `json:"lastReward,omitempty"`
}

// State 完整状态，包含个人统计与玩法状态
type State struct {
	Buttons             []Button               `json:"buttons"`
	TotalVotes          int64                  `json:"totalVotes"`
	Leaderboard         []LeaderboardEntry     `json:"leaderboard"`
	UserStats           *UserStats             `json:"userStats,omitempty"`
	Boss                *Boss                  `json:"boss,omitempty"`
	BossLeaderboard     []BossLeaderboardEntry `json:"bossLeaderboard"`
	BossLoot            []BossLootEntry        `json:"bossLoot,omitempty"`
	AnnouncementVersion string                 `json:"announcementVersion,omitempty"`
	LatestAnnouncement  *Announcement          `json:"latestAnnouncement,omitempty"`
	MyBossStats         *BossUserStats         `json:"myBossStats,omitempty"`
	Inventory           []InventoryItem        `json:"inventory"`
	Loadout             Loadout                `json:"loadout"`
	CombatStats         CombatStats            `json:"combatStats"`
	Gems                int64                  `json:"gems"`
	RecentRewards       []Reward               `json:"recentRewards,omitempty"`
	LastReward          *Reward                `json:"lastReward,omitempty"`
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
	RecentRewards    []Reward               `json:"recentRewards,omitempty"`
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
	equipmentIndexKey    string
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
	inventoryPrefix      string
	loadoutPrefix        string
	lastRewardPrefix     string
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
	"image_path",
	"image_alt",
}

// NewStore 创建 Redis 投票存储实例
func NewStore(client redis.UniversalClient, prefix string, options StoreOptions, validator interface{ Validate(string) error }) *Store {
	namespace := deriveNamespace(prefix)
	luaCache := newLuaScriptCache()

	return &Store{
		client:               client,
		prefix:               prefix,
		namespace:            namespace,
		buttonIndexKey:       namespace + "buttons:index",
		equipmentIndexKey:    namespace + "equipment:index",
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
		inventoryPrefix:      namespace + "user-inventory:",
		loadoutPrefix:        namespace + "user-loadout:",
		lastRewardPrefix:     namespace + "user-last-reward:",
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
	totalVotes := int64(0)
	for _, button := range buttons {
		totalVotes += button.Count
	}

	leaderboard, err := s.ListLeaderboard(ctx, 10)
	if err != nil {
		return Snapshot{}, err
	}

	boss, err := s.currentBoss(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	var bossLeaderboard []BossLeaderboardEntry
	if boss != nil {
		bossLeaderboard, err = s.ListBossLeaderboard(ctx, boss.ID, 10)
		if err != nil {
			return Snapshot{}, err
		}
	}

	announcementVersion, err := s.GetLatestAnnouncementVersion(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	return Snapshot{
		Buttons:             buttons,
		TotalVotes:          totalVotes,
		Leaderboard:         leaderboard,
		Boss:                boss,
		BossLeaderboard:     bossLeaderboard,
		AnnouncementVersion: announcementVersion,
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
		Inventory:     []InventoryItem{},
		Loadout:       Loadout{},
		CombatStats:   s.baseCombatStats(),
		RecentRewards: []Reward{},
	}

	trimmedNickname, hasNickname := normalizeNickname(nickname)
	if !hasNickname {
		return userState, nil
	}

	normalizedNickname, err := s.validatedNickname(trimmedNickname)
	if err != nil {
		return UserState{}, err
	}

	gems, err := s.gemsForNickname(ctx, normalizedNickname)
	if err != nil {
		return UserState{}, err
	}
	userState.Gems = gems

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


	combatStats, err := s.combatStatsForNickname(ctx, normalizedNickname, loadout)
	if err != nil {
		return UserState{}, err
	}
	userState.CombatStats = combatStats

	recentRewards, err := s.recentRewardsForNickname(ctx, normalizedNickname)
	if err != nil {
		return UserState{}, err
	}
	userState.RecentRewards = recentRewards
	if len(recentRewards) > 0 {
		userState.LastReward = new(recentRewards[len(recentRewards)-1])
	}

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

	if result.BroadcastUserAll {
		recentRewards, err := s.recentRewardsForNickname(ctx, normalizedNickname)
		if err != nil {
			return ClickResult{}, err
		}
		result.RecentRewards = recentRewards
		if len(recentRewards) > 0 {
			result.LastReward = new(recentRewards[len(recentRewards)-1])
		}
	}

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

	redisKey := s.userPrefix + normalizedNickname
	values, err := s.client.HMGet(ctx, redisKey, "nickname", "click_count").Result()
	if err != nil {
		if !isRedisWrongTypeError(err) {
			return UserStats{}, err
		}

		legacyValue, legacyErr := s.client.Get(ctx, redisKey).Result()
		if legacyErr != nil {
			return UserStats{}, legacyErr
		}

		return UserStats{
			Nickname:   normalizedNickname,
			ClickCount: int64FromString(legacyValue),
		}, nil
	}

	return UserStats{
		Nickname:   normalizedNickname,
		ClickCount: int64Value(values, 1),
	}, nil
}

// ComposeState 将公共快照与个人态组合成完整状态。
func ComposeState(snapshot Snapshot, userState UserState) State {
	return State{
		Buttons:             snapshot.Buttons,
		TotalVotes:          snapshot.TotalVotes,
		Leaderboard:         snapshot.Leaderboard,
		UserStats:           userState.UserStats,
		Boss:                snapshot.Boss,
		BossLeaderboard:     snapshot.BossLeaderboard,
		AnnouncementVersion: snapshot.AnnouncementVersion,
		MyBossStats:         userState.MyBossStats,
		Inventory:           userState.Inventory,
		Loadout:             userState.Loadout,
		CombatStats:         userState.CombatStats,
		Gems:                userState.Gems,
		RecentRewards:       userState.RecentRewards,
		LastReward:          userState.LastReward,
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
	if len(boss.Parts) > 0 {
		return s.applyBossPartClick(ctx, current, boss, nickname, delta, critical)
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

func (s *Store) applyBossPartClick(ctx context.Context, current Button, boss *Boss, nickname string, delta int64, critical bool) (ClickResult, error) {
	result, err := s.applyVoteOnlyClick(ctx, current.RedisKey, nickname, delta, critical)
	if err != nil {
		return ClickResult{}, err
	}

	quantities, err := s.inventoryQuantities(ctx, nickname)
	if err != nil {
		return result, nil
	}
	loadout, _, err := s.loadoutForNickname(ctx, nickname, quantities)
	if err != nil {
		return result, nil
	}

	combatStats, err := s.combatStatsForNickname(ctx, nickname, loadout)
	if err != nil {
		return result, nil
	}

	targetIdx := s.selectTargetPart(boss.Parts, nickname)
	if targetIdx < 0 {
		return result, nil
	}
	part := &boss.Parts[targetIdx]

	aliveCount := 0
	for _, p := range boss.Parts {
		if p.Alive {
			aliveCount++
		}
	}

	damageStats := CalcBossPartDamage(combatStats, part.Type, part.Armor, true, aliveCount)
	partDamage := damageStats.NormalDamage
	if critical {
		partDamage = damageStats.CriticalDamage
	}

	part.CurrentHP -= partDamage
	if part.CurrentHP < 0 {
		part.CurrentHP = 0
	}
	if part.CurrentHP <= 0 {
		part.Alive = false
	}

	partsRaw, marshalErr := sonic.Marshal(boss.Parts)
	if marshalErr != nil {
		return result, nil
	}

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, s.bossCurrentKey, "parts", string(partsRaw))
	pipe.ZIncrBy(ctx, s.bossDamageKey(boss.ID), float64(partDamage), nickname)
	pipe.Exec(ctx)

	result.Delta = partDamage
	result.Critical = critical

	boss.CurrentHP -= partDamage
	if boss.CurrentHP < 0 {
		boss.CurrentHP = 0
	}
	result.Boss = boss

	allDead := true
	for _, p := range boss.Parts {
		if p.Alive {
			allDead = false
			break
		}
	}

	if allDead {
		boss.Status = bossStatusDefeated
		result.BroadcastUserAll = true
		nextBoss, finalizeErr := s.finalizeBossKill(ctx, boss)
		if finalizeErr != nil {
			return result, nil
		}
		if nextBoss != nil {
			result.Boss = nextBoss
		}
	}

	return result, nil
}

func (s *Store) selectTargetPart(parts []BossPart, nickname string) int {
	if len(parts) == 0 {
		return -1
	}
	alive := make([]int, 0, len(parts))
	for i, p := range parts {
		if p.Alive {
			alive = append(alive, i)
		}
	}
	if len(alive) == 0 {
		return -1
	}
	if len(alive) == 1 {
		return alive[0]
	}
	return alive[s.roll(len(alive))]
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
	if len(lootEntries) > 0 {
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

			rewards := make([]Reward, 0, 2)
			if reward := s.chooseLoot(lootEntries); reward != nil {
				pipe.HIncrBy(ctx, s.inventoryKey(nickname), reward.ItemID, 1)
				rewards = append(rewards, Reward{
					BossID:    bossID,
					BossName:  bossName,
					ItemID:    reward.ItemID,
					ItemName:  reward.ItemName,
					GrantedAt: now,
				})
			}
			if len(rewards) > 0 {
				pipe.HSet(ctx, s.lastRewardKey(nickname), rewardRecordValues(rewards))
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
			return new(entry)
		}
	}

	return new(entries[len(entries)-1])
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

	rollLimit, threshold := criticalRollPlan(combatStats.CriticalChancePercent)
	if s.roll(rollLimit) < threshold {
		return combatStats.CriticalDamage, true, nil
	}

	return delta, false, nil
}

func (s *Store) combatStatsForNickname(ctx context.Context, nickname string, loadout Loadout) (CombatStats, error) {
	stats := s.baseCombatStats()

	attackPower, armorPen, critDmgMult, bossDmg := loadoutBonuses(loadout)
	stats.AttackPower += attackPower
	stats.ArmorPenPercent = clampFloat(stats.ArmorPenPercent+armorPen, 0, 0.80)
	stats.CritDamageMultiplier += critDmgMult
	stats.BossDamagePercent += bossDmg

	result := deriveCombatStats(stats)
	return result, nil
}

func (s *Store) baseCombatStats() CombatStats {
	return deriveCombatStats(CombatStats{
		CriticalChancePercent: clampFloat(float64(s.critical.CriticalChancePercent), 0, 100),
		CriticalCount:         s.critical.CriticalCount,
		AttackPower:           0,
		ArmorPenPercent:       0,
		CritDamageMultiplier:  1.0,
		BossDamagePercent:     0,
		AllDamageAmplify:      0,
	})
}

func loadoutBonuses(loadout Loadout) (attackPower int64, armorPen float64, critDmgMult float64, bossDmg float64) {
	items := []*InventoryItem{loadout.Weapon, loadout.Armor, loadout.Accessory}
	for _, item := range items {
		if item == nil {
			continue
		}
		attackPower += item.AttackPower
		armorPen += item.ArmorPenPercent
		critDmgMult += item.CritDamageMultiplier
		bossDmg += item.BossDamagePercent
	}
	return
}

func deriveCombatStats(stats CombatStats) CombatStats {
	stats.EffectiveIncrement = max(1, stats.AttackPower)
	stats.NormalDamage = stats.EffectiveIncrement

	if stats.CriticalCount <= 1 {
		stats.CriticalCount = 1
	}

	stats.CriticalDamage = max(stats.NormalDamage+stats.CriticalCount-1, stats.NormalDamage)

	if stats.CritDamageMultiplier < 1.0 {
		stats.CritDamageMultiplier = 1.0
	}

	return stats
}

// CalcBossPartDamage 计算对 Boss 部位的伤害（新减法公式）。
//   partType: 部位类型
//   partArmor: 部位护甲值
//   isBoss: 是否 Boss 目标（影响 Boss 增伤）
//   alivePartCount: 存活的部位数量（围剿技能用）
func CalcBossPartDamage(stats CombatStats, partType PartType, partArmor int64, isBoss bool, alivePartCount int) CombatStats {
	// 基础攻击力
	atk := max(1, stats.AttackPower)

	// 部位伤害系数
	coeff := partType.DamageCoefficient()

	// 有效护甲 = partArmor * (1 - 破甲率)，上限 80% 减免
	effectiveArmor := int64(float64(partArmor) * (1.0 - clampFloat(stats.ArmorPenPercent, 0, 0.80)))
	if effectiveArmor < 0 {
		effectiveArmor = 0
	}

	// 基础伤害 = max(攻击力 * 系数 - 护甲, 1)
	baseDamage := max(atk*int64(coeff*100)/100-effectiveArmor, 1)

	// 增伤乘区 = (1 + 全伤害增幅 + Boss 增伤)
	amplify := 1.0 + stats.AllDamageAmplify
	if isBoss {
		amplify += stats.BossDamagePercent
	}

	// 暴击乘区
	critMult := 1.0
	rollLimit, threshold := criticalRollPlan(stats.CriticalChancePercent)
	if rollLimit > 0 && stats.CriticalCount > 1 && s_roll(rollLimit) < threshold {
		critMult = max(1.0, stats.CritDamageMultiplier)
	}

	// 最终伤害
	finalDamage := int64(float64(baseDamage) * amplify * critMult)
	if finalDamage < 1 {
		finalDamage = 1
	}

	normalDamage := int64(float64(baseDamage) * amplify)
	if normalDamage < 1 {
		normalDamage = 1
	}

	criticalDamage := int64(float64(baseDamage) * amplify * critMult)
	if criticalDamage < 1 {
		criticalDamage = 1
	}

	return CombatStats{
		NormalDamage:          normalDamage,
		CriticalDamage:        criticalDamage,
		CriticalChancePercent: stats.CriticalChancePercent,
		CriticalCount:         stats.CriticalCount,
		AttackPower:           atk,
		ArmorPenPercent:       stats.ArmorPenPercent,
		CritDamageMultiplier:  critMult,
		AllDamageAmplify:      amplify - 1.0,
		BossDamagePercent:     stats.BossDamagePercent,
	}
}

// s_roll returns random int in [0, n), uses nil-safe global rand.
func s_roll(n int) int {
	if n <= 0 {
		return 0
	}
	return globalRand.IntN(n)
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

	var parts []BossPart
	if partsRaw, ok := values["parts"]; ok && partsRaw != "" {
		_ = sonic.Unmarshal([]byte(partsRaw), &parts)
	}

	return &Boss{
		ID:         id,
		TemplateID: strings.TrimSpace(values["template_id"]),
		Name:       name,
		Status:     strings.TrimSpace(values["status"]),
		MaxHP:      int64FromString(values["max_hp"]),
		CurrentHP:  int64FromString(values["current_hp"]),
		Parts:      parts,
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
		Rarity:                     normalizeEquipmentRarity(values["rarity"]),

		AttackPower:                int64FromString(values["attack_power"]),
		ArmorPenPercent:            float64FromString(values["armor_pen_percent"]),
		CritDamageMultiplier:       float64FromString(values["crit_damage_multiplier"]),
		BossDamagePercent:          float64FromString(values["boss_damage_percent"]),
		PartTypeDamageSoft:         float64FromString(values["part_type_damage_soft"]),
		PartTypeDamageHeavy:        float64FromString(values["part_type_damage_heavy"]),
		PartTypeDamageWeak:         float64FromString(values["part_type_damage_weak"]),
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

		item := buildInventoryItem(definition, quantities[itemID], true)

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

		items = append(items, buildInventoryItem(definition, quantity, equipped[itemID] != ""))
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
	recentRewards, err := s.recentRewardsForNickname(ctx, nickname)
	if err != nil {
		return nil, err
	}
	if len(recentRewards) == 0 {
		return nil, nil
	}

	return new(recentRewards[len(recentRewards)-1]), nil
}

func (s *Store) recentRewardsForNickname(ctx context.Context, nickname string) ([]Reward, error) {
	values, err := s.client.HGetAll(ctx, s.lastRewardKey(nickname)).Result()
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return []Reward{}, nil
	}

	if raw := strings.TrimSpace(values["recent_rewards"]); raw != "" {
		var rewards []Reward
		if err := sonic.Unmarshal([]byte(raw), &rewards); err == nil {
			normalized := make([]Reward, 0, len(rewards))
			for _, reward := range rewards {
				if strings.TrimSpace(reward.ItemID) == "" {
					continue
				}
				normalized = append(normalized, reward)
			}
			return normalized, nil
		}
	}

	legacyReward := rewardFromRecordValues(values)
	if legacyReward == nil {
		return []Reward{}, nil
	}

	return []Reward{*legacyReward}, nil
}

func rewardFromRecordValues(values map[string]string) *Reward {
	if len(values) == 0 || strings.TrimSpace(values["item_id"]) == "" {
		return nil
	}

	return &Reward{
		BossID:    strings.TrimSpace(values["boss_id"]),
		BossName:  strings.TrimSpace(values["boss_name"]),
		ItemID:    strings.TrimSpace(values["item_id"]),
		ItemName:  strings.TrimSpace(values["item_name"]),
		GrantedAt: int64FromString(values["granted_at"]),
	}
}

func rewardRecordValues(rewards []Reward) map[string]any {
	if len(rewards) == 0 {
		return map[string]any{}
	}

	lastReward := rewards[len(rewards)-1]
	encoded, err := sonic.Marshal(rewards)
	if err != nil {
		encoded = []byte("[]")
	}

	return map[string]any{
		"boss_id":        lastReward.BossID,
		"boss_name":      lastReward.BossName,
		"item_id":        lastReward.ItemID,
		"item_name":      lastReward.ItemName,
		"granted_at":     strconv.FormatInt(lastReward.GrantedAt, 10),
		"recent_rewards": string(encoded),
	}
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
			Rarity:                     normalizeEquipmentRarity(definition.Rarity),
			Weight:                     int64(entry.Score),
			DropRatePercent:            dropRatePercent,

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
		Key:       slug,
		RedisKey:  redisKey,
		Label:     label,
		Count:     int64Value(values, 1),
		Sort:      int(int64Value(values, 2)),
		Enabled:   stringValue(values, 3) != "0",
		Tags:      decodeStringList(stringValue(values, 4)),
		ImagePath: imagePath,
		ImageAlt:  imageAlt,
	}
}

func decodeStringList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var items []string
	if err := sonic.Unmarshal([]byte(raw), &items); err != nil {
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

	encoded, err := sonic.Marshal(normalized)
	if err != nil {
		return "[]"
	}
	return string(encoded)
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

func (s *Store) loadoutKey(nickname string) string {
	return s.loadoutPrefix + nickname
}


func (s *Store) gemKey(nickname string) string {
	return s.namespace + "gem:" + nickname
}

func (s *Store) gemsForNickname(ctx context.Context, nickname string) (int64, error) {
	val, err := s.client.HGet(ctx, s.gemKey(nickname), "gems").Result()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return int64FromString(val), nil
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


func (s *Store) equipmentKey(itemID string) string {
	return s.equipmentDefPrefix + itemID
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

func buildInventoryItem(definition EquipmentDefinition, quantity int64, equipped bool) InventoryItem {
	return InventoryItem{
		ItemID:              definition.ItemID,
		Name:                definition.Name,
		Slot:                definition.Slot,
		Rarity:              normalizeEquipmentRarity(definition.Rarity),
		ImagePath:           definition.ImagePath,
		ImageAlt:            definition.ImageAlt,
		Quantity:            quantity,
		Equipped:            equipped,
		AttackPower:         definition.AttackPower,
		ArmorPenPercent:     definition.ArmorPenPercent,
		CritDamageMultiplier: definition.CritDamageMultiplier,
		BossDamagePercent:   definition.BossDamagePercent,
		PartTypeDamageSoft:  definition.PartTypeDamageSoft,
		PartTypeDamageHeavy: definition.PartTypeDamageHeavy,
		PartTypeDamageWeak:  definition.PartTypeDamageWeak,
	}
}


func isRedisWrongTypeError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "WRONGTYPE")
}

func float64FromString(raw string) float64 {
	if strings.TrimSpace(raw) == "" {
		return 0
	}

	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0
	}

	return value
}

func clampFloat(value float64, min float64, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func ceilFloat(value float64) float64 {
	return math.Ceil(value)
}

func roundToDecimals(value float64, decimals int) float64 {
	if decimals <= 0 {
		return math.Round(value)
	}
	scale := math.Pow(10, float64(decimals))
	return math.Round(value*scale) / scale
}

func criticalRollPlan(chance float64) (int, int) {
	if chance <= 0 {
		return 100, 0
	}
	if roundToDecimals(chance, 6) == roundToDecimals(chance, 0) {
		return 100, int(chance)
	}
	return 600, int(roundToDecimals(chance*6, 0))
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
