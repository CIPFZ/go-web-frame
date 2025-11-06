package system

import (
	"github.com/CIPFZ/gowebframe/internal/model/common/request"
	"github.com/CIPFZ/gowebframe/internal/model/system"
	systemReq "github.com/CIPFZ/gowebframe/internal/model/system/request"
	"github.com/CIPFZ/gowebframe/internal/svc"
)

//@author: [granty1](https://github.com/granty1)
//@function: CreateSysOperationRecord
//@description: 创建记录
//@param: sysOperationRecord model.SysOperationRecord
//@return: err error

type IOperationRecordService interface {
	DeleteSysOperationRecordByIds(ids request.IdsReq) error
	DeleteSysOperationRecord(sysOperationRecord system.SysOperationRecord) error
	GetSysOperationRecord(id uint) (system.SysOperationRecord, error)
	GetSysOperationRecordInfoList(info systemReq.SysOperationRecordSearch) (interface{}, int64, error)
}

type OperationRecordService struct {
	serviceCtx *svc.ServiceContext
}

func NewOperationRecordService(ctx *svc.ServiceContext) IOperationRecordService {
	return &OperationRecordService{serviceCtx: ctx}
}

// DeleteSysOperationRecordByIds 批量删除记录
func (s *OperationRecordService) DeleteSysOperationRecordByIds(ids request.IdsReq) error {
	return s.serviceCtx.DB.Delete(&[]system.SysOperationRecord{}, "id in (?)", ids.Ids).Error
}

// DeleteSysOperationRecord 删除操作记录
func (s *OperationRecordService) DeleteSysOperationRecord(sysOperationRecord system.SysOperationRecord) error {
	return s.serviceCtx.DB.Delete(&sysOperationRecord).Error
}

// GetSysOperationRecord 根据id获取单条操作记录
func (s *OperationRecordService) GetSysOperationRecord(id uint) (system.SysOperationRecord, error) {
	var sysOperationRecord system.SysOperationRecord
	var err error
	err = s.serviceCtx.DB.Where("id = ?", id).First(&sysOperationRecord).Error
	return sysOperationRecord, err
}

// GetSysOperationRecordInfoList 分页获取操作记录列表
func (s *OperationRecordService) GetSysOperationRecordInfoList(info systemReq.SysOperationRecordSearch) (interface{}, int64, error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	// 创建db
	db := s.serviceCtx.DB.Model(&system.SysOperationRecord{})
	var sysOperationRecords []system.SysOperationRecord
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.Method != "" {
		db = db.Where("method = ?", info.Method)
	}
	if info.Path != "" {
		db = db.Where("path LIKE ?", "%"+info.Path+"%")
	}
	if info.Status != 0 {
		db = db.Where("status = ?", info.Status)
	}
	var err error
	var total int64
	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.Order("id desc").Limit(limit).Offset(offset).Preload("User").Find(&sysOperationRecords).Error
	return sysOperationRecords, total, err
}
