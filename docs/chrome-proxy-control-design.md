# Chrome 浏览器代理控制服务设计

## 目标

在不启动本地代理应用的情况下，只让 Chrome 使用 my-hk 服务器的代理服务。

Chrome 插件负责：

- 设置 Chrome 的 fixed proxy 或 PAC 配置；
- 实现规则分流、全局代理、境内直连和关闭代理；
- 选择节点、保存本地规则、显示当前网站的命中结果；
- 通过 HTTPS 控制接口同步节点和规则。

my-hk 服务端负责：

- 提供浏览器可直接使用的 HTTP CONNECT 代理入口；
- 将 HTTP CONNECT 请求转发到 sing-box 或 Xray 的指定出口；
- 为插件提供节点、规则版本和健康状态；
- 使用现有 API Token、限流、审计和操作日志能力。

## 当前仓库结论

仓库位置：`/home/claude/go-web-frame`

开发工作树：`my-hk-server`

GitHub 远端存在 `my-hk-server` 分支，当前提交为 `117da0e`；服务器工作树已在对应的本地 `my-hk-server` 分支上，但尚未建立 `origin/my-hk-server` 远程跟踪引用。后续提交应明确推送到该分支。

已有能力：

- `internal/modules/proxy`：sing-box/Xray 实例管理；
- 配置读取、校验、原子保存、备份和回滚；
- systemd 启停、状态读取和流量指标；
- JWT/Casbin 管理后台；
- API Token、API 授权列表、并发限制和审计日志；
- 前端动态菜单和现有代理管理页面。

当前运行配置中，sing-box/Xray 只有 Hysteria2、Trojan、VLESS、Shadowsocks 等协议入口，没有 HTTP CONNECT 或 SOCKS5 入口。Chrome 插件不能直接使用这些协议，因此必须增加一个标准代理入口。

## 组件边界

### 现有代理管理模块

保留 `internal/modules/proxy` 的职责：

- 管理 sing-box/Xray 进程；
- 编辑和校验核心配置；
- 读取 systemd 状态；
- 读取运行指标。

不把浏览器规则、设备 Token 或 PAC 内容塞入核心实例模型。

### 新增浏览器代理模块

建议路径：

```text
backend/internal/modules/browserproxy/
  api.go
  model.go
  service.go
  types.go
  service_test.go
```

职责：

- 管理浏览器可用的代理入口(profile)；
- 返回插件 bootstrap 数据；
- 管理规则集元数据；
- 检查入口健康状态；
- 生成不包含服务端私密配置的公开响应。

第一版可以将 profile 配置放在服务端 YAML 中，稳定后再迁移到数据库模型；不要自动解析所有 VLESS/Trojan 密钥并暴露给插件。

## 数据面

每个可选节点提供一个独立的 HTTP CONNECT 入口：

```text
Chrome -> HTTP CONNECT -> my-hk:41xxx -> sing-box/Xray -> 目标出口
```

推荐每个节点一个固定入口，插件切换节点时只切换 PAC 返回的 host/port。不要通过修改全局 sing-box/Xray selector 来切节点，否则多个设备会互相影响。

入口建议：

- scheme: `http` 或 `https`；
- host: 服务器公开域名；
- port: 独立端口；
- auth: HTTP Proxy-Authorization；
- TLS 和证书由 Nginx 或 sing-box/Xray 入口负责；
- 入口仅开放给浏览器代理用途，独立限流。

Chrome 插件通过 `webRequest.onAuthRequired` 仅在 `details.isProxy` 且 challenger 匹配当前入口时返回代理凭证，不能将凭证用于普通网站认证。

## 控制面 API

控制接口统一挂在现有 API 前缀下，并使用 `ApiTokenAuth`：

```text
GET /api/v1/proxy/browser/bootstrap
GET /api/v1/proxy/browser/health
```

bootstrap 返回：

```json
{
  "revision": "2026-09-19T13:00:00Z",
  "defaultMode": "rule",
  "defaultNodeId": "hk-01",
  "nodes": [
    {
      "id": "hk-01",
      "name": "香港 01",
      "scheme": "http",
      "host": "proxy.example.com",
      "port": 41001,
      "username": "device-scoped-user",
      "password": "short-lived-secret",
      "status": "ready",
      "latencyMs": 82
    }
  ],
  "ruleSets": [
    {
      "id": "cn-direct",
      "name": "中国大陆直连",
      "version": "2026-09-19",
      "url": "/api/v1/proxy/browser/rules/cn-direct"
    }
  ]
}
```

安全约束：

- 控制 API 只返回插件需要的入口信息；
- 不返回 sing-box/Xray 原始配置、VLESS 私钥、Reality 私钥或管理端口；
- Token 只授予上述只读接口；
- 节点凭证使用设备范围、可过期的代理认证信息；
- bootstrap 和规则响应加入版本号，插件本地缓存；
- 服务端只记录设备标识、错误和更新时间，不记录完整访问 URL。

## API Token 接入

复用现有系统 API Token，不新增第二套认证系统。

需要在迁移/种子中注册：

```text
GET /api/v1/proxy/browser/bootstrap
GET /api/v1/proxy/browser/health
```

插件安装时由用户在 CMS 的 API Token 页面创建一个只读 Token，并粘贴到插件设置中。后续若需要二维码配对，再增加设备授权流程，但不改变现有 Token 中间件。

## 规则与模式

插件本地生成 PAC：

- 规则模式：用户规则优先，之后是系统绕过规则，最后使用默认代理；
- 全局模式：全部使用当前节点；
- 境内直连：中国大陆域名/IP、内网和 localhost 直连，其他请求走代理；
- 关闭代理：恢复插件接管前的 Chrome 配置。

服务端只提供规则数据，不下发远程 JavaScript。插件的自定义规则保存在 `chrome.storage.local`，规则更新使用版本号和增量数据。

## 管理后台 UI

现有“代理管理”页面保留进程管理功能，新增“浏览器代理”区域：

1. 入口管理
   - 名称、节点 ID、HTTP CONNECT 地址、端口；
   - 启用状态、凭证轮换、健康检查；
   - 延迟、最近错误、最近使用时间。

2. 规则集
   - 中国大陆直连规则；
   - 自定义代理规则；
   - 自定义直连规则；
   - 规则版本和发布状态。

3. 插件接入
   - 创建只读 API Token；
   - 显示 bootstrap 地址；
   - 复制配置或生成配对二维码；
   - 撤销设备凭证。

Chrome 插件弹窗只保留高频操作：

```text
状态开关
模式：规则分流 / 全局代理 / 境内直连 / 关闭代理
当前节点和延迟
当前网站的命中结果
快速加入代理 / 快速加入直连
设置、节点、规则管理入口
```

## 低占用与稳定性

- 使用 `chrome.proxy` 和 PAC，不监听全部请求；
- Manifest V3 Service Worker 不常驻轮询；
- bootstrap 和规则只在启动、用户操作或版本变化时请求；
- 节点健康检查使用服务端缓存结果；
- 代理配置应用失败时回滚上一份配置；
- 保存 Chrome 原有代理设置，关闭插件时恢复；
- 监听其他扩展修改代理设置并提示冲突；
- 服务器入口和控制 API 分离限流；
- 规则和凭证均不写入核心 sing-box/Xray 配置备份。

## 实施顺序

### 第 1 阶段：服务端接口和入口

- 确认 HTTP CONNECT 入站使用 sing-box 还是 Xray；
- 为至少一个节点增加独立 HTTP 代理入口；
- 增加 browserproxy 配置结构和 bootstrap API；
- 接入 API Token allowlist；
- 增加节点健康检查和脱敏测试。

### 第 2 阶段：Chrome 插件 MVP

- Manifest V3、Service Worker、popup；
- 四种模式；
- 单节点切换；
- 本地自定义规则；
- HTTP Proxy-Authorization；
- 配置冲突检测和恢复。

### 第 3 阶段：后台管理和规则同步

- CMS 浏览器代理管理界面；
- 多节点入口和健康状态；
- 规则集版本管理；
- 二维码配对和凭证轮换；
- 自动选择延迟最低节点。

## 验收条件

- 不启动本地代理应用时，Chrome 可以单独访问外网；
- 关闭插件后系统其他应用不受影响；
- 规则、全局、境内直连、关闭四种模式行为可验证；
- 节点切换不改变其他设备的出口；
- 代理认证失败不会向普通网站泄露凭证；
- 服务端重启后入口配置和插件 bootstrap 仍可用；
- 现有 sing-box/Xray 管理页、配置回滚和监控测试不回归。
