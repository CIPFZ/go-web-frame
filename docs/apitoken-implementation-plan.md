# API Token 实施计划

## 阶段划分

### 阶段 1：后端基础能力

- 新增 token 核心工具
- 新增模型、DTO、仓储、服务、接口
- 注册路由与模块依赖
- 更新 AutoMigrate

### 阶段 2：鉴权链路

- 新增 `ApiTokenAuth` 中间件
- 新增并发控制器
- 为诗词只读接口增加 token 访问路由

### 阶段 3：前端管理页

- 新增 `sys/api-token` 页面
- 新增 service API
- 新增动态菜单与种子数据

### 阶段 4：测试与验证

- 后端单元测试
- 中间件集成测试
- 前端路由与页面测试
- Go test 与前端 Jest 验证

## 文件规划

后端预计新增或修改：

- `backend/internal/core/token/*`
- `backend/internal/middleware/api_token.go`
- `backend/internal/modules/system/model/api_token.go`
- `backend/internal/modules/system/dto/api_token.go`
- `backend/internal/modules/system/repository/api_token_repo.go`
- `backend/internal/modules/system/service/api_token_service.go`
- `backend/internal/modules/system/api/api_token.go`
- `backend/internal/modules/system/router/router.go`
- `backend/internal/core/server/router.go`
- `backend/cmd/migrate/main.go`
- `backend/cmd/server/seed_admin.go`

前端预计新增或修改：

- `front-end/src/services/api/apiToken.ts`
- `front-end/src/pages/sys/api-token/index.tsx`
- `front-end/config/routes.ts`
- 必要的测试文件

## 测试优先顺序

按 TDD 执行：

1. 先补 token 工具测试
2. 再补 service 与 repository 测试
3. 再补 middleware 测试
4. 再补前端路由和页面测试
5. 最后跑完整验证命令

## 闭环标准

达到以下条件才算完成：

- 数据表可迁移
- 后台可管理 token
- 外部脚本可用 token 调用样板业务接口
- 未授权、过期、禁用、超并发请求均被正确拒绝
- 创建和重置只返回一次明文 token
- 后端和前端测试通过
