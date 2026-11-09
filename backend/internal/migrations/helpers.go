package migrations

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/CIPFZ/gowebframe/internal/core/claims"
	"github.com/CIPFZ/gowebframe/internal/core/config"
	sysModel "github.com/CIPFZ/gowebframe/internal/modules/system/model"
)

// step 是一次性迁移：执行成功后在 sys_schema_migrations 记下 Version，之后不再执行。
// MySQL 的 DDL 会隐式提交，Run 必须可以重复执行（中途失败后下次从头再跑这一步）。
type step struct {
	Version string
	Run     func(db *gorm.DB, cfg *config.Config, logger *zap.Logger) error
}

func applied(db *gorm.DB, version string) (bool, error) {
	var count int64
	err := db.Model(&schemaMigration{}).Where("name = ?", version).Count(&count).Error
	return count > 0, err
}

// runSteps 按顺序执行还没执行过的迁移，每一步成功后单独记录版本号。
func runSteps(db *gorm.DB, cfg *config.Config, logger *zap.Logger, steps []step) error {
	for _, s := range steps {
		done, err := applied(db, s.Version)
		if err != nil {
			return err
		}
		if done {
			continue
		}
		if err := s.Run(db, cfg, logger); err != nil {
			return fmt.Errorf("migration %s: %w", s.Version, err)
		}
		if err := db.Create(&schemaMigration{Name: s.Version, AppliedAt: time.Now()}).Error; err != nil {
			return err
		}
	}
	return nil
}

// ensureMenu 按 path 查找菜单，不存在时创建。已有的菜单原样返回：管理员改过的名称、图标、排序都保留。
func ensureMenu(db *gorm.DB, want sysModel.SysMenu) (*sysModel.SysMenu, bool, error) {
	var menu sysModel.SysMenu
	err := db.Where("path = ?", want.Path).First(&menu).Error
	if err == nil {
		return &menu, false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}
	if err := db.Create(&want).Error; err != nil {
		return nil, false, err
	}
	return &want, true, nil
}

// grantMenu 把菜单授予这些角色；不存在的角色跳过。
func grantMenu(db *gorm.DB, menuID uint, authorities ...uint) error {
	for _, a := range authorities {
		exists, err := authorityExists(db, a)
		if err != nil {
			return err
		}
		if !exists {
			continue
		}
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&sysModel.SysAuthorityMenu{MenuId: menuID, AuthorityId: a}).Error; err != nil {
			return err
		}
	}
	return nil
}

// removeMenu 删除菜单及其授权（菜单下线时用）。
func removeMenu(db *gorm.DB, path string) error {
	var ids []uint
	if err := db.Model(&sysModel.SysMenu{}).Where("path = ?", path).Pluck("id", &ids).Error; err != nil || len(ids) == 0 {
		return err
	}
	if err := db.Where("menu_id IN ?", ids).Delete(&sysModel.SysAuthorityMenu{}).Error; err != nil {
		return err
	}
	return db.Unscoped().Where("id IN ?", ids).Delete(&sysModel.SysMenu{}).Error
}

// ensureAPI 按 path + method 登记接口，不存在时创建；已有的原样返回。
func ensureAPI(db *gorm.DB, want sysModel.SysApi) (sysModel.SysApi, error) {
	api := want
	err := db.Where(sysModel.SysApi{Path: want.Path, Method: want.Method}).Attrs(want).FirstOrCreate(&api).Error
	return api, err
}

// grantAPI 授予角色接口（API 目录关联 + Casbin 规则）；不存在的角色跳过。
// 必须在 claims.PolicyTransaction 里调用，运行中的实例才会重新加载策略。
func grantAPI(tx *gorm.DB, authorityID uint, api sysModel.SysApi) error {
	exists, err := authorityExists(tx, authorityID)
	if err != nil || !exists {
		return err
	}
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&sysModel.SysAuthorityApi{AuthorityId: authorityID, ApiId: api.ID}).Error; err != nil {
		return err
	}
	rule := sysModel.SysCasbinRule{Ptype: "p", V0: strconv.FormatUint(uint64(authorityID), 10), V1: api.Path, V2: api.Method}
	return tx.Where(&rule).FirstOrCreate(&rule).Error
}

// grantAccess 登记接口，并把菜单与接口授予这些角色（不存在的角色跳过）；授权和策略版本在同一事务里提交。
func grantAccess(db *gorm.DB, roles []uint, menuIDs []uint, apis []sysModel.SysApi) error {
	if err := ensurePolicyRevision(db); err != nil {
		return err
	}
	return claims.PolicyTransaction(db.Statement.Context, db, func(tx *gorm.DB) error {
		registered := make([]sysModel.SysApi, 0, len(apis))
		for _, want := range apis {
			api, err := ensureAPI(tx, want)
			if err != nil {
				return err
			}
			registered = append(registered, api)
		}
		for _, role := range roles {
			for _, id := range menuIDs {
				if err := grantMenu(tx, id, role); err != nil {
					return err
				}
			}
			for _, api := range registered {
				if err := grantAPI(tx, role, api); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// rolesWithAPI 返回已授予某接口的角色，用于“新接口授予能调用某个旧接口的角色”。
func rolesWithAPI(tx *gorm.DB, path, method string) ([]uint, error) {
	var roles []uint
	err := tx.Model(&sysModel.SysAuthorityApi{}).Joins("JOIN sys_apis ON sys_apis.id = sys_authority_apis.api_id").
		Where("sys_apis.path = ? AND sys_apis.method = ?", path, method).Pluck("sys_authority_apis.authority_id", &roles).Error
	return roles, err
}

func authorityExists(db *gorm.DB, authorityID uint) (bool, error) {
	var n int64
	err := db.Model(&sysModel.SysAuthority{}).Where("authority_id = ?", authorityID).Count(&n).Error
	return n > 0, err
}

// ensurePolicyRevision 确保策略版本表与它唯一的一行存在（PolicyTransaction 递增它）。
func ensurePolicyRevision(db *gorm.DB) error {
	if err := db.AutoMigrate(&claims.PolicyRevision{}); err != nil {
		return err
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&claims.PolicyRevision{ID: 1, Version: 1}).Error
}

// tokenSeedGroup 是 seed.EnsureTokenEndpoints 登记 Token 可调用接口时使用的分组。
const tokenSeedGroup = "external"

// catalogEntry 是模块接口表里的一个接口，以及新建时默认授予的角色。
type catalogEntry struct {
	Path, Method, Group, Name string
	Authorities               []uint
}

// syncAPICatalog 使 API 目录与一个模块（路径前缀 modulePrefix）的接口表一致：新增接口并按默认角色授权、
// 更新分组与说明、删除已下线的接口及其授权。可以每次迁移都执行（幂等），不修改其他模块的接口与授权。
// 已存在的接口不再补授权：管理员撤销的授权保持撤销。规则与策略版本在同一事务里提交。
func syncAPICatalog(db *gorm.DB, modulePrefix string, entries []catalogEntry) error {
	if err := ensurePolicyRevision(db); err != nil {
		return err
	}
	return claims.PolicyTransaction(db.Statement.Context, db, func(tx *gorm.DB) error {
		current := make(map[string]bool, len(entries))
		for _, e := range entries {
			current[e.Method+" "+e.Path] = true
			var api sysModel.SysApi
			res := tx.Where("path = ? AND method = ?", e.Path, e.Method).Limit(1).Find(&api)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected > 0 {
				// Token 白名单会先以 external 分组登记接口；首次被模块目录接管时按新接口授权
				adopted := api.ApiGroup != tokenSeedGroup
				if api.ApiGroup != e.Group || api.Description != e.Name {
					if err := tx.Model(&api).Updates(map[string]any{"api_group": e.Group, "description": e.Name}).Error; err != nil {
						return err
					}
				}
				if adopted {
					continue
				}
			} else {
				api = sysModel.SysApi{Path: e.Path, Method: e.Method, ApiGroup: e.Group, Description: e.Name}
				if err := tx.Create(&api).Error; err != nil {
					return err
				}
			}
			for _, a := range e.Authorities {
				if err := grantAPI(tx, a, api); err != nil {
					return err
				}
			}
		}

		var stale []sysModel.SysApi
		if err := tx.Where("path LIKE ?", strings.TrimRight(modulePrefix, "/")+"/%").Find(&stale).Error; err != nil {
			return err
		}
		for _, api := range stale {
			if current[api.Method+" "+api.Path] {
				continue
			}
			if err := tx.Where("api_id = ?", api.ID).Delete(&sysModel.SysAuthorityApi{}).Error; err != nil {
				return err
			}
			if err := tx.Where("api_id = ?", api.ID).Delete(&sysModel.SysApiTokenApi{}).Error; err != nil {
				return err
			}
			if err := tx.Where("ptype = ? AND v1 = ? AND v2 = ?", "p", api.Path, api.Method).Delete(&sysModel.SysCasbinRule{}).Error; err != nil {
				return err
			}
			if err := tx.Unscoped().Delete(&api).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
