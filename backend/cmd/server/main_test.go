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
