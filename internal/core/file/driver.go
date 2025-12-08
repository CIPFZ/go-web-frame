package file

import (
	"context" // ✨ 引入 context
	"mime/multipart"
)

type OSS interface {
	Upload(ctx context.Context, file *multipart.FileHeader, fileName string) (string, string, error)
	Delete(ctx context.Context, key string) error
}
