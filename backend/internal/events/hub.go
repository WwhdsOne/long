package events

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"long/internal/vote"
)

const (
	publicStateEventName = "public_state"
	userStateEventName   = "user_state"
)

// StateReader 提供 SSE 初始状态所需的公共态与个人态读取能力。
type StateReader interface {
	GetSnapshot(context.Context) (vote.Snapshot, error)
	GetUserState(context.Context, string) (vote.UserState, error)
}

// ServerEvent 是发往浏览器的一条 SSE 事件。
type ServerEvent struct {
	Name    string
	Payload []byte
}

type subscriber struct {
	nickname string
	ch       chan ServerEvent
}

// Hub 按事件类型向浏览器广播公共态和个人态。
type Hub struct {
	mu      sync.RWMutex
	clients map[*subscriber]struct{}
}

// NewHub creates an in-memory broadcaster for browser subscribers.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*subscriber]struct{}),
	}
}

// Subscribe 注册一个订阅者；昵称为空表示只接收公共态。
func (h *Hub) Subscribe(nickname string) (<-chan ServerEvent, func()) {
	client := &subscriber{
		nickname: strings.TrimSpace(nickname),
		ch:       make(chan ServerEvent, 4),
	}

	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		if _, ok := h.clients[client]; ok {
			delete(h.clients, client)
			close(client.ch)
		}
		h.mu.Unlock()
	}

	return client.ch, unsubscribe
}

// BroadcastPublic 向所有订阅者广播公共态。
func (h *Hub) BroadcastPublic(snapshot vote.Snapshot) error {
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		deliverEvent(client.ch, ServerEvent{
			Name:    publicStateEventName,
			Payload: payload,
		})
	}

	return nil
}

// BroadcastUser 向指定昵称的订阅者广播个人态。
func (h *Hub) BroadcastUser(nickname string, state vote.UserState) error {
	normalizedNickname := strings.TrimSpace(nickname)
	if normalizedNickname == "" {
		return nil
	}

	payload, err := json.Marshal(state)
	if err != nil {
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.nickname != normalizedNickname {
			continue
		}
		deliverEvent(client.ch, ServerEvent{
			Name:    userStateEventName,
			Payload: payload,
		})
	}

	return nil
}

// ActiveNicknames 返回当前在线且带昵称的订阅者集合。
func (h *Hub) ActiveNicknames() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	seen := make(map[string]struct{}, len(h.clients))
	for client := range h.clients {
		if client.nickname == "" {
			continue
		}
		seen[client.nickname] = struct{}{}
	}

	nicknames := make([]string, 0, len(seen))
	for nickname := range seen {
		nicknames = append(nicknames, nickname)
	}
	return nicknames
}

// NewHandler exposes the SSE endpoint used by the frontend EventSource client.
func NewHandler(hub *Hub, reader StateReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		nickname := strings.TrimSpace(r.URL.Query().Get("nickname"))

		snapshot, err := reader.GetSnapshot(r.Context())
		if err != nil {
			http.Error(w, "STATE_FETCH_FAILED", http.StatusInternalServerError)
			return
		}
		if err := writeEvent(w, publicStateEventName, snapshot); err != nil {
			return
		}

		if nickname != "" {
			userState, err := reader.GetUserState(r.Context(), nickname)
			if err != nil {
				http.Error(w, "STATE_FETCH_FAILED", http.StatusInternalServerError)
				return
			}
			if err := writeEvent(w, userStateEventName, userState); err != nil {
				return
			}
		}
		flusher.Flush()

		client, unsubscribe := hub.Subscribe(nickname)
		defer unsubscribe()

		heartbeat := time.NewTicker(25 * time.Second)
		defer heartbeat.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case event, ok := <-client:
				if !ok {
					return
				}
				if err := writeRawEvent(w, event.Name, event.Payload); err != nil {
					return
				}
				flusher.Flush()
			case <-heartbeat.C:
				if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
					return
				}
				flusher.Flush()
			}
		}
	}
}

func deliverEvent(client chan ServerEvent, event ServerEvent) {
	select {
	case client <- event:
	default:
		select {
		case <-client:
		default:
		}

		select {
		case client <- event:
		default:
		}
	}
}

func writeEvent(w http.ResponseWriter, name string, payload any) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return writeRawEvent(w, name, encoded)
}

func writeRawEvent(w http.ResponseWriter, name string, payload []byte) error {
	if _, err := fmt.Fprintf(w, "event: %s\n", name); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "data: %s\n\n", payload)
	return err
}
