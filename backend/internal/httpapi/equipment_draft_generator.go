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

	body, err := json.Marshal(chatCompletionRequest{
		Model: g.model,
		Store: false,
		Messages: []chatMessage{
			{Role: "system", Content: equipmentDraftSystemPrompt()},
			{Role: "user", Content: strings.TrimSpace(prompt)},
		},
		ResponseFormat: equipmentDraftResponseFormat(),
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
		return vote.EquipmentDefinition{}, fmt.Errorf("llm provider returned status %d", response.StatusCode)
	}

	var payload chatCompletionResponse
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return vote.EquipmentDefinition{}, fmt.Errorf("%w: decode llm response", ErrInvalidEquipmentDraft)
	}
	if len(payload.Choices) == 0 {
		return vote.EquipmentDefinition{}, fmt.Errorf("%w: empty llm choices", ErrInvalidEquipmentDraft)
	}

	return parseEquipmentDraftJSON([]byte(payload.Choices[0].Message.Content))
}

type chatCompletionRequest struct {
	Model          string                 `json:"model"`
	Store          bool                   `json:"store"`
	Messages       []chatMessage          `json:"messages"`
	ResponseFormat map[string]interface{} `json:"response_format"`
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

func parseEquipmentDraftJSON(raw []byte) (vote.EquipmentDefinition, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return vote.EquipmentDefinition{}, fmt.Errorf("%w: draft is not valid json", ErrInvalidEquipmentDraft)
	}

	required := map[string]struct{}{
		"itemId":               {},
		"name":                 {},
		"slot":                 {},
		"rarity":               {},
		"imagePath":            {},
		"imageAlt":             {},
		"attackPower":          {},
		"armorPenPercent":      {},
		"critDamageMultiplier": {},
		"bossDamagePercent":    {},
		"partTypeDamageSoft":   {},
		"partTypeDamageHeavy":  {},
		"partTypeDamageWeak":   {},
		"talentAffinity":       {},
	}
	for key := range fields {
		if _, ok := required[key]; !ok {
			return vote.EquipmentDefinition{}, fmt.Errorf("%w: unsupported field %s", ErrInvalidEquipmentDraft, key)
		}
	}
	for key := range required {
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
	if err := validateEquipmentDraft(draft); err != nil {
		return vote.EquipmentDefinition{}, err
	}
	return draft, nil
}

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
	if draft.AttackPower < 0 ||
		draft.ArmorPenPercent < 0 || draft.ArmorPenPercent > 0.8 ||
		draft.CritDamageMultiplier < 0 ||
		draft.BossDamagePercent < 0 ||
		draft.PartTypeDamageSoft < 0 ||
		draft.PartTypeDamageHeavy < 0 ||
		draft.PartTypeDamageWeak < 0 {
		return fmt.Errorf("%w: numeric value out of range", ErrInvalidEquipmentDraft)
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

func equipmentDraftSystemPrompt() string {
	return strings.Join([]string{
		"你是装备数值策划，只能输出符合 JSON Schema 的装备草稿。",
		"三主系：normal=均衡攻势，armor=碎盾攻坚，crit=致命洞察。",
		"部位类型只允许 weapon、helmet、chest、gloves、legs、accessory。",
		"稀有度只允许 普通、优秀、稀有、史诗、传说、至臻。",
		"禁止生成攻速、点击间隔、额外点击概率、bonusClicks、bonusCriticalChancePercent、bonusCriticalCount、enhanceCap 等旧词缀或字段。",
		"armorPenPercent 不得超过 0.8；所有百分比用 0.2 这类小数表示。",
		"返回必须是完整 JSON 对象，不要 Markdown，不要解释。",
	}, "\n")
}

func equipmentDraftResponseFormat() map[string]interface{} {
	properties := map[string]interface{}{
		"itemId":               map[string]interface{}{"type": "string"},
		"name":                 map[string]interface{}{"type": "string"},
		"slot":                 map[string]interface{}{"type": "string", "enum": []string{"weapon", "helmet", "chest", "gloves", "legs", "accessory"}},
		"rarity":               map[string]interface{}{"type": "string", "enum": []string{"普通", "优秀", "稀有", "史诗", "传说", "至臻"}},
		"imagePath":            map[string]interface{}{"type": "string"},
		"imageAlt":             map[string]interface{}{"type": "string"},
		"attackPower":          map[string]interface{}{"type": "integer", "minimum": 0},
		"armorPenPercent":      map[string]interface{}{"type": "number", "minimum": 0, "maximum": 0.8},
		"critDamageMultiplier": map[string]interface{}{"type": "number", "minimum": 0},
		"bossDamagePercent":    map[string]interface{}{"type": "number", "minimum": 0},
		"partTypeDamageSoft":   map[string]interface{}{"type": "number", "minimum": 0},
		"partTypeDamageHeavy":  map[string]interface{}{"type": "number", "minimum": 0},
		"partTypeDamageWeak":   map[string]interface{}{"type": "number", "minimum": 0},
		"talentAffinity":       map[string]interface{}{"type": "string", "enum": []string{"", "normal", "armor", "crit"}},
	}
	required := []string{
		"itemId", "name", "slot", "rarity", "imagePath", "imageAlt",
		"attackPower", "armorPenPercent", "critDamageMultiplier", "bossDamagePercent",
		"partTypeDamageSoft", "partTypeDamageHeavy", "partTypeDamageWeak", "talentAffinity",
	}

	return map[string]interface{}{
		"type": "json_schema",
		"json_schema": map[string]interface{}{
			"name":   "equipment_draft",
			"strict": true,
			"schema": map[string]interface{}{
				"type":                 "object",
				"additionalProperties": false,
				"properties":           properties,
				"required":             required,
			},
		},
	}
}
