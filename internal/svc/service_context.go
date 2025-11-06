package svc

import (
	"net/http"
	"sync"

	"github.com/CIPFZ/gowebframe/internal/config"
	"github.com/CIPFZ/gowebframe/internal/utils/timer"
	"github.com/CIPFZ/gowebframe/pkg/i18n"

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
	Logger             *zap.Logger
	I18n               *i18n.Service
	DB                 *gorm.DB
	Redis              redis.UniversalClient
	Mongo              *qmgo.QmgoClient
	Routers            gin.RoutesInfo
	Timer              timer.Timer
	ConcurrencyControl *singleflight.Group
	lock               sync.RWMutex
}

func NewServiceContext() *ServiceContext {
	return &ServiceContext{
		ConcurrencyControl: &singleflight.Group{},
	}
}
