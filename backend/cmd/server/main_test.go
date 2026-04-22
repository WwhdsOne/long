package main

import (
	"testing"

	"github.com/cloudwego/hertz/pkg/common/ut"

	"long/internal/httpapi"
)

func TestHertzServerRegistersPprofAndAPI(t *testing.T) {
	server := httpapi.NewHertzServer(":2333", httpapi.Options{})

	pprofResponse := ut.PerformRequest(server.Engine, "GET", "/debug/pprof/", nil).Result()
	if pprofResponse.StatusCode() != 200 {
		t.Fatalf("expected 200 from pprof index, got %d", pprofResponse.StatusCode())
	}
	if body := string(pprofResponse.Body()); body == "" {
		t.Fatal("expected pprof body")
	}
}

func TestServerAddressUsesLoopback(t *testing.T) {
	if addr := serverAddress(2333); addr != "127.0.0.1:2333" {
		t.Fatalf("expected loopback listen address, got %q", addr)
	}
}

func TestServerAddressUsesOverrideEnv(t *testing.T) {
	t.Setenv("LONG_LISTEN_HOST", "0.0.0.0")
	t.Setenv("LONG_LISTEN_PORT", "18080")

	if addr := serverAddress(2333); addr != "0.0.0.0:18080" {
		t.Fatalf("expected overridden listen address, got %q", addr)
	}
}
