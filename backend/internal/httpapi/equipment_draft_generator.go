package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"long/internal/vote"

	"github.com/bytedance/sonic"
)

var ErrInvalidEquipmentDraft = errors.New("invalid equipment draft")

type EquipmentDraftGeneratorConfig struct {
	APIKey  string
	BaseURL string
	Model   string
	Timeout time.Duration
}

type OpenAIEquipmentDraftGenerator struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewOpenAIEquipmentDraftGenerator(config EquipmentDraftGeneratorConfig) *OpenAIEquipmentDraftGenerator {
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &OpenAIEquipmentDraftGenerator{
		apiKey:  strings.TrimSpace(config.APIKey),
		baseURL: strings.TrimRight(strings.TrimSpace(config.BaseURL), "/"),
		model:   strings.TrimSpace(config.Model),
		client:  &http.Client{Timeout: timeout},
	}
}

func (g *OpenAIEquipmentDraftGenerator) GenerateEquipmentDraft(ctx context.Context, prompt string) (vote.EquipmentDefinition, error) {
	if strings.TrimSpace(prompt) == "" {
		return vote.EquipmentDefinition{}, fmt.Errorf("%w: prompt is required", ErrInvalidEquipmentDraft)
	}

	body, err := sonic.Marshal(chatCompletionRequest{
		Model: g.model,
		Store: false,
		Messages: []chatMessage{
			{Role: "system", Content: equipmentDraftSystemPrompt()},
			{Role: "user", Content: strings.TrimSpace(prompt)},
		},
		ResponseFormat: jsonObjectFormat{Type: "json_object"},
		Thinking: struct {
			Type string `json:"type"`
		}{
			Type: "disabled",
		},
	})
	if err != nil {
		return vote.EquipmentDefinition{}, fmt.Errorf("encode llm request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return vote.EquipmentDefinition{}, fmt.Errorf("build llm request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+g.apiKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := g.client.Do(request)
	if err != nil {
		return vote.EquipmentDefinition{}, fmt.Errorf("call llm provider: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return vote.EquipmentDefinition{}, fmt.Errorf("read llm response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return vote.EquipmentDefinition{}, fmt.Errorf("llm provider returned status %d: %s", response.StatusCode, string(responseBody))
	}

	var payload chatCompletionResponse
	if err := sonic.Unmarshal(responseBody, &payload); err != nil {
		return vote.EquipmentDefinition{}, fmt.Errorf("%w: decode llm response", ErrInvalidEquipmentDraft)
	}
	if len(payload.Choices) == 0 {
		return vote.EquipmentDefinition{}, fmt.Errorf("%w: empty llm choices", ErrInvalidEquipmentDraft)
	}

	return parseEquipmentDraftJSON([]byte(payload.Choices[0].Message.Content))
}

// ============================================================================
// 请求/响应结构体
// ============================================================================

type chatCompletionRequest struct {
	Model          string           `json:"model"`
	Store          bool             `json:"store"`
	Messages       []chatMessage    `json:"messages"`
	ResponseFormat jsonObjectFormat `json:"response_format"`
	Thinking       struct {
		Type string `json:"type"`
	} `json:"thinking"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type jsonObjectFormat struct {
	Type string `json:"type"`
}

// ============================================================================
// 解析与校验
// ============================================================================

func parseEquipmentDraftJSON(raw []byte) (vote.EquipmentDefinition, error) {
	var fields map[string]json.RawMessage
	if err := sonic.Unmarshal(raw, &fields); err != nil {
		return vote.EquipmentDefinition{}, fmt.Errorf("%w: draft is not valid json", ErrInvalidEquipmentDraft)
	}

	requiredFields := map[string]struct{}{
		"itemId":               {},
		"name":                 {},
		"slot":                 {},
		"description":          {},
		"rarity":               {},
		"attackPower":          {},
		"armorPenPercent":      {},
		"critDamageMultiplier": {},
		"partTypeDamageSoft":   {},
		"partTypeDamageHeavy":  {},
		"partTypeDamageWeak":   {},
		"talentAffinity":       {},
	}

	numericFields := []string{
		"attackPower",
		"armorPenPercent",
		"critDamageMultiplier",
		"partTypeDamageSoft",
		"partTypeDamageHeavy",
		"partTypeDamageWeak",
	}
	for _, key := range numericFields {
		if _, ok := fields[key]; !ok {
			fields[key] = json.RawMessage("0")
		}
	}
	if _, ok := fields["talentAffinity"]; !ok {
		fields["talentAffinity"] = json.RawMessage(`""`)
	}

	raw, _ = sonic.Marshal(fields)

	for key := range fields {
		if _, ok := requiredFields[key]; !ok {
			return vote.EquipmentDefinition{}, fmt.Errorf("%w: unsupported field %s", ErrInvalidEquipmentDraft, key)
		}
	}
	for key := range requiredFields {
		if _, ok := fields[key]; !ok {
			return vote.EquipmentDefinition{}, fmt.Errorf("%w: missing field %s", ErrInvalidEquipmentDraft, key)
		}
	}

	var draft vote.EquipmentDefinition
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&draft); err != nil {
		return vote.EquipmentDefinition{}, fmt.Errorf("%w: decode draft", ErrInvalidEquipmentDraft)
	}

	fmt.Printf("draft parsed: %+v\n", draft)

	if err := validateEquipmentDraft(draft); err != nil {
		return vote.EquipmentDefinition{}, err
	}
	return draft, nil
}

// ============================================================================
// 稀有度 + 部位 数值范围表
// ============================================================================

type rarityBounds struct {
	attackMin      int64
	attackMax      int64
	critMax        float64
	armorPenMax    float64
	partTypeDmgMax float64
}

var rarityBaseBounds = map[string]rarityBounds{
	"普通": {attackMin: 5, attackMax: 30, critMax: 0, armorPenMax: 0, partTypeDmgMax: 0},
	"优秀": {attackMin: 20, attackMax: 60, critMax: 0.3, armorPenMax: 0.1, partTypeDmgMax: 0.05},
	"稀有": {attackMin: 50, attackMax: 120, critMax: 0.6, armorPenMax: 0.25, partTypeDmgMax: 0.1},
	"史诗": {attackMin: 100, attackMax: 250, critMax: 1.0, armorPenMax: 0.4, partTypeDmgMax: 0.2},
	"传说": {attackMin: 200, attackMax: 500, critMax: 1.5, armorPenMax: 0.6, partTypeDmgMax: 0.35},
	"至臻": {attackMin: 400, attackMax: 1000, critMax: 2.0, armorPenMax: 0.8, partTypeDmgMax: 0.5},
}

type slotModifiers struct {
	attackMinRatio float64
	attackMaxRatio float64
	critRatio      float64
	armorPenRatio  float64
	partDmgRatio   float64
}

var slotModifierMap = map[string]slotModifiers{
	"weapon":    {attackMinRatio: 1.0, attackMaxRatio: 1.0, critRatio: 1.0, armorPenRatio: 1.0, partDmgRatio: 1.0},
	"gloves":    {attackMinRatio: 0.55, attackMaxRatio: 0.70, critRatio: 1.0, armorPenRatio: 1.0, partDmgRatio: 1.0},
	"helmet":    {attackMinRatio: 0.30, attackMaxRatio: 0.45, critRatio: 0.3, armorPenRatio: 0.3, partDmgRatio: 0.4},
	"chest":     {attackMinRatio: 0.20, attackMaxRatio: 0.35, critRatio: 0.3, armorPenRatio: 0.3, partDmgRatio: 0.4},
	"legs":      {attackMinRatio: 0.30, attackMaxRatio: 0.45, critRatio: 0.3, armorPenRatio: 0.3, partDmgRatio: 0.4},
	"accessory": {attackMinRatio: 0.10, attackMaxRatio: 0.25, critRatio: 1.0, armorPenRatio: 1.0, partDmgRatio: 0.4},
}

func getSlotBounds(rarity string, slot string) (rarityBounds, error) {
	base, ok := rarityBaseBounds[rarity]
	if !ok {
		return rarityBounds{}, fmt.Errorf("unknown rarity: %s", rarity)
	}
	mod, ok := slotModifierMap[slot]
	if !ok {
		return rarityBounds{}, fmt.Errorf("unknown slot: %s", slot)
	}

	attackMin := int(float64(base.attackMin) * mod.attackMinRatio)
	attackMax := int(float64(base.attackMax) * mod.attackMaxRatio)
	if attackMax < attackMin {
		attackMax = attackMin
	}

	return rarityBounds{
		attackMin:      int64(attackMin),
		attackMax:      int64(attackMax),
		critMax:        base.critMax * mod.critRatio,
		armorPenMax:    base.armorPenMax * mod.armorPenRatio,
		partTypeDmgMax: base.partTypeDmgMax * mod.partDmgRatio,
	}, nil
}

// ============================================================================
// 校验函数
// ============================================================================

func validateEquipmentDraft(draft vote.EquipmentDefinition) error {
	if strings.TrimSpace(draft.ItemID) == "" || strings.TrimSpace(draft.Name) == "" {
		return fmt.Errorf("%w: itemId and name are required", ErrInvalidEquipmentDraft)
	}
	if !allowedString(draft.Slot, []string{"weapon", "helmet", "chest", "gloves", "legs", "accessory"}) {
		return fmt.Errorf("%w: invalid slot", ErrInvalidEquipmentDraft)
	}
	if !allowedString(draft.Rarity, []string{"普通", "优秀", "稀有", "史诗", "传说", "至臻"}) {
		return fmt.Errorf("%w: invalid rarity", ErrInvalidEquipmentDraft)
	}
	if !allowedString(draft.TalentAffinity, []string{"", "normal", "armor", "crit"}) {
		return fmt.Errorf("%w: invalid talentAffinity", ErrInvalidEquipmentDraft)
	}

	bounds, err := getSlotBounds(draft.Rarity, draft.Slot)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidEquipmentDraft, err)
	}

	if draft.AttackPower < bounds.attackMin || draft.AttackPower > bounds.attackMax {
		return fmt.Errorf("%w: attackPower %d out of range [%d, %d] for %s %s",
			ErrInvalidEquipmentDraft, draft.AttackPower, bounds.attackMin, bounds.attackMax, draft.Rarity, draft.Slot)
	}

	if draft.CritDamageMultiplier < 0 || draft.CritDamageMultiplier > bounds.critMax {
		return fmt.Errorf("%w: critDamageMultiplier %.2f out of range [0, %.2f] for %s %s",
			ErrInvalidEquipmentDraft, draft.CritDamageMultiplier, bounds.critMax, draft.Rarity, draft.Slot)
	}

	if draft.ArmorPenPercent < 0 || draft.ArmorPenPercent > bounds.armorPenMax {
		return fmt.Errorf("%w: armorPenPercent %.2f out of range [0, %.2f] for %s %s",
			ErrInvalidEquipmentDraft, draft.ArmorPenPercent, bounds.armorPenMax, draft.Rarity, draft.Slot)
	}

	partFields := map[string]float64{
		"partTypeDamageSoft":  draft.PartTypeDamageSoft,
		"partTypeDamageHeavy": draft.PartTypeDamageHeavy,
		"partTypeDamageWeak":  draft.PartTypeDamageWeak,
	}
	for field, value := range partFields {
		if value < 0 || value > bounds.partTypeDmgMax {
			return fmt.Errorf("%w: %s %.2f out of range [0, %.2f] for %s %s",
				ErrInvalidEquipmentDraft, field, value, bounds.partTypeDmgMax, draft.Rarity, draft.Slot)
		}
	}

	if draft.Rarity == "普通" {
		if draft.CritDamageMultiplier != 0 {
			return fmt.Errorf("%w: 普通装备 critDamageMultiplier 必须为 0", ErrInvalidEquipmentDraft)
		}
		if draft.ArmorPenPercent != 0 {
			return fmt.Errorf("%w: 普通装备 armorPenPercent 必须为 0", ErrInvalidEquipmentDraft)
		}
		if draft.PartTypeDamageSoft != 0 || draft.PartTypeDamageHeavy != 0 || draft.PartTypeDamageWeak != 0 {
			return fmt.Errorf("%w: 普通装备 partTypeDamage 三项必须为 0", ErrInvalidEquipmentDraft)
		}
	}

	// 通用装备：attackPower 应接近中位值
	if draft.TalentAffinity == "" {
		midAttack := bounds.attackMin + (bounds.attackMax-bounds.attackMin)/2
		lower := bounds.attackMin + (bounds.attackMax-bounds.attackMin)/4
		upper := bounds.attackMin + (bounds.attackMax-bounds.attackMin)*3/4
		if draft.AttackPower < lower || draft.AttackPower > upper {
			return fmt.Errorf("%w: 通用装备 attackPower %d 应接近 %s %s 中位值 %d，允许范围 [%d, %d]",
				ErrInvalidEquipmentDraft, draft.AttackPower, draft.Rarity, draft.Slot, midAttack, lower, upper)
		}
		if draft.CritDamageMultiplier > bounds.critMax*0.5 {
			return fmt.Errorf("%w: 通用装备 critDamageMultiplier 不应超过上限的 50%%", ErrInvalidEquipmentDraft)
		}
		if draft.ArmorPenPercent > bounds.armorPenMax*0.5 {
			return fmt.Errorf("%w: 通用装备 armorPenPercent 不应超过上限的 50%%", ErrInvalidEquipmentDraft)
		}
	}

	return nil
}

func allowedString(value string, allowed []string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

// ============================================================================
// System Prompt
// ============================================================================

func equipmentDraftSystemPrompt() string {
	return strings.Join([]string{
		"你是装备数值策划，只能输出符合 JSON Schema 的装备草稿。",
		"三主系：normal=均衡攻势，armor=碎盾攻坚，crit=致命洞察。",
		"部位类型只允许 weapon、helmet、chest、gloves、legs、accessory。",
		"稀有度只允许 普通、优秀、稀有、史诗、传说、至臻。",
		"armorPenPercent 不得超过 0.8；所有百分比用 0.2 这类小数表示。",
		"",
		"【稀有度数值规则 必须遵守】",
		"以下为基础范围，具体数值还需根据部位乘以修正系数：",
		"普通：attackPower 5~30，无 critDamageMultiplier，无 armorPenPercent，partTypeDamage 三项均为 0",
		"优秀：attackPower 20~60，critDamageMultiplier 0~0.3，armorPenPercent 0~0.1，partTypeDamage 三项 0~0.05",
		"稀有：attackPower 50~120，critDamageMultiplier 0~0.6，armorPenPercent 0~0.25，partTypeDamage 三项 0~0.1",
		"史诗：attackPower 100~250，critDamageMultiplier 0~1.0，armorPenPercent 0~0.4，partTypeDamage 三项 0~0.2",
		"传说：attackPower 200~500，critDamageMultiplier 0~1.5，armorPenPercent 0~0.6，partTypeDamage 三项 0~0.35",
		"至臻：attackPower 400~1000，critDamageMultiplier 0~2.0，armorPenPercent 0~0.8，partTypeDamage 三项 0~0.5",
		"",
		"【部位修正系数 必须遵守】",
		"不同部位对 attackPower 的修正：",
		"weapon：取基础范围的 100%（全额）",
		"gloves：取基础范围的 55~70%",
		"helmet：取基础范围的 30~45%",
		"chest：取基础范围的 20~35%",
		"legs：取基础范围的 30~45%",
		"accessory：取基础范围的 10~25%",
		"",
		"不同部位对 critDamageMultiplier 的修正：",
		"weapon、gloves、accessory：可达到稀有度上限的 100%",
		"helmet、chest、legs：不得超过稀有度上限的 30%",
		"",
		"不同部位对 armorPenPercent 的修正：",
		"weapon、gloves、accessory：可达到稀有度上限的 100%",
		"helmet、chest、legs：不得超过稀有度上限的 30%",
		"",
		"不同部位对 partTypeDamage 的修正：",
		"weapon、gloves：可达到稀有度上限的 100%",
		"helmet、chest、legs、accessory：不得超过稀有度上限的 40%",
		"",
		"talentAffinity 为空字符串表示通用装备，否则必须是 normal/armor/crit 之一。",
		"通用装备各项数值取该稀有度中位值附近（attackPower 取基础范围中间值，critDamageMultiplier 和 armorPenPercent 取基础范围中间值）。",
		"description 是用中文描述装备外观和特点的文本。",
		"返回必须是完整 JSON 对象，不要 Markdown，不要解释。",
		"输出的 JSON 要求：itemId 必须是字符串（如 \"wood_sword_001\"），其余字段为 name, slot, rarity, description, attackPower, armorPenPercent, critDamageMultiplier, partTypeDamageSoft, partTypeDamageHeavy, partTypeDamageWeak, talentAffinity。",
	}, "\n")
}
