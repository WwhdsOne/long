package config

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"gopkg.in/yaml.v3"

	"long/internal/xlog"
)

// RedisConfig holds the connection settings for the Redis instance.
type RedisConfig struct {
	Host         string
	Port         int
	Username     string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	TLSEnabled   bool
}

type RateLimitWindowConfig struct {
	Limit  int
	Window time.Duration
}

type AntiScriptPointsConfig struct {
	ClickRateLimitHit        int64
	LoginTurnstileInvalid    int64
	StaminaTurnstileInvalid  int64
	PostStaminaPurchaseClick int64
}

type AntiScriptClickRateLimitConfig struct {
	NicknameWhitelist []string
	Rules             []RateLimitWindowConfig
}

type AntiScriptConfig struct {
	ScoreWindow           time.Duration
	PurchaseClickCooldown time.Duration
	BanThreshold8h        int64
	BanThreshold24h       int64
	BanThreshold72h       int64
	Points                AntiScriptPointsConfig
	ClickRateLimit        AntiScriptClickRateLimitConfig
}

// AdminConfig 管理后台鉴权配置
type AdminConfig struct {
	Username      string
	Password      string
	SessionSecret string
}

// PlayerAuthConfig 玩家账号 JWT 配置。
type PlayerAuthConfig struct {
	JWTSecret string
	JWTTTL    time.Duration
}

// OSSConfig 阿里云 OSS 直传配置
type OSSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	Bucket          string
	Region          string
	PublicBaseURL   string
	UploadDirPrefix string
	ExpireSeconds   int
}

// LLMConfig 控制后台装备草稿生成的大模型调用。
type LLMConfig struct {
	Enabled bool
	APIKey  string
	BaseURL string
	Model   string
	Timeout time.Duration
}

// RealtimeConfig 控制实时链路行为。
type RealtimeConfig struct {
	DebounceMs int
}

// TurnstileConfig 控制购买体力的人机验证。
type TurnstileConfig struct {
	Enabled                   bool
	SiteKey                   string
	SecretKey                 string
	PurchaseStaminaSampleRate float64
	VerifyTimeoutMS           int
}

// LogConfig 控制结构化日志输出行为。
type LogConfig struct {
	Level         string
	Format        string
	IncludeCaller bool
}

// MongoConfig 控制 MongoDB 冷数据存储。
type MongoConfig struct {
	Enabled        bool
	URI            string
	Database       string
	ConnectTimeout time.Duration
	WriteTimeout   time.Duration
	ReadTimeout    time.Duration
}

// ArchiveConfig 控制冷数据归档与读源切换。
type ArchiveConfig struct{}

// RoomConfig 控制房间分线配置。
type RoomConfig struct {
	Enabled        bool
	Count          int
	DefaultRoom    string
	SwitchCooldown time.Duration
}

// Config 运行时配置集合
type Config struct {
	Port        int
	Redis       RedisConfig
	AntiScript  AntiScriptConfig
	Admin       AdminConfig
	PlayerAuth  PlayerAuthConfig
	OSS         OSSConfig
	LLM         LLMConfig
	Realtime    RealtimeConfig
	Turnstile   TurnstileConfig
	Log         LogConfig
	Mongo       MongoConfig
	Archive     ArchiveConfig
	Room        RoomConfig
	RedisPrefix string
}

type fileConfig struct {
	Port  int `yaml:"port"`
	Redis struct {
		Host         string `yaml:"host"`
		Port         int    `yaml:"port"`
		Username     string `yaml:"username"`
		Password     string `yaml:"password"`
		DB           int    `yaml:"db"`
		PoolSize     int    `yaml:"pool_size"`
		MinIdleConns int    `yaml:"min_idle_conns"`
		TLSEnabled   bool   `yaml:"tls_enabled"`
	} `yaml:"redis"`
	RedisPrefix string `yaml:"redis_prefix"`
	AntiScript  struct {
		ScoreWindowSeconds           int64 `yaml:"score_window_seconds"`
		PurchaseClickCooldownSeconds int64 `yaml:"purchase_click_cooldown_seconds"`
		BanThreshold8h               int64 `yaml:"ban_threshold_8h"`
		BanThreshold24h              int64 `yaml:"ban_threshold_24h"`
		BanThreshold72h              int64 `yaml:"ban_threshold_72h"`
		Points                       struct {
			ClickRateLimitHit        int64 `yaml:"click_rate_limit_hit"`
			LoginTurnstileInvalid    int64 `yaml:"login_turnstile_invalid"`
			StaminaTurnstileInvalid  int64 `yaml:"stamina_turnstile_invalid"`
			PostStaminaPurchaseClick int64 `yaml:"post_stamina_purchase_click"`
		} `yaml:"points"`
		ClickRateLimit struct {
			NicknameWhitelist []string `yaml:"nickname_whitelist"`
				Rules             []struct {
					Limit    int `yaml:"limit"`
					WindowMS int `yaml:"window_ms"`
				} `yaml:"rules"`
		} `yaml:"click_rate_limit"`
	} `yaml:"anti_script"`
	Admin struct {
		Username      string `yaml:"username"`
		Password      string `yaml:"password"`
		SessionSecret string `yaml:"session_secret"`
	} `yaml:"admin"`
	PlayerAuth struct {
		JWTSecret    string `yaml:"jwt_secret"`
		JWTTTLSecond int    `yaml:"jwt_ttl_seconds"`
	} `yaml:"player_auth"`
	OSS struct {
		AccessKeyID     string `yaml:"access_key_id"`
		AccessKeySecret string `yaml:"access_key_secret"`
		Bucket          string `yaml:"bucket"`
		Region          string `yaml:"region"`
		PublicBaseURL   string `yaml:"public_base_url"`
		UploadDirPrefix string `yaml:"upload_dir_prefix"`
		ExpireSeconds   int    `yaml:"expire_seconds"`
	} `yaml:"oss"`
	LLM struct {
		Enabled   bool   `yaml:"enabled"`
		APIKey    string `yaml:"api_key"`
		BaseURL   string `yaml:"base_url"`
		Model     string `yaml:"model"`
		TimeoutMS int    `yaml:"timeout_ms"`
	} `yaml:"llm"`
	Realtime struct {
		DebounceMs int `yaml:"debounce_ms"`
	} `yaml:"realtime"`
	Turnstile struct {
		Enabled                   bool    `yaml:"enabled"`
		SiteKey                   string  `yaml:"site_key"`
		SecretKey                 string  `yaml:"secret_key"`
		PurchaseStaminaSampleRate float64 `yaml:"purchase_stamina_sample_rate"`
		VerifyTimeoutMS           int     `yaml:"verify_timeout_ms"`
	} `yaml:"turnstile"`
	Room struct {
		Enabled               bool   `yaml:"enabled"`
		Count                 int    `yaml:"count"`
		DefaultRoom           string `yaml:"default_room"`
		SwitchCooldownSeconds int    `yaml:"switch_cooldown_seconds"`
	} `yaml:"room"`
	Log struct {
		Level         string `yaml:"level"`
		Format        string `yaml:"format"`
		IncludeCaller bool   `yaml:"include_caller"`
	} `yaml:"log"`
	Mongo struct {
		Enabled          bool   `yaml:"enabled"`
		URI              string `yaml:"uri"`
		Database         string `yaml:"database"`
		ConnectTimeoutMS int    `yaml:"connect_timeout_ms"`
		WriteTimeoutMS   int    `yaml:"write_timeout_ms"`
		ReadTimeoutMS    int    `yaml:"read_timeout_ms"`
	} `yaml:"mongo"`
	Archive struct{} `yaml:"archive"`
}

type consulKV struct {
	Value string `json:"Value"`
}

var exitProcess = os.Exit

// Load 从 Consul 加载运行时配置
func Load() (Config, error) {
	cfg, source, err := loadFromConsul()
	if err != nil {
		return Config{}, err
	}

	go watchConsulConfig(source.addr, source.key, source.index)

	return cfg, nil
}

// LoadTest reads configuration from Consul without starting the watcher.
func LoadTest() (Config, error) {
	cfg, _, err := loadFromConsul()
	return cfg, err
}

type consulSource struct {
	addr  string
	key   string
	index string
}

func loadFromConsul() (Config, consulSource, error) {
	source, err := consulSourceFromEnv()
	if err != nil {
		return Config{}, consulSource{}, err
	}

	payload, index, err := fetchConfigPayload(context.Background(), source.addr, source.key, "")
	if err != nil {
		return Config{}, consulSource{}, err
	}

	var parsed fileConfig
	if err := yaml.Unmarshal(payload, &parsed); err != nil {
		return Config{}, consulSource{}, fmt.Errorf("parse consul config: %w", err)
	}

	config := Config{
		Port: parsed.Port,
		Redis: RedisConfig{
			Host:         parsed.Redis.Host,
			Port:         parsed.Redis.Port,
			Username:     parsed.Redis.Username,
			Password:     parsed.Redis.Password,
			DB:           parsed.Redis.DB,
			PoolSize:     parsed.Redis.PoolSize,
			MinIdleConns: parsed.Redis.MinIdleConns,
			TLSEnabled:   parsed.Redis.TLSEnabled,
		},
		AntiScript: AntiScriptConfig{
			ScoreWindow:           time.Duration(parsed.AntiScript.ScoreWindowSeconds) * time.Second,
			PurchaseClickCooldown: time.Duration(parsed.AntiScript.PurchaseClickCooldownSeconds) * time.Second,
			BanThreshold8h:        parsed.AntiScript.BanThreshold8h,
			BanThreshold24h:       parsed.AntiScript.BanThreshold24h,
			BanThreshold72h:       parsed.AntiScript.BanThreshold72h,
			Points: AntiScriptPointsConfig{
				ClickRateLimitHit:        parsed.AntiScript.Points.ClickRateLimitHit,
				LoginTurnstileInvalid:    parsed.AntiScript.Points.LoginTurnstileInvalid,
				StaminaTurnstileInvalid:  parsed.AntiScript.Points.StaminaTurnstileInvalid,
				PostStaminaPurchaseClick: parsed.AntiScript.Points.PostStaminaPurchaseClick,
			},
			ClickRateLimit: AntiScriptClickRateLimitConfig{
				NicknameWhitelist: normalizeStringList(parsed.AntiScript.ClickRateLimit.NicknameWhitelist),
				Rules:             convertRateLimitRules(parsed.AntiScript.ClickRateLimit.Rules),
			},
		},
		Admin: AdminConfig{
			Username:      parsed.Admin.Username,
			Password:      parsed.Admin.Password,
			SessionSecret: parsed.Admin.SessionSecret,
		},
		PlayerAuth: PlayerAuthConfig{
			JWTSecret: parsed.PlayerAuth.JWTSecret,
			JWTTTL:    time.Duration(parsed.PlayerAuth.JWTTTLSecond) * time.Second,
		},
		OSS: OSSConfig{
			AccessKeyID:     parsed.OSS.AccessKeyID,
			AccessKeySecret: parsed.OSS.AccessKeySecret,
			Bucket:          parsed.OSS.Bucket,
			Region:          parsed.OSS.Region,
			PublicBaseURL:   parsed.OSS.PublicBaseURL,
			UploadDirPrefix: parsed.OSS.UploadDirPrefix,
			ExpireSeconds:   parsed.OSS.ExpireSeconds,
		},
		LLM: LLMConfig{
			Enabled: parsed.LLM.Enabled,
			APIKey:  strings.TrimSpace(parsed.LLM.APIKey),
			BaseURL: normalizeLLMBaseURL(parsed.LLM.BaseURL),
			Model:   strings.TrimSpace(parsed.LLM.Model),
			Timeout: time.Duration(parsed.LLM.TimeoutMS) * time.Millisecond,
		},
		Realtime: RealtimeConfig{
			DebounceMs: parsed.Realtime.DebounceMs,
		},
		Turnstile: TurnstileConfig{
			Enabled:                   parsed.Turnstile.Enabled,
			SiteKey:                   strings.TrimSpace(parsed.Turnstile.SiteKey),
			SecretKey:                 strings.TrimSpace(parsed.Turnstile.SecretKey),
			PurchaseStaminaSampleRate: parsed.Turnstile.PurchaseStaminaSampleRate,
			VerifyTimeoutMS:           parsed.Turnstile.VerifyTimeoutMS,
		},
		Log: LogConfig{
			Level:         normalizeLogLevel(parsed.Log.Level),
			Format:        normalizeLogFormat(parsed.Log.Format),
			IncludeCaller: parsed.Log.IncludeCaller,
		},
		Mongo: MongoConfig{
			Enabled:        parsed.Mongo.Enabled,
			URI:            strings.TrimSpace(parsed.Mongo.URI),
			Database:       strings.TrimSpace(parsed.Mongo.Database),
			ConnectTimeout: time.Duration(parsed.Mongo.ConnectTimeoutMS) * time.Millisecond,
			WriteTimeout:   time.Duration(parsed.Mongo.WriteTimeoutMS) * time.Millisecond,
			ReadTimeout:    time.Duration(parsed.Mongo.ReadTimeoutMS) * time.Millisecond,
		},
		Room: RoomConfig{
			Enabled:        parsed.Room.Enabled,
			Count:          parsed.Room.Count,
			DefaultRoom:    strings.TrimSpace(parsed.Room.DefaultRoom),
			SwitchCooldown: time.Duration(parsed.Room.SwitchCooldownSeconds) * time.Second,
		},
		Archive:     ArchiveConfig{},
		RedisPrefix: parsed.RedisPrefix,
	}
	if config.Realtime.DebounceMs <= 0 {
		config.Realtime.DebounceMs = 50
	}
	if config.Turnstile.VerifyTimeoutMS <= 0 {
		config.Turnstile.VerifyTimeoutMS = 3000
	}
	config.Room = normalizeRoomConfig(config.Room)

	if err := validate(config); err != nil {
		return Config{}, consulSource{}, fmt.Errorf("validate consul config: %w", err)
	}

	return config, consulSource{
		addr:  source.addr,
		key:   source.key,
		index: index,
	}, nil
}

func validate(config Config) error {
	switch {
	case config.Port <= 0:
		return errors.New("port must be greater than 0")
	case config.Redis.Host == "":
		return errors.New("redis.host is required")
	case config.Redis.Port <= 0:
		return errors.New("redis.port must be greater than 0")
	case config.RedisPrefix == "":
		return errors.New("redis_prefix is required")
	case config.AntiScript.ScoreWindow <= 0:
		return errors.New("anti_script.score_window_seconds must be greater than 0")
	case config.AntiScript.PurchaseClickCooldown < 0:
		return errors.New("anti_script.purchase_click_cooldown_seconds must be greater than or equal to 0")
	case config.AntiScript.BanThreshold8h <= 0:
		return errors.New("anti_script.ban_threshold_8h must be greater than 0")
	case config.AntiScript.BanThreshold24h <= config.AntiScript.BanThreshold8h:
		return errors.New("anti_script.ban_threshold_24h must be greater than ban_threshold_8h")
	case config.AntiScript.BanThreshold72h <= config.AntiScript.BanThreshold24h:
		return errors.New("anti_script.ban_threshold_72h must be greater than ban_threshold_24h")
	case config.AntiScript.Points.ClickRateLimitHit <= 0:
		return errors.New("anti_script.points.click_rate_limit_hit must be greater than 0")
	case config.AntiScript.Points.LoginTurnstileInvalid <= 0:
		return errors.New("anti_script.points.login_turnstile_invalid must be greater than 0")
	case config.AntiScript.Points.StaminaTurnstileInvalid <= 0:
		return errors.New("anti_script.points.stamina_turnstile_invalid must be greater than 0")
	case config.AntiScript.Points.PostStaminaPurchaseClick <= 0:
		return errors.New("anti_script.points.post_stamina_purchase_click must be greater than 0")
	case len(config.AntiScript.ClickRateLimit.Rules) == 0:
		return errors.New("anti_script.click_rate_limit.rules must not be empty")
	case !validateClickRateLimitRules(config.AntiScript.ClickRateLimit.Rules):
		return errors.New("anti_script.click_rate_limit.rules[].limit must be greater than 0 and window_ms must be greater than 0")
	case strings.TrimSpace(config.Admin.Username) == "":
		return errors.New("admin.username is required")
	case strings.TrimSpace(config.Admin.Password) == "":
		return errors.New("admin.password is required")
	case strings.TrimSpace(config.Admin.SessionSecret) == "":
		return errors.New("admin.session_secret is required")
	case strings.TrimSpace(config.PlayerAuth.JWTSecret) == "":
		return errors.New("player_auth.jwt_secret is required")
	case config.PlayerAuth.JWTTTL <= 0:
		return errors.New("player_auth.jwt_ttl_seconds must be greater than 0")
	case config.OSS.Enabled() && strings.TrimSpace(config.OSS.AccessKeyID) == "":
		return errors.New("oss.access_key_id is required when oss is configured")
	case config.OSS.Enabled() && strings.TrimSpace(config.OSS.AccessKeySecret) == "":
		return errors.New("oss.access_key_secret is required when oss is configured")
	case config.OSS.Enabled() && strings.TrimSpace(config.OSS.Bucket) == "":
		return errors.New("oss.bucket is required when oss is configured")
	case config.OSS.Enabled() && strings.TrimSpace(config.OSS.Region) == "":
		return errors.New("oss.region is required when oss is configured")
	case config.LLM.Enabled && strings.TrimSpace(config.LLM.APIKey) == "":
		return errors.New("llm.api_key is required when llm is enabled")
	case config.LLM.Enabled && strings.TrimSpace(config.LLM.Model) == "":
		return errors.New("llm.model is required when llm is enabled")
	case config.LLM.Enabled && strings.TrimSpace(config.LLM.BaseURL) == "":
		return errors.New("llm.base_url is required when llm is enabled")
	case config.LLM.Enabled && config.LLM.Timeout <= 0:
		return errors.New("llm.timeout_ms must be greater than 0 when llm is enabled")
	case config.Turnstile.Enabled && strings.TrimSpace(config.Turnstile.SiteKey) == "":
		return errors.New("turnstile.site_key is required when turnstile is enabled")
	case config.Turnstile.Enabled && strings.TrimSpace(config.Turnstile.SecretKey) == "":
		return errors.New("turnstile.secret_key is required when turnstile is enabled")
	case config.Turnstile.PurchaseStaminaSampleRate < 0 || config.Turnstile.PurchaseStaminaSampleRate > 1:
		return errors.New("turnstile.purchase_stamina_sample_rate must be between 0 and 1")
	case config.Turnstile.VerifyTimeoutMS <= 0:
		return errors.New("turnstile.verify_timeout_ms must be greater than 0")
	case config.Mongo.Enabled && strings.TrimSpace(config.Mongo.URI) == "":
		return errors.New("mongo.uri is required when mongo is enabled")
	case config.Mongo.Enabled && strings.TrimSpace(config.Mongo.Database) == "":
		return errors.New("mongo.database is required when mongo is enabled")
	case config.Mongo.Enabled && config.Mongo.ConnectTimeout <= 0:
		return errors.New("mongo.connect_timeout_ms must be greater than 0 when mongo is enabled")
	case config.Mongo.Enabled && config.Mongo.WriteTimeout <= 0:
		return errors.New("mongo.write_timeout_ms must be greater than 0 when mongo is enabled")
	case config.Mongo.Enabled && config.Mongo.ReadTimeout <= 0:
		return errors.New("mongo.read_timeout_ms must be greater than 0 when mongo is enabled")
	}

	return nil
}

func normalizeRoomConfig(cfg RoomConfig) RoomConfig {
	if cfg.Count <= 0 {
		cfg.Count = 1
	}
	cfg.DefaultRoom = strings.TrimSpace(cfg.DefaultRoom)
	if cfg.SwitchCooldown < 0 {
		cfg.SwitchCooldown = 0
	}
	return cfg
}

func normalizeStringList(items []string) []string {
	if len(items) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func convertRateLimitRules(rules []struct {
	Limit    int `yaml:"limit"`
	WindowMS int `yaml:"window_ms"`
}) []RateLimitWindowConfig {
	if len(rules) == 0 {
		return nil
	}
	result := make([]RateLimitWindowConfig, 0, len(rules))
	for _, r := range rules {
		result = append(result, RateLimitWindowConfig{
			Limit:  r.Limit,
			Window: time.Duration(r.WindowMS) * time.Millisecond,
		})
	}
	return result
}

func validateClickRateLimitRules(rules []RateLimitWindowConfig) bool {
	for _, r := range rules {
		if r.Limit <= 0 || r.Window <= 0 {
			return false
		}
	}
	return true
}


func normalizeLLMBaseURL(baseURL string) string {
	trimmed := strings.TrimSpace(baseURL)
	if trimmed == "" {
		return "https://api.openai.com/v1"
	}
	return strings.TrimRight(trimmed, "/")
}

func normalizeLogLevel(level string) string {
	trimmed := strings.ToLower(strings.TrimSpace(level))
	if trimmed == "" {
		return "info"
	}
	return trimmed
}

func normalizeLogFormat(format string) string {
	trimmed := strings.ToLower(strings.TrimSpace(format))
	if trimmed == "" {
		return "json"
	}
	return trimmed
}

// Enabled reports whether OSS direct-upload has been configured.
func (c OSSConfig) Enabled() bool {
	return strings.TrimSpace(c.AccessKeyID) != "" ||
		strings.TrimSpace(c.AccessKeySecret) != "" ||
		strings.TrimSpace(c.Bucket) != "" ||
		strings.TrimSpace(c.Region) != "" ||
		strings.TrimSpace(c.PublicBaseURL) != ""
}

func consulSourceFromEnv() (consulSource, error) {
	addr := strings.TrimSpace(os.Getenv("CONSUL_ADDR"))
	if addr == "" {
		return consulSource{}, errors.New("CONSUL_ADDR is required")
	}

	key := strings.TrimSpace(os.Getenv("CONSUL_CONFIG_KEY"))
	if key == "" {
		return consulSource{}, errors.New("CONSUL_CONFIG_KEY is required")
	}

	return consulSource{
		addr: normalizeConsulAddr(addr),
		key:  strings.TrimPrefix(key, "/"),
	}, nil
}

func normalizeConsulAddr(addr string) string {
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return strings.TrimRight(addr, "/")
	}

	return "http://" + strings.TrimRight(addr, "/")
}

func fetchConfigPayload(ctx context.Context, consulAddr, configKey, index string) ([]byte, string, error) {
	requestURL := fmt.Sprintf("%s/v1/kv/%s", consulAddr, configKey)
	if index != "" {
		requestURL = fmt.Sprintf("%s?wait=5m&index=%s", requestURL, index)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("build consul request: %w", err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, "", fmt.Errorf("fetch consul config: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("consul returned status %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read consul response: %w", err)
	}

	var kvs []consulKV
	if err := sonic.Unmarshal(body, &kvs); err != nil {
		return nil, "", fmt.Errorf("decode consul response: %w", err)
	}
	if len(kvs) == 0 {
		return nil, "", errors.New("consul response is empty")
	}

	decoded, err := base64.StdEncoding.DecodeString(kvs[0].Value)
	if err != nil {
		return nil, "", fmt.Errorf("decode consul config value: %w", err)
	}

	return decoded, response.Header.Get("X-Consul-Index"), nil
}

func watchConsulConfig(consulAddr, configKey, lastIndex string) {
	for {
		_, nextIndex, err := fetchConfigPayload(context.Background(), consulAddr, configKey, lastIndex)
		if err != nil {
			xlog.L().Error("watch consul config failed", xlog.Err(err))
			time.Sleep(10 * time.Second)
			continue
		}

		if nextIndex == "" || nextIndex == lastIndex {
			continue
		}

		xlog.L().Info("consul config changed, exiting for restart")
		exitProcess(0)
	}
}
