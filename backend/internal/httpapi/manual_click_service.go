package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"long/internal/vote"
)

const (
	clickEntryHTTP = "http"
	clickEntryWS   = "ws"
)

// ClickTicket 描述单次手动点击用的短时票据。
type ClickTicket struct {
	Value          string `json:"ticket"`
	IssuedAt       int64  `json:"issuedAt"`
	ExpiresAt      int64  `json:"expiresAt"`
	ChallengeNonce string `json:"challengeNonce"`
}

// TicketIssueRequest 描述签发点击票据所需的最小上下文。
type TicketIssueRequest struct {
	Nickname        string
	Slug            string
	ClientID        string
	FingerprintHash string
}

// ClickPointerSample 描述一次点击中的轨迹点。
type ClickPointerSample struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	T int64   `json:"t"`
}

// ClickBehavior 描述前端上报的交互行为信号。
type ClickBehavior struct {
	PointerType     string               `json:"pointerType"`
	PressDurationMS int64                `json:"pressDurationMs"`
	Trajectory      []ClickPointerSample `json:"trajectory"`
}

// ManualClickRequest 描述一次手动点击协议载荷。
type ManualClickRequest struct {
	Nickname         string
	Slug             string
	Ticket           string
	ClientID         string
	EntryType        string
	FingerprintHash  string
	FingerprintProof string
	Behavior         ClickBehavior
}

// ManualClickConfig 为单体版票据与风控提供可调参数。
type ManualClickConfig struct {
	TicketTTL             time.Duration
	IssueLimitPerSecond   int
	ConsumeLimitPerSecond int
	RiskThreshold         int
	BanDuration           time.Duration
	MinPressDuration      time.Duration
	MaxPressDuration      time.Duration
	MinTrajectoryPoints   int
	MaxTrajectoryPoints   int
	MinPathDistance       float64
	MinDisplacement       float64
	MinCurvature          float64
	MinSpeedVariance      float64
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
	Nickname        string
	Slug            string
	IssuedAt        time.Time
	ExpiresAt       time.Time
	Nonce           string
	FingerprintHash string
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
	s.config.Store(config)
}

func (s *ManualClickService) currentConfig() ManualClickConfig {
	if raw := s.config.Load(); raw != nil {
		return raw.(ManualClickConfig)
	}
	return ManualClickConfig{}
}

// IssueTicket 为指定玩家和按钮签发一次性短时票据。
func (s *ManualClickService) IssueTicket(_ context.Context, request TicketIssueRequest) (ClickTicket, error) {
	nickname := strings.TrimSpace(request.Nickname)
	slug := strings.TrimSpace(request.Slug)
	fingerprintHash := strings.TrimSpace(request.FingerprintHash)
	if nickname == "" || slug == "" || fingerprintHash == "" {
		return ClickTicket{}, &manualClickError{kind: manualClickErrorRetryRequired}
	}

	now := s.now()
	config := s.currentConfig()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupLocked(now)
	userState := s.userStateLocked(nickname)
	if retryErr := s.blockedErrorLocked(now, userState); retryErr != nil {
		s.recordTicketEventLocked(now, request, "ticket_banned", userState)
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
		Nickname:        nickname,
		Slug:            slug,
		IssuedAt:        now,
		ExpiresAt:       now.Add(config.TicketTTL),
		Nonce:           nonce,
		FingerprintHash: fingerprintHash,
	}
	s.tickets[token] = record
	userState.issuedAt = append(userState.issuedAt, now)
	userState.abnormalCount = 0

	return ClickTicket{
		Value:          token,
		IssuedAt:       record.IssuedAt.Unix(),
		ExpiresAt:      record.ExpiresAt.Unix(),
		ChallengeNonce: record.Nonce,
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
		s.recordManualEventLocked(now, request, "click_banned", userState)
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
	if strings.TrimSpace(request.FingerprintHash) == "" {
		err := s.markAbnormalLocked(now, request, userState, manualClickErrorRetryRequired, "fingerprint_missing", 0)
		s.mu.Unlock()
		return vote.ClickResult{}, err
	}
	if record.FingerprintHash != strings.TrimSpace(request.FingerprintHash) {
		err := s.markAbnormalLocked(now, request, userState, manualClickErrorRetryRequired, "fingerprint_mismatch", 0)
		s.mu.Unlock()
		return vote.ClickResult{}, err
	}
	if !validFingerprintProof(record.FingerprintHash, ticketValue, record.Nonce, request.FingerprintProof) {
		err := s.markAbnormalLocked(now, request, userState, manualClickErrorRetryRequired, "fingerprint_proof_invalid", 0)
		s.mu.Unlock()
		return vote.ClickResult{}, err
	}
	if err := validateClickBehavior(request.Behavior, config); err != nil {
		err = s.markAbnormalLocked(now, request, userState, manualClickErrorRetryRequired, err.Error(), 0)
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
		Nickname:        request.Nickname,
		Slug:            request.Slug,
		ClientID:        request.ClientID,
		EntryType:       clickEntryHTTP,
		FingerprintHash: request.FingerprintHash,
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
	s.recordManualEventLocked(now, request, reason, state)
	return &manualClickError{
		kind:       kind,
		retryAfter: retryAfter,
	}
}

func (s *ManualClickService) recordTicketEventLocked(now time.Time, request TicketIssueRequest, reason string, state *manualClickUserState) {
	s.recordEventLocked(now, request.Nickname, request.ClientID, request.Slug, clickEntryHTTP, reason, state)
}

func (s *ManualClickService) recordManualEventLocked(now time.Time, request ManualClickRequest, reason string, state *manualClickUserState) {
	entryType := strings.TrimSpace(request.EntryType)
	if entryType == "" {
		entryType = clickEntryHTTP
	}
	s.recordEventLocked(now, request.Nickname, request.ClientID, request.Slug, entryType, reason, state)
}

func (s *ManualClickService) recordEventLocked(now time.Time, nickname string, clientID string, slug string, entryType string, reason string, state *manualClickUserState) {
	event := ManualClickRiskEvent{
		Nickname:      strings.TrimSpace(nickname),
		ClientID:      strings.TrimSpace(clientID),
		Slug:          strings.TrimSpace(slug),
		EntryType:     strings.TrimSpace(entryType),
		Reason:        reason,
		AbnormalCount: state.abnormalCount,
		CreatedAt:     now.Unix(),
	}
	if state.bannedUntil.After(now) {
		event.BanStartedAt = now.Unix()
		event.BanEndedAt = state.bannedUntil.Unix()
	}
	s.events = append(s.events, event)
	log.Printf("manual_click_risk nickname=%s ip=%s slug=%s entry=%s reason=%s abnormal=%d ban_end=%d", event.Nickname, event.ClientID, event.Slug, event.EntryType, event.Reason, event.AbnormalCount, event.BanEndedAt)
}

func validFingerprintProof(fingerprintHash string, ticket string, challengeNonce string, provided string) bool {
	expected := fingerprintProof(fingerprintHash, ticket, challengeNonce)
	if expected == "" || strings.TrimSpace(provided) == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(strings.TrimSpace(provided))) == 1
}

func fingerprintProof(fingerprintHash string, ticket string, challengeNonce string) string {
	fingerprintHash = strings.TrimSpace(fingerprintHash)
	ticket = strings.TrimSpace(ticket)
	challengeNonce = strings.TrimSpace(challengeNonce)
	if fingerprintHash == "" || ticket == "" || challengeNonce == "" {
		return ""
	}

	sum := sha256.Sum256([]byte(fingerprintHash + ":" + ticket + ":" + challengeNonce))
	return hex.EncodeToString(sum[:])
}

func validateClickBehavior(behavior ClickBehavior, config ManualClickConfig) error {
	if behavior.PressDurationMS < config.MinPressDuration.Milliseconds() {
		return errors.New("press_duration_too_short")
	}
	if behavior.PressDurationMS > config.MaxPressDuration.Milliseconds() {
		return errors.New("press_duration_too_long")
	}

	points := normalizeTrajectory(behavior.Trajectory, config.MaxTrajectoryPoints)
	if len(points) < config.MinTrajectoryPoints {
		return errors.New("trajectory_points_too_few")
	}

	metrics, ok := computeTrajectoryMetrics(points)
	if !ok {
		return errors.New("trajectory_invalid")
	}
	if metrics.pathDistance < config.MinPathDistance {
		return errors.New("trajectory_path_too_short")
	}
	if metrics.displacement < config.MinDisplacement {
		return errors.New("trajectory_displacement_too_short")
	}
	if metrics.curvature < config.MinCurvature {
		return errors.New("trajectory_curvature_too_low")
	}
	if metrics.speedVariance < config.MinSpeedVariance {
		return errors.New("trajectory_speed_variance_too_low")
	}

	return nil
}

type trajectoryMetrics struct {
	pathDistance  float64
	displacement  float64
	curvature     float64
	speedVariance float64
}

func normalizeTrajectory(points []ClickPointerSample, maxPoints int) []ClickPointerSample {
	if len(points) == 0 {
		return nil
	}
	normalized := make([]ClickPointerSample, 0, len(points))
	for _, point := range points {
		if math.IsNaN(point.X) || math.IsNaN(point.Y) || math.IsInf(point.X, 0) || math.IsInf(point.Y, 0) {
			continue
		}
		normalized = append(normalized, point)
	}
	if maxPoints > 0 && len(normalized) > maxPoints {
		normalized = normalized[len(normalized)-maxPoints:]
	}
	return normalized
}

func computeTrajectoryMetrics(points []ClickPointerSample) (trajectoryMetrics, bool) {
	if len(points) < 2 {
		return trajectoryMetrics{}, false
	}

	metrics := trajectoryMetrics{}
	speeds := make([]float64, 0, len(points)-1)
	for index := 1; index < len(points); index++ {
		dx := points[index].X - points[index-1].X
		dy := points[index].Y - points[index-1].Y
		dt := points[index].T - points[index-1].T
		if dt <= 0 {
			return trajectoryMetrics{}, false
		}

		distance := math.Hypot(dx, dy)
		metrics.pathDistance += distance
		speeds = append(speeds, distance/float64(dt))

		if index >= 2 {
			prevDX := points[index-1].X - points[index-2].X
			prevDY := points[index-1].Y - points[index-2].Y
			prevAngle := math.Atan2(prevDY, prevDX)
			nextAngle := math.Atan2(dy, dx)
			turn := math.Abs(nextAngle - prevAngle)
			if turn > math.Pi {
				turn = 2*math.Pi - turn
			}
			metrics.curvature += turn
		}
	}

	metrics.displacement = math.Hypot(points[len(points)-1].X-points[0].X, points[len(points)-1].Y-points[0].Y)

	mean := 0.0
	for _, speed := range speeds {
		mean += speed
	}
	mean /= float64(len(speeds))
	if mean <= 0 {
		return trajectoryMetrics{}, false
	}

	variance := 0.0
	for _, speed := range speeds {
		diff := speed - mean
		variance += diff * diff
	}
	variance /= float64(len(speeds))
	metrics.speedVariance = math.Sqrt(variance) / mean
	return metrics, true
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
