#!/usr/bin/env python3
"""Private, consistent DB/files/config snapshots. Restore only to random disposable projects."""
import argparse
import datetime
import fcntl
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import uuid
from urllib.request import build_opener, ProxyHandler, Request

LOCAL = Path(__file__).resolve().parent
BASE = ['docker', 'compose', '--env-file', str(LOCAL / '.env'), '-f', str(LOCAL / 'compose.yaml')]

def run(args, **kw):
    return subprocess.run(args, check=True, **kw)

def snapshot(retain):
    root = LOCAL / 'runtime/backups'
    root.mkdir(mode=0o700, parents=True, exist_ok=True)
    with (root / '.lock').open('w') as lock:
        fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        name = datetime.datetime.now(datetime.timezone.utc).strftime('%Y%m%dT%H%M%SZ')
        temp = root / ('.partial-' + name)
        temp.mkdir(mode=0o700)
        running = subprocess.check_output(BASE + ['ps', '--status', 'running', '--services'], text=True).splitlines()
        backend_ids = subprocess.check_output(BASE + ['ps', '-q', 'backend'], text=True).splitlines()
        paused = False
        try:
            # Stop HTTP writes so the SQL snapshot and uploaded files describe one state.
            if 'backend' in running:
                run(BASE + ['stop', 'backend']); paused = True
            with (temp / 'database.sql').open('wb') as out:
                run(BASE + ['exec', '-T', 'mysql', 'sh', '-c',
                    'MYSQL_PWD="$MYSQL_PASSWORD" exec mysqldump -unexus_cms --single-transaction --quick --no-tablespaces --routines --triggers --set-gtid-purged=OFF nexus_cms'], stdout=out)
            with (temp / 'uploads.tar').open('wb') as out:
                run(BASE + ['exec', '-T', 'frontend', 'tar', 'cf', '-', '-C', '/srv/uploads', '.'], stdout=out)
            shutil.copyfile(LOCAL / '.env', temp / 'environment.env')
            shutil.copyfile(LOCAL / 'runtime/backend.yaml', temp / 'backend.yaml')
            files = {p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in temp.iterdir()}
            meta = {'format': 1, 'created_at': name, 'files': files,
                    'commit': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=LOCAL, text=True).strip()}
            (temp / 'manifest.json').write_text(json.dumps(meta, indent=2))
            for p in temp.iterdir(): p.chmod(0o600)
            result = root / name
            temp.rename(result)
            for old in sorted((p for p in root.iterdir() if p.is_dir() and not p.name.startswith('.')), reverse=True)[retain:]:
                # Delete only recognized, checksum-manifested snapshots owned by this utility.
                if (old / 'manifest.json').is_file(): shutil.rmtree(old)
            print('Backup complete:', result)
            return result
        finally:
            if paused: run(['docker', 'start', *backend_ids])
            if temp.exists(): shutil.rmtree(temp)

def verify(path):
    manifest = json.loads((path / 'manifest.json').read_text())
    expected = {'database.sql', 'uploads.tar', 'environment.env', 'backend.yaml'}
    assert manifest['format'] == 1 and set(manifest['files']) == expected, 'Invalid backup manifest'
    for name, digest in manifest['files'].items():
        assert hashlib.sha256((path / name).read_bytes()).hexdigest() == digest, 'Checksum mismatch: ' + name
    print('Checksums: OK')

def rehearse(path, mysql_image):
    verify(path)
    project = 'nexus-cms-restore-' + uuid.uuid4().hex[:10]
    work = Path(tempfile.mkdtemp(prefix=project, dir=LOCAL / 'runtime'))
    config = json.loads((path / 'backend.yaml').read_text())
    config['observable']['exporter'] = 'none'
    (work / 'backend.yaml').write_text(json.dumps(config)); (work / 'backend.yaml').chmod(0o600)
    mount = json.dumps(str(work / 'backend.yaml') + ':/app/configs/local.yaml:ro')
    override = work / 'compose.yaml'
    override.write_text('services:\n  mysql:\n    image: ' + mysql_image + '\n'
        '  migrator:\n    volumes: !override\n      - ' + mount + '\n'
        '  backend:\n    ports: !reset []\n    volumes: !override\n      - ' + mount + '\n      - uploads:/app/uploads\n'
        '  frontend:\n    ports: !override\n      - "127.0.0.1::80"\n')
    compose = ['docker','compose','-p',project,'--env-file',str(path/'environment.env'),'-f',str(LOCAL/'compose.yaml'),'-f',str(override)]
    try:
        run(compose + ['up','-d','--wait','--wait-timeout','180','mysql','redis'])
        with (path / 'database.sql').open('rb') as source:
            run(compose + ['exec','-T','mysql','sh','-c','MYSQL_PWD="$MYSQL_PASSWORD" exec mysql -unexus_cms nexus_cms'], stdin=source)
        with (path / 'uploads.tar').open('rb') as source:
            run(compose + ['run','--rm','--no-deps','-T','--entrypoint','tar','backend','xf','-','-C','/app/uploads'], stdin=source)
        # Verify restored bytes before HTTP startup can create new files.
        restored = subprocess.check_output(compose + ['run','--rm','--no-deps','-T','--entrypoint','tar','backend','cf','-','-C','/app/uploads','.'])
        import io, tarfile
        def file_hashes(raw):
            with tarfile.open(fileobj=io.BytesIO(raw)) as archive:
                return {m.name:hashlib.sha256(archive.extractfile(m).read()).hexdigest() for m in archive if m.isfile()}
        assert file_hashes(restored) == file_hashes((path/'uploads.tar').read_bytes()), 'Restored upload mismatch'
        run(compose + ['up','-d','--no-build','--wait','--wait-timeout','180'])
        base='http://'+subprocess.check_output(compose+['port','frontend','80'],text=True).strip()
        opener=build_opener(ProxyHandler({}))
        def api(route, data=None, token=None):
            headers={'Content-Type':'application/json'}
            if token:headers['x-token']=token
            with opener.open(Request(base+route,data=json.dumps(data).encode() if data is not None else None,headers=headers),timeout=20) as res:return json.load(res)
        values=dict(line.split('=',1) for line in (path/'environment.env').read_text().splitlines() if line and not line.startswith('#'))
        assert api('/ready')['status']=='ok'
        login=api('/api/v1/user/login',{'username':'admin','password':values['ADMIN_PASSWORD']});assert login['code']==0
        token=login['data']['token'];assert api('/api/v1/sys/menu/getMenu',token=token)['code']==0
        listing=api('/api/v1/sys/user/getUserList',{'page':1,'pageSize':100},token);assert listing['code']==0
        print('Restore rehearsal: DB, migrations, uploads, readiness, admin login, dynamic menus OK; engine:',mysql_image)
        print('Restored user count:',listing['data']['total'])
    finally:
        run(compose+['down','-v','--remove-orphans']);shutil.rmtree(work)

if __name__ == '__main__':
    os.umask(0o077)
    parser=argparse.ArgumentParser();parser.add_argument('action',choices=['create','verify','rehearse']);parser.add_argument('snapshot',nargs='?',type=Path);parser.add_argument('--retain',type=int,default=14);parser.add_argument('--mysql-image',default='mysql:8.0');args=parser.parse_args()
    if args.retain<1:parser.error('--retain must be positive')
    if args.action=='create':snapshot(args.retain)
    elif args.snapshot is None:parser.error('snapshot path is required')
    elif args.action=='verify':verify(args.snapshot.resolve())
    else:rehearse(args.snapshot.resolve(),args.mysql_image)
