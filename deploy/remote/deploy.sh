#!/usr/bin/env bash
# 发布到一台服务器：本机构建镜像 → 经 SSH 传到服务器 → 同步部署目录（编排、.env、运行配置）→ 服务器上启动（迁移随启动执行）
# → 只清理本项目被替换下来的旧镜像 → 检查站点。服务器只运行服务，不装构建工具链。
# 所有参数都从环境变量读取，没有默认值；缺了就退出并列出缺哪几项。说明见同目录 README.md。
# 用法：deploy/remote/deploy.sh [docker compose build 的参数，如 --no-cache]
set -euo pipefail
cd "$(dirname "$0")/../.."

required=(DEPLOY_HOST DEPLOY_SSH_USER DEPLOY_SSH_PORT DEPLOY_SSH_KEY DEPLOY_DIR DEPLOY_BUNDLE DEPLOY_PROJECT DEPLOY_SITE)
missing=()
for name in "${required[@]}"; do
  [[ -n "${!name:-}" ]] || missing+=("$name")
done
if ((${#missing[@]})); then
  echo "缺少环境变量：${missing[*]}（见 deploy/remote/README.md）" >&2
  exit 1
fi
for file in "$DEPLOY_SSH_KEY" "$DEPLOY_BUNDLE/compose.yaml" "$DEPLOY_BUNDLE/.env"; do
  [[ -e "$file" ]] || { echo "缺少 $file" >&2; exit 1; }
done

SSH=(ssh -i "$DEPLOY_SSH_KEY" -o IdentitiesOnly=yes -o ConnectTimeout=15 -p "$DEPLOY_SSH_PORT" "$DEPLOY_SSH_USER@$DEPLOY_HOST")
RSYNC=(rsync -rlpt --delete -e "ssh -i $DEPLOY_SSH_KEY -o IdentitiesOnly=yes -p $DEPLOY_SSH_PORT")
COMPOSE=(docker compose -p "$DEPLOY_PROJECT" --env-file "$DEPLOY_BUNDLE/.env" -f "$DEPLOY_BUNDLE/compose.yaml")

echo "== 构建镜像"
"${COMPOSE[@]}" build "$@"
mapfile -t images < <("${COMPOSE[@]}" config --images | sort -u)

echo "== 传镜像到 $DEPLOY_HOST：${images[*]}"
docker save "${images[@]}" | gzip -1 | "${SSH[@]}" 'gunzip | docker load'

echo "== 同步部署目录"
"${SSH[@]}" "mkdir -p '$DEPLOY_DIR'"
# 备份与本机的缓存不同步；.env 与运行配置里有凭据，同步后只让本人可读
"${RSYNC[@]}" --exclude 'runtime/backups/' --exclude '__pycache__/' "$DEPLOY_BUNDLE/" "$DEPLOY_SSH_USER@$DEPLOY_HOST:$DEPLOY_DIR/"
"${SSH[@]}" "chmod 600 '$DEPLOY_DIR/.env'; find '$DEPLOY_DIR' -name '*.yaml' -path '*runtime*' -exec chmod 600 {} +"

echo "== 服务器上启动"
"${SSH[@]}" bash -s <<REMOTE
set -euo pipefail
cd '$DEPLOY_DIR'
docker compose -p '$DEPLOY_PROJECT' --env-file .env -f compose.yaml up -d --no-build --wait --wait-timeout 300 --remove-orphans
# 只清理本项目载入新版后留下的、没有标签的旧镜像
docker image prune -f --filter 'label=com.docker.compose.project=$DEPLOY_PROJECT' >/dev/null
docker compose -p '$DEPLOY_PROJECT' --env-file .env -f compose.yaml ps
REMOTE

echo "== 检查站点"
for path in /healthz /ready; do
  code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 15 "$DEPLOY_SITE$path" || true)
  echo "$DEPLOY_SITE$path -> $code"
  [[ "$code" == 200 ]] || { echo "站点检查失败" >&2; exit 1; }
done
