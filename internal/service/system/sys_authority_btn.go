package system

import (
	"errors"

	systemModel "github.com/CIPFZ/gowebframe/internal/model/system"
	systemReq "github.com/CIPFZ/gowebframe/internal/model/system/request"
	systemRes "github.com/CIPFZ/gowebframe/internal/model/system/response"
	"github.com/CIPFZ/gowebframe/internal/svc"

	"gorm.io/gorm"
)

type IAuthorityBtnService interface {
	GetAuthorityBtn(req systemReq.SysAuthorityBtnReq) (res systemRes.SysAuthorityBtnRes, err error)
	SetAuthorityBtn(req systemReq.SysAuthorityBtnReq) (err error)
	CanRemoveAuthorityBtn(ID string) (err error)
}

type AuthorityBtnService struct {
	svcCtx *svc.ServiceContext
}

func NewAuthorityBtnService(svcCtx *svc.ServiceContext) IAuthorityBtnService {
	return &AuthorityBtnService{
		svcCtx: svcCtx,
	}
}

func (a *AuthorityBtnService) GetAuthorityBtn(req systemReq.SysAuthorityBtnReq) (res systemRes.SysAuthorityBtnRes, err error) {
	var authorityBtn []systemModel.SysAuthorityBtn
	err = a.svcCtx.DB.Find(&authorityBtn, "authority_id = ? and sys_menu_id = ?", req.AuthorityId, req.MenuID).Error
	if err != nil {
		return
	}
	var selected []uint
	for _, v := range authorityBtn {
		selected = append(selected, v.SysBaseMenuBtnID)
	}
	res.Selected = selected
	return res, err
}

func (a *AuthorityBtnService) SetAuthorityBtn(req systemReq.SysAuthorityBtnReq) (err error) {
	return a.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		var authorityBtn []systemModel.SysAuthorityBtn
		err = tx.Delete(&[]systemModel.SysAuthorityBtn{}, "authority_id = ? and sys_menu_id = ?", req.AuthorityId, req.MenuID).Error
		if err != nil {
			return err
		}
		for _, v := range req.Selected {
			authorityBtn = append(authorityBtn, systemModel.SysAuthorityBtn{
				AuthorityId:      req.AuthorityId,
				SysMenuID:        req.MenuID,
				SysBaseMenuBtnID: v,
			})
		}
		if len(authorityBtn) > 0 {
			err = tx.Create(&authorityBtn).Error
		}
		if err != nil {
			return err
		}
		return err
	})
}

func (a *AuthorityBtnService) CanRemoveAuthorityBtn(ID string) (err error) {
	fErr := a.svcCtx.DB.First(&systemModel.SysAuthorityBtn{}, "sys_base_menu_btn_id = ?", ID).Error
	if errors.Is(fErr, gorm.ErrRecordNotFound) {
		return nil
	}
	return errors.New("此按钮正在被使用无法删除")
}
