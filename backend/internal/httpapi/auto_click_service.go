package httpapi

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"long/internal/vote"
)

type autoClickTask struct {
	Nickname  string
	Slug      string
	StartedAt time.Time
	UpdatedAt time.Time
}

// AutoClickStatus 描述当前玩家的挂机状态。
type AutoClickStatus struct {
	Active        bool   `json:"active"`
	ButtonKey     string `json:"buttonKey,omitempty"`
	StartedAt     int64  `json:"startedAt,omitempty"`
	UpdatedAt     int64  `json:"updatedAt,omitempty"`
	IntervalMs    int64  `json:"intervalMs"`
	RatePerSecond int    `json:"ratePerSecond"`
}

// AutoClickServiceOptions 描述挂机调度器依赖。
type AutoClickServiceOptions struct {
	Store           ButtonStore
	ChangePublisher ChangePublisher
	Interval        time.Duration
	Now             func() time.Time
	AutoStart       bool
}

// AutoClickService 以单体内存任务表托管官方挂机。
type AutoClickService struct {
	mu              sync.RWMutex
	store           ButtonStore
	changePublisher ChangePublisher
	interval        time.Duration
	now             func() time.Time
	tasks           map[string]autoClickTask
	stopCh          chan struct{}
	doneCh          chan struct{}
	closed          bool
}

// NewAutoClickService 创建挂机调度器。
func NewAutoClickService(options AutoClickServiceOptions) *AutoClickService {
	interval := options.Interval
	if interval <= 0 {
		interval = time.Second / 3
	}
	now := options.Now
	if now == nil {
		now = time.Now
	}

	service := &AutoClickService{
		store:           options.Store,
		changePublisher: options.ChangePublisher,
		interval:        interval,
		now:             now,
		tasks:           make(map[string]autoClickTask),
		stopCh:          make(chan struct{}),
		doneCh:          make(chan struct{}),
	}

	if options.AutoStart {
		go service.loop()
	} else {
		close(service.doneCh)
	}

	return service
}

// Start 启动或更新当前玩家的挂机目标。
func (s *AutoClickService) Start(_ context.Context, nickname string, slug string) (AutoClickStatus, error) {
	nickname = strings.TrimSpace(nickname)
	slug = strings.TrimSpace(slug)
	if nickname == "" || slug == "" {
		return AutoClickStatus{}, errors.New("invalid auto click target")
	}

	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[nickname]
	if !ok {
		task = autoClickTask{
			Nickname:  nickname,
			StartedAt: now,
		}
	}
	task.Slug = slug
	task.UpdatedAt = now
	s.tasks[nickname] = task
	return s.statusLocked(nickname), nil
}

// Stop 停止当前玩家的挂机任务。
func (s *AutoClickService) Stop(nickname string) AutoClickStatus {
	nickname = strings.TrimSpace(nickname)
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tasks, nickname)
	return s.inactiveStatus()
}

// Status 返回当前玩家的挂机状态。
func (s *AutoClickService) Status(nickname string) AutoClickStatus {
	nickname = strings.TrimSpace(nickname)
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.statusLocked(nickname)
}

// Close 关闭后台调度循环。
func (s *AutoClickService) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	close(s.stopCh)
	doneCh := s.doneCh
	s.mu.Unlock()

	select {
	case <-doneCh:
	default:
	}
	return nil
}

func (s *AutoClickService) loop() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	defer close(s.doneCh)

	for {
		select {
		case <-ticker.C:
			s.runOnce(context.Background())
		case <-s.stopCh:
			return
		}
	}
}

func (s *AutoClickService) runOnce(ctx context.Context) {
	s.mu.RLock()
	tasks := make([]autoClickTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	s.mu.RUnlock()

	for _, task := range tasks {
		result, err := s.store.AutoClickBossPart(ctx, task.Slug, task.Nickname)
		if err != nil {
			if errors.Is(err, vote.ErrInvalidNickname) || errors.Is(err, vote.ErrSensitiveNickname) {
				s.Stop(task.Nickname)
			}
			log.Printf("auto_click_failed nickname=%s slug=%s err=%v", task.Nickname, task.Slug, err)
			continue
		}

		changeType := vote.StateChangeButtonClicked
		if result.BroadcastUserAll {
			changeType = vote.StateChangeBossChanged
		}
		change := vote.StateChange{
			Type:      changeType,
			Nickname:  task.Nickname,
			Timestamp: s.now().Unix(),
		}
		if result.BroadcastUserAll {
			change.BroadcastUserAll = true
		}
		publishChange(ctx, s.changePublisher, change)
	}
}

func (s *AutoClickService) statusLocked(nickname string) AutoClickStatus {
	task, ok := s.tasks[nickname]
	if !ok {
		return s.inactiveStatus()
	}
	return AutoClickStatus{
		Active:        true,
		ButtonKey:     task.Slug,
		StartedAt:     task.StartedAt.Unix(),
		UpdatedAt:     task.UpdatedAt.Unix(),
		IntervalMs:    s.interval.Milliseconds(),
		RatePerSecond: int(time.Second / s.interval),
	}
}

func (s *AutoClickService) inactiveStatus() AutoClickStatus {
	return AutoClickStatus{
		Active:        false,
		IntervalMs:    s.interval.Milliseconds(),
		RatePerSecond: int(time.Second / s.interval),
	}
}
