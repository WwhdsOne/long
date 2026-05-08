package httpapi

import (
	"context"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"go.uber.org/zap"

	"long/internal/xlog"
)

const cloudflareTurnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

const (
	StaminaPurchaseTurnstileAllow       = "allow"
	StaminaPurchaseTurnstileRequire     = "require"
	StaminaPurchaseTurnstileInvalid     = "invalid"
	StaminaPurchaseTurnstileUnavailable = "unavailable"

	PlayerLoginTurnstileAllow       = "allow"
	PlayerLoginTurnstileRequire     = "require"
	PlayerLoginTurnstileInvalid     = "invalid"
	PlayerLoginTurnstileUnavailable = "unavailable"
)

type StaminaPurchaseTurnstileRequest struct {
	Nickname string
	Token    string
	RemoteIP string
}

type StaminaPurchaseTurnstileResult struct {
	Decision string
	SiteKey  string
}

type StaminaPurchaseTurnstileConfig struct {
	Enabled                   bool
	SiteKey                   string
	SecretKey                 string
	PurchaseStaminaSampleRate float64
	VerifyTimeoutMS           int
	HTTPClient                *http.Client
	RandomFloat64             func() float64
}

type PlayerLoginTurnstileRequest struct {
	Nickname string
	Token    string
	RemoteIP string
}

type PlayerLoginTurnstileResult struct {
	Decision string
	SiteKey  string
}

type PlayerLoginTurnstileConfig struct {
	Enabled         bool
	SiteKey         string
	SecretKey       string
	VerifyTimeoutMS int
	HTTPClient      *http.Client
}

type staminaPurchaseTurnstileService struct {
	enabled                   bool
	siteKey                   string
	secretKey                 string
	purchaseStaminaSampleRate float64
	verifyTimeout             time.Duration
	httpClient                *http.Client
	randomFloat64             func() float64
}

type playerLoginTurnstileService struct {
	enabled       bool
	siteKey       string
	verifyTimeout time.Duration
	verifier      *turnstileVerifier
}

type turnstileVerifier struct {
	secretKey     string
	verifyTimeout time.Duration
	httpClient    *http.Client
}

type turnstileVerifyResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

func NewStaminaPurchaseTurnstile(cfg StaminaPurchaseTurnstileConfig) StaminaPurchaseTurnstile {
	client := cfg.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	randomFloat64 := cfg.RandomFloat64
	if randomFloat64 == nil {
		randomFloat64 = rand.Float64
	}
	timeout := time.Duration(cfg.VerifyTimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &staminaPurchaseTurnstileService{
		enabled:                   cfg.Enabled,
		siteKey:                   strings.TrimSpace(cfg.SiteKey),
		secretKey:                 strings.TrimSpace(cfg.SecretKey),
		purchaseStaminaSampleRate: cfg.PurchaseStaminaSampleRate,
		verifyTimeout:             timeout,
		httpClient:                client,
		randomFloat64:             randomFloat64,
	}
}

func NewPlayerLoginTurnstile(cfg PlayerLoginTurnstileConfig) PlayerLoginTurnstile {
	client := cfg.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	timeout := time.Duration(cfg.VerifyTimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &playerLoginTurnstileService{
		enabled:       cfg.Enabled,
		siteKey:       strings.TrimSpace(cfg.SiteKey),
		verifyTimeout: timeout,
		verifier: &turnstileVerifier{
			secretKey:     strings.TrimSpace(cfg.SecretKey),
			verifyTimeout: timeout,
			httpClient:    client,
		},
	}
}

func (s *staminaPurchaseTurnstileService) CheckPurchaseStamina(ctx context.Context, req StaminaPurchaseTurnstileRequest) (StaminaPurchaseTurnstileResult, error) {
	token := strings.TrimSpace(req.Token)
	if !s.enabled {
		return StaminaPurchaseTurnstileResult{Decision: StaminaPurchaseTurnstileAllow}, nil
	}

	if token == "" {
		if !s.shouldRequireVerification() {
			xlog.L().Info("stamina purchase turnstile skipped",
				zap.String("nickname", strings.TrimSpace(req.Nickname)),
				zap.String("remote_ip", strings.TrimSpace(req.RemoteIP)))
			return StaminaPurchaseTurnstileResult{Decision: StaminaPurchaseTurnstileAllow}, nil
		}
		xlog.L().Info("stamina purchase turnstile required",
			zap.String("nickname", strings.TrimSpace(req.Nickname)),
			zap.String("remote_ip", strings.TrimSpace(req.RemoteIP)))
		return StaminaPurchaseTurnstileResult{
			Decision: StaminaPurchaseTurnstileRequire,
			SiteKey:  s.siteKey,
		}, nil
	}

	return s.verifyToken(ctx, req, token), nil
}

func (s *staminaPurchaseTurnstileService) shouldRequireVerification() bool {
	return s.randomFloat64() < s.purchaseStaminaSampleRate
}

func (s *playerLoginTurnstileService) CheckPlayerLogin(ctx context.Context, req PlayerLoginTurnstileRequest) (PlayerLoginTurnstileResult, error) {
	token := strings.TrimSpace(req.Token)
	if !s.enabled {
		return PlayerLoginTurnstileResult{Decision: PlayerLoginTurnstileAllow}, nil
	}
	if token == "" {
		xlog.L().Info("player login turnstile required",
			zap.String("nickname", strings.TrimSpace(req.Nickname)),
			zap.String("remote_ip", strings.TrimSpace(req.RemoteIP)))
		return PlayerLoginTurnstileResult{
			Decision: PlayerLoginTurnstileRequire,
			SiteKey:  s.siteKey,
		}, nil
	}

	switch s.verifier.verifyToken(ctx, req.Nickname, req.RemoteIP, token, "player login turnstile") {
	case turnstileVerifyAllow:
		return PlayerLoginTurnstileResult{Decision: PlayerLoginTurnstileAllow}, nil
	case turnstileVerifyInvalid:
		return PlayerLoginTurnstileResult{Decision: PlayerLoginTurnstileInvalid}, nil
	default:
		return PlayerLoginTurnstileResult{Decision: PlayerLoginTurnstileUnavailable}, nil
	}
}

func (s *staminaPurchaseTurnstileService) verifyToken(ctx context.Context, req StaminaPurchaseTurnstileRequest, token string) StaminaPurchaseTurnstileResult {
	verifier := &turnstileVerifier{
		secretKey:     s.secretKey,
		verifyTimeout: s.verifyTimeout,
		httpClient:    s.httpClient,
	}
	switch verifier.verifyToken(ctx, req.Nickname, req.RemoteIP, token, "stamina purchase turnstile") {
	case turnstileVerifyAllow:
		return StaminaPurchaseTurnstileResult{Decision: StaminaPurchaseTurnstileAllow}
	case turnstileVerifyInvalid:
		return StaminaPurchaseTurnstileResult{Decision: StaminaPurchaseTurnstileInvalid}
	default:
		return StaminaPurchaseTurnstileResult{Decision: StaminaPurchaseTurnstileUnavailable}
	}
}

const (
	turnstileVerifyAllow = iota + 1
	turnstileVerifyInvalid
	turnstileVerifyUnavailable
)

func (v *turnstileVerifier) verifyToken(ctx context.Context, nickname string, remoteIP string, token string, logPrefix string) int {
	verifyCtx, cancel := context.WithTimeout(ctx, v.verifyTimeout)
	defer cancel()

	values := url.Values{}
	values.Set("secret", v.secretKey)
	values.Set("response", token)
	if trimmedRemoteIP := strings.TrimSpace(remoteIP); trimmedRemoteIP != "" {
		values.Set("remoteip", trimmedRemoteIP)
	}

	httpReq, err := http.NewRequestWithContext(verifyCtx, http.MethodPost, cloudflareTurnstileVerifyURL, strings.NewReader(values.Encode()))
	if err != nil {
		xlog.L().Warn("build "+logPrefix+" request failed", xlog.Err(err))
		return turnstileVerifyUnavailable
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.httpClient.Do(httpReq)
	if err != nil {
		xlog.L().Warn(logPrefix+" verify failed", xlog.Err(err))
		return turnstileVerifyUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		xlog.L().Warn(logPrefix+" verify returned non-200", zap.Int("status_code", resp.StatusCode))
		return turnstileVerifyUnavailable
	}

	var payload turnstileVerifyResponse
	if err := sonic.ConfigDefault.NewDecoder(resp.Body).Decode(&payload); err != nil {
		xlog.L().Warn("decode "+logPrefix+" verify response failed", xlog.Err(err))
		return turnstileVerifyUnavailable
	}
	if payload.Success {
		xlog.L().Info(logPrefix+" verified",
			zap.String("nickname", strings.TrimSpace(nickname)),
			zap.String("remote_ip", strings.TrimSpace(remoteIP)))
		return turnstileVerifyAllow
	}
	if containsTurnstileServerError(payload.ErrorCodes) {
		xlog.L().Warn(logPrefix+" verify unavailable", zap.Strings("error_codes", payload.ErrorCodes))
		return turnstileVerifyUnavailable
	}
	xlog.L().Info(logPrefix+" invalid",
		zap.String("nickname", strings.TrimSpace(nickname)),
		zap.Strings("error_codes", payload.ErrorCodes))
	return turnstileVerifyInvalid
}

func containsTurnstileServerError(codes []string) bool {
	for _, code := range codes {
		if strings.TrimSpace(code) == "internal-error" {
			return true
		}
	}
	return false
}
