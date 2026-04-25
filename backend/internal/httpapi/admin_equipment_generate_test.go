package httpapi

import (
	"context"
	"errors"
	"testing"

	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"long/internal/admin"
	"long/internal/vote"
)

type fakeEquipmentDraftGenerator struct {
	draft      vote.EquipmentDefinition
	err        error
	lastPrompt string
}

func (f *fakeEquipmentDraftGenerator) GenerateEquipmentDraft(_ context.Context, prompt string) (vote.EquipmentDefinition, error) {
	f.lastPrompt = prompt
	if f.err != nil {
		return vote.EquipmentDefinition{}, f.err
	}
	return f.draft, nil
}

func TestAdminEquipmentGenerateRequiresLogin(t *testing.T) {
	handler := newNativeTestHandler(Options{
		Store: &mockStore{},
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		}),
		EquipmentDraftGenerator: &fakeEquipmentDraftGenerator{},
	})

	response := performJSONRequest(t, handler, consts.MethodPost, "/api/admin/equipment/generate", `{"prompt":"做一把武器"}`)
	if response.StatusCode() != consts.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d", response.StatusCode())
	}
}

func TestAdminEquipmentGenerateReturnsDraftWithoutSaving(t *testing.T) {
	store := &mockStore{}
	generator := &fakeEquipmentDraftGenerator{
		draft: vote.EquipmentDefinition{
			ItemID:               "soft-blade",
			Name:                 "软组织切割刃",
			Slot:                 "weapon",
			Rarity:               "史诗",
			ImagePath:            "/images/equipment/soft-blade.png",
			ImageAlt:             "软组织切割刃",
			AttackPower:          12,
			ArmorPenPercent:      0.2,
			CritDamageMultiplier: 1.5,
			PartTypeDamageSoft:   0.35,
			TalentAffinity:       "normal",
		},
	}
	handler := newNativeTestHandler(Options{
		Store: store,
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		}),
		EquipmentDraftGenerator: generator,
	})
	sessionCookie := loginAdminForEquipmentGenerate(t, handler)

	response := performJSONRequest(
		t,
		handler,
		consts.MethodPost,
		"/api/admin/equipment/generate",
		`{"prompt":"做一把偏普攻流的软组织武器"}`,
		ut.Header{Key: "Cookie", Value: sessionCookie},
	)
	if response.StatusCode() != consts.StatusOK {
		t.Fatalf("expected generate 200, got %d body=%s", response.StatusCode(), response.Body())
	}
	if generator.lastPrompt != "做一把偏普攻流的软组织武器" {
		t.Fatalf("expected prompt passed to generator, got %q", generator.lastPrompt)
	}
	if store.lastEquipment.ItemID != "" {
		t.Fatalf("expected generate not to save equipment, got %+v", store.lastEquipment)
	}

	payload := decodeJSONResponse[struct {
		Draft vote.EquipmentDefinition `json:"draft"`
	}](t, response)
	if payload.Draft.ItemID != "soft-blade" || payload.Draft.TalentAffinity != "normal" {
		t.Fatalf("unexpected draft payload: %+v", payload.Draft)
	}
}

func TestAdminEquipmentGenerateRejectsInvalidDraft(t *testing.T) {
	generator := &fakeEquipmentDraftGenerator{
		err: ErrInvalidEquipmentDraft,
	}
	handler := newNativeTestHandler(Options{
		Store: &mockStore{},
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		}),
		EquipmentDraftGenerator: generator,
	})
	sessionCookie := loginAdminForEquipmentGenerate(t, handler)

	response := performJSONRequest(
		t,
		handler,
		consts.MethodPost,
		"/api/admin/equipment/generate",
		`{"prompt":"给我攻速很快的武器"}`,
		ut.Header{Key: "Cookie", Value: sessionCookie},
	)
	if response.StatusCode() != consts.StatusUnprocessableEntity {
		t.Fatalf("expected generate 422, got %d body=%s", response.StatusCode(), response.Body())
	}
}

func TestAdminEquipmentGenerateHandlesProviderFailure(t *testing.T) {
	generator := &fakeEquipmentDraftGenerator{
		err: errors.New("provider failed"),
	}
	handler := newNativeTestHandler(Options{
		Store: &mockStore{},
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		}),
		EquipmentDraftGenerator: generator,
	})
	sessionCookie := loginAdminForEquipmentGenerate(t, handler)

	response := performJSONRequest(
		t,
		handler,
		consts.MethodPost,
		"/api/admin/equipment/generate",
		`{"prompt":"做一把武器"}`,
		ut.Header{Key: "Cookie", Value: sessionCookie},
	)
	if response.StatusCode() != consts.StatusBadGateway {
		t.Fatalf("expected generate 502, got %d body=%s", response.StatusCode(), response.Body())
	}
}

func loginAdminForEquipmentGenerate(t *testing.T, handler *nativeTestHandler) string {
	t.Helper()

	loginResponse := performJSONRequest(t, handler, consts.MethodPost, "/api/admin/login", `{"username":"admin","password":"secret"}`)
	if loginResponse.StatusCode() != consts.StatusOK {
		t.Fatalf("expected login 200, got %d", loginResponse.StatusCode())
	}
	sessionCookie := string(loginResponse.Header.Peek("Set-Cookie"))
	if sessionCookie == "" {
		t.Fatal("expected session cookie from admin login")
	}
	return sessionCookie
}
