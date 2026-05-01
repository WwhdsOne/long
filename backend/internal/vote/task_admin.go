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
	return s.taskDefinitionStore.ListTaskDefinitions(ctx)
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
	if item.TaskType == "" {
		item.TaskType = TaskTypeDaily
	}
	if item.Status == "" {
		item.Status = TaskStatusDraft
	}
	if item.ConditionKind == "" {
		item.ConditionKind = TaskConditionDailyClicks
	}
	nowUnix := s.now().Unix()
	existing, err := s.taskDefinitionStore.GetTaskDefinition(ctx, item.TaskID)
	if err != nil {
		return err
	}
	if existing != nil {
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
	nowUnix := s.now().Unix()
	copyTask := *task
	if strings.TrimSpace(newTaskID) == "" {
		copyTask.TaskID = fmt.Sprintf("%s-copy-%d", task.TaskID, nowUnix)
	} else {
		copyTask.TaskID = strings.TrimSpace(newTaskID)
	}
	copyTask.Status = TaskStatusDraft
	copyTask.CreatedAt = nowUnix
	copyTask.UpdatedAt = nowUnix
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
		if task.TaskType == TaskTypeLimited && task.EndAt > 0 && task.EndAt < nowTime.Unix() && task.Status == TaskStatusActive {
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
	return s.taskCycleArchiveStore.ListTaskCycleArchives(ctx, taskID)
}

func (s *Store) GetTaskCycleResults(ctx context.Context, taskID string, cycleKey string) (TaskCycleResultsView, error) {
	if s.taskCycleArchiveStore == nil {
		return TaskCycleResultsView{}, nil
	}
	return s.taskCycleArchiveStore.GetTaskCycleResults(ctx, taskID, cycleKey)
}

func (s *Store) archiveTaskCycle(ctx context.Context, task TaskDefinition, cycleKey string, nowUnix int64) (TaskCycleArchive, error) {
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
	return existing.TaskType != next.TaskType ||
		existing.ConditionKind != next.ConditionKind ||
		existing.TargetValue != next.TargetValue ||
		existing.StartAt != next.StartAt ||
		existing.EndAt != next.EndAt
}

func archiveCycleKeyForTask(task TaskDefinition, nowTime time.Time) (string, bool) {
	localNow := nowTime.In(taskTimeLocation)
	switch task.TaskType {
	case TaskTypeDaily:
		prev := localNow.AddDate(0, 0, -1)
		return prev.Format("2006-01-02"), true
	case TaskTypeWeekly:
		prev := localNow.AddDate(0, 0, -7)
		year, week := prev.ISOWeek()
		return fmt.Sprintf("%04d-W%02d", year, week), true
	case TaskTypeLimited:
		if task.EndAt <= 0 || nowTime.Unix() <= task.EndAt {
			return "", false
		}
		return fmt.Sprintf("%s:%d:%d", task.TaskID, task.StartAt, task.EndAt), true
	default:
		return "", false
	}
}
