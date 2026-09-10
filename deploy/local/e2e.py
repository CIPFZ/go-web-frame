#!/usr/bin/env python3
"""Run browser tests against disposable containers, never the shared CMS database."""

import importlib.util
import json
import os
from pathlib import Path
import subprocess
import uuid


def main():
    local = Path(__file__).resolve().parent
    repo = local.parent.parent
    project = 'nexus-cms-e2e-' + uuid.uuid4().hex[:10]
    run_dir = local / 'runtime' / project
    run_dir.mkdir(parents=True, mode=0o700)
    spec = importlib.util.spec_from_file_location('cms_init', local / 'init.py')
    initializer = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(initializer)
    initializer.ROOT = run_dir
    initializer.main()

    override = run_dir / 'compose.yaml'
    override.write_text('services:\n  backend:\n    ports: !reset []\n'
                        '    volumes: !override\n'
                        f'      - {json.dumps(str(run_dir / "runtime/backend.yaml") + ":/app/configs/local.yaml:ro")}\n'
                        '      - uploads:/app/uploads\n'
                        '  frontend:\n    ports: !override\n      - "127.0.0.1::80"\n')
    compose = ['docker', 'compose', '-p', project, '--env-file', str(run_dir / '.env'),
               '-f', str(local / 'compose.yaml'), '-f', str(override)]
    try:
        subprocess.run(compose + ['up', '-d', '--no-build', '--wait', '--wait-timeout', '180'], check=True)
        address = subprocess.check_output(compose + ['port', 'frontend', '80'], text=True).strip()
        values = dict(line.split('=', 1) for line in (run_dir / '.env').read_text().splitlines()
                      if line and not line.startswith('#'))
        env = os.environ.copy()
        env.update(E2E_BASE_URL='http://' + address, E2E_API_BASE='http://' + address,
                   E2E_ADMIN_USER='admin', E2E_ADMIN_PASS=values['ADMIN_PASSWORD'], E2E_ISOLATED='1')
        result = subprocess.run(['npx', 'playwright', 'test', '--workers=1', '--reporter=line',
                                 '--output=' + str(run_dir / 'playwright')],
                                cwd=repo / 'front-end', env=env)
        return result.returncode
    finally:
        # The random project owns only this run's containers and volumes.
        subprocess.run(compose + ['down', '-v', '--remove-orphans'], check=True)


if __name__ == '__main__':
    raise SystemExit(main())
