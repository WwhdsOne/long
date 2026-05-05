package core

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"sync"
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
var ErrTalentTierLocked = errors.New("talent tier locked")
var ErrTalentNotFound = errors.New("talent not found")
var ErrInvalidTalentTree = errors.New("invalid talent tree")
var ErrTalentPointsInsufficient = errors.New("talent points insufficient")
var ErrTalentInvalidCost = errors.New("talent invalid cost")
var ErrTalentInvalidLevel = errors.New("invalid level")
var ErrTalentMaxLevel = errors.New("already at max level")
var ErrTaskNotFound = errors.New("task not found")
var ErrTaskNotClaimable = errors.New("task not claimable")
var ErrTaskAlreadyClaimed = errors.New("task already claimed")
var ErrTaskImmutable = errors.New("task immutable after activation")
var ErrShopItemNotFound = errors.New("shop item not found")
var ErrShopItemNotPurchasable = errors.New("shop item not purchasable")
var ErrShopItemAlreadyOwned = errors.New("shop item already owned")
var ErrShopItemNotOwned = errors.New("shop item not owned")
var ErrShopInsufficientGold = errors.New("shop insufficient gold")
var ErrShopUnsupportedItemType = errors.New("shop unsupported item type")

const (
	bossStatusActive   = "active"
	bossStatusDefeated = "defeated"

	bossPartClickSlugPrefix     = "boss-part:"
	equipmentDefinitionCacheTTL = 30 * time.Second
)

var loadoutSlots = []string{"weapon", "helmet", "chest", "gloves", "legs", "accessory"}

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
	RoomID             string     `json:"roomId,omitempty"`
	QueueID            string     `json:"queueId,omitempty"`
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

// EquipmentDraftFailureLog 记录装备草稿生成失败上下文，便于回放和调参。
type EquipmentDraftFailureLog struct {
	Prompt       string              `json:"prompt"`
	Draft        EquipmentDefinition `json:"draft"`
	ErrorMessage string              `json:"errorMessage"`
	RawResponse  string              `json:"rawResponse"`
	CreatedAt    int64               `json:"createdAt"`
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
	RoomID             string          `json:"roomId,omitempty"`
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
	RoomID              string                 `json:"roomId,omitempty"`
	Boss                *Boss                  `json:"boss,omitempty"`
	BossLeaderboard     []BossLeaderboardEntry `json:"bossLeaderboard"`
	AnnouncementVersion string                 `json:"announcementVersion,omitempty"`
}

// UserState 个人实时状态，只推送给对应昵称的连接
type UserState struct {
	UserStats                          *UserStats           `json:"userStats,omitempty"`
	MyBossStats                        *BossUserStats       `json:"myBossStats,omitempty"`
	MyBossKills                        int64                `json:"myBossKills"`
	TotalBossKills                     int64                `json:"totalBossKills"`
	RoomID                             string               `json:"roomId,omitempty"`
	Inventory                          []InventoryItem      `json:"inventory"`
	Loadout                            Loadout              `json:"loadout"`
	CombatStats                        CombatStats          `json:"combatStats"`
	Gold                               int64                `json:"gold"`
	Stones                             int64                `json:"stones"`
	TalentPoints                       int64                `json:"talentPoints"`
	RecentRewards                      []Reward             `json:"recentRewards,omitempty"`
	Tasks                              []PlayerTask         `json:"tasks,omitempty"`
	TalentEvents                       []TalentTriggerEvent `json:"talentEvents,omitempty"`
	TalentCombatState                  *TalentCombatState   `json:"talentCombatState,omitempty"`
	EquippedBattleClickSkinID          string               `json:"equippedBattleClickSkinId"`
	EquippedBattleClickCursorImagePath string               `json:"equippedBattleClickCursorImagePath"`
}

// State 完整状态，包含个人统计与玩法状态
type State struct {
	TotalVotes                         int64                  `json:"totalVotes"`
	Leaderboard                        []LeaderboardEntry     `json:"leaderboard"`
	UserStats                          *UserStats             `json:"userStats,omitempty"`
	RoomID                             string                 `json:"roomId,omitempty"`
	Boss                               *Boss                  `json:"boss,omitempty"`
	BossLeaderboard                    []BossLeaderboardEntry `json:"bossLeaderboard"`
	BossLoot                           []BossLootEntry        `json:"bossLoot,omitempty"`
	AnnouncementVersion                string                 `json:"announcementVersion,omitempty"`
	LatestAnnouncement                 *Announcement          `json:"latestAnnouncement,omitempty"`
	MyBossStats                        *BossUserStats         `json:"myBossStats,omitempty"`
	MyBossKills                        int64                  `json:"myBossKills"`
	TotalBossKills                     int64                  `json:"totalBossKills"`
	Inventory                          []InventoryItem        `json:"inventory"`
	Loadout                            Loadout                `json:"loadout"`
	CombatStats                        CombatStats            `json:"combatStats"`
	Gold                               int64                  `json:"gold"`
	Stones                             int64                  `json:"stones"`
	TalentPoints                       int64                  `json:"talentPoints"`
	RecentRewards                      []Reward               `json:"recentRewards,omitempty"`
	EquippedBattleClickSkinID          string                 `json:"equippedBattleClickSkinId"`
	EquippedBattleClickCursorImagePath string                 `json:"equippedBattleClickCursorImagePath"`
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
	Delta                int64                  `json:"delta"`
	RoomID               string                 `json:"roomId,omitempty"`
	BossDamage           int64                  `json:"bossDamage,omitempty"`
	MyBossDamage         int64                  `json:"myBossDamage,omitempty"`
	BossLeaderboardCount int                    `json:"bossLeaderboardCount,omitempty"`
	DamageType           string                 `json:"damageType,omitempty"`
	Critical             bool                   `json:"critical"`
	UserStats            UserStats              `json:"userStats"`
	Boss                 *Boss                  `json:"boss,omitempty"`
	BossLeaderboard      []BossLeaderboardEntry `json:"bossLeaderboard,omitempty"`
	MyBossStats          *BossUserStats         `json:"myBossStats,omitempty"`
	RecentRewards        []Reward               `json:"recentRewards,omitempty"`
	TalentEvents         []TalentTriggerEvent   `json:"talentEvents,omitempty"`
	TalentCombatState    *TalentCombatState     `json:"talentCombatState,omitempty"`
	PartStateDeltas      []BossPartStateDelta   `json:"partStateDeltas,omitempty"`
	BroadcastUserAll     bool                   `json:"-"`
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
	RoomID           string          `json:"roomId,omitempty"`
	QueueID          string          `json:"queueId,omitempty"`
	BroadcastUserAll bool            `json:"broadcastUserAll,omitempty"`
	Timestamp        int64           `json:"timestamp"`
}

// StoreOptions 暴击机制配置
type StoreOptions struct {
	CriticalChancePercent int
	Room                  RoomConfig
	BossHistoryArchiver   interface{ Enqueue(BossHistoryEntry) bool }
	BossHistoryStore      interface {
		SaveBossHistory(context.Context, BossHistoryEntry) error
		ListAdminBossHistoryPage(context.Context, int64, int64) (AdminBossHistoryPage, error)
		ListBossHistory(context.Context) ([]BossHistoryEntry, error)
	}
	MessageStore interface {
		CreateMessage(context.Context, string, string) (*Message, error)
		ListMessages(context.Context, string, int64) (MessagePage, error)
		DeleteMessage(context.Context, string) error
	}
	TaskDefinitionStore interface {
		ListActiveTaskDefinitions(context.Context, int64) ([]TaskDefinition, error)
		ListTaskDefinitions(context.Context) ([]TaskDefinition, error)
		GetTaskDefinition(context.Context, string) (*TaskDefinition, error)
		UpsertTaskDefinition(context.Context, TaskDefinition) error
	}
	TaskClaimLogStore interface {
		HasTaskClaimed(context.Context, string, string, string) (bool, error)
		WriteTaskClaimLog(context.Context, TaskClaimLog) error
		ListTaskClaimLogs(context.Context, string, string) ([]TaskClaimLog, error)
		HasTaskClaimLog(context.Context, string, string) (bool, error)
	}
	TaskCycleArchiveStore interface {
		UpsertTaskCycleArchive(context.Context, TaskCycleArchive) error
		UpsertTaskCyclePlayerResults(context.Context, []TaskCyclePlayerResult) error
		ListTaskCycleArchives(context.Context, string) ([]TaskCycleArchive, error)
		GetTaskCycleResults(context.Context, string, string) (TaskCycleResultsView, error)
	}
	ShopCatalogStore interface {
		ListActiveShopItems(context.Context) ([]ShopItem, error)
		ListShopItems(context.Context) ([]ShopItem, error)
		GetShopItem(context.Context, string) (*ShopItem, error)
		UpsertShopItem(context.Context, ShopItem) error
		DeleteShopItem(context.Context, string) error
	}
	ShopPurchaseLogStore interface {
		WriteShopPurchaseLog(context.Context, ShopPurchaseLog) error
	}
}

// Store Redis 投票存储，管理按钮列表、点击计数、Boss 与装备状态
type Store struct {
	client                        redis.UniversalClient
	namespace                     string
	equipmentIndexKey             string
	playerIndexKey                string
	userPrefix                    string
	leaderboardKey                string
	totalVotesKey                 string
	bossCurrentKey                string
	bossHistoryKey                string
	bossHistoryPrefix             string
	bossTemplateIndexKey          string
	bossTemplatePrefix            string
	bossCycleKey                  string
	bossInstanceSeqKey            string
	playerRoomPrefix              string
	playerRoomCooldownPrefix      string
	announcementSeqKey            string
	announcementKey               string
	announcementPrefix            string
	messageSeqKey                 string
	messageKey                    string
	messagePrefix                 string
	equipmentDefPrefix            string
	equipmentInstancePrefix       string
	inventoryPrefix               string
	playerInstancesPrefix         string
	loadoutPrefix                 string
	lastRewardPrefix              string
	equipmentInstanceSeqKey       string
	equipmentSpentPrefix          string
	equipmentEnhancePrefix        string
	ownedBattleClickSkinsPrefix   string
	equippedBattleClickSkinPrefix string
	taskProgressPrefix            string
	taskParticipantsPrefix        string
	critical                      StoreOptions
	roomConfig                    RoomConfig
	luaRunner                     luaScriptRunner
	clickCountScript              *cachedLuaScript
	bossClickScript               *cachedLuaScript
	roll                          func(int) int
	now                           func() time.Time
	validator                     interface{ Validate(string) error }
	bossHistoryArchiver           interface{ Enqueue(BossHistoryEntry) bool }
	bossHistoryStore              interface {
		SaveBossHistory(context.Context, BossHistoryEntry) error
		ListAdminBossHistoryPage(context.Context, int64, int64) (AdminBossHistoryPage, error)
		ListBossHistory(context.Context) ([]BossHistoryEntry, error)
	}
	messageStore interface {
		CreateMessage(context.Context, string, string) (*Message, error)
		ListMessages(context.Context, string, int64) (MessagePage, error)
		DeleteMessage(context.Context, string) error
	}
	taskDefinitionStore interface {
		ListActiveTaskDefinitions(context.Context, int64) ([]TaskDefinition, error)
		ListTaskDefinitions(context.Context) ([]TaskDefinition, error)
		GetTaskDefinition(context.Context, string) (*TaskDefinition, error)
		UpsertTaskDefinition(context.Context, TaskDefinition) error
	}
	taskClaimLogStore interface {
		HasTaskClaimed(context.Context, string, string, string) (bool, error)
		WriteTaskClaimLog(context.Context, TaskClaimLog) error
		ListTaskClaimLogs(context.Context, string, string) ([]TaskClaimLog, error)
		HasTaskClaimLog(context.Context, string, string) (bool, error)
	}
	taskCycleArchiveStore interface {
		UpsertTaskCycleArchive(context.Context, TaskCycleArchive) error
		UpsertTaskCyclePlayerResults(context.Context, []TaskCyclePlayerResult) error
		ListTaskCycleArchives(context.Context, string) ([]TaskCycleArchive, error)
		GetTaskCycleResults(context.Context, string, string) (TaskCycleResultsView, error)
	}
	shopCatalogStore interface {
		ListActiveShopItems(context.Context) ([]ShopItem, error)
		ListShopItems(context.Context) ([]ShopItem, error)
		GetShopItem(context.Context, string) (*ShopItem, error)
		UpsertShopItem(context.Context, ShopItem) error
		DeleteShopItem(context.Context, string) error
	}
	shopPurchaseLogStore interface {
		WriteShopPurchaseLog(context.Context, ShopPurchaseLog) error
	}

	combatStatsCache   map[string]CombatStats
	combatStatsCacheMu sync.RWMutex

	equipmentDefinitionCache   map[string]cachedEquipmentDefinition
	equipmentDefinitionCacheMu sync.RWMutex

	compiledTalentCache   map[string]*CompiledTalentSet
	compiledTalentCacheMu sync.RWMutex
}

type cachedEquipmentDefinition struct {
	definition EquipmentDefinition
	expiresAt  time.Time
}

type equipmentDefinitionRequestCache struct {
	definitions map[string]EquipmentDefinition
	missing     map[string]struct{}
}

// NewStore 创建 Redis 投票存储实例
func NewStore(client redis.UniversalClient, namespace string, options StoreOptions, validator interface{ Validate(string) error }) *Store {
	luaCache := newLuaScriptCache()
	roomConfig := normalizeRoomConfig(options.Room)

	return &Store{
		client:                        client,
		namespace:                     namespace,
		equipmentIndexKey:             namespace + "equipment:index",
		playerIndexKey:                namespace + "players:index",
		userPrefix:                    namespace + "user:",
		leaderboardKey:                namespace + "leaderboard",
		totalVotesKey:                 namespace + "total:votes",
		bossCurrentKey:                namespace + "boss:current",
		bossHistoryKey:                namespace + "boss:history",
		bossHistoryPrefix:             namespace + "boss:history:",
		bossTemplateIndexKey:          namespace + "boss:pool:index",
		bossTemplatePrefix:            namespace + "boss:pool:",
		bossCycleKey:                  namespace + "boss:cycle",
		bossInstanceSeqKey:            namespace + "boss:instance:seq",
		playerRoomPrefix:              namespace + "player:room:",
		playerRoomCooldownPrefix:      namespace + "player:room:cd:",
		announcementSeqKey:            namespace + "announcement:seq",
		announcementKey:               namespace + "announcements",
		announcementPrefix:            namespace + "announcement:",
		messageSeqKey:                 namespace + "message:seq",
		messageKey:                    namespace + "messages",
		messagePrefix:                 namespace + "message:",
		equipmentDefPrefix:            namespace + "equip:def:",
		equipmentInstancePrefix:       namespace + "instance:",
		inventoryPrefix:               namespace + "user-inventory:",
		playerInstancesPrefix:         namespace + "player-instances:",
		loadoutPrefix:                 namespace + "user-loadout:",
		lastRewardPrefix:              namespace + "user-last-reward:",
		equipmentInstanceSeqKey:       namespace + "instance:seq",
		equipmentSpentPrefix:          namespace + "user-equipment-spent:",
		equipmentEnhancePrefix:        namespace + "user-equipment-enhance:",
		ownedBattleClickSkinsPrefix:   namespace + "player:owned:click-skins:",
		equippedBattleClickSkinPrefix: namespace + "player:equipped:click-skin:",
		taskProgressPrefix:            namespace + "task:progress:",
		taskParticipantsPrefix:        namespace + "task:participants:",
		critical:                      options,
		roomConfig:                    roomConfig,
		luaRunner: redisLuaRunner{
			client: client,
		},
		clickCountScript: newCachedLuaScript("click-count", clickCountLuaSource, luaCache),
		bossClickScript:  newCachedLuaScript("boss-click", bossClickLuaSource, luaCache),
		roll: func(limit int) int {
			return rand.IntN(limit)
		},
		now:                      time.Now,
		validator:                validator,
		bossHistoryArchiver:      options.BossHistoryArchiver,
		bossHistoryStore:         options.BossHistoryStore,
		messageStore:             options.MessageStore,
		taskDefinitionStore:      options.TaskDefinitionStore,
		taskClaimLogStore:        options.TaskClaimLogStore,
		taskCycleArchiveStore:    options.TaskCycleArchiveStore,
		shopCatalogStore:         options.ShopCatalogStore,
		shopPurchaseLogStore:     options.ShopPurchaseLogStore,
		combatStatsCache:         make(map[string]CombatStats),
		equipmentDefinitionCache: make(map[string]cachedEquipmentDefinition),
		compiledTalentCache:      make(map[string]*CompiledTalentSet),
	}
}

func (s *Store) cachedCombatStats(nickname string) (CombatStats, bool) {
	s.combatStatsCacheMu.RLock()
	defer s.combatStatsCacheMu.RUnlock()

	stats, ok := s.combatStatsCache[nickname]
	return stats, ok
}

func (s *Store) storeCombatStatsCache(nickname string, stats CombatStats) {
	s.combatStatsCacheMu.Lock()
	defer s.combatStatsCacheMu.Unlock()

	s.combatStatsCache[nickname] = stats
}

func (s *Store) invalidateCombatStatsCache(nickname string) {
	s.combatStatsCacheMu.Lock()
	defer s.combatStatsCacheMu.Unlock()

	delete(s.combatStatsCache, nickname)
}

func (s *Store) invalidateAllCombatStatsCaches() {
	s.combatStatsCacheMu.Lock()
	defer s.combatStatsCacheMu.Unlock()

	clear(s.combatStatsCache)
}

func (s *Store) cachedEquipmentDefinition(itemID string) (EquipmentDefinition, bool) {
	now := s.now()

	s.equipmentDefinitionCacheMu.RLock()
	entry, ok := s.equipmentDefinitionCache[itemID]
	s.equipmentDefinitionCacheMu.RUnlock()
	if !ok {
		return EquipmentDefinition{}, false
	}
	if !entry.expiresAt.After(now) {
		s.equipmentDefinitionCacheMu.Lock()
		delete(s.equipmentDefinitionCache, itemID)
		s.equipmentDefinitionCacheMu.Unlock()
		return EquipmentDefinition{}, false
	}
	return entry.definition, true
}

func (s *Store) storeEquipmentDefinitionCache(itemID string, definition EquipmentDefinition) {
	s.equipmentDefinitionCacheMu.Lock()
	defer s.equipmentDefinitionCacheMu.Unlock()

	s.equipmentDefinitionCache[itemID] = cachedEquipmentDefinition{
		definition: definition,
		expiresAt:  s.now().Add(equipmentDefinitionCacheTTL),
	}
}

func (s *Store) invalidateEquipmentDefinitionCache(itemID string) {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return
	}

	s.equipmentDefinitionCacheMu.Lock()
	defer s.equipmentDefinitionCacheMu.Unlock()

	delete(s.equipmentDefinitionCache, itemID)
}

func newEquipmentDefinitionRequestCache() *equipmentDefinitionRequestCache {
	return &equipmentDefinitionRequestCache{
		definitions: make(map[string]EquipmentDefinition),
		missing:     make(map[string]struct{}),
	}
}

func (s *Store) cachedCompiledTalentSet(nickname string) (*CompiledTalentSet, bool) {
	s.compiledTalentCacheMu.RLock()
	defer s.compiledTalentCacheMu.RUnlock()

	compiled, ok := s.compiledTalentCache[nickname]
	return compiled, ok
}

func (s *Store) storeCompiledTalentCache(nickname string, compiled *CompiledTalentSet) {
	s.compiledTalentCacheMu.Lock()
	defer s.compiledTalentCacheMu.Unlock()

	s.compiledTalentCache[nickname] = compiled
}

func (s *Store) invalidateCompiledTalentCache(nickname string) {
	s.compiledTalentCacheMu.Lock()
	defer s.compiledTalentCacheMu.Unlock()

	delete(s.compiledTalentCache, nickname)
}

func (s *Store) invalidatePlayerCombatCaches(nickname string) {
	s.invalidateCombatStatsCache(nickname)
	s.invalidateCompiledTalentCache(nickname)
}

// ValidateNickname checks whether the provided nickname is usable.
func (s *Store) ValidateNickname(_ context.Context, nickname string) error {
	_, err := s.validatedNickname(nickname)
	return err
}

// GetSnapshot 获取公共快照（公共排行榜 + Boss 状态）
func (s *Store) GetSnapshot(ctx context.Context) (Snapshot, error) {
	return s.GetSnapshotForRoom(ctx, s.defaultRoomID())
}

// GetSnapshotForNickname 获取玩家当前房间的公共快照。
func (s *Store) GetSnapshotForNickname(ctx context.Context, nickname string) (Snapshot, error) {
	roomID, err := s.ResolvePlayerRoom(ctx, nickname)
	if err != nil {
		return Snapshot{}, err
	}
	return s.GetSnapshotForRoom(ctx, roomID)
}

// GetSnapshotForRoom 获取指定房间的公共快照。
func (s *Store) GetSnapshotForRoom(ctx context.Context, roomID string) (Snapshot, error) {
	if isHallRoomID(roomID) {
		totalVotes, err := s.totalClickCount(ctx)
		if err != nil {
			return Snapshot{}, err
		}
		leaderboard, err := s.ListLeaderboard(ctx, 10)
		if err != nil {
			return Snapshot{}, err
		}
		announcementVersion, err := s.GetLatestAnnouncementVersion(ctx)
		if err != nil {
			return Snapshot{}, err
		}
		return Snapshot{
			TotalVotes:          totalVotes,
			Leaderboard:         leaderboard,
			RoomID:              hallRoomID,
			Boss:                nil,
			BossLeaderboard:     []BossLeaderboardEntry{},
			AnnouncementVersion: announcementVersion,
		}, nil
	}
	roomID = s.normalizeRoomID(roomID)
	totalVotes, err := s.totalClickCount(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	leaderboard, err := s.ListLeaderboard(ctx, 10)
	if err != nil {
		return Snapshot{}, err
	}

	boss, err := s.currentBossForRoom(ctx, roomID)
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
		RoomID:              roomID,
		Boss:                boss,
		BossLeaderboard:     bossLeaderboard,
		AnnouncementVersion: announcementVersion,
	}, nil
}

// GetState 获取完整状态（公共快照 + 个人统计）
func (s *Store) GetState(ctx context.Context, nickname string) (State, error) {
	snapshot, err := s.GetSnapshotForNickname(ctx, nickname)
	if err != nil {
		return State{}, err
	}

	userState, err := s.GetUserState(ctx, nickname)
	if err != nil {
		return State{}, err
	}

	return ComposeState(snapshot, userState), nil
}

// GetUserState 获取仅与指定用户相关的完整个人态。
func (s *Store) GetUserState(ctx context.Context, nickname string) (UserState, error) {
	return s.getUserState(ctx, nickname, true)
}

// GetRealtimeUserState 获取增量实时推送所需的轻量个人态。
func (s *Store) GetRealtimeUserState(ctx context.Context, nickname string) (UserState, error) {
	return s.getUserState(ctx, nickname, false)
}

func (s *Store) getUserState(ctx context.Context, nickname string, includeProfile bool) (UserState, error) {
	userState := UserState{
		Inventory:     []InventoryItem{},
		Loadout:       Loadout{},
		CombatStats:   s.baseCombatStats(),
		RecentRewards: []Reward{},
		Tasks:         []PlayerTask{},
	}
	userState.RoomID = hallRoomID
	totalBossKills, err := s.totalBossKills(ctx)
	if err != nil {
		return UserState{}, err
	}
	userState.TotalBossKills = totalBossKills

	trimmedNickname, hasNickname := normalizeNickname(nickname)
	if !hasNickname {
		return userState, nil
	}

	normalizedNickname, err := s.validatedNickname(trimmedNickname)
	if err != nil {
		return UserState{}, err
	}
	roomID, err := s.ResolvePlayerRoom(ctx, normalizedNickname)
	if err != nil {
		return UserState{}, err
	}
	userState.RoomID = roomID

	resources, err := s.resourcesForNickname(ctx, normalizedNickname)
	if err != nil {
		return UserState{}, err
	}
	userState.Gold = resources.Gold
	userState.Stones = resources.Stones
	userState.TalentPoints = resources.TalentPoints
	userState.MyBossKills = resources.BossKills
	if userState.MyBossKills <= 0 {
		fallbackKills, fallbackErr := s.historyBossKillsForNickname(ctx, normalizedNickname)
		if fallbackErr != nil {
			return UserState{}, fallbackErr
		}
		userState.MyBossKills = fallbackKills
	}

	userStats, err := s.GetUserStats(ctx, normalizedNickname)
	if err != nil {
		return UserState{}, err
	}
	userState.UserStats = &userStats

	if includeProfile {
		instances, err := s.itemInstancesByIDForNickname(ctx, normalizedNickname)
		if err != nil {
			return UserState{}, err
		}
		loadoutRefs, err := s.loadoutRefsForNickname(ctx, normalizedNickname)
		if err != nil {
			return UserState{}, err
		}
		definitionCache := newEquipmentDefinitionRequestCache()

		loadout, equipped := s.loadoutFromRefs(ctx, loadoutRefs, instances, definitionCache)
		userState.Loadout = loadout

		inventory := s.inventoryFromInstances(ctx, instances, equipped, definitionCache)
		userState.Inventory = inventory

		combatStats, err := s.combatStatsForNickname(ctx, normalizedNickname, loadout)
		if err != nil {
			return UserState{}, err
		}
		userState.CombatStats = combatStats
	}

	recentRewards, err := s.recentRewardsForNickname(ctx, normalizedNickname)
	if err != nil {
		return UserState{}, err
	}
	userState.RecentRewards = recentRewards

	if includeProfile {
		tasks, err := s.ListTasksForPlayer(ctx, normalizedNickname)
		if err != nil {
			return UserState{}, err
		}
		userState.Tasks = tasks
	}

	equippedBattleClickSkinID, equippedBattleClickCursorImagePath, err := s.equippedBattleClickSkinState(ctx, normalizedNickname)
	if err != nil {
		return UserState{}, err
	}
	userState.EquippedBattleClickSkinID = equippedBattleClickSkinID
	userState.EquippedBattleClickCursorImagePath = equippedBattleClickCursorImagePath

	combatRoomID := s.combatRoomID(roomID)
	boss, err := s.currentBossForRoom(ctx, combatRoomID)
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
		pendingTalentEvents, err := s.consumePendingTalentEvents(ctx, normalizedNickname, boss.ID)
		if err != nil {
			return UserState{}, err
		}
		userState.TalentEvents = pendingTalentEvents
	}

	return userState, nil
}

// GetPlayerResources 获取玩家资源快照，不消费 pending 天赋事件。
func (s *Store) GetPlayerResources(ctx context.Context, nickname string) (PlayerResources, error) {
	trimmedNickname, hasNickname := normalizeNickname(nickname)
	if !hasNickname {
		return PlayerResources{}, nil
	}

	normalizedNickname, err := s.validatedNickname(trimmedNickname)
	if err != nil {
		return PlayerResources{}, err
	}

	resources, err := s.resourcesForNickname(ctx, normalizedNickname)
	if err != nil {
		return PlayerResources{}, err
	}
	return PlayerResources{
		Gold:         resources.Gold,
		Stones:       resources.Stones,
		TalentPoints: resources.TalentPoints,
	}, nil
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
	s.invalidatePlayerCombatCaches(normalizedNickname)
	if err := s.recordTaskEvent(ctx, normalizedNickname, TaskEventEnhance, 1); err != nil {
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
	s.invalidatePlayerCombatCaches(normalizedNickname)

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
	s.invalidatePlayerCombatCaches(normalizedNickname)

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

// GetCurrentBossForRoom 返回指定房间当前 Boss。
func (s *Store) GetCurrentBossForRoom(ctx context.Context, roomID string) (*Boss, error) {
	return s.currentBossForRoom(ctx, roomID)
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

// ListLeaderboard 获取排行榜前 N 名。
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

// ListLeaderboardIncludingZeroClickPlayers 返回包含 0 点击玩家的排行榜窗口。
func (s *Store) ListLeaderboardIncludingZeroClickPlayers(ctx context.Context, offset int64, limit int64) ([]LeaderboardEntry, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 40
	}

	clickedLeaderboard, err := s.ListLeaderboard(ctx, 1000000)
	if err != nil {
		return nil, err
	}
	clickedSet := make(map[string]struct{}, len(clickedLeaderboard))
	allEntries := make([]LeaderboardEntry, 0, len(clickedLeaderboard))
	for _, entry := range clickedLeaderboard {
		if strings.TrimSpace(entry.Nickname) == "" {
			continue
		}
		clickedSet[entry.Nickname] = struct{}{}
		allEntries = append(allEntries, entry)
	}

	nicknames, err := s.listPlayerNicknames(ctx)
	if err != nil {
		return nil, err
	}
	rank := len(allEntries) + 1
	for _, nickname := range nicknames {
		normalized := strings.TrimSpace(nickname)
		if normalized == "" {
			continue
		}
		if _, ok := clickedSet[normalized]; ok {
			continue
		}
		allEntries = append(allEntries, LeaderboardEntry{
			Rank:       rank,
			Nickname:   normalized,
			ClickCount: 0,
		})
		rank++
	}

	if offset >= int64(len(allEntries)) {
		return []LeaderboardEntry{}, nil
	}
	end := min(int(offset+limit), len(allEntries))
	return append([]LeaderboardEntry(nil), allEntries[offset:end]...), nil
}

func (s *Store) totalClickCount(ctx context.Context) (int64, error) {
	totalVotes, err := s.client.Get(ctx, s.totalVotesKey).Int64()
	if err == nil {
		return totalVotes, nil
	}
	if err != nil && !errors.Is(err, redis.Nil) {
		return 0, err
	}

	scores, err := s.client.ZRangeWithScores(ctx, s.leaderboardKey, 0, -1).Result()
	if err != nil {
		return 0, err
	}
	total := int64(0)
	for _, score := range scores {
		total += int64(score.Score)
	}
	if setErr := s.client.Set(ctx, s.totalVotesKey, strconv.FormatInt(total, 10), 0).Err(); setErr != nil {
		return 0, setErr
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
		TotalVotes:                         snapshot.TotalVotes,
		Leaderboard:                        snapshot.Leaderboard,
		UserStats:                          userState.UserStats,
		RoomID:                             firstNonEmpty(userState.RoomID, snapshot.RoomID),
		Boss:                               snapshot.Boss,
		BossLeaderboard:                    snapshot.BossLeaderboard,
		AnnouncementVersion:                snapshot.AnnouncementVersion,
		MyBossStats:                        userState.MyBossStats,
		MyBossKills:                        userState.MyBossKills,
		TotalBossKills:                     userState.TotalBossKills,
		Inventory:                          userState.Inventory,
		Loadout:                            userState.Loadout,
		CombatStats:                        userState.CombatStats,
		Gold:                               userState.Gold,
		Stones:                             userState.Stones,
		TalentPoints:                       userState.TalentPoints,
		RecentRewards:                      userState.RecentRewards,
		EquippedBattleClickSkinID:          userState.EquippedBattleClickSkinID,
		EquippedBattleClickCursorImagePath: userState.EquippedBattleClickCursorImagePath,
	}
}

func (s *Store) applyClickCountOnly(ctx context.Context, nickname string, delta int64, critical bool) (ClickResult, error) {
	now := time.Now().Unix()
	reply, err := s.clickCountScript.Run(ctx, s.luaRunner,
		[]string{
			s.userPrefix + nickname,
			s.leaderboardKey,
			s.playerIndexKey,
			s.totalVotesKey,
		},
		delta,
		nickname,
		now,
	)
	if err != nil {
		return ClickResult{}, err
	}
	values, ok := reply.([]any)
	if !ok || len(values) < 2 {
		return ClickResult{}, fmt.Errorf("invalid click count script reply")
	}
	userCount := int64FromLuaValue(values[0])

	return ClickResult{
		Delta:    delta,
		Critical: critical,
		UserStats: UserStats{
			Nickname:   nickname,
			ClickCount: userCount,
		},
	}, nil
}

func (s *Store) AutoClickBossPart(ctx context.Context, _ string, nickname string) (ClickResult, error) {
	return s.AttackBossPartAFK(ctx, nickname)
}

// AttackBossPartAFK 执行一次挂机攻击，不增加点击数，伤害按攻击力*0.5 向下取整。
func (s *Store) AttackBossPartAFK(ctx context.Context, nickname string) (ClickResult, error) {
	roomID, err := s.ResolvePlayerRoom(ctx, nickname)
	if err != nil {
		return ClickResult{}, err
	}
	return s.AttackBossPartAFKInRoom(ctx, nickname, s.combatRoomID(roomID))
}

// AttackBossPartAFKInRoom 在指定房间执行一次挂机攻击。
func (s *Store) AttackBossPartAFKInRoom(ctx context.Context, nickname string, roomID string) (ClickResult, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return ClickResult{}, err
	}
	roomID = s.normalizeRoomID(roomID)
	boss, err := s.currentBossForRoom(ctx, roomID)
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

	damage := max(int64(math.Floor(float64(maxInt64(0, combatStats.AttackPower))*0.5)), 0)
	_, actualDamage, _ := applyBossPartDamageDelta(boss, part, damage)
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
	pipe.HSet(ctx, s.bossCurrentKeyForRoom(roomID), bossValues)
	if actualDamage > 0 {
		pipe.ZIncrBy(ctx, s.bossDamageKey(boss.ID), float64(actualDamage), normalizedNickname)
	}
	if _, execErr := pipe.Exec(ctx); execErr != nil {
		return ClickResult{}, nil
	}

	result := ClickResult{
		Delta:      0,
		RoomID:     roomID,
		Boss:       boss,
		BossDamage: actualDamage,
		DamageType: resolveBossDamageType(resolveBossDamageTypeInput{
			PartType:    part.Type,
			Critical:    false,
			BossDamage:  actualDamage,
			BossMaxHP:   boss.MaxHP,
			IsCollapsed: false,
			IsAfkAttack: true,
		}),
		UserStats: UserStats{
			Nickname: normalizedNickname,
		},
	}
	if actualDamage > 0 {
		if myBossDamage, summaryErr := s.client.ZScore(ctx, s.bossDamageKey(boss.ID), normalizedNickname).Result(); summaryErr == nil {
			result.MyBossDamage = int64(math.Round(myBossDamage))
		}
		if bossLeaderboardCount, summaryErr := s.client.ZCard(ctx, s.bossDamageKey(boss.ID)).Result(); summaryErr == nil {
			result.BossLeaderboardCount = int(bossLeaderboardCount)
		}
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

	roomID, err := s.ResolvePlayerRoom(ctx, nickname)
	if err != nil {
		return ClickResult{}, err
	}
	roomID = s.combatRoomID(roomID)
	boss, err := s.currentBossForRoom(ctx, roomID)
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
	result.RoomID = roomID
	result, err = s.applyBossPartDamage(ctx, boss, nickname, critical, result, targetIdx, comboCount)
	if err != nil {
		return ClickResult{}, err
	}
	if recordErr := s.recordTaskEvent(ctx, nickname, TaskEventClick, 1); recordErr != nil {
		return ClickResult{}, recordErr
	}
	return result, nil
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

	nowTime := s.now()
	now := nowTime.Unix()
	nowMs := nowTime.UnixMilli()

	aliveCount := 0
	for _, p := range boss.Parts {
		if p.Alive {
			aliveCount++
		}
	}

	compiledTalents, _ := s.compiledTalentSetForNickname(ctx, nickname)
	if compiledTalents == nil {
		compiledTalents = compileTalentSet(nil)
	}

	combatState, _ := s.GetTalentCombatState(ctx, nickname, boss.ID)
	if combatState == nil {
		combatState = NewTalentCombatState()
	}
	if combatState.SilverStormActive {
		if combatState.SilverStormEndsAt > 0 && now >= combatState.SilverStormEndsAt {
			combatState.SilverStormActive = false
			combatState.SilverStormRemaining = 0
			combatState.SilverStormEndsAt = 0
		} else if combatState.SilverStormEndsAt > 0 {
			combatState.SilverStormRemaining = int(maxInt64(0, combatState.SilverStormEndsAt-now))
		}
	}

	effectivePartType := part.Type
	partKey := TalentPartKey(part.X, part.Y)
	if endsAt, ok := combatState.SkinnerParts[partKey]; ok && now < endsAt {
		effectivePartType = PartTypeWeak
	}

	effectiveArmor := part.Armor
	inCollapse := false
	if slices.Contains(combatState.CollapseParts, targetIdx) {
		effectiveArmor = 0
		inCollapse = true
	}

	damageStats := CalcBossPartDamage(combatStats, effectivePartType, effectiveArmor, aliveCount, boss.CurrentHP, boss.MaxHP)
	partDamage := damageStats.NormalDamage
	partDamage = applyComboDamageAmplify(partDamage, comboCount)
	if critical {
		partDamage = damageStats.CriticalDamage

	}

	if inCollapse && compiledTalents.Armor.CollapseAmp > 1 {
		partDamage = int64(float64(partDamage) * compiledTalents.Armor.CollapseAmp)
	}

	hpRatio := float64(part.CurrentHP) / float64(maxInt64(1, boss.MaxHP))
	if compiledTalents.Has("crit_omen_kill") && combatState.OmenStacks > 0 {
		if hpRatio < compiledTalents.Crit.OmenKillThreshold {
			partDamage = int64(float64(partDamage) * (1.0 + float64(combatState.OmenStacks)*compiledTalents.Crit.OmenKillDmgPerOmen))
		}
	}

	if critical && compiledTalents.Has("crit_core") && combatState.OmenStacks > 0 {
		partDamage = int64(float64(partDamage) * (1.0 + float64(combatState.OmenStacks)*compiledTalents.Crit.OmenResonatePerOmen))
	}
	if critical && effectivePartType == PartTypeWeak && compiledTalents.Crit.WeakspotInsightMult > 1 {
		partDamage = int64(float64(partDamage) * compiledTalents.Crit.WeakspotInsightMult)
	}

	// 死兆收割被动：根据死兆层数档位提供增伤（不消耗层数）
	if compiledTalents.Has("crit_omen_reap") && len(compiledTalents.Crit.OmenReapThresholds) > 0 {
		reapMult := 1.0
		for i := len(compiledTalents.Crit.OmenReapThresholds) - 1; i >= 0; i-- {
			if combatState.OmenStacks >= compiledTalents.Crit.OmenReapThresholds[i] {
				reapMult = compiledTalents.Crit.OmenReapDamageMults[i]
				break
			}
		}
		if reapMult > 1.0 {
			partDamage = int64(float64(partDamage) * reapMult)
		}
	}

	partWasAlive := part.CurrentHP > 0
	omenGain := 0
	if compiledTalents.Has("crit_core") {
		if critical && effectivePartType == PartTypeWeak && partWasAlive {
			omenGain = compiledTalents.Crit.OmenPerWeakCrit
		}
	}
	beforeHP, actualDamage, _ := applyBossPartDamageDelta(boss, part, partDamage)

	// 死兆层数获取：弱点暴击 +2，普通暴击 +1，击碎部位 +5
	if compiledTalents.Has("crit_core") && omenGain > 0 {
		combatState.OmenStacks, _ = applyOmenStackDelta(combatState.OmenStacks, omenGain)
	}

	if critical && compiledTalents.Has("crit_skinner") && now >= combatState.SkinnerCooldownEndsAt {
		var candidates []BossPart
		for _, candidate := range boss.Parts {
			if !candidate.Alive || candidate.Type == PartTypeWeak {
				continue
			}
			candidateKey := TalentPartKey(candidate.X, candidate.Y)
			if endsAt, ok := combatState.SkinnerParts[candidateKey]; ok && endsAt > now {
				continue
			}
			candidates = append(candidates, candidate)
		}
		if len(candidates) > 0 && (s.roll == nil || s.roll(100) < int(math.Round(compiledTalents.Crit.SkinnerChance*100))) {
			target := candidates[0]
			if s.roll != nil && len(candidates) > 1 {
				target = candidates[s.roll(len(candidates))]
			}
			targetKey := TalentPartKey(target.X, target.Y)
			combatState.SkinnerParts[targetKey] = now + compiledTalents.Crit.SkinnerDuration
			combatState.SkinnerDurationByPart[targetKey] = compiledTalents.Crit.SkinnerDuration
			combatState.SkinnerCooldownEndsAt = now + compiledTalents.Crit.SkinnerCooldown
			combatState.SkinnerCooldownDuration = compiledTalents.Crit.SkinnerCooldown
		}
	}

	totalDamage := actualDamage
	result.PartStateDeltas = append(result.PartStateDeltas, BossPartStateDelta{
		X:        part.X,
		Y:        part.Y,
		Damage:   actualDamage,
		BeforeHP: beforeHP,
		AfterHP:  part.CurrentHP,
		PartType: string(part.Type),
	})

	extraDamage, talentEvents, damageTypeOverride := s.applyTriggeredTalentDamage(ctx, boss, part, nickname, result.UserStats.ClickCount, actualDamage, critical, targetIdx, combatStats, effectivePartType, compiledTalents, combatState, now, nowMs)
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

	if combatState.SilverStormActive && compiledTalents.Normal.SilverStormDamageRatio > 0 && part.Alive {
		silverStormDamage := int64(float64(maxInt64(1, totalDamage)) * compiledTalents.Normal.SilverStormDamageRatio)
		if silverStormDamage > 0 {
			silverBeforeHP, silverActualDamage, _ := applyBossPartDamageDelta(boss, part, silverStormDamage)
			if silverActualDamage > 0 {
				totalDamage += silverActualDamage
				result.PartStateDeltas = append(result.PartStateDeltas, BossPartStateDelta{
					X:        part.X,
					Y:        part.Y,
					Damage:   silverActualDamage,
					BeforeHP: silverBeforeHP,
					AfterHP:  part.CurrentHP,
					PartType: string(part.Type),
				})
				result.TalentEvents = append(result.TalentEvents, TalentTriggerEvent{
					TalentID:    "normal_ultimate",
					Name:        "白银风暴",
					ExtraDamage: silverActualDamage,
					Message:     "银风斩击",
					PartX:       part.X,
					PartY:       part.Y,
				})
			}
		}
	}

	partDiedThisClick := partWasAlive && !part.Alive
	if partDiedThisClick && compiledTalents.Has("normal_ultimate") {
		combatState.SilverStormActive = true
		duration := int(compiledTalents.Normal.SilverStormDuration)
		combatState.SilverStormRemaining = duration
		combatState.SilverStormEndsAt = now + int64(duration)
		result.TalentEvents = append(result.TalentEvents, TalentTriggerEvent{
			TalentID:   "normal_ultimate",
			Name:       "silverstorm",
			EffectType: "silver_storm",
			Message:    fmt.Sprintf("白银风暴激活！持续 %d 秒", duration),
		})
	}

	result.BossDamage = totalDamage
	result.Critical = critical
	isCollapsed := slices.Contains(combatState.CollapseParts, targetIdx)
	result.DamageType = resolveBossDamageType(resolveBossDamageTypeInput{
		PartType:    part.Type,
		Critical:    critical,
		BossDamage:  totalDamage,
		BossMaxHP:   boss.MaxHP,
		IsCollapsed: isCollapsed,
		IsAfkAttack: false,
	})
	if damageTypeOverride != "" {
		result.DamageType = damageTypeOverride
	}

	// 计算动态触发阈值（受天赋影响）
	if compiledTalents.Has("normal_core") {
		combatState.NormalTriggerCount = compiledTalents.Normal.TriggerCount
	}
	if compiledTalents.Has("armor_core") {
		combatState.ArmorTriggerCount = compiledTalents.Armor.CollapseTrigger
	}
	if compiledTalents.Has("armor_auto_strike") {
		combatState.AutoStrikeTriggerCount = compiledTalents.Armor.AutoStrikeTrigger
		combatState.AutoStrikeWindowSec = TalentAutoStrikeWindowSec
	}
	if compiledTalents.Has("armor_ultimate") {
		combatState.JudgmentDayTriggerCount = compiledTalents.Armor.UltimateTrigger
		combatState.JudgmentDayCooldownSec = compiledTalents.Armor.UltimateCooldown
	}

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
	combatStateRaw, marshalCombatStateErr := sonic.Marshal(combatState)
	if marshalCombatStateErr != nil {
		return result, nil
	}

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, s.bossCurrentKeyForRoom(boss.RoomID), bossValues)
	if totalDamage > 0 {
		pipe.ZIncrBy(ctx, s.bossDamageKey(boss.ID), float64(totalDamage), nickname)
	}
	pipe.HSet(ctx, s.talentCombatStateKey(nickname, boss.ID), "state", string(combatStateRaw))
	pipe.SAdd(ctx, s.talentCombatStateIndexKey(boss.ID), nickname)
	if _, execErr := pipe.Exec(ctx); execErr != nil {
		return result, nil
	}

	if totalDamage > 0 {
		if myBossDamage, summaryErr := s.client.ZScore(ctx, s.bossDamageKey(boss.ID), nickname).Result(); summaryErr == nil {
			result.MyBossDamage = int64(math.Round(myBossDamage))
		}
		if bossLeaderboardCount, summaryErr := s.client.ZCard(ctx, s.bossDamageKey(boss.ID)).Result(); summaryErr == nil {
			result.BossLeaderboardCount = int(bossLeaderboardCount)
		}
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

func applyTalentBleedTicks(boss *Boss, combatState *TalentCombatState, nowMs int64) (int64, []TalentTriggerEvent, []BossPartStateDelta, bool) {
	if boss == nil || combatState == nil || len(combatState.Bleeds) == 0 {
		return 0, nil, nil, false
	}

	totalDamage := int64(0)
	var events []TalentTriggerEvent
	var deltas []BossPartStateDelta
	changed := false
	for partKey, bleed := range combatState.Bleeds {
		if bleed.TickIntervalMs <= 0 || bleed.TotalTicks <= 0 || bleed.TotalDamage <= 0 || bleed.EndsAtMs <= bleed.StartedAtMs {
			delete(combatState.Bleeds, partKey)
			changed = true
			continue
		}

		partX, partY, ok := ParseTalentPartKey(partKey)
		if !ok {
			delete(combatState.Bleeds, partKey)
			changed = true
			continue
		}

		var part *BossPart
		for i := range boss.Parts {
			if boss.Parts[i].X == partX && boss.Parts[i].Y == partY {
				part = &boss.Parts[i]
				break
			}
		}
		if part == nil || !part.Alive || part.CurrentHP <= 0 {
			delete(combatState.Bleeds, partKey)
			changed = true
			continue
		}

		if nowMs < bleed.NextTickAtMs {
			continue
		}

		dueTicks := ((nowMs - bleed.NextTickAtMs) / bleed.TickIntervalMs) + 1
		remainingTicks := bleed.TotalTicks - bleed.AppliedTicks
		if dueTicks > remainingTicks {
			dueTicks = remainingTicks
		}
		if dueTicks <= 0 {
			continue
		}

		prevTicks := bleed.AppliedTicks
		nextTicks := prevTicks + dueTicks
		prevDamage := (bleed.TotalDamage * prevTicks) / bleed.TotalTicks
		nextDamage := (bleed.TotalDamage * nextTicks) / bleed.TotalTicks
		pendingDamage := nextDamage - prevDamage
		changed = true
		if pendingDamage > 0 {
			beforeHP, actualDamage, _ := applyBossPartDamageDelta(boss, part, pendingDamage)
			if actualDamage > 0 {
				totalDamage += actualDamage
				deltas = append(deltas, BossPartStateDelta{
					X:        part.X,
					Y:        part.Y,
					Damage:   actualDamage,
					BeforeHP: beforeHP,
					AfterHP:  part.CurrentHP,
					PartType: string(part.Type),
				})
				events = append(events, TalentTriggerEvent{
					TalentID:    "crit_bleed",
					Name:        "致命出血",
					EffectType:  "bleed",
					ExtraDamage: actualDamage,
					Message:     "出血结算",
					PartX:       part.X,
					PartY:       part.Y,
				})
			}
		}

		bleed.AppliedTicks = nextTicks
		bleed.AppliedDamage = nextDamage
		bleed.NextTickAtMs += dueTicks * bleed.TickIntervalMs
		if bleed.AppliedTicks >= bleed.TotalTicks || nowMs >= bleed.EndsAtMs || !part.Alive || bleed.AppliedDamage >= bleed.TotalDamage {
			delete(combatState.Bleeds, partKey)
			continue
		}
		combatState.Bleeds[partKey] = bleed
	}
	return totalDamage, events, deltas, changed
}

func (s *Store) ProcessTalentBleedTicks(ctx context.Context) ([]StateChange, error) {
	boss, err := s.currentBoss(ctx)
	if err != nil {
		return nil, err
	}
	if boss == nil || boss.Status != bossStatusActive {
		return nil, nil
	}

	nicknames, err := s.listTalentCombatStateNicknames(ctx, boss.ID)
	if err != nil {
		return nil, err
	}
	if len(nicknames) == 0 {
		return nil, nil
	}

	nowTime := s.now()
	nowMs := nowTime.UnixMilli()
	now := nowTime.Unix()
	changes := make([]StateChange, 0, len(nicknames))
	for _, nickname := range nicknames {
		combatState, err := s.GetTalentCombatState(ctx, nickname, boss.ID)
		if err != nil || combatState == nil || len(combatState.Bleeds) == 0 {
			continue
		}

		damage, events, _, changed := applyTalentBleedTicks(boss, combatState, nowMs)
		if !changed {
			continue
		}
		if err := s.persistTalentTickState(ctx, boss, nickname, combatState, damage); err != nil {
			return nil, err
		}
		if err := s.appendPendingTalentEvents(ctx, nickname, boss.ID, events); err != nil {
			return nil, err
		}

		change := StateChange{
			Type:      StateChangeBossChanged,
			Nickname:  nickname,
			Timestamp: now,
		}
		if boss.Status == bossStatusActive {
			changes = append(changes, change)
			continue
		}

		change.BroadcastUserAll = true
		if _, _, err := s.finalizeBossKill(ctx, boss, false, ""); err != nil {
			return nil, err
		}
		changes = append(changes, change)
		break
	}
	return changes, nil
}

func (s *Store) listTalentCombatStateNicknames(ctx context.Context, bossID string) ([]string, error) {
	return s.client.SMembers(ctx, s.talentCombatStateIndexKey(bossID)).Result()
}

func (s *Store) persistTalentTickState(ctx context.Context, boss *Boss, nickname string, combatState *TalentCombatState, totalDamage int64) error {
	if boss == nil || combatState == nil {
		return nil
	}

	allDead := true
	for _, part := range boss.Parts {
		if part.Alive {
			allDead = false
			break
		}
	}
	if allDead {
		boss.Status = bossStatusDefeated
		boss.DefeatedAt = s.now().Unix()
	}

	partsRaw, err := sonic.Marshal(boss.Parts)
	if err != nil {
		return err
	}
	combatStateRaw, err := sonic.Marshal(combatState)
	if err != nil {
		return err
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
	pipe.HSet(ctx, s.bossCurrentKeyForRoom(boss.RoomID), bossValues)
	if totalDamage > 0 {
		pipe.ZIncrBy(ctx, s.bossDamageKey(boss.ID), float64(totalDamage), nickname)
	}
	pipe.HSet(ctx, s.talentCombatStateKey(nickname, boss.ID), "state", string(combatStateRaw))
	pipe.SAdd(ctx, s.talentCombatStateIndexKey(boss.ID), nickname)
	_, err = pipe.Exec(ctx)
	return err
}

func applyComboDamageAmplify(baseDamage int64, comboCount int64) int64 {
	if baseDamage <= 0 {
		return baseDamage
	}
	if comboCount < 25 {
		return baseDamage
	}
	comboAmplify := float64(comboCount/25) * 0.10
	return int64(float64(baseDamage) * (1.0 + comboAmplify))
}

func applyBossPartDamageDelta(boss *Boss, part *BossPart, damage int64) (beforeHP int64, actualDamage int64, partJustDied bool) {
	if boss == nil || part == nil {
		return 0, 0, false
	}

	beforeHP = maxInt64(0, part.CurrentHP)
	actualDamage = min(maxInt64(damage, 0), beforeHP)
	part.CurrentHP = beforeHP - actualDamage
	partWasAlive := part.Alive
	if part.CurrentHP <= 0 {
		part.CurrentHP = 0
		part.Alive = false
	}
	if boss.CurrentHP > 0 {
		boss.CurrentHP -= actualDamage
		if boss.CurrentHP < 0 {
			boss.CurrentHP = 0
		}
	}

	return beforeHP, actualDamage, partWasAlive && !part.Alive
}

func (s *Store) applyTriggeredTalentDamage(ctx context.Context, boss *Boss, part *BossPart, nickname string, clickCount int64, baseDamage int64, isCritical bool, partIndex int, combatStats CombatStats, effectivePartType PartType, compiledTalents *CompiledTalentSet, combatState *TalentCombatState, now, nowMs int64) (int64, []TalentTriggerEvent, string) {
	if boss == nil || part == nil || strings.TrimSpace(nickname) == "" || clickCount <= 0 {
		return 0, nil, ""
	}
	if compiledTalents == nil {
		compiledTalents = compileTalentSet(nil)
	}

	triggerCtx := &talentTriggerContext{
		boss:              boss,
		part:              part,
		nickname:          nickname,
		clickCount:        clickCount,
		baseDamage:        baseDamage,
		isCritical:        isCritical,
		partIndex:         partIndex,
		combatStats:       combatStats,
		effectivePartType: effectivePartType,
		compiledTalents:   compiledTalents,
		combatState:       combatState,
		now:               now,
		nowMs:             nowMs,
		roll:              s.roll,
	}
	for _, trigger := range compiledTalents.triggers {
		trigger(triggerCtx)
	}

	if combatState.CollapseEndsAt > 0 && now >= combatState.CollapseEndsAt {
		combatState.CollapseParts = nil
		combatState.CollapseEndsAt = 0
	}

	return triggerCtx.totalExtra, triggerCtx.events, triggerCtx.damageTypeOverride
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
	IsCollapsed bool
	IsAfkAttack bool
}

func resolveBossDamageType(input resolveBossDamageTypeInput) string {
	damage := maxInt64(0, input.BossDamage)
	maxHP := maxInt64(1, input.BossMaxHP)
	damageRatio := float64(damage) / float64(maxHP)

	if damageRatio >= 0.2 {
		return "doomsday"
	}
	if input.Critical && damageRatio >= 0.10 {
		return "judgement"
	}
	if input.Critical && input.PartType == PartTypeWeak {
		return "weakCritical"
	}
	if input.Critical {
		return "critical"
	}
	if input.IsCollapsed {
		return "trueDamage"
	}
	if input.PartType == PartTypeHeavy {
		return "heavy"
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
		current, currentErr := s.currentBossForRoom(ctx, boss.RoomID)
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
	pipe.Incr(ctx, s.totalBossKillsKey())
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
	qualifiedNicknames := make([]string, 0, len(participants))
	for _, participant := range participants {
		nickname, ok := participant.Member.(string)
		if !ok || nickname == "" || participant.Score < float64(minDamage) {
			continue
		}
		qualifiedNicknames = append(qualifiedNicknames, nickname)

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
		pipe.HIncrBy(ctx, s.resourceKey(nickname), "boss_kills", 1)

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
	for _, nickname := range qualifiedNicknames {
		if recordErr := s.recordTaskEvent(ctx, nickname, TaskEventBossKill, 1); recordErr != nil {
			return nil, nil, recordErr
		}
	}

	if err := s.SaveBossToHistory(ctx, boss); err != nil {
		return nil, nil, err
	}

	enabled, err := s.bossCycleEnabledForRoom(ctx, boss.RoomID)
	if err != nil {
		return nil, nil, err
	}
	if enabled {
		nextBoss, err := s.activateNextBossFromCycleForRoom(ctx, boss.RoomID, boss.TemplateID)
		if err != nil && !errors.Is(err, ErrBossPoolEmpty) && !errors.Is(err, ErrBossCycleQueueEmpty) {
			return nil, nil, err
		}
		if nextBoss != nil {
			return nextBoss, rewardForNickname, nil
		}
	}

	current, currentErr := s.currentBossForRoom(ctx, boss.RoomID)
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
	if cached, ok := s.cachedCombatStats(nickname); ok {
		return cached, nil
	}

	stats := s.baseCombatStats()

	attackPower, armorPen, critRate, critDmgMult, partTypeSoft, partTypeHeavy, partTypeWeak := loadoutBonuses(loadout)
	stats.AttackPower += attackPower
	stats.ArmorPenPercent = clampFloat(stats.ArmorPenPercent+armorPen, 0, 1.0)
	stats.CriticalChancePercent += critRate * 100
	stats.CritDamageMultiplier += critDmgMult
	stats.PartTypeDamageSoft += partTypeSoft
	stats.PartTypeDamageHeavy += partTypeHeavy
	stats.PartTypeDamageWeak += partTypeWeak

	compiledTalents, err := s.compiledTalentSetForNickname(ctx, nickname)
	if err != nil {
		return CombatStats{}, err
	}
	mods := compiledTalents.Modifiers
	if mods != nil {
		stats.AttackPower = int64(float64(stats.AttackPower) * (1 + max(0.0, mods.AttackPowerPercent)))
		stats.ArmorPenPercent = clampFloat(stats.ArmorPenPercent+mods.ArmorPenExtra, 0, 1.0)
		stats.AllDamageAmplify += mods.AllDamageAmplify
		stats.CriticalChancePercent += mods.CritRateBonus
		stats.CritDamageMultiplier += mods.CritDamagePercentBonus
		stats.PerPartDamagePercent = max(0.0, mods.PerPartDamagePercent)
		stats.LowHpMultiplier = max(1.0, mods.LowHpMultiplier)
		stats.LowHpThreshold = clampFloat(mods.LowHpThreshold, 0, 1)
		stats.PartTypeDamageSoft += max(0.0, mods.PartTypeBonus[PartTypeSoft])
		stats.PartTypeDamageHeavy += max(0.0, mods.PartTypeBonus[PartTypeHeavy])
		stats.PartTypeDamageWeak += max(0.0, mods.PartTypeBonus[PartTypeWeak])
		if mods.PenToAmplifyRatio > 0 {
			stats.AllDamageAmplify += stats.ArmorPenPercent * mods.PenToAmplifyRatio
		}
		if mods.OverflowToCritDmgRatio > 0 && stats.CriticalChancePercent > 100 {
			overflow := stats.CriticalChancePercent - 100
			stats.CritDamageMultiplier += overflow * mods.OverflowToCritDmgRatio
			stats.CriticalChancePercent = 100
		}
	}
	stats.CriticalChancePercent = clampFloat(stats.CriticalChancePercent, 0, 100)

	result := deriveCombatStats(stats)
	s.storeCombatStatsCache(nickname, result)
	return result, nil
}

func (s *Store) baseCombatStats() CombatStats {
	return deriveCombatStats(CombatStats{
		CriticalChancePercent: clampFloat(float64(s.critical.CriticalChancePercent), 0, 100),
		AttackPower:           5,
		ArmorPenPercent:       0,
		CritDamageMultiplier:  1.5,
		AllDamageAmplify:      0,
		LowHpMultiplier:       1,
	})
}

func loadoutBonuses(loadout Loadout) (attackPower int64, armorPen float64, critRate float64, critDmgMult float64, partTypeSoft float64, partTypeHeavy float64, partTypeWeak float64) {
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
		partTypeSoft += item.PartTypeDamageSoft
		partTypeHeavy += item.PartTypeDamageHeavy
		partTypeWeak += item.PartTypeDamageWeak
	}
	return
}

func deriveCombatStats(stats CombatStats) CombatStats {
	stats.EffectiveIncrement = max(1, stats.AttackPower)
	stats.NormalDamage = stats.EffectiveIncrement

	if stats.CritDamageMultiplier < 1.0 {
		stats.CritDamageMultiplier = 1.0
	}
	multiplierBasedCriticalDamage := max(int64(float64(stats.NormalDamage)*stats.CritDamageMultiplier), stats.NormalDamage)
	stats.CriticalDamage = multiplierBasedCriticalDamage

	return stats
}

func hasCriticalBonus(stats CombatStats) bool {
	return stats.CritDamageMultiplier > 1.0
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
	effectiveArmor := max(int64(float64(partArmor)*(1.0-clampFloat(stats.ArmorPenPercent, 0, 0.80))), 0)

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
			amplifyBonus += stats.LowHpMultiplier - 1
		}
	}

	amplify := 1.0 + amplifyBonus

	// 暴击乘区（这里只计算“命中暴击时应造成多少伤害”，不在这里做第二次暴击判定）
	critMult := max(1.0, stats.CritDamageMultiplier)

	// 最终伤害
	normalDamage := max(int64(float64(baseDamage)*amplify), 1)

	criticalDamage := max(int64(float64(baseDamage)*amplify*critMult), 1)

	return CombatStats{
		NormalDamage:          normalDamage,
		CriticalDamage:        criticalDamage,
		CriticalChancePercent: stats.CriticalChancePercent,
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
	return s.currentBossForRoom(ctx, s.defaultRoomID())
}

func (s *Store) currentBossForRoom(ctx context.Context, roomID string) (*Boss, error) {
	if isHallRoomID(roomID) {
		return nil, nil
	}
	return s.currentBossFromCmdable(ctx, s.client, roomID)
}

func (s *Store) currentBossFromCmdable(ctx context.Context, client redis.Cmdable, roomID string) (*Boss, error) {
	if isHallRoomID(roomID) {
		return nil, nil
	}
	roomID = s.normalizeRoomID(roomID)
	values, err := client.HMGet(ctx, s.bossCurrentKeyForRoom(roomID),
		"id",
		"template_id",
		"room_id",
		"queue_id",
		"name",
		"status",
		"max_hp",
		"current_hp",
		"gold_on_kill",
		"stone_on_kill",
		"talent_points_on_kill",
		"parts",
		"started_at",
		"defeated_at",
	).Result()
	if err != nil {
		return nil, err
	}
	id := strings.TrimSpace(stringValue(values, 0))
	name := strings.TrimSpace(stringValue(values, 4))
	if id == "" && name == "" {
		return nil, nil
	}

	var parts []BossPart
	if partsRaw := strings.TrimSpace(stringValue(values, 11)); partsRaw != "" {
		_ = sonic.Unmarshal([]byte(partsRaw), &parts)
	}

	return &Boss{
		ID:                 id,
		TemplateID:         strings.TrimSpace(stringValue(values, 1)),
		RoomID:             firstNonEmpty(strings.TrimSpace(stringValue(values, 2)), roomID),
		QueueID:            firstNonEmpty(strings.TrimSpace(stringValue(values, 3)), s.queueIDForRoom(roomID)),
		Name:               name,
		Status:             strings.TrimSpace(stringValue(values, 5)),
		MaxHP:              int64Value(values, 6),
		CurrentHP:          int64Value(values, 7),
		GoldOnKill:         int64Value(values, 8),
		StoneOnKill:        int64Value(values, 9),
		TalentPointsOnKill: int64Value(values, 10),
		Parts:              parts,
		StartedAt:          int64Value(values, 12),
		DefeatedAt:         int64Value(values, 13),
	}, nil
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
		RoomID:             strings.TrimSpace(values["room_id"]),
		QueueID:            strings.TrimSpace(values["queue_id"]),
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

	pipe := s.client.Pipeline()
	scoreCmd := pipe.ZScore(ctx, s.bossDamageKey(bossID), nickname)
	rankCmd := pipe.ZRevRank(ctx, s.bossDamageKey(bossID), nickname)
	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	if scoreErr := scoreCmd.Err(); scoreErr != nil {
		if errors.Is(scoreErr, redis.Nil) {
			return nil, nil
		}
		return nil, scoreErr
	}

	stats := &BossUserStats{
		Nickname: nickname,
		Damage:   int64(scoreCmd.Val()),
	}

	if rankErr := rankCmd.Err(); rankErr != nil {
		if !errors.Is(rankErr, redis.Nil) {
			return nil, rankErr
		}
	} else {
		stats.Rank = int(rankCmd.Val()) + 1
	}

	return stats, nil
}

func (s *Store) getEquipmentDefinition(ctx context.Context, itemID string) (EquipmentDefinition, error) {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return EquipmentDefinition{}, ErrEquipmentNotFound
	}
	if definition, ok := s.cachedEquipmentDefinition(itemID); ok {
		return definition, nil
	}

	definition, err := s.loadEquipmentDefinitionFromRedis(ctx, itemID)
	if err != nil {
		return EquipmentDefinition{}, err
	}
	s.storeEquipmentDefinitionCache(itemID, definition)
	return definition, nil
}

func (s *Store) loadEquipmentDefinitionFromRedis(ctx context.Context, itemID string) (EquipmentDefinition, error) {
	values, err := s.client.HMGet(ctx, s.equipmentKey(itemID),
		"name",
		"slot",
		"rarity",
		"image_path",
		"image_alt",
		"attack_power",
		"armor_pen_percent",
		"crit_rate",
		"crit_damage_multiplier",
		"part_type_damage_soft",
		"part_type_damage_heavy",
		"part_type_damage_weak",
		"talent_affinity",
	).Result()
	if err != nil {
		return EquipmentDefinition{}, err
	}
	if stringValue(values, 0) == "" &&
		stringValue(values, 1) == "" &&
		stringValue(values, 2) == "" &&
		stringValue(values, 5) == "" {
		return EquipmentDefinition{}, ErrEquipmentNotFound
	}

	return EquipmentDefinition{
		ItemID:               itemID,
		Name:                 firstNonEmpty(strings.TrimSpace(stringValue(values, 0)), itemID),
		Slot:                 normalizeEquipmentSlot(stringValue(values, 1)),
		Rarity:               normalizeEquipmentRarity(stringValue(values, 2)),
		ImagePath:            strings.TrimSpace(stringValue(values, 3)),
		ImageAlt:             strings.TrimSpace(stringValue(values, 4)),
		AttackPower:          int64Value(values, 5),
		ArmorPenPercent:      float64FromString(stringValue(values, 6)),
		CritRate:             float64FromString(stringValue(values, 7)),
		CritDamageMultiplier: float64FromString(stringValue(values, 8)),
		PartTypeDamageSoft:   float64FromString(stringValue(values, 9)),
		PartTypeDamageHeavy:  float64FromString(stringValue(values, 10)),
		PartTypeDamageWeak:   float64FromString(stringValue(values, 11)),
		TalentAffinity:       strings.TrimSpace(stringValue(values, 12)),
	}, nil
}

func (s *Store) getEquipmentDefinitionFromRequestCache(ctx context.Context, itemID string, requestCache *equipmentDefinitionRequestCache) (EquipmentDefinition, error) {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return EquipmentDefinition{}, ErrEquipmentNotFound
	}
	if requestCache == nil {
		return s.getEquipmentDefinition(ctx, itemID)
	}
	if definition, ok := requestCache.definitions[itemID]; ok {
		return definition, nil
	}
	if _, ok := requestCache.missing[itemID]; ok {
		return EquipmentDefinition{}, ErrEquipmentNotFound
	}

	definition, err := s.getEquipmentDefinition(ctx, itemID)
	if err != nil {
		if errors.Is(err, ErrEquipmentNotFound) {
			requestCache.missing[itemID] = struct{}{}
		}
		return EquipmentDefinition{}, err
	}
	requestCache.definitions[itemID] = definition
	return definition, nil
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

	filteredIDs := make([]string, 0, len(instanceIDs))
	pipe := s.client.Pipeline()
	instanceCmds := make(map[string]*redis.SliceCmd, len(instanceIDs))
	for _, instanceID := range instanceIDs {
		instanceID = strings.TrimSpace(instanceID)
		if instanceID == "" {
			continue
		}
		filteredIDs = append(filteredIDs, instanceID)
		instanceCmds[instanceID] = pipe.HMGet(ctx, s.equipmentInstanceKey(instanceID),
			"item_id",
			"enhance_level",
			"spent_stones",
			"bound",
			"locked",
			"created_at",
		)
	}
	if len(filteredIDs) == 0 {
		return map[string]ItemInstance{}, nil
	}
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	instances := make(map[string]ItemInstance, len(instanceIDs))
	for _, instanceID := range filteredIDs {
		cmd := instanceCmds[instanceID]
		if cmd == nil {
			continue
		}
		if cmdErr := cmd.Err(); cmdErr != nil {
			if errors.Is(cmdErr, redis.Nil) {
				continue
			}
			return nil, cmdErr
		}
		values := cmd.Val()
		itemID := strings.TrimSpace(stringValue(values, 0))
		if itemID == "" {
			continue
		}
		instance := ItemInstance{
			InstanceID:   instanceID,
			ItemID:       itemID,
			EnhanceLevel: int(int64Value(values, 1)),
			SpentStones:  int64Value(values, 2),
			Bound:        int64Value(values, 3) > 0,
			Locked:       int64Value(values, 4) > 0,
			CreatedAt:    int64Value(values, 5),
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

	values, err := s.client.HMGet(ctx, s.equipmentInstanceKey(ref),
		"item_id",
		"enhance_level",
		"spent_stones",
		"bound",
		"locked",
		"created_at",
	).Result()
	if err != nil {
		return nil, err
	}
	if stringValue(values, 0) == "" {
		return nil, ErrEquipmentNotOwned
	}
	itemID := strings.TrimSpace(stringValue(values, 0))
	if itemID == "" {
		return nil, ErrEquipmentNotOwned
	}

	instance := &ItemInstance{
		InstanceID:   ref,
		ItemID:       itemID,
		EnhanceLevel: int(int64Value(values, 1)),
		SpentStones:  int64Value(values, 2),
		Bound:        int64Value(values, 3) > 0,
		Locked:       int64Value(values, 4) > 0,
		CreatedAt:    int64Value(values, 5),
	}
	return instance, nil
}

func (s *Store) loadoutRefsForNickname(ctx context.Context, nickname string) (map[string]string, error) {
	values, err := s.client.HMGet(ctx, s.loadoutKey(nickname), loadoutSlots...).Result()
	if err != nil {
		return nil, err
	}

	loadoutRefs := make(map[string]string, len(loadoutSlots))
	for index, slot := range loadoutSlots {
		equippedRef := strings.TrimSpace(stringValue(values, index))
		if equippedRef == "" {
			continue
		}
		loadoutRefs[slot] = equippedRef
	}
	return loadoutRefs, nil
}

func (s *Store) loadoutFromRefs(ctx context.Context, loadoutRefs map[string]string, instances map[string]ItemInstance, requestCache *equipmentDefinitionRequestCache) (Loadout, map[string]string) {
	loadout := Loadout{}
	equipped := make(map[string]string, len(loadoutRefs))
	for _, slot := range loadoutSlots {
		equippedRef := strings.TrimSpace(loadoutRefs[slot])
		if equippedRef == "" {
			continue
		}

		instance, ok := instances[equippedRef]
		if !ok {
			continue
		}
		definition, defErr := s.getEquipmentDefinitionFromRequestCache(ctx, instance.ItemID, requestCache)
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

	return loadout, equipped
}

func (s *Store) loadoutForNickname(ctx context.Context, nickname string) (Loadout, map[string]string, error) {
	loadoutRefs, err := s.loadoutRefsForNickname(ctx, nickname)
	if err != nil {
		return Loadout{}, nil, err
	}
	instances, err := s.itemInstancesByIDForNickname(ctx, nickname)
	if err != nil {
		return Loadout{}, nil, err
	}
	loadout, equipped := s.loadoutFromRefs(ctx, loadoutRefs, instances, newEquipmentDefinitionRequestCache())
	return loadout, equipped, nil
}

func (s *Store) inventoryFromInstances(ctx context.Context, instances map[string]ItemInstance, equipped map[string]string, requestCache *equipmentDefinitionRequestCache) []InventoryItem {
	items := make([]InventoryItem, 0, len(instances))
	for _, instance := range instances {
		definition, err := s.getEquipmentDefinitionFromRequestCache(ctx, instance.ItemID, requestCache)
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
		return []InventoryItem{}
	}

	slices.SortFunc(items, func(left, right InventoryItem) int {
		if left.Slot == right.Slot {
			return strings.Compare(left.Name, right.Name)
		}
		return strings.Compare(left.Slot, right.Slot)
	})

	return items
}

func (s *Store) inventoryForNickname(ctx context.Context, nickname string, equipped map[string]string) ([]InventoryItem, error) {
	instances, err := s.itemInstancesByIDForNickname(ctx, nickname)
	if err != nil {
		return nil, err
	}
	return s.inventoryFromInstances(ctx, instances, equipped, newEquipmentDefinitionRequestCache()), nil
}

func (s *Store) recentRewardsForNickname(ctx context.Context, nickname string) ([]Reward, error) {
	values, err := s.client.HMGet(ctx, s.lastRewardKey(nickname),
		"boss_id",
		"boss_name",
		"item_id",
		"item_name",
		"granted_at",
		"recent_rewards",
	).Result()
	if err != nil {
		return nil, err
	}
	if stringValue(values, 2) == "" && stringValue(values, 5) == "" {
		return []Reward{}, nil
	}

	if raw := strings.TrimSpace(stringValue(values, 5)); raw != "" {
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

func rewardFromRecordValues(values []any) *Reward {
	if len(values) == 0 || strings.TrimSpace(stringValue(values, 2)) == "" {
		return nil
	}

	return &Reward{
		BossID:    strings.TrimSpace(stringValue(values, 0)),
		BossName:  strings.TrimSpace(stringValue(values, 1)),
		ItemID:    strings.TrimSpace(stringValue(values, 2)),
		ItemName:  strings.TrimSpace(stringValue(values, 3)),
		GrantedAt: int64Value(values, 4),
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
	BossKills    int64
}

type PlayerResources struct {
	Gold         int64
	Stones       int64
	TalentPoints int64
}

type BossKillBackfillStats struct {
	HistoryBosses int64
	PlayerCount   int64
}

func (s *Store) resourcesForNickname(ctx context.Context, nickname string) (playerResources, error) {
	resourceKey := s.resourceKey(nickname)
	values, err := s.client.HMGet(ctx, resourceKey, "gold", "stones", "talent_points", "boss_kills").Result()
	if err != nil {
		return playerResources{}, err
	}

	return playerResources{
		Gold:         int64Value(values, 0),
		Stones:       int64Value(values, 1),
		TalentPoints: int64Value(values, 2),
		BossKills:    int64Value(values, 3),
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

func (s *Store) totalBossKillsKey() string {
	return s.namespace + "total:boss:kills"
}

func (s *Store) totalBossKills(ctx context.Context) (int64, error) {
	value, err := s.client.Get(ctx, s.totalBossKillsKey()).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			if s.bossHistoryStore != nil {
				items, listErr := s.bossHistoryStore.ListBossHistory(ctx)
				if listErr != nil {
					return 0, listErr
				}
				return int64(len(items)), nil
			}
			count, countErr := s.client.ZCard(ctx, s.bossHistoryKey).Result()
			if countErr != nil {
				return 0, countErr
			}
			return count, nil
		}
		return 0, err
	}
	return int64FromString(value), nil
}

func (s *Store) historyBossKillsForNickname(ctx context.Context, nickname string) (int64, error) {
	entries, err := s.ListBossHistory(ctx)
	if err != nil {
		return 0, err
	}
	normalizedNickname := strings.TrimSpace(nickname)
	if normalizedNickname == "" {
		return 0, nil
	}
	var total int64
	for _, entry := range entries {
		counted := false
		if score, scoreErr := s.client.ZScore(ctx, s.bossDamageKey(entry.Boss.ID), normalizedNickname).Result(); scoreErr == nil && score > 0 {
			total++
			continue
		}
		for _, item := range entry.Damage {
			if item.Nickname == normalizedNickname && item.Damage > 0 {
				counted = true
				break
			}
		}
		if counted {
			total++
		}
	}
	return total, nil
}

func (s *Store) RebuildBossKillCounters(ctx context.Context) (BossKillBackfillStats, error) {
	entries, err := s.ListBossHistory(ctx)
	if err != nil {
		return BossKillBackfillStats{}, err
	}

	perPlayer := make(map[string]int64)
	for _, entry := range entries {
		seen := make(map[string]struct{})
		damageEntries, damageErr := s.client.ZRangeWithScores(ctx, s.bossDamageKey(entry.ID), 0, -1).Result()
		if damageErr == nil && len(damageEntries) > 0 {
			for _, item := range damageEntries {
				nickname, ok := item.Member.(string)
				if !ok || strings.TrimSpace(nickname) == "" || item.Score <= 0 {
					continue
				}
				seen[strings.TrimSpace(nickname)] = struct{}{}
			}
		} else {
			for _, item := range entry.Damage {
				if strings.TrimSpace(item.Nickname) == "" || item.Damage <= 0 {
					continue
				}
				seen[strings.TrimSpace(item.Nickname)] = struct{}{}
			}
		}
		for nickname := range seen {
			perPlayer[nickname]++
		}
	}

	allNicknames, err := s.listPlayerNicknames(ctx)
	if err != nil {
		return BossKillBackfillStats{}, err
	}
	for nickname := range perPlayer {
		if !slices.Contains(allNicknames, nickname) {
			allNicknames = append(allNicknames, nickname)
		}
	}

	pipe := s.client.TxPipeline()
	for _, nickname := range allNicknames {
		trimmedNickname := strings.TrimSpace(nickname)
		if trimmedNickname == "" {
			continue
		}
		if count, ok := perPlayer[trimmedNickname]; ok && count > 0 {
			pipe.HSet(ctx, s.resourceKey(trimmedNickname), "boss_kills", strconv.FormatInt(count, 10))
			continue
		}
		pipe.HDel(ctx, s.resourceKey(trimmedNickname), "boss_kills")
	}
	pipe.Set(ctx, s.totalBossKillsKey(), strconv.FormatInt(int64(len(entries)), 10), 0)
	if _, err := pipe.Exec(ctx); err != nil {
		return BossKillBackfillStats{}, err
	}

	return BossKillBackfillStats{
		HistoryBosses: int64(len(entries)),
		PlayerCount:   int64(len(perPlayer)),
	}, nil
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
