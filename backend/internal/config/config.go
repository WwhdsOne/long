package config

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"gopkg.in/yaml.v3"
)

// ButtonSeed 内置按钮种子数据
type ButtonSeed struct {
	Slug      string // 按钮标识
	Label     string // 显示名称
	Sort      int    // 排序权重
	ImagePath string // 图片路径
	ImageAlt  string // 图片描述
}

// RedisConfig holds the connection settings for the Redis instance.
type RedisConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	DB         int
	TLSEnabled bool
}

// RateLimitConfig controls the in-memory anti-abuse policy for click requests.
type RateLimitConfig struct {
	Limit             int
	Window            time.Duration
	BlacklistDuration time.Duration
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

// Config 运行时配置集合
type Config struct {
	Port               int
	Redis              RedisConfig
	RateLimit          RateLimitConfig
	Admin              AdminConfig
	PlayerAuth         PlayerAuthConfig
	OSS                OSSConfig
	LLM                LLMConfig
	RedisPrefix        string
	ButtonPollInterval time.Duration // 低频兜底按钮索引同步间隔
	PublicDir          string
}

type fileConfig struct {
	Port  int `yaml:"port"`
	Redis struct {
		Host       string `yaml:"host"`
		Port       int    `yaml:"port"`
		Username   string `yaml:"username"`
		Password   string `yaml:"password"`
		DB         int    `yaml:"db"`
		TLSEnabled bool   `yaml:"tls_enabled"`
	} `yaml:"redis"`
	RedisPrefix          string `yaml:"redis_prefix"`
	ButtonPollIntervalMS int    `yaml:"button_poll_interval_ms"`
	RateLimit            struct {
		Limit       int `yaml:"limit"`
		WindowMS    int `yaml:"window_ms"`
		BlacklistMS int `yaml:"blacklist_ms"`
	} `yaml:"rate_limit"`
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
}

type consulKV struct {
	Value string `json:"Value"`
}

// DefaultButtons 内置默认按钮列表
var DefaultButtons = []ButtonSeed{
	{Slug: "feel", Label: "有感觉吗", Sort: 10},
	{Slug: "understand", Label: "有没有懂的", Sort: 20},
	{
		Slug:      "wechat-pity",
		Label:     "微信[可怜]表情",
		Sort:      30,
		ImagePath: "/images/emojipedia-wechat-whimper.png",
		ImageAlt:  "微信可怜表情",
	},
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
			Host:       parsed.Redis.Host,
			Port:       parsed.Redis.Port,
			Username:   parsed.Redis.Username,
			Password:   parsed.Redis.Password,
			DB:         parsed.Redis.DB,
			TLSEnabled: parsed.Redis.TLSEnabled,
		},
		RateLimit: RateLimitConfig{
			Limit:             parsed.RateLimit.Limit,
			Window:            time.Duration(parsed.RateLimit.WindowMS) * time.Millisecond,
			BlacklistDuration: time.Duration(parsed.RateLimit.BlacklistMS) * time.Millisecond,
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
		RedisPrefix:        parsed.RedisPrefix,
		ButtonPollInterval: time.Duration(parsed.ButtonPollIntervalMS) * time.Millisecond,
		PublicDir:          resolvePublicDir(),
	}

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
	case config.ButtonPollInterval <= 0:
		return errors.New("button_poll_interval_ms must be greater than 0")
	case config.RateLimit.Limit <= 0:
		return errors.New("rate_limit.limit must be greater than 0")
	case config.RateLimit.Window <= 0:
		return errors.New("rate_limit.window_ms must be greater than 0")
	case config.RateLimit.BlacklistDuration <= 0:
		return errors.New("rate_limit.blacklist_ms must be greater than 0")
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
	}

	return nil
}

func normalizeLLMBaseURL(baseURL string) string {
	trimmed := strings.TrimSpace(baseURL)
	if trimmed == "" {
		return "https://api.openai.com/v1"
	}
	return strings.TrimRight(trimmed, "/")
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
			log.Printf("watch consul config failed: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		if nextIndex == "" || nextIndex == lastIndex {
			continue
		}

		log.Printf("consul config changed, exiting for restart")
		exitProcess(0)
	}
}

func resolvePublicDir() string {
	return "public"
}
