package config

import "testing"

func TestLoadTestReadsConfigTestYAML(t *testing.T) {
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
}
