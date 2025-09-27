package archiver

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/klauspost/compress/zstd"
	"github.com/yeka/zip"
)

//================================================================
// 1. 内部接口和注册表 (策略模式)
//================================================================

// processor defines the interface for a specific archive format.
// The `opts` parameter is the original Task object, allowing access to all settings.
type processor interface {
	Compress(ctx context.Context, sources []string, destination string, opts *Task) error
	Decompress(ctx context.Context, source string, destination string, opts *Task) error
}

// processors holds the registered implementations.
var processors = make(map[Format]processor)

// Go 的 init 函数会在包首次被使用时执行，非常适合用来注册实现
func init() {
	processors[ZIP] = &zipProcessor{}
	processors[TARZST] = &tarZstProcessor{}
	processors[TARGZ] = &tarGzProcessor{}
	// 在此注册更多格式...
}

// getProcessor is a factory function to retrieve the correct processor.
func getProcessor(f Format) (processor, error) {
	p, ok := processors[f]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrFormatNotSupported, f)
	}
	return p, nil
}

// preflightScan 预扫描源路径以计算总文件数和总大小
func preflightScan(sources []string) (totalFiles int64, totalBytes int64, err error) {
	for _, source := range sources {
		walkErr := filepath.WalkDir(source, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.Type().IsRegular() {
				info, err := d.Info()
				if err != nil {
					return err
				}
				atomic.AddInt64(&totalFiles, 1)
				atomic.AddInt64(&totalBytes, info.Size())
			}
			return nil
		})
		if walkErr != nil {
			return 0, 0, walkErr
		}
	}
	return
}

//================================================================
// 2. ZIP 格式的具体实现
//================================================================

type zipProcessor struct{}

func (z *zipProcessor) Compress(ctx context.Context, sources []string, destination string, opts *Task) error {
	opts.logger.Debug("开始 ZIP 压缩任务", "目标文件", destination)

	// --- 进度回调：预扫描 ---
	var totalFiles, totalBytes int64
	var filesProcessed, bytesProcessed atomic.Int64
	if opts.progressFunc != nil {
		var err error
		totalFiles, totalBytes, err = preflightScan(sources)
		if err != nil {
			return fmt.Errorf("预扫描源文件失败: %w", err)
		}
	}
	// --- 结束 ---

	// 1. 创建目标压缩文件
	zipFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("创建 zip 文件失败: %w", err)
	}
	defer func(f *os.File) {
		_ = zipFile.Close()
	}(zipFile)

	zipWriter := zip.NewWriter(zipFile)
	defer func(w *zip.Writer) {
		_ = w.Close()
	}(zipWriter)

	// 2. 初始化工作池和并发控制工具
	pool := NewPool(opts.cpuLimit)
	pool.Start()
	defer pool.Stop()

	var writerMutex sync.Mutex
	// 将 errChan 的容量设置为 totalFiles，如果 totalFiles 为 0，则至少为 1
	errChanSize := totalFiles
	if errChanSize == 0 {
		errChanSize = 1
	}
	errChan := make(chan error, errChanSize)

	// 3. 遍历所有源路径（文件或目录）
	for _, source := range sources {
		// 检查上下文是否已被取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		source := source // 捕获循环变量
		baseDir := filepath.Dir(source)
		if fi, err := os.Stat(source); err == nil && !fi.IsDir() {
			// 如果源是一个文件，则其基准目录就是它所在的目录
		} else {
			// 如果源是一个目录，则其本身就是基准目录
			baseDir = source
		}

		// 4. 遍历目录和文件
		walkErr := filepath.WalkDir(source, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// 检查上下文
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// 只处理普通文件
			if !d.Type().IsRegular() {
				return nil
			}
			// 获取文件信息
			fileInfo, err := d.Info()
			if err != nil {
				return err
			}

			// 5. 将文件压缩任务提交到工作池
			pool.Submit(func() {
				// 计算文件在 zip 中的相对路径
				relPath, err := filepath.Rel(baseDir, path)
				if err != nil {
					errChan <- fmt.Errorf("计算相对路径失败 '%s': %w", path, err)
					return
				}
				// 统一路径分隔符为 '/'
				relPath = filepath.ToSlash(relPath)

				opts.logger.Debug("正在压缩文件", "路径", relPath)

				// --- 进度回调：执行并更新 ---
				if opts.progressFunc != nil {
					opts.progressFunc(Progress{
						TotalFiles:     totalFiles,
						FilesProcessed: filesProcessed.Load(),
						TotalBytes:     totalBytes,
						BytesProcessed: bytesProcessed.Load(),
						CurrentFile:    relPath,
					})
				}

				// 打开源文件
				file, err := os.Open(path)
				if err != nil {
					errChan <- fmt.Errorf("打开文件失败 '%s': %w", path, err)
					return
				}
				defer func(f *os.File) {
					_ = f.Close()
				}(file)

				// 加锁以安全地写入 zip 文件
				writerMutex.Lock()
				defer writerMutex.Unlock()

				var zipEntryWriter io.Writer
				if opts.password != "" {
					// 使用密码创建加密的条目
					zipEntryWriter, err = zipWriter.Encrypt(relPath, opts.password, zip.AES256Encryption)
				} else {
					// 创建普通的条目
					zipEntryWriter, err = zipWriter.Create(relPath)
				}
				if err != nil {
					errChan <- fmt.Errorf("在 zip 中创建条目失败 '%s': %w", relPath, err)
					return
				}

				// 使用流式拷贝，将文件内容写入 zip
				_, err = io.Copy(zipEntryWriter, file)
				if err != nil {
					errChan <- fmt.Errorf("拷贝文件内容失败 '%s': %w", relPath, err)
				}
				// --- 进度回调：在成功后更新原子计数器 ---
				if opts.progressFunc != nil {
					filesProcessed.Add(1)
					bytesProcessed.Add(fileInfo.Size())
				}
			})
			return nil
		})

		if walkErr != nil {
			return fmt.Errorf("遍历源路径 '%s' 失败: %w", source, walkErr)
		}
	}

	// 6. 等待所有压缩任务完成
	pool.Wait()
	close(errChan)

	// 7. 检查是否有任何 goroutine 发生了错误
	for err := range errChan {
		return err // 返回第一个遇到的错误
	}

	opts.logger.Info("ZIP 压缩任务成功完成")
	return nil
}

func (z *zipProcessor) Decompress(ctx context.Context, source string, destination string, opts *Task) error {
	opts.logger.Debug("开始 ZIP 解压任务", "源文件", source)

	// 1. 打开 zip 归档文件
	reader, err := zip.OpenReader(source)
	if err != nil {
		return fmt.Errorf("打开 zip 文件失败: %w", err)
	}
	defer func(r *zip.ReadCloser) {
		_ = r.Close()
	}(reader)

	// --- 进度回调：预扫描 ---
	var totalFiles int64
	var totalBytes int64
	for _, f := range reader.File {
		if !f.FileInfo().IsDir() {
			totalFiles++
			totalBytes += int64(f.UncompressedSize64)
		}
	}
	var filesProcessed, bytesProcessed int64

	// 2. 确保目标目录存在
	if err := os.MkdirAll(destination, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 3. 遍历 zip 中的每一个文件
	for _, file := range reader.File {
		// 检查上下文
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 4. 构建解压后的完整路径，并进行安全检查 (!!!关键!!!)
		destPath, err := secureJoin(destination, file.Name)
		if err != nil {
			return err // 返回路径不安全的错误
		}
		opts.logger.Debug("正在解压文件", "路径", destPath)

		// 5. 如果是目录，则创建它
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, file.Mode()); err != nil {
				return fmt.Errorf("创建解压目录失败 '%s': %w", destPath, err)
			}
			continue
		}

		// --- 进度回调：执行并更新 ---
		if opts.progressFunc != nil {
			opts.progressFunc(Progress{
				TotalFiles:     totalFiles,
				FilesProcessed: filesProcessed,
				TotalBytes:     totalBytes,
				BytesProcessed: bytesProcessed,
				CurrentFile:    file.Name,
			})
		}

		// 6. 如果是文件，创建其所在的目录
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("创建父目录失败 '%s': %w", filepath.Dir(destPath), err)
		}

		// 7. 打开 zip 中的文件条目
		if file.IsEncrypted() {
			if opts.password == "" {
				return fmt.Errorf("文件 '%s' 已加密, 但未提供密码", file.Name)
			}
			file.SetPassword(opts.password)
		}
		srcFile, err := file.Open()
		if err != nil {
			// 检查是否是密码相关的错误
			if errors.Is(err, zip.ErrPassword) || strings.Contains(err.Error(), "password") {
				return fmt.Errorf("解密 '%s' 失败: 密码错误或加密算法不支持", file.Name)
			}
			return fmt.Errorf("打开 zip 内文件失败 '%s': %w", file.Name, err)
		}
		// 创建目标文件
		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			_ = srcFile.Close()
			return fmt.Errorf("创建目标文件失败 '%s': %w", destPath, err)
		}

		// 8. 流式拷贝数据
		_, err = io.Copy(destFile, srcFile)

		// 确保文件都已关闭
		_ = srcFile.Close()
		_ = destFile.Close()

		if err != nil {
			return fmt.Errorf("解压文件内容失败 '%s': %w", file.Name, err)
		}
		// --- 进度回调：在成功后更新计数器 ---
		if opts.progressFunc != nil {
			filesProcessed++
			bytesProcessed += int64(file.UncompressedSize64)
		}
	}

	// --- 进度回调：触发最后一次完成状态的回调 ---
	if opts.progressFunc != nil {
		opts.progressFunc(Progress{
			TotalFiles:     totalFiles,
			FilesProcessed: filesProcessed,
			TotalBytes:     totalBytes,
			BytesProcessed: bytesProcessed,
			CurrentFile:    "", // 操作完成，没有当前文件
		})
	}

	opts.logger.Info("ZIP 解压任务成功完成")
	return nil
}

// secureJoin 用于安全地拼接路径，防止路径遍历攻击 (Zip Slip)
func secureJoin(base, target string) (string, error) {
	// 将目标路径清理干净
	cleanTarget := filepath.Clean(target)

	// 拼接绝对路径
	destPath := filepath.Join(base, cleanTarget)

	// 检查最终路径是否仍然在基准目录之内
	if !strings.HasPrefix(destPath, filepath.Clean(base)) {
		return "", fmt.Errorf("检测到不安全的路径: %s", target)
	}
	return destPath, nil
}

//================================================================
// 3. TAR.ZST 格式的具体实现
//================================================================

type tarZstProcessor struct{}

func (tz *tarZstProcessor) Compress(ctx context.Context, sources []string, destination string, opts *Task) error {
	opts.logger.Debug("开始 TAR.ZST 压缩任务", "目标文件", destination)

	// --- 进度回调：预扫描 ---
	var totalFiles, totalBytes int64
	var filesProcessed, bytesProcessed atomic.Int64
	if opts.progressFunc != nil {
		var err error
		totalFiles, totalBytes, err = preflightScan(sources)
		if err != nil {
			return fmt.Errorf("预扫描源文件失败: %w", err)
		}
	}

	// 1. 创建 I/O 链: os.File -> zstd.Writer -> tar.Writer
	destFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(destFile)

	var topWriter io.WriteCloser = destFile // 顶层写入器
	if opts.password != "" {
		opts.logger.Debug("启用流加密")
		// 如果有密码，在文件写入器之上包装一层加密写入器
		encryptedWriter, err := encryptStream(destFile, opts.password)
		if err != nil {
			return err
		}
		topWriter = encryptedWriter
		defer func(w io.WriteCloser) {
			_ = w.Close()
		}(topWriter) // 确保加密写入器被关闭以写入认证标签
	}

	zstdWriter, err := zstd.NewWriter(topWriter)
	if err != nil {
		return fmt.Errorf("创建 zstd写入器失败: %w", err)
	}
	defer func(w *zstd.Encoder) {
		_ = w.Close()
	}(zstdWriter)

	tarWriter := tar.NewWriter(zstdWriter)
	defer func(w *tar.Writer) {
		_ = w.Close()
	}(tarWriter)

	// 2. 初始化工作池 (逻辑与 zip 类似)
	pool := NewPool(opts.cpuLimit)
	pool.Start()
	defer pool.Stop()

	var writerMutex sync.Mutex
	errChanSize := totalFiles
	if errChanSize == 0 {
		errChanSize = 1
	}
	errChan := make(chan error, errChanSize)

	// 3. 遍历源路径
	for _, source := range sources {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		source := source
		baseDir := filepath.Dir(source)
		if fi, err := os.Stat(source); err == nil && !fi.IsDir() {
			baseDir = filepath.Dir(source)
		} else {
			baseDir = source
		}

		// 4. 遍历文件树
		walkErr := filepath.WalkDir(source, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// tar 需要包含目录信息
			if d.IsDir() {
				return nil
			}

			// 5. 提交压缩任务到工作池
			pool.Submit(func() {
				// 计算相对路径
				relPath, err := filepath.Rel(baseDir, path)
				if err != nil {
					errChan <- fmt.Errorf("计算相对路径失败 '%s': %w", path, err)
					return
				}
				relPath = filepath.ToSlash(relPath)
				opts.logger.Debug("正在归档并压缩文件", "路径", relPath)

				if opts.progressFunc != nil {
					opts.progressFunc(Progress{
						TotalFiles:     totalFiles,
						FilesProcessed: filesProcessed.Load(),
						TotalBytes:     totalBytes,
						BytesProcessed: bytesProcessed.Load(),
						CurrentFile:    relPath,
					})
				}

				// 获取文件信息以创建 tar 头
				fileInfo, err := os.Stat(path)
				if err != nil {
					errChan <- fmt.Errorf("获取文件信息失败 '%s': %w", path, err)
					return
				}
				hdr, err := tar.FileInfoHeader(fileInfo, "")
				if err != nil {
					errChan <- fmt.Errorf("创建 tar头失败 '%s': %w", path, err)
					return
				}
				hdr.Name = relPath

				// 打开源文件
				file, err := os.Open(path)
				if err != nil {
					errChan <- fmt.Errorf("打开文件失败 '%s': %w", path, err)
					return
				}
				defer func(f *os.File) {
					_ = f.Close()
				}(file)

				// 加锁写入 tar
				writerMutex.Lock()
				defer writerMutex.Unlock()

				if err := tarWriter.WriteHeader(hdr); err != nil {
					errChan <- fmt.Errorf("写入 tar头失败 '%s': %w", relPath, err)
					return
				}
				if _, err := io.Copy(tarWriter, file); err != nil {
					errChan <- fmt.Errorf("拷贝文件内容失败 '%s': %w", relPath, err)
				}
				if opts.progressFunc != nil {
					filesProcessed.Add(1)
					bytesProcessed.Add(fileInfo.Size())
				}
			})
			return nil
		})

		if walkErr != nil {
			return fmt.Errorf("遍历源路径 '%s' 失败: %w", source, walkErr)
		}
	}

	// 6. 等待并检查错误
	pool.Wait()
	close(errChan)

	for err := range errChan {
		return err
	}

	if opts.progressFunc != nil {
		opts.progressFunc(Progress{
			TotalFiles:     totalFiles,
			FilesProcessed: filesProcessed.Load(),
			TotalBytes:     totalBytes,
			BytesProcessed: bytesProcessed.Load(),
			CurrentFile:    "",
		})
	}

	opts.logger.Info("TAR.ZST 压缩任务成功完成")
	return nil
}

func (tz *tarZstProcessor) Decompress(ctx context.Context, source string, destination string, opts *Task) error {
	opts.logger.Debug("开始 TAR.ZST 解压任务", "源文件", source)

	// 1. 创建 I/O 链: os.File -> zstd.Reader -> tar.Reader
	sourceFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(sourceFile)

	var topReader io.Reader = sourceFile // 顶层读取器
	if opts.password != "" {
		opts.logger.Debug("启用流解密")
		// 如果有密码，在文件读取器之上包装一层解密读取器
		decryptedReader, err := decryptStream(sourceFile, opts.password)
		if err != nil {
			return err
		}
		topReader = decryptedReader
	}

	zstdReader, err := zstd.NewReader(topReader)
	if err != nil {
		return fmt.Errorf("创建 zstd读取器失败: %w", err)
	}
	defer zstdReader.Close()

	tarReader := tar.NewReader(zstdReader)

	// 2. 确保目标目录存在
	if err := os.MkdirAll(destination, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	var filesProcessed, bytesProcessed int64
	// 3. 循环读取 tar 流中的文件
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break // 读取完毕
		}
		if err != nil {
			return fmt.Errorf("读取 tar流失败: %w", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 4. 安全构建目标路径
		destPath, err := secureJoin(destination, hdr.Name)
		if err != nil {
			return err
		}
		opts.logger.Debug("正在解压文件", "路径", destPath)

		if opts.progressFunc != nil {
			opts.progressFunc(Progress{
				TotalFiles:     -1, // -1 表示未知
				FilesProcessed: filesProcessed,
				TotalBytes:     -1, // -1 表示未知
				BytesProcessed: bytesProcessed,
				CurrentFile:    hdr.Name,
			})
		}

		// 5. 根据头信息类型处理文件或目录
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(destPath, os.FileMode(hdr.Mode)); err != nil {
				return fmt.Errorf("创建解压目录失败 '%s': %w", destPath, err)
			}
		case tar.TypeReg:
			// 创建父目录
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("创建父目录失败 '%s': %w", filepath.Dir(destPath), err)
			}
			// 创建并写入文件
			destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return fmt.Errorf("创建目标文件失败 '%s': %w", destPath, err)
			}
			if _, err := io.Copy(destFile, tarReader); err != nil {
				_ = destFile.Close()
				return fmt.Errorf("解压文件内容失败 '%s': %w", hdr.Name, err)
			}
			_ = destFile.Close()

			if opts.progressFunc != nil {
				filesProcessed++
				bytesProcessed += hdr.Size
			}
		default:
			opts.logger.Warn("跳过不支持的 tar条目类型", "类型", hdr.Typeflag, "名称", hdr.Name)
		}
	}

	if opts.progressFunc != nil {
		opts.progressFunc(Progress{
			TotalFiles:     -1,
			FilesProcessed: filesProcessed,
			TotalBytes:     -1,
			BytesProcessed: bytesProcessed,
			CurrentFile:    "",
		})
	}

	opts.logger.Info("TAR.ZST 解压任务成功完成")
	return nil
}

//================================================================
// 4. TAR.GZ 格式的具体实现
//================================================================

//================================================================
// 4. TAR.GZ 格式的具体实现 (完整版)
//================================================================

type tarGzProcessor struct{}

func (tz *tarGzProcessor) Compress(ctx context.Context, sources []string, destination string, opts *Task) error {
	opts.logger.Debug("开始 TAR.GZ 压缩任务", "目标文件", destination)

	// --- 进度回调：预扫描 ---
	var totalFiles, totalBytes int64
	var filesProcessed, bytesProcessed atomic.Int64
	if opts.progressFunc != nil {
		var err error
		totalFiles, totalBytes, err = preflightScan(sources)
		if err != nil {
			return fmt.Errorf("预扫描源文件失败: %w", err)
		}
	}
	// --- 结束 ---

	destFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(destFile)

	// 1. 创建 I/O 链: os.File -> gzip.Writer -> tar.Writer
	gzWriter := gzip.NewWriter(destFile)
	defer func(w *gzip.Writer) {
		_ = w.Close()
	}(gzWriter)

	tarWriter := tar.NewWriter(gzWriter)
	defer func(w *tar.Writer) {
		_ = w.Close()
	}(tarWriter)

	// 2. 初始化工作池
	pool := NewPool(opts.cpuLimit)
	pool.Start()
	defer pool.Stop()

	var writerMutex sync.Mutex
	errChanSize := totalFiles
	if errChanSize == 0 {
		errChanSize = 1
	}
	errChan := make(chan error, errChanSize)

	// 3. 遍历源路径
	for _, source := range sources {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		source := source
		baseDir := filepath.Dir(source)
		if fi, err := os.Stat(source); err == nil && !fi.IsDir() {
			baseDir = filepath.Dir(source)
		} else {
			baseDir = source
		}

		// 4. 遍历文件树
		walkErr := filepath.WalkDir(source, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			if !d.Type().IsRegular() {
				return nil
			}
			fileInfo, err := d.Info()
			if err != nil {
				return err
			}

			// 5. 提交任务到工作池
			pool.Submit(func() {
				relPath, err := filepath.Rel(baseDir, path)
				if err != nil {
					errChan <- fmt.Errorf("计算相对路径失败 '%s': %w", path, err)
					return
				}
				relPath = filepath.ToSlash(relPath)
				opts.logger.Debug("正在归档并压缩文件", "路径", relPath)

				if opts.progressFunc != nil {
					opts.progressFunc(Progress{
						TotalFiles:     totalFiles,
						FilesProcessed: filesProcessed.Load(),
						TotalBytes:     totalBytes,
						BytesProcessed: bytesProcessed.Load(),
						CurrentFile:    relPath,
					})
				}

				hdr, err := tar.FileInfoHeader(fileInfo, "")
				if err != nil {
					errChan <- fmt.Errorf("创建 tar头失败 '%s': %w", path, err)
					return
				}
				hdr.Name = relPath

				file, err := os.Open(path)
				if err != nil {
					errChan <- fmt.Errorf("打开文件失败 '%s': %w", path, err)
					return
				}
				defer func(f *os.File) {
					_ = f.Close()
				}(file)

				writerMutex.Lock()
				defer writerMutex.Unlock()

				if err := tarWriter.WriteHeader(hdr); err != nil {
					errChan <- fmt.Errorf("写入 tar头失败 '%s': %w", relPath, err)
					return
				}
				if _, err := io.Copy(tarWriter, file); err != nil {
					errChan <- fmt.Errorf("拷贝文件内容失败 '%s': %w", relPath, err)
					return
				}

				if opts.progressFunc != nil {
					filesProcessed.Add(1)
					bytesProcessed.Add(fileInfo.Size())
				}
			})
			return nil
		})

		if walkErr != nil {
			return fmt.Errorf("遍历源路径 '%s' 失败: %w", source, walkErr)
		}
	}

	// 6. 等待并检查错误
	pool.Wait()
	close(errChan)
	for err := range errChan {
		return err
	}

	// --- 进度回调：触发最后一次完成状态的回调 ---
	if opts.progressFunc != nil {
		opts.progressFunc(Progress{
			TotalFiles:     totalFiles,
			FilesProcessed: filesProcessed.Load(),
			TotalBytes:     totalBytes,
			BytesProcessed: bytesProcessed.Load(),
			CurrentFile:    "",
		})
	}
	// --- 结束 ---

	opts.logger.Info("TAR.GZ 压缩任务成功完成")
	return nil
}

func (tz *tarGzProcessor) Decompress(ctx context.Context, source string, destination string, opts *Task) error {
	opts.logger.Debug("开始 TAR.GZ 解压任务", "源文件", source)

	sourceFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(sourceFile)

	// 1. 创建 I/O 链: os.File -> gzip.Reader -> tar.Reader
	gzReader, err := gzip.NewReader(sourceFile)
	if err != nil {
		return fmt.Errorf("创建 gzip读取器失败: %w", err)
	}
	defer func(r *gzip.Reader) {
		_ = r.Close()
	}(gzReader)

	tarReader := tar.NewReader(gzReader)

	if err := os.MkdirAll(destination, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	var filesProcessed, bytesProcessed int64

	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取 tar流失败: %w", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		destPath, err := secureJoin(destination, hdr.Name)
		if err != nil {
			return err
		}
		opts.logger.Debug("正在解压文件", "路径", destPath)

		if opts.progressFunc != nil {
			opts.progressFunc(Progress{
				TotalFiles:     -1, // tar 流格式总数未知
				FilesProcessed: filesProcessed,
				TotalBytes:     -1, // tar 流格式总字节数未知
				BytesProcessed: bytesProcessed,
				CurrentFile:    hdr.Name,
			})
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(destPath, os.FileMode(hdr.Mode)); err != nil {
				return fmt.Errorf("创建解压目录失败 '%s': %w", destPath, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("创建父目录失败 '%s': %w", filepath.Dir(destPath), err)
			}
			destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return fmt.Errorf("创建目标文件失败 '%s': %w", destPath, err)
			}
			if _, err := io.Copy(destFile, tarReader); err != nil {
				_ = destFile.Close()
				return fmt.Errorf("解压文件内容失败 '%s': %w", hdr.Name, err)
			}
			_ = destFile.Close()

			if opts.progressFunc != nil {
				filesProcessed++
				bytesProcessed += hdr.Size
			}
		default:
			opts.logger.Warn("跳过不支持的 tar条目类型", "类型", hdr.Typeflag, "名称", hdr.Name)
		}
	}

	if opts.progressFunc != nil {
		opts.progressFunc(Progress{
			TotalFiles:     -1,
			FilesProcessed: filesProcessed,
			TotalBytes:     -1,
			BytesProcessed: bytesProcessed,
			CurrentFile:    "",
		})
	}

	opts.logger.Info("TAR.GZ 解压任务成功完成")
	return nil
}
