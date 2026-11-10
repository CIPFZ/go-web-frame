# 数据库迁移

迁移在 `backend/internal/migrations`，由 `cmd/migrate` 执行；服务启动前用 `migrations.Check` 确认已经迁移到 `Latest`。
多个实例同时执行迁移时由数据库锁串行化（MySQL `GET_LOCK`、PostgreSQL advisory lock、SQLite 事务）。

## 新增一次迁移

在 `runner.go` 的 `steps` 末尾追加一步，版本号用“日期_内容_v序号”：

```go
var steps = []step{
	{Version: "20261201_orders_v1", Run: migrateOrders},
}
```

- 每一步成功后单独记录版本号，已记录的不再执行；失败的步骤不记录，下次从这一步重新开始。
- MySQL 的 DDL 会隐式提交，所以 `Run` 必须可以重复执行（先判断列、表是否存在等）。
- 已发布的步骤不要修改或调换顺序；要改行为就再追加一步。

## 菜单与授权

`helpers.go` 提供下面这些函数，业务模块的迁移用它们登记菜单和接口：

| 函数 | 作用 |
|---|---|
| `ensureMenu` | 按 path 创建菜单；已有的原样保留（管理员改过的名称、图标、排序不会被覆盖） |
| `grantMenu` / `removeMenu` | 授予菜单（不存在的角色跳过）/ 删除菜单及其授权 |
| `ensureAPI` | 按 path + method 登记接口 |
| `grantAPI` | 授予角色接口（API 目录关联 + Casbin 规则），**必须在 `claims.PolicyTransaction` 里调用** |
| `rolesWithAPI` | 已授予某接口的角色，用于“新接口授予能调用某个旧接口的角色” |
| `syncAPICatalog` | 让 API 目录与模块的接口表保持一致：新接口按默认角色授权，已有接口不补授权（管理员撤销的保持撤销），下线的接口连同授权一起删除；可以每次迁移都执行 |

授权变更必须和策略版本在同一个事务里提交（`claims.PolicyTransaction`）。运行中的实例每次鉴权都会比较版本号，版本变了才会重新加载 Casbin 策略；直接写 `sys_casbin_rules` 而不递增版本，其他实例要等重启才能看到。
