package core

// TaskType 任务周期类型。
type TaskType string

const (
	TaskTypeDaily    TaskType = "daily"
	TaskTypeWeekly   TaskType = "weekly"
	TaskTypeLimited  TaskType = "limited"
	TaskTypeLongTerm TaskType = "long_term"
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

// TaskEventKind 任务统计的行为类型。
type TaskEventKind string

const (
	TaskEventClick    TaskEventKind = "click"
	TaskEventBossKill TaskEventKind = "boss_kill"
	TaskEventEnhance  TaskEventKind = "enhance"
)

// TaskWindowKind 任务累计窗口类型。
type TaskWindowKind string

const (
	TaskWindowDaily      TaskWindowKind = "daily"
	TaskWindowWeekly     TaskWindowKind = "weekly"
	TaskWindowFixedRange TaskWindowKind = "fixed_range"
	TaskWindowLifetime   TaskWindowKind = "lifetime"
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

// EquipmentRewardDetail 装备奖励详情（含属性，供前端展示悬浮窗口）。
type EquipmentRewardDetail struct {
	ItemID               string  `json:"itemId"`
	Name                 string  `json:"name"`
	Rarity               string  `json:"rarity"`
	ImagePath            string  `json:"imagePath,omitempty"`
	ImageAlt             string  `json:"imageAlt,omitempty"`
	AttackPower          int64   `json:"attackPower,omitempty"`
	ArmorPenPercent      float64 `json:"armorPenPercent,omitempty"`
	CritRate             float64 `json:"critRate"`
	CritDamageMultiplier float64 `json:"critDamageMultiplier,omitempty"`
	PartTypeDamageSoft   float64 `json:"partTypeDamageSoft,omitempty"`
	PartTypeDamageHeavy  float64 `json:"partTypeDamageHeavy,omitempty"`
	PartTypeDamageWeak   float64 `json:"partTypeDamageWeak,omitempty"`
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
	EventKind     TaskEventKind     `json:"eventKind"`
	WindowKind    TaskWindowKind    `json:"windowKind"`
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
	EventKind             TaskEventKind     `json:"eventKind"`
	WindowKind            TaskWindowKind    `json:"windowKind"`
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
	TaskID                 string                  `json:"taskId"`
	Title                  string                  `json:"title"`
	Description            string                  `json:"description"`
	TaskType               TaskType                `json:"taskType"`
	EventKind              TaskEventKind           `json:"eventKind"`
	WindowKind             TaskWindowKind          `json:"windowKind"`
	ConditionKind          TaskConditionKind       `json:"conditionKind"`
	TargetValue            int64                   `json:"targetValue"`
	Rewards                TaskRewards             `json:"rewards"`
	DisplayOrder           int64                   `json:"displayOrder"`
	StartAt                int64                   `json:"startAt,omitempty"`
	EndAt                  int64                   `json:"endAt,omitempty"`
	CycleKey               string                  `json:"cycleKey"`
	Progress               int64                   `json:"progress"`
	Status                 TaskPlayerStatus        `json:"status"`
	CompletedAt            int64                   `json:"completedAt,omitempty"`
	ClaimedAt              int64                   `json:"claimedAt,omitempty"`
	CanClaim               bool                    `json:"canClaim"`
	EquipmentRewardDetails []EquipmentRewardDetail `json:"equipmentRewardDetails,omitempty"`
}

// NormalizeTaskDefinitionModel 将任务定义补齐为新旧字段并存的兼容模型。
func NormalizeTaskDefinitionModel(item TaskDefinition) TaskDefinition {
	if item.EventKind == "" {
		item.EventKind = taskEventKindFromLegacy(item.ConditionKind)
	}
	if item.WindowKind == "" {
		item.WindowKind = taskWindowKindFromLegacy(item.TaskType, item.ConditionKind)
	}
	item.TaskType = legacyTaskTypeFromWindowKind(item.WindowKind)
	item.ConditionKind = legacyConditionKindFromModel(item.EventKind, item.WindowKind)
	return item
}

// NormalizeTaskArchiveModel 将任务归档补齐为新旧字段并存的兼容模型。
func NormalizeTaskArchiveModel(item TaskCycleArchive) TaskCycleArchive {
	if item.EventKind == "" {
		item.EventKind = taskEventKindFromLegacy(item.ConditionKind)
	}
	if item.WindowKind == "" {
		item.WindowKind = taskWindowKindFromLegacy(item.TaskType, item.ConditionKind)
	}
	item.TaskType = legacyTaskTypeFromWindowKind(item.WindowKind)
	item.ConditionKind = legacyConditionKindFromModel(item.EventKind, item.WindowKind)
	return item
}

func taskEventKindFromLegacy(conditionKind TaskConditionKind) TaskEventKind {
	switch conditionKind {
	case TaskConditionDailyClicks, TaskConditionWeeklyClicks:
		return TaskEventClick
	case TaskConditionBossKills:
		return TaskEventBossKill
	case TaskConditionEnhanceCount:
		return TaskEventEnhance
	default:
		return TaskEventClick
	}
}

func taskWindowKindFromLegacy(taskType TaskType, conditionKind TaskConditionKind) TaskWindowKind {
	switch conditionKind {
	case TaskConditionDailyClicks:
		if taskType != TaskTypeLimited {
			return TaskWindowDaily
		}
	case TaskConditionWeeklyClicks:
		if taskType != TaskTypeLimited {
			return TaskWindowWeekly
		}
	}
	switch taskType {
	case TaskTypeWeekly:
		return TaskWindowWeekly
	case TaskTypeLimited:
		return TaskWindowFixedRange
	case TaskTypeLongTerm:
		return TaskWindowLifetime
	default:
		return TaskWindowDaily
	}
}

func legacyTaskTypeFromWindowKind(windowKind TaskWindowKind) TaskType {
	switch windowKind {
	case TaskWindowWeekly:
		return TaskTypeWeekly
	case TaskWindowFixedRange:
		return TaskTypeLimited
	case TaskWindowLifetime:
		return TaskTypeLongTerm
	default:
		return TaskTypeDaily
	}
}

func legacyConditionKindFromModel(eventKind TaskEventKind, windowKind TaskWindowKind) TaskConditionKind {
	switch eventKind {
	case TaskEventBossKill:
		return TaskConditionBossKills
	case TaskEventEnhance:
		return TaskConditionEnhanceCount
	default:
		if windowKind == TaskWindowWeekly {
			return TaskConditionWeeklyClicks
		}
		return TaskConditionDailyClicks
	}
}
