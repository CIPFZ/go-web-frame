package file

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	cos "github.com/tencentyun/cos-go-sdk-v5"
	sts "github.com/tencentyun/qcloud-cos-sts-sdk/go"

	"github.com/CIPFZ/gowebframe/internal/core/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	minCredentialTTL = 15 * time.Minute
	maxCredentialTTL = 2 * time.Hour
)

// uploadActions 覆盖简单上传与分块上传（>5GB 对象需要分块）。
var uploadActions = []string{
	"name/cos:PutObject",
	"name/cos:PostObject",
	"name/cos:InitiateMultipartUpload",
	"name/cos:ListMultipartUploads",
	"name/cos:ListParts",
	"name/cos:UploadPart",
	"name/cos:CompleteMultipartUpload",
	"name/cos:AbortMultipartUpload",
}

type COSDriver struct {
	cfg       config.COSConfig
	appID     string
	keyPrefix string
	client    *cos.Client
	sts       *sts.Client
	tracer    trace.Tracer
}

var (
	_ OSS                    = (*COSDriver)(nil)
	_ ObjectStater           = (*COSDriver)(nil)
	_ UploadCredentialIssuer = (*COSDriver)(nil)
	_ ObjectURLResolver      = (*COSDriver)(nil)
)

// NewCOSDriver 创建 COS 驱动；配置已由 config.ValidateServer 校验。
func NewCOSDriver(cfg config.COSConfig) (*COSDriver, error) {
	appID, err := cfg.AppID()
	if err != nil {
		return nil, err
	}
	bucketURL, err := url.Parse(publicBaseURL(cfg))
	if err != nil {
		return nil, fmt.Errorf("cos bucket url: %w", err)
	}
	httpClient := &http.Client{Transport: &cos.AuthorizationTransport{SecretID: cfg.SecretID, SecretKey: cfg.SecretKey}}
	return &COSDriver{
		cfg:       cfg,
		appID:     appID,
		keyPrefix: normalizePrefix(cfg.KeyPrefix),
		client:    cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, httpClient),
		sts:       sts.NewClient(cfg.SecretID, cfg.SecretKey, &http.Client{Timeout: 10 * time.Second}),
		tracer:    otel.Tracer("core.file.cos"),
	}, nil
}

// Upload 将文件写入 <key_prefix><key>。
func (d *COSDriver) Upload(ctx context.Context, file *multipart.FileHeader, key string) (string, string, error) {
	storedKey := d.keyPrefix + strings.TrimLeft(key, "/")
	ctx, span := d.startSpan(ctx, "COS.PutObject", storedKey)
	defer span.End()

	src, err := file.Open()
	if err != nil {
		recordError(span, err)
		return "", "", err
	}
	defer src.Close()

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	opt := &cos.ObjectPutOptions{ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{ContentType: contentType, ContentLength: file.Size}}
	if _, err := d.client.Object.Put(ctx, storedKey, src, opt); err != nil {
		recordError(span, err)
		return "", "", fmt.Errorf("cos put object: %w", err)
	}
	return d.PublicURL(storedKey), storedKey, nil
}

func (d *COSDriver) Delete(ctx context.Context, key string) error {
	ctx, span := d.startSpan(ctx, "COS.DeleteObject", key)
	defer span.End()
	if _, err := d.client.Object.Delete(ctx, key); err != nil {
		recordError(span, err)
		return fmt.Errorf("cos delete object: %w", err)
	}
	return nil
}

func (d *COSDriver) Stat(ctx context.Context, key string) (ObjectInfo, error) {
	ctx, span := d.startSpan(ctx, "COS.HeadObject", key)
	defer span.End()
	resp, err := d.client.Object.Head(ctx, key, nil)
	if err != nil {
		if cos.IsNotFoundError(err) {
			return ObjectInfo{}, ErrObjectNotFound
		}
		recordError(span, err)
		return ObjectInfo{}, fmt.Errorf("cos head object: %w", err)
	}
	info := ObjectInfo{Key: key, SizeBytes: resp.ContentLength, ETag: strings.Trim(resp.Header.Get("ETag"), `"`)}
	if t, err := http.ParseTime(resp.Header.Get("Last-Modified")); err == nil {
		info.LastModified = t
	}
	return info, nil
}

// IssueUploadCredential 通过 STS GetFederationToken 签发只能写入 key 的临时密钥。
// TTL 截断到 [15m, 2h]。
func (d *COSDriver) IssueUploadCredential(ctx context.Context, key string, ttl time.Duration) (UploadCredential, error) {
	key = strings.TrimLeft(key, "/")
	if key == "" || strings.ContainsAny(key, "*?") {
		return UploadCredential{}, errors.New("file: upload key must be a concrete object key")
	}
	_, span := d.startSpan(ctx, "STS.GetFederationToken", key)
	defer span.End()

	ttl = min(max(ttl, minCredentialTTL), maxCredentialTTL)
	result, err := d.sts.GetCredential(&sts.CredentialOptions{
		Region:          d.cfg.Region,
		DurationSeconds: int64(ttl / time.Second),
		Policy:          d.uploadPolicy(key),
	})
	if err != nil {
		recordError(span, err)
		return UploadCredential{}, fmt.Errorf("sts get credential: %w", err)
	}
	if result.Credentials == nil {
		err := errors.New("sts returned no credentials")
		recordError(span, err)
		return UploadCredential{}, err
	}
	return UploadCredential{
		Bucket:       d.cfg.Bucket,
		Region:       d.cfg.Region,
		Key:          key,
		TmpSecretID:  result.Credentials.TmpSecretID,
		TmpSecretKey: result.Credentials.TmpSecretKey,
		SessionToken: result.Credentials.SessionToken,
		ExpiresAt:    time.Unix(int64(result.ExpiredTime), 0).UTC(),
	}, nil
}

func (d *COSDriver) uploadPolicy(key string) *sts.CredentialPolicy {
	resource := fmt.Sprintf("qcs::cos:%s:uid/%s:%s/%s", d.cfg.Region, d.appID, d.cfg.Bucket, key)
	return &sts.CredentialPolicy{
		Version: "2.0",
		Statement: []sts.CredentialPolicyStatement{{
			Effect:   "allow",
			Action:   uploadActions,
			Resource: []string{resource},
		}},
	}
}

func (d *COSDriver) PublicURL(key string) string {
	return publicBaseURL(d.cfg) + "/" + escapeKey(key)
}

func (d *COSDriver) InternalURL(key string) string {
	return fmt.Sprintf("https://%s.cos-internal.%s.tencentcos.cn/%s", d.cfg.Bucket, d.cfg.Region, escapeKey(key))
}

func (d *COSDriver) startSpan(ctx context.Context, name, key string) (context.Context, trace.Span) {
	return d.tracer.Start(ctx, name, trace.WithAttributes(
		attribute.String("db.system", "cos"),
		attribute.String("cos.bucket", d.cfg.Bucket),
		attribute.String("cos.key", key),
	))
}

func publicBaseURL(cfg config.COSConfig) string {
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.Bucket, cfg.Region)
}

func escapeKey(key string) string {
	parts := strings.Split(strings.TrimLeft(key, "/"), "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func normalizePrefix(prefix string) string {
	prefix = strings.Trim(strings.TrimSpace(prefix), "/")
	if prefix == "" {
		return ""
	}
	return prefix + "/"
}
