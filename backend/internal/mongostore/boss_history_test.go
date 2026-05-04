package mongostore

import (
	"errors"
	"testing"

	"long/internal/core"
)

func TestCollectAllBossHistoryPagesReadsPastFirstHundred(t *testing.T) {
	const total = 250

	items, err := collectAllBossHistoryPages(func(page int64, pageSize int64) (core.AdminBossHistoryPage, error) {
		start := int((page - 1) * pageSize)
		if start >= total {
			return core.AdminBossHistoryPage{
				Items:      []core.BossHistoryEntry{},
				Page:       page,
				PageSize:   pageSize,
				Total:      total,
				TotalPages: 3,
			}, nil
		}

		end := min(start+int(pageSize), total)

		rows := make([]core.BossHistoryEntry, 0, end-start)
		for index := start; index < end; index++ {
			rows = append(rows, core.BossHistoryEntry{
				Boss: core.Boss{ID: "boss"},
			})
		}

		return core.AdminBossHistoryPage{
			Items:      rows,
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: 3,
		}, nil
	}, 100)
	if err != nil {
		t.Fatalf("collect all boss history pages: %v", err)
	}
	if len(items) != total {
		t.Fatalf("expected %d items, got %d", total, len(items))
	}
}

func TestCollectAllBossHistoryPagesStopsOnFetchError(t *testing.T) {
	wantErr := errors.New("boom")

	_, err := collectAllBossHistoryPages(func(page int64, pageSize int64) (core.AdminBossHistoryPage, error) {
		if page == 2 {
			return core.AdminBossHistoryPage{}, wantErr
		}
		return core.AdminBossHistoryPage{
			Items:      []core.BossHistoryEntry{{Boss: core.Boss{ID: "boss-1"}}},
			Page:       page,
			PageSize:   pageSize,
			Total:      2,
			TotalPages: 2,
		}, nil
	}, 100)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}
