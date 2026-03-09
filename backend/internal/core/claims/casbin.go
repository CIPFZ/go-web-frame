package claims

import (
	"sync"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

// 1. 将单例变量定义在函数外部（包级别）
var (
	cachedEnforcer *casbin.SyncedCachedEnforcer
	once           sync.Once
)

// InitCasbin 初始化 Casbin Enforcer (单例模式)
func InitCasbin(db *gorm.DB) *casbin.SyncedCachedEnforcer {
	once.Do(func() {
		// ✨ 2. 适配自定义表名 "sys_casbin_rules"
		// 如果不这样做，它会去读 casbin_rule 表，导致权限丢失
		a, err := gormadapter.NewAdapterByDBUseTableName(db, "sys_", "casbin_rules")
		if err != nil {
			// ✨ 3. 启动阶段的关键组件失败，应该直接 panic
			// 这样运维人员能立刻知道服务没起动起来，而不是起动了一个“无权限”的残废服务
			panic("Casbin适配数据库失败: " + err.Error())
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
			panic("Casbin模型加载失败: " + err.Error())
		}

		// 使用 SyncedCachedEnforcer (支持并发安全 + 缓存)
		cachedEnforcer, err = casbin.NewSyncedCachedEnforcer(m, a)
		if err != nil {
			panic("Casbin Enforcer初始化失败: " + err.Error())
		}

		cachedEnforcer.EnableAutoSave(true)

		// 设置缓存过期时间 (防止权限修改后长时间不生效)
		// 生产环境建议 10-30 分钟，或者在修改权限时手动调用 LoadPolicy
		cachedEnforcer.SetExpireTime(60 * time.Minute)

		// 开启自动加载策略 (可选，如果有其他实例修改数据库，这里可以自动同步)
		cachedEnforcer.StartAutoLoadPolicy(5 * time.Second)

		// 初始加载
		if err := cachedEnforcer.LoadPolicy(); err != nil {
			panic("Casbin策略加载失败: " + err.Error())
		}
	})

	return cachedEnforcer
}
