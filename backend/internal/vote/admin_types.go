package vote

// AdminState 管理后台聚合状态
type AdminState struct {
	Boss              *Boss                  `json:"boss,omitempty"`
	BossLeaderboard   []BossLeaderboardEntry `json:"bossLeaderboard"`
	Equipment         []EquipmentDefinition  `json:"equipment,omitempty"`
	Loot              []BossLootEntry        `json:"loot"`
	BossCycleEnabled  bool                   `json:"bossCycleEnabled"`
	BossPool          []BossTemplate         `json:"bossPool"`
	BossCycleQueue    []string               `json:"bossCycleQueue"`
	PlayerCount       int64                  `json:"playerCount"`
	RecentPlayerCount int64                  `json:"recentPlayerCount"`
	Players           []AdminPlayerOverview  `json:"players,omitempty"`
}

// AdminPlayerOverview 管理后台的玩家概览
type AdminPlayerOverview struct {
	Nickname   string          `json:"nickname"`
	ClickCount int64           `json:"clickCount"`
	Inventory  []InventoryItem `json:"inventory"`
	Loadout    Loadout         `json:"loadout"`
}

// AdminPlayerPage 后台玩家分页结果
type AdminPlayerPage struct {
	Items      []AdminPlayerOverview `json:"items"`
	NextCursor string                `json:"nextCursor,omitempty"`
	Total      int64                 `json:"total"`
}

// BossUpsert 管理后台 Boss 启动载荷
type BossUpsert struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	MaxHP       int64      `json:"maxHp"`
	GoldOnKill  int64      `json:"goldOnKill"`
	StoneOnKill int64      `json:"stoneOnKill"`
	Parts       []BossPart `json:"parts,omitempty"`
}

// BossTemplate Boss 池模板。
type BossTemplate struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	MaxHP       int64           `json:"maxHp"`
	GoldOnKill  int64           `json:"goldOnKill"`
	StoneOnKill int64           `json:"stoneOnKill"`
	Loot        []BossLootEntry `json:"loot"`
	Layout      []BossPart      `json:"layout,omitempty"` // 部位布局
}

// BossTemplateUpsert 后台 Boss 模板保存载荷。
type BossTemplateUpsert struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	MaxHP       int64      `json:"maxHp"`
	GoldOnKill  int64      `json:"goldOnKill"`
	StoneOnKill int64      `json:"stoneOnKill"`
	Layout      []BossPart `json:"layout,omitempty"`
}

// BossHistoryEntry 历史 Boss 概览
type BossHistoryEntry struct {
	Boss
	Loot   []BossLootEntry        `json:"loot"`
	Damage []BossLeaderboardEntry `json:"damage"`
}
