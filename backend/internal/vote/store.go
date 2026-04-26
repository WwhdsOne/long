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
var ErrTalentTreeNotSet = errors.New("talent tree not set")
var ErrTalentAlreadyLearned = errors.New("talent already learned")
var ErrTalentPrerequisite = errors.New("talent prerequisite not met")
var ErrTalentNotFound = errors.New("talent not found")
var ErrTalentMaxLevel = errors.New("talent max level reached")
var ErrInvalidTalentTree = errors.New("invalid talent tree")

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
	ID          string     `json:"id"`
	TemplateID  string     `json:"templateId,omitempty"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	MaxHP       int64      `json:"maxHp"`
	CurrentHP   int64      `json:"currentHp"`
	GoldOnKill  int64      `json:"goldOnKill"`
	StoneOnKill int64      `json:"stoneOnKill"`
	Parts       []BossPart `json:"parts,omitempty"`
	StartedAt   int64      `json:"startedAt,omitempty"`
	DefeatedAt  int64      `json:"defeatedAt,omitempty"`
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
	BossID     string          `json:"bossId,omitempty"`
	TemplateID string          `json:"templateId,omitempty"`
	Status     string          `json:"status,omitempty"`
	GoldRange  ResourceRange   `json:"goldRange"`
	StoneRange ResourceRange   `json:"stoneRange"`
	BossLoot   []BossLootEntry `json:"bossLoot"`
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
	UserStats     *UserStats      `json:"userStats,omitempty"`
	MyBossStats   *BossUserStats  `json:"myBossStats,omitempty"`
	Inventory     []InventoryItem `json:"inventory"`
	Loadout       Loadout         `json:"loadout"`
	CombatStats   CombatStats     `json:"combatStats"`
	Gems          int64           `json:"gems"`
	Gold          int64           `json:"gold"`
	Stones        int64           `json:"stones"`
	RecentRewards []Reward        `json:"recentRewards,omitempty"`
	LastReward    *Reward         `json:"lastReward,omitempty"`
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
	Gems                int64                  `json:"gems"`
	Gold                int64                  `json:"gold"`
	Stones              int64                  `json:"stones"`
	RecentRewards       []Reward               `json:"recentRewards,omitempty"`
	LastReward          *Reward                `json:"lastReward,omitempty"`
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
	RefundedStones int64  `json:"refundedStones"`
	Stones         int64  `json:"stones"`
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
	userState.Gems = resources.Gems
	userState.Gold = resources.Gold
	userState.Stones = resources.Stones

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

// ClickButton 处理 Boss 部位点击。slug 必须以 boss-part: 开头。
func (s *Store) ClickButton(ctx context.Context, slug string, nickname string) (ClickResult, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return ClickResult{}, err
	}
	slug = strings.TrimSpace(slug)
	if !strings.HasPrefix(slug, bossPartClickSlugPrefix) {
		return ClickResult{}, fmt.Errorf("button not available")
	}
	return s.clickBossPart(ctx, slug, normalizedNickname)
}

// ClickBossPart 处理不绑定按钮的 Boss 部位手动点击。
func (s *Store) ClickBossPart(ctx context.Context, target string, nickname string) (ClickResult, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return ClickResult{}, err
	}
	return s.clickBossPart(ctx, target, normalizedNickname)
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

// SalvageItem 分解装备实例，返还已消耗强化石的 60%（向下取整）。
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
	definition, err := s.getEquipmentDefinition(ctx, instance.ItemID)
	if err != nil {
		return SalvageResult{}, err
	}

	refund := int64(math.Floor(float64(maxInt64(0, instance.SpentStones)) * 0.6))

	pipe := s.client.TxPipeline()
	pipe.SRem(ctx, s.playerInstancesKey(normalizedNickname), instance.InstanceID)
	pipe.Del(ctx, s.equipmentInstanceKey(instance.InstanceID))
	if refund > 0 {
		pipe.HIncrBy(ctx, s.resourceKey(normalizedNickname), "stones", refund)
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
		RefundedStones: refund,
		Stones:         resources.Stones,
	}, nil
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
		Gems:                userState.Gems,
		Gold:                userState.Gold,
		Stones:              userState.Stones,
		RecentRewards:       userState.RecentRewards,
		LastReward:          userState.LastReward,
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
		nextBoss, finalizeErr := s.finalizeBossKill(ctx, boss, true)
		if finalizeErr != nil {
			return result, nil
		}
		if nextBoss != nil {
			result.Boss = nextBoss
		}
	}
	return result, nil
}

func (s *Store) clickBossPart(ctx context.Context, target string, nickname string) (ClickResult, error) {
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
	return s.applyBossPartDamage(ctx, boss, nickname, critical, result, targetIdx)
}

func (s *Store) applyBossPartDamage(ctx context.Context, boss *Boss, nickname string, critical bool, result ClickResult, targetIdx int) (ClickResult, error) {
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

	aliveCount := 0
	for _, p := range boss.Parts {
		if p.Alive {
			aliveCount++
		}
	}

	damageStats := CalcBossPartDamage(combatStats, part.Type, part.Armor, aliveCount)
	partDamage := damageStats.NormalDamage
	if critical {
		partDamage = damageStats.CriticalDamage
	}

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
	if part.CurrentHP <= 0 {
		part.Alive = false
	}

	boss.CurrentHP = sumBossPartCurrentHP(boss.Parts)
	result.BossDamage = actualDamage
	result.Critical = critical
	result.DamageType = resolveBossDamageType(resolveBossDamageTypeInput{
		PartType:    part.Type,
		Critical:    critical,
		BossDamage:  actualDamage,
		BossMaxHP:   boss.MaxHP,
		IsAfkAttack: false,
	})

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
	if actualDamage > 0 {
		pipe.ZIncrBy(ctx, s.bossDamageKey(boss.ID), float64(actualDamage), nickname)
	}
	if _, execErr := pipe.Exec(ctx); execErr != nil {
		return result, nil
	}

	result.Boss = boss

	if allDead {
		result.BroadcastUserAll = true
		nextBoss, finalizeErr := s.finalizeBossKill(ctx, boss, false)
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

func (s *Store) finalizeBossKill(ctx context.Context, boss *Boss, afkMode bool) (*Boss, error) {
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
	participants, err := s.client.ZRevRangeWithScores(ctx, s.bossDamageKey(bossID), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	pipe := s.client.Pipeline()
	now := s.now().Unix()
	minDamage := (maxInt64(1, boss.MaxHP) + 99) / 100
	goldBase := boss.GoldOnKill
	stoneBase := boss.StoneOnKill
	if afkMode {
		goldBase = int64(math.Floor(float64(goldBase) * 0.5))
		stoneBase = int64(math.Floor(float64(stoneBase) * 0.5))
	}
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

		if len(lootEntries) == 0 {
			continue
		}
		rewards := make([]Reward, 0, len(lootEntries))
		for _, reward := range s.rollLootDrops(lootEntries) {
			instanceID, createErr := s.newEquipmentInstanceID(ctx)
			if createErr != nil {
				return nil, createErr
			}
			pipe.HSet(ctx, s.equipmentInstanceKey(instanceID), map[string]any{
				"item_id":       reward.ItemID,
				"enhance_level": "0",
				"spent_stones":  "0",
				"bound":         "0",
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
		}
	}

	if _, err = pipe.Exec(ctx); err != nil {
		return nil, err
	}

	if err := s.SaveBossToHistory(ctx, boss); err != nil {
		return nil, err
	}

	enabled, err := s.bossCycleEnabled(ctx)
	if err != nil {
		return nil, err
	}
	if enabled {
		nextBoss, err := s.activateNextBossFromCycle(ctx, boss.TemplateID)
		if err != nil && !errors.Is(err, ErrBossPoolEmpty) && !errors.Is(err, ErrBossCycleQueueEmpty) {
			return nil, err
		}
		if nextBoss != nil {
			return nextBoss, nil
		}
	}

	return s.currentBoss(ctx)
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
func CalcBossPartDamage(stats CombatStats, partType PartType, partArmor int64, alivePartCount int) CombatStats {
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

	// 增伤乘区 = (1 + 全伤害增幅)
	amplify := 1.0 + stats.AllDamageAmplify

	// 暴击乘区（这里只计算“命中暴击时应造成多少伤害”，不在这里做第二次暴击判定）
	critMult := max(1.0, stats.CritDamageMultiplier)

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
		ID:          id,
		TemplateID:  strings.TrimSpace(values["template_id"]),
		Name:        name,
		Status:      strings.TrimSpace(values["status"]),
		MaxHP:       int64FromString(values["max_hp"]),
		CurrentHP:   int64FromString(values["current_hp"]),
		GoldOnKill:  int64FromString(values["gold_on_kill"]),
		StoneOnKill: int64FromString(values["stone_on_kill"]),
		Parts:       parts,
		StartedAt:   int64FromString(values["started_at"]),
		DefeatedAt:  int64FromString(values["defeated_at"]),
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
		item := buildInventoryItem(definition, 1, true, instance.EnhanceLevel, instance.InstanceID, instance.Bound)
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
			})
			continue
		}
		items = append(items, buildInventoryItem(definition, 1, equipped[instance.InstanceID] != "", instance.EnhanceLevel, instance.InstanceID, instance.Bound))
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

func (s *Store) legacyGemKey(nickname string) string {
	return s.namespace + "gem:" + nickname
}

type playerResources struct {
	Gems   int64
	Gold   int64
	Stones int64
}

func (s *Store) resourcesForNickname(ctx context.Context, nickname string) (playerResources, error) {
	resourceKey := s.resourceKey(nickname)
	values, err := s.client.HMGet(ctx, resourceKey, "gems", "gold", "stones").Result()
	if err != nil {
		return playerResources{}, err
	}

	current := playerResources{
		Gems:   int64Value(values, 0),
		Gold:   int64Value(values, 1),
		Stones: int64Value(values, 2),
	}

	legacyKey := s.legacyGemKey(nickname)
	legacyValues, err := s.client.HMGet(ctx, legacyKey, "gems", "gold", "stones").Result()
	if err != nil {
		return playerResources{}, err
	}
	if !hasAnyHMGetValue(legacyValues) {
		return current, nil
	}

	legacy := playerResources{
		Gems:   int64Value(legacyValues, 0),
		Gold:   int64Value(legacyValues, 1),
		Stones: int64Value(legacyValues, 2),
	}
	merged := legacy
	if hasAnyHMGetValue(values) {
		merged = playerResources{
			Gems:   current.Gems + legacy.Gems,
			Gold:   current.Gold + legacy.Gold,
			Stones: current.Stones + legacy.Stones,
		}
	}

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, resourceKey, map[string]any{
		"gems":   strconv.FormatInt(merged.Gems, 10),
		"gold":   strconv.FormatInt(merged.Gold, 10),
		"stones": strconv.FormatInt(merged.Stones, 10),
	})
	pipe.Del(ctx, legacyKey)
	if _, err := pipe.Exec(ctx); err != nil {
		return playerResources{}, err
	}

	return merged, nil
}

func hasAnyHMGetValue(values []any) bool {
	for _, value := range values {
		if value != nil {
			return true
		}
	}
	return false
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

func buildInventoryItem(definition EquipmentDefinition, quantity int64, equipped bool, enhanceLevel int, instanceID string, bound bool) InventoryItem {
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
