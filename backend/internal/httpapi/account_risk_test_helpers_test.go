package httpapi

import (
	"context"
	"time"

	"long/internal/core"
)

type mockClickRiskDetector struct {
	hit      bool
	hitCount int
	err      error
	calls    []string
}

func (m *mockClickRiskDetector) Detect(key string) (bool, error) {
	m.calls = append(m.calls, key)
	return m.hit, m.err
}

func (m *mockClickRiskDetector) DetectCount(key string) (int, error) {
	m.calls = append(m.calls, key)
	if m.err != nil {
		return 0, m.err
	}
	if m.hitCount > 0 {
		return m.hitCount, nil
	}
	if m.hit {
		return 1, nil
	}
	return 0, nil
}

type mockAccountRiskManager struct {
	entries  []core.AccountRiskState
	recorded []struct {
		nickname string
		event    core.AccountRiskEvent
	}
	lastClearedNickname string
	recordErr           error
	listErr             error
	clearErr            error
}

func (m *mockAccountRiskManager) RecordAccountRiskEvent(_ context.Context, nickname string, event core.AccountRiskEvent) (core.AccountRiskState, error) {
	if m.recordErr != nil {
		return core.AccountRiskState{}, m.recordErr
	}
	m.recorded = append(m.recorded, struct {
		nickname string
		event    core.AccountRiskEvent
	}{nickname: nickname, event: event})
	for index := range m.entries {
		if m.entries[index].Nickname == nickname {
			m.entries[index].Score++
			return m.entries[index], nil
		}
	}
	state := core.AccountRiskState{Nickname: nickname, Score: 1}
	m.entries = append(m.entries, state)
	return state, nil
}

func (m *mockAccountRiskManager) GetAccountRiskState(_ context.Context, nickname string) (core.AccountRiskState, error) {
	for _, entry := range m.entries {
		if entry.Nickname == nickname {
			return entry, nil
		}
	}
	return core.AccountRiskState{Nickname: nickname}, nil
}

func (m *mockAccountRiskManager) ListAccountRiskEntries(_ context.Context) ([]core.AccountRiskState, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return append([]core.AccountRiskState(nil), m.entries...), nil
}

func (m *mockAccountRiskManager) ClearAccountRiskState(_ context.Context, nickname string) error {
	if m.clearErr != nil {
		return m.clearErr
	}
	m.lastClearedNickname = nickname
	next := make([]core.AccountRiskState, 0, len(m.entries))
	for _, entry := range m.entries {
		if entry.Nickname != nickname {
			next = append(next, entry)
		}
	}
	m.entries = next
	return nil
}

var _ = time.Second
