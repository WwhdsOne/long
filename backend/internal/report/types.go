package report

// ReportSummary 一份完整报表
type ReportSummary struct {
	Title    string              // 报表标题
	From     int64               // 起始 Unix 秒
	To       int64               // 结束 Unix 秒
	Player   PlayerActivityStats // 玩家活跃
	Boss     BossStats           // Boss 战况
	Economy  EconomyStats        // 经济系统
}

// PlayerActivityStats 玩家活跃度
type PlayerActivityStats struct {
	UniqueIPs      int64            // 独立访问用户数（access_logs 去重 client_ip）
	ActivePlayers  int64            // 活跃玩家数（domain_events 去重 nickname）
	NewPlayers     int64            // 新增玩家数（区间内首次出现的 nickname）
	TotalRequests  int64            // 总请求量
	P95LatencyMs   float64          // P95 延迟（毫秒）
	TopPlayers     []PlayerActivity // Top 10 活跃玩家
}

// PlayerActivity 单个玩家活跃记录
type PlayerActivity struct {
	Nickname   string
	EventCount int64
}

// BossStats Boss 战况
type BossStats struct {
	SpawnCount      int64            // Boss 生成次数
	KillCount       int64            // Boss 击杀次数
	KillRate        float64          // 击杀率（百分比）
	TotalDamage     int64            // 总伤害量
	AvgSurvivalSecs float64          // 平均存活时间（秒）
	TopDamagers     []DamageRankItem // Top 10 伤害榜
	TopLoot         []LootRankItem   // Top 5 掉落装备
}

// DamageRankItem 伤害榜条目
type DamageRankItem struct {
	Nickname    string
	TotalDamage int64
}

// LootRankItem 掉落榜条目
type LootRankItem struct {
	ItemName string
	Count    int64
}

// EconomyStats 经济系统
type EconomyStats struct {
	ShopTotalGold    int64            // 商店总销售额
	ShopPurchaseCnt  int64            // 商店购买次数
	TopShopItems     []ShopRankItem   // 热销 Top 10
	TaskClaimCnt      int64            // 任务完成数
	TaskRewardGold   int64            // 任务奖励总金币
	TaskRewardStones int64            // 任务奖励总石头
	TaskRewardTP     int64            // 任务奖励总天赋点
	TaskParticipants int64            // 任务参与人数（去重）
}

// ShopRankItem 热销条目
type ShopRankItem struct {
	ItemID string
	Count  int64
}
