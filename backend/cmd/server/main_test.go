package main

import (
	"crypto/tls"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/common/ut"

	"long/internal/config"
	"long/internal/httpapi"
)

func TestHertzServerRegistersPprofAndAPI(t *testing.T) {
	server := httpapi.NewHertzServer(":2333", httpapi.Options{}, nil)

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

func TestResolveServerTLSConfigDisabledByDefault(t *testing.T) {
	tlsConfig, err := resolveServerTLSConfig()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if tlsConfig != nil {
		t.Fatalf("expected nil tls config, got %+v", tlsConfig)
	}
}

func TestResolveServerTLSConfigRequiresCertAndKeyTogether(t *testing.T) {
	t.Setenv("LONG_TLS_CERT_FILE", "/tmp/server.crt")

	_, err := resolveServerTLSConfig()
	if !errors.Is(err, errIncompleteTLSFiles) {
		t.Fatalf("expected errIncompleteTLSFiles, got %v", err)
	}
}

func TestResolveServerTLSConfigBuildsConfig(t *testing.T) {
	t.Setenv("LONG_TLS_CERT_FILE", "/tmp/server.crt")
	t.Setenv("LONG_TLS_KEY_FILE", "/tmp/server.key")

	tlsConfig, err := resolveServerTLSConfig()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if tlsConfig == nil {
		t.Fatal("expected tls config")
	}
	if tlsConfig.MinVersion != tls.VersionTLS12 {
		t.Fatalf("expected min tls 1.2, got %v", tlsConfig.MinVersion)
	}
	if len(tlsConfig.NextProtos) < 2 || tlsConfig.NextProtos[0] != "h2" || tlsConfig.NextProtos[1] != "http/1.1" {
		t.Fatalf("expected alpn h2/http1.1, got %+v", tlsConfig.NextProtos)
	}
	if tlsConfig.GetCertificate == nil {
		t.Fatal("expected dynamic certificate loader")
	}
}

func TestConfigureLocalTimezoneSetsAsiaShanghai(t *testing.T) {
	originalLocal := time.Local
	originalTZ, hadTZ := os.LookupEnv("TZ")
	defer func() {
		time.Local = originalLocal
		if hadTZ {
			_ = os.Setenv("TZ", originalTZ)
			return
		}
		_ = os.Unsetenv("TZ")
	}()

	if err := configureLocalTimezone(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if time.Local == nil {
		t.Fatal("expected time.Local to be set")
	}
	if time.Local.String() != "Asia/Shanghai" {
		t.Fatalf("expected Asia/Shanghai, got %q", time.Local.String())
	}
	if got := os.Getenv("TZ"); got != "Asia/Shanghai" {
		t.Fatalf("expected TZ=Asia/Shanghai, got %q", got)
	}
}

func TestRenderStartupInfoIncludesBannerAndSummary(t *testing.T) {
	cfg := config.Config{
		Port:        2333,
		Redis:       config.RedisConfig{Host: "127.0.0.1", Port: 6379, DB: 2, TLSEnabled: true},
		Mongo:       config.MongoConfig{Enabled: true, Database: "vote_wall"},
		Log:         config.LogConfig{Level: "info", Format: "json"},
		RedisPrefix: "hai-world:",
	}

	info := renderStartupInfo(cfg, "127.0.0.1:2333", true)
	assertContains(t, godAnimalArt(), "神兽保佑")
	assertContains(t, godAnimalArt(), "代码无BUG")
	assertContains(t, info, "监听地址: 127.0.0.1:2333")
	assertContains(t, info, "HTTPS: true")
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
