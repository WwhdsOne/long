package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	adminauth "long/internal/admin"
	ossupload "long/internal/oss"
	"long/internal/ratelimit"
	"long/internal/vote"
)

// ButtonStore 投票存储接口，HTTP 层所需的最小契约
type ButtonStore interface {
	GetState(context.Context, string) (vote.State, error)
	GetSnapshot(context.Context) (vote.Snapshot, error)
	GetUserState(context.Context, string) (vote.UserState, error)
	ClickButton(context.Context, string, string) (vote.ClickResult, error)
	ValidateNickname(context.Context, string) error
	EquipItem(context.Context, string, string) (vote.State, error)
	UnequipItem(context.Context, string, string) (vote.State, error)
	SynthesizeItem(context.Context, string, string) (vote.State, error)
	EquipHero(context.Context, string, string) (vote.State, error)
	UnequipHero(context.Context, string, string) (vote.State, error)
	GetAdminState(context.Context) (vote.AdminState, error)
	ListAdminPlayers(context.Context, string, int64) (vote.AdminPlayerPage, error)
	GetAdminPlayer(context.Context, string) (*vote.AdminPlayerOverview, error)
	SaveButton(context.Context, vote.ButtonUpsert) error
	SaveEquipmentDefinition(context.Context, vote.EquipmentDefinition) error
	SaveHeroDefinition(context.Context, vote.HeroDefinition) error
	DeleteEquipmentDefinition(context.Context, string) error
	DeleteHeroDefinition(context.Context, string) error
	ActivateBoss(context.Context, vote.BossUpsert) (*vote.Boss, error)
	DeactivateBoss(context.Context) error
	SetBossLoot(context.Context, string, []vote.BossLootEntry) error
	SaveBossTemplate(context.Context, vote.BossTemplateUpsert) error
	DeleteBossTemplate(context.Context, string) error
	SetBossTemplateLoot(context.Context, string, []vote.BossLootEntry) error
	SetBossTemplateHeroLoot(context.Context, string, []vote.BossHeroLootEntry) error
	SetBossCycleEnabled(context.Context, bool) (*vote.Boss, error)
	ListBossHistory(context.Context) ([]vote.BossHistoryEntry, error)
	GetLatestAnnouncement(context.Context) (*vote.Announcement, error)
	ListAnnouncements(context.Context, bool) ([]vote.Announcement, error)
	SaveAnnouncement(context.Context, vote.AnnouncementUpsert) (*vote.Announcement, error)
	DeleteAnnouncement(context.Context, string) error
	CreateMessage(context.Context, string, string) (*vote.Message, error)
	ListMessages(context.Context, string, int64) (vote.MessagePage, error)
	DeleteMessage(context.Context, string) error
}

// StateView 为只读聚合提供公共态、个人态和完整态读取能力。
type StateView interface {
	GetState(context.Context, string) (vote.State, error)
	GetSnapshot(context.Context) (vote.Snapshot, error)
	GetUserState(context.Context, string) (vote.UserState, error)
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

// ClickGuard 点击频率限制接口
type ClickGuard interface {
	Allow(string) (time.Duration, error)
}

// Options 路由配置，注入业务逻辑、实时更新和静态资源
type Options struct {
	Store              ButtonStore
	StateView          StateView
	Broadcaster        Broadcaster
	ChangePublisher    ChangePublisher
	ClickGuard         ClickGuard
	Events             http.Handler
	PublicDir          string
	AdminAuthenticator *adminauth.Authenticator
	OSSSigner          OSSSigner
}

const adminSessionCookieName = "long_admin_session"

// NewHandler 构建 API 路由和 SPA 静态文件回退处理器
func NewHandler(options Options) http.Handler {
	stateView := options.StateView
	if stateView == nil {
		stateView = options.Store
	}

	apiMux := http.NewServeMux()

	apiMux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})

	apiMux.HandleFunc("GET /api/buttons", func(w http.ResponseWriter, r *http.Request) {
		state, err := stateView.GetState(r.Context(), r.URL.Query().Get("nickname"))
		if err != nil {
			if writeNicknameError(w, err) {
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "STATE_FETCH_FAILED"})
			return
		}

		writeJSON(w, http.StatusOK, state)
	})

	apiMux.HandleFunc("GET /api/boss/history", func(w http.ResponseWriter, r *http.Request) {
		history, err := options.Store.ListBossHistory(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "BOSS_HISTORY_FAILED"})
			return
		}

		writeJSON(w, http.StatusOK, history)
	})

	apiMux.HandleFunc("GET /api/announcements/latest", func(w http.ResponseWriter, r *http.Request) {
		item, err := options.Store.GetLatestAnnouncement(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "ANNOUNCEMENT_FETCH_FAILED"})
			return
		}
		writeJSON(w, http.StatusOK, item)
	})

	apiMux.HandleFunc("GET /api/announcements", func(w http.ResponseWriter, r *http.Request) {
		items, err := options.Store.ListAnnouncements(r.Context(), false)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "ANNOUNCEMENT_LIST_FAILED"})
			return
		}
		writeJSON(w, http.StatusOK, items)
	})

	apiMux.HandleFunc("GET /api/messages", func(w http.ResponseWriter, r *http.Request) {
		page, err := options.Store.ListMessages(r.Context(), r.URL.Query().Get("cursor"), 50)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "MESSAGE_LIST_FAILED"})
			return
		}
		writeJSON(w, http.StatusOK, page)
	})

	apiMux.HandleFunc("POST /api/messages", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Nickname string `json:"nickname"`
			Content  string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "INVALID_REQUEST"})
			return
		}

		message, err := options.Store.CreateMessage(r.Context(), body.Nickname, body.Content)
		if err != nil {
			if writeNicknameError(w, err) || writeContentError(w, err) {
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "MESSAGE_CREATE_FAILED"})
			return
		}

		publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
			Type:      vote.StateChangeMessageCreated,
			Nickname:  strings.TrimSpace(body.Nickname),
			Timestamp: time.Now().Unix(),
		})
		writeJSON(w, http.StatusOK, message)
	})

	apiMux.HandleFunc("POST /api/nickname/validate", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Nickname string `json:"nickname"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "INVALID_REQUEST",
				"message": "昵称没有带上，先报个名再试试。",
			})
			return
		}

		if err := options.Store.ValidateNickname(r.Context(), body.Nickname); err != nil {
			if writeNicknameError(w, err) {
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "NICKNAME_VALIDATE_FAILED"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})

	apiMux.HandleFunc("POST /api/buttons/{slug}/click", func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		var body struct {
			Nickname string `json:"nickname"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "INVALID_REQUEST",
				"message": "昵称没有带上，先报个名再开点。",
			})
			return
		}
		if strings.TrimSpace(body.Nickname) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "INVALID_NICKNAME",
				"message": "昵称还没填好，先起个名字再点。",
			})
			return
		}

		if options.ClickGuard != nil {
			retryAfter, err := options.ClickGuard.Allow(clientIdentifier(r))
			if err != nil {
				if errors.Is(err, ratelimit.ErrTooManyRequests) {
					w.Header().Set("Retry-After", strconv.FormatInt(int64(retryAfter/time.Second), 10))
					writeJSON(w, http.StatusTooManyRequests, map[string]string{
						"error":   "TOO_MANY_REQUESTS",
						"message": "点得太快了，先歇 10 分钟再来。",
					})
					return
				}

				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "RATE_LIMIT_FAILED"})
				return
			}
		}

		result, err := options.Store.ClickButton(r.Context(), slug, body.Nickname)
		if err != nil {
			if errors.Is(err, vote.ErrButtonNotFound) {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "BUTTON_NOT_FOUND"})
				return
			}
			if writeNicknameError(w, err) {
				return
			}

			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "INCREMENT_FAILED"})
			return
		}

		change := vote.StateChange{
			Type:      vote.StateChangeButtonClicked,
			Nickname:  strings.TrimSpace(body.Nickname),
			Timestamp: time.Now().Unix(),
		}
		if result.BroadcastUserAll {
			change.BroadcastUserAll = true
		}
		publishChange(r.Context(), options.ChangePublisher, change)
		writeJSON(w, http.StatusOK, map[string]any{
			"button":          result.Button,
			"userStats":       result.UserStats,
			"delta":           result.Delta,
			"critical":        result.Critical,
			"boss":            result.Boss,
			"bossLeaderboard": result.BossLeaderboard,
			"myBossStats":     result.MyBossStats,
			"recentRewards":   result.RecentRewards,
			"lastReward":      result.LastReward,
		})
	})

	apiMux.HandleFunc("POST /api/equipment/{itemId}/equip", func(w http.ResponseWriter, r *http.Request) {
		itemID := r.PathValue("itemId")
		var body struct {
			Nickname string `json:"nickname"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "INVALID_REQUEST",
				"message": "昵称没有带上，先报个名再穿装备。",
			})
			return
		}

		state, err := options.Store.EquipItem(r.Context(), body.Nickname, itemID)
		if err != nil {
			if writeNicknameError(w, err) {
				return
			}
			if errors.Is(err, vote.ErrEquipmentNotFound) {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "EQUIPMENT_NOT_FOUND"})
				return
			}
			if errors.Is(err, vote.ErrEquipmentNotOwned) {
				writeJSON(w, http.StatusBadRequest, map[string]string{
					"error":   "EQUIPMENT_NOT_OWNED",
					"message": "这件装备还不在你的背包里。",
				})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "EQUIP_FAILED"})
			return
		}

		publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
			Type:      vote.StateChangeEquipmentChanged,
			Nickname:  strings.TrimSpace(body.Nickname),
			Timestamp: time.Now().Unix(),
		})
		writeJSON(w, http.StatusOK, state)
	})

	apiMux.HandleFunc("POST /api/equipment/{itemId}/unequip", func(w http.ResponseWriter, r *http.Request) {
		itemID := r.PathValue("itemId")
		var body struct {
			Nickname string `json:"nickname"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "INVALID_REQUEST",
				"message": "昵称没有带上，先报个名再卸装备。",
			})
			return
		}

		state, err := options.Store.UnequipItem(r.Context(), body.Nickname, itemID)
		if err != nil {
			if writeNicknameError(w, err) {
				return
			}
			if errors.Is(err, vote.ErrEquipmentNotFound) {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "EQUIPMENT_NOT_FOUND"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "UNEQUIP_FAILED"})
			return
		}

		publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
			Type:      vote.StateChangeEquipmentChanged,
			Nickname:  strings.TrimSpace(body.Nickname),
			Timestamp: time.Now().Unix(),
		})
		writeJSON(w, http.StatusOK, state)
	})

	apiMux.HandleFunc("POST /api/equipment/{itemId}/synthesize", func(w http.ResponseWriter, r *http.Request) {
		itemID := r.PathValue("itemId")
		var body struct {
			Nickname string `json:"nickname"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "INVALID_REQUEST",
				"message": "昵称没有带上，先报个名再升星。",
			})
			return
		}

		state, err := options.Store.SynthesizeItem(r.Context(), body.Nickname, itemID)
		if err != nil {
			if writeNicknameError(w, err) {
				return
			}
			switch {
			case errors.Is(err, vote.ErrEquipmentNotFound):
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "EQUIPMENT_NOT_FOUND"})
			case errors.Is(err, vote.ErrEquipmentNotOwned), errors.Is(err, vote.ErrEquipmentNotEnough):
				writeJSON(w, http.StatusBadRequest, map[string]string{
					"error":   "EQUIPMENT_NOT_ENOUGH",
					"message": "至少要有 3 件同名装备才能升星。",
				})
			case errors.Is(err, vote.ErrEquipmentMaxStar):
				writeJSON(w, http.StatusBadRequest, map[string]string{
					"error":   "EQUIPMENT_MAX_STAR",
					"message": "这件装备已经满星了。",
				})
			default:
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "SYNTHESIZE_FAILED"})
			}
			return
		}

		publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
			Type:      vote.StateChangeEquipmentChanged,
			Nickname:  strings.TrimSpace(body.Nickname),
			Timestamp: time.Now().Unix(),
		})
		writeJSON(w, http.StatusOK, state)
	})

	apiMux.HandleFunc("POST /api/heroes/{heroId}/equip", func(w http.ResponseWriter, r *http.Request) {
		heroID := r.PathValue("heroId")
		var body struct {
			Nickname string `json:"nickname"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "INVALID_REQUEST",
				"message": "昵称没有带上，先报个名再派出英雄。",
			})
			return
		}

		state, err := options.Store.EquipHero(r.Context(), body.Nickname, heroID)
		if err != nil {
			if writeNicknameError(w, err) {
				return
			}
			switch {
			case errors.Is(err, vote.ErrHeroNotFound):
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "HERO_NOT_FOUND"})
			case errors.Is(err, vote.ErrHeroNotOwned):
				writeJSON(w, http.StatusBadRequest, map[string]string{
					"error":   "HERO_NOT_OWNED",
					"message": "这位小小英雄还没加入你的队伍。",
				})
			default:
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "HERO_EQUIP_FAILED"})
			}
			return
		}

		publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
			Type:      vote.StateChangeEquipmentChanged,
			Nickname:  strings.TrimSpace(body.Nickname),
			Timestamp: time.Now().Unix(),
		})
		writeJSON(w, http.StatusOK, state)
	})

	apiMux.HandleFunc("POST /api/heroes/{heroId}/unequip", func(w http.ResponseWriter, r *http.Request) {
		heroID := r.PathValue("heroId")
		var body struct {
			Nickname string `json:"nickname"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "INVALID_REQUEST",
				"message": "昵称没有带上，先报个名再收回英雄。",
			})
			return
		}

		state, err := options.Store.UnequipHero(r.Context(), body.Nickname, heroID)
		if err != nil {
			if writeNicknameError(w, err) {
				return
			}
			switch {
			case errors.Is(err, vote.ErrHeroNotFound):
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "HERO_NOT_FOUND"})
			default:
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "HERO_UNEQUIP_FAILED"})
			}
			return
		}

		publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
			Type:      vote.StateChangeEquipmentChanged,
			Nickname:  strings.TrimSpace(body.Nickname),
			Timestamp: time.Now().Unix(),
		})
		writeJSON(w, http.StatusOK, state)
	})

	if options.AdminAuthenticator != nil {
		apiMux.HandleFunc("POST /api/admin/login", func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "INVALID_REQUEST"})
				return
			}

			token, ok := options.AdminAuthenticator.Login(body.Username, body.Password)
			if !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{
					"error":   "INVALID_CREDENTIALS",
					"message": "账号或口令不对。",
				})
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     adminSessionCookieName,
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		})

		apiMux.HandleFunc("POST /api/admin/logout", func(w http.ResponseWriter, _ *http.Request) {
			http.SetCookie(w, &http.Cookie{
				Name:     adminSessionCookieName,
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				MaxAge:   -1,
				SameSite: http.SameSiteLaxMode,
			})
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		})

		apiMux.HandleFunc("GET /api/admin/session", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			writeJSON(w, http.StatusOK, map[string]bool{"authenticated": true})
		})

		apiMux.HandleFunc("GET /api/admin/state", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			state, err := options.Store.GetAdminState(r.Context())
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "ADMIN_STATE_FAILED"})
				return
			}

			writeJSON(w, http.StatusOK, state)
		})

		apiMux.HandleFunc("GET /api/admin/players", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			limit := int64(50)
			if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
				parsedLimit, err := strconv.ParseInt(rawLimit, 10, 64)
				if err != nil {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "INVALID_LIMIT"})
					return
				}
				limit = parsedLimit
			}

			page, err := options.Store.ListAdminPlayers(r.Context(), r.URL.Query().Get("cursor"), limit)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "ADMIN_PLAYERS_FAILED"})
				return
			}

			writeJSON(w, http.StatusOK, page)
		})

		apiMux.HandleFunc("GET /api/admin/players/{nickname}", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			player, err := options.Store.GetAdminPlayer(r.Context(), r.PathValue("nickname"))
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "ADMIN_PLAYER_FAILED"})
				return
			}
			if player == nil {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "PLAYER_NOT_FOUND"})
				return
			}

			writeJSON(w, http.StatusOK, player)
		})

		apiMux.HandleFunc("POST /api/admin/boss/activate", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			var body vote.BossUpsert
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "INVALID_REQUEST"})
				return
			}

			boss, err := options.Store.ActivateBoss(r.Context(), body)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "BOSS_ACTIVATE_FAILED"})
				return
			}

			publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
				Type:             vote.StateChangeBossChanged,
				BroadcastUserAll: true,
				Timestamp:        time.Now().Unix(),
			})
			writeJSON(w, http.StatusOK, boss)
		})

		apiMux.HandleFunc("POST /api/admin/boss/deactivate", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			if err := options.Store.DeactivateBoss(r.Context()); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "BOSS_DEACTIVATE_FAILED"})
				return
			}

			publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
				Type:             vote.StateChangeBossChanged,
				BroadcastUserAll: true,
				Timestamp:        time.Now().Unix(),
			})
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		})

		apiMux.HandleFunc("PUT /api/admin/boss/loot", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			var body struct {
				BossID string               `json:"bossId"`
				Loot   []vote.BossLootEntry `json:"loot"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "INVALID_REQUEST"})
				return
			}

			if err := options.Store.SetBossLoot(r.Context(), body.BossID, body.Loot); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "BOSS_LOOT_FAILED"})
				return
			}

			publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
				Type:             vote.StateChangeBossChanged,
				BroadcastUserAll: true,
				Timestamp:        time.Now().Unix(),
			})
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		})

		apiMux.HandleFunc("POST /api/admin/boss/pool", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			var body vote.BossTemplateUpsert
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "INVALID_REQUEST"})
				return
			}

			if err := options.Store.SaveBossTemplate(r.Context(), body); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "BOSS_POOL_SAVE_FAILED"})
				return
			}

			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		})

		apiMux.HandleFunc("PUT /api/admin/boss/pool/{templateId}", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			var body vote.BossTemplateUpsert
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "INVALID_REQUEST"})
				return
			}
			if templateID := strings.TrimSpace(r.PathValue("templateId")); templateID != "" {
				body.ID = templateID
			}

			if err := options.Store.SaveBossTemplate(r.Context(), body); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "BOSS_POOL_SAVE_FAILED"})
				return
			}

			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		})

		apiMux.HandleFunc("DELETE /api/admin/boss/pool/{templateId}", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			if err := options.Store.DeleteBossTemplate(r.Context(), r.PathValue("templateId")); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "BOSS_POOL_DELETE_FAILED"})
				return
			}

			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		})

		apiMux.HandleFunc("PUT /api/admin/boss/pool/{templateId}/loot", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			var body struct {
				Loot []vote.BossLootEntry `json:"loot"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "INVALID_REQUEST"})
				return
			}

			if err := options.Store.SetBossTemplateLoot(r.Context(), r.PathValue("templateId"), body.Loot); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "BOSS_POOL_LOOT_FAILED"})
				return
			}

			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		})

		apiMux.HandleFunc("PUT /api/admin/boss/pool/{templateId}/hero-loot", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			var body struct {
				Loot []vote.BossHeroLootEntry `json:"loot"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "INVALID_REQUEST"})
				return
			}

			if err := options.Store.SetBossTemplateHeroLoot(r.Context(), r.PathValue("templateId"), body.Loot); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "BOSS_POOL_HERO_LOOT_FAILED"})
				return
			}

			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		})

		apiMux.HandleFunc("POST /api/admin/boss/cycle/enable", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			boss, err := options.Store.SetBossCycleEnabled(r.Context(), true)
			if err != nil {
				if errors.Is(err, vote.ErrBossPoolEmpty) {
					writeJSON(w, http.StatusBadRequest, map[string]string{
						"error":   "BOSS_POOL_EMPTY",
						"message": "Boss 池还是空的，先加模板再开启循环。",
					})
					return
				}
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "BOSS_CYCLE_ENABLE_FAILED"})
				return
			}

			publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
				Type:             vote.StateChangeBossChanged,
				BroadcastUserAll: true,
				Timestamp:        time.Now().Unix(),
			})
			writeJSON(w, http.StatusOK, boss)
		})

		apiMux.HandleFunc("POST /api/admin/boss/cycle/disable", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			boss, err := options.Store.SetBossCycleEnabled(r.Context(), false)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "BOSS_CYCLE_DISABLE_FAILED"})
				return
			}

			publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
				Type:             vote.StateChangeBossChanged,
				BroadcastUserAll: true,
				Timestamp:        time.Now().Unix(),
			})
			writeJSON(w, http.StatusOK, boss)
		})

		apiMux.HandleFunc("GET /api/admin/boss/history", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			history, err := options.Store.ListBossHistory(r.Context())
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "BOSS_HISTORY_FAILED"})
				return
			}

			writeJSON(w, http.StatusOK, history)
		})

		apiMux.HandleFunc("GET /api/admin/announcements", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			items, err := options.Store.ListAnnouncements(r.Context(), true)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "ANNOUNCEMENT_LIST_FAILED"})
				return
			}
			writeJSON(w, http.StatusOK, items)
		})

		apiMux.HandleFunc("POST /api/admin/announcements", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			var body vote.AnnouncementUpsert
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "INVALID_REQUEST"})
				return
			}

			item, err := options.Store.SaveAnnouncement(r.Context(), body)
			if err != nil {
				if writeContentError(w, err) {
					return
				}
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "ANNOUNCEMENT_SAVE_FAILED"})
				return
			}
			publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
				Type:      vote.StateChangeAnnouncementChanged,
				Timestamp: time.Now().Unix(),
			})
			writeJSON(w, http.StatusOK, item)
		})

		apiMux.HandleFunc("DELETE /api/admin/announcements/{id}", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			if err := options.Store.DeleteAnnouncement(r.Context(), r.PathValue("id")); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "ANNOUNCEMENT_DELETE_FAILED"})
				return
			}
			publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
				Type:      vote.StateChangeAnnouncementChanged,
				Timestamp: time.Now().Unix(),
			})
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		})

		apiMux.HandleFunc("GET /api/admin/messages", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			page, err := options.Store.ListMessages(r.Context(), r.URL.Query().Get("cursor"), 50)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "MESSAGE_LIST_FAILED"})
				return
			}
			writeJSON(w, http.StatusOK, page)
		})

		apiMux.HandleFunc("DELETE /api/admin/messages/{id}", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			if err := options.Store.DeleteMessage(r.Context(), r.PathValue("id")); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "MESSAGE_DELETE_FAILED"})
				return
			}
			publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
				Type:      vote.StateChangeMessageDeleted,
				Timestamp: time.Now().Unix(),
			})
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		})

		apiMux.HandleFunc("POST /api/admin/oss/sts", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}
			if options.OSSSigner == nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{
					"error":   "OSS_NOT_CONFIGURED",
					"message": "OSS 直传还没配置，先手动填图片 URL。",
				})
				return
			}

			policy, err := options.OSSSigner.CreatePolicy(r.Context())
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "OSS_POLICY_FAILED"})
				return
			}
			writeJSON(w, http.StatusOK, policy)
		})

		apiMux.HandleFunc("POST /api/admin/buttons", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			var body vote.ButtonUpsert
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "INVALID_REQUEST"})
				return
			}

			if err := options.Store.SaveButton(r.Context(), body); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "BUTTON_SAVE_FAILED"})
				return
			}

			publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
				Type:      vote.StateChangeButtonMetaChanged,
				Timestamp: time.Now().Unix(),
			})
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		})

		apiMux.HandleFunc("PUT /api/admin/buttons/{slug}", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			var body vote.ButtonUpsert
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "INVALID_REQUEST"})
				return
			}
			body.Slug = r.PathValue("slug")

			if err := options.Store.SaveButton(r.Context(), body); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "BUTTON_SAVE_FAILED"})
				return
			}

			publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
				Type:      vote.StateChangeButtonMetaChanged,
				Timestamp: time.Now().Unix(),
			})
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		})

		apiMux.HandleFunc("POST /api/admin/equipment", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			var body vote.EquipmentDefinition
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "INVALID_REQUEST"})
				return
			}

			if err := options.Store.SaveEquipmentDefinition(r.Context(), body); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "EQUIPMENT_SAVE_FAILED"})
				return
			}

			publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
				Type:             vote.StateChangeEquipmentMetaChanged,
				BroadcastUserAll: true,
				Timestamp:        time.Now().Unix(),
			})
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		})

		apiMux.HandleFunc("POST /api/admin/heroes", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			var body vote.HeroDefinition
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "INVALID_REQUEST"})
				return
			}

			if err := options.Store.SaveHeroDefinition(r.Context(), body); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "HERO_SAVE_FAILED"})
				return
			}

			publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
				Type:             vote.StateChangeEquipmentMetaChanged,
				BroadcastUserAll: true,
				Timestamp:        time.Now().Unix(),
			})
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		})

		apiMux.HandleFunc("PUT /api/admin/heroes/{heroId}", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			var body vote.HeroDefinition
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "INVALID_REQUEST"})
				return
			}
			body.HeroID = r.PathValue("heroId")

			if err := options.Store.SaveHeroDefinition(r.Context(), body); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "HERO_SAVE_FAILED"})
				return
			}

			publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
				Type:             vote.StateChangeEquipmentMetaChanged,
				BroadcastUserAll: true,
				Timestamp:        time.Now().Unix(),
			})
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		})

		apiMux.HandleFunc("DELETE /api/admin/heroes/{heroId}", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			if err := options.Store.DeleteHeroDefinition(r.Context(), r.PathValue("heroId")); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "HERO_DELETE_FAILED"})
				return
			}

			publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
				Type:             vote.StateChangeEquipmentMetaChanged,
				BroadcastUserAll: true,
				Timestamp:        time.Now().Unix(),
			})
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		})

		apiMux.HandleFunc("PUT /api/admin/equipment/{itemId}", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			var body vote.EquipmentDefinition
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "INVALID_REQUEST"})
				return
			}
			body.ItemID = r.PathValue("itemId")

			if err := options.Store.SaveEquipmentDefinition(r.Context(), body); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "EQUIPMENT_SAVE_FAILED"})
				return
			}

			publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
				Type:             vote.StateChangeEquipmentMetaChanged,
				BroadcastUserAll: true,
				Timestamp:        time.Now().Unix(),
			})
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		})

		apiMux.HandleFunc("DELETE /api/admin/equipment/{itemId}", func(w http.ResponseWriter, r *http.Request) {
			if !isAdminAuthenticated(r, options.AdminAuthenticator) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
				return
			}

			if err := options.Store.DeleteEquipmentDefinition(r.Context(), r.PathValue("itemId")); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "EQUIPMENT_DELETE_FAILED"})
				return
			}

			publishChange(r.Context(), options.ChangePublisher, vote.StateChange{
				Type:             vote.StateChangeEquipmentMetaChanged,
				BroadcastUserAll: true,
				Timestamp:        time.Now().Unix(),
			})
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		})
	}

	if options.Events != nil {
		apiMux.Handle("GET /api/events", options.Events)
	}

	if options.PublicDir == "" {
		return apiMux
	}

	fileServer := http.FileServer(http.Dir(options.PublicDir))
	indexFile := filepath.Join(options.PublicDir, "index.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/api" {
			apiMux.ServeHTTP(w, r)
			return
		}

		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}

		cleanedPath := filepath.Clean("/" + strings.TrimPrefix(r.URL.Path, "/"))
		if cleanedPath == "/" {
			http.ServeFile(w, r, indexFile)
			return
		}

		target := filepath.Join(options.PublicDir, cleanedPath)
		if stat, err := os.Stat(target); err == nil && !stat.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}

		http.ServeFile(w, r, indexFile)
	})
}

// writeJSON keeps API responses consistent across handlers.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeNicknameError(w http.ResponseWriter, err error) bool {
	if errors.Is(err, vote.ErrInvalidNickname) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "INVALID_NICKNAME",
			"message": "昵称还没填好，先起个名字再点。",
		})
		return true
	}

	if errors.Is(err, vote.ErrSensitiveNickname) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "SENSITIVE_NICKNAME",
			"message": "昵称包含敏感词，请换一个试试。",
		})
		return true
	}

	return false
}

func writeContentError(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, vote.ErrSensitiveContent):
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "SENSITIVE_CONTENT",
			"message": "内容包含敏感词，请改一下再发。",
		})
		return true
	case errors.Is(err, vote.ErrMessageEmpty):
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "EMPTY_CONTENT",
			"message": "内容不能为空。",
		})
		return true
	case errors.Is(err, vote.ErrMessageTooLong):
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "CONTENT_TOO_LONG",
			"message": "内容最多 200 个字。",
		})
		return true
	}

	return false
}

// clientIdentifier extracts the best-effort real client address for rate limiting.
func clientIdentifier(request *http.Request) string {
	if forwardedFor := request.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		if len(parts) > 0 {
			candidate := strings.TrimSpace(parts[0])
			if candidate != "" {
				return candidate
			}
		}
	}

	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err == nil && host != "" {
		return host
	}

	return request.RemoteAddr
}

func isAdminAuthenticated(request *http.Request, authenticator *adminauth.Authenticator) bool {
	if authenticator == nil {
		return false
	}

	cookie, err := request.Cookie(adminSessionCookieName)
	if err != nil {
		return false
	}

	return authenticator.Verify(cookie.Value)
}

func publishChange(ctx context.Context, publisher ChangePublisher, change vote.StateChange) {
	if publisher == nil {
		return
	}
	_ = publisher.PublishChange(ctx, change)
}
