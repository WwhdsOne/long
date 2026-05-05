package httpapi

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"

	adminauth "long/internal/admin"
	"long/internal/archive"
	"long/internal/core"
	"long/internal/events"
	ossupload "long/internal/oss"
	"long/internal/xlog"
)

// ButtonStore 投票存储接口，HTTP 层所需的最小契约。
type ButtonStore interface {
	GetState(context.Context, string) (core.State, error)
	GetSnapshot(context.Context) (core.Snapshot, error)
	ListLeaderboard(context.Context, int64) ([]core.LeaderboardEntry, error)
	ListLeaderboardIncludingZeroClickPlayers(context.Context, int64, int64) ([]core.LeaderboardEntry, error)
	GetUserState(context.Context, string) (core.UserState, error)
	GetPlayerResources(context.Context, string) (core.PlayerResources, error)
	ClickButton(context.Context, string, string, int64) (core.ClickResult, error)
	ClickBossPart(context.Context, string, string) (core.ClickResult, error)
	AttackBossPartAFK(context.Context, string) (core.ClickResult, error)
	AutoClickBossPart(context.Context, string, string) (core.ClickResult, error)
	ValidateNickname(context.Context, string) error
	EquipItem(context.Context, string, string) (core.State, error)
	UnequipItem(context.Context, string, string) (core.State, error)
	EnhanceItem(context.Context, string, string) (core.State, error)
	SalvageItem(context.Context, string, string) (core.SalvageResult, error)
	BulkSalvageUnequipped(context.Context, string) (core.BulkSalvageResult, error)
	LockItem(context.Context, string, string) (core.State, error)
	UnlockItem(context.Context, string, string) (core.State, error)
	GetAdminState(context.Context) (core.AdminState, error)
	ListAdminEquipmentPage(context.Context, int64, int64) (core.AdminEquipmentPage, error)
	ListAdminBossHistoryPage(context.Context, int64, int64) (core.AdminBossHistoryPage, error)
	ListAdminPlayers(context.Context, string, int64) (core.AdminPlayerPage, error)
	GetAdminPlayer(context.Context, string) (*core.AdminPlayerOverview, error)
	SaveEquipmentDefinition(context.Context, core.EquipmentDefinition) error
	DeleteEquipmentDefinition(context.Context, string) error
	ActivateBoss(context.Context, core.BossUpsert) (*core.Boss, error)
	DeactivateBoss(context.Context) error
	SetBossLoot(context.Context, string, []core.BossLootEntry) error
	SaveBossTemplate(context.Context, core.BossTemplateUpsert) error
	DeleteBossTemplate(context.Context, string) error
	SetBossTemplateLoot(context.Context, string, []core.BossLootEntry) error
	SetBossCycleQueue(context.Context, []string) ([]string, error)
	SetBossCycleEnabled(context.Context, bool) (*core.Boss, error)
	ListBossHistory(context.Context) ([]core.BossHistoryEntry, error)
	GetBossResources(context.Context) (core.BossResources, error)
	GetLatestAnnouncement(context.Context) (*core.Announcement, error)
	ListAnnouncements(context.Context, bool) ([]core.Announcement, error)
	ListTasksForPlayer(context.Context, string) ([]core.PlayerTask, error)
	ClaimTaskReward(context.Context, string, string) (core.UserState, error)
	ListTaskDefinitions(context.Context) ([]core.TaskDefinition, error)
	SaveTaskDefinition(context.Context, core.TaskDefinition) error
	ActivateTaskDefinition(context.Context, string) error
	DeactivateTaskDefinition(context.Context, string) error
	DuplicateTaskDefinition(context.Context, string, string) (*core.TaskDefinition, error)
	ArchiveExpiredTaskCycles(context.Context, time.Time) ([]core.TaskCycleArchive, error)
	ListTaskCycleArchives(context.Context, string) ([]core.TaskCycleArchive, error)
	GetTaskCycleResults(context.Context, string, string) (core.TaskCycleResultsView, error)
	ListShopCatalogItemsForPlayer(context.Context, string) ([]core.ShopCatalogItemView, error)
	PurchaseShopItem(context.Context, string, string) (core.UserState, error)
	EquipShopItem(context.Context, string, string) (core.UserState, error)
	UnequipShopItem(context.Context, string) (core.UserState, error)
	ListShopItems(context.Context) ([]core.ShopItem, error)
	SaveShopItem(context.Context, core.ShopItem) error
	DeleteShopItem(context.Context, string) error
	SaveAnnouncement(context.Context, core.AnnouncementUpsert) (*core.Announcement, error)
	DeleteAnnouncement(context.Context, string) error
	CreateMessage(context.Context, string, string) (*core.Message, error)
	ListMessages(context.Context, string, int64) (core.MessagePage, error)
	DeleteMessage(context.Context, string) error
	// 天赋系统
	GetTalentState(context.Context, string) (*core.TalentState, error)
	UpgradeTalent(context.Context, string, string, int) error
	ResetTalents(context.Context, string) error
	ComputeTalentModifiers(context.Context, string) (*core.TalentModifiers, error)
}

// StateView 为只读聚合提供公共态、个人态和完整态读取能力。
type StateView interface {
	GetState(context.Context, string) (core.State, error)
	GetSnapshot(context.Context) (core.Snapshot, error)
	GetUserState(context.Context, string) (core.UserState, error)
	ListRooms(context.Context, string) (core.RoomList, error)
}

// RealtimeHub 为 SSE 与 WebSocket 共享的订阅中心。
type RealtimeHub interface {
	Subscribe(string) (<-chan events.ServerEvent, func())
	SubscriberCount() int
}

// OSSSigner 负责生成 OSS 直传短时凭证。
type OSSSigner interface {
	CreatePolicy(context.Context, string) (ossupload.Policy, error)
}

// Broadcaster 保留旧接口，避免现有调用点和测试同时变更。
type Broadcaster interface {
	BroadcastSnapshot(core.Snapshot) error
}

// ChangePublisher 负责将业务变更发布到实时层。
type ChangePublisher interface {
	PublishChange(context.Context, core.StateChange) error
}

// AdminBossHistoryReader 负责后台 Boss 历史分页读取，可选替换 Redis 默认读源。
type AdminBossHistoryReader interface {
	ListAdminBossHistoryPage(context.Context, int64, int64) (core.AdminBossHistoryPage, error)
}

// MessageStore 负责留言墙读写，可选替换 Redis 默认实现。
type MessageStore interface {
	CreateMessage(context.Context, string, string) (*core.Message, error)
	ListMessages(context.Context, string, int64) (core.MessagePage, error)
	DeleteMessage(context.Context, string) error
}

// AdminAuditWriter 负责后台审计日志写入。
type AdminAuditWriter interface {
	WriteAdminAuditLog(context.Context, core.AdminAuditLog) error
}

// DomainEventWriter 负责业务事件写入。
type DomainEventWriter interface {
	WriteDomainEvent(context.Context, core.DomainEvent) error
}

// ClickGuard 点击频率限制接口。
type ClickGuard interface {
	Allow(string) (time.Duration, error)
	ListBlacklist() []core.BlacklistEntry
	Unblock(string) bool
}

// AfkController 负责离页挂机状态流转与结算。
type AfkController interface {
	ReportPresence(context.Context, string, bool) error
	ConsumeSettlement(string) core.AfkSettlement
	Close() error
}

// AutoClickStatus 保留兼容类型，旧接口已下线。
type AutoClickStatus struct {
	Active        bool   `json:"active"`
	ButtonKey     string `json:"buttonKey,omitempty"`
	StartedAt     int64  `json:"startedAt,omitempty"`
	UpdatedAt     int64  `json:"updatedAt,omitempty"`
	IntervalMs    int64  `json:"intervalMs"`
	RatePerSecond int    `json:"ratePerSecond"`
}

// AutoClickController 保留兼容接口，旧接口已下线。
type AutoClickController interface {
	Start(context.Context, string, string) (AutoClickStatus, error)
	Stop(string) AutoClickStatus
	Status(string) AutoClickStatus
	Close() error
}

// PlayerAuthenticator 负责玩家昵称密码登录、JWT 校验和后台密码重置。
type PlayerAuthenticator interface {
	Login(context.Context, string, string) (string, string, error)
	Verify(context.Context, string) (string, error)
	ResetPassword(context.Context, string, string) error
}

// EquipmentDraftGenerator 根据自然语言生成装备草稿，不负责持久化。
type EquipmentDraftGenerator interface {
	GenerateEquipmentDraft(context.Context, string) (core.EquipmentDefinition, error)
}

// EquipmentDraftFailureWriter 负责记录装备草稿生成失败上下文。
type EquipmentDraftFailureWriter interface {
	WriteEquipmentDraftFailure(context.Context, core.EquipmentDraftFailureLog) error
}

// Options 路由配置，注入业务逻辑、实时更新和静态资源。
type Options struct {
	Store                       ButtonStore
	StateView                   StateView
	Broadcaster                 Broadcaster
	ChangePublisher             ChangePublisher
	ClickGuard                  ClickGuard
	RateLimitNicknameWhitelist  []string
	AutoClick                   AutoClickController
	Afk                         AfkController
	PlayerAuthenticator         PlayerAuthenticator
	Events                      app.HandlerFunc
	RealtimeHub                 RealtimeHub
	AdminAuthenticator          *adminauth.Authenticator
	OSSSigner                   OSSSigner
	EquipmentDraftGenerator     EquipmentDraftGenerator
	EquipmentDraftFailureWriter EquipmentDraftFailureWriter
	AdminBossHistoryReader      AdminBossHistoryReader
	MessageStore                MessageStore
	AdminAuditWriter            AdminAuditWriter
	DomainEventWriter           DomainEventWriter
	AccessLogQueue              *archive.AsyncQueue[xlog.AccessLogEntry]
}

const adminSessionCookieName = "long_admin_session"
const playerSessionCookieName = "long_player_session"

func registerRoutes(engine *route.Engine, options Options) {
	stateView := options.StateView
	if stateView == nil {
		if candidate, ok := options.Store.(StateView); ok {
			stateView = candidate
		}
	}

	registerPublicRoutes(engine, options, stateView)
	registerPlayerAuthRoutes(engine, options)
	registerPlayerActionRoutes(engine, options)
	registerTalentRoutes(engine, options)
	registerAdminRoutes(engine, options)
	registerRealtimeRoutes(engine, options)
	registerStaticRoutes(engine, options)
}
