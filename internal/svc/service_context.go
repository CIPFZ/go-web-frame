package svc

import (
	"context"
	"net/http"
	"sync"

	"github.com/CIPFZ/gowebframe/internal/config"
	systemModel "github.com/CIPFZ/gowebframe/internal/model/system"
	systemReq "github.com/CIPFZ/gowebframe/internal/model/system/request"
	"github.com/CIPFZ/gowebframe/internal/utils"
	"github.com/CIPFZ/gowebframe/internal/utils/timer"
	"github.com/CIPFZ/gowebframe/pkg/i18n"
	"github.com/casbin/casbin/v2"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/qmgo"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type OperationLogPort interface {
	Push(log systemModel.SysOperationLog)
	GetOperationLogList(ctx context.Context, req systemReq.SearchOperationLogReq) ([]systemModel.SysOperationLog, int64, error)
	DeleteOperationLogByIds(ctx context.Context, ids []uint) error
}

type ServiceContext struct {
	SRV                 *http.Server
	Config              *config.Config
	Viper               *viper.Viper
	JWT                 *utils.JWT
	Logger              *zap.Logger
	I18n                *i18n.Service
	DB                  *gorm.DB
	Redis               redis.UniversalClient
	Mongo               *qmgo.QmgoClient
	Routers             gin.RoutesInfo
	Timer               timer.Timer
	ConcurrencyControl  *singleflight.Group
	CasbinEnforcer      *casbin.SyncedCachedEnforcer
	lock                sync.RWMutex
	OperationLogService OperationLogPort
}

func NewServiceContext() *ServiceContext {
	return &ServiceContext{
		ConcurrencyControl: &singleflight.Group{},
	}
}
