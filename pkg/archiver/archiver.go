package archiver

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"runtime"
	"strings"
)

//================================================================
// 1. 公共常量和错误
//================================================================

type Format string

const (
	ZIP    Format = "zip"
	TAR    Format = "tar"     // 仅打包，不压缩
	TARZST Format = "tar.zst" // tar + zstd
	TARGZ  Format = "tar.gz"  // tar + gzip
)

var (
	ErrFormatNotSpecified  = errors.New("未指定压缩格式，且无法自动检测")
	ErrFormatNotSupported  = errors.New("不支持的压缩格式")
	ErrSourceRequired      = errors.New("必须指定输入文件或目录")
	ErrDestinationRequired = errors.New("必须指定输出文件或目录")
)

//================================================================
// 2. 进度结构
//================================================================

// Progress 封装了压缩或解压操作的当前进度
type Progress struct {
	TotalFiles     int64  // 要处理的总文件数
	FilesProcessed int64  // 已处理的文件数
	TotalBytes     int64  // 要处理的总字节数
	BytesProcessed int64  // 已处理的字节数
	CurrentFile    string // 当前正在处理的文件名
}

// ProgressFunc 是用户定义的回调函数类型，用于接收进度更新
type ProgressFunc func(p Progress)

//================================================================
// 2. Task 构建器 (核心 API)
//================================================================

type Task struct {
	isCompress   bool
	sources      []string
	destination  string
	format       Format
	cpuLimit     int
	password     string
	progressFunc ProgressFunc
	logger       *slog.Logger
	ctx          context.Context
}

// Compress 压缩任务
func Compress(sources ...string) *Task {
	return &Task{
		isCompress: true,
		sources:    sources,
		cpuLimit:   max(1, runtime.NumCPU()/2),
		logger:     slog.New(slog.NewTextHandler(io.Discard, nil)), // 默认丢弃日志
		ctx:        context.Background(),
	}
}

// Decompress 解压任务
func Decompress(source string) *Task {
	return &Task{
		isCompress: false,
		sources:    []string{source},
		cpuLimit:   max(1, runtime.NumCPU()/2),
		logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
		ctx:        context.Background(),
	}
}

func (t *Task) To(destination string) *Task {
	t.destination = destination
	return t
}

func (t *Task) WithFormat(format Format) *Task {
	t.format = format
	return t
}

func (t *Task) WithCPU(limit int) *Task {
	if limit <= 0 {
		t.cpuLimit = runtime.NumCPU()
	} else {
		t.cpuLimit = limit
	}
	return t
}

func (t *Task) WithPassword(password string) *Task {
	t.password = password
	return t
}

func (t *Task) WithLogger(logger *slog.Logger) *Task {
	if logger != nil {
		t.logger = logger
	}
	return t
}

func (t *Task) WithContext(ctx context.Context) *Task {
	if ctx != nil {
		t.ctx = ctx
	}
	return t
}

// WithProgress 注册一个进度回调函数。
// 在压缩或解压过程中，该函数会被多次调用以报告进度。
func (t *Task) WithProgress(fn ProgressFunc) *Task {
	t.progressFunc = fn
	return t
}

//================================================================
// 3. Execute 主流程
//================================================================

func (t *Task) Execute() error {
	if len(t.sources) == 0 || t.sources[0] == "" {
		return ErrSourceRequired
	}
	if t.destination == "" {
		return ErrDestinationRequired
	}
	if t.format == "" {
		filename := t.destination
		if !t.isCompress {
			filename = t.sources[0]
		}
		if f := AutodetectByFilename(filename); f != "" {
			t.format = f
		} else {
			return ErrFormatNotSpecified
		}
	}

	process, err := getProcessor(t.format) // 从 processor.go 获取
	if err != nil {
		return err
	}

	op := "解压"
	if t.isCompress {
		op = "压缩"
	}
	t.logger.Info("开始任务", "操作", op, "格式", t.format, "CPU限制", t.cpuLimit)

	if t.isCompress {
		return process.Compress(t.ctx, t.sources, t.destination, t)
	}
	return process.Decompress(t.ctx, t.sources[0], t.destination, t)
}

//================================================================
// 4. 自动检测格式
//================================================================

func AutodetectByFilename(filename string) Format {
	switch {
	case strings.HasSuffix(filename, ".zip"):
		return ZIP
	case strings.HasSuffix(filename, ".tar"):
		return TAR
	case strings.HasSuffix(filename, ".tar.zst"), strings.HasSuffix(filename, ".tzst"):
		return TARZST
	case strings.HasSuffix(filename, ".tar.gz"), strings.HasSuffix(filename, ".tgz"):
		return TARGZ
	default:
		return ""
	}
}
