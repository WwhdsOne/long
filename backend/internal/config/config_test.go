package config

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
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
            redis_prefix: "vote:"
            rate_limit:
              limit: 30
              window_ms: 2000
              blacklist_ms: 600000
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
            llm:
              enabled: true
              api_key: "sk-test"
              base_url: "https://llm.example.com/v1/"
              model: "gpt-test"
              timeout_ms: 15000
            log:
              level: "debug"
              format: "console"
              include_caller: true
            mongo:
              enabled: true
              uri: "mongodb://127.0.0.1:27017"
              database: "vote_wall"
              connect_timeout_ms: 3000
              write_timeout_ms: 2000
              read_timeout_ms: 4000
            archive:
              boss_history_dual_write: true
              boss_history_read_source: "mongo"
              admin_audit_enabled: true
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
	if cfg.RedisPrefix != "vote:" {
		t.Fatalf("expected redis prefix vote:button:, got %q", cfg.RedisPrefix)
	}
	if cfg.RateLimit.Limit != 30 {
		t.Fatalf("expected rate limit 30, got %d", cfg.RateLimit.Limit)
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
	if !cfg.LLM.Enabled {
		t.Fatal("expected llm enabled")
	}
	if cfg.LLM.APIKey != "sk-test" {
		t.Fatalf("expected llm api key sk-test, got %q", cfg.LLM.APIKey)
	}
	if cfg.LLM.BaseURL != "https://llm.example.com/v1" {
		t.Fatalf("expected normalized llm base url, got %q", cfg.LLM.BaseURL)
	}
	if cfg.LLM.Model != "gpt-test" {
		t.Fatalf("expected llm model gpt-test, got %q", cfg.LLM.Model)
	}
	if cfg.LLM.Timeout != 15*time.Second {
		t.Fatalf("expected llm timeout 15s, got %s", cfg.LLM.Timeout)
	}
	if cfg.Log.Level != "debug" {
		t.Fatalf("expected log level debug, got %q", cfg.Log.Level)
	}
	if cfg.Log.Format != "console" {
		t.Fatalf("expected log format console, got %q", cfg.Log.Format)
	}
	if !cfg.Log.IncludeCaller {
		t.Fatal("expected include caller enabled")
	}
	if !cfg.Mongo.Enabled {
		t.Fatal("expected mongo enabled")
	}
	if cfg.Mongo.URI != "mongodb://127.0.0.1:27017" {
		t.Fatalf("expected mongo uri, got %q", cfg.Mongo.URI)
	}
	if cfg.Mongo.Database != "vote_wall" {
		t.Fatalf("expected mongo database vote_wall, got %q", cfg.Mongo.Database)
	}
	if cfg.Mongo.ConnectTimeout != 3*time.Second {
		t.Fatalf("expected mongo connect timeout 3s, got %s", cfg.Mongo.ConnectTimeout)
	}
	if cfg.Mongo.WriteTimeout != 2*time.Second {
		t.Fatalf("expected mongo write timeout 2s, got %s", cfg.Mongo.WriteTimeout)
	}
	if cfg.Mongo.ReadTimeout != 4*time.Second {
		t.Fatalf("expected mongo read timeout 4s, got %s", cfg.Mongo.ReadTimeout)
	}
	if !cfg.Archive.BossHistoryDualWrite {
		t.Fatal("expected boss history dual write enabled")
	}
	if cfg.Archive.BossHistoryReadSource != "mongo" {
		t.Fatalf("expected boss history read source mongo, got %q", cfg.Archive.BossHistoryReadSource)
	}
	if !cfg.Archive.AdminAuditEnabled {
		t.Fatal("expected admin audit enabled")
	}
}

func TestValidateAllowsMissingLLMConfigWhenDisabled(t *testing.T) {
	cfg := validConfigForTest()
	cfg.LLM.Enabled = false
	cfg.LLM.APIKey = ""
	cfg.LLM.Model = ""
	cfg.LLM.BaseURL = ""
	cfg.LLM.Timeout = 0

	if err := validate(cfg); err != nil {
		t.Fatalf("expected disabled llm config to be optional, got %v", err)
	}
}

func TestValidateRequiresLLMFieldsWhenEnabled(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{
			name: "api key",
			mutate: func(cfg *Config) {
				cfg.LLM.APIKey = ""
			},
			wantErr: "llm.api_key is required when llm is enabled",
		},
		{
			name: "model",
			mutate: func(cfg *Config) {
				cfg.LLM.Model = ""
			},
			wantErr: "llm.model is required when llm is enabled",
		},
		{
			name: "base url",
			mutate: func(cfg *Config) {
				cfg.LLM.BaseURL = ""
			},
			wantErr: "llm.base_url is required when llm is enabled",
		},
		{
			name: "timeout",
			mutate: func(cfg *Config) {
				cfg.LLM.Timeout = 0
			},
			wantErr: "llm.timeout_ms must be greater than 0 when llm is enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfigForTest()
			cfg.LLM = LLMConfig{
				Enabled: true,
				APIKey:  "sk-test",
				BaseURL: "https://api.openai.com/v1",
				Model:   "gpt-test",
				Timeout: 5 * time.Second,
			}
			tt.mutate(&cfg)

			err := validate(cfg)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("expected %q, got %v", tt.wantErr, err)
			}
		})
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

func TestValidateRequiresMongoFieldsWhenEnabled(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{
			name: "uri",
			mutate: func(cfg *Config) {
				cfg.Mongo.URI = ""
			},
			wantErr: "mongo.uri is required when mongo is enabled",
		},
		{
			name: "database",
			mutate: func(cfg *Config) {
				cfg.Mongo.Database = ""
			},
			wantErr: "mongo.database is required when mongo is enabled",
		},
		{
			name: "connect timeout",
			mutate: func(cfg *Config) {
				cfg.Mongo.ConnectTimeout = 0
			},
			wantErr: "mongo.connect_timeout_ms must be greater than 0 when mongo is enabled",
		},
		{
			name: "write timeout",
			mutate: func(cfg *Config) {
				cfg.Mongo.WriteTimeout = 0
			},
			wantErr: "mongo.write_timeout_ms must be greater than 0 when mongo is enabled",
		},
		{
			name: "read timeout",
			mutate: func(cfg *Config) {
				cfg.Mongo.ReadTimeout = 0
			},
			wantErr: "mongo.read_timeout_ms must be greater than 0 when mongo is enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfigForTest()
			cfg.Mongo = MongoConfig{
				Enabled:        true,
				URI:            "mongodb://127.0.0.1:27017",
				Database:       "vote_wall",
				ConnectTimeout: 3 * time.Second,
				WriteTimeout:   2 * time.Second,
				ReadTimeout:    4 * time.Second,
			}
			tt.mutate(&cfg)

			err := validate(cfg)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("expected %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestValidateRejectsUnknownBossHistoryReadSource(t *testing.T) {
	cfg := validConfigForTest()
	cfg.Archive.BossHistoryReadSource = "postgres"

	err := validate(cfg)
	if err == nil || err.Error() != "archive.boss_history_read_source must be one of redis,mongo" {
		t.Fatalf("expected archive read source validation error, got %v", err)
	}
}

func validConfigForTest() Config {
	return Config{
		Port: 2333,
		Redis: RedisConfig{
			Host: "127.0.0.1",
			Port: 6379,
		},
		RedisPrefix: "vote:",
		RateLimit: RateLimitConfig{
			Limit:             30,
			Window:            2 * time.Second,
			BlacklistDuration: 10 * time.Minute,
		},
		Admin: AdminConfig{
			Username:      "admin",
			Password:      "secret",
			SessionSecret: "session-secret",
		},
		PlayerAuth: PlayerAuthConfig{
			JWTSecret: "player-secret",
			JWTTTL:    7 * 24 * time.Hour,
		},
	}
}
