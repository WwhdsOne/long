package httpapi

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestOpenAIEquipmentDraftGeneratorCallsChatCompletions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer sk-test" {
			t.Fatalf("unexpected authorization header: %q", r.Header.Get("Authorization"))
		}

		rawBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		body := string(rawBody)
		for _, expected := range []string{
			`"model":"gpt-test"`,
			`"store":false`,
			`"response_format"`,
			`自然语言描述`,
		} {
			if !strings.Contains(body, expected) {
				t.Fatalf("expected request body to contain %q, got %s", expected, body)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [{
				"message": {
					"content": "{\"itemId\":\"soft-blade\",\"name\":\"软组织切割刃\",\"slot\":\"weapon\",\"rarity\":\"史诗\",\"description\":\"测试描述\",\"attackPower\":150,\"armorPenPercent\":0.4,\"critDamageMultiplier\":1.0,\"partTypeDamageSoft\":0.2,\"partTypeDamageHeavy\":0.2,\"partTypeDamageWeak\":0.2,\"talentAffinity\":\"normal\"}"
				}
			}]
		}`))
	}))
	defer server.Close()

	generator := NewOpenAIEquipmentDraftGenerator(EquipmentDraftGeneratorConfig{
		APIKey:  "sk-test",
		BaseURL: server.URL + "/v1",
		Model:   "gpt-test",
		Timeout: time.Second,
	})

	draft, err := generator.GenerateEquipmentDraft(context.Background(), "自然语言描述")
	if err != nil {
		t.Fatalf("generate equipment draft: %v", err)
	}
	if draft.ItemID != "soft-blade" || draft.Slot != "weapon" || draft.TalentAffinity != "normal" {
		t.Fatalf("unexpected draft: %+v", draft)
	}
}

func TestOpenAIEquipmentDraftGeneratorReturnsRawResponseOnInvalidDraft(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [{
				"message": {
					"content": "{\"itemId\":\"bad-blade\",\"name\":\"坏刀\",\"slot\":\"weapon\",\"rarity\":\"史诗\",\"description\":\"测试描述\",\"attackPower\":150,\"armorPenPercent\":0.95,\"critDamageMultiplier\":1.0,\"partTypeDamageSoft\":0.2,\"partTypeDamageHeavy\":0.2,\"partTypeDamageWeak\":0.2,\"talentAffinity\":\"normal\"}"
				}
			}]
		}`))
	}))
	defer server.Close()

	generator := NewOpenAIEquipmentDraftGenerator(EquipmentDraftGeneratorConfig{
		APIKey:  "sk-test",
		BaseURL: server.URL + "/v1",
		Model:   "gpt-test",
		Timeout: time.Second,
	})

	_, err := generator.GenerateEquipmentDraft(context.Background(), "自然语言描述")
	if err == nil {
		t.Fatal("expected invalid draft error")
	}
	var generateErr *EquipmentDraftGenerateError
	if !errors.As(err, &generateErr) {
		t.Fatalf("expected equipment draft generate error, got %T %v", err, err)
	}
	if generateErr.RawResponse == "" {
		t.Fatalf("expected raw response in error, got %+v", generateErr)
	}
	if generateErr.Draft.ItemID != "bad-blade" {
		t.Fatalf("expected partial draft in error, got %+v", generateErr.Draft)
	}
}

func TestOpenAIEquipmentDraftGeneratorRejectsForbiddenField(t *testing.T) {
	_, err := parseEquipmentDraftJSON([]byte(`{
		"itemId":"speed-blade",
		"name":"攻速刀",
		"slot":"weapon",
		"rarity":"稀有",
		"imagePath":"",
		"imageAlt":"",
		"attackPower":8,
		"armorPenPercent":0.1,
		"critDamageMultiplier":1.2,
		"bossDamagePercent":0.05,
		"partTypeDamageSoft":0,
		"partTypeDamageHeavy":0,
		"partTypeDamageWeak":0,
		"talentAffinity":"armor",
		"attackSpeed":1.5
	}`))
	if err == nil {
		t.Fatal("expected forbidden extra field to be rejected")
	}
}

func TestOpenAIEquipmentDraftGeneratorRejectsArmorPenOverflow(t *testing.T) {
	_, err := parseEquipmentDraftJSON([]byte(`{
		"itemId":"piercer",
		"name":"穿甲锥",
		"slot":"weapon",
		"rarity":"传说",
		"imagePath":"",
		"imageAlt":"",
		"attackPower":10,
		"armorPenPercent":0.81,
		"critDamageMultiplier":1.2,
		"bossDamagePercent":0.05,
		"partTypeDamageSoft":0,
		"partTypeDamageHeavy":0,
		"partTypeDamageWeak":0,
		"talentAffinity":"armor"
	}`))
	if err == nil {
		t.Fatal("expected armorPenPercent overflow to be rejected")
	}
}
