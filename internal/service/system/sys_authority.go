package system

import (
	"errors"
	"strconv"

	"github.com/CIPFZ/gowebframe/internal/model/common/request"
	systemModel "github.com/CIPFZ/gowebframe/internal/model/system"
	systemReq "github.com/CIPFZ/gowebframe/internal/model/system/request"
	systemRes "github.com/CIPFZ/gowebframe/internal/model/system/response"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/logger"
	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
)

var ErrRoleExistence = errors.New("存在相同角色id")

type IAuthorityService interface {
	CreateAuthority(auth systemModel.SysAuthority) (authority systemModel.SysAuthority, err error)
	CopyAuthority(adminAuthorityID uint, copyInfo systemRes.SysAuthorityCopyResponse) (authority systemModel.SysAuthority, err error)
	UpdateAuthority(c *gin.Context, auth systemModel.SysAuthority) (authority systemModel.SysAuthority, err error)
	DeleteAuthority(auth *systemModel.SysAuthority) error
	GetAuthorityInfoList(authorityID uint) (list []systemModel.SysAuthority, err error)
	GetStructAuthorityList(authorityID uint) (list []uint, err error)
	CheckAuthorityIDAuth(authorityID, targetID uint) (err error)
	GetAuthorityInfo(auth systemModel.SysAuthority) (sa systemModel.SysAuthority, err error)
	SetDataAuthority(adminAuthorityID uint, auth systemModel.SysAuthority) (err error)
	SetMenuAuthority(auth *systemModel.SysAuthority) error
	findChildrenAuthority(authority *systemModel.SysAuthority) (err error)
	GetParentAuthorityID(authorityID uint) (parentID uint, err error)
}

type AuthorityService struct {
	svcCtx *svc.ServiceContext
}

func NewAuthorityService(svcCtx *svc.ServiceContext) IAuthorityService {
	return &AuthorityService{
		svcCtx: svcCtx,
	}
}

// CreateAuthority 创建一个角色
func (s *AuthorityService) CreateAuthority(auth systemModel.SysAuthority) (authority systemModel.SysAuthority, err error) {

	if err = s.svcCtx.DB.Where("authority_id = ?", auth.AuthorityId).First(&systemModel.SysAuthority{}).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		return auth, ErrRoleExistence
	}

	e := s.svcCtx.DB.Transaction(func(tx *gorm.DB) error {

		if err = tx.Create(&auth).Error; err != nil {
			return err
		}

		auth.SysBaseMenus = systemReq.DefaultMenu()
		if err = tx.Model(&auth).Association("SysBaseMenus").Replace(&auth.SysBaseMenus); err != nil {
			return err
		}
		casbinInfos := systemReq.DefaultCasbin()
		authorityId := strconv.Itoa(int(auth.AuthorityId))
		rules := [][]string{}
		for _, v := range casbinInfos {
			rules = append(rules, []string{authorityId, v.Path, v.Method})
		}
		return NewCasbinService(s.svcCtx).AddPolicies(tx, rules)
	})

	return auth, e
}

// CopyAuthority 复制一个角色
func (s *AuthorityService) CopyAuthority(adminAuthorityID uint, copyInfo systemRes.SysAuthorityCopyResponse) (authority systemModel.SysAuthority, err error) {
	var authorityBox systemModel.SysAuthority
	if !errors.Is(s.svcCtx.DB.Where("authority_id = ?", copyInfo.Authority.AuthorityId).First(&authorityBox).Error, gorm.ErrRecordNotFound) {
		return authority, ErrRoleExistence
	}
	copyInfo.Authority.Children = []systemModel.SysAuthority{}
	menuService := NewMenuService(s.svcCtx)
	menus, err := menuService.GetMenuAuthority(&request.GetAuthorityId{AuthorityId: copyInfo.OldAuthorityId})
	if err != nil {
		return
	}
	var baseMenu []systemModel.SysBaseMenu
	for _, v := range menus {
		intNum := v.MenuId
		v.SysBaseMenu.ID = uint(intNum)
		baseMenu = append(baseMenu, v.SysBaseMenu)
	}
	copyInfo.Authority.SysBaseMenus = baseMenu
	err = s.svcCtx.DB.Create(&copyInfo.Authority).Error
	if err != nil {
		return
	}

	var btns []systemModel.SysAuthorityBtn

	err = s.svcCtx.DB.Find(&btns, "authority_id = ?", copyInfo.OldAuthorityId).Error
	if err != nil {
		return
	}
	if len(btns) > 0 {
		for i := range btns {
			btns[i].AuthorityId = copyInfo.Authority.AuthorityId
		}
		err = s.svcCtx.DB.Create(&btns).Error

		if err != nil {
			return
		}
	}
	casbinService := NewCasbinService(s.svcCtx)
	paths := casbinService.GetPolicyPathByAuthorityId(copyInfo.OldAuthorityId)
	err = casbinService.UpdateCasbin(adminAuthorityID, copyInfo.Authority.AuthorityId, paths)
	if err != nil {
		_ = s.DeleteAuthority(&copyInfo.Authority)
	}
	return copyInfo.Authority, err
}

// UpdateAuthority 更改一个角色
func (s *AuthorityService) UpdateAuthority(c *gin.Context, auth systemModel.SysAuthority) (authority systemModel.SysAuthority, err error) {
	var oldAuthority systemModel.SysAuthority
	err = s.svcCtx.DB.Where("authority_id = ?", auth.AuthorityId).First(&oldAuthority).Error
	log := logger.GetLogger(c)
	if err != nil {
		log.Debug(err.Error())
		return systemModel.SysAuthority{}, errors.New("查询角色数据失败")
	}
	err = s.svcCtx.DB.Model(&oldAuthority).Updates(&auth).Error
	return auth, err
}

// DeleteAuthority 删除角色
func (s *AuthorityService) DeleteAuthority(auth *systemModel.SysAuthority) error {
	if errors.Is(s.svcCtx.DB.Debug().Preload("Users").First(&auth).Error, gorm.ErrRecordNotFound) {
		return errors.New("该角色不存在")
	}
	if len(auth.Users) != 0 {
		return errors.New("此角色有用户正在使用禁止删除")
	}
	if !errors.Is(s.svcCtx.DB.Where("authority_id = ?", auth.AuthorityId).First(&systemModel.SysUser{}).Error, gorm.ErrRecordNotFound) {
		return errors.New("此角色有用户正在使用禁止删除")
	}
	if !errors.Is(s.svcCtx.DB.Where("parent_id = ?", auth.AuthorityId).First(&systemModel.SysAuthority{}).Error, gorm.ErrRecordNotFound) {
		return errors.New("此角色存在子角色不允许删除")
	}

	return s.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		if err = tx.Preload("SysBaseMenus").Preload("DataAuthorityId").Where("authority_id = ?", auth.AuthorityId).First(auth).Unscoped().Delete(auth).Error; err != nil {
			return err
		}

		if len(auth.SysBaseMenus) > 0 {
			if err = tx.Model(auth).Association("SysBaseMenus").Delete(auth.SysBaseMenus); err != nil {
				return err
			}
			// err = db.Association("SysBaseMenus").Delete(&auth)
		}
		if len(auth.DataAuthorityId) > 0 {
			if err = tx.Model(auth).Association("DataAuthorityId").Delete(auth.DataAuthorityId); err != nil {
				return err
			}
		}

		if err = tx.Delete(&systemModel.SysUserAuthority{}, "sys_authority_authority_id = ?", auth.AuthorityId).Error; err != nil {
			return err
		}
		if err = tx.Where("authority_id = ?", auth.AuthorityId).Delete(&[]systemModel.SysAuthorityBtn{}).Error; err != nil {
			return err
		}

		authorityId := strconv.Itoa(int(auth.AuthorityId))

		casbinService := NewCasbinService(s.svcCtx)
		if err = casbinService.RemoveFilteredPolicy(tx, authorityId); err != nil {
			return err
		}

		return nil
	})
}

// GetAuthorityInfoList 分页获取数据
func (s *AuthorityService) GetAuthorityInfoList(authorityID uint) (list []systemModel.SysAuthority, err error) {
	var authority systemModel.SysAuthority
	err = s.svcCtx.DB.Where("authority_id = ?", authorityID).First(&authority).Error
	if err != nil {
		return nil, err
	}
	var authorities []systemModel.SysAuthority
	db := s.svcCtx.DB.Model(&systemModel.SysAuthority{})
	if s.svcCtx.Config.System.UseStrictAuth {
		// 当开启了严格树形结构后
		if *authority.ParentId == 0 {
			// 只有顶级角色可以修改自己的权限和以下权限
			err = db.Preload("DataAuthorityId").Where("authority_id = ?", authorityID).Find(&authorities).Error
		} else {
			// 非顶级角色只能修改以下权限
			err = db.Debug().Preload("DataAuthorityId").Where("parent_id = ?", authorityID).Find(&authorities).Error
		}
	} else {
		err = db.Preload("DataAuthorityId").Where("parent_id = ?", "0").Find(&authorities).Error
	}

	for k := range authorities {
		err = s.findChildrenAuthority(&authorities[k])
	}
	return authorities, err
}

func (s *AuthorityService) GetStructAuthorityList(authorityID uint) (list []uint, err error) {
	var auth systemModel.SysAuthority
	_ = s.svcCtx.DB.First(&auth, "authority_id = ?", authorityID).Error
	var authorities []systemModel.SysAuthority
	err = s.svcCtx.DB.Preload("DataAuthorityId").Where("parent_id = ?", authorityID).Find(&authorities).Error
	if len(authorities) > 0 {
		for k := range authorities {
			list = append(list, authorities[k].AuthorityId)
			childrenList, err := s.GetStructAuthorityList(authorities[k].AuthorityId)
			if err == nil {
				list = append(list, childrenList...)
			}
		}
	}
	if *auth.ParentId == 0 {
		list = append(list, authorityID)
	}
	return list, err
}

func (s *AuthorityService) CheckAuthorityIDAuth(authorityID, targetID uint) (err error) {
	if !s.svcCtx.Config.System.UseStrictAuth {
		return nil
	}
	authIDS, err := s.GetStructAuthorityList(authorityID)
	if err != nil {
		return err
	}
	hasAuth := false
	for _, v := range authIDS {
		if v == targetID {
			hasAuth = true
			break
		}
	}
	if !hasAuth {
		return errors.New("您提交的角色ID不合法")
	}
	return nil
}

// GetAuthorityInfo 获取所有角色信息
func (s *AuthorityService) GetAuthorityInfo(auth systemModel.SysAuthority) (sa systemModel.SysAuthority, err error) {
	err = s.svcCtx.DB.Preload("DataAuthorityId").Where("authority_id = ?", auth.AuthorityId).First(&sa).Error
	return sa, err
}

// SetDataAuthority 设置角色资源权限
func (s *AuthorityService) SetDataAuthority(adminAuthorityID uint, auth systemModel.SysAuthority) error {
	var checkIDs []uint
	checkIDs = append(checkIDs, auth.AuthorityId)
	for i := range auth.DataAuthorityId {
		checkIDs = append(checkIDs, auth.DataAuthorityId[i].AuthorityId)
	}

	for i := range checkIDs {
		err := s.CheckAuthorityIDAuth(adminAuthorityID, checkIDs[i])
		if err != nil {
			return err
		}
	}

	var res systemModel.SysAuthority
	s.svcCtx.DB.Preload("DataAuthorityId").First(&res, "authority_id = ?", auth.AuthorityId)
	err := s.svcCtx.DB.Model(&s).Association("DataAuthorityId").Replace(&auth.DataAuthorityId)
	return err
}

// SetMenuAuthority 菜单与角色绑定
func (s *AuthorityService) SetMenuAuthority(auth *systemModel.SysAuthority) error {
	var res systemModel.SysAuthority
	s.svcCtx.DB.Preload("SysBaseMenus").First(&res, "authority_id = ?", auth.AuthorityId)
	err := s.svcCtx.DB.Model(&s).Association("SysBaseMenus").Replace(&auth.SysBaseMenus)
	return err
}

// findChildrenAuthority 查询子角色
func (s *AuthorityService) findChildrenAuthority(authority *systemModel.SysAuthority) (err error) {
	err = s.svcCtx.DB.Preload("DataAuthorityId").Where("parent_id = ?", authority.AuthorityId).Find(&authority.Children).Error
	if len(authority.Children) > 0 {
		for k := range authority.Children {
			err = s.findChildrenAuthority(&authority.Children[k])
		}
	}
	return err
}

func (s *AuthorityService) GetParentAuthorityID(authorityID uint) (parentID uint, err error) {
	var authority systemModel.SysAuthority
	err = s.svcCtx.DB.Where("authority_id = ?", authorityID).First(&authority).Error
	if err != nil {
		return
	}
	return *authority.ParentId, nil
}
