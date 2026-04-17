package vote

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"

	"long/internal/config"
)

var ErrButtonNotFound = errors.New("button not found")

// Button is the normalized payload returned to the frontend and SSE clients.
type Button struct {
	Key       string `json:"key"`
	RedisKey  string `json:"redisKey"`
	Label     string `json:"label"`
	Count     int64  `json:"count"`
	Sort      int    `json:"sort"`
	Enabled   bool   `json:"enabled"`
	ImagePath string `json:"imagePath,omitempty"`
	ImageAlt  string `json:"imageAlt,omitempty"`
}

// ClickResult describes the post-click snapshot plus the applied increment.
type ClickResult struct {
	Button   Button `json:"button"`
	Delta    int64  `json:"delta"`
	Critical bool   `json:"critical"`
}

// StoreOptions controls how much a crit adds and how often it happens.
type StoreOptions struct {
	CriticalChancePercent int
	CriticalCount         int64
}

type buttonFallback struct {
	Label     string
	ImagePath string
	ImageAlt  string
}

// Store wraps Redis access for listing, incrementing, and seeding buttons.
type Store struct {
	client    redis.UniversalClient
	prefix    string
	fallbacks map[string]buttonFallback
	critical  StoreOptions
	roll      func(int) int
}

var hashFields = []string{
	"label",
	"count",
	"sort",
	"enabled",
	"image_path",
	"image_alt",
}

// NewStore creates a Redis-backed vote store with fallback image metadata.
func NewStore(client redis.UniversalClient, prefix string, options StoreOptions) *Store {
	return &Store{
		client: client,
		prefix: prefix,
		fallbacks: map[string]buttonFallback{
			"wechat-pity": {
				ImagePath: "/images/emojipedia-wechat-whimper.png",
				ImageAlt:  "微信可怜表情",
			},
		},
		critical: options,
		roll: func(limit int) int {
			return rand.IntN(limit)
		},
	}
}

// ListButtons scans Redis, filters hidden buttons, and returns them in display order.
func (s *Store) ListButtons(ctx context.Context) ([]Button, error) {
	keys, err := s.scanKeys(ctx)
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return []Button{}, nil
	}

	pipe := s.client.Pipeline()
	cmds := make([]*redis.SliceCmd, len(keys))
	for index, redisKey := range keys {
		cmds[index] = pipe.HMGet(ctx, redisKey, hashFields...)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	buttons := make([]Button, 0, len(keys))
	for index, redisKey := range keys {
		button := s.normalizeButton(redisKey, cmds[index].Val())
		if button.Enabled {
			buttons = append(buttons, button)
		}
	}

	slices.SortFunc(buttons, func(left, right Button) int {
		if left.Sort == right.Sort {
			return strings.Compare(left.Key, right.Key)
		}
		if left.Sort < right.Sort {
			return -1
		}
		return 1
	})

	return buttons, nil
}

// ClickButton applies a normal or critical increment and returns the new snapshot.
func (s *Store) ClickButton(ctx context.Context, slug string) (ClickResult, error) {
	redisKey := s.prefix + slug

	exists, err := s.client.Exists(ctx, redisKey).Result()
	if err != nil {
		return ClickResult{}, err
	}
	if exists == 0 {
		return ClickResult{}, ErrButtonNotFound
	}

	currentValues, err := s.client.HMGet(ctx, redisKey, hashFields...).Result()
	if err != nil {
		return ClickResult{}, err
	}

	current := s.normalizeButton(redisKey, currentValues)
	if !current.Enabled {
		return ClickResult{}, ErrButtonNotFound
	}

	delta, critical := s.nextIncrement()
	if _, err := s.client.HIncrBy(ctx, redisKey, "count", delta).Result(); err != nil {
		return ClickResult{}, err
	}

	updatedValues, err := s.client.HMGet(ctx, redisKey, hashFields...).Result()
	if err != nil {
		return ClickResult{}, err
	}

	return ClickResult{
		Button:   s.normalizeButton(redisKey, updatedValues),
		Delta:    delta,
		Critical: critical,
	}, nil
}

// EnsureDefaults seeds the built-in buttons only when Redis is currently empty.
func (s *Store) EnsureDefaults(ctx context.Context, buttons []config.ButtonSeed) error {
	keys, err := s.scanKeys(ctx)
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return nil
	}

	pipe := s.client.Pipeline()
	for _, button := range buttons {
		values := map[string]any{
			"label":   button.Label,
			"count":   "0",
			"sort":    strconv.Itoa(button.Sort),
			"enabled": "1",
		}
		if button.ImagePath != "" {
			values["image_path"] = button.ImagePath
		}
		if button.ImageAlt != "" {
			values["image_alt"] = button.ImageAlt
		}
		pipe.HSet(ctx, s.prefix+button.Slug, values)
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (s *Store) normalizeButton(redisKey string, values []any) Button {
	slug := strings.TrimPrefix(redisKey, s.prefix)
	fallback := s.fallbacks[slug]

	label := stringValue(values, 0)
	if label == "" {
		if fallback.Label != "" {
			label = fallback.Label
		} else {
			label = slug
		}
	}

	imagePath := stringValue(values, 4)
	if imagePath == "" {
		imagePath = fallback.ImagePath
	}

	imageAlt := stringValue(values, 5)
	if imageAlt == "" {
		imageAlt = fallback.ImageAlt
	}

	return Button{
		Key:       slug,
		RedisKey:  redisKey,
		Label:     label,
		Count:     int64Value(values, 1),
		Sort:      int(int64Value(values, 2)),
		Enabled:   stringValue(values, 3) != "0",
		ImagePath: imagePath,
		ImageAlt:  imageAlt,
	}
}

func (s *Store) scanKeys(ctx context.Context) ([]string, error) {
	var (
		cursor uint64
		keys   []string
	)

	for {
		foundKeys, nextCursor, err := s.client.Scan(ctx, cursor, s.prefix+"*", 100).Result()
		if err != nil {
			return nil, err
		}

		keys = append(keys, foundKeys...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

func (s *Store) nextIncrement() (int64, bool) {
	if s.critical.CriticalChancePercent <= 0 || s.critical.CriticalCount <= 1 {
		return 1, false
	}

	if s.roll(100) < s.critical.CriticalChancePercent {
		return s.critical.CriticalCount, true
	}

	return 1, false
}

func stringValue(values []any, index int) string {
	if index >= len(values) || values[index] == nil {
		return ""
	}

	switch value := values[index].(type) {
	case string:
		return value
	case []byte:
		return string(value)
	default:
		return fmt.Sprint(value)
	}
}

func int64Value(values []any, index int) int64 {
	raw := stringValue(values, index)
	if raw == "" {
		return 0
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0
	}

	return value
}
