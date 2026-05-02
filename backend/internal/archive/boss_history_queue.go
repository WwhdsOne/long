package archive

import (
	"context"
	"sync"
	"time"

	"long/internal/core"
	"long/internal/xlog"
)

// BossHistoryWriter 负责将 Boss 历史写入冷数据存储。
type BossHistoryWriter interface {
	SaveBossHistory(context.Context, core.BossHistoryEntry) error
}

// BossHistoryQueueConfig 控制归档队列行为。
type BossHistoryQueueConfig struct {
	BufferSize   int
	WorkerCount  int
	WriteTimeout time.Duration
}

// BossHistoryQueue 用进程内 channel + worker 异步归档 Boss 历史。
type BossHistoryQueue struct {
	writer       BossHistoryWriter
	workerCount  int
	writeTimeout time.Duration
	ch           chan core.BossHistoryEntry
	wg           sync.WaitGroup
	closeOnce    sync.Once
}

// NewBossHistoryQueue 创建 Boss 历史归档队列。
func NewBossHistoryQueue(cfg BossHistoryQueueConfig, writer BossHistoryWriter) *BossHistoryQueue {
	bufferSize := cfg.BufferSize
	if bufferSize <= 0 {
		bufferSize = 64
	}

	return &BossHistoryQueue{
		writer:       writer,
		workerCount:  maxInt(cfg.WorkerCount, 1),
		writeTimeout: cfg.WriteTimeout,
		ch:           make(chan core.BossHistoryEntry, bufferSize),
	}
}

// Start 启动固定数量 worker。
func (q *BossHistoryQueue) Start() {
	for i := 0; i < q.workerCount; i++ {
		q.wg.Add(1)
		go q.worker()
	}
}

// Enqueue 投递一条 Boss 历史；队列满时返回 false。
func (q *BossHistoryQueue) Enqueue(entry core.BossHistoryEntry) bool {
	select {
	case q.ch <- entry:
		return true
	default:
		return false
	}
}

// Close 停止接收新任务并等待 worker 退出。
func (q *BossHistoryQueue) Close() error {
	q.closeOnce.Do(func() {
		close(q.ch)
		q.wg.Wait()
	})
	return nil
}

func (q *BossHistoryQueue) worker() {
	defer q.wg.Done()

	for entry := range q.ch {
		if q.writer == nil {
			continue
		}

		ctx := context.Background()
		cancel := func() {}
		if q.writeTimeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, q.writeTimeout)
		}

		if err := q.writer.SaveBossHistory(ctx, entry); err != nil {
			xlog.L().Error("archive boss history failed", xlog.Err(err))
		}
		cancel()
	}
}

func maxInt(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}
