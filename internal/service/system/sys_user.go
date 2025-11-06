package system

import (
	"fmt"

	"github.com/CIPFZ/gowebframe/internal/model/system"
	"github.com/CIPFZ/gowebframe/internal/model/system/request"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/internal/utils"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

type IUserService interface {
	Login(ctx *gin.Context, u *system.SysUser) (*system.SysUser, error)
	GetUserInfoList(ctx *gin.Context, info request.GetUserList) (list interface{}, total int64, err error)
	GetUserInfo(ctx *gin.Context, uuid uuid.UUID) (user system.SysUser, err error)
}

type UserService struct {
	serviceCtx *svc.ServiceContext
}

func NewUserService(ctx *svc.ServiceContext) IUserService {
	return &UserService{serviceCtx: ctx}
}

func (s *UserService) Login(ctx *gin.Context, u *system.SysUser) (*system.SysUser, error) {
	if s.serviceCtx.DB == nil {
		return nil, fmt.Errorf("db not init")
	}

	var user system.SysUser
	err := s.serviceCtx.DB.Where("username = ?", u.Username).Preload("Authorities").Preload("Authority").First(&user).Error
	if err == nil {
		if ok := utils.BcryptCheck(u.Password, user.Password); !ok {
			return nil, fmt.Errorf("password error")
		}
		// TODO 用户角色默认路由检查
	}
	return &user, err
}

func (s *UserService) GetUserInfoList(ctx *gin.Context, info request.GetUserList) (list interface{}, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	db := s.serviceCtx.DB.Model(&system.SysUser{})
	var userList []system.SysUser

	if info.NickName != "" {
		db = db.Where("nick_name LIKE ?", "%"+info.NickName+"%")
	}
	if info.Phone != "" {
		db = db.Where("phone LIKE ?", "%"+info.Phone+"%")
	}
	if info.Username != "" {
		db = db.Where("username LIKE ?", "%"+info.Username+"%")
	}
	if info.Email != "" {
		db = db.Where("email LIKE ?", "%"+info.Email+"%")
	}

	err = db.Count(&total).Error
	if err != nil {
		return
	}
	err = db.Limit(limit).Offset(offset).Preload("Authorities").Preload("Authority").Find(&userList).Error
	return userList, total, err
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(ctx *gin.Context, uuid uuid.UUID) (user system.SysUser, err error) {
	var reqUser system.SysUser
	err = s.serviceCtx.DB.Preload("Authorities").Preload("Authority").First(&reqUser, "uuid = ?", uuid).Error
	if err != nil {
		return reqUser, err
	}
	// 获取菜单信息
	//NewMenuService(s.serviceCtx).UserAuthorityDefaultRouter(ctx, &reqUser)
	return reqUser, err
}
