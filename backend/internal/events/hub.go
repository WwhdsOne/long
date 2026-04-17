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

// StateProvider fetches the latest state for one browser subscriber.
type StateProvider func(context.Context, string) (vote.State, error)

// Hub broadcasts live SSE snapshots to all connected clients.
type Hub struct {
	mu      sync.RWMutex
	clients map[chan struct{}]struct{}
}

// NewHub creates an in-memory broadcaster for browser subscribers.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[chan struct{}]struct{}),
	}
}

// Subscribe registers a new client channel and returns a cleanup callback.
func (h *Hub) Subscribe() (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)

	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		if _, ok := h.clients[ch]; ok {
			delete(h.clients, ch)
			close(ch)
		}
		h.mu.Unlock()
	}

	return ch, unsubscribe
}

// BroadcastSnapshot sends the newest vote wall snapshot to every listener.
func (h *Hub) BroadcastSnapshot() error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		deliverSnapshot(client)
	}

	return nil
}

// NewHandler exposes the SSE endpoint used by the frontend EventSource client.
func NewHandler(hub *Hub, current StateProvider) http.HandlerFunc {
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
		state, err := current(r.Context(), nickname)
		if err != nil {
			http.Error(w, "STATE_FETCH_FAILED", http.StatusInternalServerError)
			return
		}

		initialPayload, err := snapshotPayload(state)
		if err != nil {
			http.Error(w, "STATE_FETCH_FAILED", http.StatusInternalServerError)
			return
		}

		if err := writeEvent(w, initialPayload); err != nil {
			return
		}
		flusher.Flush()

		client, unsubscribe := hub.Subscribe()
		defer unsubscribe()

		heartbeat := time.NewTicker(25 * time.Second)
		defer heartbeat.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case _, ok := <-client:
				if !ok {
					return
				}
				state, err := current(r.Context(), nickname)
				if err != nil {
					return
				}
				payload, err := snapshotPayload(state)
				if err != nil {
					return
				}
				if err := writeEvent(w, payload); err != nil {
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

func deliverSnapshot(client chan struct{}) {
	select {
	case client <- struct{}{}:
	default:
		select {
		case <-client:
		default:
		}

		select {
		case client <- struct{}{}:
		default:
		}
	}
}

func snapshotPayload(state vote.State) ([]byte, error) {
	return json.Marshal(state)
}

func writeEvent(w http.ResponseWriter, payload []byte) error {
	_, err := fmt.Fprintf(w, "data: %s\n\n", payload)
	return err
}
