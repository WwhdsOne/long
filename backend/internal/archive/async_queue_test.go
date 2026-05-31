package archive

import (
	"context"
	"testing"
	"time"
)

func TestAsyncQueueEnqueueAfterCloseReturnsFalseWithoutPanic(t *testing.T) {
	queue := NewAsyncQueue[int](AsyncQueueConfig{
		Name:         "test",
		BufferSize:   1,
		WorkerCount:  0,
		WriteTimeout: 50 * time.Millisecond,
	}, func(context.Context, int) error {
		return nil
	})

	if err := queue.Close(); err != nil {
		t.Fatalf("close queue: %v", err)
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("expected enqueue after close to avoid panic, got %v", recovered)
		}
	}()

	if queue.Enqueue(1) {
		t.Fatal("expected enqueue after close to return false")
	}
}
