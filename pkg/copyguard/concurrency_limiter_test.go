package copyguard

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestRateLimiting 测试带宽限速功能是否生效
func TestRateLimiting(t *testing.T) {
	// 1. 准备环境
	sourceDir := t.TempDir()
	destDir := t.TempDir()
	sourceFile := filepath.Join(sourceDir, "largefile.bin")

	// --- 修正点 1: 增大文件大小，使测试更稳定 ---
	fileSize := 200 * 1024 // 200 KB
	fileContent := make([]byte, fileSize)
	// (省略了填充内容的循环，因为内容本身不重要)
	err := os.WriteFile(sourceFile, fileContent, 0644)
	require.NoError(t, err)

	// 2. 创建一个限速为 50 KB/s 的 Manager
	limit := float64(50 * 1024) // 50 KB/s
	manager, err := NewManager(WithBandwidthLimit(limit))
	require.NoError(t, err)
	defer manager.Close()

	// 3. 执行并计时
	startTime := time.Now()
	err = manager.Copy(context.Background(), sourceFile, filepath.Join(destDir, "dest.bin"), nil)
	duration := time.Since(startTime)

	// 4. 验证结果
	require.NoError(t, err)

	// --- 修正点 2: 使用更科学的公式计算期望的最小耗时 ---
	// 理论耗时 = (总大小 - 突发容量) / 速率
	// 突发容量就是我们的 copyChunkSize
	throttledSize := float64(fileSize - copyChunkSize)
	theoreticalDuration := time.Duration(throttledSize/limit) * time.Second

	// 我们允许一些误差（比如 10%），所以期望的最小耗时是理论值的 90%
	expectedMinDuration := theoreticalDuration * 9 / 10

	assert.GreaterOrEqual(t, duration, expectedMinDuration, "拷贝耗时应受到限速影响")
	t.Logf("文件大小: %d B, 限速: %.2f B/s, 理论耗时: %v, 实际耗时: %v", fileSize, limit, theoreticalDuration, duration)
}

// TestConcurrency 测试并发拷贝功能 (此测试用例无需改动)
func TestConcurrency(t *testing.T) {
	// 1. 准备环境：创建多个源文件
	sourceDir := t.TempDir()
	destDir := t.TempDir()
	numFiles := 10
	fileSize := 10 * 1024 // 10 KB
	fileContent := make([]byte, fileSize)

	for i := 0; i < numFiles; i++ {
		sourceFile := filepath.Join(sourceDir, fmt.Sprintf("file%d.txt", i))
		err := os.WriteFile(sourceFile, fileContent, 0644)
		require.NoError(t, err)
	}

	// 2. 创建一个并发数为 4 的 Manager
	manager, err := NewManager(WithConcurrency(4))
	require.NoError(t, err)
	defer manager.Close()

	// 3. 并发执行所有拷贝任务
	var wg sync.WaitGroup
	wg.Add(numFiles)

	for i := 0; i < numFiles; i++ {
		// 必须在循环内创建局部变量，否则会因为闭包问题导致所有 goroutine 使用相同的变量值
		src := filepath.Join(sourceDir, fmt.Sprintf("file%d.txt", i))
		dst := filepath.Join(destDir, fmt.Sprintf("file%d.txt", i))

		go func() {
			defer wg.Done()
			err := manager.Copy(context.Background(), src, dst, nil)
			// 在 goroutine 内部直接断言，如果出错测试会失败
			assert.NoError(t, err, "并发拷贝不应出错: "+src)
		}()
	}

	// 等待所有拷贝任务完成
	wg.Wait()

	// 4. 验证结果：检查所有目标文件是否都已正确创建
	for i := 0; i < numFiles; i++ {
		destFile := filepath.Join(destDir, fmt.Sprintf("file%d.txt", i))
		stat, err := os.Stat(destFile)
		assert.NoError(t, err, "目标文件应存在: "+destFile)
		assert.Equal(t, int64(fileSize), stat.Size(), "目标文件大小应正确")
	}
}
