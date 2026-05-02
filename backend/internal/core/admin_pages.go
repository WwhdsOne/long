package core

import "context"

const (
	defaultAdminPageSize = 20
	maxAdminPageSize     = 100
)

type AdminEquipmentPage struct {
	Items      []EquipmentDefinition `json:"items"`
	Page       int64                 `json:"page"`
	PageSize   int64                 `json:"pageSize"`
	Total      int64                 `json:"total"`
	TotalPages int64                 `json:"totalPages"`
}

type AdminBossHistoryPage struct {
	Items      []BossHistoryEntry `json:"items"`
	Page       int64              `json:"page"`
	PageSize   int64              `json:"pageSize"`
	Total      int64              `json:"total"`
	TotalPages int64              `json:"totalPages"`
}

func (s *Store) ListAdminEquipmentPage(ctx context.Context, page int64, pageSize int64) (AdminEquipmentPage, error) {
	items, err := s.ListEquipmentDefinitions(ctx)
	if err != nil {
		return AdminEquipmentPage{}, err
	}

	items, page, pageSize, total, totalPages := paginateAdminItems(items, page, pageSize)
	return AdminEquipmentPage{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *Store) ListAdminBossHistoryPage(ctx context.Context, page int64, pageSize int64) (AdminBossHistoryPage, error) {
	if s.bossHistoryStore != nil {
		return s.bossHistoryStore.ListAdminBossHistoryPage(ctx, page, pageSize)
	}

	items, err := s.ListBossHistory(ctx)
	if err != nil {
		return AdminBossHistoryPage{}, err
	}

	items, page, pageSize, total, totalPages := paginateAdminItems(items, page, pageSize)
	return AdminBossHistoryPage{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func normalizeAdminPage(page int64, pageSize int64) (int64, int64) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultAdminPageSize
	}
	if pageSize > maxAdminPageSize {
		pageSize = maxAdminPageSize
	}
	return page, pageSize
}

func paginateAdminItems[T any](items []T, page int64, pageSize int64) ([]T, int64, int64, int64, int64) {
	page, pageSize = normalizeAdminPage(page, pageSize)

	total := int64(len(items))
	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	start := (page - 1) * pageSize
	if start >= total {
		return []T{}, page, pageSize, total, totalPages
	}

	end := min(start+pageSize, total)

	return items[start:end], page, pageSize, total, totalPages
}
