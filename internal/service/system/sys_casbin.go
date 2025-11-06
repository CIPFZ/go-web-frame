package system

import (
	"errors"
	"strconv"

	systemReq "github.com/CIPFZ/gowebframe/internal/model/system/request"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/internal/utils"
	gormadapter "github.com/casbin/gorm-adapter/v3"

	"gorm.io/gorm"
)

type ICasbinService interface {
	UpdateCasbin(adminAuthorityID, AuthorityID uint, casbinInfos []systemReq.CasbinInfo) error
	UpdateCasbinApi(oldPath string, newPath string, oldMethod string, newMethod string) error
	GetPolicyPathByAuthorityId(AuthorityID uint) (pathMaps []systemReq.CasbinInfo)
	RemoveFilteredPolicy(db *gorm.DB, authorityId string) error
	ClearCasbin(v int, p ...string) bool
	SyncPolicy(db *gorm.DB, authorityId string, rules [][]string) error
	AddPolicies(db *gorm.DB, rules [][]string) error
	FreshCasbin() (err error)
}

type CasbinService struct {
	svcCtx *svc.ServiceContext
}

func NewCasbinService(svcCtx *svc.ServiceContext) ICasbinService {
	return &CasbinService{
		svcCtx: svcCtx,
	}
}

// UpdateCasbin 更新casbin权限
func (s *CasbinService) UpdateCasbin(adminAuthorityID, AuthorityID uint, casbinInfos []systemReq.CasbinInfo) error {

	authorityService := NewAuthorityService(s.svcCtx)
	err := authorityService.CheckAuthorityIDAuth(adminAuthorityID, AuthorityID)
	if err != nil {
		return err
	}

	if s.svcCtx.Config.System.UseStrictAuth {
		apiService := NewApiService(s.svcCtx)
		apis, e := apiService.GetAllApis(adminAuthorityID)
		if e != nil {
			return e
		}

		for i := range casbinInfos {
			hasApi := false
			for j := range apis {
				if apis[j].Path == casbinInfos[i].Path && apis[j].Method == casbinInfos[i].Method {
					hasApi = true
					break
				}
			}
			if !hasApi {
				return errors.New("存在api不在权限列表中")
			}
		}
	}

	authorityId := strconv.Itoa(int(AuthorityID))
	s.ClearCasbin(0, authorityId)
	rules := [][]string{}
	//做权限去重处理
	deduplicateMap := make(map[string]bool)
	for _, v := range casbinInfos {
		key := authorityId + v.Path + v.Method
		if _, ok := deduplicateMap[key]; !ok {
			deduplicateMap[key] = true
			rules = append(rules, []string{authorityId, v.Path, v.Method})
		}
	}
	if len(rules) == 0 {
		return nil
	} // 设置空权限无需调用 AddPolicies 方法
	e := utils.GetCasbin(s.svcCtx)
	success, _ := e.AddPolicies(rules)
	if !success {
		return errors.New("存在相同api,添加失败,请联系管理员")
	}
	return nil
}

// UpdateCasbinApi  API更新随动
func (s *CasbinService) UpdateCasbinApi(oldPath string, newPath string, oldMethod string, newMethod string) error {
	err := s.svcCtx.DB.Model(&gormadapter.CasbinRule{}).Where("v1 = ? AND v2 = ?", oldPath, oldMethod).Updates(map[string]interface{}{
		"v1": newPath,
		"v2": newMethod,
	}).Error
	if err != nil {
		return err
	}

	e := utils.GetCasbin(s.svcCtx)
	return e.LoadPolicy()
}

// GetPolicyPathByAuthorityId 获取权限列表
func (s *CasbinService) GetPolicyPathByAuthorityId(AuthorityID uint) (pathMaps []systemReq.CasbinInfo) {
	e := utils.GetCasbin(s.svcCtx)
	authorityId := strconv.Itoa(int(AuthorityID))
	list, _ := e.GetFilteredPolicy(0, authorityId)
	for _, v := range list {
		pathMaps = append(pathMaps, systemReq.CasbinInfo{
			Path:   v[1],
			Method: v[2],
		})
	}
	return pathMaps
}

// ClearCasbin 清除匹配的权限
func (s *CasbinService) ClearCasbin(v int, p ...string) bool {
	e := utils.GetCasbin(s.svcCtx)
	success, _ := e.RemoveFilteredPolicy(v, p...)
	return success
}

// RemoveFilteredPolicy 使用数据库方法清理筛选的politicy 此方法需要调用FreshCasbin方法才可以在系统中即刻生效
func (s *CasbinService) RemoveFilteredPolicy(db *gorm.DB, authorityId string) error {
	return db.Delete(&gormadapter.CasbinRule{}, "v0 = ?", authorityId).Error
}

// SyncPolicy 同步目前数据库的policy 此方法需要调用FreshCasbin方法才可以在系统中即刻生效
func (s *CasbinService) SyncPolicy(db *gorm.DB, authorityId string, rules [][]string) error {
	err := s.RemoveFilteredPolicy(db, authorityId)
	if err != nil {
		return err
	}
	return s.AddPolicies(db, rules)
}

// AddPolicies 添加匹配的权限
func (s *CasbinService) AddPolicies(db *gorm.DB, rules [][]string) error {
	var casbinRules []gormadapter.CasbinRule
	for i := range rules {
		casbinRules = append(casbinRules, gormadapter.CasbinRule{
			Ptype: "p",
			V0:    rules[i][0],
			V1:    rules[i][1],
			V2:    rules[i][2],
		})
	}
	return db.Create(&casbinRules).Error
}

func (s *CasbinService) FreshCasbin() (err error) {
	e := utils.GetCasbin(s.svcCtx)
	err = e.LoadPolicy()
	return err
}
