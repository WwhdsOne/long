package events

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"long/internal/vote"
)

// SnapshotProvider fetches the latest button list for a newly connected stream.
type SnapshotProvider func(context.Context) ([]vote.Button, error)

// Hub broadcasts live SSE snapshots to all connected clients.
type Hub struct {
	mu      sync.RWMutex
	clients map[chan []byte]struct{}
}

// NewHub creates an in-memory broadcaster for browser subscribers.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[chan []byte]struct{}),
	}
}

// Subscribe registers a new client channel and returns a cleanup callback.
func (h *Hub) Subscribe() (<-chan []byte, func()) {
	ch := make(chan []byte, 1)

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
func (h *Hub) BroadcastSnapshot(buttons []vote.Button) error {
	payload, err := snapshotPayload(buttons)
	if err != nil {
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		deliverSnapshot(client, payload)
	}

	return nil
}

// NewHandler exposes the SSE endpoint used by the frontend EventSource client.
func NewHandler(hub *Hub, current SnapshotProvider) http.HandlerFunc {
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

		buttons, err := current(r.Context())
		if err != nil {
			http.Error(w, "BUTTONS_FETCH_FAILED", http.StatusInternalServerError)
			return
		}

		initialPayload, err := snapshotPayload(buttons)
		if err != nil {
			http.Error(w, "BUTTONS_FETCH_FAILED", http.StatusInternalServerError)
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
			case payload, ok := <-client:
				if !ok {
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

func deliverSnapshot(client chan []byte, payload []byte) {
	cloned := append([]byte(nil), payload...)

	select {
	case client <- cloned:
	default:
		select {
		case <-client:
		default:
		}

		select {
		case client <- cloned:
		default:
		}
	}
}

func snapshotPayload(buttons []vote.Button) ([]byte, error) {
	return json.Marshal(struct {
		Buttons []vote.Button `json:"buttons"`
	}{
		Buttons: buttons,
	})
}

func writeEvent(w http.ResponseWriter, payload []byte) error {
	_, err := fmt.Fprintf(w, "data: %s\n\n", payload)
	return err
}
