package vote

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var taskTimeLocation = time.FixedZone("CST", 8*60*60)

// ListTasksForPlayer 返回当前玩家可见任务及进度。
func (s *Store) ListTasksForPlayer(ctx context.Context, nickname string) ([]PlayerTask, error) {
	if s.taskDefinitionStore == nil {
		return []PlayerTask{}, nil
	}

	nowTime := s.now()
	taskDefs, err := s.taskDefinitionStore.ListActiveTaskDefinitions(ctx, nowTime.Unix())
	if err != nil {
		return nil, err
	}
	if len(taskDefs) == 0 {
		return []PlayerTask{}, nil
	}

	normalizedNickname, hasNickname := normalizeNickname(nickname)
	if !hasNickname {
		return []PlayerTask{}, nil
	}
	normalizedNickname, err = s.validatedNickname(normalizedNickname)
	if err != nil {
		return nil, err
	}

	items := make([]PlayerTask, 0, len(taskDefs))
	for _, taskDef := range taskDefs {
		taskDef = NormalizeTaskDefinitionModel(taskDef)
		cycleKey := currentTaskCycleKey(taskDef, nowTime)
		progress, status, completedAt, claimedAt, err := s.taskProgressForPlayer(ctx, normalizedNickname, taskDef, cycleKey)
		if err != nil {
			return nil, err
		}
		items = append(items, PlayerTask{
			TaskID:        taskDef.TaskID,
			Title:         taskDef.Title,
			Description:   taskDef.Description,
			TaskType:      taskDef.TaskType,
			EventKind:     taskDef.EventKind,
			WindowKind:    taskDef.WindowKind,
			ConditionKind: taskDef.ConditionKind,
			TargetValue:   taskDef.TargetValue,
			Rewards:       taskDef.Rewards,
			DisplayOrder:  taskDef.DisplayOrder,
			StartAt:       taskDef.StartAt,
			EndAt:         taskDef.EndAt,
			CycleKey:      cycleKey,
			Progress:      progress,
			Status:        status,
			CompletedAt:   completedAt,
			ClaimedAt:     claimedAt,
			CanClaim:      status == TaskPlayerStatusCompletedUnclaimed,
		})
	}
	return items, nil
}

// ClaimTaskReward 领取一条任务的奖励，并返回最新个人状态。
func (s *Store) ClaimTaskReward(ctx context.Context, nickname string, taskID string) (UserState, error) {
	if s.taskDefinitionStore == nil {
		return UserState{}, ErrTaskNotFound
	}
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return UserState{}, err
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return UserState{}, ErrTaskNotFound
	}
	taskDef, err := s.taskDefinitionStore.GetTaskDefinition(ctx, taskID)
	if err != nil {
		return UserState{}, err
	}
	if taskDef == nil {
		return UserState{}, ErrTaskNotFound
	}
	nowTime := s.now()
	*taskDef = NormalizeTaskDefinitionModel(*taskDef)
	if !taskIsActiveAt(*taskDef, nowTime.Unix()) {
		return UserState{}, ErrTaskNotClaimable
	}
	cycleKey := currentTaskCycleKey(*taskDef, nowTime)
	progress, status, completedAt, _, err := s.taskProgressForPlayer(ctx, normalizedNickname, *taskDef, cycleKey)
	if err != nil {
		return UserState{}, err
	}
	if status == TaskPlayerStatusClaimed {
		return UserState{}, ErrTaskAlreadyClaimed
	}
	if progress < taskDef.TargetValue || status != TaskPlayerStatusCompletedUnclaimed {
		return UserState{}, ErrTaskNotClaimable
	}

	nowUnix := nowTime.Unix()
	rewardItems := make([]Reward, 0, len(taskDef.Rewards.EquipmentItems))
	pipe := s.client.TxPipeline()
	if taskDef.Rewards.Gold > 0 {
		pipe.HIncrBy(ctx, s.resourceKey(normalizedNickname), "gold", taskDef.Rewards.Gold)
	}
	if taskDef.Rewards.Stones > 0 {
		pipe.HIncrBy(ctx, s.resourceKey(normalizedNickname), "stones", taskDef.Rewards.Stones)
	}
	if taskDef.Rewards.TalentPoints > 0 {
		pipe.HIncrBy(ctx, s.resourceKey(normalizedNickname), "talent_points", taskDef.Rewards.TalentPoints)
	}
	for _, reward := range taskDef.Rewards.EquipmentItems {
		if strings.TrimSpace(reward.ItemID) == "" || reward.Quantity <= 0 {
			continue
		}
		definition, defErr := s.getEquipmentDefinition(ctx, reward.ItemID)
		if defErr != nil {
			return UserState{}, defErr
		}
		for count := int64(0); count < reward.Quantity; count++ {
			instanceID, createErr := s.newEquipmentInstanceID(ctx)
			if createErr != nil {
				return UserState{}, createErr
			}
			pipe.HSet(ctx, s.equipmentInstanceKey(instanceID), map[string]any{
				"item_id":       reward.ItemID,
				"enhance_level": "0",
				"spent_stones":  "0",
				"bound":         "0",
				"locked":        "0",
				"created_at":    strconv.FormatInt(nowUnix, 10),
			})
			pipe.SAdd(ctx, s.playerInstancesKey(normalizedNickname), instanceID)
			rewardItems = append(rewardItems, Reward{
				ItemID:    reward.ItemID,
				ItemName:  definition.Name,
				GrantedAt: nowUnix,
			})
		}
	}
	progressKey := s.taskProgressKey(normalizedNickname, taskDef.TaskID, cycleKey)
	pipe.HSet(ctx, progressKey, map[string]any{
		"status":       string(TaskPlayerStatusClaimed),
		"progress":     strconv.FormatInt(progress, 10),
		"target_value": strconv.FormatInt(taskDef.TargetValue, 10),
		"completed_at": strconv.FormatInt(completedAt, 10),
		"claimed_at":   strconv.FormatInt(nowUnix, 10),
	})
	pipe.ZAdd(ctx, s.playerIndexKey, redis.Z{
		Score:  float64(nowUnix),
		Member: normalizedNickname,
	})
	if len(rewardItems) > 0 {
		pipe.HSet(ctx, s.lastRewardKey(normalizedNickname), rewardRecordValues(rewardItems))
	}
	if ttl := taskCycleTTL(*taskDef, nowUnix); ttl > 0 {
		pipe.Expire(ctx, progressKey, ttl)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return UserState{}, err
	}

	if s.taskClaimLogStore != nil {
		if err := s.taskClaimLogStore.WriteTaskClaimLog(ctx, TaskClaimLog{
			TaskID:    taskDef.TaskID,
			CycleKey:  cycleKey,
			Nickname:  normalizedNickname,
			Rewards:   taskDef.Rewards,
			ClaimedAt: nowUnix,
		}); err != nil {
			return UserState{}, err
		}
	}

	return s.GetUserState(ctx, normalizedNickname)
}

func (s *Store) recordTaskEvent(ctx context.Context, nickname string, eventKind TaskEventKind, delta int64) error {
	if s.taskDefinitionStore == nil || delta <= 0 {
		return nil
	}
	normalizedNickname, hasNickname := normalizeNickname(nickname)
	if !hasNickname {
		return nil
	}
	normalizedNickname, err := s.validatedNickname(normalizedNickname)
	if err != nil {
		return err
	}

	nowTime := s.now()
	taskDefs, err := s.taskDefinitionStore.ListActiveTaskDefinitions(ctx, nowTime.Unix())
	if err != nil {
		return err
	}
	for _, taskDef := range taskDefs {
		taskDef = NormalizeTaskDefinitionModel(taskDef)
		if !taskMatchesEvent(taskDef, eventKind) {
			continue
		}
		cycleKey := currentTaskCycleKey(taskDef, nowTime)
		if err := s.incrementTaskProgress(ctx, normalizedNickname, taskDef, cycleKey, delta, nowTime.Unix()); err != nil {
			return err
		}
	}
	return nil
}

func taskMatchesEvent(task TaskDefinition, eventKind TaskEventKind) bool {
	task = NormalizeTaskDefinitionModel(task)
	return task.EventKind == eventKind
}

func taskIsActiveAt(task TaskDefinition, nowUnix int64) bool {
	if task.Status != TaskStatusActive {
		return false
	}
	if task.StartAt > 0 && nowUnix < task.StartAt {
		return false
	}
	if task.EndAt > 0 && nowUnix > task.EndAt {
		return false
	}
	return true
}

func currentTaskCycleKey(task TaskDefinition, nowTime time.Time) string {
	task = NormalizeTaskDefinitionModel(task)
	localNow := nowTime.In(taskTimeLocation)
	switch task.WindowKind {
	case TaskWindowDaily:
		return localNow.Format("2006-01-02")
	case TaskWindowWeekly:
		year, week := localNow.ISOWeek()
		return fmt.Sprintf("%04d-W%02d", year, week)
	case TaskWindowFixedRange:
		return fmt.Sprintf("%s:%d:%d", task.TaskID, task.StartAt, task.EndAt)
	default:
		return localNow.Format("2006-01-02")
	}
}

func (s *Store) incrementTaskProgress(ctx context.Context, nickname string, task TaskDefinition, cycleKey string, delta int64, nowUnix int64) error {
	progressKey := s.taskProgressKey(nickname, task.TaskID, cycleKey)
	participantKey := s.taskParticipantsKey(task.TaskID, cycleKey)

	pipe := s.client.TxPipeline()
	progressCmd := pipe.HIncrBy(ctx, progressKey, "progress", delta)
	pipe.HSet(ctx, progressKey, "target_value", strconv.FormatInt(task.TargetValue, 10))
	pipe.SAdd(ctx, participantKey, nickname)
	if ttl := taskCycleTTL(task, nowUnix); ttl > 0 {
		pipe.Expire(ctx, progressKey, ttl)
		pipe.Expire(ctx, participantKey, ttl)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	progress := progressCmd.Val()
	_, status, completedAt, claimedAt, err := s.taskProgressForPlayer(ctx, nickname, task, cycleKey)
	if err != nil {
		return err
	}
	if status == TaskPlayerStatusClaimed {
		return nil
	}

	nextStatus := TaskPlayerStatusInProgress
	if progress >= task.TargetValue && task.TargetValue > 0 {
		nextStatus = TaskPlayerStatusCompletedUnclaimed
		if completedAt == 0 {
			completedAt = nowUnix
		}
	}
	values := map[string]any{
		"status":       string(nextStatus),
		"progress":     strconv.FormatInt(progress, 10),
		"target_value": strconv.FormatInt(task.TargetValue, 10),
		"completed_at": strconv.FormatInt(completedAt, 10),
	}
	if claimedAt > 0 {
		values["claimed_at"] = strconv.FormatInt(claimedAt, 10)
	}
	return s.client.HSet(ctx, progressKey, values).Err()
}

func (s *Store) taskProgressForPlayer(ctx context.Context, nickname string, task TaskDefinition, cycleKey string) (int64, TaskPlayerStatus, int64, int64, error) {
	progressKey := s.taskProgressKey(nickname, task.TaskID, cycleKey)
	values, err := s.client.HMGet(ctx, progressKey, "progress", "status", "completed_at", "claimed_at").Result()
	if err != nil {
		return 0, TaskPlayerStatusInProgress, 0, 0, err
	}
	progress := int64Value(values, 0)
	status := TaskPlayerStatus(strings.TrimSpace(stringValue(values, 1)))
	if status == "" {
		status = TaskPlayerStatusInProgress
		if progress >= task.TargetValue && task.TargetValue > 0 {
			status = TaskPlayerStatusCompletedUnclaimed
		}
	}
	completedAt := int64Value(values, 2)
	claimedAt := int64Value(values, 3)
	if claimedAt == 0 && s.taskClaimLogStore != nil {
		claimed, claimErr := s.taskClaimLogStore.HasTaskClaimed(ctx, task.TaskID, cycleKey, nickname)
		if claimErr != nil {
			return 0, TaskPlayerStatusInProgress, 0, 0, claimErr
		}
		if claimed {
			status = TaskPlayerStatusClaimed
		}
	}
	return progress, status, completedAt, claimedAt, nil
}

func (s *Store) taskProgressKey(nickname string, taskID string, cycleKey string) string {
	return s.taskProgressPrefix + strings.TrimSpace(nickname) + ":" + strings.TrimSpace(taskID) + ":" + strings.TrimSpace(cycleKey)
}

func (s *Store) taskParticipantsKey(taskID string, cycleKey string) string {
	return s.taskParticipantsPrefix + strings.TrimSpace(taskID) + ":" + strings.TrimSpace(cycleKey)
}

func taskCycleTTL(task TaskDefinition, nowUnix int64) time.Duration {
	task = NormalizeTaskDefinitionModel(task)
	nowTime := time.Unix(nowUnix, 0).In(taskTimeLocation)
	switch task.WindowKind {
	case TaskWindowDaily:
		expireAt := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day()+3, 0, 0, 0, 0, taskTimeLocation)
		return time.Until(expireAt)
	case TaskWindowWeekly:
		expireAt := nowTime.AddDate(0, 0, 10)
		return time.Until(expireAt)
	case TaskWindowFixedRange:
		if task.EndAt > 0 {
			return time.Until(time.Unix(task.EndAt, 0).Add(72 * time.Hour))
		}
		return 72 * time.Hour
	default:
		return 72 * time.Hour
	}
}
