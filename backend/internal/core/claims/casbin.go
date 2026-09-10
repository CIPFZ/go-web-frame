package claims

import (
	"fmt"
	"time"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

// InitCasbin owns one enforcer per service context; failures are returned to the
// startup cleanup path instead of panicking or sharing global state.
func InitCasbin(db *gorm.DB) (*casbin.SyncedCachedEnforcer, error) {
	// ✨ 2. 适配自定义表名 "sys_casbin_rules"
	// 如果不这样做，它会去读 casbin_rule 表，导致权限丢失
	a, err := gormadapter.NewAdapterByDBUseTableName(db, "sys_", "casbin_rules")
	if err != nil {
		// Return startup errors so the caller can close initialized resources.
		return nil, fmt.Errorf("casbin adapter: %w", err)
	}

	// 定义 RBAC 模型
	// sub: 角色ID (string)
	// obj: URL路径 (string)
	// act: HTTP方法 (string)
	text := `
		[request_definition]
		r = sub, obj, act

		[policy_definition]
		p = sub, obj, act

		[role_definition]
		g = _, _

		[policy_effect]
		e = some(where (p.eft == allow))

		[matchers]
		# keyMatch2 支持 /api/v1/user/:id 这种路径匹配
		m = r.sub == p.sub && keyMatch2(r.obj, p.obj) && r.act == p.act
		`

	m, err := model.NewModelFromString(text)
	if err != nil {
		return nil, fmt.Errorf("casbin model: %w", err)
	}

	// 使用 SyncedCachedEnforcer (支持并发安全 + 缓存)
	cachedEnforcer, err := casbin.NewSyncedCachedEnforcer(m, a)
	if err != nil {
		return nil, fmt.Errorf("casbin enforcer: %w", err)
	}

	cachedEnforcer.EnableAutoSave(true)

	// 设置缓存过期时间 (防止权限修改后长时间不生效)
	// 生产环境建议 10-30 分钟，或者在修改权限时手动调用 LoadPolicy
	cachedEnforcer.SetExpireTime(60 * time.Minute)

	// 初始加载
	if err := cachedEnforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("casbin policy: %w", err)
	}
	cachedEnforcer.StartAutoLoadPolicy(5 * time.Second)
	return cachedEnforcer, nil
}
