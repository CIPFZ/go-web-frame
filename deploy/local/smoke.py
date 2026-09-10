#!/usr/bin/env python3
"""Check the deployed frontend proxy, admin login, menus and system APIs."""

import base64
import json
from pathlib import Path
import secrets
from urllib.error import HTTPError
from urllib.request import build_opener, ProxyHandler, Request


def main():
    root = Path(__file__).resolve().parent
    env = dict(line.split('=', 1) for line in (root / '.env').read_text().splitlines()
               if line and not line.startswith('#'))
    base = f'http://127.0.0.1:{env.get("CMS_WEB_PORT", "8080")}'
    opener = build_opener(ProxyHandler({}))

    def request(path, payload=None, token=None):
        headers = {'Content-Type': 'application/json'}
        if token:
            headers['x-token'] = token
        body = json.dumps(payload).encode() if payload is not None else None
        with opener.open(Request(base + path, data=body, headers=headers), timeout=20) as response:
            return json.load(response)

    with opener.open(base, timeout=20) as response:
        assert response.status == 200 and '<html' in response.read().decode().lower()
    assert request('/health')['status'] == 'ok'
    readiness = request('/ready')
    assert readiness['status'] == 'ok'
    assert readiness['checks'].get('database') == 'ok'
    assert readiness['checks'].get('redis') == 'ok'
    print('Frontend and proxied backend health: OK')

    login = request('/api/v1/user/login', {'username': 'admin', 'password': env['ADMIN_PASSWORD']})
    assert login.get('code') == 0, f'Login failed with code {login.get("code")}'
    token = login['data']['token']
    assert token
    print('Admin login: OK')
    for path, payload in (
        ('/api/v1/sys/user/getSelfInfo', None),
        ('/api/v1/sys/menu/getMenu', None),
        ('/api/v1/sys/system/getServerInfo', {}),
    ):
        result = request(path, payload, token)
        assert result.get('code') == 0, f'{path}: code {result.get("code")}'
        if path.endswith('getMenu'):
            assert result['data'], 'No menus were seeded'
            assert not any(word in json.dumps(result['data']) for word in ('plugin', 'poetry'))
        print(f'{path}: OK')

    for path, method in (
        ('/api/v1/plugin/public/getPublishedPluginList', 'POST'),
        ('/api/v1/plugin/plugin/getPluginList', 'POST'),
        ('/api/v1/plugin/release/claim', 'POST'),
        ('/api/v1/plugin/product/getProductList', 'POST'),
        ('/api/v1/poetry/poem/list', 'GET'),
        ('/api/v1/poetry/dynasty/list', 'GET'),
    ):
        try:
            opener.open(Request(base + path, method=method, headers={'x-token': token}), timeout=20)
        except HTTPError as error:
            assert error.code == 404, f'{path}: expected 404, got {error.code}'
        else:
            raise AssertionError(f'Removed API is still registered: {path}')
    swagger = request('/api/v1/swagger/doc.json')
    assert not any(word in json.dumps(swagger).lower() for word in ('autocodeplugin', 'installplugin', '/poetry/', '/plugin/'))
    print('Removed module APIs and Swagger entries: OK')

    image = base64.b64decode('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII=')
    boundary = 'cms-smoke-' + secrets.token_hex(8)
    upload_body = (
        f'--{boundary}\r\nContent-Disposition: form-data; name="file"; filename="deployment-smoke.png"\r\n'
        'Content-Type: image/png\r\n\r\n'
    ).encode() + image + f'\r\n--{boundary}--\r\n'.encode()
    upload_request = Request(base + '/api/v1/sys/file/upload', data=upload_body, headers={
        'x-token': token, 'Content-Type': f'multipart/form-data; boundary={boundary}',
    })
    with opener.open(upload_request, timeout=20) as response:
        uploaded = json.load(response)
    assert uploaded.get('code') == 0, f'Upload failed with code {uploaded.get("code")}'
    url = uploaded['data']['url']
    assert url.startswith('/uploads/file/')
    with opener.open(base + url, timeout=20) as response:
        assert response.read() == image, 'Uploaded image did not round-trip through Nginx'
    print('Local upload and Nginx download: OK')
    print('Deployment smoke checks passed')


if __name__ == '__main__':
    main()
