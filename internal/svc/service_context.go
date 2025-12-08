package svc

import (
	"github.com/CIPFZ/gowebframe/internal/core/file"
	"net/http"
	"sync"
	"time"

	"github.com/CIPFZ/gowebframe/internal/core/audit"
	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/core/i18n"
	"github.com/CIPFZ/gowebframe/internal/core/jwt"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/qiniu/qmgo"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type ServiceContext struct {
	SRV                *http.Server
	Config             *config.Config
	Viper              *viper.Viper
	JWT                *jwt.JWT
	Logger             *zap.Logger
	I18n               *i18n.Service
	DB                 *gorm.DB
	Redis              redis.UniversalClient
	Mongo              *qmgo.QmgoClient
	Routers            gin.RoutesInfo
	Timer              time.Timer
	ConcurrencyControl *singleflight.Group
	CasbinEnforcer     *casbin.SyncedCachedEnforcer
	lock               sync.RWMutex
	AuditRecorder      *audit.AuditRecorder
	OSS                file.OSS
}

func NewServiceContext() *ServiceContext {
	return &ServiceContext{
		ConcurrencyControl: &singleflight.Group{},
	}
}
