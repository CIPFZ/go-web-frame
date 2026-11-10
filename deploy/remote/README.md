# 发布到一台服务器

`deploy.sh` 在本机构建镜像，经 SSH 传到服务器并启动；服务器只需要 Docker 与 Docker Compose，不装构建工具链。

1. 本机 `docker compose build`（后端镜像构建时会跑 Go 测试）。
2. `docker save | gzip | ssh … docker load` 把镜像传过去，不经过镜像仓库。
3. rsync 部署目录（`compose.yaml`、`.env`、运行配置）到服务器；`.env` 与 `runtime/*.yaml` 改为只有本人可读。
4. 服务器上 `docker compose up -d --no-build --wait`：迁移随启动执行，失败时不会启动新版本。
5. 只清理本项目被替换下来的旧镜像，然后检查站点的 `/healthz` 与 `/ready`。

## 参数

全部从环境变量读取，**没有默认值**，缺了就退出并列出缺哪几项。可以写在一个不提交的文件里，用前 `source`：

| 变量 | 含义 |
|---|---|
| `DEPLOY_HOST` / `DEPLOY_SSH_PORT` / `DEPLOY_SSH_USER` | 服务器地址、SSH 端口与用户 |
| `DEPLOY_SSH_KEY` | SSH 私钥文件 |
| `DEPLOY_DIR` | 服务器上的部署目录 |
| `DEPLOY_BUNDLE` | 本机的部署目录：里面要有 `compose.yaml` 与 `.env`，以及编排引用的运行配置（如 `deploy/local`，先运行它的 `init.py`） |
| `DEPLOY_PROJECT` | Compose 项目名（决定容器、卷与镜像清理的范围） |
| `DEPLOY_SITE` | 站点地址，如 `https://cms.example.com`，用于发布后检查 |

```bash
deploy/remote/deploy.sh            # 正常发布
deploy/remote/deploy.sh --no-cache # 参数原样交给 docker compose build
```

服务器上的数据卷由 Compose 项目名决定：同一台服务器部署多份时用不同的 `DEPLOY_PROJECT` 与 `DEPLOY_DIR`。
