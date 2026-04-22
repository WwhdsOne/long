package playerauth

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidToken       = errors.New("invalid token")
)

type Config struct {
	Namespace string
	JWTSecret string
	TokenTTL  time.Duration
	Now       func() time.Time
}

type Service struct {
	client        redis.UniversalClient
	accountPrefix string
	jwtSecret     []byte
	tokenTTL      time.Duration
	now           func() time.Time
	validator     interface{ Validate(string) error }
	passwordCost  int
}

type claims struct {
	jwt.RegisteredClaims
}

func NewService(client redis.UniversalClient, cfg Config, validator interface{ Validate(string) error }) *Service {
	tokenTTL := cfg.TokenTTL
	if tokenTTL <= 0 {
		tokenTTL = 7 * 24 * time.Hour
	}

	now := cfg.Now
	if now == nil {
		now = time.Now
	}

	return &Service{
		client:        client,
		accountPrefix: strings.TrimSpace(cfg.Namespace) + "player-auth:",
		jwtSecret:     []byte(cfg.JWTSecret),
		tokenTTL:      tokenTTL,
		now:           now,
		validator:     validator,
		passwordCost:  bcrypt.DefaultCost,
	}
}

func (s *Service) Login(ctx context.Context, nickname string, password string) (string, string, error) {
	normalizedNickname, ok := normalizeNickname(nickname)
	if !ok {
		return "", "", ErrInvalidCredentials
	}
	if strings.TrimSpace(password) == "" {
		return "", "", ErrInvalidPassword
	}

	accountKey := s.accountKey(normalizedNickname)
	record, err := s.client.HGetAll(ctx, accountKey).Result()
	if err != nil {
		return "", "", err
	}

	if len(record) == 0 {
		if s.validator != nil {
			if err := s.validator.Validate(normalizedNickname); err != nil {
				return "", "", err
			}
		}
		if err := s.upsertPassword(ctx, accountKey, normalizedNickname, password, true); err != nil {
			return "", "", err
		}
	} else {
		passwordHash := record["password_hash"]
		if passwordHash == "" || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) != nil {
			return "", "", ErrInvalidCredentials
		}
	}

	token, err := s.issueToken(ctx, normalizedNickname)
	if err != nil {
		return "", "", err
	}
	return token, normalizedNickname, nil
}

func (s *Service) Verify(ctx context.Context, token string) (string, error) {
	parsed, err := jwt.ParseWithClaims(strings.TrimSpace(token), &claims{}, func(token *jwt.Token) (any, error) {
		if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return "", ErrInvalidToken
	}

	registered, ok := parsed.Claims.(*claims)
	if !ok || !parsed.Valid {
		return "", ErrInvalidToken
	}

	nickname, ok := normalizeNickname(registered.Subject)
	if !ok {
		return "", ErrInvalidToken
	}

	resetUnix, err := s.currentResetUnix(ctx, nickname)
	if err != nil {
		return "", err
	}
	if resetUnix > 0 && registered.IssuedAt != nil && registered.IssuedAt.Time.Unix() < resetUnix {
		return "", ErrInvalidToken
	}

	return nickname, nil
}

func (s *Service) ResetPassword(ctx context.Context, nickname string, password string) error {
	normalizedNickname, ok := normalizeNickname(nickname)
	if !ok {
		return ErrInvalidCredentials
	}
	if strings.TrimSpace(password) == "" {
		return ErrInvalidPassword
	}
	if s.validator != nil {
		if err := s.validator.Validate(normalizedNickname); err != nil {
			return err
		}
	}

	return s.upsertPassword(ctx, s.accountKey(normalizedNickname), normalizedNickname, password, false)
}

func (s *Service) upsertPassword(ctx context.Context, accountKey string, nickname string, password string, createOnly bool) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), s.passwordCost)
	if err != nil {
		return err
	}

	now := s.now().Unix()
	values := map[string]any{
		"nickname":          nickname,
		"password_hash":     string(passwordHash),
		"updated_at":        now,
		"password_reset_at": now,
	}
	if createOnly {
		values["created_at"] = now
	} else {
		exists, err := s.client.Exists(ctx, accountKey).Result()
		if err != nil {
			return err
		}
		if exists == 0 {
			values["created_at"] = now
		}
	}

	return s.client.HSet(ctx, accountKey, values).Err()
}

func (s *Service) issueToken(ctx context.Context, nickname string) (string, error) {
	issuedAt := s.now()
	claims := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   nickname,
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(s.tokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *Service) currentResetUnix(ctx context.Context, nickname string) (int64, error) {
	record, err := s.client.HGet(ctx, s.accountKey(nickname), "password_reset_at").Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, ErrInvalidToken
		}
		return 0, err
	}

	resetUnix, err := strconv.ParseInt(record, 10, 64)
	if err != nil {
		return 0, ErrInvalidToken
	}
	return resetUnix, nil
}

func (s *Service) accountKey(nickname string) string {
	return s.accountPrefix + nickname
}

func normalizeNickname(nickname string) (string, bool) {
	trimmed := strings.TrimSpace(nickname)
	if trimmed == "" {
		return "", false
	}
	return trimmed, true
}
