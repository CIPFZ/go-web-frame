package system

import (
	"context"

	systemModel "github.com/CIPFZ/gowebframe/internal/model/system"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/logger"
	"github.com/gin-gonic/gin"

	"go.uber.org/zap"
)

type IJwtService interface {
	JsonInBlacklist(jwtList systemModel.JwtBlacklist) (err error)
	GetRedisJWT(userName string) (redisJWT string, err error)
	LoadAll(c *gin.Context)
}

type JwtService struct {
	svcCtx *svc.ServiceContext
}

func NewJwtService(svcCtx *svc.ServiceContext) *JwtService {
	return &JwtService{
		svcCtx: svcCtx,
	}
}

// JsonInBlacklist 拉黑jwt
func (jwtService *JwtService) JsonInBlacklist(jwtList systemModel.JwtBlacklist) (err error) {
	err = jwtService.svcCtx.DB.Create(&jwtList).Error
	if err != nil {
		return
	}
	// TODO 加入黑名单
	return
}

// GetRedisJWT 从redis取jwt
func (jwtService *JwtService) GetRedisJWT(userName string) (redisJWT string, err error) {
	redisJWT, err = jwtService.svcCtx.Redis.Get(context.Background(), userName).Result()
	return redisJWT, err
}

func (jwtService *JwtService) LoadAll(c *gin.Context) {
	var data []string
	err := jwtService.svcCtx.DB.Model(&systemModel.JwtBlacklist{}).Select("jwt").Find(&data).Error
	log := logger.GetLogger(c)
	if err != nil {
		log.Error("加载数据库jwt黑名单失败!", zap.Error(err))
		return
	}
	// TODO jwt黑名单 加入 BlackCache 中
}
