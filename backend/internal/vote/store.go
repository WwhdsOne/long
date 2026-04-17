package vote

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"long/internal/config"
	nicknamefilter "long/internal/nickname"
)

// 错误定义
var ErrButtonNotFound = errors.New("button not found")
var ErrInvalidNickname = errors.New("invalid nickname")
var ErrSensitiveNickname = errors.New("sensitive nickname")

// Button 按钮数据结构，返回给前端和 SSE 客户端
type Button struct {
	Key       string `json:"key"`                 // 按钮标识
	RedisKey  string `json:"redisKey"`            // Redis 键名
	Label     string `json:"label"`               // 显示名称
	Count     int64  `json:"count"`               // 当前票数
	Sort      int    `json:"sort"`                // 排序权重
	Enabled   bool   `json:"enabled"`             // 是否启用
	ImagePath string `json:"imagePath,omitempty"` // 图片路径
	ImageAlt  string `json:"imageAlt,omitempty"`  // 图片描述
}

// UserStats 用户统计信息
type UserStats struct {
	Nickname   string `json:"nickname"`   // 昵称
	ClickCount int64  `json:"clickCount"` // 累计点击数
}

// LeaderboardEntry 排行榜条目
type LeaderboardEntry struct {
	Rank       int    `json:"rank"`       // 排名
	Nickname   string `json:"nickname"`   // 昵称
	ClickCount int64  `json:"clickCount"` // 点击数
}

// Snapshot 公共实时状态，广播给所有连接的客户端
type Snapshot struct {
	Buttons     []Button           `json:"buttons"`     // 按钮列表
	Leaderboard []LeaderboardEntry `json:"leaderboard"` // 排行榜
}

// State 完整状态，包含个人统计信息
type State struct {
	Buttons     []Button           `json:"buttons"`             // 按钮列表
	Leaderboard []LeaderboardEntry `json:"leaderboard"`         // 排行榜
	UserStats   *UserStats         `json:"userStats,omitempty"` // 个人统计
}

// ClickResult 点击结果，包含更新后的快照和增量信息
type ClickResult struct {
	Button    Button    `json:"button"`    // 更新后的按钮
	Delta     int64     `json:"delta"`     // 本次增量（普通点击为1，暴击为更多）
	Critical  bool      `json:"critical"`  // 是否触发暴击
	UserStats UserStats `json:"userStats"` // 更新后的用户统计
}

// StoreOptions 暴击机制配置
type StoreOptions struct {
	CriticalChancePercent int   // 暴击概率（百分比）
	CriticalCount         int64 // 暴击时的增量
}

// buttonFallback 按钮回退数据（用于图片等元数据）
type buttonFallback struct {
	Label     string
	ImagePath string
	ImageAlt  string
}

// Store Redis 投票存储，管理按钮列表、点击计数和排行榜
type Store struct {
	client         redis.UniversalClient
	prefix         string                    // 按钮键前缀
	userPrefix     string                    // 用户键前缀
	leaderboardKey string                    // 排行榜键名
	fallbacks      map[string]buttonFallback // 回退数据
	critical       StoreOptions              // 暴击配置
	roll           func(int) int             // 随机数生成器
	validator      interface{ Validate(string) error }
}

// hashFields Redis Hash 中存储的字段列表
var hashFields = []string{
	"label",
	"count",
	"sort",
	"enabled",
	"image_path",
	"image_alt",
}

// NewStore 创建 Redis 投票存储实例
func NewStore(client redis.UniversalClient, prefix string, options StoreOptions, validator interface{ Validate(string) error }) *Store {
	namespace := deriveNamespace(prefix)

	return &Store{
		client:         client,
		prefix:         prefix,
		userPrefix:     namespace + "user:",
		leaderboardKey: namespace + "leaderboard",
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
		validator: validator,
	}
}

// ValidateNickname checks whether the provided nickname is usable.
func (s *Store) ValidateNickname(_ context.Context, nickname string) error {
	_, err := s.validatedNickname(nickname)
	return err
}

// GetSnapshot 获取公共快照（按钮列表 + 排行榜）
func (s *Store) GetSnapshot(ctx context.Context) (Snapshot, error) {
	buttons, err := s.ListButtons(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	leaderboard, err := s.ListLeaderboard(ctx, 10)
	if err != nil {
		return Snapshot{}, err
	}

	return Snapshot{
		Buttons:     buttons,
		Leaderboard: leaderboard,
	}, nil
}

// GetState 获取完整状态（公共快照 + 个人统计）
func (s *Store) GetState(ctx context.Context, nickname string) (State, error) {
	snapshot, err := s.GetSnapshot(ctx)
	if err != nil {
		return State{}, err
	}

	state := State{
		Buttons:     snapshot.Buttons,
		Leaderboard: snapshot.Leaderboard,
	}

	trimmedNickname, hasNickname := normalizeNickname(nickname)
	if !hasNickname {
		return state, nil
	}

	normalizedNickname, err := s.validatedNickname(trimmedNickname)
	if err != nil {
		return State{}, err
	}

	userStats, err := s.GetUserStats(ctx, normalizedNickname)
	if err != nil {
		return State{}, err
	}
	state.UserStats = &userStats

	return state, nil
}

// ListButtons 扫描 Redis，过滤禁用按钮，按排序权重返回
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

// ClickButton 处理按钮点击，支持普通点击和暴击
func (s *Store) ClickButton(ctx context.Context, slug string, nickname string) (ClickResult, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return ClickResult{}, err
	}

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
	pipe := s.client.TxPipeline()
	pipe.HIncrBy(ctx, redisKey, "count", delta)
	userCountCmd := pipe.HIncrBy(ctx, s.userPrefix+normalizedNickname, "click_count", delta)
	pipe.HSet(ctx, s.userPrefix+normalizedNickname, map[string]any{
		"nickname":   normalizedNickname,
		"updated_at": strconv.FormatInt(time.Now().Unix(), 10),
	})
	pipe.ZIncrBy(ctx, s.leaderboardKey, float64(delta), normalizedNickname)

	if _, err := pipe.Exec(ctx); err != nil {
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
		UserStats: UserStats{
			Nickname:   normalizedNickname,
			ClickCount: userCountCmd.Val(),
		},
	}, nil
}

// EnsureDefaults 在 Redis 为空时初始化默认按钮
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

// normalizeButton 将 Redis 数据转换为 Button 结构
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

// scanKeys 扫描 Redis 中匹配前缀的所有键
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

// nextIncrement 计算本次点击的增量值和是否暴击
func (s *Store) nextIncrement() (int64, bool) {
	if s.critical.CriticalChancePercent <= 0 || s.critical.CriticalCount <= 1 {
		return 1, false
	}

	if s.roll(100) < s.critical.CriticalChancePercent {
		return s.critical.CriticalCount, true
	}

	return 1, false
}

// ListLeaderboard 获取排行榜前 N 名
func (s *Store) ListLeaderboard(ctx context.Context, limit int64) ([]LeaderboardEntry, error) {
	if limit <= 0 {
		limit = 10
	}

	scores, err := s.client.ZRevRangeWithScores(ctx, s.leaderboardKey, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	leaderboard := make([]LeaderboardEntry, 0, len(scores))
	for index, score := range scores {
		nickname, ok := score.Member.(string)
		if !ok || nickname == "" {
			continue
		}

		leaderboard = append(leaderboard, LeaderboardEntry{
			Rank:       index + 1,
			Nickname:   nickname,
			ClickCount: int64(score.Score),
		})
	}

	return leaderboard, nil
}

// GetUserStats 获取指定用户的统计信息
func (s *Store) GetUserStats(ctx context.Context, nickname string) (UserStats, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return UserStats{}, err
	}

	values, err := s.client.HMGet(ctx, s.userPrefix+normalizedNickname, "nickname", "click_count").Result()
	if err != nil {
		return UserStats{}, err
	}

	return UserStats{
		Nickname:   normalizedNickname,
		ClickCount: int64Value(values, 1),
	}, nil
}

func (s *Store) validatedNickname(nickname string) (string, error) {
	normalizedNickname, ok := normalizeNickname(nickname)
	if !ok {
		return "", ErrInvalidNickname
	}

	if s.validator != nil {
		if err := s.validator.Validate(normalizedNickname); err != nil {
			if errors.Is(err, nicknamefilter.ErrSensitiveNickname) {
				return "", ErrSensitiveNickname
			}
			return "", err
		}
	}

	return normalizedNickname, nil
}

// deriveNamespace 从按钮前缀推导命名空间
func deriveNamespace(prefix string) string {
	if strings.HasSuffix(prefix, "button:") {
		return strings.TrimSuffix(prefix, "button:")
	}

	return prefix
}

// normalizeNickname 规范化昵称（去除首尾空格）
func normalizeNickname(nickname string) (string, bool) {
	trimmed := strings.TrimSpace(nickname)
	if trimmed == "" {
		return "", false
	}

	return trimmed, true
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
