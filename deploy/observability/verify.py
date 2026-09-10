#!/usr/bin/env python3
"""Query stored telemetry, never infer delivery from exporter configuration alone."""
import json
from pathlib import Path
import subprocess
import time
import uuid
from urllib.parse import urlencode
from urllib.request import build_opener, ProxyHandler, Request
root=Path(__file__).resolve().parents[2]
compose=['docker','compose','--env-file',str(root/'deploy/local/.env'),'-f',str(root/'deploy/local/compose.yaml'),'-f',str(root/'deploy/observability/compose.yaml')]
opener=build_opener(ProxyHandler({}))
def endpoint(service,port):
    cid=subprocess.check_output(compose+['ps','-q',service],text=True).strip()
    raw=subprocess.check_output(['docker','inspect','--format','{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}',cid],text=True).strip()
    return 'http://'+raw+':'+str(port)
def read(url,headers=None):
    with opener.open(Request(url,headers=headers or {}),timeout=10) as response:return json.load(response)
prom=endpoint('prometheus',9090);loki=endpoint('loki',3100);tempo=endpoint('tempo',3200)
trace=uuid.uuid4().hex
read('http://127.0.0.1:8080/api/v1/public/config',{'traceparent':f'00-{trace}-0123456789abcdef-01'})
for attempt in range(24):
    try:
        metric=read(prom+'/api/v1/query?'+urlencode({'query':'http_server_request_duration_seconds_count'}))['data']['result']
        if not metric:
            metric=read(prom+'/api/v1/query?'+urlencode({'query':'http_server_duration_milliseconds_count'}))['data']['result']
        logs=read(loki+'/loki/api/v1/query_range?'+urlencode({'query':'{service_name="nexus-cms"}','limit':1}))['data']['result']
        spans=read(tempo+'/api/traces/'+trace)
        assert metric and logs and spans, 'Waiting for signals'
        print('Stored metrics: OK; application log streams: OK; propagated trace lookup: OK')
        break
    except Exception:
        if attempt==23:raise
        time.sleep(5)
readiness=read(prom+'/api/v1/query?'+urlencode({'query':'httpcheck_status'}))['data']['result']
assert readiness, 'No readiness telemetry'
rules=read(prom+'/api/v1/rules')['data']['groups']
assert rules and all(rule['health']=='ok' for group in rules for rule in group['rules'])
print('Readiness metrics and evaluated alert rules: OK')
