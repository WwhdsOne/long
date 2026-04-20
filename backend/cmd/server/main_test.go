package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildRootHandlerRegistersPprof(t *testing.T) {
	appHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("app"))
	})

	handler := buildRootHandler(appHandler)

	pprofRequest := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	pprofResponse := httptest.NewRecorder()
	handler.ServeHTTP(pprofResponse, pprofRequest)

	if pprofResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 from pprof index, got %d", pprofResponse.Code)
	}
	if body := pprofResponse.Body.String(); !strings.Contains(body, "profile") {
		t.Fatalf("expected pprof body, got %q", body)
	}

	appRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	appResponse := httptest.NewRecorder()
	handler.ServeHTTP(appResponse, appRequest)

	if appResponse.Code != http.StatusTeapot {
		t.Fatalf("expected app handler status, got %d", appResponse.Code)
	}
}
