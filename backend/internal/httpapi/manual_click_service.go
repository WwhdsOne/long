package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"long/internal/ratelimit"
	"long/internal/vote"
)

const (
	clickEntryHTTP = "http"
	clickEntryWS   = "ws"
)

// ClickTicket 描述单次手动点击用的短时票据。
type ClickTicket struct {
	Value     string `json:"ticket"`
	IssuedAt  int64  `json:"issuedAt"`
	ExpiresAt int64  `json:"expiresAt"`
}

// TicketIssueRequest 描述签发点击票据所需的最小上下文。
type TicketIssueRequest struct {
	Nickname string
	Slug     string
	ClientID string
}

// ManualClickRequest 描述一次手动点击协议载荷。
type ManualClickRequest struct {
	Nickname  string
	Slug      string
	Ticket    string
	ClientID  string
	EntryType string
}

// ManualClickConfig 为单体版票据与风控提供最小可调参数。
type ManualClickConfig struct {
	TicketTTL             time.Duration
	IssueLimitPerSecond   int
	ConsumeLimitPerSecond int
	RiskThreshold         int
	BanDuration           time.Duration
}

// ManualClickServiceOptions 描述手动点击服务依赖。
type ManualClickServiceOptions struct {
	Store      ButtonStore
	ClickGuard ClickGuard
	Config     ManualClickConfig
	Now        func() time.Time
	RandReader io.Reader
}

type manualClickTicketRecord struct {
	Nickname  string
	Slug      string
	IssuedAt  time.Time
	ExpiresAt time.Time
	Nonce     string
}

type manualClickUserState struct {
	issuedAt      []time.Time
	consumedAt    []time.Time
	abnormalCount int
	bannedUntil   time.Time
}

// ManualClickRiskEvent 记录内存风控事件，便于日志或后续后台读取。
type ManualClickRiskEvent struct {
	Nickname      string
	ClientID      string
	Slug          string
	EntryType     string
	Reason        string
	AbnormalCount int
	CreatedAt     int64
	BanStartedAt  int64
	BanEndedAt    int64
}

type manualClickErrorKind string

const (
	manualClickErrorRetryRequired manualClickErrorKind = "retry_required"
	manualClickErrorTooFrequent   manualClickErrorKind = "too_frequent"
)

type manualClickError struct {
	kind       manualClickErrorKind
	retryAfter time.Duration
}

func (e *manualClickError) Error() string {
	if e == nil {
		return ""
	}
	return string(e.kind)
}

func manualClickRequiresRetry(err error) bool {
	var manualErr *manualClickError
	return errors.As(err, &manualErr) && manualErr.kind == manualClickErrorRetryRequired
}

func manualClickTooFrequent(err error) bool {
	var manualErr *manualClickError
	return errors.As(err, &manualErr) && manualErr.kind == manualClickErrorTooFrequent
}

func manualClickRetryAfter(err error) time.Duration {
	var manualErr *manualClickError
	if errors.As(err, &manualErr) {
		return manualErr.retryAfter
	}
	return 0
}

// ManualClickService 在单体进程内维护票据、异常计数与短时封禁状态。
type ManualClickService struct {
	mu         sync.Mutex
	store      ButtonStore
	clickGuard ClickGuard
	now        func() time.Time
	randReader io.Reader
	config     atomic.Value
	tickets    map[string]manualClickTicketRecord
	users      map[string]*manualClickUserState
	events     []ManualClickRiskEvent
}

// NewManualClickService 创建手动点击服务。
func NewManualClickService(options ManualClickServiceOptions) *ManualClickService {
	now := options.Now
	if now == nil {
		now = time.Now
	}
	reader := options.RandReader
	if reader == nil {
		reader = rand.Reader
	}

	service := &ManualClickService{
		store:      options.Store,
		clickGuard: options.ClickGuard,
		now:        now,
		randReader: reader,
		tickets:    make(map[string]manualClickTicketRecord),
		users:      make(map[string]*manualClickUserState),
		events:     make([]ManualClickRiskEvent, 0, 32),
	}
	service.UpdateConfig(options.Config)
	return service
}

// UpdateConfig 允许上层在运行时替换风控参数。
func (s *ManualClickService) UpdateConfig(config ManualClickConfig) {
	s.config.Store(normalizeManualClickConfig(config))
}

func (s *ManualClickService) currentConfig() ManualClickConfig {
	if raw := s.config.Load(); raw != nil {
		return raw.(ManualClickConfig)
	}
	return normalizeManualClickConfig(ManualClickConfig{})
}

// IssueTicket 为指定玩家和按钮签发一次性短时票据。
func (s *ManualClickService) IssueTicket(_ context.Context, request TicketIssueRequest) (ClickTicket, error) {
	nickname := strings.TrimSpace(request.Nickname)
	slug := strings.TrimSpace(request.Slug)
	if nickname == "" || slug == "" {
		return ClickTicket{}, &manualClickError{kind: manualClickErrorRetryRequired}
	}

	now := s.now()
	config := s.currentConfig()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupLocked(now)
	userState := s.userStateLocked(nickname)
	if retryErr := s.blockedErrorLocked(now, userState); retryErr != nil {
		s.recordEventLocked(now, request, "ticket_banned", userState)
		return ClickTicket{}, retryErr
	}
	if retryErr := s.allowIssueLocked(now, request, userState, config); retryErr != nil {
		return ClickTicket{}, retryErr
	}

	token, nonce, err := s.generateToken()
	if err != nil {
		return ClickTicket{}, err
	}

	record := manualClickTicketRecord{
		Nickname:  nickname,
		Slug:      slug,
		IssuedAt:  now,
		ExpiresAt: now.Add(config.TicketTTL),
		Nonce:     nonce,
	}
	s.tickets[token] = record
	userState.issuedAt = append(userState.issuedAt, now)
	userState.abnormalCount = 0

	return ClickTicket{
		Value:     token,
		IssuedAt:  record.IssuedAt.Unix(),
		ExpiresAt: record.ExpiresAt.Unix(),
	}, nil
}

// Click 校验票据后再进入业务结算核心。
func (s *ManualClickService) Click(ctx context.Context, request ManualClickRequest) (vote.ClickResult, error) {
	nickname := strings.TrimSpace(request.Nickname)
	slug := strings.TrimSpace(request.Slug)
	ticketValue := strings.TrimSpace(request.Ticket)
	if nickname == "" || slug == "" || ticketValue == "" {
		return vote.ClickResult{}, &manualClickError{kind: manualClickErrorRetryRequired}
	}

	now := s.now()
	config := s.currentConfig()

	s.mu.Lock()
	s.cleanupLocked(now)
	userState := s.userStateLocked(nickname)
	if retryErr := s.blockedErrorLocked(now, userState); retryErr != nil {
		s.recordEventLocked(now, TicketIssueRequest{
			Nickname: nickname,
			Slug:     slug,
			ClientID: request.ClientID,
		}, "click_banned", userState)
		s.mu.Unlock()
		return vote.ClickResult{}, retryErr
	}
	if retryErr := s.allowConsumeLocked(now, request, userState, config); retryErr != nil {
		s.mu.Unlock()
		return vote.ClickResult{}, retryErr
	}

	record, ok := s.tickets[ticketValue]
	if !ok {
		err := s.markAbnormalLocked(now, request, userState, manualClickErrorRetryRequired, "ticket_missing", 0)
		s.mu.Unlock()
		return vote.ClickResult{}, err
	}
	delete(s.tickets, ticketValue)

	if record.ExpiresAt.Before(now) {
		err := s.markAbnormalLocked(now, request, userState, manualClickErrorRetryRequired, "ticket_expired", 0)
		s.mu.Unlock()
		return vote.ClickResult{}, err
	}
	if record.Nickname != nickname {
		err := s.markAbnormalLocked(now, request, userState, manualClickErrorRetryRequired, "ticket_nickname_mismatch", 0)
		s.mu.Unlock()
		return vote.ClickResult{}, err
	}
	if record.Slug != slug {
		err := s.markAbnormalLocked(now, request, userState, manualClickErrorRetryRequired, "ticket_slug_mismatch", 0)
		s.mu.Unlock()
		return vote.ClickResult{}, err
	}

	userState.consumedAt = append(userState.consumedAt, now)
	userState.abnormalCount = 0
	s.mu.Unlock()

	if apiErr := enforceClickRateLimitForClient(s.clickGuard, request.ClientID, nickname); apiErr != nil {
		if apiErr.Status == 429 {
			return vote.ClickResult{}, &manualClickError{
				kind:       manualClickErrorTooFrequent,
				retryAfter: apiErr.RetryAfter,
			}
		}
		return vote.ClickResult{}, errors.New(apiErr.Code)
	}

	return s.store.ClickButton(ctx, slug, nickname)
}

func (s *ManualClickService) generateToken() (string, string, error) {
	buf := make([]byte, 24)
	if _, err := io.ReadFull(s.randReader, buf); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(buf)
	nonce := base64.RawURLEncoding.EncodeToString(buf[:8])
	return token, nonce, nil
}

func (s *ManualClickService) userStateLocked(nickname string) *manualClickUserState {
	state := s.users[nickname]
	if state == nil {
		state = &manualClickUserState{}
		s.users[nickname] = state
	}
	return state
}

func (s *ManualClickService) cleanupLocked(now time.Time) {
	config := s.currentConfig()
	cutoff := now.Add(-time.Second)
	for nickname, state := range s.users {
		state.issuedAt = filterRecentTimes(state.issuedAt, cutoff)
		state.consumedAt = filterRecentTimes(state.consumedAt, cutoff)
		if !state.bannedUntil.After(now) && len(state.issuedAt) == 0 && len(state.consumedAt) == 0 && state.abnormalCount == 0 {
			delete(s.users, nickname)
		}
	}

	for token, record := range s.tickets {
		if !record.ExpiresAt.After(now) {
			delete(s.tickets, token)
		}
	}

	if len(s.events) > 256 {
		s.events = append([]ManualClickRiskEvent(nil), s.events[len(s.events)-128:]...)
	}

	if config.TicketTTL <= 0 {
		s.UpdateConfig(config)
	}
}

func (s *ManualClickService) blockedErrorLocked(now time.Time, state *manualClickUserState) error {
	if state == nil || !state.bannedUntil.After(now) {
		return nil
	}
	return &manualClickError{
		kind:       manualClickErrorTooFrequent,
		retryAfter: state.bannedUntil.Sub(now),
	}
}

func (s *ManualClickService) allowIssueLocked(now time.Time, request TicketIssueRequest, state *manualClickUserState, config ManualClickConfig) error {
	state.issuedAt = filterRecentTimes(state.issuedAt, now.Add(-time.Second))
	if len(state.issuedAt) < config.IssueLimitPerSecond {
		return nil
	}
	return s.markAbnormalLocked(now, ManualClickRequest{
		Nickname:  request.Nickname,
		Slug:      request.Slug,
		ClientID:  request.ClientID,
		EntryType: clickEntryHTTP,
	}, state, manualClickErrorTooFrequent, "ticket_issue_rate_limited", time.Second)
}

func (s *ManualClickService) allowConsumeLocked(now time.Time, request ManualClickRequest, state *manualClickUserState, config ManualClickConfig) error {
	state.consumedAt = filterRecentTimes(state.consumedAt, now.Add(-time.Second))
	if len(state.consumedAt) < config.ConsumeLimitPerSecond {
		return nil
	}
	return s.markAbnormalLocked(now, request, state, manualClickErrorTooFrequent, "ticket_consume_rate_limited", time.Second)
}

func (s *ManualClickService) markAbnormalLocked(now time.Time, request ManualClickRequest, state *manualClickUserState, kind manualClickErrorKind, reason string, retryAfter time.Duration) error {
	config := s.currentConfig()
	state.abnormalCount++
	if state.abnormalCount >= config.RiskThreshold {
		state.bannedUntil = now.Add(config.BanDuration)
		retryAfter = config.BanDuration
	}
	if retryAfter <= 0 {
		retryAfter = time.Second
	}
	s.recordEventLocked(now, TicketIssueRequest{
		Nickname: request.Nickname,
		Slug:     request.Slug,
		ClientID: request.ClientID,
	}, reason, state)
	return &manualClickError{
		kind:       kind,
		retryAfter: retryAfter,
	}
}

func (s *ManualClickService) recordEventLocked(now time.Time, request TicketIssueRequest, reason string, state *manualClickUserState) {
	event := ManualClickRiskEvent{
		Nickname:      strings.TrimSpace(request.Nickname),
		ClientID:      strings.TrimSpace(request.ClientID),
		Slug:          strings.TrimSpace(request.Slug),
		EntryType:     clickEntryHTTP,
		Reason:        reason,
		AbnormalCount: state.abnormalCount,
		CreatedAt:     now.Unix(),
	}
	if state.bannedUntil.After(now) {
		event.BanStartedAt = now.Unix()
		event.BanEndedAt = state.bannedUntil.Unix()
	}
	s.events = append(s.events, event)
	log.Printf("manual_click_risk nickname=%s ip=%s slug=%s reason=%s abnormal=%d ban_end=%d", event.Nickname, event.ClientID, event.Slug, event.Reason, event.AbnormalCount, event.BanEndedAt)
}

func normalizeManualClickConfig(config ManualClickConfig) ManualClickConfig {
	if config.TicketTTL <= 0 {
		config.TicketTTL = 2 * time.Second
	}
	if config.IssueLimitPerSecond <= 0 {
		config.IssueLimitPerSecond = 6
	}
	if config.ConsumeLimitPerSecond <= 0 {
		config.ConsumeLimitPerSecond = 6
	}
	if config.RiskThreshold <= 0 {
		config.RiskThreshold = 4
	}
	if config.BanDuration <= 0 {
		config.BanDuration = 10 * time.Minute
	}
	return config
}

func filterRecentTimes(items []time.Time, cutoff time.Time) []time.Time {
	filtered := items[:0]
	for _, item := range items {
		if !item.Before(cutoff) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

var _ = ratelimit.ErrTooManyRequests
