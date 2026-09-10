# 本地观测部署

入口为现有 CMS 的 `/observability/`，使用独立的 Grafana 登录。首次部署的管理员用户名为 admin，密码取本机私有 `.env` 的 `ADMIN_PASSWORD`。Grafana 后续改密独立于 CMS。所有存储和采集端口只在 Docker 网络内可见。

```bash
# 前提：已构建 CMS 镜像、完成初始化与备份。
python3 deploy/observability/enable.py
python3 deploy/observability/verify.py
```

脚本保留数据库与 JWT 密钥，将后端配置为 OTLP HTTP，并启动可选 Compose overlay。配置文件是启动快照，修改后需重启后端。基础 `deploy/local/compose.yaml` 可以独立运行；不使用观测平台时，将 `observable.exporter` 设回 `none` 并重启后端。

## 数据边界

| 数据 | 当前采集/查询路径 |
|---|---|
| 请求量、HTTP 状态、耗时分布 | otelgin → Collector → Prometheus → Grafana |
| Go 运行时指标 | OTel runtime instrumentation → Prometheus |
| 就绪状态 | Collector 每 15 秒请求 `/ready`，检测 SQL/Redis/可选 Mongo |
| 审计丢弃 | `cms.audit.dropped` 指标，发生丢弃才有样本；队列不是可靠业务事件存储 |
| 应用日志 | Zap → OTLP → Loki，包含可用的 trace/span 关联 |
| 调用链 | HTTP + GORM/Redis instrumentation → Tempo，支持 W3C traceparent |
| 告警状态 | Prometheus 评估 Collector 不可达、后端未就绪、审计丢弃规则；Grafana 显示 ALERTS |

没有接入主机 node_exporter/cAdvisor，故不承诺整机或所有容器的 CPU、内存、磁盘、网络监控。没有接入 Nexus 业务，因此没有 bench/轮次/case 通过率。没有配置向邮件、聊天工具发送告警。

CMS 的“系统状态”页提供即时服务概览及观测入口。界面中的“观测已启用”表示后端已配置导出；导出链路健康应以仪表盘和 verify.py 的真实查询为准。

默认指标保留 14 天且限 1 GiB；日志保留 7 天；trace 保留 48 小时。日志和 trace 的磁盘量随流量变化。小团队默认 100% trace 采样，流量增长后通过 `trace_sample_ratio` 调整。五个观测容器各限制 512 MiB 内存。Docker 日志也应使用主部署的轮转策略。

版本：Collector contrib 0.160.0、Prometheus 3.14.0、Grafana 13.2.1、Loki 3.7.7、Tempo 2.9.5。版本通过官方 release/tag 与 Docker 拉取核对。Tempo 使用支持当前单进程部署的 2.9 补丁线；3.0 更改了存储/组件架构，应另行设计迁移，不能仅替换版本号。

## 告警规则验证

```bash
docker compose --env-file deploy/local/.env -f deploy/local/compose.yaml \
  -f deploy/observability/compose.yaml exec -T prometheus \
  promtool check rules /etc/prometheus/alerts.yaml
```

`verify.py` 发送带 traceparent 的真实请求，随后查询 Prometheus 样本、Loki 日志和 Tempo 指定 trace，最后检查规则评估健康。阈值触发与恢复用 `alert-tests.yaml` 验证，不停止共享数据库制造故障。
