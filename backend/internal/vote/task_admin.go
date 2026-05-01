package vote

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ListTaskDefinitions 返回后台任务列表。
func (s *Store) ListTaskDefinitions(ctx context.Context) ([]TaskDefinition, error) {
	if s.taskDefinitionStore == nil {
		return []TaskDefinition{}, nil
	}
	items, err := s.taskDefinitionStore.ListTaskDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	for index := range items {
		items[index] = NormalizeTaskDefinitionModel(items[index])
	}
	return items, nil
}

// SaveTaskDefinition 创建或更新任务定义。
func (s *Store) SaveTaskDefinition(ctx context.Context, item TaskDefinition) error {
	if s.taskDefinitionStore == nil {
		return ErrTaskNotFound
	}
	item.TaskID = strings.TrimSpace(item.TaskID)
	item.Title = strings.TrimSpace(item.Title)
	item.Description = strings.TrimSpace(item.Description)
	if item.TaskID == "" || item.Title == "" || item.TargetValue <= 0 {
		return ErrTaskNotClaimable
	}
	if item.Status == "" {
		item.Status = TaskStatusDraft
	}
	item = NormalizeTaskDefinitionModel(item)
	nowUnix := s.now().Unix()
	existing, err := s.taskDefinitionStore.GetTaskDefinition(ctx, item.TaskID)
	if err != nil {
		return err
	}
	if existing != nil {
		*existing = NormalizeTaskDefinitionModel(*existing)
		if existing.Status == TaskStatusActive && taskCoreChanged(*existing, item) {
			return ErrTaskImmutable
		}
		item.CreatedAt = existing.CreatedAt
		if item.Status == "" {
			item.Status = existing.Status
		}
	} else if item.CreatedAt == 0 {
		item.CreatedAt = nowUnix
	}
	item.Rewards = normalizeTaskRewards(item.Rewards)
	if !taskDefinitionHasRewards(item.Rewards) || !taskDefinitionTimeWindowValid(item) {
		return ErrTaskNotClaimable
	}
	item.UpdatedAt = nowUnix
	return s.taskDefinitionStore.UpsertTaskDefinition(ctx, item)
}

func (s *Store) ActivateTaskDefinition(ctx context.Context, taskID string) error {
	if s.taskDefinitionStore == nil {
		return ErrTaskNotFound
	}
	task, err := s.taskDefinitionStore.GetTaskDefinition(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}
	*task = NormalizeTaskDefinitionModel(*task)
	task.Rewards = normalizeTaskRewards(task.Rewards)
	if !taskDefinitionHasRewards(task.Rewards) || !taskDefinitionTimeWindowValid(*task) {
		return ErrTaskNotClaimable
	}
	task.Status = TaskStatusActive
	task.UpdatedAt = s.now().Unix()
	return s.taskDefinitionStore.UpsertTaskDefinition(ctx, *task)
}

func (s *Store) DeactivateTaskDefinition(ctx context.Context, taskID string) error {
	if s.taskDefinitionStore == nil {
		return ErrTaskNotFound
	}
	task, err := s.taskDefinitionStore.GetTaskDefinition(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}
	*task = NormalizeTaskDefinitionModel(*task)
	task.Status = TaskStatusInactive
	task.UpdatedAt = s.now().Unix()
	return s.taskDefinitionStore.UpsertTaskDefinition(ctx, *task)
}

func (s *Store) DuplicateTaskDefinition(ctx context.Context, taskID string, newTaskID string) (*TaskDefinition, error) {
	if s.taskDefinitionStore == nil {
		return nil, ErrTaskNotFound
	}
	task, err := s.taskDefinitionStore.GetTaskDefinition(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}
	*task = NormalizeTaskDefinitionModel(*task)
	newTaskID = strings.TrimSpace(newTaskID)
	if newTaskID != "" {
		existing, err := s.taskDefinitionStore.GetTaskDefinition(ctx, newTaskID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, ErrTaskImmutable
		}
	}
	nowUnix := s.now().Unix()
	copyTask := *task
	if newTaskID == "" {
		copyTask.TaskID = fmt.Sprintf("%s-copy-%d", task.TaskID, nowUnix)
	} else {
		copyTask.TaskID = newTaskID
	}
	copyTask.Status = TaskStatusDraft
	copyTask.CreatedAt = nowUnix
	copyTask.UpdatedAt = nowUnix
	copyTask.Rewards = normalizeTaskRewards(copyTask.Rewards)
	if err := s.taskDefinitionStore.UpsertTaskDefinition(ctx, copyTask); err != nil {
		return nil, err
	}
	return &copyTask, nil
}

func (s *Store) ArchiveExpiredTaskCycles(ctx context.Context, nowTime time.Time) ([]TaskCycleArchive, error) {
	if s.taskDefinitionStore == nil || s.taskCycleArchiveStore == nil {
		return []TaskCycleArchive{}, nil
	}
	taskDefs, err := s.taskDefinitionStore.ListTaskDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	archives := make([]TaskCycleArchive, 0)
	for _, task := range taskDefs {
		task = NormalizeTaskDefinitionModel(task)
		cycleKey, shouldArchive := archiveCycleKeyForTask(task, nowTime)
		if !shouldArchive {
			continue
		}
		archive, err := s.archiveTaskCycle(ctx, task, cycleKey, nowTime.Unix())
		if err != nil {
			return nil, err
		}
		if archive.TaskID != "" {
			archives = append(archives, archive)
		}
		if task.WindowKind == TaskWindowFixedRange && task.EndAt > 0 && task.EndAt < nowTime.Unix() && task.Status == TaskStatusActive {
			task.Status = TaskStatusExpired
			task.UpdatedAt = nowTime.Unix()
			if err := s.taskDefinitionStore.UpsertTaskDefinition(ctx, task); err != nil {
				return nil, err
			}
		}
	}
	return archives, nil
}

func (s *Store) ListTaskCycleArchives(ctx context.Context, taskID string) ([]TaskCycleArchive, error) {
	if s.taskCycleArchiveStore == nil {
		return []TaskCycleArchive{}, nil
	}
	items, err := s.taskCycleArchiveStore.ListTaskCycleArchives(ctx, taskID)
	if err != nil {
		return nil, err
	}
	for index := range items {
		items[index] = NormalizeTaskArchiveModel(items[index])
	}
	return items, nil
}

func (s *Store) GetTaskCycleResults(ctx context.Context, taskID string, cycleKey string) (TaskCycleResultsView, error) {
	if s.taskCycleArchiveStore == nil {
		return TaskCycleResultsView{}, nil
	}
	result, err := s.taskCycleArchiveStore.GetTaskCycleResults(ctx, taskID, cycleKey)
	if err != nil {
		return TaskCycleResultsView{}, err
	}
	result.Archive = NormalizeTaskArchiveModel(result.Archive)
	return result, nil
}

func normalizeTaskRewards(rewards TaskRewards) TaskRewards {
	rewards.EquipmentItems = normalizeTaskEquipmentRewards(rewards.EquipmentItems)
	return rewards
}

func normalizeTaskEquipmentRewards(items []TaskEquipmentReward) []TaskEquipmentReward {
	if len(items) == 0 {
		return nil
	}
	result := make([]TaskEquipmentReward, 0, len(items))
	for _, item := range items {
		itemID := strings.TrimSpace(item.ItemID)
		if itemID == "" || item.Quantity <= 0 {
			continue
		}
		result = append(result, TaskEquipmentReward{
			ItemID:   itemID,
			Quantity: item.Quantity,
		})
	}
	return result
}

func taskDefinitionHasRewards(rewards TaskRewards) bool {
	return rewards.Gold > 0 || rewards.Stones > 0 || rewards.TalentPoints > 0 || len(rewards.EquipmentItems) > 0
}

func taskDefinitionTimeWindowValid(item TaskDefinition) bool {
	item = NormalizeTaskDefinitionModel(item)
	if item.WindowKind != TaskWindowFixedRange {
		return true
	}
	return item.StartAt > 0 && item.EndAt > 0 && item.EndAt > item.StartAt
}

func (s *Store) archiveTaskCycle(ctx context.Context, task TaskDefinition, cycleKey string, nowUnix int64) (TaskCycleArchive, error) {
	task = NormalizeTaskDefinitionModel(task)
	participantKey := s.taskParticipantsKey(task.TaskID, cycleKey)
	participants, err := s.client.SMembers(ctx, participantKey).Result()
	if err != nil {
		return TaskCycleArchive{}, err
	}
	allPlayers, err := s.client.ZRevRange(ctx, s.playerIndexKey, 0, -1).Result()
	if err != nil {
		return TaskCycleArchive{}, err
	}
	var claimLogs []TaskClaimLog
	if s.taskClaimLogStore != nil {
		claimLogs, err = s.taskClaimLogStore.ListTaskClaimLogs(ctx, task.TaskID, cycleKey)
		if err != nil {
			return TaskCycleArchive{}, err
		}
	}

	participantSet := make(map[string]struct{}, len(participants))
	for _, nickname := range participants {
		trimmed := strings.TrimSpace(nickname)
		if trimmed == "" {
			continue
		}
		participantSet[trimmed] = struct{}{}
	}
	claimLogMap := make(map[string]TaskClaimLog, len(claimLogs))
	for _, item := range claimLogs {
		claimLogMap[item.Nickname] = item
	}

	results := make([]TaskCyclePlayerResult, 0, len(allPlayers))
	archive := TaskCycleArchive{
		TaskID:        task.TaskID,
		CycleKey:      cycleKey,
		TaskType:      task.TaskType,
		EventKind:     task.EventKind,
		WindowKind:    task.WindowKind,
		ConditionKind: task.ConditionKind,
		TargetValue:   task.TargetValue,
		StartAt:       task.StartAt,
		EndAt:         task.EndAt,
		ArchivedAt:    nowUnix,
	}
	for _, nickname := range allPlayers {
		trimmed := strings.TrimSpace(nickname)
		if trimmed == "" {
			continue
		}
		result := TaskCyclePlayerResult{
			TaskID:      task.TaskID,
			CycleKey:    cycleKey,
			Nickname:    trimmed,
			TargetValue: task.TargetValue,
			ArchivedAt:  nowUnix,
			Status:      TaskPlayerStatusNotParticipated,
		}
		if _, ok := participantSet[trimmed]; ok {
			progress, status, completedAt, claimedAt, err := s.taskProgressForPlayer(ctx, trimmed, task, cycleKey)
			if err != nil {
				return TaskCycleArchive{}, err
			}
			result.Progress = progress
			result.CompletedAt = completedAt
			result.ClaimedAt = claimedAt
			switch status {
			case TaskPlayerStatusClaimed:
				result.Status = TaskPlayerStatusClaimed
				archive.ClaimedTotal++
				archive.CompletedTotal++
			case TaskPlayerStatusCompletedUnclaimed:
				result.Status = TaskPlayerStatusCompletedUnclaimed
				archive.CompletedTotal++
				archive.ExpiredUnclaimedTotal++
			default:
				result.Status = TaskPlayerStatusUnfinished
				archive.UnfinishedTotal++
			}
			archive.ParticipantsTotal++
		} else if claimLog, ok := claimLogMap[trimmed]; ok {
			result.Status = TaskPlayerStatusClaimed
			result.Progress = task.TargetValue
			result.ClaimedAt = claimLog.ClaimedAt
			archive.ParticipantsTotal++
			archive.CompletedTotal++
			archive.ClaimedTotal++
		} else {
			archive.NotParticipatedTotal++
		}
		results = append(results, result)
	}
	if err := s.taskCycleArchiveStore.UpsertTaskCycleArchive(ctx, archive); err != nil {
		return TaskCycleArchive{}, err
	}
	if err := s.taskCycleArchiveStore.UpsertTaskCyclePlayerResults(ctx, results); err != nil {
		return TaskCycleArchive{}, err
	}
	return archive, nil
}

func taskCoreChanged(existing TaskDefinition, next TaskDefinition) bool {
	existing = NormalizeTaskDefinitionModel(existing)
	next = NormalizeTaskDefinitionModel(next)
	return existing.EventKind != next.EventKind ||
		existing.WindowKind != next.WindowKind ||
		existing.TargetValue != next.TargetValue ||
		existing.StartAt != next.StartAt ||
		existing.EndAt != next.EndAt
}

func archiveCycleKeyForTask(task TaskDefinition, nowTime time.Time) (string, bool) {
	task = NormalizeTaskDefinitionModel(task)
	localNow := nowTime.In(taskTimeLocation)
	switch task.WindowKind {
	case TaskWindowDaily:
		prev := localNow.AddDate(0, 0, -1)
		return prev.Format("2006-01-02"), true
	case TaskWindowWeekly:
		if localNow.Weekday() != time.Monday {
			return "", false
		}
		prev := localNow.AddDate(0, 0, -7)
		year, week := prev.ISOWeek()
		return fmt.Sprintf("%04d-W%02d", year, week), true
	case TaskWindowFixedRange:
		if task.EndAt <= 0 || nowTime.Unix() <= task.EndAt {
			return "", false
		}
		return fmt.Sprintf("%s:%d:%d", task.TaskID, task.StartAt, task.EndAt), true
	default:
		return "", false
	}
}
