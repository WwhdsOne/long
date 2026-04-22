package config

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestLoadTestReadsConfigFromConsul(t *testing.T) {
	t.Setenv("CONSUL_CONFIG_KEY", "vote-wall/test")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/kv/vote-wall/test" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		payload := base64.StdEncoding.EncodeToString([]byte(`
            port: 2333
            redis:
              host: 47.93.83.136
              port: 6379
              username: ""
              password: "Wwh852456"
              db: 3
              tls_enabled: false
            redis_prefix: "vote:button:"
            button_poll_interval_ms: 3000
            rate_limit:
              limit: 30
              window_ms: 2000
              blacklist_ms: 600000
            critical_hit:
              chance_percent: 5
              count: 5
            admin:
              username: "admin"
              password: "secret"
              session_secret: "session-secret"
            player_auth:
              jwt_secret: "player-secret"
              jwt_ttl_seconds: 604800
            oss:
              access_key_id: "test-ak"
              access_key_secret: "test-secret"
              bucket: "vote-wall"
              region: "cn-beijing"
              public_base_url: "https://cdn.example.com"
              upload_dir_prefix: "buttons"
        `))

		w.Header().Set("X-Consul-Index", "7")
		fmt.Fprintf(w, `[{"Value":%q}]`, payload)
	}))
	defer server.Close()

	t.Setenv("CONSUL_ADDR", server.URL)

	cfg, err := LoadTest()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Port != 2333 {
		t.Fatalf("expected port 2333, got %d", cfg.Port)
	}
	if cfg.RedisPrefix != "vote:button:" {
		t.Fatalf("expected redis prefix vote:button:, got %q", cfg.RedisPrefix)
	}
	if cfg.RateLimit.Limit != 30 {
		t.Fatalf("expected rate limit 30, got %d", cfg.RateLimit.Limit)
	}
	if cfg.CriticalHit.ChancePercent != 5 {
		t.Fatalf("expected critical chance 5, got %d", cfg.CriticalHit.ChancePercent)
	}
	if cfg.CriticalHit.Count != 5 {
		t.Fatalf("expected critical count 5, got %d", cfg.CriticalHit.Count)
	}
	if cfg.Admin.Username != "admin" {
		t.Fatalf("expected admin username admin, got %q", cfg.Admin.Username)
	}
	if cfg.Admin.Password != "secret" {
		t.Fatalf("expected admin password secret, got %q", cfg.Admin.Password)
	}
	if cfg.Admin.SessionSecret != "session-secret" {
		t.Fatalf("expected admin session secret session-secret, got %q", cfg.Admin.SessionSecret)
	}
	if cfg.PlayerAuth.JWTSecret != "player-secret" {
		t.Fatalf("expected player jwt secret player-secret, got %q", cfg.PlayerAuth.JWTSecret)
	}
	if cfg.OSS.AccessKeyID != "test-ak" {
		t.Fatalf("expected oss access key id test-ak, got %q", cfg.OSS.AccessKeyID)
	}
	if cfg.OSS.Bucket != "vote-wall" {
		t.Fatalf("expected oss bucket vote-wall, got %q", cfg.OSS.Bucket)
	}
	if cfg.OSS.PublicBaseURL != "https://cdn.example.com" {
		t.Fatalf("expected oss public base url, got %q", cfg.OSS.PublicBaseURL)
	}
}

func TestLoadTestRequiresConsulEnv(t *testing.T) {
	if err := os.Unsetenv("CONSUL_ADDR"); err != nil {
		t.Fatalf("unset CONSUL_ADDR: %v", err)
	}
	if err := os.Unsetenv("CONSUL_CONFIG_KEY"); err != nil {
		t.Fatalf("unset CONSUL_CONFIG_KEY: %v", err)
	}

	if _, err := LoadTest(); err == nil {
		t.Fatal("expected missing env error")
	}
}
