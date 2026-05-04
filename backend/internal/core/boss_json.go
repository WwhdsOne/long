package core

import (
	"encoding/json"
	"strconv"
	"strings"
)

func parseFlexibleInt64Value(raw json.RawMessage) (int64, error) {
	text := strings.TrimSpace(string(raw))
	if text == "" || text == "null" {
		return 0, nil
	}
	if strings.HasPrefix(text, `"`) {
		unquoted, err := strconv.Unquote(text)
		if err != nil {
			return 0, err
		}
		text = strings.TrimSpace(unquoted)
	}
	if text == "" {
		return 0, nil
	}
	return strconv.ParseInt(text, 10, 64)
}

func formatInt64String(value int64) string {
	return strconv.FormatInt(value, 10)
}

func (p BossPart) MarshalJSON() ([]byte, error) {
	type bossPartJSON struct {
		X           int      `json:"x"`
		Y           int      `json:"y"`
		Type        PartType `json:"type"`
		DisplayName string   `json:"displayName,omitempty"`
		ImagePath   string   `json:"imagePath,omitempty"`
		MaxHP       string   `json:"maxHp"`
		CurrentHP   string   `json:"currentHp"`
		Armor       string   `json:"armor"`
		Alive       bool     `json:"alive"`
	}
	return json.Marshal(bossPartJSON{
		X:           p.X,
		Y:           p.Y,
		Type:        p.Type,
		DisplayName: p.DisplayName,
		ImagePath:   p.ImagePath,
		MaxHP:       formatInt64String(p.MaxHP),
		CurrentHP:   formatInt64String(p.CurrentHP),
		Armor:       formatInt64String(p.Armor),
		Alive:       p.Alive,
	})
}

func (p *BossPart) UnmarshalJSON(data []byte) error {
	type bossPartJSON struct {
		X           int             `json:"x"`
		Y           int             `json:"y"`
		Type        PartType        `json:"type"`
		DisplayName string          `json:"displayName,omitempty"`
		ImagePath   string          `json:"imagePath,omitempty"`
		MaxHP       json.RawMessage `json:"maxHp"`
		CurrentHP   json.RawMessage `json:"currentHp"`
		Armor       json.RawMessage `json:"armor"`
		Alive       bool            `json:"alive"`
	}
	var aux bossPartJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	maxHP, err := parseFlexibleInt64Value(aux.MaxHP)
	if err != nil {
		return err
	}
	currentHP, err := parseFlexibleInt64Value(aux.CurrentHP)
	if err != nil {
		return err
	}
	armor, err := parseFlexibleInt64Value(aux.Armor)
	if err != nil {
		return err
	}
	*p = BossPart{
		X:           aux.X,
		Y:           aux.Y,
		Type:        aux.Type,
		DisplayName: aux.DisplayName,
		ImagePath:   aux.ImagePath,
		MaxHP:       maxHP,
		CurrentHP:   currentHP,
		Armor:       armor,
		Alive:       aux.Alive,
	}
	return nil
}

func (b Boss) MarshalJSON() ([]byte, error) {
	type bossJSON struct {
		ID                 string     `json:"id"`
		TemplateID         string     `json:"templateId,omitempty"`
		RoomID             string     `json:"roomId,omitempty"`
		QueueID            string     `json:"queueId,omitempty"`
		Name               string     `json:"name"`
		Status             string     `json:"status"`
		MaxHP              string     `json:"maxHp"`
		CurrentHP          string     `json:"currentHp"`
		GoldOnKill         int64      `json:"goldOnKill"`
		StoneOnKill        int64      `json:"stoneOnKill"`
		TalentPointsOnKill int64      `json:"talentPointsOnKill"`
		Parts              []BossPart `json:"parts,omitempty"`
		StartedAt          int64      `json:"startedAt,omitempty"`
		DefeatedAt         int64      `json:"defeatedAt,omitempty"`
	}
	return json.Marshal(bossJSON{
		ID:                 b.ID,
		TemplateID:         b.TemplateID,
		RoomID:             b.RoomID,
		QueueID:            b.QueueID,
		Name:               b.Name,
		Status:             b.Status,
		MaxHP:              formatInt64String(b.MaxHP),
		CurrentHP:          formatInt64String(b.CurrentHP),
		GoldOnKill:         b.GoldOnKill,
		StoneOnKill:        b.StoneOnKill,
		TalentPointsOnKill: b.TalentPointsOnKill,
		Parts:              b.Parts,
		StartedAt:          b.StartedAt,
		DefeatedAt:         b.DefeatedAt,
	})
}

func (b *Boss) UnmarshalJSON(data []byte) error {
	type bossJSON struct {
		ID                 string          `json:"id"`
		TemplateID         string          `json:"templateId,omitempty"`
		RoomID             string          `json:"roomId,omitempty"`
		QueueID            string          `json:"queueId,omitempty"`
		Name               string          `json:"name"`
		Status             string          `json:"status"`
		MaxHP              json.RawMessage `json:"maxHp"`
		CurrentHP          json.RawMessage `json:"currentHp"`
		GoldOnKill         int64           `json:"goldOnKill"`
		StoneOnKill        int64           `json:"stoneOnKill"`
		TalentPointsOnKill int64           `json:"talentPointsOnKill"`
		Parts              []BossPart      `json:"parts,omitempty"`
		StartedAt          int64           `json:"startedAt,omitempty"`
		DefeatedAt         int64           `json:"defeatedAt,omitempty"`
	}
	var aux bossJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	maxHP, err := parseFlexibleInt64Value(aux.MaxHP)
	if err != nil {
		return err
	}
	currentHP, err := parseFlexibleInt64Value(aux.CurrentHP)
	if err != nil {
		return err
	}
	*b = Boss{
		ID:                 aux.ID,
		TemplateID:         aux.TemplateID,
		RoomID:             aux.RoomID,
		QueueID:            aux.QueueID,
		Name:               aux.Name,
		Status:             aux.Status,
		MaxHP:              maxHP,
		CurrentHP:          currentHP,
		GoldOnKill:         aux.GoldOnKill,
		StoneOnKill:        aux.StoneOnKill,
		TalentPointsOnKill: aux.TalentPointsOnKill,
		Parts:              aux.Parts,
		StartedAt:          aux.StartedAt,
		DefeatedAt:         aux.DefeatedAt,
	}
	return nil
}

func (t BossTemplate) MarshalJSON() ([]byte, error) {
	type bossTemplateJSON struct {
		ID                 string          `json:"id"`
		Name               string          `json:"name"`
		MaxHP              string          `json:"maxHp"`
		GoldOnKill         int64           `json:"goldOnKill"`
		StoneOnKill        int64           `json:"stoneOnKill"`
		TalentPointsOnKill int64           `json:"talentPointsOnKill"`
		Loot               []BossLootEntry `json:"loot"`
		Layout             []BossPart      `json:"layout,omitempty"`
	}
	return json.Marshal(bossTemplateJSON{
		ID:                 t.ID,
		Name:               t.Name,
		MaxHP:              formatInt64String(t.MaxHP),
		GoldOnKill:         t.GoldOnKill,
		StoneOnKill:        t.StoneOnKill,
		TalentPointsOnKill: t.TalentPointsOnKill,
		Loot:               t.Loot,
		Layout:             t.Layout,
	})
}

func (e BossHistoryEntry) MarshalJSON() ([]byte, error) {
	type bossHistoryJSON struct {
		ID                 string                 `json:"id"`
		TemplateID         string                 `json:"templateId,omitempty"`
		RoomID             string                 `json:"roomId,omitempty"`
		QueueID            string                 `json:"queueId,omitempty"`
		Name               string                 `json:"name"`
		Status             string                 `json:"status"`
		MaxHP              string                 `json:"maxHp"`
		CurrentHP          string                 `json:"currentHp"`
		GoldOnKill         int64                  `json:"goldOnKill"`
		StoneOnKill        int64                  `json:"stoneOnKill"`
		TalentPointsOnKill int64                  `json:"talentPointsOnKill"`
		Parts              []BossPart             `json:"parts,omitempty"`
		StartedAt          int64                  `json:"startedAt,omitempty"`
		DefeatedAt         int64                  `json:"defeatedAt,omitempty"`
		Loot               []BossLootEntry        `json:"loot"`
		Damage             []BossLeaderboardEntry `json:"damage"`
	}
	return json.Marshal(bossHistoryJSON{
		ID:                 e.ID,
		TemplateID:         e.TemplateID,
		RoomID:             e.RoomID,
		QueueID:            e.QueueID,
		Name:               e.Name,
		Status:             e.Status,
		MaxHP:              formatInt64String(e.MaxHP),
		CurrentHP:          formatInt64String(e.CurrentHP),
		GoldOnKill:         e.GoldOnKill,
		StoneOnKill:        e.StoneOnKill,
		TalentPointsOnKill: e.TalentPointsOnKill,
		Parts:              e.Parts,
		StartedAt:          e.StartedAt,
		DefeatedAt:         e.DefeatedAt,
		Loot:               e.Loot,
		Damage:             e.Damage,
	})
}

func (e *BossHistoryEntry) UnmarshalJSON(data []byte) error {
	type bossHistoryJSON struct {
		ID                 string                 `json:"id"`
		TemplateID         string                 `json:"templateId,omitempty"`
		RoomID             string                 `json:"roomId,omitempty"`
		QueueID            string                 `json:"queueId,omitempty"`
		Name               string                 `json:"name"`
		Status             string                 `json:"status"`
		MaxHP              json.RawMessage        `json:"maxHp"`
		CurrentHP          json.RawMessage        `json:"currentHp"`
		GoldOnKill         int64                  `json:"goldOnKill"`
		StoneOnKill        int64                  `json:"stoneOnKill"`
		TalentPointsOnKill int64                  `json:"talentPointsOnKill"`
		Parts              []BossPart             `json:"parts,omitempty"`
		StartedAt          int64                  `json:"startedAt,omitempty"`
		DefeatedAt         int64                  `json:"defeatedAt,omitempty"`
		Loot               []BossLootEntry        `json:"loot"`
		Damage             []BossLeaderboardEntry `json:"damage"`
	}
	var aux bossHistoryJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	maxHP, err := parseFlexibleInt64Value(aux.MaxHP)
	if err != nil {
		return err
	}
	currentHP, err := parseFlexibleInt64Value(aux.CurrentHP)
	if err != nil {
		return err
	}
	*e = BossHistoryEntry{
		Boss: Boss{
			ID:                 aux.ID,
			TemplateID:         aux.TemplateID,
			RoomID:             aux.RoomID,
			QueueID:            aux.QueueID,
			Name:               aux.Name,
			Status:             aux.Status,
			MaxHP:              maxHP,
			CurrentHP:          currentHP,
			GoldOnKill:         aux.GoldOnKill,
			StoneOnKill:        aux.StoneOnKill,
			TalentPointsOnKill: aux.TalentPointsOnKill,
			Parts:              aux.Parts,
			StartedAt:          aux.StartedAt,
			DefeatedAt:         aux.DefeatedAt,
		},
		Loot:   aux.Loot,
		Damage: aux.Damage,
	}
	return nil
}

func (r RoomInfo) MarshalJSON() ([]byte, error) {
	type roomInfoJSON struct {
		ID                 string `json:"id"`
		Current            bool   `json:"current"`
		Joinable           bool   `json:"joinable"`
		OnlineCount        int    `json:"onlineCount"`
		CycleEnabled       bool   `json:"cycleEnabled"`
		QueueID            string `json:"queueId"`
		CurrentBossID      string `json:"currentBossId,omitempty"`
		CurrentBossName    string `json:"currentBossName,omitempty"`
		CurrentBossStatus  string `json:"currentBossStatus,omitempty"`
		CurrentBossHP      string `json:"currentBossHp,omitempty"`
		CurrentBossMaxHP   string `json:"currentBossMaxHp,omitempty"`
		CurrentBossAvgHP   string `json:"currentBossAvgHp,omitempty"`
		CooldownRemainingS int64  `json:"cooldownRemainingSeconds,omitempty"`
	}
	return json.Marshal(roomInfoJSON{
		ID:                 r.ID,
		Current:            r.Current,
		Joinable:           r.Joinable,
		OnlineCount:        r.OnlineCount,
		CycleEnabled:       r.CycleEnabled,
		QueueID:            r.QueueID,
		CurrentBossID:      r.CurrentBossID,
		CurrentBossName:    r.CurrentBossName,
		CurrentBossStatus:  r.CurrentBossStatus,
		CurrentBossHP:      formatInt64String(r.CurrentBossHP),
		CurrentBossMaxHP:   formatInt64String(r.CurrentBossMaxHP),
		CurrentBossAvgHP:   formatInt64String(r.CurrentBossAvgHP),
		CooldownRemainingS: r.CooldownRemainingS,
	})
}
