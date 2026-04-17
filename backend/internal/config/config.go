package config

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ButtonSeed describes one built-in button that should exist in an empty Redis.
type ButtonSeed struct {
	Slug      string
	Label     string
	Sort      int
	ImagePath string
	ImageAlt  string
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

// CriticalHitConfig controls the optional crit mechanic for button clicks.
type CriticalHitConfig struct {
	ChancePercent int
	Count         int64
}

// Config bundles every runtime setting needed by the backend service.
type Config struct {
	Port               int
	Redis              RedisConfig
	RateLimit          RateLimitConfig
	CriticalHit        CriticalHitConfig
	RedisPrefix        string
	ButtonPollInterval time.Duration
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
	CriticalHit struct {
		ChancePercent int   `yaml:"chance_percent"`
		Count         int64 `yaml:"count"`
	} `yaml:"critical_hit"`
}

type consulKV struct {
	Value string `json:"Value"`
}

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

// Load reads the runtime configuration from Consul and starts a config watcher.
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
		CriticalHit: CriticalHitConfig{
			ChancePercent: parsed.CriticalHit.ChancePercent,
			Count:         parsed.CriticalHit.Count,
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
	case config.CriticalHit.ChancePercent <= 0 || config.CriticalHit.ChancePercent > 100:
		return errors.New("critical_hit.chance_percent must be between 1 and 100")
	case config.CriticalHit.Count <= 1:
		return errors.New("critical_hit.count must be greater than 1")
	}

	return nil
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
	if err := json.Unmarshal(body, &kvs); err != nil {
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
