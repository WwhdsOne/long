package core

import (
	"strings"
	"testing"

	"github.com/bytedance/sonic"
)

func TestBossJSONMarshalsHPFieldsAsStrings(t *testing.T) {
	payload, err := sonic.Marshal(Boss{
		ID:        "boss-1",
		Name:      "巨像",
		MaxHP:     9223372036854775800,
		CurrentHP: 9223372036854775799,
		Parts: []BossPart{
			{
				X:         0,
				Y:         0,
				Type:      PartTypeSoft,
				MaxHP:     9223372036854775800,
				CurrentHP: 9223372036854775799,
				Armor:     12,
				Alive:     true,
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal boss: %v", err)
	}

	text := string(payload)
	if !strings.Contains(text, `"maxHp":"9223372036854775800"`) {
		t.Fatalf("expected boss maxHp string, got %s", text)
	}
	if !strings.Contains(text, `"currentHp":"9223372036854775799"`) {
		t.Fatalf("expected boss currentHp string, got %s", text)
	}
	if !strings.Contains(text, `"armor":"12"`) {
		t.Fatalf("expected boss part armor string, got %s", text)
	}
}

func TestRoomInfoJSONMarshalsBossHPFieldsAsStrings(t *testing.T) {
	payload, err := sonic.Marshal(RoomInfo{
		ID:               "2",
		CurrentBossHP:    9223372036854775799,
		CurrentBossMaxHP: 9223372036854775800,
		CurrentBossAvgHP: 1844674407370955160,
	})
	if err != nil {
		t.Fatalf("marshal room info: %v", err)
	}

	text := string(payload)
	if !strings.Contains(text, `"currentBossHp":"9223372036854775799"`) {
		t.Fatalf("expected currentBossHp string, got %s", text)
	}
	if !strings.Contains(text, `"currentBossMaxHp":"9223372036854775800"`) {
		t.Fatalf("expected currentBossMaxHp string, got %s", text)
	}
	if !strings.Contains(text, `"currentBossAvgHp":"1844674407370955160"`) {
		t.Fatalf("expected currentBossAvgHp string, got %s", text)
	}
}
