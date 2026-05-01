package admin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

// Config 管理后台基础认证配置
type Config struct {
	Username      string
	Password      string
	SessionSecret string
}

// Authenticator 负责签发和校验简单会话令牌。
type Authenticator struct {
	username string
	password string
	secret   []byte
}

// NewAuthenticator 创建一个基于固定口令的认证器。
func NewAuthenticator(cfg Config) *Authenticator {
	return &Authenticator{
		username: strings.TrimSpace(cfg.Username),
		password: cfg.Password,
		secret:   []byte(cfg.SessionSecret),
	}
}

// Login 校验账号口令并生成签名会话令牌。
func (a *Authenticator) Login(username string, password string) (string, bool) {
	if strings.TrimSpace(username) != a.username || password != a.password {
		return "", false
	}

	payload := a.username
	signature := a.sign(payload)
	return base64.RawURLEncoding.EncodeToString([]byte(payload + "." + signature)), true
}

// Verify 校验会话令牌是否由当前配置签发。
func (a *Authenticator) Verify(token string) bool {
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(token))
	if err != nil {
		return false
	}

	parts := strings.Split(string(decoded), ".")
	if len(parts) != 2 {
		return false
	}
	if parts[0] != a.username {
		return false
	}

	expected := a.sign(parts[0])
	return hmac.Equal([]byte(parts[1]), []byte(expected))
}

// Username 返回当前固定管理员账号名。
func (a *Authenticator) Username() string {
	return a.username
}

func (a *Authenticator) sign(payload string) string {
	mac := hmac.New(sha256.New, a.secret)
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
