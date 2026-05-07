package core

import (
	"context"
	"testing"
	"time"

	"long/internal/nickname"
)

type stubTaskDefinitionStore struct {
	items    []TaskDefinition
	upserted []TaskDefinition
}

func (s *stubTaskDefinitionStore) ListActiveTaskDefinitions(_ context.Context, _ int64) ([]TaskDefinition, error) {
	return s.items, nil
}

func (s *stubTaskDefinitionStore) ListTaskDefinitions(_ context.Context) ([]TaskDefinition, error) {
	return s.items, nil
}

func (s *stubTaskDefinitionStore) GetTaskDefinition(_ context.Context, taskID string) (*TaskDefinition, error) {
	for _, item := range s.items {
		if item.TaskID == taskID {
			copyItem := item
			return &copyItem, nil
		}
	}
	return nil, nil
}

func (s *stubTaskDefinitionStore) UpsertTaskDefinition(_ context.Context, item TaskDefinition) error {
	s.upserted = append(s.upserted, item)
	replaced := false
	for index := range s.items {
		if s.items[index].TaskID == item.TaskID {
			s.items[index] = item
			replaced = true
			break
		}
	}
	if !replaced {
		s.items = append(s.items, item)
	}
	return nil
}

type stubTaskClaimLogStore struct {
	claimed map[string]bool
}

func (s stubTaskClaimLogStore) HasTaskClaimed(_ context.Context, taskID string, cycleKey string, nickname string) (bool, error) {
	if s.claimed == nil {
		return false, nil
	}
	return s.claimed[taskID+":"+cycleKey+":"+nickname], nil
}

func (s stubTaskClaimLogStore) WriteTaskClaimLog(_ context.Context, _ TaskClaimLog) error {
	return nil
}

func (s stubTaskClaimLogStore) ListTaskClaimLogs(_ context.Context, _ string, _ string) ([]TaskClaimLog, error) {
	return []TaskClaimLog{}, nil
}

func (s stubTaskClaimLogStore) HasTaskClaimLog(_ context.Context, _ string, _ string) (bool, error) {
	return false, nil
}

func newTaskTestStore(t *testing.T, defs []TaskDefinition) (*Store, func()) {
	t.Helper()
	baseStore, cleanup := newTestStore(t)
	taskDefStore := &stubTaskDefinitionStore{items: defs}
	taskStore := NewStore(baseStore.client, "vote:", StoreOptions{
		CriticalChancePercent: 5,
		TaskDefinitionStore:   taskDefStore,
		TaskClaimLogStore:     stubTaskClaimLogStore{},
	}, nickname.NewValidator([]string{"习近平", "xjp"}))
	return taskStore, cleanup
}

func TestListTasksForPlayerReturnsRedisProgress(t *testing.T) {
	store, cleanup := newTaskTestStore(t, []TaskDefinition{{
		TaskID:        "daily-click-1",
		Title:         "今日点击",
		TaskType:      TaskTypeDaily,
		Status:        TaskStatusActive,
		ConditionKind: TaskConditionDailyClicks,
		TargetValue:   3,
	}})
	defer cleanup()

	nowTime := time.Date(2026, 5, 1, 10, 0, 0, 0, taskTimeLocation)
	store.now = func() time.Time { return nowTime }
	ctx := context.Background()
	cycleKey := currentTaskCycleKey(TaskDefinition{TaskID: "daily-click-1", TaskType: TaskTypeDaily}, nowTime)
	if err := store.client.HSet(ctx, store.taskProgressKey("阿明", "daily-click-1", cycleKey), map[string]any{
		"progress":     "3",
		"status":       string(TaskPlayerStatusCompletedUnclaimed),
		"completed_at": "123",
	}).Err(); err != nil {
		t.Fatalf("seed task progress: %v", err)
	}

	items, err := store.ListTasksForPlayer(ctx, "阿明")
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one task, got %+v", items)
	}
	if items[0].Progress != 3 || items[0].Status != TaskPlayerStatusCompletedUnclaimed || !items[0].CanClaim {
		t.Fatalf("unexpected task view: %+v", items[0])
	}
}

func TestRecordTaskEventUpdatesProgressAndParticipants(t *testing.T) {
	store, cleanup := newTaskTestStore(t, []TaskDefinition{{
		TaskID:        "daily-click-1",
		Title:         "今日点击",
		TaskType:      TaskTypeDaily,
		Status:        TaskStatusActive,
		ConditionKind: TaskConditionDailyClicks,
		TargetValue:   2,
	}})
	defer cleanup()

	nowTime := time.Date(2026, 5, 1, 10, 0, 0, 0, taskTimeLocation)
	store.now = func() time.Time { return nowTime }
	ctx := context.Background()

	if err := store.recordTaskEvent(ctx, "阿明", TaskEventClick, 1); err != nil {
		t.Fatalf("record first event: %v", err)
	}
	if err := store.recordTaskEvent(ctx, "阿明", TaskEventClick, 1); err != nil {
		t.Fatalf("record second event: %v", err)
	}

	items, err := store.ListTasksForPlayer(ctx, "阿明")
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one task, got %+v", items)
	}
	if items[0].Progress != 2 || items[0].Status != TaskPlayerStatusCompletedUnclaimed {
		t.Fatalf("unexpected task progress after events: %+v", items[0])
	}

	participantKey := store.taskParticipantsKey("daily-click-1", currentTaskCycleKey(TaskDefinition{TaskID: "daily-click-1", TaskType: TaskTypeDaily}, nowTime))
	exists, err := store.client.SIsMember(ctx, participantKey, "阿明").Result()
	if err != nil {
		t.Fatalf("check participant set: %v", err)
	}
	if !exists {
		t.Fatal("expected participant set to contain 阿明")
	}
}

func TestSaveTaskDefinitionRejectsCoreChangeAfterActivation(t *testing.T) {
	store, cleanup := newTaskTestStore(t, []TaskDefinition{{
		TaskID:        "daily-click-1",
		Title:         "今日点击",
		TaskType:      TaskTypeDaily,
		Status:        TaskStatusActive,
		ConditionKind: TaskConditionDailyClicks,
		TargetValue:   3,
		CreatedAt:     1,
	}})
	defer cleanup()

	err := store.SaveTaskDefinition(context.Background(), TaskDefinition{
		TaskID:        "daily-click-1",
		Title:         "今日点击改版",
		TaskType:      TaskTypeDaily,
		Status:        TaskStatusActive,
		ConditionKind: TaskConditionEnhanceCount,
		TargetValue:   3,
	})
	if err != ErrTaskImmutable {
		t.Fatalf("expected ErrTaskImmutable, got %v", err)
	}
}

func TestSaveTaskDefinitionRejectsLimitedTaskWithoutValidWindow(t *testing.T) {
	store, cleanup := newTaskTestStore(t, nil)
	defer cleanup()

	err := store.SaveTaskDefinition(context.Background(), TaskDefinition{
		TaskID:        "limited-click",
		Title:         "限时点击",
		TaskType:      TaskTypeLimited,
		Status:        TaskStatusDraft,
		ConditionKind: TaskConditionDailyClicks,
		TargetValue:   10,
		Rewards: TaskRewards{
			Gold: 100,
		},
		StartAt: 200,
		EndAt:   100,
	})
	if err != ErrTaskNotClaimable {
		t.Fatalf("expected ErrTaskNotClaimable, got %v", err)
	}
}

func TestSaveTaskDefinitionRejectsTaskWithoutRewards(t *testing.T) {
	store, cleanup := newTaskTestStore(t, nil)
	defer cleanup()

	err := store.SaveTaskDefinition(context.Background(), TaskDefinition{
		TaskID:        "daily-click",
		Title:         "今日点击",
		TaskType:      TaskTypeDaily,
		Status:        TaskStatusDraft,
		ConditionKind: TaskConditionDailyClicks,
		TargetValue:   10,
	})
	if err != ErrTaskNotClaimable {
		t.Fatalf("expected ErrTaskNotClaimable, got %v", err)
	}
}

func TestSaveTaskDefinitionNormalizesLegacyFields(t *testing.T) {
	baseStore, cleanup := newTestStore(t)
	defer cleanup()

	taskDefStore := &stubTaskDefinitionStore{}
	store := NewStore(baseStore.client, "vote:", StoreOptions{
		CriticalChancePercent: 5,
		TaskDefinitionStore:   taskDefStore,
		TaskClaimLogStore:     stubTaskClaimLogStore{},
	}, nickname.NewValidator([]string{"习近平", "xjp"}))

	err := store.SaveTaskDefinition(context.Background(), TaskDefinition{
		TaskID:        "daily-click",
		Title:         "今日点击",
		TaskType:      TaskTypeDaily,
		Status:        TaskStatusDraft,
		ConditionKind: TaskConditionDailyClicks,
		TargetValue:   10,
		Rewards: TaskRewards{
			Gold: 100,
		},
	})
	if err != nil {
		t.Fatalf("save task definition: %v", err)
	}
	if len(taskDefStore.upserted) != 1 {
		t.Fatalf("expected one upserted task, got %d", len(taskDefStore.upserted))
	}
	item := taskDefStore.upserted[0]
	if item.EventKind != TaskEventClick {
		t.Fatalf("expected event kind click, got %q", item.EventKind)
	}
	if item.WindowKind != TaskWindowDaily {
		t.Fatalf("expected window kind daily, got %q", item.WindowKind)
	}
}

func TestSaveTaskDefinitionRejectsFixedRangeWithoutValidWindow(t *testing.T) {
	store, cleanup := newTaskTestStore(t, nil)
	defer cleanup()

	err := store.SaveTaskDefinition(context.Background(), TaskDefinition{
		TaskID:      "campaign-click",
		Title:       "活动点击",
		Status:      TaskStatusDraft,
		EventKind:   TaskEventClick,
		WindowKind:  TaskWindowFixedRange,
		TargetValue: 10,
		Rewards: TaskRewards{
			Gold: 100,
		},
		StartAt: 200,
		EndAt:   100,
	})
	if err != ErrTaskNotClaimable {
		t.Fatalf("expected ErrTaskNotClaimable, got %v", err)
	}
}

func TestRecordTaskEventUpdatesFixedRangeTaskProgress(t *testing.T) {
	store, cleanup := newTaskTestStore(t, []TaskDefinition{{
		TaskID:      "campaign-click-1",
		Title:       "活动点击",
		Status:      TaskStatusActive,
		EventKind:   TaskEventClick,
		WindowKind:  TaskWindowFixedRange,
		TargetValue: 2,
		StartAt:     time.Date(2026, 5, 2, 0, 0, 0, 0, taskTimeLocation).Unix(),
		EndAt:       time.Date(2026, 5, 5, 23, 59, 59, 0, taskTimeLocation).Unix(),
	}})
	defer cleanup()

	nowTime := time.Date(2026, 5, 3, 10, 0, 0, 0, taskTimeLocation)
	store.now = func() time.Time { return nowTime }
	ctx := context.Background()

	if err := store.recordTaskEvent(ctx, "阿明", TaskEventClick, 1); err != nil {
		t.Fatalf("record first event: %v", err)
	}
	if err := store.recordTaskEvent(ctx, "阿明", TaskEventClick, 1); err != nil {
		t.Fatalf("record second event: %v", err)
	}

	items, err := store.ListTasksForPlayer(ctx, "阿明")
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one task, got %+v", items)
	}
	if items[0].Progress != 2 || items[0].Status != TaskPlayerStatusCompletedUnclaimed {
		t.Fatalf("unexpected task progress after events: %+v", items[0])
	}
	expectedCycleKey := "campaign-click-1:1777651200:1777996799"
	if items[0].CycleKey != expectedCycleKey {
		t.Fatalf("expected cycle key %q, got %q", expectedCycleKey, items[0].CycleKey)
	}
}

func TestSaveTaskDefinitionSupportsLifetimeWindow(t *testing.T) {
	baseStore, cleanup := newTestStore(t)
	defer cleanup()

	taskDefStore := &stubTaskDefinitionStore{}
	store := NewStore(baseStore.client, "vote:", StoreOptions{
		CriticalChancePercent: 5,
		TaskDefinitionStore:   taskDefStore,
		TaskClaimLogStore:     stubTaskClaimLogStore{},
	}, nickname.NewValidator([]string{"习近平", "xjp"}))

	err := store.SaveTaskDefinition(context.Background(), TaskDefinition{
		TaskID:      "longterm-click",
		Title:       "长期点击",
		Status:      TaskStatusDraft,
		EventKind:   TaskEventClick,
		WindowKind:  TaskWindowLifetime,
		TargetValue: 100,
		Rewards: TaskRewards{
			Gold: 100,
		},
	})
	if err != nil {
		t.Fatalf("save lifetime task definition: %v", err)
	}
	if len(taskDefStore.upserted) != 1 {
		t.Fatalf("expected one upserted task, got %d", len(taskDefStore.upserted))
	}
	item := taskDefStore.upserted[0]
	if item.WindowKind != TaskWindowLifetime {
		t.Fatalf("expected window kind lifetime, got %q", item.WindowKind)
	}
	if item.TaskType != TaskTypeLongTerm {
		t.Fatalf("expected task type long_term, got %q", item.TaskType)
	}
}

func TestRecordTaskEventUpdatesLifetimeTaskProgress(t *testing.T) {
	store, cleanup := newTaskTestStore(t, []TaskDefinition{{
		TaskID:      "longterm-click-1",
		Title:       "长期点击",
		Status:      TaskStatusActive,
		EventKind:   TaskEventClick,
		WindowKind:  TaskWindowLifetime,
		TargetValue: 2,
	}})
	defer cleanup()

	nowTime := time.Date(2026, 5, 3, 10, 0, 0, 0, taskTimeLocation)
	store.now = func() time.Time { return nowTime }
	ctx := context.Background()

	if err := store.recordTaskEvent(ctx, "阿明", TaskEventClick, 1); err != nil {
		t.Fatalf("record first event: %v", err)
	}
	if err := store.recordTaskEvent(ctx, "阿明", TaskEventClick, 1); err != nil {
		t.Fatalf("record second event: %v", err)
	}

	items, err := store.ListTasksForPlayer(ctx, "阿明")
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one task, got %+v", items)
	}
	if items[0].Progress != 2 || items[0].Status != TaskPlayerStatusCompletedUnclaimed {
		t.Fatalf("unexpected lifetime task progress after events: %+v", items[0])
	}
	if items[0].CycleKey != "longterm-click-1" {
		t.Fatalf("expected lifetime cycle key to use task id, got %q", items[0].CycleKey)
	}
}

func TestEnhanceItemBatchRecordsEnhanceTaskProgressByLevels(t *testing.T) {
	store, cleanup := newTaskTestStore(t, []TaskDefinition{{
		TaskID:        "enhance-3",
		Title:         "强化三次",
		Status:        TaskStatusActive,
		EventKind:     TaskEventEnhance,
		ConditionKind: TaskConditionEnhanceCount,
		WindowKind:    TaskWindowDaily,
		TargetValue:   3,
		Rewards: TaskRewards{
			Gold: 100,
		},
	}})
	defer cleanup()

	ctx := context.Background()
	nickname := "阿明"
	nowTime := time.Date(2026, 5, 8, 10, 0, 0, 0, taskTimeLocation)
	store.now = func() time.Time { return nowTime }
	seedEquipmentDefinitionWithRarity(t, store, ctx, "wood-sword", "weapon", "普通", 10)
	instanceID := seedOwnedInstance(t, store, ctx, nickname, "wood-sword")

	if err := store.client.HSet(ctx, store.resourceKey(nickname), map[string]any{
		"gold":   "5000",
		"stones": "30",
	}).Err(); err != nil {
		t.Fatalf("seed resource: %v", err)
	}

	if _, err := store.EnhanceItemBatch(ctx, nickname, instanceID, 3); err != nil {
		t.Fatalf("enhance item batch: %v", err)
	}

	items, err := store.ListTasksForPlayer(ctx, nickname)
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one task, got %+v", items)
	}
	if items[0].Progress != 3 {
		t.Fatalf("expected enhance task progress 3, got %+v", items[0])
	}
	if items[0].Status != TaskPlayerStatusCompletedUnclaimed || !items[0].CanClaim {
		t.Fatalf("expected enhance task claimable after batch enhance, got %+v", items[0])
	}
}

func TestDuplicateTaskDefinitionRejectsExistingTargetID(t *testing.T) {
	store, cleanup := newTaskTestStore(t, []TaskDefinition{
		{
			TaskID:        "daily-click",
			Title:         "今日点击",
			TaskType:      TaskTypeDaily,
			Status:        TaskStatusDraft,
			ConditionKind: TaskConditionDailyClicks,
			TargetValue:   10,
			Rewards: TaskRewards{
				Gold: 100,
			},
		},
		{
			TaskID:        "existing-copy",
			Title:         "旧副本",
			TaskType:      TaskTypeDaily,
			Status:        TaskStatusDraft,
			ConditionKind: TaskConditionDailyClicks,
			TargetValue:   5,
			Rewards: TaskRewards{
				Gold: 10,
			},
		},
	})
	defer cleanup()

	_, err := store.DuplicateTaskDefinition(context.Background(), "daily-click", "existing-copy")
	if err != ErrTaskImmutable {
		t.Fatalf("expected ErrTaskImmutable, got %v", err)
	}
}
