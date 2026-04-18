package admin

import "testing"

func TestAuthenticatorIssuesAndVerifiesSignedSession(t *testing.T) {
	auth := NewAuthenticator(Config{
		Username:      "admin",
		Password:      "secret",
		SessionSecret: "session-secret",
	})

	token, ok := auth.Login("admin", "secret")
	if !ok {
		t.Fatal("expected login to succeed")
	}

	if token == "" {
		t.Fatal("expected non-empty session token")
	}

	if !auth.Verify(token) {
		t.Fatal("expected issued token to verify")
	}
}

func TestAuthenticatorRejectsWrongPasswordAndTampering(t *testing.T) {
	auth := NewAuthenticator(Config{
		Username:      "admin",
		Password:      "secret",
		SessionSecret: "session-secret",
	})

	if _, ok := auth.Login("admin", "wrong"); ok {
		t.Fatal("expected login to fail with wrong password")
	}

	token, ok := auth.Login("admin", "secret")
	if !ok {
		t.Fatal("expected login to succeed")
	}

	if auth.Verify(token + "tampered") {
		t.Fatal("expected tampered token to fail verification")
	}
}
