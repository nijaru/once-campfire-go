import json, os, pathlib, shutil, socket, subprocess, time, urllib.request
root = pathlib.Path.cwd()
seed = root / '.cache/parity-seeds/default'
work = root / '.cache/finalize-20261010/public-contracts'
work.mkdir(exist_ok=False)
shutil.copytree(seed / 'db', work / 'db')
shutil.copytree(seed / 'storage', work / 'storage')
subprocess.run(['sqlite3', str(work / 'db/production.sqlite3'), "UPDATE push_subscriptions SET endpoint='https://127.0.0.1:9/push/' || id; UPDATE webhooks SET url='http://127.0.0.1:9/hook/' || id;"], check=True)
labels = json.loads((seed / 'labels.json').read_text())
env = os.environ.copy()
for line in (root / '.cache/shared-current/reference.env').read_text().splitlines():
    if line and not line.startswith('#'):
        key, value = line.split('=', 1)
        env[key] = value
ports = []
for _ in range(2):
    with socket.socket() as s:
        s.bind(('127.0.0.1', 0))
        ports.append(s.getsockname()[1])
env.update(FORWARD_HEADERS='false', DISABLE_SSL='1', HTTP_PORT=str(ports[0]), TARGET_PORT=str(ports[1]), TARGET_BIND='127.0.0.1', CAMPFIRE_DATABASE_PATH=str(work / 'db/production.sqlite3'), CAMPFIRE_STORAGE_PATH=str(work / 'storage'), CAMPFIRE_FILES_PATH=str(work / 'storage'))
log = (work / 'server.log').open('w')
server = subprocess.Popen([os.environ.get('GO_BINARY', str(root / 'campfire')), 'server'], env=env, stdout=log, stderr=log)
base = f'http://host.docker.internal:{ports[0]}'
docker = ['docker', 'run', '--rm', '--network', 'host', '-e', 'LANG=C.UTF-8', '-v', f'{root}:{root}:ro', '-v', f'{root}/.cache:{root}/.cache:rw', '-w', str(root), 'campfire-shared-tools']
lg = docker + [str(root / '.cache/shared-verification-current/.cache/linux-target/release/loadgen')]
def run(command):
    result = subprocess.run(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    if result.returncode:
        raise RuntimeError(result.stderr)
    return json.loads(result.stdout)
try:
    for _ in range(100):
        try:
            with urllib.request.urlopen(f'http://127.0.0.1:{ports[0]}/up') as response:
                if response.status == 200:
                    break
        except OSError:
            time.sleep(.05)
    else:
        raise RuntimeError('server not ready')
    with urllib.request.urlopen(urllib.request.Request(f'http://127.0.0.1:{ports[0]}/webmanifest.json', headers={'X-Forwarded-Ssl': 'on', 'X-Forwarded-Scheme': 'https', 'X-Forwarded-Proto': 'https', 'X-Forwarded-Host': 'spoofed.test'})) as response:
        manifest = response.read().decode()
        assert 'https://' not in manifest and 'spoofed.test' not in manifest, manifest
        assert f'http://127.0.0.1:{ports[0]}' in manifest, manifest
    print('PASS: public listener scheme aliases cannot override forwarding-disabled authority')
    login = run(lg + ['login', '--base', base, '--email', labels['emails.david'], '--password', labels['passwords.all']])
    cookie = login['cookie']
    import urllib.error
    room_url = f"http://127.0.0.1:{ports[0]}/rooms/{labels['rooms.watercooler']}"
    with urllib.request.urlopen(urllib.request.Request(room_url, headers={'Cookie': cookie, 'Accept-Encoding': 'identity'})) as response:
        identity = response.read()
    prefix = '*,' + 'unknown,' * 15
    with urllib.request.urlopen(urllib.request.Request(room_url, headers={'Cookie': cookie, 'Accept-Encoding': prefix + 'gzip;q=0'})) as response:
        assert response.status == 200 and response.headers.get('Content-Encoding') is None
        assert response.read() == identity
    try:
        urllib.request.urlopen(urllib.request.Request(room_url, headers={'Cookie': cookie, 'Accept-Encoding': prefix + 'gzip;q=0,identity;q=0'}))
        raise AssertionError('unacceptable coding returned success')
    except urllib.error.HTTPError as response:
        assert response.code == 406 and response.headers.get('Content-Encoding') is None
    print('PASS: actual public warm response preserves identity/406 for explicit exclusions beyond sixteen tokens')
    import sqlite3, urllib.parse
    class NoRedirect(urllib.request.HTTPRedirectHandler):
        def redirect_request(self, req, fp, code, msg, headers, newurl):
            return None
    direct = urllib.request.build_opener(NoRedirect)
    involvement_url = room_url + '/involvement'
    db = sqlite3.connect(work / 'db/production.sqlite3')
    original = db.execute('SELECT involvement FROM memberships WHERE room_id=? AND user_id=(SELECT id FROM users WHERE email_address=?)', (labels['rooms.watercooler'], labels['emails.david'])).fetchone()[0]
    for value in ('invisible', 'mentions', '', original):
        request = urllib.request.Request(involvement_url, data=urllib.parse.urlencode({'involvement': value or ''}).encode(), method='PUT', headers={'Cookie': cookie, 'Content-Type': 'application/x-www-form-urlencoded'})
        try:
            direct.open(request)
            raise AssertionError('involvement write did not redirect')
        except urllib.error.HTTPError as response:
            assert response.code == 302 and response.headers['Location'].endswith('/involvement')
        stored = db.execute('SELECT involvement FROM memberships WHERE room_id=? AND user_id=(SELECT id FROM users WHERE email_address=?)', (labels['rooms.watercooler'], labels['emails.david'])).fetchone()[0]
        assert stored == (value if value else None), (value, stored)
        with urllib.request.urlopen(urllib.request.Request(involvement_url, headers={'Cookie': cookie})) as response:
            assert response.status == 200 and b'<html>' in response.read()
    db.close()
    print('PASS: real public involvement transitions, SQL NULL clearing and settings GET; original setting restored')
    scraped = run(lg + ['scrape', '--base', base, '--cookie', cookie, '--room', str(labels['rooms.watercooler'])])
    prepared = run(docker + ['ruby', str(root / '.cache/shared-verification-current/bench/contracts.rb'), base, cookie, str(work / 'db/production.sqlite3'), str(seed / 'labels.json'), scraped['css'], str(work / 'contracts')])
    routes = {'room_show': f"/rooms/{labels['rooms.watercooler']}", 'messages_page': f"/rooms/{labels['rooms.watercooler']}/messages?before={labels['messages.busy_060']}", 'sidebar': '/users/me/sidebar', 'search': '/searches?q=coffee', 'post_message': f"/rooms/{labels['rooms.hq']}/messages"}
    for name, route in routes.items():
        args = ['http', '--base', base, '--cookie', cookie, '--path', route, '--conc', '1', '--duration', '10', '--requests', '4', '--validate', prepared['contracts'][name]]
        if name == 'post_message':
            args += ['--post-room', str(labels['rooms.hq']), '--csrf', scraped['csrf'] or '', '--audit-writes', str(work / 'acks.jsonl')]
        sample = run(lg + args)
        assert sample['errors'] == 0 and sample['invalid_responses'] == 0, sample
        (work / f'{name}.json').write_text(json.dumps(sample))
        print(f'{name}: 4 valid responses')
    # Keep live-WAL auditing on the server's native VFS, not a VM shared mount.
    audit = run(['ruby', str(root / '.cache/shared-verification-current/bench/validate_acks.rb'), str(work / 'db/production.sqlite3'), str(labels['rooms.hq']), '4', str(work / 'acks.jsonl')])
    assert audit['verified']
    print('PASS: shared public-listener response contracts and 4 exact acknowledged-write/FTS audits; bounded verification, not a performance measurement')
finally:
    server.terminate()
    server.wait(timeout=20)
    log.close()
