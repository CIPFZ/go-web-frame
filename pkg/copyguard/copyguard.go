package copyguard

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"sync"

	"github.com/panjf2000/ants/v2"
	"golang.org/x/time/rate"
)

// Manager 是所有拷贝操作的核心控制器。
// 它管理资源、执行限制策略，并协调整个拷贝过程。
type Manager struct {
	config Config
	logger *zap.Logger

	// --- 新增组件 ---
	pool             *ants.Pool    // 用于管理拷贝任务的 goroutine 池
	bandwidthLimiter *rate.Limiter // 带宽限制器 (字节/秒)
	iopsLimiter      *rate.Limiter // IOPS 限制器 (次/秒)

	// 后续里程碑中会用到这些组件
	// diskBudget *atomic.Int64
	// globalCtx   context.Context
	// cancelFunc  context.CancelFunc
}

// NewManager 使用指定的选项创建并初始化一个新的 Manager 实例。
func NewManager(opts ...Option) (*Manager, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, fmt.Errorf("应用选项失败: %w", err)
		}
	}

	// --- 初始化新增组件 ---

	// 初始化带宽限制器
	bwLimit := rate.Limit(cfg.BandwidthLimit)
	if cfg.BandwidthLimit <= 0 {
		bwLimit = rate.Inf
	}
	// --- 修正点 ---
	// Burst 的大小应该与单次 I/O 的块大小匹配，以实现平滑限速
	bwLimiter := rate.NewLimiter(bwLimit, copyChunkSize)

	// 初始化 IOPS 限制器
	iopsLimit := rate.Limit(cfg.IOPSLimit)
	if cfg.IOPSLimit <= 0 {
		iopsLimit = rate.Inf
	}
	// IOPS 的突发通常设置为1，代表允许一次突发的 I/O 操作
	iopsLimiter := rate.NewLimiter(iopsLimit, 1)

	// ... (ants 池的初始化不变) ...
	pool, err := ants.NewPool(cfg.Concurrency)
	if err != nil {
		return nil, fmt.Errorf("创建 ants goroutine 池失败: %w", err)
	}

	m := &Manager{
		config:           *cfg,
		logger:           cfg.Logger.Named("copyguard"),
		pool:             pool,
		bandwidthLimiter: bwLimiter,
		iopsLimiter:      iopsLimiter,
	}

	m.logger.Info("Copyguard 管理器已初始化", zap.Any("配置", m.config))
	return m, nil
}

// Copy 向管理器提交一个文件拷贝操作。
// 此方法会阻塞，直到拷贝完成或因错误、上下文取消而中止。
func (m *Manager) Copy(ctx context.Context, src, dst string, onProgress func(Progress)) error {
	m.logger.Info("收到拷贝请求，提交到任务池",
		zap.String("源", src),
		zap.String("目标", dst),
	)

	// 使用 WaitGroup 和 error channel 来实现阻塞和错误传递
	var wg sync.WaitGroup
	var taskErr error
	wg.Add(1)

	// 将任务提交到 ants 池
	err := m.pool.Submit(func() {
		defer wg.Done()
		// 在池的 goroutine 中执行实际的拷贝逻辑
		taskErr = m.executeCopy(ctx, src, dst, onProgress)
	})
	if err != nil {
		// 如果池已关闭或满了，Submit会返回错误
		wg.Done() // 必须调用 Done 来释放 WaitGroup
		return fmt.Errorf("提交任务到池失败: %w", err)
	}

	// 等待任务完成
	wg.Wait()
	return taskErr
}

// Close 平滑地关闭管理器，释放所有资源。
// 它会等待所有已提交的任务执行完毕。
func (m *Manager) Close() error {
	m.logger.Info("收到关闭请求，正在释放资源...")
	if m.pool != nil {
		m.pool.Release() // 释放 ants 池
	}
	m.logger.Info("管理器已关闭")
	return nil
}
