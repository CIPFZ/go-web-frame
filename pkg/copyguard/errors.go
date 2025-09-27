package copyguard

import "errors"

// 公共错误，由 copyguard 库返回。
var (
	// ErrInsufficientDiskSpace 当预估的拷贝大小超过可用的磁盘预算时返回。
	ErrInsufficientDiskSpace = errors.New("copyguard: 磁盘空间不足")

	// ErrSourceIsDirectory 当源路径是一个目录而非文件时返回。
	ErrSourceIsDirectory = errors.New("copyguard: 源路径不能是目录")
)
