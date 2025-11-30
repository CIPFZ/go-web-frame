package system

//
//import (
//	"github.com/CIPFZ/gowebframe/internal/model/common/response"
//	systemModel "github.com/CIPFZ/gowebframe/internal/model/system"
//	systemService "github.com/CIPFZ/gowebframe/internal/service/system"
//	"github.com/CIPFZ/gowebframe/internal/svc"
//	"github.com/CIPFZ/gowebframe/internal/utils"
//	"github.com/CIPFZ/gowebframe/pkg/logger"
//
//	"github.com/gin-gonic/gin"
//	"go.uber.org/zap"
//)
//
//type JwtApi struct {
//	svcCtx  *svc.ServiceContext
//	service systemService.IJwtService
//}
//
//func NewJwtApi(svcCtx *svc.ServiceContext) *JwtApi {
//	return &JwtApi{
//		svcCtx:  svcCtx,
//		service: systemService.NewJwtService(svcCtx),
//	}
//}
//
//// JsonInBlacklist jwt加入黑名单
//func (j *JwtApi) JsonInBlacklist(c *gin.Context) {
//	token := utils.GetToken(c, j.svcCtx)
//	jwt := systemModel.JwtBlacklist{Jwt: token}
//	err := j.service.JsonInBlacklist(jwt)
//	log := logger.GetLogger(c)
//	if err != nil {
//		log.Error("jwt作废失败!", zap.Error(err))
//		response.FailWithMessage("jwt作废失败", c)
//		return
//	}
//	utils.ClearToken(c)
//	response.OkWithMessage("jwt作废成功", c)
//}
