package storage

import (
	"context"
	"hash/crc64"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCloudProviderAdaptersGenerateSignedURLs(t *testing.T) {
	tests := []struct {
		name   string
		config ConnectionConfig
	}{
		{
			name: "r2",
			config: ConnectionConfig{
				ProviderCode:    ProviderCloudflareR2,
				AccountID:       "account",
				AccessKeyID:     "access",
				SecretAccessKey: "secret",
				Bucket:          "media",
				Endpoint:        "https://account.r2.cloudflarestorage.com",
			},
		},
		{
			name: "aliyun-oss",
			config: ConnectionConfig{
				ProviderCode:    ProviderAliyunOSS,
				AccessKeyID:     "access",
				SecretAccessKey: "secret",
				Bucket:          "media",
				Region:          "cn-hangzhou",
				Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
			},
		},
		{
			name: "tencent-cos",
			config: ConnectionConfig{
				ProviderCode: ProviderTencentCOS,
				AppID:        "123456",
				SecretID:     "access",
				SecretKey:    "secret",
				Bucket:       "media-123456",
				Region:       "ap-guangzhou",
				Endpoint:     "https://media-123456.cos.ap-guangzhou.myqcloud.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NormalizeConnectionConfig(tt.config)
			if err != nil {
				t.Fatal(err)
			}
			provider, err := newProviderFromConfig(config, nil)
			if err != nil {
				t.Fatal(err)
			}

			url, err := provider.CreateSignedURL(context.Background(), "images/cat.png", time.Hour)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(url, "images") || !strings.Contains(url, "cat.png") {
				t.Fatalf("signed URL = %q, want object key", url)
			}
		})
	}
}

func TestCloudProviderAdaptersExposeRealImplementations(t *testing.T) {
	for _, code := range []string{ProviderCloudflareR2, ProviderAliyunOSS, ProviderTencentCOS} {
		definition := ProviderDefinitions()[code]
		if definition == nil || !definition.AdapterAvailable {
			t.Fatalf("provider %q is not marked as available", code)
		}
	}
}

func TestCloudProviderAdaptersPerformObjectLifecycleAgainstHTTPServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodHead {
			writer.Header().Set("Content-Length", "5")
			writer.Header().Set("Content-Type", "image/png")
			writer.WriteHeader(http.StatusOK)
			return
		}
		if request.Method == http.MethodGet {
			writer.Header().Set("Content-Type", "image/png")
			_, _ = writer.Write([]byte("hello"))
			return
		}
		if request.Method == http.MethodDelete {
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		if request.Method == http.MethodPut {
			body, _ := io.ReadAll(request.Body)
			crc := crc64.Update(0, crc64.MakeTable(crc64.ECMA), body)
			writer.Header().Set("x-cos-hash-crc64ecma", strconv.FormatUint(crc, 10))
			writer.Header().Set("Content-Type", "application/xml")
			_, _ = writer.Write([]byte("<CopyObjectResult><ETag>\"test\"</ETag></CopyObjectResult>"))
			return
		}
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}))
	defer server.Close()

	configs := []ConnectionConfig{
		{ProviderCode: ProviderCloudflareR2, AccountID: "account", AccessKeyID: "access", SecretAccessKey: "secret", Bucket: "media", Endpoint: server.URL},
		{ProviderCode: ProviderAliyunOSS, AccessKeyID: "access", SecretAccessKey: "secret", Bucket: "media", Region: "cn-hangzhou", Endpoint: server.URL},
		{ProviderCode: ProviderTencentCOS, AppID: "123456", SecretID: "access", SecretKey: "secret", Bucket: "media-123456", Region: "ap-guangzhou", Endpoint: server.URL},
	}

	for _, input := range configs {
		t.Run(input.ProviderCode, func(t *testing.T) {
			config, err := NormalizeConnectionConfig(input)
			if err != nil {
				t.Fatal(err)
			}
			provider, err := newProviderFromConfig(config, server.Client())
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			if err := provider.Put(ctx, "images/cat.png", "image/png", []byte("hello")); err != nil {
				t.Fatal(err)
			}
			content, err := provider.Get(ctx, "images/cat.png")
			if err != nil || string(content) != "hello" {
				t.Fatalf("get = %q/%v", content, err)
			}
			metadata, err := provider.GetMetadata(ctx, "images/cat.png")
			if err != nil || metadata.SizeBytes != 5 || metadata.ContentType != "image/png" {
				t.Fatalf("metadata = %+v/%v", metadata, err)
			}
			exists, err := provider.Exists(ctx, "images/cat.png")
			if err != nil || !exists {
				t.Fatalf("exists = %t/%v", exists, err)
			}
			if err := provider.Copy(ctx, "images/cat.png", "images/copy.png"); err != nil {
				t.Fatal(err)
			}
			if err := provider.Delete(ctx, "images/cat.png"); err != nil {
				t.Fatal(err)
			}
		})
	}
}
