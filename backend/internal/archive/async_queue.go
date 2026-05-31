package archive

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// AsyncQueueConfig 控制通用异步队列行为。
type AsyncQueueConfig struct {
	Name         string
	BufferSize   int
	WorkerCount  int
	WriteTimeout time.Duration
}

// AsyncQueue 为 Mongo 日志/事件写入提供进程内 channel + worker 模式。
type AsyncQueue[T any] struct {
	name         string
	ch           chan T
	workerCount  int
	writeTimeout time.Duration
	handle       func(context.Context, T) error
	wg           sync.WaitGroup
	closeOnce    sync.Once
	closed       atomic.Bool
}

// NewAsyncQueue 创建通用异步队列。
func NewAsyncQueue[T any](cfg AsyncQueueConfig, handle func(context.Context, T) error) *AsyncQueue[T] {
	bufferSize := cfg.BufferSize
	if bufferSize <= 0 {
		bufferSize = 128
	}
	workerCount := cfg.WorkerCount
	if workerCount <= 0 {
		workerCount = 1
	}
	return &AsyncQueue[T]{
		name:         cfg.Name,
		ch:           make(chan T, bufferSize),
		workerCount:  workerCount,
		writeTimeout: cfg.WriteTimeout,
		handle:       handle,
	}
}

// Start 启动 worker。
func (q *AsyncQueue[T]) Start() {
	for i := 0; i < q.workerCount; i++ {
		q.wg.Add(1)
		go q.worker()
	}
}

// Enqueue 投递任务；队列满或队列已关闭时返回 false。
func (q *AsyncQueue[T]) Enqueue(item T) (ok bool) {
	if q.closed.Load() {
		return false
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			// Close 与发送并发时，向已关闭 channel 发送会 panic，这里统一降级为返回 false。
			ok = false
		}
	}()
	select {
	case q.ch <- item:
		return true
	default:
		log.Printf("async queue is full: %s", q.name)
		return false
	}
}

// Close 关闭队列并等待 worker 退出。
func (q *AsyncQueue[T]) Close() error {
	q.closeOnce.Do(func() {
		q.closed.Store(true)
		close(q.ch)
		q.wg.Wait()
	})
	return nil
}

func (q *AsyncQueue[T]) worker() {
	defer q.wg.Done()
	for item := range q.ch {
		ctx := context.Background()
		cancel := func() {}
		if q.writeTimeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, q.writeTimeout)
		}
		if err := q.handle(ctx, item); err != nil {
			log.Printf("async queue handle failed: %s: %v", q.name, err)
		}
		cancel()
	}
}
