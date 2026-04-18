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
	"long/internal/ratelimit"
	"long/internal/vote"
)

// ButtonStore 投票存储接口，HTTP 层所需的最小契约
type ButtonStore interface {
	GetState(context.Context, string) (vote.State, error)
	GetSnapshot(context.Context) (vote.Snapshot, error)
	ClickButton(context.Context, string, string) (vote.ClickResult, error)
	ValidateNickname(context.Context, string) error
	EquipItem(context.Context, string, string) (vote.State, error)
	UnequipItem(context.Context, string, string) (vote.State, error)
	GetAdminState(context.Context) (vote.AdminState, error)
	SaveButton(context.Context, vote.ButtonUpsert) error
	SaveEquipmentDefinition(context.Context, vote.EquipmentDefinition) error
	DeleteEquipmentDefinition(context.Context, string) error
	ActivateBoss(context.Context, vote.BossUpsert) (*vote.Boss, error)
	DeactivateBoss(context.Context) error
	SetBossLoot(context.Context, string, []vote.BossLootEntry) error
	ListBossHistory(context.Context) ([]vote.BossHistoryEntry, error)
}

// Broadcaster 广播接口，点击成功后推送更新快照
type Broadcaster interface {
	BroadcastSnapshot(vote.Snapshot) error
}

// ClickGuard 点击频率限制接口
type ClickGuard interface {
	Allow(string) (time.Duration, error)
}

// Options 路由配置，注入业务逻辑、实时更新和静态资源
type Options struct {
	Store              ButtonStore
	Broadcaster        Broadcaster
	ClickGuard         ClickGuard
	Events             http.Handler
	PublicDir          string
	AdminAuthenticator *adminauth.Authenticator
}

const adminSessionCookieName = "long_admin_session"

// NewHandler 构建 API 路由和 SPA 静态文件回退处理器
func NewHandler(options Options) http.Handler {
	apiMux := http.NewServeMux()

	apiMux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})

	apiMux.HandleFunc("GET /api/buttons", func(w http.ResponseWriter, r *http.Request) {
		state, err := options.Store.GetState(r.Context(), r.URL.Query().Get("nickname"))
		if err != nil {
			if writeNicknameError(w, err) {
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "STATE_FETCH_FAILED"})
			return
		}

		writeJSON(w, http.StatusOK, state)
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

		state, err := options.Store.GetState(r.Context(), body.Nickname)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "STATE_FETCH_FAILED"})
			return
		}

		_ = options.Broadcaster.BroadcastSnapshot(vote.Snapshot{
			Buttons:     state.Buttons,
			Leaderboard: state.Leaderboard,
		})
		writeJSON(w, http.StatusOK, map[string]any{
			"button":          result.Button,
			"buttons":         state.Buttons,
			"leaderboard":     state.Leaderboard,
			"userStats":       result.UserStats,
			"delta":           result.Delta,
			"critical":        result.Critical,
			"boss":            state.Boss,
			"bossLeaderboard": state.BossLeaderboard,
			"myBossStats":     state.MyBossStats,
			"inventory":       state.Inventory,
			"loadout":         state.Loadout,
			"combatStats":     state.CombatStats,
			"lastReward":      state.LastReward,
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

			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
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
