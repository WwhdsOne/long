package core

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const fallbackRoomID = "1"
const hallRoomID = "hall"

var ErrRoomNotFound = errors.New("room not found")
var ErrRoomSwitchCooldown = errors.New("room switch cooldown")
var ErrRoomNotJoinable = errors.New("room not joinable")

// RoomConfig 控制可加入房间列表与切房冷却。
type RoomConfig struct {
	Enabled        bool
	Count          int
	DefaultRoom    string
	SwitchCooldown time.Duration
}

// RoomInfo 描述一个可选房间的当前运行态。
type RoomInfo struct {
	ID                 string `json:"id"`
	DisplayName        string `json:"displayName"`
	Current            bool   `json:"current"`
	Joinable           bool   `json:"joinable"`
	OnlineCount        int    `json:"onlineCount"`
	CycleEnabled       bool   `json:"cycleEnabled"`
	QueueID            string `json:"queueId"`
	CurrentBossID      string `json:"currentBossId,omitempty"`
	CurrentBossName    string `json:"currentBossName,omitempty"`
	CurrentBossStatus  string `json:"currentBossStatus,omitempty"`
	CurrentBossHP      int64  `json:"currentBossHp,omitempty"`
	CurrentBossMaxHP   int64  `json:"currentBossMaxHp,omitempty"`
	CurrentBossAvgHP   int64  `json:"currentBossAvgHp,omitempty"`
	CooldownRemainingS int64  `json:"cooldownRemainingSeconds,omitempty"`
}

// RoomList 是玩家视角的房间列表。
type RoomList struct {
	CurrentRoomID                  string     `json:"currentRoomId"`
	SwitchCooldownRemainingSeconds int64      `json:"switchCooldownRemainingSeconds"`
	Rooms                          []RoomInfo `json:"rooms"`
}

// RoomSwitchResult 是切房成功后的返回载荷。
type RoomSwitchResult struct {
	CurrentRoomID            string     `json:"currentRoomId"`
	CooldownUntil            int64      `json:"cooldownUntil,omitempty"`
	CooldownRemainingSeconds int64      `json:"cooldownRemainingSeconds,omitempty"`
	Rooms                    []RoomInfo `json:"rooms"`
}

func normalizeRoomConfig(cfg RoomConfig) RoomConfig {
	if cfg.Count <= 0 {
		cfg.Count = 1
	}
	defaultRoom := strings.TrimSpace(cfg.DefaultRoom)
	if !slices.Contains(configuredRoomIDsFromCount(cfg.Count), defaultRoom) {
		defaultRoom = fallbackRoomID
	}
	if cfg.SwitchCooldown < 0 {
		cfg.SwitchCooldown = 0
	}
	cfg.DefaultRoom = defaultRoom
	return cfg
}

func (s *Store) configuredRoomIDs() []string {
	return configuredRoomIDsFromCount(s.roomConfig.Count)
}

func configuredRoomIDsFromCount(count int) []string {
	if count <= 0 {
		count = 1
	}
	ids := make([]string, 0, count)
	for i := 1; i <= count; i++ {
		ids = append(ids, strconv.Itoa(i))
	}
	return ids
}

func (s *Store) defaultRoomID() string {
	roomID := strings.TrimSpace(s.roomConfig.DefaultRoom)
	if roomID == "" {
		return fallbackRoomID
	}
	return roomID
}

func (s *Store) isKnownRoom(roomID string) bool {
	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		return false
	}
	return slices.Contains(s.configuredRoomIDs(), roomID)
}

func isHallRoomID(roomID string) bool {
	return strings.EqualFold(strings.TrimSpace(roomID), hallRoomID)
}

func (s *Store) normalizeRoomID(roomID string) string {
	roomID = strings.TrimSpace(roomID)
	if isHallRoomID(roomID) {
		return hallRoomID
	}
	if s.isKnownRoom(roomID) {
		return roomID
	}
	return s.defaultRoomID()
}

func (s *Store) playerRoomKey(nickname string) string {
	return s.playerRoomPrefix + strings.TrimSpace(nickname)
}

func (s *Store) playerRoomCooldownKey(nickname string) string {
	return s.playerRoomCooldownPrefix + strings.TrimSpace(nickname)
}

func (s *Store) roomNamesKey() string {
	return s.namespace + "room:names"
}

func (s *Store) bossCurrentKeyForRoom(roomID string) string {
	return s.bossCurrentKey + ":" + s.normalizeRoomID(roomID)
}

func (s *Store) bossCycleKeyForRoom(roomID string) string {
	return s.bossCycleKey + ":" + s.normalizeRoomID(roomID)
}

func (s *Store) queueIDForRoom(roomID string) string {
	return s.normalizeRoomID(roomID)
}

func (s *Store) combatRoomID(roomID string) string {
	if isHallRoomID(roomID) {
		return s.defaultRoomID()
	}
	return s.normalizeRoomID(roomID)
}

func defaultRoomDisplayName(roomID string) string {
	return "房间 " + strings.TrimSpace(roomID)
}

func (s *Store) GetRoomDisplayName(ctx context.Context, roomID string) (string, error) {
	roomID = strings.TrimSpace(roomID)
	if !s.isKnownRoom(roomID) {
		return "", ErrRoomNotFound
	}
	displayName, err := s.client.HGet(ctx, s.roomNamesKey(), roomID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return defaultRoomDisplayName(roomID), nil
		}
		return "", err
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return defaultRoomDisplayName(roomID), nil
	}
	return displayName, nil
}

func (s *Store) SetRoomDisplayName(ctx context.Context, roomID string, displayName string) error {
	roomID = strings.TrimSpace(roomID)
	if !s.isKnownRoom(roomID) {
		return ErrRoomNotFound
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return s.client.HDel(ctx, s.roomNamesKey(), roomID).Err()
	}
	return s.client.HSet(ctx, s.roomNamesKey(), roomID, displayName).Err()
}

func (s *Store) ResolvePlayerRoom(ctx context.Context, nickname string) (string, error) {
	normalizedNickname, ok := normalizeNickname(nickname)
	if !ok {
		return hallRoomID, nil
	}
	value, err := s.client.Get(ctx, s.playerRoomKey(normalizedNickname)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return hallRoomID, nil
		}
		return "", err
	}
	return s.normalizeRoomID(value), nil
}

func (s *Store) ListRooms(ctx context.Context, nickname string) (RoomList, error) {
	currentRoomID, err := s.ResolvePlayerRoom(ctx, nickname)
	if err != nil {
		return RoomList{}, err
	}
	remaining, err := s.roomSwitchCooldownRemaining(ctx, nickname)
	if err != nil {
		return RoomList{}, err
	}
	rooms, err := s.roomInfos(ctx, currentRoomID, remaining)
	if err != nil {
		return RoomList{}, err
	}
	return RoomList{
		CurrentRoomID:                  currentRoomID,
		SwitchCooldownRemainingSeconds: remaining,
		Rooms:                          rooms,
	}, nil
}

func (s *Store) SwitchPlayerRoom(ctx context.Context, nickname string, targetRoomID string) (RoomSwitchResult, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return RoomSwitchResult{}, err
	}
	targetRoomID = strings.TrimSpace(targetRoomID)
	if !isHallRoomID(targetRoomID) && !s.isKnownRoom(targetRoomID) {
		return RoomSwitchResult{}, ErrRoomNotFound
	}

	nowUnix := s.now().Unix()
	currentRoomID, err := s.ResolvePlayerRoom(ctx, normalizedNickname)
	if err != nil {
		return RoomSwitchResult{}, err
	}
	cooldownUntil, err := s.roomSwitchCooldownUntil(ctx, normalizedNickname)
	if err != nil {
		return RoomSwitchResult{}, err
	}
	if !isHallRoomID(targetRoomID) && targetRoomID != currentRoomID && cooldownUntil > nowUnix {
		return RoomSwitchResult{}, ErrRoomSwitchCooldown
	}
	if !isHallRoomID(targetRoomID) && targetRoomID != currentRoomID {
		joinable, err := s.roomJoinable(ctx, targetRoomID)
		if err != nil {
			return RoomSwitchResult{}, err
		}
		if !joinable {
			return RoomSwitchResult{}, ErrRoomNotJoinable
		}
	}

	nextCooldownUntil := cooldownUntil
	if nextCooldownUntil <= nowUnix {
		nextCooldownUntil = 0
	}
	pipe := s.client.TxPipeline()
	pipe.Set(ctx, s.playerRoomKey(normalizedNickname), targetRoomID, 0)
	if !isHallRoomID(targetRoomID) && targetRoomID != currentRoomID && s.roomConfig.SwitchCooldown > 0 {
		nextCooldownUntil = nowUnix + int64(s.roomConfig.SwitchCooldown.Seconds())
		pipe.Set(ctx, s.playerRoomCooldownKey(normalizedNickname), strconv.FormatInt(nextCooldownUntil, 10), s.roomConfig.SwitchCooldown)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return RoomSwitchResult{}, err
	}

	remaining := maxInt64(0, nextCooldownUntil-nowUnix)
	rooms, err := s.roomInfos(ctx, targetRoomID, remaining)
	if err != nil {
		return RoomSwitchResult{}, err
	}
	return RoomSwitchResult{
		CurrentRoomID:            targetRoomID,
		CooldownUntil:            nextCooldownUntil,
		CooldownRemainingSeconds: remaining,
		Rooms:                    rooms,
	}, nil
}

func (s *Store) roomSwitchCooldownRemaining(ctx context.Context, nickname string) (int64, error) {
	until, err := s.roomSwitchCooldownUntil(ctx, nickname)
	if err != nil {
		return 0, err
	}
	return maxInt64(0, until-s.now().Unix()), nil
}

func (s *Store) roomSwitchCooldownUntil(ctx context.Context, nickname string) (int64, error) {
	normalizedNickname, ok := normalizeNickname(nickname)
	if !ok {
		return 0, nil
	}
	value, err := s.client.Get(ctx, s.playerRoomCooldownKey(normalizedNickname)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, err
	}
	return int64FromString(value), nil
}

func (s *Store) roomInfos(ctx context.Context, currentRoomID string, cooldownRemaining int64) ([]RoomInfo, error) {
	ids := s.configuredRoomIDs()
	rooms := make([]RoomInfo, 0, len(ids))
	for _, id := range ids {
		boss, err := s.currentBossForRoom(ctx, id)
		if err != nil {
			return nil, err
		}
		onlineCount := 0
		if boss != nil {
			count, err := s.client.ZCard(ctx, s.bossDamageKey(boss.ID)).Result()
			if err != nil {
				return nil, err
			}
			onlineCount = int(count)
		}
		avgBossHP, err := s.roomCycleAverageBossHP(ctx, id)
		if err != nil {
			return nil, err
		}
		enabled, err := s.bossCycleEnabledForRoom(ctx, id)
		if err != nil {
			return nil, err
		}
		displayName, err := s.GetRoomDisplayName(ctx, id)
		if err != nil {
			return nil, err
		}
		info := RoomInfo{
			ID:                 id,
			DisplayName:        displayName,
			Current:            id == currentRoomID,
			Joinable:           enabled || (boss != nil && boss.Status == bossStatusActive),
			OnlineCount:        onlineCount,
			CycleEnabled:       enabled,
			QueueID:            s.queueIDForRoom(id),
			CooldownRemainingS: cooldownRemaining,
		}
		if boss != nil {
			info.CurrentBossID = boss.ID
			info.CurrentBossName = boss.Name
			info.CurrentBossStatus = boss.Status
			info.CurrentBossHP = boss.CurrentHP
			info.CurrentBossMaxHP = boss.MaxHP
		}
		info.CurrentBossAvgHP = avgBossHP
		rooms = append(rooms, info)
	}
	return rooms, nil
}

func (s *Store) roomCycleAverageBossHP(ctx context.Context, roomID string) (int64, error) {
	queue, err := s.loadBossTemplateQueueForRoom(ctx, roomID)
	if err != nil {
		if errors.Is(err, ErrBossCycleQueueEmpty) || errors.Is(err, ErrBossPoolEmpty) {
			return 0, nil
		}
		return 0, err
	}
	var total int64
	for _, template := range queue {
		total += template.MaxHP
	}
	return total / int64(len(queue)), nil
}

func (s *Store) roomJoinable(ctx context.Context, roomID string) (bool, error) {
	boss, err := s.currentBossForRoom(ctx, roomID)
	if err != nil {
		return false, err
	}
	if boss != nil && boss.Status == bossStatusActive {
		return true, nil
	}
	return s.bossCycleEnabledForRoom(ctx, roomID)
}
