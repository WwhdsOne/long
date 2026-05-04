package toolconfig

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"long/internal/config"
)

type Options struct {
	NeedMongo   bool
	IncludeRoom bool
}

func Load(opts Options) (config.Config, error) {
	redisPort, err := requiredInt("REDIS_PORT")
	if err != nil {
		return config.Config{}, err
	}

	redisDB, err := optionalInt("REDIS_DB", 0)
	if err != nil {
		return config.Config{}, err
	}

	redisTLS, err := optionalBool("REDIS_TLS_ENABLED", false)
	if err != nil {
		return config.Config{}, err
	}

	cfg := config.Config{
		Redis: config.RedisConfig{
			Host:       requiredString("REDIS_HOST"),
			Port:       redisPort,
			Username:   strings.TrimSpace(os.Getenv("REDIS_USERNAME")),
			Password:   os.Getenv("REDIS_PASSWORD"),
			DB:         redisDB,
			TLSEnabled: redisTLS,
		},
		RedisPrefix: requiredString("REDIS_PREFIX"),
	}

	if cfg.Redis.Host == "" {
		return config.Config{}, fmt.Errorf("环境变量 %s 必填", "REDIS_HOST")
	}
	if cfg.RedisPrefix == "" {
		return config.Config{}, fmt.Errorf("环境变量 %s 必填", "REDIS_PREFIX")
	}

	mongoEnabled := opts.NeedMongo
	if !mongoEnabled {
		mongoEnabled, err = optionalBool("MONGO_ENABLED", false)
		if err != nil {
			return config.Config{}, err
		}
	}
	if mongoEnabled {
		connectTimeoutMS, err := optionalInt("MONGO_CONNECT_TIMEOUT_MS", 3000)
		if err != nil {
			return config.Config{}, err
		}
		writeTimeoutMS, err := optionalInt("MONGO_WRITE_TIMEOUT_MS", 5000)
		if err != nil {
			return config.Config{}, err
		}
		readTimeoutMS, err := optionalInt("MONGO_READ_TIMEOUT_MS", 5000)
		if err != nil {
			return config.Config{}, err
		}

		cfg.Mongo = config.MongoConfig{
			Enabled:        true,
			URI:            requiredString("MONGO_URI"),
			Database:       requiredString("MONGO_DATABASE"),
			ConnectTimeout: time.Duration(connectTimeoutMS) * time.Millisecond,
			WriteTimeout:   time.Duration(writeTimeoutMS) * time.Millisecond,
			ReadTimeout:    time.Duration(readTimeoutMS) * time.Millisecond,
		}
		if cfg.Mongo.URI == "" {
			return config.Config{}, fmt.Errorf("环境变量 %s 必填", "MONGO_URI")
		}
		if cfg.Mongo.Database == "" {
			return config.Config{}, fmt.Errorf("环境变量 %s 必填", "MONGO_DATABASE")
		}
	}

	if opts.IncludeRoom {
		roomEnabled, err := optionalBool("ROOM_ENABLED", true)
		if err != nil {
			return config.Config{}, err
		}
		switchCooldownSeconds, err := optionalInt("ROOM_SWITCH_COOLDOWN_SECONDS", 0)
		if err != nil {
			return config.Config{}, err
		}
		roomCount, err := optionalInt("ROOM_COUNT", 1)
		if err != nil {
			return config.Config{}, err
		}
		if roomCount <= 0 {
			roomCount = 1
		}
		defaultRoom := strings.TrimSpace(os.Getenv("ROOM_DEFAULT"))
		if defaultRoom == "" {
			defaultRoom = "1"
		}
		cfg.Room = config.RoomConfig{
			Enabled:        roomEnabled,
			Count:          roomCount,
			DefaultRoom:    defaultRoom,
			SwitchCooldown: time.Duration(switchCooldownSeconds) * time.Second,
		}
	}

	return cfg, nil
}

func requiredString(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func requiredInt(key string) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return 0, fmt.Errorf("环境变量 %s 必填", key)
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("环境变量 %s 不是合法整数: %w", key, err)
	}
	return value, nil
}

func optionalInt(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("环境变量 %s 不是合法整数: %w", key, err)
	}
	return value, nil
}

func optionalBool(key string, fallback bool) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("环境变量 %s 不是合法布尔值: %w", key, err)
	}
	return value, nil
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}
