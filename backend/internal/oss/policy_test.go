package oss

import (
	"context"
	"strings"
	"testing"
)

func TestSignerCreatePolicyBuildsUploadPayload(t *testing.T) {
	signer := NewSigner(Config{
		AccessKeyID:     "test-ak",
		AccessKeySecret: "test-secret",
		Bucket:          "vote-wall",
		Region:          "cn-beijing",
		PublicBaseURL:   "https://cdn.example.com",
		UploadDirPrefix: "buttons",
		ExpireSeconds:   300,
	})

	policy, err := signer.CreatePolicy(context.Background())
	if err != nil {
		t.Fatalf("create policy: %v", err)
	}

	if policy.AccessKeyID != "test-ak" {
		t.Fatalf("expected access key id test-ak, got %q", policy.AccessKeyID)
	}
	if !strings.Contains(policy.Host, "vote-wall.oss-cn-beijing.aliyuncs.com") {
		t.Fatalf("unexpected host: %q", policy.Host)
	}
	if policy.Dir == "" || !strings.HasPrefix(policy.Dir, "buttons/") {
		t.Fatalf("expected upload dir under buttons/, got %q", policy.Dir)
	}
	if policy.Policy == "" || policy.Signature == "" {
		t.Fatalf("expected policy and signature, got %+v", policy)
	}
	if policy.PublicBaseURL != "https://cdn.example.com" {
		t.Fatalf("unexpected public base url: %q", policy.PublicBaseURL)
	}
}
