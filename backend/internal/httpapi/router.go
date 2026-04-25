package httpapi

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"

	adminauth "long/internal/admin"
	"long/internal/events"
	ossupload "long/internal/oss"
	"long/internal/vote"
)

// ButtonStore 投票存储接口，HTTP 层所需的最小契约。
type ButtonStore interface {
	GetState(context.Context, string) (vote.State, error)
	GetSnapshot(context.Context) (vote.Snapshot, error)
	GetUserState(context.Context, string) (vote.UserState, error)
	ClickButton(context.Context, string, string) (vote.ClickResult, error)
	AutoClickBossPart(context.Context, string, string) (vote.ClickResult, error)
	ValidateNickname(context.Context, string) error
	EquipItem(context.Context, string, string) (vote.State, error)
	UnequipItem(context.Context, string, string) (vote.State, error)
	GetAdminState(context.Context) (vote.AdminState, error)
	ListAdminButtonsPage(context.Context, int64, int64) (vote.AdminButtonPage, error)
	ListAdminEquipmentPage(context.Context, int64, int64) (vote.AdminEquipmentPage, error)
	ListAdminBossHistoryPage(context.Context, int64, int64) (vote.AdminBossHistoryPage, error)
	ListAdminPlayers(context.Context, string, int64) (vote.AdminPlayerPage, error)
	GetAdminPlayer(context.Context, string) (*vote.AdminPlayerOverview, error)
	SaveButton(context.Context, vote.ButtonUpsert) error
	SaveEquipmentDefinition(context.Context, vote.EquipmentDefinition) error
	DeleteEquipmentDefinition(context.Context, string) error
	ActivateBoss(context.Context, vote.BossUpsert) (*vote.Boss, error)
	DeactivateBoss(context.Context) error
	SetBossLoot(context.Context, string, []vote.BossLootEntry) error
	SaveBossTemplate(context.Context, vote.BossTemplateUpsert) error
	DeleteBossTemplate(context.Context, string) error
	SetBossTemplateLoot(context.Context, string, []vote.BossLootEntry) error
	SetBossCycleEnabled(context.Context, bool) (*vote.Boss, error)
	ListBossHistory(context.Context) ([]vote.BossHistoryEntry, error)
	GetBossResources(context.Context) (vote.BossResources, error)
	GetLatestAnnouncement(context.Context) (*vote.Announcement, error)
	ListAnnouncements(context.Context, bool) ([]vote.Announcement, error)
	SaveAnnouncement(context.Context, vote.AnnouncementUpsert) (*vote.Announcement, error)
	DeleteAnnouncement(context.Context, string) error
	CreateMessage(context.Context, string, string) (*vote.Message, error)
	ListMessages(context.Context, string, int64) (vote.MessagePage, error)
	DeleteMessage(context.Context, string) error
	// 天赋系统
	SelectTalentTree(context.Context, string, vote.TalentTree, vote.TalentTree) error
	GetTalentState(context.Context, string) (*vote.TalentState, error)
	LearnTalent(context.Context, string, string) error
	ResetTalents(context.Context, string) error
	ComputeTalentModifiers(context.Context, string) (*vote.TalentModifiers, error)
}

// StateView 为只读聚合提供公共态、个人态和完整态读取能力。
type StateView interface {
	GetState(context.Context, string) (vote.State, error)
	GetSnapshot(context.Context) (vote.Snapshot, error)
	GetUserState(context.Context, string) (vote.UserState, error)
}

// RealtimeHub 为 SSE 与 WebSocket 共享的订阅中心。
type RealtimeHub interface {
	Subscribe(string) (<-chan events.ServerEvent, func())
	SubscriberCount() int
}

// OSSSigner 负责生成 OSS 直传短时凭证。
type OSSSigner interface {
	CreatePolicy(context.Context) (ossupload.Policy, error)
}

// Broadcaster 保留旧接口，避免现有调用点和测试同时变更。
type Broadcaster interface {
	BroadcastSnapshot(vote.Snapshot) error
}

// ChangePublisher 负责将业务变更发布到实时层。
type ChangePublisher interface {
	PublishChange(context.Context, vote.StateChange) error
}

// ClickGuard 点击频率限制接口。
type ClickGuard interface {
	Allow(string) (time.Duration, error)
}

// AutoClickController 负责服务端托管挂机任务的生命周期。
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
	GenerateEquipmentDraft(context.Context, string) (vote.EquipmentDefinition, error)
}

// Options 路由配置，注入业务逻辑、实时更新和静态资源。
type Options struct {
	Store                   ButtonStore
	StateView               StateView
	Broadcaster             Broadcaster
	ChangePublisher         ChangePublisher
	ClickGuard              ClickGuard
	AutoClick               AutoClickController
	PlayerAuthenticator     PlayerAuthenticator
	Events                  app.HandlerFunc
	RealtimeHub             RealtimeHub
	PublicDir               string
	AdminAuthenticator      *adminauth.Authenticator
	OSSSigner               OSSSigner
	EquipmentDraftGenerator EquipmentDraftGenerator
}

const adminSessionCookieName = "long_admin_session"
const playerSessionCookieName = "long_player_session"

func registerRoutes(engine *route.Engine, options Options) {
	stateView := options.StateView
	if stateView == nil {
		stateView = options.Store
	}

	registerPublicRoutes(engine, options, stateView)
	registerPlayerAuthRoutes(engine, options)
	registerPlayerActionRoutes(engine, options)
	registerTalentRoutes(engine, options)
	registerAdminRoutes(engine, options)
	registerRealtimeRoutes(engine, options)
	registerStaticRoutes(engine, options)
}
