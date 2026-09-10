#!/usr/bin/env python3
"""Enable optional local telemetry while preserving the deployment's private configuration."""
import json
import os
from pathlib import Path
import subprocess
root=Path(__file__).resolve().parents[2]
p=root/'deploy/local/runtime/backend.yaml'
cfg=json.loads(p.read_text())
cfg['observable'].update(exporter='otel',service_name='nexus-cms',service_version='base-frame-20260910',trace_sample_ratio=1.0)
cfg['observable']['otel_exporter']={'protocol':'http','endpoint':'otel-collector:4318','insecure':True,'authorization':'','organization':'','stream_name':''}
cfg['system'].setdefault('allow_registration',False)
cfg['system'].setdefault('trusted_proxies',['172.16.0.0/12','192.168.0.0/16','10.0.0.0/8'])
p.write_text(json.dumps(cfg,indent=2)+'\n');p.chmod(0o600)
subprocess.run(['docker','compose','--env-file',str(root/'deploy/local/.env'),'-f',str(root/'deploy/local/compose.yaml'),'-f',str(root/'deploy/observability/compose.yaml'),'up','-d','--no-build','--wait','--wait-timeout','180'],check=True)
# File content changes do not trigger Compose container recreation.
subprocess.run(['docker','compose','--env-file',str(root/'deploy/local/.env'),'-f',str(root/'deploy/local/compose.yaml'),'restart','backend'],check=True)
