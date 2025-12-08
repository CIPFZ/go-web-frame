package audit

import (
	"context"
	"gorm.io/gorm"
	"sync"
	"time"

	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"go.uber.org/zap"
)

const (
	logChanCapacity = 10000           // logChan 的缓冲区容量，用于暂存待处理的日志
	batchSize       = 100             // 批量写入数据库的日志数量阈值
	flushInterval   = 2 * time.Second // 定时将缓冲区日志写入数据库的时间间隔
)

// AuditRecorder 它使用一个带缓冲的 channel 和一个后台 goroutine 来实现操作日志的异步、批量写入，
// 从而避免阻塞 HTTP 请求的正常流程。
type AuditRecorder struct {
	db      *gorm.DB
	logger  *zap.Logger                // 日志记录器
	logChan chan model.SysOperationLog // 用于接收日志的带缓冲通道
	wg      sync.WaitGroup             // 用于等待后台 worker goroutine 优雅退出
}

// NewAuditRecorder 创建并启动一个新的 OperationLogService 实例。
// 它会立即启动一个后台 worker goroutine 来消费日志。
func NewAuditRecorder(db *gorm.DB, logger *zap.Logger) *AuditRecorder {
	s := &AuditRecorder{
		db:      db,
		logger:  logger,
		logChan: make(chan model.SysOperationLog, logChanCapacity),
	}

	s.wg.Add(1)
	go s.startWorker()

	return s
}

// Push 是一个非阻塞方法，用于将操作日志推送到处理队列。
// 如果队列已满，它会记录一条警告并丢弃该日志，以防止阻塞调用方（通常是中间件）。
func (r *AuditRecorder) Push(log model.SysOperationLog) {
	select {
	case r.logChan <- log:
		// 成功推送到通道
	default:
		// 通道已满，执行降级策略：丢弃日志并记录警告
		r.logger.Warn("operation_log_dropped_queue_full", zap.String("path", log.Path))
	}
}

// startWorker 是一个长期运行的后台 goroutine，负责从 logChan 消费日志并批量写入数据库。
func (r *AuditRecorder) startWorker() {
	defer r.wg.Done() // 当 goroutine 退出时，通知 WaitGroup

	// 预分配批处理切片的容量，以减少后续 append 操作的内存分配次数
	batch := make([]model.SysOperationLog, 0, batchSize)
	// 创建一个定时器，用于周期性地强制刷写缓冲区
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case log, ok := <-r.logChan:
			if !ok {
				// 通道已被关闭 (由 Close 方法触发)，这是退出的信号。
				// 在退出前，处理并刷写缓冲区中剩余的所有日志。
				if len(batch) > 0 {
					r.flush(batch)
				}
				return // 退出 goroutine
			}

			batch = append(batch, log)
			// 当缓冲区大小达到阈值时，立即刷写
			if len(batch) >= batchSize {
				r.flush(batch)
				batch = batch[:0] // 重置切片，保持底层数组以复用内存
			}

		case <-ticker.C:
			// 定时器触发，强制刷写缓冲区中积攒的日志
			if len(batch) > 0 {
				r.flush(batch)
				batch = batch[:0]
			}
		}
	}
}

// flush 调用数据仓库将一批日志批量写入数据库。
func (r *AuditRecorder) flush(logs []model.SysOperationLog) {
	if err := r.db.CreateInBatches(logs, len(logs)).Error; err != nil {
		// 此时数据库可能挂了，记录到文件日志作为最后的备份
		r.logger.Error("flush_operation_logs_failed", zap.Error(err), zap.Int("count", len(logs)))
	}
}

// Close 优雅地关闭日志服务。
func (r *AuditRecorder) Close(ctx context.Context) error {
	// 1. 关闭 logChan，这将向 startWorker 发出停止接收新日志并准备退出的信号。
	close(r.logChan)

	// 2. 使用一个 goroutine 等待 WaitGroup 完成，这表示 startWorker 已处理完所有剩余日志并退出。
	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()

	// 3. 等待 worker 完成，或等待外部上下文超时/取消。
	select {
	case <-done:
		// worker 正常关闭
		r.logger.Info("OperationLogService closed gracefully")
		return nil
	case <-ctx.Done():
		// 等待超时
		r.logger.Warn("OperationLogService close timed out", zap.Error(ctx.Err()))
		return ctx.Err()
	}
}
