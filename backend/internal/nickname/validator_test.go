package nickname

import "testing"

func TestValidatorAllowsCleanNickname(t *testing.T) {
	validator := NewValidator([]string{"习近平", "xjp"})

	if err := validator.Validate("阿明在前排"); err != nil {
		t.Fatalf("expected clean nickname to pass, got %v", err)
	}
}

func TestValidatorRejectsSensitiveNickname(t *testing.T) {
	validator := NewValidator([]string{"习近平", "xjp"})

	if err := validator.Validate("我是习近平粉丝"); err != ErrSensitiveNickname {
		t.Fatalf("expected ErrSensitiveNickname, got %v", err)
	}
}

func TestValidatorRejectsCaseInsensitiveSensitiveNickname(t *testing.T) {
	validator := NewValidator([]string{"xjp"})

	if err := validator.Validate("XJP今天来了"); err != ErrSensitiveNickname {
		t.Fatalf("expected ErrSensitiveNickname for case-insensitive match, got %v", err)
	}
}

func TestSensitiveLexiconValidatorLoadsTermsFromVendoredRepository(t *testing.T) {
	validator := NewSensitiveLexiconValidator()

	if err := validator.Validate("今天想找兼职"); err != ErrSensitiveNickname {
		t.Fatalf("expected vendored repository terms to be loaded, got %v", err)
	}
}
