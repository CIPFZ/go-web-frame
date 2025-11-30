package system

import (
	"context"
	"sync"
	"time"

	systemModel "github.com/CIPFZ/gowebframe/internal/model/system"
	systemReq "github.com/CIPFZ/gowebframe/internal/model/system/request"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	logChanCapacity = 10000 // 缓冲区容量
	batchSize       = 100   // 批次大小
	flushInterval   = 2 * time.Second
)

type OperationLogService struct {
	db     *gorm.DB
	logger *zap.Logger

	logChan chan systemModel.SysOperationLog
	wg      sync.WaitGroup
}

func NewOperationLogService(db *gorm.DB, logger *zap.Logger) *OperationLogService {
	s := &OperationLogService{
		db:      db,
		logger:  logger,
		logChan: make(chan systemModel.SysOperationLog, logChanCapacity),
	}

	s.wg.Add(1)
	go s.startWorker()

	return s
}

// Push 推入队列 (非阻塞，高性能)
func (s *OperationLogService) Push(log systemModel.SysOperationLog) {
	select {
	case s.logChan <- log:
	default:
		// 队列满时的降级策略：记录错误指标或日志，但不阻塞主业务
		// 在高负载系统中，丢弃日志保业务是标准做法
		s.logger.Warn("operation_log_dropped_queue_full")
	}
}

// startWorker 消费者
func (s *OperationLogService) startWorker() {
	// ✨ 确保 Worker 退出时通知 WaitGroup
	defer s.wg.Done()

	var batch []systemModel.SysOperationLog
	// 预分配切片容量，减少 append 时的内存重新分配
	batch = make([]systemModel.SysOperationLog, 0, batchSize)

	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case log, ok := <-s.logChan:
			if !ok {
				// Channel 已关闭
				// 必须把 batch 中残留的数据写入数据库
				if len(batch) > 0 {
					s.flush(batch)
				}
				return // 退出 Worker
			}

			batch = append(batch, log)
			if len(batch) >= batchSize {
				s.flush(batch)
				batch = batch[:0] // 清空并复用底层数组
			}

		case <-ticker.C:
			if len(batch) > 0 {
				s.flush(batch)
				batch = batch[:0]
			}
		}
	}
}

// flush 批量写入
func (s *OperationLogService) flush(logs []systemModel.SysOperationLog) {
	// 使用 GORM 的 CreateInBatches 提高性能
	if err := s.db.CreateInBatches(logs, len(logs)).Error; err != nil {
		// 此时数据库可能挂了，记录到文件日志作为最后的备份
		s.logger.Error("flush_operation_logs_failed", zap.Error(err), zap.Int("count", len(logs)))
	}
}

// Close 企业级优雅关闭
// 1. 关闭 channel 停止接收新日志
// 2. 等待 worker 处理完积压数据并写入数据库
// 3. 支持 Context 超时控制
// Close 优雅关闭
func (s *OperationLogService) Close(ctx context.Context) error {
	close(s.logChan)

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("OperationLogService closed gracefully")
		return nil
	case <-ctx.Done():
		s.logger.Warn("OperationLogService close timed out", zap.Error(ctx.Err()))
		return ctx.Err()
	}
}

// GetOperationLogList 分页获取日志
func (s *OperationLogService) GetOperationLogList(ctx context.Context, req systemReq.SearchOperationLogReq) ([]systemModel.SysOperationLog, int64, error) {
	limit := req.PageSize
	offset := req.PageSize * (req.Page - 1)

	db := s.db.WithContext(ctx).Model(&systemModel.SysOperationLog{})
	var list []systemModel.SysOperationLog
	var total int64

	// 动态查询条件
	if req.Method != "" {
		db = db.Where("method = ?", req.Method)
	}
	if req.Path != "" {
		db = db.Where("path LIKE ?", "%"+req.Path+"%")
	}
	if req.Ip != "" {
		db = db.Where("ip LIKE ?", "%"+req.Ip+"%")
	}
	if req.Status != nil {
		db = db.Where("status = ?", *req.Status)
	}
	if req.UserID != 0 {
		db = db.Where("user_id = ?", req.UserID)
	}
	if req.TraceID != "" {
		db = db.Where("trace_id = ?", req.TraceID)
	}
	if req.StartDate != nil && req.EndDate != nil {
		db = db.Where("created_at BETWEEN ? AND ?", req.StartDate, req.EndDate)
	}

	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表 (关联 User 获取昵称, 按时间倒序)
	// 注意：Resp 和 Body 可能很大，列表页如果不展示详情，可以考虑 Omit 掉这两个字段以提升性能
	// 但为了简单，这里先 Select *
	err := db.Limit(limit).Offset(offset).
		Preload("User").
		Order("id desc").
		Find(&list).Error

	return list, total, err
}

// DeleteOperationLogByIds 批量删除
func (s *OperationLogService) DeleteOperationLogByIds(ctx context.Context, ids []uint) error {
	// 硬删除 (日志数据量大，通常不需要软删除)
	return s.db.WithContext(ctx).Unscoped().Delete(&systemModel.SysOperationLog{}, ids).Error
}
