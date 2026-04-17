package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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

// Load reads the runtime configuration from backend/config.yaml.
func Load() (Config, error) {
	return loadFromFile("config.yaml")
}

// LoadTest reads the dedicated test configuration from backend/config.test.yaml.
func LoadTest() (Config, error) {
	return loadFromFile("config.test.yaml")
}

func loadFromFile(filename string) (Config, error) {
	configPath := resolveConfigPath(filename)

	payload, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("read %s: %w", filename, err)
	}

	var source fileConfig
	if err := yaml.Unmarshal(payload, &source); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", filename, err)
	}

	config := Config{
		Port: source.Port,
		Redis: RedisConfig{
			Host:       source.Redis.Host,
			Port:       source.Redis.Port,
			Username:   source.Redis.Username,
			Password:   source.Redis.Password,
			DB:         source.Redis.DB,
			TLSEnabled: source.Redis.TLSEnabled,
		},
		RateLimit: RateLimitConfig{
			Limit:             source.RateLimit.Limit,
			Window:            time.Duration(source.RateLimit.WindowMS) * time.Millisecond,
			BlacklistDuration: time.Duration(source.RateLimit.BlacklistMS) * time.Millisecond,
		},
		CriticalHit: CriticalHitConfig{
			ChancePercent: source.CriticalHit.ChancePercent,
			Count:         source.CriticalHit.Count,
		},
		RedisPrefix:        source.RedisPrefix,
		ButtonPollInterval: time.Duration(source.ButtonPollIntervalMS) * time.Millisecond,
		PublicDir:          resolvePublicDir(),
	}

	if err := validate(config); err != nil {
		return Config{}, fmt.Errorf("validate %s: %w", filename, err)
	}

	return config, nil
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

func resolveConfigPath(filename string) string {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return filename
	}

	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", filename))
}

func resolvePublicDir() string {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Clean(filepath.Join("..", "..", "public"))
	}

	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "public"))
}
