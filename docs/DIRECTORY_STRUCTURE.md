# 推荐目录结构与演进规则

结论：当前项目适合**后端模块化单体 + 前端按业务组织**。按业务模块找到完整功能，再在模块内部区分 API、服务和数据访问。目录深度应跟随实际复杂度，基础 CMS 不需要每层都增加一套空接口或一个独立服务。

## 当前已落实的整理

- 后端已有 `internal/modules/system/{api,dto,service,repository,model,router,seed}`，保留这一合理分层。
- 前端动态路由相关文件从 `src/utils` 集中到 **`src/routing`**：菜单缓存、后端菜单转换、组件注册与图标适配放在一起。
- 前端业务请求从笼统的 `src/services/api` 改成 **`src/services/system`**，与后端 system 模块对应。HTTP 地址、后端组件标识和业务路由没有改变。
- 删除未注册、无外部引用的模板 `Admin`、`Welcome`、`profile`、`result`、重复 `exception` 页面；保留实际 404 和账号页面。删除工作台旧样例组件、请求、类型和样式，以及注册页未启用的 mock。
- 架构和维护文档放入 `docs`，CI 放入 `.github/workflows`；本机配置、备份和测试产物继续隔离在被 Git 忽略的 `deploy/local/runtime`。

## 后端

```text
backend/
├── cmd/
│   ├── server/                 # HTTP 进程入口、组装
│   ├── migrate/                # 现有迁移命令和版本脚本
│   └── worker/                 # Nexus 分支新增，独立执行进程
├── internal/
│   ├── modules/
│   │   ├── system/
│   │   │   ├── api/            # HTTP 参数/响应，调用 service
│   │   │   ├── dto/            # 模块接口输入输出
│   │   │   ├── service/        # 业务规则、跨仓库事务边界
│   │   │   ├── repository/     # 数据查询/持久化
│   │   │   ├── model/          # 当前 GORM 数据模型
│   │   │   ├── router/         # 本模块路由注册
│   │   │   └── seed/           # 基础数据默认值
│   │   ├── bench/             # 下列为业务分支建议新增
│   │   ├── run/               # 轮次/case 结果先内聚于 run
│   │   └── agent/             # agent 配置、阶段执行记录
│   ├── core/                  # config/db/file/auth/audit/observability
│   ├── middleware/            # Gin 横切行为
│   ├── svc/                   # 现有依赖组装上下文
│   └── docs/                  # Swagger 生成产物，不手工写业务
├── pkg/                       # 现有通用工具，避免继续成为杂物目录
└── configs/
```

后端不建议当前阶段把所有 model、service、repository 合并成根目录下三个大包，也不建议提前拆成多个 Go module。新业务模块不能直接读写其他模块的内部表来绕过其业务规则。

随业务增长逐步处理：

1. 将 `cmd/server` 中的组装逻辑提取为 `internal/bootstrap`，使 main 只处理参数和进程信号；迁移实现移入 `internal/migrations`，`cmd/migrate` 只负责运行。当前命令仍能正常工作，不为目录名称进行整体搬家。
2. 给新模块显式注入最小依赖。当前 ServiceContext 暂时兼容；不要让新的 run service 随意访问所有数据库、HTTP 服务和其他模块状态。
3. `core` 保持基础设施语义。如果某项需要导入业务 model（例如目前的审计），后续通过接口/独立事件 DTO 收窄依赖。可使用 `platform` 命名，但改目录名称本身不会改变依赖方向。
4. 只有复杂领域才把业务实体与 GORM model 分开；简单 CRUD 没必要增加转换层。事务涉及同一业务操作时，在明确边界内一次提交。
5. 当前 `pkg/utils/gin_context.go` 等依赖框架内部类型，实际上不是可独立复用的公共库。新工具优先放到实际使用的 internal 包；若未来对外发布 SDK，单独定义公共 API。

## 前端

```text
front-end/
├── config/                    # Umi 构建、代理、公共路由骨架
├── src/
│   ├── app.tsx                # Umi 应用入口
│   ├── requestErrorConfig.ts  # 请求/响应策略入口
│   ├── routing/               # 后端驱动动态路由
│   ├── pages/                 # Umi 页面入口，按实际页面领域组织
│   │   ├── user/              # 登录、注册等公共入口
│   │   ├── account/           # 当前账号设置
│   │   ├── sys/               # 系统管理，页面专用组件就近存放
│   │   ├── dashboard/
│   │   ├── state/
│   │   └── about/
│   ├── services/
│   │   ├── system/            # 当前 CMS API 与类型
│   │   ├── bench/             # 新业务模块新增时创建
│   │   ├── run/
│   │   └── agent/
│   ├── features/              # 多页面复用的复杂业务能力；按需创建
│   │   └── run-progress/      # 例如轮次图、事件订阅与恢复
│   ├── components/            # 跨业务通用 UI
│   ├── utils/                 # 真正通用的纯函数
│   └── locales/
├── e2e/                       # 浏览器与 API 端到端场景
└── tests/                     # 测试环境初始化
```

`features/bench` 与 `pages/bench` 不应同时放两份完整实现。简单页面把专用组件、hooks、测试放在页面附近即可；只有出现多页面复用的业务能力时，才提取到 features。`components` 不接收所有业务组件，`utils` 不接收状态机或业务服务。

动态路由由 **后端路径/组件标识 → routing/componentMap → pages** 组成。移动源文件时更新 componentMap 的 import，不必改变已有后端菜单记录；因此本轮目录调整不需要菜单数据库迁移。

接口类型下一步改为在 `services/<domain>/types.ts` 显式导出，逐渐替代全局 API namespace；类型生成以本仓库 Swagger 为来源。单元测试紧邻被测文件，E2E 按业务场景组织。

## 仓库与部署

继续使用一个仓库、一个前端包和一个 Go module。`backend`、`front-end` 的既有名称可以保留；统一成 `apps/api` / `apps/web` 只有在真的增加多个可独立发布的应用时才有价值。

- `docs/`：架构、数据模型、运行与维护说明。
- `deploy/local/`：当前可运行且已验证的 Compose 入口。
- `deploy/k3s/`：已有备用示例，尚未按本轮变更验收；不把它当成当前生产支持承诺。
- `scripts/`：跨应用维护脚本；浏览器测试由 `deploy/local/e2e.py` 管理隔离环境。
- 不提交 `node_modules`、构建产物、数据库文件、配置密钥、备份或测试运行目录。

判断拆目录是否值得的标准：**能否更容易找到完整功能、看清依赖、独立测试和减少修改范围**。当前最有价值的是业务归属和依赖边界，层级越深并不意味着架构越好。
