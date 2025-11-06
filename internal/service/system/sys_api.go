package system

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CIPFZ/gowebframe/internal/model/common/request"
	systemModel "github.com/CIPFZ/gowebframe/internal/model/system"
	systemRes "github.com/CIPFZ/gowebframe/internal/model/system/response"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"gorm.io/gorm"
)

type IApiService interface {
	CreateApi(api systemModel.SysApi) (err error)
	GetApiGroups() (groups []string, groupApiMap map[string]string, err error)
	SyncApi() (newApis, deleteApis, ignoreApis []systemModel.SysApi, err error)
	IgnoreApi(ignoreApi systemModel.SysIgnoreApi) (err error)
	EnterSyncApi(syncApis systemRes.SysSyncApis) (err error)
	DeleteApi(api systemModel.SysApi) (err error)
	GetAPIInfoList(api systemModel.SysApi, info request.PageInfo, order string, desc bool) (list interface{}, total int64, err error)
	GetAllApis(authorityID uint) (apis []systemModel.SysApi, err error)
	GetApiById(id int) (api systemModel.SysApi, err error)
	UpdateApi(api systemModel.SysApi) (err error)
	DeleteApisByIds(ids request.IdsReq) (err error)
}

type ApiService struct {
	svcCtx        *svc.ServiceContext
	casbinService *CasbinService
}

func NewApiService(svcCtx *svc.ServiceContext) IApiService {
	return &ApiService{
		svcCtx:        svcCtx,
		casbinService: &CasbinService{svcCtx: svcCtx},
	}
}

// CreateApi 新增基础api
func (s *ApiService) CreateApi(api systemModel.SysApi) (err error) {
	if !errors.Is(s.svcCtx.DB.Where("path = ? AND method = ?", api.Path, api.Method).First(&systemModel.SysApi{}).Error, gorm.ErrRecordNotFound) {
		return errors.New("存在相同api")
	}
	return s.svcCtx.DB.Create(&api).Error
}

func (s *ApiService) GetApiGroups() (groups []string, groupApiMap map[string]string, err error) {
	var apis []systemModel.SysApi
	err = s.svcCtx.DB.Find(&apis).Error
	if err != nil {
		return
	}
	groupApiMap = make(map[string]string, 0)
	for i := range apis {
		pathArr := strings.Split(apis[i].Path, "/")
		newGroup := true
		for i2 := range groups {
			if groups[i2] == apis[i].ApiGroup {
				newGroup = false
			}
		}
		if newGroup {
			groups = append(groups, apis[i].ApiGroup)
		}
		groupApiMap[pathArr[1]] = apis[i].ApiGroup
	}
	return
}

func (s *ApiService) SyncApi() (newApis, deleteApis, ignoreApis []systemModel.SysApi, err error) {
	newApis = make([]systemModel.SysApi, 0)
	deleteApis = make([]systemModel.SysApi, 0)
	ignoreApis = make([]systemModel.SysApi, 0)
	var apis []systemModel.SysApi
	err = s.svcCtx.DB.Find(&apis).Error
	if err != nil {
		return
	}
	var ignores []systemModel.SysIgnoreApi
	err = s.svcCtx.DB.Find(&ignores).Error
	if err != nil {
		return
	}

	for i := range ignores {
		ignoreApis = append(ignoreApis, systemModel.SysApi{
			Path:        ignores[i].Path,
			Description: "",
			ApiGroup:    "",
			Method:      ignores[i].Method,
		})
	}

	var cacheApis []systemModel.SysApi
	routers := s.svcCtx.Routers
	for i := range routers {
		ignoresFlag := false
		for j := range ignores {
			if ignores[j].Path == routers[i].Path && ignores[j].Method == routers[i].Method {
				ignoresFlag = true
			}
		}
		if !ignoresFlag {
			cacheApis = append(cacheApis, systemModel.SysApi{
				Path:   routers[i].Path,
				Method: routers[i].Method,
			})
		}
	}

	//对比数据库中的api和内存中的api，如果数据库中的api不存在于内存中，则把api放入删除数组，如果内存中的api不存在于数据库中，则把api放入新增数组
	for i := range cacheApis {
		var flag bool
		// 如果存在于内存不存在于api数组中
		for j := range apis {
			if cacheApis[i].Path == apis[j].Path && cacheApis[i].Method == apis[j].Method {
				flag = true
			}
		}
		if !flag {
			newApis = append(newApis, systemModel.SysApi{
				Path:        cacheApis[i].Path,
				Description: "",
				ApiGroup:    "",
				Method:      cacheApis[i].Method,
			})
		}
	}

	for i := range apis {
		var flag bool
		// 如果存在于api数组不存在于内存
		for j := range cacheApis {
			if cacheApis[j].Path == apis[i].Path && cacheApis[j].Method == apis[i].Method {
				flag = true
			}
		}
		if !flag {
			deleteApis = append(deleteApis, apis[i])
		}
	}
	return
}

func (s *ApiService) IgnoreApi(ignoreApi systemModel.SysIgnoreApi) (err error) {
	if ignoreApi.Flag {
		return s.svcCtx.DB.Create(&ignoreApi).Error
	}
	return s.svcCtx.DB.Unscoped().Delete(&ignoreApi, "path = ? AND method = ?", ignoreApi.Path, ignoreApi.Method).Error
}

func (s *ApiService) EnterSyncApi(syncApis systemRes.SysSyncApis) (err error) {
	return s.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		var txErr error
		if len(syncApis.NewApis) > 0 {
			txErr = tx.Create(&syncApis.NewApis).Error
			if txErr != nil {
				return txErr
			}
		}
		for i := range syncApis.DeleteApis {
			s.casbinService.ClearCasbin(1, syncApis.DeleteApis[i].Path, syncApis.DeleteApis[i].Method)
			txErr = tx.Delete(&systemModel.SysApi{}, "path = ? AND method = ?", syncApis.DeleteApis[i].Path, syncApis.DeleteApis[i].Method).Error
			if txErr != nil {
				return txErr
			}
		}
		return nil
	})
}

// DeleteApi 删除基础api
func (s *ApiService) DeleteApi(api systemModel.SysApi) (err error) {
	var entity systemModel.SysApi
	err = s.svcCtx.DB.First(&entity, "id = ?", api.ID).Error // 根据id查询api记录
	if errors.Is(err, gorm.ErrRecordNotFound) {              // api记录不存在
		return err
	}
	err = s.svcCtx.DB.Delete(&entity).Error
	if err != nil {
		return err
	}
	s.casbinService.ClearCasbin(1, entity.Path, entity.Method)
	return nil
}

// GetAPIInfoList 分页获取数据
func (s *ApiService) GetAPIInfoList(api systemModel.SysApi, info request.PageInfo, order string, desc bool) (list interface{}, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	db := s.svcCtx.DB.Model(&systemModel.SysApi{})
	var apiList []systemModel.SysApi

	if api.Path != "" {
		db = db.Where("path LIKE ?", "%"+api.Path+"%")
	}

	if api.Description != "" {
		db = db.Where("description LIKE ?", "%"+api.Description+"%")
	}

	if api.Method != "" {
		db = db.Where("method = ?", api.Method)
	}

	if api.ApiGroup != "" {
		db = db.Where("api_group = ?", api.ApiGroup)
	}

	err = db.Count(&total).Error

	if err != nil {
		return apiList, total, err
	}

	db = db.Limit(limit).Offset(offset)
	OrderStr := "id desc"
	if order != "" {
		orderMap := make(map[string]bool, 5)
		orderMap["id"] = true
		orderMap["path"] = true
		orderMap["api_group"] = true
		orderMap["description"] = true
		orderMap["method"] = true
		if !orderMap[order] {
			err = fmt.Errorf("非法的排序字段: %v", order)
			return apiList, total, err
		}
		OrderStr = order
		if desc {
			OrderStr = order + " desc"
		}
	}
	err = db.Order(OrderStr).Find(&apiList).Error
	return apiList, total, err
}

// GetAllApis 获取所有的api
func (s *ApiService) GetAllApis(authorityID uint) (apis []systemModel.SysApi, err error) {
	authorityService := NewAuthorityService(s.svcCtx)
	parentAuthorityID, err := authorityService.GetParentAuthorityID(authorityID)
	if err != nil {
		return nil, err
	}
	err = s.svcCtx.DB.Order("id desc").Find(&apis).Error
	if parentAuthorityID == 0 || !s.svcCtx.Config.System.UseStrictAuth {
		return
	}
	paths := s.casbinService.GetPolicyPathByAuthorityId(authorityID)
	// 挑选 apis里面的path和method也在paths里面的api
	var authApis []systemModel.SysApi
	for i := range apis {
		for j := range paths {
			if paths[j].Path == apis[i].Path && paths[j].Method == apis[i].Method {
				authApis = append(authApis, apis[i])
			}
		}
	}
	return authApis, err
}

// GetApiById 根据id获取api
func (s *ApiService) GetApiById(id int) (api systemModel.SysApi, err error) {
	err = s.svcCtx.DB.First(&api, "id = ?", id).Error
	return
}

// UpdateApi 根据id更新api
func (s *ApiService) UpdateApi(api systemModel.SysApi) (err error) {
	var oldA systemModel.SysApi
	err = s.svcCtx.DB.First(&oldA, "id = ?", api.ID).Error
	if oldA.Path != api.Path || oldA.Method != api.Method {
		var duplicateApi systemModel.SysApi
		if ferr := s.svcCtx.DB.First(&duplicateApi, "path = ? AND method = ?", api.Path, api.Method).Error; ferr != nil {
			if !errors.Is(ferr, gorm.ErrRecordNotFound) {
				return ferr
			}
		} else {
			if duplicateApi.ID != api.ID {
				return errors.New("存在相同api路径")
			}
		}

	}
	if err != nil {
		return err
	}

	err = s.casbinService.UpdateCasbinApi(oldA.Path, api.Path, oldA.Method, api.Method)
	if err != nil {
		return err
	}

	return s.svcCtx.DB.Save(&api).Error
}

// DeleteApisByIds 删除选中API
func (s *ApiService) DeleteApisByIds(ids request.IdsReq) (err error) {
	return s.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		var apis []systemModel.SysApi
		err = tx.Find(&apis, "id in ?", ids.Ids).Error
		if err != nil {
			return err
		}
		err = tx.Delete(&[]systemModel.SysApi{}, "id in ?", ids.Ids).Error
		if err != nil {
			return err
		}
		for _, sysApi := range apis {
			s.casbinService.ClearCasbin(1, sysApi.Path, sysApi.Method)
		}
		return err
	})
}
