#!/usr/bin/env python3
"""Fail when a literal frontend request has no matching local OpenAPI operation."""
import json
from pathlib import Path
import re
root=Path(__file__).resolve().parent.parent
spec=json.loads((root/'backend/internal/docs/swagger.json').read_text())
paths=spec['paths'];errors=[];checked=0
for source in (root/'front-end/src/services/system').glob('*.ts'):
    text=source.read_text()
    for match in re.finditer(r"request(?:<[\s\S]*?>)?\(\s*['\"](/api/v1/[^'\"]+)['\"]\s*[,)]",text):
        route=match.group(1).removeprefix('/api/v1')
        options=text[match.end():].split('\n}',1)[0] if text[match.end()-1]==',' else ''
        method=re.search(r"method:\s*['\"]([A-Z]+)['\"]",options)
        verb=method.group(1).lower() if method else 'get'
        checked+=1
        if verb not in paths.get(route,{}): errors.append(f'{source.name}: {verb.upper()} {route}')
if errors:raise SystemExit('Missing OpenAPI contracts:\n'+'\n'.join(errors))
print(f'Frontend/OpenAPI contracts: {checked} operations verified')
