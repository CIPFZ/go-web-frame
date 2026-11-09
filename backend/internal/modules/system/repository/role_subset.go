package repository

import (
	"errors"
	"strconv"

	"gorm.io/gorm"

	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
)

// 子角色是父角色权限的一部分：给子角色分配的接口、菜单只能从父角色已有的里面选；父角色收回的，各级子角色一起收回。
// 子角色不继承父角色的权限（拥有子角色的人只有子角色自己的权限）。超级管理员（1）拥有全部，它的子角色可以选任何权限。

const superAdminRole = 1

// parentAPIs 返回父角色的接口（“方法 路径”）；没有父角色或父角色是超级管理员时返回 nil（不限）。
func parentAPIs(tx *gorm.DB, role model.SysAuthority) (map[string]bool, error) {
	if role.ParentId == 0 || role.ParentId == superAdminRole {
		return nil, nil
	}
	var rules []model.SysCasbinRule
	if err := tx.Where("v0 = ? AND ptype = ?", strconv.FormatUint(uint64(role.ParentId), 10), "p").Find(&rules).Error; err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(rules))
	for _, r := range rules {
		out[r.V2+" "+r.V1] = true
	}
	return out, nil
}

// parentMenus 返回父角色的菜单；没有父角色或父角色是超级管理员时返回 nil（不限）。
func parentMenus(tx *gorm.DB, role model.SysAuthority) (map[uint]bool, error) {
	if role.ParentId == 0 || role.ParentId == superAdminRole {
		return nil, nil
	}
	var ids []uint
	if err := tx.Model(&model.SysAuthorityMenu{}).Where("authority_id = ?", role.ParentId).Pluck("menu_id", &ids).Error; err != nil {
		return nil, err
	}
	out := make(map[uint]bool, len(ids))
	for _, id := range ids {
		out[id] = true
	}
	return out, nil
}

// 子集校验失败的提示（已登记到响应词条，见 pkg/response/message_keys.go）
var (
	errMenusNotInParent = errors.New("子角色的菜单只能从父角色已有的里面选")
	errAPIsNotInParent  = errors.New("子角色的接口只能从父角色已有的里面选")
)

// pruneChildren 把各级子角色超出其父角色的接口与菜单收回（父角色改了权限之后）。
func pruneChildren(tx *gorm.DB, roleID uint) error {
	var children []model.SysAuthority
	if err := tx.Where("parent_id = ? AND deleted_at IS NULL", roleID).Find(&children).Error; err != nil {
		return err
	}
	for _, c := range children {
		if c.AuthorityId == roleID {
			return errors.New("角色不能是自身的父角色")
		}
		apis, err := parentAPIs(tx, c)
		if err != nil {
			return err
		}
		if apis != nil {
			subject := strconv.FormatUint(uint64(c.AuthorityId), 10)
			var rules []model.SysCasbinRule
			if err := tx.Where("v0 = ? AND ptype = ?", subject, "p").Find(&rules).Error; err != nil {
				return err
			}
			for _, r := range rules {
				if apis[r.V2+" "+r.V1] {
					continue
				}
				if err := tx.Delete(&r).Error; err != nil {
					return err
				}
				if err := tx.Where("authority_id = ? AND api_id IN (?)", c.AuthorityId,
					tx.Model(&model.SysApi{}).Select("id").Where("path = ? AND method = ?", r.V1, r.V2)).Delete(&model.SysAuthorityApi{}).Error; err != nil {
					return err
				}
			}
		}
		menus, err := parentMenus(tx, c)
		if err != nil {
			return err
		}
		if menus != nil {
			keep := make([]uint, 0, len(menus))
			for id := range menus {
				keep = append(keep, id)
			}
			q := tx.Where("authority_id = ?", c.AuthorityId)
			if len(keep) > 0 {
				q = q.Where("menu_id NOT IN ?", keep)
			}
			if err := q.Delete(&model.SysAuthorityMenu{}).Error; err != nil {
				return err
			}
		}
		if err := pruneChildren(tx, c.AuthorityId); err != nil {
			return err
		}
	}
	return nil
}

// PruneChildRoles 把所有子角色超出父角色的接口与菜单收回（升级到“子角色是父角色的子集”时执行一次）。
// 必须在 claims.PolicyTransaction 里调用。
func PruneChildRoles(tx *gorm.DB) error {
	var roots []uint
	if err := tx.Model(&model.SysAuthority{}).Where("parent_id = 0 AND deleted_at IS NULL").Pluck("authority_id", &roots).Error; err != nil {
		return err
	}
	for _, id := range roots {
		if err := pruneChildren(tx, id); err != nil {
			return err
		}
	}
	return nil
}
