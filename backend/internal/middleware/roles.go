package middleware

import (
	"context"
	"slices"
	"strconv"

	"gorm.io/gorm"

	"github.com/CIPFZ/gowebframe/internal/svc"
)

// SuperAdminAuthorityID 是超级管理员角色：不经 Casbin 判断，拥有全部接口。
const SuperAdminAuthorityID = 1

// UserRoles 返回用户拥有的全部角色（升序）。一个用户可以有多个角色，它们同时生效，权限取并集。
// 子角色只能拥有父角色权限的一部分（分配时校验，见 system/repository/role_subset.go），不继承父角色：只算用户直接拥有的角色。
func UserRoles(ctx context.Context, db *gorm.DB, userID uint) ([]uint, error) {
	var roles []uint
	if err := db.WithContext(ctx).Table("sys_user_authorities AS ua").
		Joins("JOIN sys_authorities AS a ON a.authority_id = ua.authority_id AND a.deleted_at IS NULL").
		Where("ua.user_id = ?", userID).Order("ua.authority_id").Pluck("ua.authority_id", &roles).Error; err != nil {
		return nil, err
	}
	return slices.Compact(roles), nil
}

// RolesAllow 判断这些角色中是否有一个能调用某接口。
func RolesAllow(ctx context.Context, svcCtx *svc.ServiceContext, roles []uint, path, method string) (bool, error) {
	if slices.Contains(roles, SuperAdminAuthorityID) {
		return true, nil
	}
	for _, role := range roles {
		ok, err := svcCtx.Policy.Enforce(ctx, strconv.FormatUint(uint64(role), 10), path, method)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

// UserAllows 判断用户现在能否调用某接口。
func UserAllows(ctx context.Context, svcCtx *svc.ServiceContext, userID uint, path, method string) (bool, error) {
	roles, err := UserRoles(ctx, svcCtx.DB, userID)
	if err != nil {
		return false, err
	}
	return RolesAllow(ctx, svcCtx, roles, path, method)
}
