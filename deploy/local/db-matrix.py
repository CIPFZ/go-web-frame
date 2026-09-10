#!/usr/bin/env python3
"""Run migration/security tests against disposable MySQL 8.4, PostgreSQL and SQLite."""
import os
from pathlib import Path
import secrets
import subprocess
import time
import uuid
local=Path(__file__).resolve().parent
repo=local.parent.parent
project='cms-matrix-'+uuid.uuid4().hex[:8]
password=secrets.token_hex(16)
containers=[]
def run(args,**kw):return subprocess.run(args,check=True,**kw)
try:
    run(['docker','network','create',project],stdout=subprocess.DEVNULL)
    for name,image,env in [('mysql','mysql:8.4',{'MYSQL_ROOT_PASSWORD':password,'MYSQL_DATABASE':'cms_test'}),('postgres','postgres:17-alpine',{'POSTGRES_PASSWORD':password,'POSTGRES_DB':'cms_test'}),('redis','redis:7.4-alpine',{})]:
        container=project+'-'+name;containers.append(container)
        args=['docker','run','-d','--name',container,'--network',project,'--network-alias',name]
        # Environment values are passed through process environment, not argv/logs.
        for key in env:args+=['-e',key]
        run(args+[image],env={**os.environ,**env},stdout=subprocess.DEVNULL)
    for attempt in range(60):
        mysql=subprocess.run(['docker','exec',containers[0],'sh','-c','MYSQL_PWD="$MYSQL_ROOT_PASSWORD" mysql -h 127.0.0.1 -uroot -e "SELECT 1"'],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
        postgres=subprocess.run(['docker','exec',containers[1],'pg_isready','-U','postgres'],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
        if mysql.returncode==0 and postgres.returncode==0:break
        time.sleep(2)
    else:raise RuntimeError('Database startup timed out')
    env={**os.environ,'TEST_MYSQL_DSN':f'root:{password}@tcp(mysql:3306)/cms_test?charset=utf8mb4&parseTime=True&loc=UTC','TEST_POSTGRES_DSN':f'host=postgres user=postgres password={password} dbname=cms_test sslmode=disable','TEST_REDIS_ADDR':'redis:6379'}
    run(['docker','run','--rm','--network',project,'-v',str(repo/'backend')+':/app','-v',str(local/'runtime/go-mod-cache')+':/go/pkg/mod','-v',str(local/'runtime/go-build-cache')+':/root/.cache/go-build','-w','/app','-e','TEST_MYSQL_DSN','-e','TEST_POSTGRES_DSN','-e','TEST_REDIS_ADDR','-e','GOMAXPROCS=2','golang:1.27.1-alpine','sh','-c','go test -p 2 -count=1 -v ./internal/migrations ./internal/core/token'],env=env)
finally:
    for container in containers:subprocess.run(['docker','rm','-fv',container],stdout=subprocess.DEVNULL)
    subprocess.run(['docker','network','rm',project],stdout=subprocess.DEVNULL)
