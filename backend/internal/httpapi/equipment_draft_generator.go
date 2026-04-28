package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"long/internal/vote"

	"github.com/bytedance/sonic"
)

var ErrInvalidEquipmentDraft = errors.New("invalid equipment draft")

// ============================================================================
// 配置与生成器
// ============================================================================

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

// ============================================================================
// 核心流程：先定稀有度，再生成装备
// ============================================================================

func (g *OpenAIEquipmentDraftGenerator) GenerateEquipmentDraft(ctx context.Context, prompt string) (vote.EquipmentDefinition, error) {
	if strings.TrimSpace(prompt) == "" {
		return vote.EquipmentDefinition{}, fmt.Errorf("%w: prompt is required", ErrInvalidEquipmentDraft)
	}

	// 第一步：由 AI 决定稀有度
	rarity, err := g.determineRarity(ctx, prompt)
	if err != nil {
		return vote.EquipmentDefinition{}, fmt.Errorf("determine rarity: %w", err)
	}

	// 第二步：根据稀有度拼接详细规则，生成装备草稿
	systemPrompt := equipmentDraftSystemPromptForRarity(rarity)
	body, err := sonic.Marshal(chatCompletionRequest{
		Model: g.model,
		Store: false,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
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

	content, err := g.chatOnce(ctx, body)
	if err != nil {
		return vote.EquipmentDefinition{}, err
	}
	return parseEquipmentDraftJSON([]byte(content))
}

// ---------------------------------------------------------------------------
// 第一阶段：判定稀有度
// ---------------------------------------------------------------------------

type rarityResponse struct {
	Rarity string `json:"rarity"`
}

func (g *OpenAIEquipmentDraftGenerator) determineRarity(ctx context.Context, userPrompt string) (string, error) {
	sysPrompt := strings.Join([]string{
		"根据用户输入，确定装备稀有度。",
		"只允许返回 普通、优秀、稀有、史诗、传说、至臻 之一。",
		"如果用户没有明确指定，可以从列表中随机选择一个。",
		"直接返回 JSON，不要多余内容。",
	}, "\n")

	body, err := sonic.Marshal(chatCompletionRequest{
		Model: g.model,
		Store: false,
		Messages: []chatMessage{
			{Role: "system", Content: sysPrompt},
			{Role: "user", Content: userPrompt},
		},
		ResponseFormat: jsonObjectFormat{Type: "json_object"},
		Thinking: struct {
			Type string `json:"type"`
		}{
			Type: "disabled",
		},
	})
	if err != nil {
		return "", err
	}

	content, err := g.chatOnce(ctx, body)
	if err != nil {
		return "", err
	}

	var rr rarityResponse
	if err := sonic.Unmarshal([]byte(content), &rr); err != nil {
		return "", fmt.Errorf("%w: decode rarity response", ErrInvalidEquipmentDraft)
	}
	rr.Rarity = strings.TrimSpace(rr.Rarity)
	if !allowedString(rr.Rarity, []string{"普通", "优秀", "稀有", "史诗", "传说", "至臻"}) {
		return "", fmt.Errorf("%w: invalid rarity %q", ErrInvalidEquipmentDraft, rr.Rarity)
	}
	return rr.Rarity, nil
}

// ---------------------------------------------------------------------------
// 通用 HTTP 调用封装
// ---------------------------------------------------------------------------

func (g *OpenAIEquipmentDraftGenerator) chatOnce(ctx context.Context, body []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build llm request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call llm provider: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read llm response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("llm provider returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var payload chatCompletionResponse
	if err := sonic.Unmarshal(respBody, &payload); err != nil {
		return "", fmt.Errorf("%w: decode llm response", ErrInvalidEquipmentDraft)
	}
	if len(payload.Choices) == 0 {
		return "", fmt.Errorf("%w: empty llm choices", ErrInvalidEquipmentDraft)
	}
	return payload.Choices[0].Message.Content, nil
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
		"critRate":             {},
	}

	numericFields := []string{
		"attackPower",
		"armorPenPercent",
		"critDamageMultiplier",
		"partTypeDamageSoft",
		"partTypeDamageHeavy",
		"partTypeDamageWeak",
		"critRate",
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
	critRateMax    float64
	critMax        float64
	armorPenMax    float64
	partTypeDmgMax float64
}

var rarityBaseBounds = map[string]rarityBounds{
	"普通": {attackMin: 5, attackMax: 30, critRateMax: 0, critMax: 0, armorPenMax: 0, partTypeDmgMax: 0},
	"优秀": {attackMin: 20, attackMax: 60, critRateMax: 0.05, critMax: 0.3, armorPenMax: 0.1, partTypeDmgMax: 0.05},
	"稀有": {attackMin: 50, attackMax: 120, critRateMax: 0.08, critMax: 0.6, armorPenMax: 0.25, partTypeDmgMax: 0.1},
	"史诗": {attackMin: 100, attackMax: 250, critRateMax: 0.12, critMax: 1.0, armorPenMax: 0.4, partTypeDmgMax: 0.2},
	"传说": {attackMin: 200, attackMax: 500, critRateMax: 0.18, critMax: 1.5, armorPenMax: 0.6, partTypeDmgMax: 0.35},
	"至臻": {attackMin: 400, attackMax: 1000, critRateMax: 0.35, critMax: 2.0, armorPenMax: 0.8, partTypeDmgMax: 0.5},
}

type slotModifiers struct {
	attackMinRatio float64
	attackMaxRatio float64
	critRateRatio  float64
	critRatio      float64
	armorPenRatio  float64
	partDmgRatio   float64
}

var slotModifierMap = map[string]slotModifiers{
	"weapon":    {attackMinRatio: 1.0, attackMaxRatio: 1.0, critRateRatio: 1.0, critRatio: 1.0, armorPenRatio: 1.0, partDmgRatio: 1.0},
	"gloves":    {attackMinRatio: 0.55, attackMaxRatio: 0.70, critRateRatio: 1.0, critRatio: 1.0, armorPenRatio: 1.0, partDmgRatio: 1.0},
	"helmet":    {attackMinRatio: 0.30, attackMaxRatio: 0.45, critRateRatio: 0.3, critRatio: 0.3, armorPenRatio: 0.3, partDmgRatio: 0.4},
	"chest":     {attackMinRatio: 0.20, attackMaxRatio: 0.35, critRateRatio: 0.3, critRatio: 0.3, armorPenRatio: 0.3, partDmgRatio: 0.4},
	"legs":      {attackMinRatio: 0.30, attackMaxRatio: 0.45, critRateRatio: 0.3, critRatio: 0.3, armorPenRatio: 0.3, partDmgRatio: 0.4},
	"accessory": {attackMinRatio: 0.10, attackMaxRatio: 0.25, critRateRatio: 1.0, critRatio: 1.0, armorPenRatio: 1.0, partDmgRatio: 0.4},
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
	attackMax := max(int(float64(base.attackMax)*mod.attackMaxRatio), attackMin)

	return rarityBounds{
		attackMin:      int64(attackMin),
		attackMax:      int64(attackMax),
		critRateMax:    base.critRateMax * mod.critRateRatio,
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

	if draft.CritRate < 0 || draft.CritRate > bounds.critRateMax {
		return fmt.Errorf("%w: critRate %.2f out of range [0, %.2f] for %s %s",
			ErrInvalidEquipmentDraft, draft.CritRate, bounds.critRateMax, draft.Rarity, draft.Slot)
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
		if draft.CritRate != 0 {
			return fmt.Errorf("%w: 普通装备 critRate 必须为 0", ErrInvalidEquipmentDraft)
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
		if draft.CritRate > bounds.critRateMax*0.5 {
			return fmt.Errorf("%w: 通用装备 critRate 不应超过上限的 50%%", ErrInvalidEquipmentDraft)
		}
	}

	return nil
}

func allowedString(value string, allowed []string) bool {
	return slices.Contains(allowed, value)
}

// ============================================================================
// 按稀有度预计算的详细 prompt（AI 无需再做乘法）
// ============================================================================

func equipmentDraftSystemPromptForRarity(rarity string) string {
	type slotRange struct {
		attackMin, attackMax       int
		critRateMax                float64
		critMax                    float64
		armorPenMax                float64
		softMax, heavyMax, weakMax float64
	}
	compute := func(r string) map[string]slotRange {
		m := make(map[string]slotRange)
		for _, slot := range []string{"weapon", "gloves", "helmet", "chest", "legs", "accessory"} {
			b, _ := getSlotBounds(r, slot)
			m[slot] = slotRange{
				attackMin:   int(b.attackMin),
				attackMax:   int(b.attackMax),
				critRateMax: b.critRateMax,
				critMax:     b.critMax,
				armorPenMax: b.armorPenMax,
				softMax:     b.partTypeDmgMax,
				heavyMax:    b.partTypeDmgMax,
				weakMax:     b.partTypeDmgMax,
			}
		}
		return m
	}

	ranges := compute(rarity)

	slotLines := make([]string, 0, 6)
	for _, slot := range []string{"weapon", "gloves", "helmet", "chest", "legs", "accessory"} {
		r := ranges[slot]
		lower := r.attackMin + (r.attackMax-r.attackMin)/4
		upper := r.attackMin + (r.attackMax-r.attackMin)*3/4

		line := fmt.Sprintf(
			"%s %s：attackPower %d~%d，critRate 0~%.2f，critDamageMultiplier 0~%.2f，armorPenPercent 0~%.2f，partTypeDamage 三项各 0~%.2f",
			rarity, slot, r.attackMin, r.attackMax, r.critRateMax, r.critMax, r.armorPenMax, r.softMax,
		)
		line += fmt.Sprintf(
			"；若 talentAffinity=\"\"（通用装备），则 attackPower 限制在 %d~%d，critRate不超过%.2f，critDamageMultiplier不超过%.2f，armorPenPercent不超过%.2f",
			lower, upper, r.critRateMax*0.5, r.critMax*0.5, r.armorPenMax*0.5,
		)
		slotLines = append(slotLines, line)
	}

	prompt := strings.Join([]string{
		"你是装备数值策划，只能输出符合 JSON Schema 的装备草稿。",
		"三主系：normal=均衡攻势，armor=碎盾攻坚，crit=致命洞察。",
		"部位类型只允许 weapon、helmet、chest、gloves、legs、accessory。",
		fmt.Sprintf("本次只生成 %s 稀有度的装备，最终数值必须严格遵守下方给出的范围。", rarity),
		"所有百分比用 0.2 这类小数表示。",
		"",
		"【所有部位最终数值范围（已含修正，直接使用，禁止自行乘法）】",
		strings.Join(slotLines, "\n"),
		"",
		"talentAffinity 为空字符串表示通用装备，否则必须是 normal/armor/crit 之一。",
		"description 是用中文描述装备外观和特点的文本。",
		"返回必须是完整 JSON 对象，不要 Markdown，不要解释。",
		"输出的 JSON 要求：itemId 必须是字符串，其余字段为 name, slot, rarity, description, attackPower, armorPenPercent, critRate, critDamageMultiplier, partTypeDamageSoft, partTypeDamageHeavy, partTypeDamageWeak, talentAffinity。",
	}, "\n")

	return prompt
}
