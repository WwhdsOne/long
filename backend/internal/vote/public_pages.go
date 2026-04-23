package vote

import (
	"context"
	"slices"
)

const (
	defaultPublicButtonPageSize int64 = 9
	maxPublicButtonPageSize     int64 = 50
)

// ButtonPage 描述公共按钮分页结果。
type ButtonPage struct {
	Items      []Button `json:"items"`
	Page       int64    `json:"page"`
	PageSize   int64    `json:"pageSize"`
	Total      int64    `json:"total"`
	TotalPages int64    `json:"totalPages"`
	TotalVotes int64    `json:"totalVotes"`
}

// ListButtonsPage 返回公共投票页。第一页固定优先展示当前星光按钮。
func (s *Store) ListButtonsPage(ctx context.Context, page int64, pageSize int64) (ButtonPage, error) {
	buttons, err := s.ListButtons(ctx)
	if err != nil {
		return ButtonPage{}, err
	}

	starlight := starlightStateForButtons(buttons, s.now())
	return buildButtonPage(buttons, starlight.ActiveKeys, page, pageSize), nil
}

func buildButtonPage(buttons []Button, activeKeys []string, page int64, pageSize int64) ButtonPage {
	page, pageSize = normalizePublicPage(page, pageSize)
	total := int64(len(buttons))
	totalVotes := int64(0)
	for _, button := range buttons {
		totalVotes += button.Count
	}

	featuredSet := make(map[string]struct{}, len(activeKeys))
	for _, key := range activeKeys {
		featuredSet[key] = struct{}{}
	}

	featured := make([]Button, 0, len(activeKeys))
	normal := make([]Button, 0, len(buttons))
	for _, button := range buttons {
		if _, ok := featuredSet[button.Key]; ok {
			featured = append(featured, button)
			continue
		}
		normal = append(normal, button)
	}

	firstNormalCount := pageSize - int64(len(featured))
	if firstNormalCount < 0 {
		firstNormalCount = 0
	}
	if firstNormalCount > int64(len(normal)) {
		firstNormalCount = int64(len(normal))
	}

	firstPageItems := make([]Button, 0, min(int(pageSize), len(featured)+int(firstNormalCount)))
	firstPageItems = append(firstPageItems, featured...)
	firstPageItems = append(firstPageItems, normal[:firstNormalCount]...)
	if int64(len(firstPageItems)) > pageSize {
		firstPageItems = firstPageItems[:pageSize]
	}

	remainingNormal := int64(len(normal)) - firstNormalCount
	totalPages := int64(1)
	if remainingNormal > 0 {
		totalPages += (remainingNormal + pageSize - 1) / pageSize
	}

	if page <= 1 {
		return ButtonPage{
			Items:      firstPageItems,
			Page:       1,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
			TotalVotes: totalVotes,
		}
	}

	offset := firstNormalCount + (page-2)*pageSize
	if offset >= int64(len(normal)) {
		return ButtonPage{
			Items:      []Button{},
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
			TotalVotes: totalVotes,
		}
	}

	end := min(offset+pageSize, int64(len(normal)))
	items := slices.Clone(normal[offset:end])
	return ButtonPage{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		TotalVotes: totalVotes,
	}
}

func normalizePublicPage(page int64, pageSize int64) (int64, int64) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultPublicButtonPageSize
	}
	if pageSize > maxPublicButtonPageSize {
		pageSize = maxPublicButtonPageSize
	}
	return page, pageSize
}
