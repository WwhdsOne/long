package vote

import (
	"context"
	"testing"
	"time"

	"long/internal/nickname"
)

type stubTaskDefinitionStore struct {
	items []TaskDefinition
}

func (s stubTaskDefinitionStore) ListActiveTaskDefinitions(_ context.Context, _ int64) ([]TaskDefinition, error) {
	return s.items, nil
}

func (s stubTaskDefinitionStore) ListTaskDefinitions(_ context.Context) ([]TaskDefinition, error) {
	return s.items, nil
}

func (s stubTaskDefinitionStore) GetTaskDefinition(_ context.Context, taskID string) (*TaskDefinition, error) {
	for _, item := range s.items {
		if item.TaskID == taskID {
			copyItem := item
			return &copyItem, nil
		}
	}
	return nil, nil
}

func (s stubTaskDefinitionStore) UpsertTaskDefinition(_ context.Context, _ TaskDefinition) error {
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
	taskStore := NewStore(baseStore.client, "vote:", StoreOptions{
		CriticalChancePercent: 5,
		CriticalCount:         5,
		TaskDefinitionStore:   stubTaskDefinitionStore{items: defs},
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

	if err := store.recordTaskEvent(ctx, "阿明", TaskConditionDailyClicks, 1); err != nil {
		t.Fatalf("record first event: %v", err)
	}
	if err := store.recordTaskEvent(ctx, "阿明", TaskConditionDailyClicks, 1); err != nil {
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
