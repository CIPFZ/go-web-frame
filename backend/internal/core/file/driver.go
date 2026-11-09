package file

import (
	"context"
	"errors"
	"mime/multipart"
	"time"
)

// OSS 存储经后端上传的文件（头像、附件）。
type OSS interface {
	// Upload 保存文件，返回访问 URL 和实际存储 key（Delete 使用该 key）。
	Upload(ctx context.Context, file *multipart.FileHeader, fileName string) (string, string, error)
	Delete(ctx context.Context, key string) error
}

// 下面是部分驱动才有的能力，业务用类型断言判断当前驱动是否支持，例如：
//
//	if issuer, ok := svcCtx.OSS.(file.UploadCredentialIssuer); ok { ... }

// ErrObjectNotFound 表示对象不存在。
var ErrObjectNotFound = errors.New("file: object not found")

// ObjectInfo 是对象元数据。
type ObjectInfo struct {
	Key          string
	SizeBytes    int64
	ETag         string
	LastModified time.Time
}

// ObjectStater 读取对象元数据，用于核对客户端直传的对象。
type ObjectStater interface {
	Stat(ctx context.Context, key string) (ObjectInfo, error)
}

// UploadCredential 是只允许写入单个 key 的临时凭据（客户端直传大文件，不经过后端）。
type UploadCredential struct {
	Bucket       string
	Region       string
	Key          string
	TmpSecretID  string
	TmpSecretKey string
	SessionToken string
	ExpiresAt    time.Time
}

// UploadCredentialIssuer 签发限定到单个 key 的临时上传凭据（支持分块上传）。
type UploadCredentialIssuer interface {
	IssueUploadCredential(ctx context.Context, key string, ttl time.Duration) (UploadCredential, error)
}

// ObjectURLResolver 生成对象的直连下载地址：公网与同地域内网。
type ObjectURLResolver interface {
	PublicURL(key string) string
	InternalURL(key string) string
}
