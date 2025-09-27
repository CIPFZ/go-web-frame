package copyguard

import (
	"context"
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"
)

const copyChunkSize = 32 * 1024 // 32 KB

// executeCopy 是实际执行单个文件拷贝任务的内部方法。
func (m *Manager) executeCopy(ctx context.Context, src, dst string, onProgress func(Progress)) error {
	// 1. 获取源文件信息
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("无法获取源文件信息 '%s': %w", src, err)
	}

	// 2. 验证源文件是否是一个普通文件
	if !sourceFileStat.Mode().IsRegular() {
		return ErrSourceIsDirectory
	}

	m.logger.Debug("开始拷贝文件", zap.String("源", src), zap.String("目标", dst), zap.Int64("大小", sourceFileStat.Size()))

	// 3. 打开源文件
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("无法打开源文件 '%s': %w", src, err)
	}
	defer sourceFile.Close()

	// 4. 创建或覆盖目标文件
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("无法创建目标文件 '%s': %w", dst, err)
	}
	defer destFile.Close()

	// 5. 创建一个缓冲区来执行分块拷贝
	buffer := make([]byte, copyChunkSize) // <--- 使用常量
	var copiedBytes int64

	// 6. 循环拷贝数据
	for {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err() // 如果上下文被取消，则中止拷贝
		default:
		}

		bytesRead, readErr := sourceFile.Read(buffer)
		if bytesRead > 0 {
			// --- 执行限速 ---
			// 在写入之前，首先等待 IOPS 令牌。这代表一次 I/O 操作。
			if err := m.iopsLimiter.Wait(ctx); err != nil {
				return err
			}
			// 然后，等待足够的带宽令牌。
			// WaitN 会阻塞，直到桶中有足够（bytesRead）的令牌。
			if err := m.bandwidthLimiter.WaitN(ctx, bytesRead); err != nil {
				return err
			}
			// --- 限速结束 ---

			// 写入数据
			_, writeErr := destFile.Write(buffer[:bytesRead])
			if writeErr != nil {
				return fmt.Errorf("写入目标文件 '%s' 时出错: %w", dst, writeErr)
			}
			copiedBytes += int64(bytesRead)
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("读取源文件 '%s' 时出错: %w", src, readErr)
		}
	}

	m.logger.Debug("文件拷贝完成", zap.String("源", src), zap.String("目标", dst))
	return nil
}
