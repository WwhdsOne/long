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

	"long/internal/ratelimit"
	"long/internal/vote"
)

// ButtonStore is the minimal vote storage contract required by the HTTP layer.
type ButtonStore interface {
	ListButtons(context.Context) ([]vote.Button, error)
	ClickButton(context.Context, string) (vote.ClickResult, error)
}

// Broadcaster emits updated snapshots after a successful click.
type Broadcaster interface {
	BroadcastSnapshot([]vote.Button) error
}

// ClickGuard decides whether the current client may submit another click.
type ClickGuard interface {
	Allow(string) (time.Duration, error)
}

// Options wires business logic, realtime updates, and static assets into one router.
type Options struct {
	Store       ButtonStore
	Broadcaster Broadcaster
	ClickGuard  ClickGuard
	Events      http.Handler
	PublicDir   string
}

// NewHandler builds the API routes plus the SPA/static-file fallback handler.
func NewHandler(options Options) http.Handler {
	apiMux := http.NewServeMux()

	apiMux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})

	apiMux.HandleFunc("GET /api/buttons", func(w http.ResponseWriter, r *http.Request) {
		buttons, err := options.Store.ListButtons(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "BUTTONS_FETCH_FAILED"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"buttons": buttons})
	})

	apiMux.HandleFunc("POST /api/buttons/{slug}/click", func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")

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

		result, err := options.Store.ClickButton(r.Context(), slug)
		if err != nil {
			if errors.Is(err, vote.ErrButtonNotFound) {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "BUTTON_NOT_FOUND"})
				return
			}

			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "INCREMENT_FAILED"})
			return
		}

		buttons, err := options.Store.ListButtons(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "BUTTONS_FETCH_FAILED"})
			return
		}

		_ = options.Broadcaster.BroadcastSnapshot(buttons)
		writeJSON(w, http.StatusOK, map[string]any{
			"button":   result.Button,
			"buttons":  buttons,
			"delta":    result.Delta,
			"critical": result.Critical,
		})
	})

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
