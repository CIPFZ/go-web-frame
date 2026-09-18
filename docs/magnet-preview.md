# 磁力预览

在 `feature/magnet-preview` 分支开发。SQLite 保存 CMS 菜单及权限；元信息与可选封面写入本地文件缓存，无 Redis、Mongo 或 Docker 依赖。

## 启用

在服务 YAML 中添加：

```yaml
magnet_preview:
  enabled: true
  # 建议使用独立绝对路径；省略则在配置文件同目录创建 magnet-preview。
  cache_dir: /home/claude/go-web-frame/runtime/magnet-preview
  metadata_timeout_sec: 45
  max_files: 1000
  fetch_cover: true
  max_cover_mb: 2
  max_concurrent: 1
```

先构建并执行 `go run ./cmd/migrate -f /path/to/config.yaml`，再启动新版本后端及前端。
迁移新增“磁力预览”菜单及 API，仅默认授予内置管理员（角色 1），其他角色通过现有权限界面分配。
未启用时不创建 BitTorrent 客户端或注册预览 API。

## 接口

- `POST /api/v1/magnet/preview`：JSON `{"magnet":"magnet:?xt=urn:btih:..."}`，复用 CMS 的 x-token 登录验证及 Casbin 权限。
- `GET /api/v1/magnet/cover/:hash`：同样需要登录和授权；页面通过认证请求加载 Blob。
- 响应包含 name、info_hash、total_size、file_count、files、content_type，以及可选 cover。
- type/content_type：video、audio、image、document、archive、disk_image、other。
- 扩展名用于文件类型提示；资源主类型结合文件体积判断，封面不覆盖视频、音频、镜像或压缩包类型。
- 多文件种子可能包含多种类型，逐文件返回 type；截断展示不会影响总数、总大小和整体分类。

## 行为与限制

支持 BitTorrent v1 的十六进制及 Base32 BTIH，包括包含 BTIH 的混合磁力；纯 v2 BTMH 暂不支持。
使用 tracker / DHT 获取真实元信息，必须存在可连接的节点。名称参数 dn 不是已验证的元信息，获取失败不会用其伪造成功响应。
元信息等待上限 45 秒；可选封面等待上限 8 秒。客户端建议等待 60 秒。
不下载完整视频、音乐或 ISO，也不从视频自动截帧。封面仅从种子内存在的 JPG、PNG、GIF、WebP 等受支持图片中挑选。
读取封面可能需要下载所在的完整 BitTorrent 分片及相邻字节；分片超过 8 MiB 时跳过。封面失败不影响元信息结果。
默认图片上限 2 MiB，配置最大 5 MiB；图片解码尺寸上限 4000 万像素。
每个请求使用独立临时存储，结束时清理分片；仅缓存 JSON 和验证过的封面。缓存有效期 24 小时，最多 64 条，成功写入后淘汰过期和较旧条目。
默认并发 1，最大 2；满载立即返回稍后重试，避免排队阻塞小内存宿主机。空闲时无常驻 torrent 客户端。
仅保留经过检查的 BTIH、名称与 tracker 参数，丢弃 xs/ws/x.pe 等来源；拒绝非公网 tracker 解析及 peer 地址。
封面缺失、下载失败或浏览器加载失败均使用类型图标。后台没有外部电影海报、音乐专辑数据库匹配。

## 验证

```sh
cd backend
GOMAXPROCS=2 GOMEMLIMIT=350MiB go test -p 1 ./...
go build -p 1 -o bin/base-frame ./cmd/server
go build -p 1 -o bin/migrate ./cmd/migrate
cd ../front-end
npm run tsc
npm run build
```
