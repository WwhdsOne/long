package httpapi

import (
	"context"
	"errors"
	"testing"

	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"long/internal/admin"
	"long/internal/core"
)

type fakeEquipmentDraftGenerator struct {
	draft        core.EquipmentDefinition
	err          error
	lastPrompt   string
	rawResponse  string
	returnDetail bool
}

func (f *fakeEquipmentDraftGenerator) GenerateEquipmentDraft(_ context.Context, prompt string) (core.EquipmentDefinition, error) {
	f.lastPrompt = prompt
	if f.err != nil {
		if f.returnDetail {
			return core.EquipmentDefinition{}, &EquipmentDraftGenerateError{
				Message:     f.err.Error(),
				Prompt:      prompt,
				Draft:       f.draft,
				RawResponse: f.rawResponse,
				Cause:       f.err,
			}
		}
		return core.EquipmentDefinition{}, f.err
	}
	return f.draft, nil
}

type fakeEquipmentDraftFailureWriter struct {
	items []core.EquipmentDraftFailureLog
	err   error
}

func (f *fakeEquipmentDraftFailureWriter) WriteEquipmentDraftFailure(ctx context.Context, item core.EquipmentDraftFailureLog) error {
	f.items = append(f.items, item)
	return f.err
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
		draft: core.EquipmentDefinition{
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
		Draft core.EquipmentDefinition `json:"draft"`
	}](t, response)
	if payload.Draft.ItemID != "soft-blade" || payload.Draft.TalentAffinity != "normal" {
		t.Fatalf("unexpected draft payload: %+v", payload.Draft)
	}
}

func TestAdminEquipmentGenerateRejectsInvalidDraft(t *testing.T) {
	writer := &fakeEquipmentDraftFailureWriter{}
	generator := &fakeEquipmentDraftGenerator{
		draft: core.EquipmentDefinition{
			ItemID: "bad-weapon",
			Name:   "坏掉的草稿",
			Slot:   "weapon",
			Rarity: "史诗",
		},
		err:          ErrInvalidEquipmentDraft,
		rawResponse:  `{"itemId":"bad-weapon"}`,
		returnDetail: true,
	}
	handler := newNativeTestHandler(Options{
		Store: &mockStore{},
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		}),
		EquipmentDraftGenerator:     generator,
		EquipmentDraftFailureWriter: writer,
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
	if len(writer.items) != 1 {
		t.Fatalf("expected one failure log, got %d", len(writer.items))
	}
	if writer.items[0].Prompt != "给我攻速很快的武器" {
		t.Fatalf("expected prompt to be logged, got %+v", writer.items[0])
	}
	if writer.items[0].Draft.ItemID != "bad-weapon" {
		t.Fatalf("expected draft to be logged, got %+v", writer.items[0].Draft)
	}
	if writer.items[0].RawResponse != `{"itemId":"bad-weapon"}` {
		t.Fatalf("expected raw response logged, got %+v", writer.items[0])
	}
}

func TestAdminEquipmentGenerateHandlesProviderFailure(t *testing.T) {
	writer := &fakeEquipmentDraftFailureWriter{}
	generator := &fakeEquipmentDraftGenerator{
		err:          errors.New("provider failed"),
		rawResponse:  `provider timeout`,
		returnDetail: true,
	}
	handler := newNativeTestHandler(Options{
		Store: &mockStore{},
		AdminAuthenticator: admin.NewAuthenticator(admin.Config{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		}),
		EquipmentDraftGenerator:     generator,
		EquipmentDraftFailureWriter: writer,
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
	if len(writer.items) != 1 {
		t.Fatalf("expected one failure log, got %d", len(writer.items))
	}
	if writer.items[0].ErrorMessage != "provider failed" {
		t.Fatalf("expected provider error logged, got %+v", writer.items[0])
	}
	if writer.items[0].RawResponse != "provider timeout" {
		t.Fatalf("expected provider raw response logged, got %+v", writer.items[0])
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
