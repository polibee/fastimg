package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	osscredentials "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	aws "github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awscredentials "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/tencentyun/cos-go-sdk-v5"
)

type cloudProviderBase struct {
	bucket string
}

func (cloudProviderBase) CompleteMultipartUpload(_ context.Context, key string, _ []MultipartPart) error {
	if err := validateObjectKey(key); err != nil {
		return err
	}
	return ErrOperationNotSupported
}

func validateCloudPut(key, contentType string, content []byte) error {
	if err := validateObjectKey(key); err != nil {
		return err
	}
	if len(content) == 0 {
		return ErrObjectNotFound
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType == "" {
		return ErrInvalidObjectType
	}
	extensionType := mime.TypeByExtension(strings.ToLower(filepath.Ext(key)))
	if extensionType != "" {
		extensionMediaType, _, _ := mime.ParseMediaType(extensionType)
		if mediaType != extensionMediaType {
			return ErrObjectTypeMismatch
		}
	}
	return nil
}

func validateCloudKey(key string) error {
	return validateObjectKey(key)
}

func readCloudBody(body io.ReadCloser) ([]byte, error) {
	defer body.Close()
	return io.ReadAll(body)
}

func validExpiry(expiresIn time.Duration) error {
	if expiresIn <= 0 {
		return errors.New("signed URL expiry must be positive")
	}
	return nil
}

func newProviderFromConfig(config ConnectionConfig, httpClient *http.Client) (StorageProvider, error) {
	switch config.ProviderCode {
	case ProviderCloudflareR2:
		return newR2Provider(config, httpClient)
	case ProviderAliyunOSS:
		return newAliyunOSSProvider(config, httpClient)
	case ProviderTencentCOS:
		return newTencentCOSProvider(config, httpClient)
	default:
		return nil, ErrUnsupportedProvider
	}
}

type r2Provider struct {
	cloudProviderBase
	client  *s3.Client
	presign *s3.PresignClient
}

func newR2Provider(config ConnectionConfig, httpClient *http.Client) (StorageProvider, error) {
	options := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithCredentialsProvider(awscredentials.NewStaticCredentialsProvider(config.AccessKeyID, config.SecretAccessKey, "")),
		awsconfig.WithRegion(config.Region),
	}
	if httpClient != nil {
		options = append(options, awsconfig.WithHTTPClient(httpClient))
	}
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), options...)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(config.Endpoint)
	})
	return &r2Provider{cloudProviderBase: cloudProviderBase{bucket: config.Bucket}, client: client, presign: s3.NewPresignClient(client)}, nil
}

func (p *r2Provider) Put(ctx context.Context, key, contentType string, content []byte) error {
	if err := validateCloudPut(key, contentType, content); err != nil {
		return err
	}
	length := int64(len(content))
	_, err := p.client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(key), Body: bytes.NewReader(content), ContentType: aws.String(contentType), ContentLength: &length})
	return err
}

func (p *r2Provider) Get(ctx context.Context, key string) ([]byte, error) {
	if err := validateCloudKey(key); err != nil {
		return nil, err
	}
	result, err := p.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(key)})
	if err != nil {
		if cloudNotFound(err) {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}
	return readCloudBody(result.Body)
}

func (p *r2Provider) GetMetadata(ctx context.Context, key string) (ObjectMetadata, error) {
	if err := validateCloudKey(key); err != nil {
		return ObjectMetadata{}, err
	}
	result, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(key)})
	if err != nil {
		if cloudNotFound(err) {
			return ObjectMetadata{}, ErrObjectNotFound
		}
		return ObjectMetadata{}, err
	}
	return ObjectMetadata{SizeBytes: aws.ToInt64(result.ContentLength), ContentType: aws.ToString(result.ContentType)}, nil
}

func (p *r2Provider) Delete(ctx context.Context, key string) error {
	if err := validateCloudKey(key); err != nil {
		return err
	}
	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(key)})
	if cloudNotFound(err) {
		return nil
	}
	return err
}

func (p *r2Provider) CreateSignedURL(ctx context.Context, key string, expiresIn time.Duration) (string, error) {
	if err := validateCloudKey(key); err != nil {
		return "", err
	}
	if err := validExpiry(expiresIn); err != nil {
		return "", err
	}
	result, err := p.presign.PresignGetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(key)}, s3.WithPresignExpires(expiresIn))
	if err != nil {
		return "", err
	}
	return result.URL, nil
}

func (p *r2Provider) Copy(ctx context.Context, from, to string) error {
	if err := validateCloudKey(from); err != nil {
		return err
	}
	if err := validateCloudKey(to); err != nil {
		return err
	}
	source := url.PathEscape(p.bucket + "/" + from)
	_, err := p.client.CopyObject(ctx, &s3.CopyObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(to), CopySource: aws.String(source)})
	return err
}

func (p *r2Provider) Exists(ctx context.Context, key string) (bool, error) {
	if err := validateCloudKey(key); err != nil {
		return false, err
	}
	_, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(key)})
	if cloudNotFound(err) {
		return false, nil
	}
	return err == nil, err
}

type aliyunOSSProvider struct {
	cloudProviderBase
	client *oss.Client
}

func newAliyunOSSProvider(config ConnectionConfig, httpClient *http.Client) (StorageProvider, error) {
	providerConfig := oss.LoadDefaultConfig().WithRegion(config.Region).WithEndpoint(config.Endpoint).WithCredentialsProvider(osscredentials.NewStaticCredentialsProvider(config.AccessKeyID, config.SecretAccessKey))
	if !strings.Contains(config.Endpoint, "aliyuncs.com") {
		providerConfig = providerConfig.WithUsePathStyle(true)
	}
	if httpClient != nil {
		providerConfig = providerConfig.WithHttpClient(httpClient)
	}
	return &aliyunOSSProvider{cloudProviderBase: cloudProviderBase{bucket: config.Bucket}, client: oss.NewClient(providerConfig)}, nil
}

func (p *aliyunOSSProvider) Put(ctx context.Context, key, contentType string, content []byte) error {
	if err := validateCloudPut(key, contentType, content); err != nil {
		return err
	}
	length := int64(len(content))
	_, err := p.client.PutObject(ctx, &oss.PutObjectRequest{Bucket: oss.Ptr(p.bucket), Key: oss.Ptr(key), Body: bytes.NewReader(content), ContentType: oss.Ptr(contentType), ContentLength: &length})
	return err
}

func (p *aliyunOSSProvider) Get(ctx context.Context, key string) ([]byte, error) {
	if err := validateCloudKey(key); err != nil {
		return nil, err
	}
	result, err := p.client.GetObject(ctx, &oss.GetObjectRequest{Bucket: oss.Ptr(p.bucket), Key: oss.Ptr(key)})
	if err != nil {
		if cloudNotFound(err) {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}
	return readCloudBody(result.Body)
}

func (p *aliyunOSSProvider) GetMetadata(ctx context.Context, key string) (ObjectMetadata, error) {
	if err := validateCloudKey(key); err != nil {
		return ObjectMetadata{}, err
	}
	result, err := p.client.HeadObject(ctx, &oss.HeadObjectRequest{Bucket: oss.Ptr(p.bucket), Key: oss.Ptr(key)})
	if err != nil {
		if cloudNotFound(err) {
			return ObjectMetadata{}, ErrObjectNotFound
		}
		return ObjectMetadata{}, err
	}
	return ObjectMetadata{SizeBytes: result.ContentLength, ContentType: oss.ToString(result.ContentType)}, nil
}

func (p *aliyunOSSProvider) Delete(ctx context.Context, key string) error {
	if err := validateCloudKey(key); err != nil {
		return err
	}
	_, err := p.client.DeleteObject(ctx, &oss.DeleteObjectRequest{Bucket: oss.Ptr(p.bucket), Key: oss.Ptr(key)})
	if cloudNotFound(err) {
		return nil
	}
	return err
}

func (p *aliyunOSSProvider) CreateSignedURL(ctx context.Context, key string, expiresIn time.Duration) (string, error) {
	if err := validateCloudKey(key); err != nil {
		return "", err
	}
	if err := validExpiry(expiresIn); err != nil {
		return "", err
	}
	result, err := p.client.Presign(ctx, &oss.GetObjectRequest{Bucket: oss.Ptr(p.bucket), Key: oss.Ptr(key)}, oss.PresignExpires(expiresIn))
	if err != nil {
		return "", err
	}
	return result.URL, nil
}

func (p *aliyunOSSProvider) Copy(ctx context.Context, from, to string) error {
	if err := validateCloudKey(from); err != nil {
		return err
	}
	if err := validateCloudKey(to); err != nil {
		return err
	}
	_, err := p.client.CopyObject(ctx, &oss.CopyObjectRequest{Bucket: oss.Ptr(p.bucket), Key: oss.Ptr(to), SourceBucket: oss.Ptr(p.bucket), SourceKey: oss.Ptr(from)})
	return err
}

func (p *aliyunOSSProvider) Exists(ctx context.Context, key string) (bool, error) {
	if err := validateCloudKey(key); err != nil {
		return false, err
	}
	_, err := p.client.HeadObject(ctx, &oss.HeadObjectRequest{Bucket: oss.Ptr(p.bucket), Key: oss.Ptr(key)})
	if cloudNotFound(err) {
		return false, nil
	}
	return err == nil, err
}

type tencentCOSProvider struct {
	cloudProviderBase
	client *cos.Client
	config ConnectionConfig
}

func newTencentCOSProvider(config ConnectionConfig, httpClient *http.Client) (StorageProvider, error) {
	endpoint, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, err
	}
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	transport := &cos.AuthorizationTransport{SecretID: config.SecretID, SecretKey: config.SecretKey, Transport: httpClient.Transport}
	clientHTTP := *httpClient
	clientHTTP.Transport = transport
	client := cos.NewClient(&cos.BaseURL{BucketURL: endpoint}, &clientHTTP)
	return &tencentCOSProvider{cloudProviderBase: cloudProviderBase{bucket: config.Bucket}, client: client, config: config}, nil
}

func (p *tencentCOSProvider) Put(ctx context.Context, key, contentType string, content []byte) error {
	if err := validateCloudPut(key, contentType, content); err != nil {
		return err
	}
	_, err := p.client.Object.Put(ctx, key, bytes.NewReader(content), &cos.ObjectPutOptions{ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{ContentType: contentType, ContentLength: int64(len(content))}})
	return err
}

func (p *tencentCOSProvider) Get(ctx context.Context, key string) ([]byte, error) {
	if err := validateCloudKey(key); err != nil {
		return nil, err
	}
	result, err := p.client.Object.Get(ctx, key, nil)
	if err != nil {
		if cloudNotFound(err) {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}
	return readCloudBody(result.Body)
}

func (p *tencentCOSProvider) GetMetadata(ctx context.Context, key string) (ObjectMetadata, error) {
	if err := validateCloudKey(key); err != nil {
		return ObjectMetadata{}, err
	}
	result, err := p.client.Object.Head(ctx, key, nil)
	if err != nil {
		if cloudNotFound(err) {
			return ObjectMetadata{}, ErrObjectNotFound
		}
		return ObjectMetadata{}, err
	}
	return ObjectMetadata{SizeBytes: result.ContentLength, ContentType: result.Header.Get("Content-Type")}, nil
}

func (p *tencentCOSProvider) Delete(ctx context.Context, key string) error {
	if err := validateCloudKey(key); err != nil {
		return err
	}
	_, err := p.client.Object.Delete(ctx, key)
	if cloudNotFound(err) {
		return nil
	}
	return err
}

func (p *tencentCOSProvider) CreateSignedURL(ctx context.Context, key string, expiresIn time.Duration) (string, error) {
	if err := validateCloudKey(key); err != nil {
		return "", err
	}
	if err := validExpiry(expiresIn); err != nil {
		return "", err
	}
	result, err := p.client.Object.GetPresignedURL(ctx, http.MethodGet, key, p.config.SecretID, p.config.SecretKey, expiresIn, nil, true)
	if err != nil {
		return "", err
	}
	return result.String(), nil
}

func (p *tencentCOSProvider) Copy(ctx context.Context, from, to string) error {
	if err := validateCloudKey(from); err != nil {
		return err
	}
	if err := validateCloudKey(to); err != nil {
		return err
	}
	_, _, err := p.client.Object.Copy(ctx, to, p.bucket+"/"+from, nil)
	return err
}

func (p *tencentCOSProvider) Exists(ctx context.Context, key string) (bool, error) {
	if err := validateCloudKey(key); err != nil {
		return false, err
	}
	return p.client.Object.IsExist(ctx, key)
}

func cloudNotFound(err error) bool {
	if err == nil {
		return false
	}
	var apiError *smithy.GenericAPIError
	if errors.As(err, &apiError) {
		code := strings.ToLower(apiError.ErrorCode())
		if code == "notfound" || code == "nosuchkey" || code == "nosuchobject" || code == "nosuchbucket" {
			return true
		}
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "status code: 404") || strings.Contains(message, "not found") || strings.Contains(message, "nosuchkey")
}

var _ StorageProvider = (*r2Provider)(nil)
var _ StorageProvider = (*aliyunOSSProvider)(nil)
var _ StorageProvider = (*tencentCOSProvider)(nil)
