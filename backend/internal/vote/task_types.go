package vote

// TaskType 任务周期类型。
type TaskType string

const (
	TaskTypeDaily   TaskType = "daily"
	TaskTypeWeekly  TaskType = "weekly"
	TaskTypeLimited TaskType = "limited"
)

// TaskStatus 任务状态。
type TaskStatus string

const (
	TaskStatusDraft    TaskStatus = "draft"
	TaskStatusActive   TaskStatus = "active"
	TaskStatusInactive TaskStatus = "inactive"
	TaskStatusExpired  TaskStatus = "expired"
)

// TaskConditionKind 任务条件类型。
type TaskConditionKind string

const (
	TaskConditionDailyClicks  TaskConditionKind = "daily_clicks"
	TaskConditionWeeklyClicks TaskConditionKind = "weekly_clicks"
	TaskConditionBossKills    TaskConditionKind = "boss_kills"
	TaskConditionEnhanceCount TaskConditionKind = "enhance_count"
)

// TaskPlayerStatus 玩家在某个任务周期下的结果状态。
type TaskPlayerStatus string

const (
	TaskPlayerStatusInProgress         TaskPlayerStatus = "in_progress"
	TaskPlayerStatusClaimed            TaskPlayerStatus = "claimed"
	TaskPlayerStatusCompletedUnclaimed TaskPlayerStatus = "completed_unclaimed"
	TaskPlayerStatusUnfinished         TaskPlayerStatus = "unfinished"
	TaskPlayerStatusNotParticipated    TaskPlayerStatus = "not_participated"
)

// TaskEquipmentReward 装备奖励项。
type TaskEquipmentReward struct {
	ItemID   string `json:"itemId"`
	Quantity int64  `json:"quantity"`
}

// TaskRewards 任务奖励。
type TaskRewards struct {
	Gold           int64                 `json:"gold"`
	Stones         int64                 `json:"stones"`
	TalentPoints   int64                 `json:"talentPoints"`
	EquipmentItems []TaskEquipmentReward `json:"equipmentItems,omitempty"`
}

// TaskDefinition 任务定义。
type TaskDefinition struct {
	TaskID        string            `json:"taskId"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	TaskType      TaskType          `json:"taskType"`
	Status        TaskStatus        `json:"status"`
	ConditionKind TaskConditionKind `json:"conditionKind"`
	TargetValue   int64             `json:"targetValue"`
	Rewards       TaskRewards       `json:"rewards"`
	DisplayOrder  int64             `json:"displayOrder"`
	StartAt       int64             `json:"startAt,omitempty"`
	EndAt         int64             `json:"endAt,omitempty"`
	CreatedAt     int64             `json:"createdAt"`
	UpdatedAt     int64             `json:"updatedAt"`
}

// TaskClaimLog 玩家领取任务奖励的事实记录。
type TaskClaimLog struct {
	TaskID     string      `json:"taskId"`
	CycleKey   string      `json:"cycleKey"`
	Nickname   string      `json:"nickname"`
	Rewards    TaskRewards `json:"rewards"`
	ClaimedAt  int64       `json:"claimedAt"`
	ArchivedAt int64       `json:"archivedAt,omitempty"`
}

// TaskCycleArchive 周期汇总归档。
type TaskCycleArchive struct {
	TaskID                string            `json:"taskId"`
	CycleKey              string            `json:"cycleKey"`
	TaskType              TaskType          `json:"taskType"`
	ConditionKind         TaskConditionKind `json:"conditionKind"`
	TargetValue           int64             `json:"targetValue"`
	StartAt               int64             `json:"startAt,omitempty"`
	EndAt                 int64             `json:"endAt,omitempty"`
	ParticipantsTotal     int64             `json:"participantsTotal"`
	CompletedTotal        int64             `json:"completedTotal"`
	ClaimedTotal          int64             `json:"claimedTotal"`
	ExpiredUnclaimedTotal int64             `json:"expiredUnclaimedTotal"`
	UnfinishedTotal       int64             `json:"unfinishedTotal"`
	NotParticipatedTotal  int64             `json:"notParticipatedTotal"`
	ArchivedAt            int64             `json:"archivedAt"`
}

// TaskCyclePlayerResult 周期内单个玩家的归档结果。
type TaskCyclePlayerResult struct {
	TaskID      string           `json:"taskId"`
	CycleKey    string           `json:"cycleKey"`
	Nickname    string           `json:"nickname"`
	Progress    int64            `json:"progress"`
	TargetValue int64            `json:"targetValue"`
	Status      TaskPlayerStatus `json:"status"`
	CompletedAt int64            `json:"completedAt,omitempty"`
	ClaimedAt   int64            `json:"claimedAt,omitempty"`
	ArchivedAt  int64            `json:"archivedAt"`
}

// TaskCycleResultsView 后台查看任务某周期的归档结果。
type TaskCycleResultsView struct {
	Archive TaskCycleArchive        `json:"archive"`
	Items   []TaskCyclePlayerResult `json:"items"`
}

// PlayerTask 玩家当前可见任务视图。
type PlayerTask struct {
	TaskID        string            `json:"taskId"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	TaskType      TaskType          `json:"taskType"`
	ConditionKind TaskConditionKind `json:"conditionKind"`
	TargetValue   int64             `json:"targetValue"`
	Rewards       TaskRewards       `json:"rewards"`
	DisplayOrder  int64             `json:"displayOrder"`
	StartAt       int64             `json:"startAt,omitempty"`
	EndAt         int64             `json:"endAt,omitempty"`
	CycleKey      string            `json:"cycleKey"`
	Progress      int64             `json:"progress"`
	Status        TaskPlayerStatus  `json:"status"`
	CompletedAt   int64             `json:"completedAt,omitempty"`
	ClaimedAt     int64             `json:"claimedAt,omitempty"`
	CanClaim      bool              `json:"canClaim"`
}
