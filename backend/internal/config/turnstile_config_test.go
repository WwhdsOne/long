package config

import "testing"

func TestValidateAllowsMissingTurnstileKeysWhenDisabled(t *testing.T) {
	cfg := validConfigForTest()
	cfg.Turnstile = TurnstileConfig{
		Enabled:                   false,
		PurchaseStaminaSampleRate: 0.5,
		VerifyTimeoutMS:           3000,
	}

	if err := validate(cfg); err != nil {
		t.Fatalf("expected disabled turnstile config to be optional, got %v", err)
	}
}

func TestValidateRequiresTurnstileFieldsWhenEnabled(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{
			name: "site key",
			mutate: func(cfg *Config) {
				cfg.Turnstile.SiteKey = ""
			},
			wantErr: "turnstile.site_key is required when turnstile is enabled",
		},
		{
			name: "secret key",
			mutate: func(cfg *Config) {
				cfg.Turnstile.SecretKey = ""
			},
			wantErr: "turnstile.secret_key is required when turnstile is enabled",
		},
		{
			name: "sample rate too small",
			mutate: func(cfg *Config) {
				cfg.Turnstile.PurchaseStaminaSampleRate = -0.1
			},
			wantErr: "turnstile.purchase_stamina_sample_rate must be between 0 and 1",
		},
		{
			name: "sample rate too large",
			mutate: func(cfg *Config) {
				cfg.Turnstile.PurchaseStaminaSampleRate = 1.1
			},
			wantErr: "turnstile.purchase_stamina_sample_rate must be between 0 and 1",
		},
		{
			name: "verify timeout",
			mutate: func(cfg *Config) {
				cfg.Turnstile.VerifyTimeoutMS = 0
			},
			wantErr: "turnstile.verify_timeout_ms must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfigForTest()
			cfg.Turnstile = TurnstileConfig{
				Enabled:                   true,
				SiteKey:                   "site-key",
				SecretKey:                 "secret-key",
				PurchaseStaminaSampleRate: 0.5,
				VerifyTimeoutMS:           3000,
			}
			tt.mutate(&cfg)

			err := validate(cfg)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("expected %q, got %v", tt.wantErr, err)
			}
		})
	}
}
