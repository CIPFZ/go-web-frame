package request

import (
	"github.com/CIPFZ/gowebframe/internal/model/common/request"
	"github.com/CIPFZ/gowebframe/internal/model/system"
)

type SysOperationRecordSearch struct {
	system.SysOperationRecord
	request.PageInfo
}
