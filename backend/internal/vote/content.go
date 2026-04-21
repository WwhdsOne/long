package vote

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	nicknamefilter "long/internal/nickname"
)

const (
	maxMessageRunes = 200
)

type equipmentUpgrade struct {
	StarLevel                  int
	BonusClicks                int64
	BonusCriticalChancePercent int
	BonusCriticalCount         int64
	ReforgePityCounter         int
}

// GetLatestAnnouncement 返回最新生效公告。
func (s *Store) GetLatestAnnouncement(ctx context.Context) (*Announcement, error) {
	items, err := s.ListAnnouncements(ctx, false)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

// ListAnnouncements 返回公告列表；公开接口只看 active=true，后台接口可以看全部。
func (s *Store) ListAnnouncements(ctx context.Context, includeInactive bool) ([]Announcement, error) {
	ids, err := s.client.ZRevRange(ctx, s.announcementKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []Announcement{}, nil
	}

	items := make([]Announcement, 0, len(ids))
	for _, id := range ids {
		announcement, loadErr := s.loadAnnouncement(ctx, id)
		if loadErr != nil || announcement == nil {
			continue
		}
		if !includeInactive && !announcement.Active {
			continue
		}
		items = append(items, *announcement)
	}

	return items, nil
}

// SaveAnnouncement 创建一条新公告。
func (s *Store) SaveAnnouncement(ctx context.Context, announcement AnnouncementUpsert) (*Announcement, error) {
	title := strings.TrimSpace(announcement.Title)
	content := strings.TrimSpace(announcement.Content)
	if title == "" {
		title = "更新公告"
	}
	if content == "" {
		return nil, ErrMessageEmpty
	}

	id, err := s.client.Incr(ctx, s.announcementSeqKey).Result()
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	item := &Announcement{
		ID:          strconv.FormatInt(id, 10),
		Title:       title,
		Content:     content,
		PublishedAt: now,
		Active:      announcement.Active,
	}

	if err := s.client.HSet(ctx, s.announcementItemKey(item.ID), map[string]any{
		"id":           item.ID,
		"title":        item.Title,
		"content":      item.Content,
		"published_at": strconv.FormatInt(item.PublishedAt, 10),
		"active":       boolToRedis(item.Active),
	}).Err(); err != nil {
		return nil, err
	}
	if err := s.client.ZAdd(ctx, s.announcementKey, redis.Z{
		Score:  float64(id),
		Member: item.ID,
	}).Err(); err != nil {
		return nil, err
	}

	return item, nil
}

// DeleteAnnouncement 删除公告。
func (s *Store) DeleteAnnouncement(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil
	}
	pipe := s.client.TxPipeline()
	pipe.ZRem(ctx, s.announcementKey, id)
	pipe.Del(ctx, s.announcementItemKey(id))
	_, err := pipe.Exec(ctx)
	return err
}

// CreateMessage 发送一条公共留言。
func (s *Store) CreateMessage(ctx context.Context, nickname string, content string) (*Message, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return nil, err
	}

	normalizedContent, err := s.validatedMessageContent(content)
	if err != nil {
		return nil, err
	}

	id, err := s.client.Incr(ctx, s.messageSeqKey).Result()
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	item := &Message{
		ID:        strconv.FormatInt(id, 10),
		Nickname:  normalizedNickname,
		Content:   normalizedContent,
		CreatedAt: now,
	}

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, s.messageItemKey(item.ID), map[string]any{
		"id":         item.ID,
		"nickname":   item.Nickname,
		"content":    item.Content,
		"created_at": strconv.FormatInt(item.CreatedAt, 10),
	})
	pipe.ZAdd(ctx, s.messageKey, redis.Z{
		Score:  float64(id),
		Member: item.ID,
	})
	pipe.ZAdd(ctx, s.playerIndexKey, redis.Z{
		Score:  float64(now),
		Member: item.Nickname,
	})
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	return item, nil
}

// ListMessages 返回公共留言分页。
func (s *Store) ListMessages(ctx context.Context, cursor string, limit int64) (MessagePage, error) {
	if limit <= 0 {
		limit = 50
	}

	rangeBy := &redis.ZRangeBy{
		Max:    "+inf",
		Min:    "-inf",
		Offset: 0,
		Count:  limit + 1,
	}
	if trimmed := strings.TrimSpace(cursor); trimmed != "" {
		if _, err := strconv.ParseInt(trimmed, 10, 64); err != nil {
			return MessagePage{}, err
		}
		rangeBy.Max = fmt.Sprintf("(%s", trimmed)
	}

	ids, err := s.client.ZRevRangeByScore(ctx, s.messageKey, rangeBy).Result()
	if err != nil {
		return MessagePage{}, err
	}
	if len(ids) == 0 {
		return MessagePage{Items: []Message{}}, nil
	}

	page := MessagePage{
		Items: make([]Message, 0, limit),
	}

	for index, id := range ids {
		if int64(index) >= limit {
			page.NextCursor = page.Items[len(page.Items)-1].ID
			break
		}
		message, loadErr := s.loadMessage(ctx, id)
		if loadErr != nil || message == nil {
			continue
		}
		page.Items = append(page.Items, *message)
	}

	return page, nil
}

// DeleteMessage 删除一条留言。
func (s *Store) DeleteMessage(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil
	}
	pipe := s.client.TxPipeline()
	pipe.ZRem(ctx, s.messageKey, id)
	pipe.Del(ctx, s.messageItemKey(id))
	_, err := pipe.Exec(ctx)
	return err
}

// SynthesizeItem 对指定装备类型执行一次升星。
func (s *Store) SynthesizeItem(ctx context.Context, nickname string, itemID string) (State, error) {
	normalizedNickname, err := s.validatedNickname(nickname)
	if err != nil {
		return State{}, err
	}

	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return State{}, ErrEquipmentNotFound
	}
	if _, err := s.getEquipmentDefinition(ctx, itemID); err != nil {
		return State{}, err
	}

	quantity, err := s.client.HGet(ctx, s.inventoryKey(normalizedNickname), itemID).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return State{}, ErrEquipmentNotOwned
		}
		return State{}, err
	}
	if quantity < 3 {
		return State{}, ErrEquipmentNotEnough
	}

	upgrade, err := s.getEquipmentUpgrade(ctx, normalizedNickname, itemID)
	if err != nil {
		return State{}, err
	}
	switch s.roll(3) {
	case 0:
		upgrade.BonusClicks++
	case 1:
		upgrade.BonusCriticalChancePercent++
	default:
		upgrade.BonusCriticalCount++
	}
	upgrade.StarLevel++

	pipe := s.client.TxPipeline()
	pipe.HIncrBy(ctx, s.inventoryKey(normalizedNickname), itemID, -2)
	pipe.HSet(ctx, s.upgradeKey(normalizedNickname, itemID), map[string]any{
		"star_level":                    strconv.Itoa(upgrade.StarLevel),
		"bonus_clicks":                  strconv.FormatInt(upgrade.BonusClicks, 10),
		"bonus_critical_chance_percent": strconv.Itoa(upgrade.BonusCriticalChancePercent),
		"bonus_critical_count":          strconv.FormatInt(upgrade.BonusCriticalCount, 10),
	})
	pipe.ZAdd(ctx, s.playerIndexKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: normalizedNickname,
	})
	if _, err := pipe.Exec(ctx); err != nil {
		return State{}, err
	}

	return s.GetState(ctx, normalizedNickname)
}

func (s *Store) loadAnnouncement(ctx context.Context, id string) (*Announcement, error) {
	values, err := s.client.HGetAll(ctx, s.announcementItemKey(id)).Result()
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	return &Announcement{
		ID:          firstNonEmpty(strings.TrimSpace(values["id"]), id),
		Title:       strings.TrimSpace(values["title"]),
		Content:     strings.TrimSpace(values["content"]),
		PublishedAt: int64FromString(values["published_at"]),
		Active:      strings.TrimSpace(values["active"]) != "0",
	}, nil
}

func (s *Store) loadMessage(ctx context.Context, id string) (*Message, error) {
	values, err := s.client.HGetAll(ctx, s.messageItemKey(id)).Result()
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	return &Message{
		ID:        firstNonEmpty(strings.TrimSpace(values["id"]), id),
		Nickname:  strings.TrimSpace(values["nickname"]),
		Content:   strings.TrimSpace(values["content"]),
		CreatedAt: int64FromString(values["created_at"]),
	}, nil
}

func (s *Store) validatedMessageContent(content string) (string, error) {
	trimmed := strings.TrimSpace(content)
	switch {
	case trimmed == "":
		return "", ErrMessageEmpty
	case len([]rune(trimmed)) > maxMessageRunes:
		return "", ErrMessageTooLong
	}

	if s.validator != nil {
		if err := s.validator.Validate(trimmed); err != nil {
			if errors.Is(err, nicknamefilter.ErrSensitiveNickname) {
				return "", ErrSensitiveContent
			}
			return "", err
		}
	}

	return trimmed, nil
}

func (s *Store) getEquipmentUpgrade(ctx context.Context, nickname string, itemID string) (equipmentUpgrade, error) {
	if strings.TrimSpace(nickname) == "" || strings.TrimSpace(itemID) == "" {
		return equipmentUpgrade{}, nil
	}

	values, err := s.client.HGetAll(ctx, s.upgradeKey(nickname, itemID)).Result()
	if err != nil {
		return equipmentUpgrade{}, err
	}
	if len(values) == 0 {
		return equipmentUpgrade{}, nil
	}

	return equipmentUpgrade{
		StarLevel:                  int(int64FromString(values["star_level"])),
		BonusClicks:                int64FromString(values["bonus_clicks"]),
		BonusCriticalChancePercent: int(int64FromString(values["bonus_critical_chance_percent"])),
		BonusCriticalCount:         int64FromString(values["bonus_critical_count"]),
		ReforgePityCounter:         int(int64FromString(values["reforge_pity_counter"])),
	}, nil
}

func (s *Store) buildInventoryItem(definition EquipmentDefinition, upgrade equipmentUpgrade, quantity int64, equipped bool) InventoryItem {
	return InventoryItem{
		ItemID:                     definition.ItemID,
		Name:                       displayItemName(definition.Name, upgrade.StarLevel),
		Slot:                       definition.Slot,
		Quantity:                   quantity,
		StarLevel:                  upgrade.StarLevel,
		ReforgePityCounter:         upgrade.ReforgePityCounter,
		BonusClicks:                definition.BonusClicks + upgrade.BonusClicks,
		BonusCriticalChancePercent: definition.BonusCriticalChancePercent + upgrade.BonusCriticalChancePercent,
		BonusCriticalCount:         definition.BonusCriticalCount + upgrade.BonusCriticalCount,
		Equipped:                   equipped,
	}
}

func unknownInventoryItem(itemID string, upgrade equipmentUpgrade, quantity int64, equipped bool) InventoryItem {
	return InventoryItem{
		ItemID:                     itemID,
		Name:                       displayItemName(itemID, upgrade.StarLevel),
		Quantity:                   quantity,
		StarLevel:                  upgrade.StarLevel,
		ReforgePityCounter:         upgrade.ReforgePityCounter,
		BonusClicks:                upgrade.BonusClicks,
		BonusCriticalChancePercent: upgrade.BonusCriticalChancePercent,
		BonusCriticalCount:         upgrade.BonusCriticalCount,
		Equipped:                   equipped,
	}
}

func displayItemName(baseName string, starLevel int) string {
	name := firstNonEmpty(strings.TrimSpace(baseName), "未命名装备")
	if starLevel <= 0 {
		return name
	}
	return fmt.Sprintf("%s +%d", name, starLevel)
}
