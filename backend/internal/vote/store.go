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

	nicknamefilter "long/internal/nickname"
)

// 错误定义

var ErrInvalidNickname = errors.New("invalid nickname")
var ErrSensitiveNickname = errors.New("sensitive nickname")
var ErrSensitiveContent = errors.New("sensitive content")
var ErrEquipmentNotFound = errors.New("equipment not found")
var ErrEquipmentNotOwned = errors.New("equipment not owned")
var ErrEquipmentLocked = errors.New("equipment locked")
var ErrEquipmentEnhanceMaxLevel = errors.New("equipment enhance max level")
var ErrEquipmentEnhanceInsufficientGold = errors.New("equipment enhance insufficient gold")
var ErrEquipmentEnhanceInsufficientStones = errors.New("equipment enhance insufficient stones")
var ErrMessageEmpty = errors.New("message empty")
var ErrMessageTooLong = errors.New("message too long")
var ErrBossTemplateNotFound = errors.New("boss template not found")
var ErrBossPoolEmpty = errors.New("boss pool empty")
var ErrBossCycleQueueEmpty = errors.New("boss cycle queue empty")
var ErrBossPartsRequired = errors.New("boss parts required")
var ErrBossPartNotFound = errors.New("boss part not found")
var ErrBossPartAlreadyDead = errors.New("boss part already dead")
var ErrTalentAlreadyLearned = errors.New("talent already learned")
var ErrTalentPrerequisite = errors.New("talent prerequisite not met")
var ErrTalentNotFound = errors.New("talent not found")
var ErrInvalidTalentTree = errors.New("invalid talent tree")
var ErrTalentPointsInsufficient = errors.New("talent points insufficient")
var ErrTalentInvalidCost = errors.New("talent invalid cost")

const (
	bossStatusActive   = "active"
	bossStatusDefeated = "defeated"

	bossPartClickSlugPrefix = "boss-part:"
)

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
	X           int      `json:"x"`
	Y           int      `json:"y"`
	Type        PartType `json:"type"`
	DisplayName string   `json:"displayName,omitempty"`
	ImagePath   string   `json:"imagePath,omitempty"`
	MaxHP       int64    `json:"maxHp"`
	CurrentHP   int64    `json:"currentHp"`
	Armor       int64    `json:"armor"`
	Alive       bool     `json:"alive"`
}

// Boss 世界 Boss 状态
type Boss struct {
	ID                 string     `json:"id"`
	TemplateID         string     `json:"templateId,omitempty"`
	Name               string     `json:"name"`
	Status             string     `json:"status"`
	MaxHP              int64      `json:"maxHp"`
	CurrentHP          int64      `json:"currentHp"`
	GoldOnKill         int64      `json:"goldOnKill"`
	StoneOnKill        int64      `json:"stoneOnKill"`
	TalentPointsOnKill int64      `json:"talentPointsOnKill"`
	Parts              []BossPart `json:"parts,omitempty"`
	StartedAt          int64      `json:"startedAt,omitempty"`
	DefeatedAt         int64      `json:"defeatedAt,omitempty"`
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
	Rank     int    `json:"rank"`
}

// EquipmentDefinition 装备模板
type EquipmentDefinition struct {
	ItemID               string  `json:"itemId"`
	Name                 string  `json:"name"`
	Slot                 string  `json:"slot"`
	Description          string  `json:"description"`
	Rarity               string  `json:"rarity"`
	ImagePath            string  `json:"imagePath,omitempty"`
	ImageAlt             string  `json:"imageAlt,omitempty"`
	AttackPower          int64   `json:"attackPower,omitempty"`
	ArmorPenPercent      float64 `json:"armorPenPercent,omitempty"`
	CritRate             float64 `json:"critRate"` // 暴击率
	CritDamageMultiplier float64 `json:"critDamageMultiplier,omitempty"`
	PartTypeDamageSoft   float64 `json:"partTypeDamageSoft,omitempty"`  // 软组织增伤
	PartTypeDamageHeavy  float64 `json:"partTypeDamageHeavy,omitempty"` // 重甲增伤
	PartTypeDamageWeak   float64 `json:"partTypeDamageWeak,omitempty"`  // 弱点增伤
	TalentAffinity       string  `json:"talentAffinity,omitempty"`      // 天赋系绑定
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
	ItemID               string  `json:"itemId"`
	InstanceID           string  `json:"instanceId,omitempty"`
	Name                 string  `json:"name"`
	Slot                 string  `json:"slot"`
	Rarity               string  `json:"rarity"`
	ImagePath            string  `json:"imagePath,omitempty"`
	ImageAlt             string  `json:"imageAlt,omitempty"`
	Quantity             int64   `json:"quantity"`
	Equipped             bool    `json:"equipped"`
	EnhanceLevel         int     `json:"enhanceLevel,omitempty"`
	Bound                bool    `json:"bound,omitempty"`
	Locked               bool    `json:"locked,omitempty"`
	AttackPower          int64   `json:"attackPower,omitempty"`
	ArmorPenPercent      float64 `json:"armorPenPercent,omitempty"`
	CritRate             float64 `json:"critRate,omitempty"`
	CritDamageMultiplier float64 `json:"critDamageMultiplier,omitempty"`
	PartTypeDamageSoft   float64 `json:"partTypeDamageSoft,omitempty"`
	PartTypeDamageHeavy  float64 `json:"partTypeDamageHeavy,omitempty"`
	PartTypeDamageWeak   float64 `json:"partTypeDamageWeak,omitempty"`
}

// ItemInstance 装备实例
type ItemInstance struct {
	InstanceID   string `json:"instanceId"`
	ItemID       string `json:"itemId"`
	EnhanceLevel int    `json:"enhanceLevel"`
	SpentStones  int64  `json:"spentStones"`
	Bound        bool   `json:"bound"`
	Locked       bool   `json:"locked"`
	CreatedAt    int64  `json:"createdAt"`
}

// Loadout 已穿戴装备
type Loadout struct {
	Weapon    *InventoryItem `json:"weapon,omitempty"`
	Helmet    *InventoryItem `json:"helmet,omitempty"`
	Chest     *InventoryItem `json:"chest,omitempty"`
	Gloves    *InventoryItem `json:"gloves,omitempty"`
	Legs      *InventoryItem `json:"legs,omitempty"`
	Accessory *InventoryItem `json:"accessory,omitempty"`
}

// CombatStats 当前生效的点击战斗属性
type CombatStats struct {
	EffectiveIncrement    int64   `json:"effectiveIncrement"`
	NormalDamage          int64   `json:"normalDamage"`
	CriticalChancePercent float64 `json:"criticalChancePercent"`
	CriticalCount         int64   `json:"criticalCount"`
	CriticalDamage        int64   `json:"criticalDamage"`
	AttackPower           int64   `json:"attackPower"`
	ArmorPenPercent       float64 `json:"armorPenPercent"`
	CritDamageMultiplier  float64 `json:"critDamageMultiplier"`
	AllDamageAmplify      float64 `json:"allDamageAmplify"`
	PartTypeDamageSoft    float64 `json:"partTypeDamageSoft,omitempty"`
	PartTypeDamageHeavy   float64 `json:"partTypeDamageHeavy,omitempty"`
	PartTypeDamageWeak    float64 `json:"partTypeDamageWeak,omitempty"`
	PerPartDamagePercent  float64 `json:"perPartDamagePercent,omitempty"`
	LowHpMultiplier       float64 `json:"lowHpMultiplier,omitempty"`
	LowHpThreshold        float64 `json:"lowHpThreshold,omitempty"`
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
	ItemID               string  `json:"itemId"`
	ItemName             string  `json:"itemName"`
	Slot                 string  `json:"slot"`
	Rarity               string  `json:"rarity"`
	ImagePath            string  `json:"imagePath,omitempty"`
	ImageAlt             string  `json:"imageAlt,omitempty"`
	Weight               int64   `json:"weight"`
	DropRatePercent      float64 `json:"dropRatePercent"`
	AttackPower          int64   `json:"attackPower,omitempty"`
	ArmorPenPercent      float64 `json:"armorPenPercent,omitempty"`
	CritRate             float64 `json:"critRate,omitempty"`
	CritDamageMultiplier float64 `json:"critDamageMultiplier,omitempty"`
	PartTypeDamageSoft   float64 `json:"partTypeDamageSoft,omitempty"`
	PartTypeDamageHeavy  float64 `json:"partTypeDamageHeavy,omitempty"`
	PartTypeDamageWeak   float64 `json:"partTypeDamageWeak,omitempty"`
	TalentAffinity       string  `json:"talentAffinity,omitempty"`
}

// BossResources 描述当前 Boss 的低频公共资源。
type BossResources struct {
	BossID             string          `json:"bossId,omitempty"`
	TemplateID         string          `json:"templateId,omitempty"`
	Status             string          `json:"status,omitempty"`
	GoldRange          ResourceRange   `json:"goldRange"`
	StoneRange         ResourceRange   `json:"stoneRange"`
	TalentPointsOnKill int64           `json:"talentPointsOnKill"`
	BossLoot           []BossLootEntry `json:"bossLoot"`
}

// ResourceRange 掉落资源显示区间。
type ResourceRange struct {
	Min int64 `json:"min"`
	Max int64 `json:"max"`
}

// Snapshot 公共实时状态，广播给所有连接的客户端
type Snapshot struct {
	TotalVotes          int64                  `json:"totalVotes"`
	Leaderboard         []LeaderboardEntry     `json:"leaderboard"`
	Boss                *Boss                  `json:"boss,omitempty"`
	BossLeaderboard     []BossLeaderboardEntry `json:"bossLeaderboard"`
	AnnouncementVersion string                 `json:"announcementVersion,omitempty"`
}

// UserState 个人实时状态，只推送给对应昵称的连接
type UserState struct {
	UserStats         *UserStats          `json:"userStats,omitempty"`
	MyBossStats       *BossUserStats      `json:"myBossStats,omitempty"`
	Inventory         []InventoryItem     `json:"inventory"`
	Loadout           Loadout             `json:"loadout"`
	CombatStats       CombatStats         `json:"combatStats"`
	Gold              int64               `json:"gold"`
	Stones            int64               `json:"stones"`
	TalentPoints      int64               `json:"talentPoints"`
	RecentRewards     []Reward            `json:"recentRewards,omitempty"`
	TalentCombatState *TalentCombatState  `json:"talentCombatState,omitempty"`
}

// State 完整状态，包含个人统计与玩法状态
type State struct {
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
	Gold                int64                  `json:"gold"`
	Stones              int64                  `json:"stones"`
	TalentPoints        int64                  `json:"talentPoints"`
	RecentRewards       []Reward               `json:"recentRewards,omitempty"`
}

// AfkSettlement 挂机结算汇总。
type AfkSettlement struct {
	Kills      int64    `json:"kills"`
	GoldTotal  int64    `json:"goldTotal"`
	StoneTotal int64    `json:"stoneTotal"`
	StartedAt  int64    `json:"startedAt"`
	EndedAt    int64    `json:"endedAt"`
	Rewards    []Reward `json:"rewards,omitempty"`
}

// SalvageResult 装备分解结果。
type SalvageResult struct {
	ItemID         string `json:"itemId"`
	GoldReward     int64  `json:"goldReward"`
	StoneReward    int64  `json:"stoneReward"`
	RefundedStones int64  `json:"refundedStones"`
	Gold           int64  `json:"gold"`
	Stones         int64  `json:"stones"`
}

// BulkSalvageResult 一键分解结果。
type BulkSalvageResult struct {
	SalvagedCount       int            `json:"salvagedCount"`
	SalvagedByRarity    map[string]int `json:"salvagedByRarity,omitempty"`
	ExcludedEquipped    int            `json:"excludedEquipped"`
	ExcludedLocked      int            `json:"excludedLocked"`
	ExcludedTopRarity   int            `json:"excludedTopRarity"`
	GoldReward          int64          `json:"goldReward"`
	StoneReward         int64          `json:"stoneReward"`
	RefundedStones      int64          `json:"refundedStones"`
	Gold                int64          `json:"gold"`
	Stones              int64          `json:"stones"`
	HasEnhancedSalvaged bool           `json:"hasEnhancedSalvaged"`
}

// ClickResult 点击结果，包含更新后的增量与状态摘要
type ClickResult struct {
	Delta            int64                  `json:"delta"`
	BossDamage       int64                  `json:"bossDamage,omitempty"`
	DamageType       string                 `json:"damageType,omitempty"`
	Critical         bool                   `json:"critical"`
	UserStats        UserStats              `json:"userStats"`
	Boss             *Boss                  `json:"boss,omitempty"`
	BossLeaderboard  []BossLeaderboardEntry `json:"bossLeaderboard,omitempty"`
	MyBossStats      *BossUserStats         `json:"myBossStats,omitempty"`
	RecentRewards    []Reward               `json:"recentRewards,omitempty"`
	TalentEvents      []TalentTriggerEvent   `json:"talentEvents,omitempty"`
	TalentCombatState *TalentCombatState     `json:"talentCombatState,omitempty"`
	PartStateDeltas   []BossPartStateDelta   `json:"partStateDeltas,omitempty"`
	BroadcastUserAll  bool                   `json:"-"`
}

// TalentTriggerEvent 描述一次天赋触发事件，供前端战斗反馈显示。
type TalentTriggerEvent struct {
	TalentID    string `json:"talentId"`
	Name        string `json:"name"`
	EffectType  string `json:"effectType"`
	ExtraDamage int64  `json:"extraDamage,omitempty"`
	Message     string `json:"message,omitempty"`
	PartX       int    `json:"partX,omitempty"`
	PartY       int    `json:"partY,omitempty"`
}

// BossPartStateDelta 描述单次点击造成的部位变化增量。
type BossPartStateDelta struct {
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Damage   int64  `json:"damage"`
	BeforeHP int64  `json:"beforeHp"`
	AfterHP  int64  `json:"afterHp"`
	PartType string `json:"partType"`
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

// Store Redis 投票存储，管理按钮列表、点击计数、Boss 与装备状态
type Store struct {
	client                  redis.UniversalClient
	namespace               string
	equipmentIndexKey       string
	playerIndexKey          string
	userPrefix              string
	leaderboardKey          string
	bossCurrentKey          string
	bossHistoryKey          string
	bossHistoryPrefix       string
	bossTemplateIndexKey    string
	bossTemplatePrefix      string
	bossCycleKey            string
	bossInstanceSeqKey      string
	announcementSeqKey      string
	announcementKey         string
	announcementPrefix      string
	messageSeqKey           string
	messageKey              string
	messagePrefix           string
	equipmentDefPrefix      string
	equipmentInstancePrefix string
	inventoryPrefix         string
	playerInstancesPrefix   string
	loadoutPrefix           string
	lastRewardPrefix        string
	equipmentInstanceSeqKey string
	equipmentSpentPrefix    string
	equipmentEnhancePrefix  string
	critical                StoreOptions
	luaRunner               luaScriptRunner
	bossClickScript         *cachedLuaScript
	roll                    func(int) int
	now                     func() time.Time
	validator               interface{ Validate(string) error }
}

// NewStore 创建 Redis 投票存储实例
func NewStore(client redis.UniversalClient, namespace string, options StoreOptions, validator interface{ Validate(string) error }) *Store {
	luaCache := newLuaScriptCache()

	return &Store{
		client:                  client,
		namespace:               namespace,
		equipmentIndexKey:       namespace + "equipment:index",
		playerIndexKey:          namespace + "players:index",
		userPrefix:              namespace + "user:",
		leaderboardKey:          namespace + "leaderboard",
		bossCurrentKey:          namespace + "boss:current",
		bossHistoryKey:          namespace + "boss:history",
		bossHistoryPrefix:       namespace + "boss:history:",
		bossTemplateIndexKey:    namespace + "boss:pool:index",
		bossTemplatePrefix:      namespace + "boss:pool:",
		bossCycleKey:            namespace + "boss:cycle",
		bossInstanceSeqKey:      namespace + "boss:instance:seq",
		announcementSeqKey:      namespace + "announcement:seq",
		announcementKey:         namespace + "announcements",
		announcementPrefix:      namespace + "announcement:",
		messageSeqKey:           namespace + "message:seq",
		messageKey:              namespace + "messages",
		messagePrefix:           namespace + "message:",
		equipmentDefPrefix:      namespace + "equip:def:",
		equipmentInstancePrefix: namespace + "instance:",
		inventoryPrefix:         namespace + "user-inventory:",
		playerInstancesPrefix:   namespace + "player-instances:",
		loadoutPrefix:           namespace + "user-loadout:",
		lastRewardPrefix:        namespace + "user-last-reward:",
		equipmentInstanceSeqKey: namespace + "instance:seq",
		equipmentSpentPrefix:    namespace + "user-equipment-spent:",
		equipmentEnhancePrefix:  namespace + "user-equipment-enhance:",
		critical:                options,
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

// GetSnapshot 获取公共快照（公共排行榜 + Boss 状态）
func (s *Store) GetSnapshot(ctx context.Context) (Snapshot, error) {
	totalVotes, err := s.totalClickCount(ctx)
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

	resources, err := s.resourcesForNickname(ctx, normalizedNickname)
	if err != nil {
		return UserState{}, err
	}
	userState.Gold = resources.Gold
	userState.Stones = resources.Stones
	userState.TalentPoints = resources.TalentPoints

	userStats, err := s.GetUserStats(ctx, normalizedNickname)
	if err != nil {
		return UserState{}, err
	}
	userState.UserStats = &userStats

	loadout, equipped, err := s.loadoutForNickname(ctx, normalizedNickname)
	if err != nil {
		return UserState{}, err
	}
	userState.Loadout = loadout

	inventory, err := s.inventoryForNickname(ctx, normalizedNickname, equipped)
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

		combatState, _ := s.GetTalentCombatState(ctx, normalizedNickname, boss.ID)
		userState.TalentCombatState = combatState
	}

	return userState, nil
}

// ClickButton 处理 Boss 部位点击。slug 必须以 boss-part: 开头。
func (s *Store) ClickButton(ctx context.Context, slug string, nickname string, comboCount int64) (ClickResult, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return ClickResult{}, err
	}
	slug = strings.TrimSpace(slug)
	if !strings.HasPrefix(slug, bossPartClickSlugPrefix) {
		return ClickResult{}, fmt.Errorf("button not available")
	}
	return s.clickBossPart(ctx, slug, normalizedNickname, comboCount)
}

// ClickBossPart 处理不绑定按钮的 Boss 部位手动点击。
func (s *Store) ClickBossPart(ctx context.Context, target string, nickname string) (ClickResult, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return ClickResult{}, err
	}
	return s.clickBossPart(ctx, target, normalizedNickname, 0)
}

// EquipItem 穿戴一件装备实例。装备效果会影响平时点击与 Boss 伤害。
func (s *Store) EquipItem(ctx context.Context, nickname string, instanceID string) (State, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return State{}, err
	}

	instanceID = strings.TrimSpace(instanceID)
	if instanceID == "" {
		return State{}, ErrEquipmentNotFound
	}

	instance, err := s.getOwnedInstance(ctx, normalizedNickname, instanceID)
	if err != nil {
		return State{}, err
	}
	definition, err := s.getEquipmentDefinition(ctx, instance.ItemID)
	if err != nil {
		return State{}, err
	}

	now := time.Now().Unix()
	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, s.loadoutKey(normalizedNickname), definition.Slot, instance.InstanceID)
	pipe.ZAdd(ctx, s.playerIndexKey, redis.Z{
		Score:  float64(now),
		Member: normalizedNickname,
	})
	if _, err := pipe.Exec(ctx); err != nil {
		return State{}, err
	}

	return s.GetState(ctx, normalizedNickname)
}

// UnequipItem 卸下一件当前已穿戴的装备实例。
func (s *Store) UnequipItem(ctx context.Context, nickname string, instanceID string) (State, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return State{}, err
	}

	instanceID = strings.TrimSpace(instanceID)
	if instanceID == "" {
		return State{}, ErrEquipmentNotFound
	}

	instance, err := s.getOwnedInstance(ctx, normalizedNickname, instanceID)
	if err != nil {
		return State{}, err
	}
	definition, err := s.getEquipmentDefinition(ctx, instance.ItemID)
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

// EnhanceItem 强化一件装备实例，消耗金币和强化石并提升等级。
func (s *Store) EnhanceItem(ctx context.Context, nickname string, instanceID string) (State, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return State{}, err
	}

	instanceID = strings.TrimSpace(instanceID)
	if instanceID == "" {
		return State{}, ErrEquipmentNotFound
	}

	instance, err := s.getOwnedInstance(ctx, normalizedNickname, instanceID)
	if err != nil {
		return State{}, err
	}

	definition, err := s.getEquipmentDefinition(ctx, instance.ItemID)
	if err != nil {
		return State{}, err
	}

	maxLevel := maxEnhanceLevel(definition.Rarity)
	if instance.EnhanceLevel >= maxLevel {
		return State{}, ErrEquipmentEnhanceMaxLevel
	}

	goldCost := enhanceGoldCost(instance.EnhanceLevel)
	stoneCost := enhanceStoneCost(instance.EnhanceLevel)
	resources, err := s.resourcesForNickname(ctx, normalizedNickname)
	if err != nil {
		return State{}, err
	}
	if resources.Gold < goldCost {
		return State{}, ErrEquipmentEnhanceInsufficientGold
	}
	if resources.Stones < stoneCost {
		return State{}, ErrEquipmentEnhanceInsufficientStones
	}

	now := time.Now().Unix()
	pipe := s.client.TxPipeline()
	pipe.HIncrBy(ctx, s.resourceKey(normalizedNickname), "gold", -goldCost)
	pipe.HIncrBy(ctx, s.resourceKey(normalizedNickname), "stones", -stoneCost)
	pipe.HIncrBy(ctx, s.equipmentInstanceKey(instance.InstanceID), "spent_stones", stoneCost)
	pipe.HIncrBy(ctx, s.equipmentInstanceKey(instance.InstanceID), "enhance_level", 1)
	pipe.ZAdd(ctx, s.playerIndexKey, redis.Z{
		Score:  float64(now),
		Member: normalizedNickname,
	})
	if _, err := pipe.Exec(ctx); err != nil {
		return State{}, err
	}

	return s.GetState(ctx, normalizedNickname)
}

// LockItem 锁定一件装备实例。
func (s *Store) LockItem(ctx context.Context, nickname string, instanceID string) (State, error) {
	return s.setItemLockState(ctx, nickname, instanceID, true)
}

// UnlockItem 解锁一件装备实例。
func (s *Store) UnlockItem(ctx context.Context, nickname string, instanceID string) (State, error) {
	return s.setItemLockState(ctx, nickname, instanceID, false)
}

func (s *Store) setItemLockState(ctx context.Context, nickname string, instanceID string, locked bool) (State, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return State{}, err
	}

	instanceID = strings.TrimSpace(instanceID)
	if instanceID == "" {
		return State{}, ErrEquipmentNotFound
	}

	instance, err := s.getOwnedInstance(ctx, normalizedNickname, instanceID)
	if err != nil {
		return State{}, err
	}

	lockedValue := "0"
	if locked {
		lockedValue = "1"
	}

	now := time.Now().Unix()
	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, s.equipmentInstanceKey(instance.InstanceID), "locked", lockedValue)
	pipe.ZAdd(ctx, s.playerIndexKey, redis.Z{
		Score:  float64(now),
		Member: normalizedNickname,
	})
	if _, err := pipe.Exec(ctx); err != nil {
		return State{}, err
	}

	return s.GetState(ctx, normalizedNickname)
}

// SalvageItem 分解装备实例，按稀有度返还金币/强化石并返还已消耗强化石的 60%（向下取整）。
func (s *Store) SalvageItem(ctx context.Context, nickname string, instanceID string) (SalvageResult, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return SalvageResult{}, err
	}

	instanceID = strings.TrimSpace(instanceID)
	if instanceID == "" {
		return SalvageResult{}, ErrEquipmentNotFound
	}

	instance, err := s.getOwnedInstance(ctx, normalizedNickname, instanceID)
	if err != nil {
		return SalvageResult{}, err
	}
	if instance.Locked {
		return SalvageResult{}, ErrEquipmentLocked
	}
	definition, err := s.getEquipmentDefinition(ctx, instance.ItemID)
	if err != nil {
		return SalvageResult{}, err
	}

	goldReward, stoneReward := salvageBaseReward(definition.Rarity)
	refund := int64(math.Floor(float64(maxInt64(0, instance.SpentStones)) * 0.6))
	stoneGain := stoneReward + refund

	pipe := s.client.TxPipeline()
	pipe.SRem(ctx, s.playerInstancesKey(normalizedNickname), instance.InstanceID)
	pipe.Del(ctx, s.equipmentInstanceKey(instance.InstanceID))
	if goldReward > 0 {
		pipe.HIncrBy(ctx, s.resourceKey(normalizedNickname), "gold", goldReward)
	}
	if stoneGain > 0 {
		pipe.HIncrBy(ctx, s.resourceKey(normalizedNickname), "stones", stoneGain)
	}
	if definition.Slot != "" {
		equippedRef, getErr := s.client.HGet(ctx, s.loadoutKey(normalizedNickname), definition.Slot).Result()
		if getErr != nil && !errors.Is(getErr, redis.Nil) {
			return SalvageResult{}, getErr
		}
		if strings.TrimSpace(equippedRef) == instance.InstanceID {
			pipe.HDel(ctx, s.loadoutKey(normalizedNickname), definition.Slot)
		}
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return SalvageResult{}, err
	}

	resources, err := s.resourcesForNickname(ctx, normalizedNickname)
	if err != nil {
		return SalvageResult{}, err
	}
	return SalvageResult{
		ItemID:         instance.ItemID,
		GoldReward:     goldReward,
		StoneReward:    stoneReward,
		RefundedStones: refund,
		Gold:           resources.Gold,
		Stones:         resources.Stones,
	}, nil
}

// BulkSalvageUnequipped 一键分解所有“未穿戴、未锁定、且非至臻”的装备。
func (s *Store) BulkSalvageUnequipped(ctx context.Context, nickname string) (BulkSalvageResult, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return BulkSalvageResult{}, err
	}

	_, equipped, err := s.loadoutForNickname(ctx, normalizedNickname)
	if err != nil {
		return BulkSalvageResult{}, err
	}
	instances, err := s.itemInstancesByIDForNickname(ctx, normalizedNickname)
	if err != nil {
		return BulkSalvageResult{}, err
	}

	type salvageCandidate struct {
		instance ItemInstance
		slot     string
		rarity   string
	}

	result := BulkSalvageResult{
		SalvagedByRarity: map[string]int{},
	}
	candidates := make([]salvageCandidate, 0, len(instances))
	for _, instance := range instances {
		if equipped[instance.InstanceID] != "" {
			result.ExcludedEquipped++
			continue
		}
		if instance.Locked {
			result.ExcludedLocked++
			continue
		}

		definition, defErr := s.getEquipmentDefinition(ctx, instance.ItemID)
		if defErr != nil {
			continue
		}

		rarity := normalizeEquipmentRarity(definition.Rarity)
		if rarity == "至臻" {
			result.ExcludedTopRarity++
			continue
		}

		candidates = append(candidates, salvageCandidate{
			instance: instance,
			slot:     definition.Slot,
			rarity:   rarity,
		})
	}

	if len(candidates) == 0 {
		resources, resourceErr := s.resourcesForNickname(ctx, normalizedNickname)
		if resourceErr != nil {
			return BulkSalvageResult{}, resourceErr
		}
		result.Gold = resources.Gold
		result.Stones = resources.Stones
		return result, nil
	}

	for _, candidate := range candidates {
		goldReward, stoneReward := salvageBaseReward(candidate.rarity)
		refund := int64(math.Floor(float64(maxInt64(0, candidate.instance.SpentStones)) * 0.6))
		result.GoldReward += goldReward
		result.StoneReward += stoneReward
		result.RefundedStones += refund
		if refund > 0 {
			result.HasEnhancedSalvaged = true
		}
		result.SalvagedByRarity[candidate.rarity]++
	}
	result.SalvagedCount = len(candidates)

	pipe := s.client.TxPipeline()
	for _, candidate := range candidates {
		pipe.SRem(ctx, s.playerInstancesKey(normalizedNickname), candidate.instance.InstanceID)
		pipe.Del(ctx, s.equipmentInstanceKey(candidate.instance.InstanceID))
		if candidate.slot != "" {
			equippedRef, getErr := s.client.HGet(ctx, s.loadoutKey(normalizedNickname), candidate.slot).Result()
			if getErr != nil && !errors.Is(getErr, redis.Nil) {
				return BulkSalvageResult{}, getErr
			}
			if strings.TrimSpace(equippedRef) == candidate.instance.InstanceID {
				pipe.HDel(ctx, s.loadoutKey(normalizedNickname), candidate.slot)
			}
		}
	}
	if result.GoldReward > 0 {
		pipe.HIncrBy(ctx, s.resourceKey(normalizedNickname), "gold", result.GoldReward)
	}
	stoneGain := result.StoneReward + result.RefundedStones
	if stoneGain > 0 {
		pipe.HIncrBy(ctx, s.resourceKey(normalizedNickname), "stones", stoneGain)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return BulkSalvageResult{}, err
	}

	resources, err := s.resourcesForNickname(ctx, normalizedNickname)
	if err != nil {
		return BulkSalvageResult{}, err
	}
	result.Gold = resources.Gold
	result.Stones = resources.Stones
	return result, nil
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

func (s *Store) totalClickCount(ctx context.Context) (int64, error) {
	scores, err := s.client.ZRangeWithScores(ctx, s.leaderboardKey, 0, -1).Result()
	if err != nil {
		return 0, err
	}
	total := int64(0)
	for _, score := range scores {
		total += int64(score.Score)
	}
	return total, nil
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
		Gold:                userState.Gold,
		Stones:              userState.Stones,
		TalentPoints:        userState.TalentPoints,
		RecentRewards:       userState.RecentRewards,
	}
}

func (s *Store) applyClickCountOnly(ctx context.Context, nickname string, delta int64, critical bool) (ClickResult, error) {
	now := time.Now().Unix()
	pipe := s.client.TxPipeline()
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

	return ClickResult{
		Delta:    delta,
		Critical: critical,
		UserStats: UserStats{
			Nickname:   nickname,
			ClickCount: userCountCmd.Val(),
		},
	}, nil
}

func (s *Store) AutoClickBossPart(ctx context.Context, _ string, nickname string) (ClickResult, error) {
	return s.AttackBossPartAFK(ctx, nickname)
}

// AttackBossPartAFK 执行一次挂机攻击，不增加点击数，伤害按攻击力*0.5 向下取整。
func (s *Store) AttackBossPartAFK(ctx context.Context, nickname string) (ClickResult, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return ClickResult{}, err
	}
	boss, err := s.currentBoss(ctx)
	if err != nil {
		return ClickResult{}, err
	}
	if boss == nil || boss.Status != bossStatusActive || len(boss.Parts) == 0 {
		return ClickResult{}, nil
	}

	loadout, _, err := s.loadoutForNickname(ctx, normalizedNickname)
	if err != nil {
		return ClickResult{}, nil
	}
	combatStats, err := s.combatStatsForNickname(ctx, normalizedNickname, loadout)
	if err != nil {
		return ClickResult{}, nil
	}

	targetIdx := s.selectTargetPart(boss.Parts, normalizedNickname)
	if targetIdx < 0 {
		return ClickResult{}, nil
	}
	part := &boss.Parts[targetIdx]
	if !part.Alive || part.CurrentHP <= 0 {
		return ClickResult{}, nil
	}

	damage := int64(math.Floor(float64(maxInt64(0, combatStats.AttackPower)) * 0.5))
	if damage < 0 {
		damage = 0
	}
	actualDamage := damage
	if actualDamage > part.CurrentHP {
		actualDamage = part.CurrentHP
	}
	part.CurrentHP -= damage
	if part.CurrentHP < 0 {
		part.CurrentHP = 0
	}
	if part.CurrentHP <= 0 {
		part.Alive = false
	}

	boss.CurrentHP = sumBossPartCurrentHP(boss.Parts)
	allDead := true
	for _, p := range boss.Parts {
		if p.Alive {
			allDead = false
			break
		}
	}
	if allDead {
		boss.Status = bossStatusDefeated
		boss.DefeatedAt = s.now().Unix()
	}

	partsRaw, marshalErr := sonic.Marshal(boss.Parts)
	if marshalErr != nil {
		return ClickResult{}, nil
	}
	bossValues := map[string]any{
		"parts":      string(partsRaw),
		"current_hp": strconv.FormatInt(boss.CurrentHP, 10),
		"status":     boss.Status,
	}
	if boss.DefeatedAt != 0 {
		bossValues["defeated_at"] = strconv.FormatInt(boss.DefeatedAt, 10)
	}

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, s.bossCurrentKey, bossValues)
	if actualDamage > 0 {
		pipe.ZIncrBy(ctx, s.bossDamageKey(boss.ID), float64(actualDamage), normalizedNickname)
	}
	if _, execErr := pipe.Exec(ctx); execErr != nil {
		return ClickResult{}, nil
	}

	result := ClickResult{
		Delta:      0,
		Boss:       boss,
		BossDamage: actualDamage,
		DamageType: resolveBossDamageType(resolveBossDamageTypeInput{
			PartType:    part.Type,
			Critical:    false,
			BossDamage:  actualDamage,
			BossMaxHP:   boss.MaxHP,
			IsAfkAttack: true,
		}),
		UserStats: UserStats{
			Nickname: normalizedNickname,
		},
	}

	if allDead {
		result.BroadcastUserAll = true
		nextBoss, earnedRewards, finalizeErr := s.finalizeBossKill(ctx, boss, true, normalizedNickname)
		if finalizeErr != nil {
			return result, nil
		}
		if len(earnedRewards) > 0 {
			result.RecentRewards = earnedRewards
		}
		if nextBoss != nil {
			result.Boss = nextBoss
		}
	}
	return result, nil
}

func (s *Store) clickBossPart(ctx context.Context, target string, nickname string, comboCount int64) (ClickResult, error) {
	x, y, ok := parseBossPartClickTarget(target)
	if !ok {
		return ClickResult{}, ErrBossPartNotFound
	}

	boss, err := s.currentBoss(ctx)
	if err != nil {
		return ClickResult{}, err
	}
	if boss == nil || boss.Status != bossStatusActive || len(boss.Parts) == 0 {
		return ClickResult{}, ErrBossPartNotFound
	}

	targetIdx := findBossPartIndex(boss.Parts, x, y)
	if targetIdx < 0 {
		return ClickResult{}, ErrBossPartNotFound
	}
	part := boss.Parts[targetIdx]
	if !part.Alive || part.CurrentHP <= 0 {
		return ClickResult{}, ErrBossPartAlreadyDead
	}

	_, critical, err := s.nextIncrement(ctx, nickname)
	if err != nil {
		return ClickResult{}, err
	}
	result, err := s.applyClickCountOnly(ctx, nickname, 1, critical)
	if err != nil {
		return ClickResult{}, err
	}
	return s.applyBossPartDamage(ctx, boss, nickname, critical, result, targetIdx, comboCount)
}

func (s *Store) applyBossPartDamage(ctx context.Context, boss *Boss, nickname string, critical bool, result ClickResult, targetIdx int, comboCount int64) (ClickResult, error) {
	loadout, _, err := s.loadoutForNickname(ctx, nickname)
	if err != nil {
		return result, nil
	}

	combatStats, err := s.combatStatsForNickname(ctx, nickname, loadout)
	if err != nil {
		return result, nil
	}

	if targetIdx < 0 {
		targetIdx = s.selectTargetPart(boss.Parts, nickname)
	}
	if targetIdx < 0 {
		return result, nil
	}
	part := &boss.Parts[targetIdx]
	if !part.Alive || part.CurrentHP <= 0 {
		return result, ErrBossPartAlreadyDead
	}

	now := s.now().Unix()

	aliveCount := 0
	for _, p := range boss.Parts {
		if p.Alive {
			aliveCount++
		}
	}

	talentState, _ := s.GetTalentState(ctx, nickname)
	learned := make(map[string]struct{})
	if talentState != nil {
		for _, id := range talentState.Talents {
			learned[id] = struct{}{}
		}
	}

	combatState, _ := s.GetTalentCombatState(ctx, nickname, boss.ID)
	if combatState == nil {
		combatState = NewTalentCombatState()
	}

	effectivePartType := part.Type
	if combatState.SilverStormActive {
		effectivePartType = PartTypeSoft
	}
	if combatState.DeathEcstasyEndsAt > now {
		effectivePartType = PartTypeWeak
	}
	partKey := TalentPartKey(part.X, part.Y)
	if endsAt, ok := combatState.SkinnerParts[partKey]; ok && now < endsAt {
		effectivePartType = PartTypeWeak
	}

	effectiveArmor := part.Armor
	inCollapse := false
	for _, idx := range combatState.CollapseParts {
		if idx == targetIdx {
			effectiveArmor = 0
			inCollapse = true
			break
		}
	}

	damageStats := CalcBossPartDamage(combatStats, effectivePartType, effectiveArmor, aliveCount, boss.CurrentHP, boss.MaxHP)
	partDamage := damageStats.NormalDamage
	if comboCount >= 50 {
		comboAmplify := float64(comboCount/50) * 0.05
		partDamage = int64(float64(partDamage) * (1.0 + comboAmplify))
	}
	if critical {
		partDamage = damageStats.CriticalDamage
		if combatState.DeathEcstasyEndsAt > now {
			partDamage = int64(float64(partDamage) * 3.0)
		}
		if combatState.DoomCritBuff {
			partDamage = int64(float64(partDamage) * 3.0)
		}
	}

	if inCollapse && hasTalent(learned, "armor_ruin") {
		partDamage = int64(float64(partDamage) * 2.0)
	}

	hpRatio := float64(part.CurrentHP) / float64(maxInt64(1, boss.MaxHP))
	if hasTalent(learned, "crit_omen_kill") && hpRatio < 0.35 && combatState.OmenStacks > 0 {
		partDamage = int64(float64(partDamage) * (1.0 + float64(combatState.OmenStacks)*0.01))
	}

	if critical && hasTalent(learned, "crit_omen_resonate") && combatState.OmenStacks > 0 {
		partDamage = int64(float64(partDamage) * (1.0 + float64(combatState.OmenStacks)*0.003))
	}

	beforeHP := part.CurrentHP
	actualDamage := partDamage
	if actualDamage > part.CurrentHP {
		actualDamage = part.CurrentHP
	}
	if actualDamage < 0 {
		actualDamage = 0
	}
	part.CurrentHP -= partDamage
	if part.CurrentHP < 0 {
		part.CurrentHP = 0
	}
	partWasAlive := part.Alive
	if part.CurrentHP <= 0 {
		part.Alive = false
	}
	partJustDied := partWasAlive && !part.Alive

	if critical && effectivePartType == PartTypeWeak && partWasAlive && hasTalent(learned, "crit_core") {
		combatState.OmenStacks++
	}

	if critical && part.Type != PartTypeWeak && effectivePartType != PartTypeWeak && hasTalent(learned, "crit_skinner") {
		if s.roll != nil && s.roll(100) < 30 {
			combatState.SkinnerParts[partKey] = now + 5
		}
	}

	boss.CurrentHP = sumBossPartCurrentHP(boss.Parts)
	totalDamage := actualDamage
	result.PartStateDeltas = append(result.PartStateDeltas, BossPartStateDelta{
		X:        part.X,
		Y:        part.Y,
		Damage:   actualDamage,
		BeforeHP: beforeHP,
		AfterHP:  part.CurrentHP,
		PartType: string(part.Type),
	})

	extraDamage, talentEvents, damageTypeOverride := s.applyTriggeredTalentDamage(ctx, boss, part, nickname, result.UserStats.ClickCount, actualDamage, critical, targetIdx, learned, combatState, now)
	if extraDamage > 0 {
		totalDamage += extraDamage
		result.PartStateDeltas = append(result.PartStateDeltas, BossPartStateDelta{
			X:        part.X,
			Y:        part.Y,
			Damage:   extraDamage,
			BeforeHP: part.CurrentHP + extraDamage,
			AfterHP:  part.CurrentHP,
			PartType: string(part.Type),
		})
	}
	if len(talentEvents) > 0 {
		result.TalentEvents = append(result.TalentEvents, talentEvents...)
	}

	if partJustDied && hasTalent(learned, "normal_ultimate") {
		combatState.SilverStormActive = true
		combatState.SilverStormRemaining = 15
		result.TalentEvents = append(result.TalentEvents, TalentTriggerEvent{
			TalentID:   "normal_ultimate",
			Name:       "silverstorm",
			EffectType: "silver_storm",
			Message:    "白银风暴激活！15 次攻击视为软组织",
		})
	}
	if combatState.SilverStormActive {
		combatState.SilverStormRemaining--
		if combatState.SilverStormRemaining <= 0 {
			combatState.SilverStormActive = false
			combatState.SilverStormRemaining = 0
		}
	}

	if hasTalent(learned, "crit_ultimate") && len(combatState.DoomMarks) == 0 && len(boss.Parts) >= 2 {
		combatState.DoomMarks = randomMarkIndices(len(boss.Parts), 2, s.roll)
	}

	result.BossDamage = totalDamage
	result.Critical = critical
	result.DamageType = resolveBossDamageType(resolveBossDamageTypeInput{
		PartType:    part.Type,
		Critical:    critical,
		BossDamage:  totalDamage,
		BossMaxHP:   boss.MaxHP,
		IsAfkAttack: false,
	})
	if damageTypeOverride != "" {
		result.DamageType = damageTypeOverride
	}

	_ = s.SaveTalentCombatState(ctx, nickname, boss.ID, combatState)
	result.TalentCombatState = combatState

	allDead := true
	for _, p := range boss.Parts {
		if p.Alive {
			allDead = false
			break
		}
	}

	if allDead {
		boss.Status = bossStatusDefeated
		boss.DefeatedAt = s.now().Unix()
	}

	partsRaw, marshalErr := sonic.Marshal(boss.Parts)
	if marshalErr != nil {
		return result, nil
	}

	bossValues := map[string]any{
		"parts":      string(partsRaw),
		"current_hp": strconv.FormatInt(boss.CurrentHP, 10),
		"status":     boss.Status,
	}
	if boss.DefeatedAt != 0 {
		bossValues["defeated_at"] = strconv.FormatInt(boss.DefeatedAt, 10)
	}

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, s.bossCurrentKey, bossValues)
	if totalDamage > 0 {
		pipe.ZIncrBy(ctx, s.bossDamageKey(boss.ID), float64(totalDamage), nickname)
	}
	if _, execErr := pipe.Exec(ctx); execErr != nil {
		return result, nil
	}

	result.Boss = boss

	if allDead {
		result.BroadcastUserAll = true
		nextBoss, _, finalizeErr := s.finalizeBossKill(ctx, boss, false, "")
		if finalizeErr != nil {
			return result, nil
		}
		result.Boss = nextBoss
	}

	return result, nil
}

func (s *Store) applyTriggeredTalentDamage(ctx context.Context, boss *Boss, part *BossPart, nickname string, clickCount int64, baseDamage int64, isCritical bool, partIndex int, learned map[string]struct{}, combatState *TalentCombatState, now int64) (int64, []TalentTriggerEvent, string) {
	if boss == nil || part == nil || strings.TrimSpace(nickname) == "" || clickCount <= 0 {
		return 0, nil, ""
	}

	var totalExtra int64
	var events []TalentTriggerEvent
	var damageTypeOverride string

	if hasTalent(learned, "normal_core") {
		def, ok := talentDefs["normal_core"]
		if ok {
			triggerCount := int64(TalentNormalStormTriggerCount)
			extraHits := int64(TalentNormalStormExtraHits)
			chaseRatio := TalentNormalStormChaseRatio
			val, _ := def.EffectValue.(map[string]any)
			if v, ok := val["triggerCount"].(float64); ok && v > 0 { triggerCount = int64(v) }
			if v, ok := val["extraHits"].(float64); ok && v > 0 { extraHits = int64(v) }
			if v, ok := val["chaseRatio"].(float64); ok && v > 0 { chaseRatio = v }
			if hasTalent(learned, "normal_chase_up") {
				if def2, ok2 := talentDefs["normal_chase_up"]; ok2 {
					if val2, ok3 := def2.EffectValue.(map[string]any); ok3 {
						if v, ok4 := val2["chaseRatio"].(float64); ok4 && v > chaseRatio { chaseRatio = v }
					}
				}
			}
			if hasTalent(learned, "normal_combo_ext") {
				if def2, ok2 := talentDefs["normal_combo_ext"]; ok2 {
					if val2, ok3 := def2.EffectValue.(map[string]any); ok3 {
						if v, ok4 := val2["extraHits"].(float64); ok4 && v > 0 { extraHits += int64(v) }
					}
				}
			}

			partKey := TalentPartKey(part.X, part.Y)
			combatState.PartStormComboCount[partKey]++
			if combatState.PartStormComboCount[partKey] >= triggerCount {
				burst := int64(math.Floor(float64(maxInt64(1, baseDamage)) * chaseRatio * float64(maxInt64(1, extraHits))))
				if burst > 0 {
					if burst > part.CurrentHP { burst = part.CurrentHP }
					part.CurrentHP -= burst
					if part.CurrentHP <= 0 { part.CurrentHP = 0; part.Alive = false }
					boss.CurrentHP = sumBossPartCurrentHP(boss.Parts)
					totalExtra += burst
					events = append(events, TalentTriggerEvent{
						TalentID: "normal_core", Name: def.Name, EffectType: def.EffectType,
						ExtraDamage: burst, Message: fmt.Sprintf("追击爆发 %d 段伤害", extraHits),
						PartX: part.X, PartY: part.Y,
					})
					if hasTalent(learned, "normal_charge") {
						combatState.PartStormComboCount[partKey] = int64(float64(triggerCount) * 0.30)
					} else {
						combatState.PartStormComboCount[partKey] = 0
					}
				}
			}
		}
	}

	if hasTalent(learned, "armor_core") && part.Type == PartTypeHeavy {
		partKey := TalentPartKey(part.X, part.Y)
		combatState.PartHeavyClickCount[partKey]++
		if combatState.PartHeavyClickCount[partKey] >= 100 {
			cd := int64(8)
			if hasTalent(learned, "armor_collapse_ext") { cd = 15 }
			combatState.CollapseParts = append(combatState.CollapseParts, partIndex)
			combatState.CollapseEndsAt = now + cd
			events = append(events, TalentTriggerEvent{
				TalentID: "armor_core", Name: "灭绝穿甲", EffectType: "collapse_trigger",
				Message:  fmt.Sprintf("结构崩塌！护甲归零 %d 秒", cd),
				PartX:    part.X,
				PartY:    part.Y,
			})
			combatState.PartHeavyClickCount[partKey] = 0 // 触发后归零，允许多次崩塌
		}
	}

	if hasTalent(learned, "armor_auto_strike") && now-combatState.LastAutoStrikeAt >= 20 {
		var best *BossPart
		for i := range boss.Parts {
			p := &boss.Parts[i]
			if !p.Alive || p.Type != PartTypeHeavy { continue }
			if best == nil || p.CurrentHP > best.CurrentHP { best = p }
		}
		if best != nil {
			sd := int64(float64(baseDamage) * 3.0)
			if sd > best.CurrentHP { sd = best.CurrentHP }
			best.CurrentHP -= sd
			if best.CurrentHP <= 0 { best.CurrentHP = 0; best.Alive = false }
			combatState.LastAutoStrikeAt = now
			totalExtra += sd
			events = append(events, TalentTriggerEvent{
				TalentID: "armor_auto_strike", Name: "自动打击触发", EffectType: "auto_strike",
				ExtraDamage: sd, Message: "自动打击触发",
			})
			damageTypeOverride = "trueDamage"
		}
	}

	if hasTalent(learned, "armor_ultimate") && part.Type == PartTypeHeavy {
		pk := TalentPartKey(part.X, part.Y)
		if !combatState.JudgmentDayUsed[pk] && combatState.PartHeavyClickCount[pk] >= 100 {
			combatState.JudgmentDayUsed[pk] = true
			cd := boss.MaxHP / 2
			if cd > part.CurrentHP { cd = part.CurrentHP }
			part.CurrentHP -= cd
			if part.CurrentHP <= 0 { part.CurrentHP = 0; part.Alive = false }
			boss.CurrentHP = sumBossPartCurrentHP(boss.Parts)
			totalExtra += cd
			events = append(events, TalentTriggerEvent{
				TalentID: "armor_ultimate", Name: "审判日触发！削除 50% 最大生命", EffectType: "judgment_day",
				ExtraDamage: cd, Message: "审判日触发！削除 50% 最大生命",
			})
			damageTypeOverride = "judgement"
		}
	}

	if hasTalent(learned, "crit_bleed") && isCritical {
		if bd := int64(float64(baseDamage) * 0.60); bd > 0 {
			totalExtra += bd
			events = append(events, TalentTriggerEvent{
				TalentID: "crit_bleed", Name: "致命出血", EffectType: "bleed",
				ExtraDamage: bd, Message: "致命出血",
				PartX: part.X, PartY: part.Y,
			})
		}
	}

	if hasTalent(learned, "crit_omen_reap") && combatState.OmenStacks >= 30 {
		rd := int64(float64(baseDamage) * 2.0)
		if rd > part.CurrentHP { rd = part.CurrentHP }
		part.CurrentHP -= rd
		if part.CurrentHP <= 0 { part.CurrentHP = 0; part.Alive = false }
		boss.CurrentHP = sumBossPartCurrentHP(boss.Parts)
		totalExtra += rd
		events = append(events, TalentTriggerEvent{
			TalentID: "crit_omen_reap", Name: "死兆收割", EffectType: "omen_harvest",
			ExtraDamage: rd, Message: "死兆收割",
			PartX: part.X, PartY: part.Y,
		})
		if rd > baseDamage*5 { damageTypeOverride = "doomsday" }
	}

	if hasTalent(learned, "crit_final_cut") && isCritical {
		combatState.CritCount++
		if combatState.CritCount >= 120 && now-combatState.LastFinalCutAt >= 30 {
			combatState.LastFinalCutAt = now
			cd := int64(float64(boss.MaxHP) * 0.12)
			if cd > part.CurrentHP { cd = part.CurrentHP }
			part.CurrentHP -= cd
			if part.CurrentHP <= 0 { part.CurrentHP = 0; part.Alive = false }
			boss.CurrentHP = sumBossPartCurrentHP(boss.Parts)
			totalExtra += cd
			events = append(events, TalentTriggerEvent{
				TalentID: "crit_final_cut", Name: "终末血斩！", EffectType: "final_cut",
				ExtraDamage: cd, Message: "终末血斩！",
				PartX: part.X, PartY: part.Y,
			})
			damageTypeOverride = "doomsday"
		}
	}

	if hasTalent(learned, "crit_death_ecstasy") && combatState.OmenStacks >= 50 && combatState.DeathEcstasyEndsAt <= now {
		combatState.OmenStacks -= 50
		combatState.DeathEcstasyEndsAt = now + 6
		events = append(events, TalentTriggerEvent{
			TalentID: "crit_death_ecstasy", Name: "死亡狂喜激活！6 秒内暴伤 +200%，全攻击视为弱点", EffectType: "death_ecstasy",
			Message: "死亡狂喜激活！6 秒内暴伤 +200%，全攻击视为弱点",
			PartX: part.X, PartY: part.Y,
		})
	}

	if hasTalent(learned, "crit_ultimate") && !part.Alive && len(combatState.DoomMarks) > 0 {
		for _, idx := range combatState.DoomMarks {
			if idx == partIndex {
				combatState.DoomDestroyed++
				if combatState.DoomDestroyed == 1 {
					combatState.OmenStacks += 100
					combatState.DoomCritBuff = true
					events = append(events, TalentTriggerEvent{
						TalentID: "crit_ultimate", Name: "末日审判", EffectType: "doom_judgment",
						Message: "终结标记击碎！+100 死兆，下次暴伤 x3",
						PartX: part.X, PartY: part.Y,
					})
					damageTypeOverride = "doomsday"
				} else if combatState.DoomDestroyed >= 2 && combatState.DoomCritBuff {
					events = append(events, TalentTriggerEvent{
						TalentID: "crit_ultimate", Name: "末日审判", EffectType: "doom_judgment",
						Message: "双标记击碎！暴伤 x6",
						PartX: part.X, PartY: part.Y,
					})
					damageTypeOverride = "doomsday"
				}
				break
			}
		}
	}

	if combatState.CollapseEndsAt > 0 && now >= combatState.CollapseEndsAt {
		for _, idx := range combatState.CollapseParts {
			if idx >= 0 && idx < len(boss.Parts) {
				pk := TalentPartKey(boss.Parts[idx].X, boss.Parts[idx].Y)
				combatState.PartHeavyClickCount[pk] = 0
			}
		}
		combatState.CollapseParts = nil
		combatState.CollapseEndsAt = 0
	}
	if combatState.DeathEcstasyEndsAt > 0 && now >= combatState.DeathEcstasyEndsAt {
		combatState.DeathEcstasyEndsAt = 0
	}
	if hasTalent(learned, "crit_omen_reap") && combatState.OmenStacks >= 30 {
		combatState.OmenStacks -= 30
	}

	return totalExtra, events, damageTypeOverride
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

func parseBossPartClickTarget(target string) (int, int, bool) {
	raw := strings.TrimSpace(target)
	raw = strings.TrimPrefix(raw, bossPartClickSlugPrefix)
	parts := strings.Split(raw, "-")
	if len(parts) != 2 {
		return 0, 0, false
	}
	x, xErr := strconv.Atoi(strings.TrimSpace(parts[0]))
	y, yErr := strconv.Atoi(strings.TrimSpace(parts[1]))
	if xErr != nil || yErr != nil || x < 0 || x >= 5 || y < 0 || y >= 5 {
		return 0, 0, false
	}
	return x, y, true
}

func findBossPartIndex(parts []BossPart, x int, y int) int {
	for index, part := range parts {
		if part.X == x && part.Y == y {
			return index
		}
	}
	return -1
}

func bossPartClickSlug(x int, y int) string {
	return bossPartClickSlugPrefix + strconv.Itoa(x) + "-" + strconv.Itoa(y)
}

func bossPartDisplayLabel(part BossPart) string {
	if label := strings.TrimSpace(part.DisplayName); label != "" {
		return label
	}
	switch part.Type {
	case PartTypeSoft:
		return "软组织"
	case PartTypeHeavy:
		return "重甲"
	case PartTypeWeak:
		return "弱点"
	default:
		return string(part.Type)
	}
}

type resolveBossDamageTypeInput struct {
	PartType    PartType
	Critical    bool
	BossDamage  int64
	BossMaxHP   int64
	IsAfkAttack bool
}

func resolveBossDamageType(input resolveBossDamageTypeInput) string {
	damage := maxInt64(0, input.BossDamage)
	maxHP := maxInt64(1, input.BossMaxHP)
	damageRatio := float64(damage) / float64(maxHP)

	if damageRatio >= 0.2 {
		return "doomsday"
	}
	if input.Critical && damageRatio >= 0.11 {
		return "judgement"
	}
	if input.Critical && input.PartType == PartTypeWeak {
		return "weakCritical"
	}
	if input.Critical {
		return "critical"
	}
	if input.PartType == PartTypeHeavy {
		return "trueDamage"
	}
	if input.IsAfkAttack {
		return "pursuit"
	}
	return "normal"
}

func (s *Store) finalizeBossKill(ctx context.Context, boss *Boss, afkMode bool, rewardNickname string) (*Boss, []Reward, error) {
	if boss == nil || strings.TrimSpace(boss.ID) == "" {
		return nil, nil, nil
	}
	bossID := strings.TrimSpace(boss.ID)
	bossName := strings.TrimSpace(boss.Name)
	rewardNickname = strings.TrimSpace(rewardNickname)

	acquired, err := s.client.SetNX(ctx, s.bossRewardLockKey(bossID), "1", 0).Result()
	if err != nil {
		return nil, nil, err
	}
	if !acquired {
		current, currentErr := s.currentBoss(ctx)
		return current, nil, currentErr
	}

	lootEntries, err := s.loadBossLoot(ctx, bossID)
	if err != nil {
		return nil, nil, err
	}
	participants, err := s.client.ZRevRangeWithScores(ctx, s.bossDamageKey(bossID), 0, -1).Result()
	if err != nil {
		return nil, nil, err
	}

	pipe := s.client.Pipeline()
	now := s.now().Unix()
	minDamage := (maxInt64(1, boss.MaxHP) + 99) / 100
	goldBase := boss.GoldOnKill
	stoneBase := boss.StoneOnKill
	talentPointBase := maxInt64(0, boss.TalentPointsOnKill)
	if afkMode {
		goldBase = int64(math.Floor(float64(goldBase) * 0.5))
		stoneBase = int64(math.Floor(float64(stoneBase) * 0.5))
	}
	rewardForNickname := make([]Reward, 0, len(lootEntries))
	for _, participant := range participants {
		nickname, ok := participant.Member.(string)
		if !ok || nickname == "" || participant.Score < float64(minDamage) {
			continue
		}

		goldDelta := rollResourceReward(s.roll, goldBase, 0.75, 1.25)
		stoneDelta := rollResourceReward(s.roll, stoneBase, 0.67, 1.33)
		if goldDelta > 0 {
			pipe.HIncrBy(ctx, s.resourceKey(nickname), "gold", goldDelta)
		}
		if stoneDelta > 0 {
			pipe.HIncrBy(ctx, s.resourceKey(nickname), "stones", stoneDelta)
		}
		if talentPointBase > 0 {
			pipe.HIncrBy(ctx, s.resourceKey(nickname), "talent_points", talentPointBase)
		}

		if len(lootEntries) == 0 {
			continue
		}
		rewards := make([]Reward, 0, len(lootEntries))
		for _, reward := range s.rollLootDrops(lootEntries) {
			instanceID, createErr := s.newEquipmentInstanceID(ctx)
			if createErr != nil {
				return nil, nil, createErr
			}
			pipe.HSet(ctx, s.equipmentInstanceKey(instanceID), map[string]any{
				"item_id":       reward.ItemID,
				"enhance_level": "0",
				"spent_stones":  "0",
				"bound":         "0",
				"locked":        "0",
				"created_at":    strconv.FormatInt(now, 10),
			})
			pipe.SAdd(ctx, s.playerInstancesKey(nickname), instanceID)
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
			if rewardNickname != "" && nickname == rewardNickname {
				rewardForNickname = append(rewardForNickname, rewards...)
			}
		}
	}

	if _, err = pipe.Exec(ctx); err != nil {
		return nil, nil, err
	}

	if err := s.SaveBossToHistory(ctx, boss); err != nil {
		return nil, nil, err
	}

	enabled, err := s.bossCycleEnabled(ctx)
	if err != nil {
		return nil, nil, err
	}
	if enabled {
		nextBoss, err := s.activateNextBossFromCycle(ctx, boss.TemplateID)
		if err != nil && !errors.Is(err, ErrBossPoolEmpty) && !errors.Is(err, ErrBossCycleQueueEmpty) {
			return nil, nil, err
		}
		if nextBoss != nil {
			return nextBoss, rewardForNickname, nil
		}
	}

	current, currentErr := s.currentBoss(ctx)
	return current, rewardForNickname, currentErr
}

func rollResourceReward(roller func(int) int, base int64, minMultiplier float64, maxMultiplier float64) int64 {
	if base <= 0 {
		return 0
	}
	const rollSteps = 10000
	delta := maxMultiplier - minMultiplier
	if delta < 0 {
		delta = 0
	}
	roll := 0
	if roller != nil {
		roll = roller(rollSteps)
	}
	multiplier := minMultiplier + float64(max(0, roll))*delta/float64(rollSteps)
	result := int64(math.Floor(float64(base) * multiplier))
	if result < 0 {
		return 0
	}
	return result
}

func (s *Store) rollLootDrops(entries []BossLootEntry) []BossLootEntry {
	if len(entries) == 0 {
		return nil
	}

	drops := make([]BossLootEntry, 0, len(entries))
	for _, entry := range entries {
		threshold := dropRateThreshold(entry.DropRatePercent)
		if threshold <= 0 {
			continue
		}
		if threshold >= dropRateRollLimit || s.roll(dropRateRollLimit) < threshold {
			drops = append(drops, entry)
		}
	}

	return drops
}

func (s *Store) nextIncrement(ctx context.Context, nickname string) (int64, bool, error) {
	loadout, _, err := s.loadoutForNickname(ctx, nickname)
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

	if combatStats.CriticalChancePercent <= 0 || !hasCriticalBonus(combatStats) {
		return delta, false, nil
	}

	rollLimit, threshold := criticalRollPlan(combatStats.CriticalChancePercent)
	if rollLimit <= 0 {
		return delta, false, nil
	}
	if s.roll(rollLimit) < threshold {
		return combatStats.CriticalDamage, true, nil
	}

	return delta, false, nil
}

func (s *Store) combatStatsForNickname(ctx context.Context, nickname string, loadout Loadout) (CombatStats, error) {
	stats := s.baseCombatStats()

	attackPower, armorPen, critRate, critDmgMult := loadoutBonuses(loadout)
	stats.AttackPower += attackPower
	stats.ArmorPenPercent = clampFloat(stats.ArmorPenPercent+armorPen, 0, 0.80)
	stats.CriticalChancePercent = clampFloat(stats.CriticalChancePercent+critRate*100, 0, 100)
	stats.CritDamageMultiplier += critDmgMult

	mods, err := s.ComputeTalentModifiers(ctx, nickname)
	if err != nil {
		return CombatStats{}, err
	}
	if mods != nil {
		stats.AttackPower = int64(float64(stats.AttackPower) * (1 + max(0.0, mods.AttackPowerPercent)))
		stats.ArmorPenPercent = clampFloat(stats.ArmorPenPercent+mods.ArmorPenExtra, 0, 0.80)
		stats.AllDamageAmplify += mods.AllDamageAmplify
		stats.CritDamageMultiplier += mods.CritDamagePercentBonus
		stats.PerPartDamagePercent = max(0.0, mods.PerPartDamagePercent)
		stats.LowHpMultiplier = max(1.0, mods.LowHpMultiplier)
		stats.LowHpThreshold = clampFloat(mods.LowHpThreshold, 0, 1)
		stats.PartTypeDamageSoft += max(0.0, mods.PartTypeBonus[PartTypeSoft])
		stats.PartTypeDamageHeavy += max(0.0, mods.PartTypeBonus[PartTypeHeavy])
		stats.PartTypeDamageWeak += max(0.0, mods.PartTypeBonus[PartTypeWeak])
	}

	result := deriveCombatStats(stats)
	return result, nil
}

func (s *Store) baseCombatStats() CombatStats {
	return deriveCombatStats(CombatStats{
		CriticalChancePercent: clampFloat(float64(s.critical.CriticalChancePercent), 0, 100),
		CriticalCount:         s.critical.CriticalCount,
		AttackPower:           5,
		ArmorPenPercent:       0,
		CritDamageMultiplier:  1.5,
		AllDamageAmplify:      0,
		LowHpMultiplier:       1,
	})
}

func loadoutBonuses(loadout Loadout) (attackPower int64, armorPen float64, critRate float64, critDmgMult float64) {
	items := []*InventoryItem{
		loadout.Weapon,
		loadout.Helmet,
		loadout.Chest,
		loadout.Gloves,
		loadout.Legs,
		loadout.Accessory,
	}
	for _, item := range items {
		if item == nil {
			continue
		}
		attackPower += item.AttackPower
		armorPen += item.ArmorPenPercent
		critRate += item.CritRate
		critDmgMult += item.CritDamageMultiplier
	}
	return
}

func deriveCombatStats(stats CombatStats) CombatStats {
	stats.EffectiveIncrement = max(1, stats.AttackPower)
	stats.NormalDamage = stats.EffectiveIncrement

	if stats.CriticalCount <= 1 {
		stats.CriticalCount = 1
	}

	if stats.CritDamageMultiplier < 1.0 {
		stats.CritDamageMultiplier = 1.0
	}

	countBasedCriticalDamage := max(stats.NormalDamage+stats.CriticalCount-1, stats.NormalDamage)
	multiplierBasedCriticalDamage := int64(float64(stats.NormalDamage) * stats.CritDamageMultiplier)
	if multiplierBasedCriticalDamage < stats.NormalDamage {
		multiplierBasedCriticalDamage = stats.NormalDamage
	}
	stats.CriticalDamage = max(countBasedCriticalDamage, multiplierBasedCriticalDamage)

	return stats
}

func hasCriticalBonus(stats CombatStats) bool {
	return stats.CriticalCount > 1 || stats.CritDamageMultiplier > 1.0
}

// CalcBossPartDamage 计算对 Boss 部位的伤害（新减法公式）。
//
//	partType: 部位类型
//	partArmor: 部位护甲值
//	alivePartCount: 存活的部位数量（围剿技能用）
func CalcBossPartDamage(stats CombatStats, partType PartType, partArmor int64, alivePartCount int, bossCurrentHP int64, bossMaxHP int64) CombatStats {
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

	// 增伤乘区 = 1 + 全局增伤 + 部位增伤 + 每存活部位增伤
	amplifyBonus := stats.AllDamageAmplify
	switch partType {
	case PartTypeSoft:
		amplifyBonus += stats.PartTypeDamageSoft
	case PartTypeHeavy:
		amplifyBonus += stats.PartTypeDamageHeavy
	case PartTypeWeak:
		amplifyBonus += stats.PartTypeDamageWeak
	}
	amplifyBonus += float64(max(0, alivePartCount)) * max(0, stats.PerPartDamagePercent)

	// 低血斩杀增伤
	if stats.LowHpMultiplier > 1 && stats.LowHpThreshold > 0 {
		hpRatio := 1.0
		if bossMaxHP > 0 {
			hpRatio = float64(max(0, bossCurrentHP)) / float64(max(1, bossMaxHP))
		}
		if hpRatio <= stats.LowHpThreshold {
			amplifyBonus += (stats.LowHpMultiplier - 1)
		}
	}

	amplify := 1.0 + amplifyBonus

	// 暴击乘区（这里只计算“命中暴击时应造成多少伤害”，不在这里做第二次暴击判定）
	critMult := max(1.0, stats.CritDamageMultiplier)

	// 最终伤害
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
		PartTypeDamageSoft:    stats.PartTypeDamageSoft,
		PartTypeDamageHeavy:   stats.PartTypeDamageHeavy,
		PartTypeDamageWeak:    stats.PartTypeDamageWeak,
		PerPartDamagePercent:  stats.PerPartDamagePercent,
		LowHpMultiplier:       stats.LowHpMultiplier,
		LowHpThreshold:        stats.LowHpThreshold,
	}
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
		ID:                 id,
		TemplateID:         strings.TrimSpace(values["template_id"]),
		Name:               name,
		Status:             strings.TrimSpace(values["status"]),
		MaxHP:              int64FromString(values["max_hp"]),
		CurrentHP:          int64FromString(values["current_hp"]),
		GoldOnKill:         int64FromString(values["gold_on_kill"]),
		StoneOnKill:        int64FromString(values["stone_on_kill"]),
		TalentPointsOnKill: int64FromString(values["talent_points_on_kill"]),
		Parts:              parts,
		StartedAt:          int64FromString(values["started_at"]),
		DefeatedAt:         int64FromString(values["defeated_at"]),
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

	stats := &BossUserStats{
		Nickname: nickname,
		Damage:   int64(score),
	}

	rank, err := s.client.ZRevRank(ctx, s.bossDamageKey(bossID), nickname).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			return nil, err
		}
	} else {
		stats.Rank = int(rank) + 1
	}

	return stats, nil
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
		ItemID:               itemID,
		Name:                 firstNonEmpty(strings.TrimSpace(values["name"]), itemID),
		Slot:                 normalizeEquipmentSlot(values["slot"]),
		Rarity:               normalizeEquipmentRarity(values["rarity"]),
		ImagePath:            strings.TrimSpace(values["image_path"]),
		ImageAlt:             strings.TrimSpace(values["image_alt"]),
		AttackPower:          int64FromString(values["attack_power"]),
		ArmorPenPercent:      float64FromString(values["armor_pen_percent"]),
		CritRate:             float64FromString(values["crit_rate"]),
		CritDamageMultiplier: float64FromString(values["crit_damage_multiplier"]),
		PartTypeDamageSoft:   float64FromString(values["part_type_damage_soft"]),
		PartTypeDamageHeavy:  float64FromString(values["part_type_damage_heavy"]),
		PartTypeDamageWeak:   float64FromString(values["part_type_damage_weak"]),
		TalentAffinity:       strings.TrimSpace(values["talent_affinity"]),
	}, nil
}

func (s *Store) itemInstancesByIDForNickname(ctx context.Context, nickname string) (map[string]ItemInstance, error) {
	instanceIDs, err := s.client.SMembers(ctx, s.playerInstancesKey(nickname)).Result()
	if err != nil {
		if isRedisWrongTypeError(err) {
			return map[string]ItemInstance{}, nil
		}
		return nil, err
	}
	if len(instanceIDs) == 0 {
		return map[string]ItemInstance{}, nil
	}

	instances := make(map[string]ItemInstance, len(instanceIDs))
	for _, instanceID := range instanceIDs {
		instanceID = strings.TrimSpace(instanceID)
		if instanceID == "" {
			continue
		}
		values, err := s.client.HGetAll(ctx, s.equipmentInstanceKey(instanceID)).Result()
		if err != nil {
			return nil, err
		}
		if len(values) == 0 {
			continue
		}
		itemID := strings.TrimSpace(values["item_id"])
		if itemID == "" {
			continue
		}
		instance := ItemInstance{
			InstanceID:   instanceID,
			ItemID:       itemID,
			EnhanceLevel: int(int64FromString(values["enhance_level"])),
			SpentStones:  int64FromString(values["spent_stones"]),
			Bound:        int64FromString(values["bound"]) > 0,
			Locked:       int64FromString(values["locked"]) > 0,
			CreatedAt:    int64FromString(values["created_at"]),
		}
		instances[instanceID] = instance
	}

	return instances, nil
}

func (s *Store) getOwnedInstance(ctx context.Context, nickname string, ref string) (*ItemInstance, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, ErrEquipmentNotFound
	}

	owned, err := s.client.SIsMember(ctx, s.playerInstancesKey(nickname), ref).Result()
	if err != nil {
		if isRedisWrongTypeError(err) {
			return nil, ErrEquipmentNotOwned
		}
		return nil, err
	}
	if !owned {
		return nil, ErrEquipmentNotOwned
	}

	values, err := s.client.HGetAll(ctx, s.equipmentInstanceKey(ref)).Result()
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, ErrEquipmentNotOwned
	}
	itemID := strings.TrimSpace(values["item_id"])
	if itemID == "" {
		return nil, ErrEquipmentNotOwned
	}

	instance := &ItemInstance{
		InstanceID:   ref,
		ItemID:       itemID,
		EnhanceLevel: int(int64FromString(values["enhance_level"])),
		SpentStones:  int64FromString(values["spent_stones"]),
		Bound:        int64FromString(values["bound"]) > 0,
		Locked:       int64FromString(values["locked"]) > 0,
		CreatedAt:    int64FromString(values["created_at"]),
	}
	return instance, nil
}

func (s *Store) loadoutForNickname(ctx context.Context, nickname string) (Loadout, map[string]string, error) {
	values, err := s.client.HGetAll(ctx, s.loadoutKey(nickname)).Result()
	if err != nil {
		return Loadout{}, nil, err
	}
	instances, err := s.itemInstancesByIDForNickname(ctx, nickname)
	if err != nil {
		return Loadout{}, nil, err
	}

	loadout := Loadout{}
	equipped := make(map[string]string, len(values))
	for slot, equippedRef := range values {
		slot = normalizeEquipmentSlot(slot)
		equippedRef = strings.TrimSpace(equippedRef)
		if equippedRef == "" || slot == "" {
			continue
		}

		instance, ok := instances[equippedRef]
		if !ok {
			continue
		}
		definition, defErr := s.getEquipmentDefinition(ctx, instance.ItemID)
		if defErr != nil {
			continue
		}
		item := buildInventoryItem(definition, 1, true, instance.EnhanceLevel, instance.InstanceID, instance.Bound, instance.Locked)
		equipped[instance.InstanceID] = slot
		switch slot {
		case "weapon":
			loadout.Weapon = &item
		case "helmet":
			loadout.Helmet = &item
		case "chest":
			loadout.Chest = &item
		case "gloves":
			loadout.Gloves = &item
		case "legs":
			loadout.Legs = &item
		case "accessory":
			loadout.Accessory = &item
		}
	}

	return loadout, equipped, nil
}

func (s *Store) inventoryForNickname(ctx context.Context, nickname string, equipped map[string]string) ([]InventoryItem, error) {
	instances, err := s.itemInstancesByIDForNickname(ctx, nickname)
	if err != nil {
		return nil, err
	}

	items := make([]InventoryItem, 0, len(instances))
	for _, instance := range instances {
		definition, err := s.getEquipmentDefinition(ctx, instance.ItemID)
		if err != nil {
			items = append(items, InventoryItem{
				ItemID:       instance.ItemID,
				InstanceID:   instance.InstanceID,
				Name:         instance.ItemID,
				Quantity:     1,
				Equipped:     equipped[instance.InstanceID] != "",
				EnhanceLevel: instance.EnhanceLevel,
				Bound:        instance.Bound,
				Locked:       instance.Locked,
			})
			continue
		}
		items = append(items, buildInventoryItem(definition, 1, equipped[instance.InstanceID] != "", instance.EnhanceLevel, instance.InstanceID, instance.Bound, instance.Locked))
	}

	if len(items) == 0 {
		return []InventoryItem{}, nil
	}

	slices.SortFunc(items, func(left, right InventoryItem) int {
		if left.Slot == right.Slot {
			return strings.Compare(left.Name, right.Name)
		}
		return strings.Compare(left.Slot, right.Slot)
	})

	return items, nil
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
	for _, entry := range entries {
		itemID, ok := entry.Member.(string)
		if !ok || strings.TrimSpace(itemID) == "" {
			continue
		}

		dropRatePercent := clampFloat(entry.Score, 0, 100)

		definition, defErr := s.getEquipmentDefinition(ctx, itemID)
		if defErr != nil {
			loot = append(loot, BossLootEntry{
				ItemID:          itemID,
				DropRatePercent: dropRatePercent,
			})
			continue
		}

		loot = append(loot, BossLootEntry{
			ItemID:               itemID,
			ItemName:             definition.Name,
			Slot:                 definition.Slot,
			Rarity:               normalizeEquipmentRarity(definition.Rarity),
			ImagePath:            definition.ImagePath,
			ImageAlt:             definition.ImageAlt,
			DropRatePercent:      dropRatePercent,
			AttackPower:          definition.AttackPower,
			ArmorPenPercent:      definition.ArmorPenPercent,
			CritRate:             definition.CritRate,
			CritDamageMultiplier: definition.CritDamageMultiplier,
			PartTypeDamageSoft:   definition.PartTypeDamageSoft,
			PartTypeDamageHeavy:  definition.PartTypeDamageHeavy,
			PartTypeDamageWeak:   definition.PartTypeDamageWeak,
			TalentAffinity:       definition.TalentAffinity,
		})
	}

	return loot, nil
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

const dropRateRollLimit = 10000

func normalizeLootDropRate(item BossLootEntry) float64 {
	if item.DropRatePercent > 0 {
		return clampFloat(item.DropRatePercent, 0, 100)
	}
	return clampFloat(float64(item.Weight), 0, 100)
}

func dropRateThreshold(dropRatePercent float64) int {
	return int(math.Round(clampFloat(dropRatePercent, 0, 100) * 100))
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

func (s *Store) playerInstancesKey(nickname string) string {
	return s.playerInstancesPrefix + nickname
}

func (s *Store) equipmentInstanceKey(instanceID string) string {
	return s.equipmentInstancePrefix + strings.TrimSpace(instanceID)
}

func (s *Store) newEquipmentInstanceID(ctx context.Context) (string, error) {
	seq, err := s.client.Incr(ctx, s.equipmentInstanceSeqKey).Result()
	if err != nil {
		return "", err
	}
	return "inst-" + strconv.FormatInt(seq, 10), nil
}

func (s *Store) resourceKey(nickname string) string {
	return s.namespace + "resource:" + nickname
}

type playerResources struct {
	Gold         int64
	Stones       int64
	TalentPoints int64
}

func (s *Store) resourcesForNickname(ctx context.Context, nickname string) (playerResources, error) {
	resourceKey := s.resourceKey(nickname)
	values, err := s.client.HMGet(ctx, resourceKey, "gold", "stones", "talent_points").Result()
	if err != nil {
		return playerResources{}, err
	}

	return playerResources{
		Gold:         int64Value(values, 0),
		Stones:       int64Value(values, 1),
		TalentPoints: int64Value(values, 2),
	}, nil
}

func (s *Store) equipmentSpentKey(nickname string) string {
	return s.equipmentSpentPrefix + nickname
}

func (s *Store) equipmentEnhanceKey(nickname string) string {
	return s.equipmentEnhancePrefix + nickname
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

// RealtimeEventChannel 返回当前命名空间对应的 Redis 实时事件通道名。
func RealtimeEventChannel(namespace string) string {
	return namespace + "events"
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

func buildInventoryItem(definition EquipmentDefinition, quantity int64, equipped bool, enhanceLevel int, instanceID string, bound bool, locked bool) InventoryItem {
	enhanceLevel = maxInt(0, enhanceLevel)
	multValue := math.Pow(1.12, float64(enhanceLevel))
	multPercent := math.Pow(1.08, float64(enhanceLevel))

	name := displayItemName(definition.Name, enhanceLevel)
	attackPower := int64(math.Round(float64(definition.AttackPower) * multValue))
	armorPenPercent := definition.ArmorPenPercent * multPercent
	critRate := definition.CritRate * multPercent
	critDamageMultiplier := definition.CritDamageMultiplier * multPercent
	partTypeDamageSoft := definition.PartTypeDamageSoft * multPercent
	partTypeDamageHeavy := definition.PartTypeDamageHeavy * multPercent
	partTypeDamageWeak := definition.PartTypeDamageWeak * multPercent

	return InventoryItem{
		ItemID:               definition.ItemID,
		InstanceID:           strings.TrimSpace(instanceID),
		Name:                 name,
		Slot:                 normalizeEquipmentSlot(definition.Slot),
		Rarity:               normalizeEquipmentRarity(definition.Rarity),
		ImagePath:            definition.ImagePath,
		ImageAlt:             definition.ImageAlt,
		Quantity:             quantity,
		Equipped:             equipped,
		EnhanceLevel:         enhanceLevel,
		Bound:                bound,
		Locked:               locked,
		AttackPower:          attackPower,
		ArmorPenPercent:      armorPenPercent,
		CritRate:             critRate,
		CritDamageMultiplier: critDamageMultiplier,
		PartTypeDamageSoft:   partTypeDamageSoft,
		PartTypeDamageHeavy:  partTypeDamageHeavy,
		PartTypeDamageWeak:   partTypeDamageWeak,
	}
}

// 获取装备的强化金币消耗。
func enhanceGoldCost(currentLevel int) int64 {
	level := maxInt(0, currentLevel)
	return int64(math.Ceil(500 * math.Pow(1.5, float64(level))))
}

// 获取装备的强化石消耗。
func enhanceStoneCost(currentLevel int) int64 {
	level := maxInt(0, currentLevel)
	// 公式：3 * 1.5^level，然后向上取整
	return int64(math.Ceil(3 * math.Pow(1.5, float64(level))))
}

// salvageBaseReward 返回装备按稀有度分解得到的基础金币与强化石。
func salvageBaseReward(rarity string) (int64, int64) {
	switch strings.TrimSpace(rarity) {
	case "神话":
		return 5000, 20
	}

	switch normalizeEquipmentRarity(rarity) {
	case "至臻":
		return 10000, 50
	case "传说":
		return 2000, 8
	case "史诗":
		return 1000, 3
	case "稀有":
		return 500, 1
	case "优秀":
		return 300, 1
	case "普通":
		fallthrough
	default:
		return 200, 0
	}
}

func maxEnhanceLevel(rarity string) int {
	// 当前版本统一基础上限 + 稀有度额外上限。
	return 5 + RarityStatsForRarity(rarity).EnhanceCapExtra
}

func maxInt(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

func normalizeEquipmentSlot(slot string) string {
	switch strings.TrimSpace(slot) {
	case "weapon", "武器":
		return "weapon"
	case "helmet", "头盔":
		return "helmet"
	case "chest", "armor", "胸甲", "护甲":
		return "chest"
	case "gloves", "手套":
		return "gloves"
	case "legs", "腿甲":
		return "legs"
	case "accessory", "饰品":
		return "accessory"
	default:
		return strings.TrimSpace(slot)
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

// hasTalent checks if a learned set contains a talent ID.
func hasTalent(learned map[string]struct{}, id string) bool {
	_, ok := learned[id]
	return ok
}

// randomMarkIndices randomly selects count unique indices from [0, n).
func randomMarkIndices(n, count int, roll func(int) int) []int {
	if n <= 0 || count <= 0 {
		return nil
	}
	if count > n {
		count = n
	}
	result := make([]int, 0, count)
	seen := make(map[int]struct{}, count)
	for len(result) < count {
		idx := roll(n)
		if _, ok := seen[idx]; ok {
			continue
		}
		seen[idx] = struct{}{}
		result = append(result, idx)
	}
	return result
}
