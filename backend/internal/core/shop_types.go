package core

import "strings"

// ShopItemType 商店商品类型。
type ShopItemType string

const (
	ShopItemTypeBattleClickSkin ShopItemType = "battle_click_skin"
)

// ShopItem 商店商品目录模型。
type ShopItem struct {
	ItemID                     string       `json:"itemId" bson:"item_id"`
	Title                      string       `json:"title" bson:"title"`
	ItemType                   ShopItemType `json:"itemType" bson:"item_type"`
	PriceGold                  int64        `json:"priceGold" bson:"price_gold"`
	ImagePath                  string       `json:"imagePath" bson:"image_path"`
	ImageAlt                   string       `json:"imageAlt" bson:"image_alt"`
	PreviewImagePath           string       `json:"previewImagePath" bson:"preview_image_path"`
	BattleClickCursorImagePath string       `json:"battleClickCursorImagePath" bson:"battle_click_cursor_image_path"`
	Description                string       `json:"description" bson:"description"`
	Active                     bool         `json:"active" bson:"active"`
	SortOrder                  int64        `json:"sortOrder" bson:"sort_order"`
	AutoEquipOnPurchase        bool         `json:"autoEquipOnPurchase" bson:"auto_equip_on_purchase"`
	CreatedAt                  int64        `json:"createdAt,omitempty" bson:"created_at"`
	UpdatedAt                  int64        `json:"updatedAt,omitempty" bson:"updated_at"`
}

// ShopCatalogItemView 是面向前台的商店商品视图。
type ShopCatalogItemView struct {
	ShopItem
	Owned    bool `json:"owned"`
	Equipped bool `json:"equipped"`
}

// ShopPurchaseLog 记录一次商店购买。
type ShopPurchaseLog struct {
	ItemID      string       `json:"itemId" bson:"item_id"`
	Nickname    string       `json:"nickname" bson:"nickname"`
	ItemType    ShopItemType `json:"itemType" bson:"item_type"`
	PriceGold   int64        `json:"priceGold" bson:"price_gold"`
	PurchasedAt int64        `json:"purchasedAt" bson:"purchased_at"`
	Equipped    bool         `json:"equipped" bson:"equipped"`
}

// ShopActionResult 用于返回购买/切换后的玩家态。
type ShopActionResult struct {
	ItemID    string    `json:"itemId"`
	UserState UserState `json:"userState"`
}

// NormalizeShopItemModel 规范化商店商品字段。
func NormalizeShopItemModel(item ShopItem) ShopItem {
	item.ItemID = strings.TrimSpace(item.ItemID)
	item.Title = strings.TrimSpace(item.Title)
	item.ImagePath = strings.TrimSpace(item.ImagePath)
	item.ImageAlt = strings.TrimSpace(item.ImageAlt)
	item.PreviewImagePath = strings.TrimSpace(item.PreviewImagePath)
	item.BattleClickCursorImagePath = strings.TrimSpace(item.BattleClickCursorImagePath)
	item.Description = strings.TrimSpace(item.Description)
	if item.ItemType == "" {
		item.ItemType = ShopItemTypeBattleClickSkin
	}
	if item.PriceGold < 0 {
		item.PriceGold = 0
	}
	if !item.AutoEquipOnPurchase {
		item.AutoEquipOnPurchase = true
	}
	return item
}
