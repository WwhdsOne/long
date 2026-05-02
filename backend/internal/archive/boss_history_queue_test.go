package archive

import (
	"context"
	"sync"
	"testing"
	"time"

	"long/internal/core"
)

type recordingBossHistoryWriter struct {
	mu      sync.Mutex
	entries []core.BossHistoryEntry
}

func (w *recordingBossHistoryWriter) SaveBossHistory(_ context.Context, entry core.BossHistoryEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.entries = append(w.entries, entry)
	return nil
}

func (w *recordingBossHistoryWriter) snapshot() []core.BossHistoryEntry {
	w.mu.Lock()
	defer w.mu.Unlock()
	return append([]core.BossHistoryEntry(nil), w.entries...)
}

func TestBossHistoryQueueConsumesSubmittedEntries(t *testing.T) {
	writer := &recordingBossHistoryWriter{}
	queue := NewBossHistoryQueue(BossHistoryQueueConfig{
		BufferSize:   4,
		WorkerCount:  1,
		WriteTimeout: 200 * time.Millisecond,
	}, writer)

	queue.Start()
	defer queue.Close()

	ok := queue.Enqueue(core.BossHistoryEntry{
		Boss: core.Boss{
			ID:        "boss-1",
			Name:      "测试 Boss",
			Status:    "defeated",
			StartedAt: 1,
		},
	})
	if !ok {
		t.Fatal("expected enqueue to succeed")
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		entries := writer.snapshot()
		if len(entries) == 1 && entries[0].ID == "boss-1" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("expected worker to consume entry, got %+v", writer.snapshot())
}

func TestBossHistoryQueueReturnsFalseWhenBufferFull(t *testing.T) {
	writer := &recordingBossHistoryWriter{}
	queue := NewBossHistoryQueue(BossHistoryQueueConfig{
		BufferSize:   1,
		WorkerCount:  0,
		WriteTimeout: 200 * time.Millisecond,
	}, writer)
	defer queue.Close()

	if !queue.Enqueue(core.BossHistoryEntry{Boss: core.Boss{ID: "boss-1"}}) {
		t.Fatal("expected first enqueue to succeed")
	}
	if queue.Enqueue(core.BossHistoryEntry{Boss: core.Boss{ID: "boss-2"}}) {
		t.Fatal("expected second enqueue to fail when buffer is full")
	}
}
