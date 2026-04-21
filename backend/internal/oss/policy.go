package oss

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bytedance/sonic"
)

// Config 定义 OSS 直传签名所需配置。
type Config struct {
	AccessKeyID     string
	AccessKeySecret string
	Bucket          string
	Region          string
	PublicBaseURL   string
	UploadDirPrefix string
	ExpireSeconds   int
}

// Policy 是前端直传 OSS 所需的短时表单凭证。
type Policy struct {
	AccessKeyID   string `json:"accessKeyId"`
	Policy        string `json:"policy"`
	Signature     string `json:"signature"`
	Host          string `json:"host"`
	Dir           string `json:"dir"`
	Bucket        string `json:"bucket"`
	Region        string `json:"region"`
	Expire        int64  `json:"expire"`
	PublicBaseURL string `json:"publicBaseUrl"`
}

// Signer 生成阿里云 OSS 直传策略。
type Signer struct {
	config Config
}

// NewSigner 创建一个 OSS 直传签名器。
func NewSigner(config Config) *Signer {
	return &Signer{config: config}
}

// CreatePolicy 生成一次短时上传策略。
func (s *Signer) CreatePolicy(_ context.Context) (Policy, error) {
	if s == nil {
		return Policy{}, errors.New("oss signer is nil")
	}
	if strings.TrimSpace(s.config.AccessKeyID) == "" ||
		strings.TrimSpace(s.config.AccessKeySecret) == "" ||
		strings.TrimSpace(s.config.Bucket) == "" ||
		strings.TrimSpace(s.config.Region) == "" {
		return Policy{}, errors.New("oss signer is not fully configured")
	}

	expireSeconds := s.config.ExpireSeconds
	if expireSeconds <= 0 {
		expireSeconds = 300
	}

	dirPrefix := strings.Trim(strings.TrimSpace(s.config.UploadDirPrefix), "/")
	if dirPrefix == "" {
		dirPrefix = "buttons"
	}

	expiresAt := time.Now().UTC().Add(time.Duration(expireSeconds) * time.Second)
	dir := fmt.Sprintf("%s/%s/", dirPrefix, time.Now().UTC().Format("20060102"))
	host := fmt.Sprintf("https://%s.oss-%s.aliyuncs.com", s.config.Bucket, s.config.Region)

	rawPolicy, err := sonic.Marshal(map[string]any{
		"expiration": expiresAt.Format(time.RFC3339),
		"conditions": []any{
			map[string]string{"bucket": s.config.Bucket},
			[]any{"starts-with", "$key", dir},
			[]any{"content-length-range", 0, 10 * 1024 * 1024},
		},
	})
	if err != nil {
		return Policy{}, err
	}

	encodedPolicy := base64.StdEncoding.EncodeToString(rawPolicy)
	mac := hmac.New(sha1.New, []byte(s.config.AccessKeySecret))
	_, _ = mac.Write([]byte(encodedPolicy))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	publicBaseURL := strings.TrimRight(strings.TrimSpace(s.config.PublicBaseURL), "/")
	if publicBaseURL == "" {
		publicBaseURL = host
	}

	return Policy{
		AccessKeyID:   s.config.AccessKeyID,
		Policy:        encodedPolicy,
		Signature:     signature,
		Host:          host,
		Dir:           dir,
		Bucket:        s.config.Bucket,
		Region:        s.config.Region,
		Expire:        expiresAt.Unix(),
		PublicBaseURL: publicBaseURL,
	}, nil
}
