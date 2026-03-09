# K3s Deployment Guide

## 1) Build Images

From project root:

```bash
docker build -t web-cms-backend:local ./backend
docker build -t web-cms-frontend:local ./front-end
```

## 2) Import Images To k3s

If your k3s uses containerd:

```bash
docker save web-cms-backend:local | sudo k3s ctr images import -
docker save web-cms-frontend:local | sudo k3s ctr images import -
```

## 3) Configure External Dependencies

This k3s manifest set does not create MySQL/Redis/MinIO/Mongo pods.
It references your existing cluster services via backend config:

- [backend-config.yaml](C:/Users/ytq/work/ai/web-cms/deploy/k3s/base/backend-config.yaml)

Before deploy, edit these fields to your real service DNS and credentials:

- `mysql.path`
- `mysql.username` / `mysql.password`
- `redis.addr` / `redis.password`
- `file.minio.endpoint` / `access_key` / `secret_key` / `bucket`
- (optional) `mongo.hosts` if enabling `system.use_mongo=true`

## 4) Deploy

```bash
kubectl apply -k deploy/k3s/base
kubectl -n web-cms get pods
kubectl -n web-cms get svc
kubectl -n web-cms get ingress
```

## 5) Local Access

Add hosts entry:

```text
127.0.0.1 cms.local
```

Then open:

- http://cms.local

Default admin account (seeded on deploy):

- username: `admin`
- password: `Admin@123456`

## 6) Seed Behavior

Backend deployment enables:

- `SEED_ADMIN_ENABLED=true`
- `SEED_ADMIN_USERNAME`
- `SEED_ADMIN_PASSWORD`
- `SEED_ADMIN_AUTHORITY_ID=1`
- `SEED_ADMIN_DEFAULT_ROUTER=dashboard/workplace`

This triggers idempotent base data initialization during startup:

- authorities/roles
- menus
- apis
- casbin policies
- admin user
- base poetry metadata

## 7) Validation Commands

```bash
kubectl -n web-cms logs deploy/backend --tail=200
kubectl -n web-cms port-forward svc/backend 8080:8080
curl -s http://127.0.0.1:8080/health
```

Login smoke test:

```bash
curl -s -X POST http://127.0.0.1:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin@123456"}'
```

## 8) Notes

- These manifests only deploy `backend/frontend/ingress` and reuse existing middleware.
- For production, move credentials to a secure secret manager and replace default passwords.
- If your cluster does not use Traefik ingress class, update `deploy/k3s/base/ingress.yaml`.
