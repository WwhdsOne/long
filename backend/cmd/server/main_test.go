package main

import (
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/common/ut"

	"long/internal/config"
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

func TestRenderStartupInfoIncludesBannerAndSummary(t *testing.T) {
	cfg := config.Config{
		Port:        2333,
		Redis:       config.RedisConfig{Host: "127.0.0.1", Port: 6379, DB: 2, TLSEnabled: true},
		Mongo:       config.MongoConfig{Enabled: true, Database: "vote_wall"},
		Log:         config.LogConfig{Level: "info", Format: "json"},
		RedisPrefix: "vote:",
	}

	info := renderStartupInfo(cfg, "127.0.0.1:2333")
	assertContains(t, godAnimalArt(), "神兽保佑")
	assertContains(t, godAnimalArt(), "代码无BUG")
	assertContains(t, info, "监听地址: 127.0.0.1:2333")
	assertContains(t, info, "Redis: 127.0.0.1:6379/2")
	assertContains(t, info, "Redis TLS: true")
	assertContains(t, info, "Mongo: true (vote_wall)")
	assertContains(t, info, "日志: level=info format=json")
}

func assertContains(t *testing.T, got string, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("expected %q to contain %q", got, want)
	}
}
