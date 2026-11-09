package api

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	logger "github.com/CIPFZ/gowebframe/internal/core/log"
	tokenCore "github.com/CIPFZ/gowebframe/internal/core/token"
	"github.com/CIPFZ/gowebframe/internal/middleware"
	"github.com/CIPFZ/gowebframe/internal/modules/system/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"github.com/CIPFZ/gowebframe/internal/modules/system/service"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/response"
)

type ApiTokenApi struct {
	svcCtx          *svc.ServiceContext
	apiTokenService service.IApiTokenService
}

func NewApiTokenApi(svcCtx *svc.ServiceContext, apiTokenService service.IApiTokenService) *ApiTokenApi {
	return &ApiTokenApi{
		svcCtx:          svcCtx,
		apiTokenService: apiTokenService,
	}
}

// actor 是当前操作的人；能调用“所有 Token”接口的人（管理员）可以管理所有人的 Token。
func (a *ApiTokenApi) actor(c *gin.Context) (service.TokenActor, error) {
	userID := c.GetUint(middleware.CtxKeyUserID)
	manage, err := middleware.UserAllows(c.Request.Context(), a.svcCtx, userID, a.svcCtx.Config.System.RouterPrefix+"/sys/api-token/all", "POST")
	return service.TokenActor{UserID: userID, Manage: manage}, err
}

// GetApiTokenList 只返回本人的 Token。
func (a *ApiTokenApi) GetApiTokenList(c *gin.Context) {
	a.listTokens(c, c.GetUint(middleware.CtxKeyUserID))
}

// GetAllApiTokens 返回所有人的 Token（管理员），可按人筛选。
func (a *ApiTokenApi) GetAllApiTokens(c *gin.Context) {
	a.listTokens(c, 0)
}

func (a *ApiTokenApi) listTokens(c *gin.Context, owner uint) {
	var req dto.SearchApiTokenReq
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}

	req.Normalize()
	if owner != 0 {
		req.OwnerID = owner
	}
	list, total, err := a.apiTokenService.GetApiTokenList(c.Request.Context(), req)
	if err != nil {
		logger.GetLogger(c).Error("get_api_token_list_error", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithPage(list, total, req.Page, req.PageSize, c)
}

func (a *ApiTokenApi) CreateApiToken(c *gin.Context) {
	var req dto.CreateApiTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}

	actor, err := a.actor(c)
	if err != nil {
		logger.GetLogger(c).Error("create_api_token_error", zap.Error(err))
		response.FailWithMessage("创建失败", c)
		return
	}
	resp, err := a.apiTokenService.CreateApiToken(c.Request.Context(), actor, req)
	if err != nil {
		logger.GetLogger(c).Error("create_api_token_error", zap.Error(err))
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithData(resp, c)
}

func (a *ApiTokenApi) GetApiTokenDetail(c *gin.Context) {
	var req dto.ApiTokenDetailReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}
	actor, err := a.actor(c)
	if err != nil {
		logger.GetLogger(c).Error("get_api_token_detail_error", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	resp, err := a.apiTokenService.GetApiTokenDetail(c.Request.Context(), actor, req.ID)
	if err != nil {
		logger.GetLogger(c).Error("get_api_token_detail_error", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithData(resp, c)
}

func (a *ApiTokenApi) UpdateApiToken(c *gin.Context) {
	var req dto.UpdateApiTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}
	actor, err := a.actor(c)
	if err == nil {
		err = a.apiTokenService.UpdateApiToken(c.Request.Context(), actor, req)
	}
	if err != nil {
		logger.GetLogger(c).Error("update_api_token_error", zap.Error(err))
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (a *ApiTokenApi) DeleteApiToken(c *gin.Context) {
	var req dto.DeleteApiTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}
	actor, err := a.actor(c)
	if err == nil {
		err = a.apiTokenService.DeleteApiToken(c.Request.Context(), actor, req)
	}
	if err != nil {
		logger.GetLogger(c).Error("delete_api_token_error", zap.Error(err))
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (a *ApiTokenApi) ResetApiToken(c *gin.Context) {
	var req dto.ToggleApiTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}
	actor, err := a.actor(c)
	if err != nil {
		logger.GetLogger(c).Error("reset_api_token_error", zap.Error(err))
		response.FailWithMessage("重置失败", c)
		return
	}
	resp, err := a.apiTokenService.ResetApiToken(c.Request.Context(), actor, req.ID)
	if err != nil {
		logger.GetLogger(c).Error("reset_api_token_error", zap.Error(err))
		response.FailWithMessage("重置失败: "+err.Error(), c)
		return
	}
	response.OkWithData(resp, c)
}

func (a *ApiTokenApi) EnableApiToken(c *gin.Context) {
	a.toggleApiToken(c, true)
}

func (a *ApiTokenApi) DisableApiToken(c *gin.Context) {
	a.toggleApiToken(c, false)
}

func (a *ApiTokenApi) toggleApiToken(c *gin.Context, enabled bool) {
	var req dto.ToggleApiTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}

	actor, err := a.actor(c)
	if err == nil {
		if enabled {
			err = a.apiTokenService.EnableApiToken(c.Request.Context(), actor, req.ID)
		} else {
			err = a.apiTokenService.DisableApiToken(c.Request.Context(), actor, req.ID)
		}
	}
	if err != nil {
		logger.GetLogger(c).Error("toggle_api_token_error", zap.Bool("enabled", enabled), zap.Error(err))
		response.FailWithMessage("操作失败: "+err.Error(), c)
		return
	}
	response.Ok(c)
}

// GetOptions 返回申请 Token 时可选的接口（可被 Token 调用、且本人现在有权限的）、本人的限制与已有的 Token 数。
// Token 自身信息接口不属于任何角色的权限，人人可选。
func (a *ApiTokenApi) GetOptions(c *gin.Context) {
	ctx := c.Request.Context()
	actor, err := a.actor(c)
	if err != nil {
		logger.GetLogger(c).Error("get_api_token_options_error", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	prefix := a.svcCtx.Config.System.RouterPrefix
	var apis []model.SysApi
	query := a.svcCtx.DB.WithContext(ctx).Where("1 = 0")
	for path, method := range tokenCore.Endpoints(prefix) {
		query = query.Or("path = ? AND method = ?", path, method)
	}
	if err := query.Order("api_group, path").Find(&apis).Error; err != nil {
		response.FailWithMessage("获取失败", c)
		return
	}
	roles, err := middleware.UserRoles(ctx, a.svcCtx.DB, actor.UserID)
	if err != nil {
		response.FailWithMessage("获取失败", c)
		return
	}
	mine := make([]model.SysApi, 0, len(apis))
	for _, api := range apis {
		ok, err := middleware.RolesAllow(ctx, a.svcCtx, roles, api.Path, api.Method)
		if err != nil {
			response.FailWithMessage("获取失败", c)
			return
		}
		if ok || actor.Manage || api.Path == tokenCore.TokenInfoPath(prefix) {
			mine = append(mine, api)
		}
	}
	limits, err := a.apiTokenService.EffectiveLimits(ctx, actor.UserID)
	if err != nil {
		response.FailWithMessage("获取失败", c)
		return
	}
	used, err := a.apiTokenService.CountTokens(ctx, actor.UserID)
	if err != nil {
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithData(gin.H{"list": mine, "total": len(mine), "limits": limits, "usedTokens": used}, c)
}

// GetLimit 返回一份限制设置（userId 为 0 是全局默认）与实际生效的值。
func (a *ApiTokenApi) GetLimit(c *gin.Context) {
	var req dto.TokenLimitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}
	detail, err := a.apiTokenService.GetLimitDetail(c.Request.Context(), req.UserID)
	if err != nil {
		logger.GetLogger(c).Error("get_token_limit_error", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithData(detail, c)
}

// UpdateLimit 保存一份限制设置；字段为空表示跟随全局默认。
func (a *ApiTokenApi) UpdateLimit(c *gin.Context) {
	var req dto.TokenLimitSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}
	if err := a.apiTokenService.UpdateLimit(c.Request.Context(), req); err != nil {
		logger.GetLogger(c).Error("update_token_limit_error", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}
