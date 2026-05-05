package httpapi

import (
	"errors"
	"maps"
	"strconv"
	"strings"

	"github.com/bytedance/sonic"
	"google.golang.org/protobuf/proto"

	"long/internal/core"
	"long/internal/realtimepb"
)

type realtimePublicDeltaPayload struct {
	TotalVotes          int64                       `json:"totalVotes"`
	Leaderboard         *[]core.LeaderboardEntry    `json:"leaderboard,omitempty"`
	RoomID              string                      `json:"roomId,omitempty"`
	Boss                *core.Boss                  `json:"boss,omitempty"`
	BossLeaderboard     []core.BossLeaderboardEntry `json:"bossLeaderboard"`
	AnnouncementVersion string                      `json:"announcementVersion,omitempty"`
}

type realtimePublicMetaPayload struct {
	Leaderboard         *[]core.LeaderboardEntry    `json:"leaderboard,omitempty"`
	BossLeaderboard     []core.BossLeaderboardEntry `json:"bossLeaderboard"`
	AnnouncementVersion string                      `json:"announcementVersion,omitempty"`
}

type realtimeUserDeltaPayload struct {
	UserStats                          *core.UserStats           `json:"userStats,omitempty"`
	MyBossStats                        *core.BossUserStats       `json:"myBossStats,omitempty"`
	MyBossKills                        int64                     `json:"myBossKills"`
	TotalBossKills                     int64                     `json:"totalBossKills"`
	RoomID                             string                    `json:"roomId,omitempty"`
	Loadout                            *core.Loadout             `json:"loadout,omitempty"`
	CombatStats                        *core.CombatStats         `json:"combatStats,omitempty"`
	Gold                               int64                     `json:"gold"`
	Stones                             int64                     `json:"stones"`
	TalentPoints                       int64                     `json:"talentPoints"`
	RecentRewards                      []core.Reward             `json:"recentRewards,omitempty"`
	TalentEvents                       []core.TalentTriggerEvent `json:"talentEvents,omitempty"`
	TalentCombatState                  *core.TalentCombatState   `json:"talentCombatState,omitempty"`
	EquippedBattleClickSkinID          string                    `json:"equippedBattleClickSkinId,omitempty"`
	EquippedBattleClickCursorImagePath string                    `json:"equippedBattleClickCursorImagePath,omitempty"`
}

type realtimeRoomStatePayload struct {
	CurrentRoomID                  string          `json:"currentRoomId"`
	SwitchCooldownRemainingSeconds int64           `json:"switchCooldownRemainingSeconds"`
	Rooms                          []core.RoomInfo `json:"rooms"`
}

type realtimeRoomStateCompatPayload struct {
	CurrentRoomID                  string                        `json:"currentRoomId"`
	SwitchCooldownRemainingSeconds int64                         `json:"switchCooldownRemainingSeconds"`
	Rooms                          []realtimeRoomInfoCompatEntry `json:"rooms"`
}

type realtimeRoomInfoCompatEntry struct {
	ID                       string `json:"id"`
	DisplayName              string `json:"displayName"`
	Current                  bool   `json:"current"`
	Joinable                 bool   `json:"joinable"`
	OnlineCount              int    `json:"onlineCount"`
	CycleEnabled             bool   `json:"cycleEnabled"`
	QueueID                  string `json:"queueId"`
	CurrentBossID            string `json:"currentBossId,omitempty"`
	CurrentBossName          string `json:"currentBossName,omitempty"`
	CurrentBossStatus        string `json:"currentBossStatus,omitempty"`
	CurrentBossHP            string `json:"currentBossHp,omitempty"`
	CurrentBossMaxHP         string `json:"currentBossMaxHp,omitempty"`
	CurrentBossAvgHP         string `json:"currentBossAvgHp,omitempty"`
	CooldownRemainingSeconds int64  `json:"cooldownRemainingSeconds,omitempty"`
}

const (
	realtimeBinaryTypeClickRequest byte = 1
	realtimeBinaryTypeClickAck     byte = 2
	realtimeBinaryTypePublicDelta  byte = 3
	realtimeBinaryTypeUserDelta    byte = 4
	realtimeBinaryTypeRoomState    byte = 5
	realtimeBinaryTypePublicMeta   byte = 6
)

var errRealtimeBinaryFrameInvalid = errors.New("invalid realtime binary frame")

func packRealtimeBinaryMessage(messageType byte, message proto.Message) ([]byte, error) {
	body, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	frame := make([]byte, 1+len(body))
	frame[0] = messageType
	copy(frame[1:], body)
	return frame, nil
}

func unpackRealtimeBinaryMessage(frame []byte, messageType byte, message proto.Message) error {
	if len(frame) < 1 || frame[0] != messageType {
		return errRealtimeBinaryFrameInvalid
	}
	return proto.Unmarshal(frame[1:], message)
}

func decodeRealtimeBinaryClickRequest(frame []byte) (*realtimepb.ClickRequest, error) {
	message := &realtimepb.ClickRequest{}
	if err := unpackRealtimeBinaryMessage(frame, realtimeBinaryTypeClickRequest, message); err != nil {
		return nil, err
	}
	return message, nil
}

func encodeRealtimeBinaryClickAck(payload realtimeClickAckPayload) ([]byte, error) {
	return packRealtimeBinaryMessage(realtimeBinaryTypeClickAck, &realtimepb.ClickAck{
		Delta:                payload.Delta,
		Critical:             payload.Critical,
		BossDamage:           payload.BossDamage,
		MyBossDamage:         payload.MyBossDamage,
		BossLeaderboardCount: int32(payload.BossLeaderboardCount),
		DamageType:           payload.DamageType,
		TalentEvents:         toProtoTalentTriggerEvents(payload.TalentEvents),
		PartStateDeltas:      toProtoBossPartStateDeltas(payload.PartStateDeltas),
		TalentCombatState:    toProtoTalentCombatState(payload.TalentCombatState),
		UserDelta:            toProtoUserDeltaPatch(payload.UserDelta),
		Button: &realtimepb.ButtonRef{
			Key: payload.Button.Key,
		},
	})
}

func encodeRealtimeBinaryPublicDelta(payload realtimePublicDeltaPayload) ([]byte, error) {
	return packRealtimeBinaryMessage(realtimeBinaryTypePublicDelta, &realtimepb.PublicDelta{
		TotalVotes:          payload.TotalVotes,
		Leaderboard:         toProtoLeaderboardEntries(payload.Leaderboard),
		RoomId:              payload.RoomID,
		Boss:                toProtoBoss(payload.Boss),
		BossLeaderboard:     toProtoBossLeaderboardEntries(payload.BossLeaderboard),
		AnnouncementVersion: payload.AnnouncementVersion,
	})
}

func encodeRealtimeBinaryPublicMeta(payload realtimePublicMetaPayload) ([]byte, error) {
	return packRealtimeBinaryMessage(realtimeBinaryTypePublicMeta, &realtimepb.PublicMeta{
		Leaderboard:         toProtoLeaderboardEntries(payload.Leaderboard),
		BossLeaderboard:     toProtoBossLeaderboardEntries(payload.BossLeaderboard),
		AnnouncementVersion: payload.AnnouncementVersion,
	})
}

func encodeRealtimeBinaryUserDelta(payload realtimeUserDeltaPayload) ([]byte, error) {
	return packRealtimeBinaryMessage(realtimeBinaryTypeUserDelta, &realtimepb.UserDelta{
		UserStats:                          toProtoUserStats(payload.UserStats),
		MyBossStats:                        toProtoBossUserStats(payload.MyBossStats),
		MyBossKills:                        payload.MyBossKills,
		TotalBossKills:                     payload.TotalBossKills,
		RoomId:                             payload.RoomID,
		Loadout:                            toProtoLoadout(payload.Loadout),
		CombatStats:                        toProtoCombatStats(payload.CombatStats),
		Gold:                               payload.Gold,
		Stones:                             payload.Stones,
		TalentPoints:                       payload.TalentPoints,
		RecentRewards:                      toProtoRewards(payload.RecentRewards),
		TalentEvents:                       toProtoTalentTriggerEvents(payload.TalentEvents),
		TalentCombatState:                  toProtoTalentCombatState(payload.TalentCombatState),
		EquippedBattleClickSkinId:          payload.EquippedBattleClickSkinID,
		EquippedBattleClickCursorImagePath: payload.EquippedBattleClickCursorImagePath,
	})
}

func encodeRealtimeBinaryRoomState(payload realtimeRoomStatePayload) ([]byte, error) {
	return packRealtimeBinaryMessage(realtimeBinaryTypeRoomState, &realtimepb.RoomState{
		CurrentRoomId:                  payload.CurrentRoomID,
		SwitchCooldownRemainingSeconds: payload.SwitchCooldownRemainingSeconds,
		Rooms:                          toProtoRoomInfos(payload.Rooms),
	})
}

func encodeRealtimeBinaryPublicDeltaFromJSON(raw []byte) ([]byte, error) {
	var payload realtimePublicDeltaPayload
	if err := sonic.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	return encodeRealtimeBinaryPublicDelta(payload)
}

func encodeRealtimeBinaryPublicMetaFromJSON(raw []byte) ([]byte, error) {
	var payload realtimePublicMetaPayload
	if err := sonic.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	return encodeRealtimeBinaryPublicMeta(payload)
}

func encodeRealtimeBinaryUserDeltaFromJSON(raw []byte) ([]byte, error) {
	var payload realtimeUserDeltaPayload
	if err := sonic.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	return encodeRealtimeBinaryUserDelta(payload)
}

func encodeRealtimeBinaryRoomStateFromJSON(raw []byte) ([]byte, error) {
	var payload realtimeRoomStateCompatPayload
	if err := sonic.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	rooms := make([]core.RoomInfo, 0, len(payload.Rooms))
	for _, room := range payload.Rooms {
		currentBossHP, err := parseRealtimeFlexibleInt64Value(room.CurrentBossHP)
		if err != nil {
			return nil, err
		}
		currentBossMaxHP, err := parseRealtimeFlexibleInt64Value(room.CurrentBossMaxHP)
		if err != nil {
			return nil, err
		}
		currentBossAvgHP, err := parseRealtimeFlexibleInt64Value(room.CurrentBossAvgHP)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, core.RoomInfo{
			ID:                 room.ID,
			DisplayName:        room.DisplayName,
			Current:            room.Current,
			Joinable:           room.Joinable,
			OnlineCount:        room.OnlineCount,
			CycleEnabled:       room.CycleEnabled,
			QueueID:            room.QueueID,
			CurrentBossID:      room.CurrentBossID,
			CurrentBossName:    room.CurrentBossName,
			CurrentBossStatus:  room.CurrentBossStatus,
			CurrentBossHP:      currentBossHP,
			CurrentBossMaxHP:   currentBossMaxHP,
			CurrentBossAvgHP:   currentBossAvgHP,
			CooldownRemainingS: room.CooldownRemainingSeconds,
		})
	}
	return encodeRealtimeBinaryRoomState(realtimeRoomStatePayload{
		CurrentRoomID:                  payload.CurrentRoomID,
		SwitchCooldownRemainingSeconds: payload.SwitchCooldownRemainingSeconds,
		Rooms:                          rooms,
	})
}

func parseRealtimeFlexibleInt64Value(raw string) (int64, error) {
	normalized := strings.TrimSpace(raw)
	if normalized == "" {
		return 0, nil
	}
	return strconv.ParseInt(normalized, 10, 64)
}

func toProtoUserDeltaPatch(delta *realtimeUserDelta) *realtimepb.UserDeltaPatch {
	if delta == nil {
		return nil
	}
	message := &realtimepb.UserDeltaPatch{}
	if delta.Gold != nil {
		message.Gold = *delta.Gold
	}
	if delta.Stones != nil {
		message.Stones = *delta.Stones
	}
	if delta.TalentPoints != nil {
		message.TalentPoints = *delta.TalentPoints
	}
	return message
}

func toProtoUserStats(stats *core.UserStats) *realtimepb.UserStats {
	if stats == nil {
		return nil
	}
	return &realtimepb.UserStats{
		Nickname:   stats.Nickname,
		ClickCount: stats.ClickCount,
	}
}

func toProtoLeaderboardEntries(entries *[]core.LeaderboardEntry) []*realtimepb.LeaderboardEntry {
	if entries == nil {
		return nil
	}
	result := make([]*realtimepb.LeaderboardEntry, 0, len(*entries))
	for _, entry := range *entries {
		result = append(result, &realtimepb.LeaderboardEntry{
			Rank:       int32(entry.Rank),
			Nickname:   entry.Nickname,
			ClickCount: entry.ClickCount,
		})
	}
	return result
}

func toProtoBossPart(part core.BossPart) *realtimepb.BossPart {
	return &realtimepb.BossPart{
		X:           int32(part.X),
		Y:           int32(part.Y),
		Type:        string(part.Type),
		DisplayName: part.DisplayName,
		ImagePath:   part.ImagePath,
		MaxHp:       part.MaxHP,
		CurrentHp:   part.CurrentHP,
		Armor:       part.Armor,
		Alive:       part.Alive,
	}
}

func toProtoBoss(boss *core.Boss) *realtimepb.Boss {
	if boss == nil {
		return nil
	}
	parts := make([]*realtimepb.BossPart, 0, len(boss.Parts))
	for _, part := range boss.Parts {
		parts = append(parts, toProtoBossPart(part))
	}
	return &realtimepb.Boss{
		Id:                 boss.ID,
		TemplateId:         boss.TemplateID,
		RoomId:             boss.RoomID,
		QueueId:            boss.QueueID,
		Name:               boss.Name,
		Status:             boss.Status,
		MaxHp:              boss.MaxHP,
		CurrentHp:          boss.CurrentHP,
		GoldOnKill:         boss.GoldOnKill,
		StoneOnKill:        boss.StoneOnKill,
		TalentPointsOnKill: boss.TalentPointsOnKill,
		Parts:              parts,
		StartedAt:          boss.StartedAt,
		DefeatedAt:         boss.DefeatedAt,
	}
}

func toProtoBossLeaderboardEntries(entries []core.BossLeaderboardEntry) []*realtimepb.BossLeaderboardEntry {
	result := make([]*realtimepb.BossLeaderboardEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, &realtimepb.BossLeaderboardEntry{
			Rank:     int32(entry.Rank),
			Nickname: entry.Nickname,
			Damage:   entry.Damage,
		})
	}
	return result
}

func toProtoBossUserStats(stats *core.BossUserStats) *realtimepb.BossUserStats {
	if stats == nil {
		return nil
	}
	return &realtimepb.BossUserStats{
		Nickname: stats.Nickname,
		Damage:   stats.Damage,
		Rank:     int32(stats.Rank),
	}
}

func toProtoRoomInfos(rooms []core.RoomInfo) []*realtimepb.RoomInfo {
	result := make([]*realtimepb.RoomInfo, 0, len(rooms))
	for _, room := range rooms {
		result = append(result, &realtimepb.RoomInfo{
			Id:                       room.ID,
			DisplayName:              room.DisplayName,
			Current:                  room.Current,
			Joinable:                 room.Joinable,
			OnlineCount:              int32(room.OnlineCount),
			CycleEnabled:             room.CycleEnabled,
			QueueId:                  room.QueueID,
			CurrentBossId:            room.CurrentBossID,
			CurrentBossName:          room.CurrentBossName,
			CurrentBossStatus:        room.CurrentBossStatus,
			CurrentBossHp:            room.CurrentBossHP,
			CurrentBossMaxHp:         room.CurrentBossMaxHP,
			CurrentBossAvgHp:         room.CurrentBossAvgHP,
			CooldownRemainingSeconds: room.CooldownRemainingS,
		})
	}
	return result
}

func toProtoInventoryItem(item *core.InventoryItem) *realtimepb.InventoryItem {
	if item == nil {
		return nil
	}
	return &realtimepb.InventoryItem{
		ItemId:               item.ItemID,
		InstanceId:           item.InstanceID,
		Name:                 item.Name,
		Slot:                 item.Slot,
		Rarity:               item.Rarity,
		ImagePath:            item.ImagePath,
		ImageAlt:             item.ImageAlt,
		Quantity:             item.Quantity,
		Equipped:             item.Equipped,
		EnhanceLevel:         int32(item.EnhanceLevel),
		Bound:                item.Bound,
		Locked:               item.Locked,
		AttackPower:          item.AttackPower,
		ArmorPenPercent:      item.ArmorPenPercent,
		CritRate:             item.CritRate,
		CritDamageMultiplier: item.CritDamageMultiplier,
		PartTypeDamageSoft:   item.PartTypeDamageSoft,
		PartTypeDamageHeavy:  item.PartTypeDamageHeavy,
		PartTypeDamageWeak:   item.PartTypeDamageWeak,
	}
}

func toProtoLoadout(loadout *core.Loadout) *realtimepb.Loadout {
	if loadout == nil {
		return nil
	}
	return &realtimepb.Loadout{
		Weapon:    toProtoInventoryItem(loadout.Weapon),
		Helmet:    toProtoInventoryItem(loadout.Helmet),
		Chest:     toProtoInventoryItem(loadout.Chest),
		Gloves:    toProtoInventoryItem(loadout.Gloves),
		Legs:      toProtoInventoryItem(loadout.Legs),
		Accessory: toProtoInventoryItem(loadout.Accessory),
	}
}

func toProtoCombatStats(stats *core.CombatStats) *realtimepb.CombatStats {
	if stats == nil {
		return nil
	}
	return &realtimepb.CombatStats{
		EffectiveIncrement:    stats.EffectiveIncrement,
		NormalDamage:          stats.NormalDamage,
		CriticalChancePercent: stats.CriticalChancePercent,
		CriticalDamage:        stats.CriticalDamage,
		AttackPower:           stats.AttackPower,
		ArmorPenPercent:       stats.ArmorPenPercent,
		CritDamageMultiplier:  stats.CritDamageMultiplier,
		AllDamageAmplify:      stats.AllDamageAmplify,
		PartTypeDamageSoft:    stats.PartTypeDamageSoft,
		PartTypeDamageHeavy:   stats.PartTypeDamageHeavy,
		PartTypeDamageWeak:    stats.PartTypeDamageWeak,
		PerPartDamagePercent:  stats.PerPartDamagePercent,
		LowHpMultiplier:       stats.LowHpMultiplier,
		LowHpThreshold:        stats.LowHpThreshold,
	}
}

func toProtoRewards(rewards []core.Reward) []*realtimepb.Reward {
	result := make([]*realtimepb.Reward, 0, len(rewards))
	for _, reward := range rewards {
		result = append(result, &realtimepb.Reward{
			BossId:    reward.BossID,
			BossName:  reward.BossName,
			ItemId:    reward.ItemID,
			ItemName:  reward.ItemName,
			GrantedAt: reward.GrantedAt,
		})
	}
	return result
}

func toProtoTalentTriggerEvents(events []core.TalentTriggerEvent) []*realtimepb.TalentTriggerEvent {
	result := make([]*realtimepb.TalentTriggerEvent, 0, len(events))
	for _, event := range events {
		result = append(result, &realtimepb.TalentTriggerEvent{
			TalentId:    event.TalentID,
			Name:        event.Name,
			EffectType:  event.EffectType,
			ExtraDamage: event.ExtraDamage,
			Message:     event.Message,
			PartX:       int32(event.PartX),
			PartY:       int32(event.PartY),
		})
	}
	return result
}

func toProtoBossPartStateDeltas(deltas []core.BossPartStateDelta) []*realtimepb.BossPartStateDelta {
	result := make([]*realtimepb.BossPartStateDelta, 0, len(deltas))
	for _, delta := range deltas {
		result = append(result, &realtimepb.BossPartStateDelta{
			X:        int32(delta.X),
			Y:        int32(delta.Y),
			Damage:   delta.Damage,
			BeforeHp: delta.BeforeHP,
			AfterHp:  delta.AfterHP,
			PartType: delta.PartType,
		})
	}
	return result
}

func toProtoTalentCombatState(state *core.TalentCombatState) *realtimepb.TalentCombatState {
	if state == nil {
		return nil
	}
	bleeds := make(map[string]*realtimepb.TalentBleedState, len(state.Bleeds))
	for key, bleed := range state.Bleeds {
		bleeds[key] = &realtimepb.TalentBleedState{
			StartedAtMs:    bleed.StartedAtMs,
			NextTickAtMs:   bleed.NextTickAtMs,
			EndsAtMs:       bleed.EndsAtMs,
			DurationMs:     bleed.DurationMs,
			TickIntervalMs: bleed.TickIntervalMs,
			TotalTicks:     bleed.TotalTicks,
			AppliedTicks:   bleed.AppliedTicks,
			TotalDamage:    bleed.TotalDamage,
			AppliedDamage:  bleed.AppliedDamage,
		}
	}
	return &realtimepb.TalentCombatState{
		OmenStacks:              int32(state.OmenStacks),
		Bleeds:                  bleeds,
		CollapseParts:           toInt32Slice(state.CollapseParts),
		CollapseEndsAt:          state.CollapseEndsAt,
		CollapseDuration:        state.CollapseDuration,
		DoomMarks:               toInt32Slice(state.DoomMarks),
		HasTriggeredDoom:        state.HasTriggeredDoom,
		DoomMarkCumDamage:       cloneInt64Map(state.DoomMarkCumDamage),
		SilverStormRemaining:    int32(state.SilverStormRemaining),
		SilverStormEndsAt:       state.SilverStormEndsAt,
		SilverStormActive:       state.SilverStormActive,
		AutoStrikeTargetPart:    state.AutoStrikeTargetPart,
		AutoStrikeComboCount:    state.AutoStrikeComboCount,
		AutoStrikeExpiresAt:     state.AutoStrikeExpiresAt,
		LastFinalCutAt:          state.LastFinalCutAt,
		JudgmentDayUsed:         cloneInt64Map(state.JudgmentDayUsed),
		JudgmentDayCooldownSec:  state.JudgmentDayCooldownSec,
		PartHeavyClickCount:     cloneInt64Map(state.PartHeavyClickCount),
		PartJudgmentDayCount:    cloneInt64Map(state.PartJudgmentDayCount),
		PartRetainedClicks:      cloneInt64Map(state.PartRetainedClicks),
		PartStormComboCount:     cloneInt64Map(state.PartStormComboCount),
		SkinnerParts:            cloneInt64Map(state.SkinnerParts),
		SkinnerDurationByPart:   cloneInt64Map(state.SkinnerDurationByPart),
		SkinnerCooldownEndsAt:   state.SkinnerCooldownEndsAt,
		SkinnerCooldownDuration: state.SkinnerCooldownDuration,
		NormalTriggerCount:      state.NormalTriggerCount,
		ArmorTriggerCount:       state.ArmorTriggerCount,
		JudgmentDayTriggerCount: state.JudgmentDayTriggerCount,
		AutoStrikeTriggerCount:  state.AutoStrikeTriggerCount,
		AutoStrikeWindowSec:     state.AutoStrikeWindowSec,
	}
}

func toInt32Slice(values []int) []int32 {
	result := make([]int32, 0, len(values))
	for _, value := range values {
		result = append(result, int32(value))
	}
	return result
}

func cloneInt64Map(values map[string]int64) map[string]int64 {
	if len(values) == 0 {
		return nil
	}
	result := make(map[string]int64, len(values))
	maps.Copy(result, values)
	return result
}
