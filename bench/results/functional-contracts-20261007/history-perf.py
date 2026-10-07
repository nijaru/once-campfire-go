import gzip, hashlib, importlib.machinery, importlib.util, json, os
from pathlib import Path
import shutil, sqlite3, subprocess, tempfile, urllib.request
root=Path('/src')
loader=importlib.machinery.SourceFileLoader('upgrade',str(root/'bin/check-upgrade'))
spec=importlib.util.spec_from_loader(loader.name,loader); upgrade=importlib.util.module_from_spec(spec);loader.exec_module(upgrade)
seed=root/'.cache/parity-seeds/default'; labels=json.loads((seed/'labels.json').read_text())
logs=Path(tempfile.mkdtemp(prefix='history-perf-',dir='/tmp'))
out=root/'.cache/correctness-current/history-perf';out.mkdir(exist_ok=True)
env={k:v for k,v in os.environ.items() if not k.startswith(('THRUSTER_','CAMPFIRE_')) and k!='TLS_DOMAIN'}
for line in (root/'reference/parity/.env.reference').read_text().splitlines():
 if line and not line.startswith('#') and '=' in line:k,v=line.split('=',1);env[k]=v
env.update(DISABLE_SSL='true',TARGET_BIND='127.0.0.1',GOMAXPROCS='4',TOKIO_WORKER_THREADS='4')
binaries={'before':root/'.cache/correctness-current/campfire-go-linux','after':root/'.cache/correctness-current/campfire-go-linux-after'}
report={'binaries':{k:hashlib.sha256(v.read_bytes()).hexdigest() for k,v in binaries.items()},'server_cpus':'0-3','client_cpus':'4-7','samples':[]}
expected={}
for repetition in range(3):
 for variant in (('before','after') if repetition%2==0 else ('after','before')):
  with tempfile.TemporaryDirectory(dir=logs) as temp:
   work=Path(temp); db=work/'database.sqlite3'
   with sqlite3.connect(f'file:{seed}/db/production.sqlite3?mode=ro',uri=True) as src,sqlite3.connect(db) as dst:src.backup(dst)
   shutil.copytree(seed/'storage',work/'files')
   wrapper=work/'server';wrapper.write_text(f'#!/bin/sh\nexec taskset -c 0-3 {binaries[variant]} "$@"\n');wrapper.chmod(0o755)
   environment=env|{'CAMPFIRE_DATABASE_PATH':str(db),'CAMPFIRE_STORAGE_PATH':str(work),'CAMPFIRE_FILES_PATH':str(work/'files')}
   with upgrade.server(wrapper,environment,logs,f'{repetition}-{variant}') as base:
    cookie=upgrade.login(base,labels)
    room=labels['rooms.watercooler'];anchor=labels['messages.busy_060']
    for case,path in (('latest',f'/rooms/{room}/messages'),('before',f'/rooms/{room}/messages?before={anchor}')):
     request=urllib.request.Request(base+path,headers={'Cookie':cookie,'Accept-Encoding':'gzip'})
     with urllib.request.urlopen(request) as response:
      assert response.status==200 and response.headers.get('Content-Encoding')=='gzip'
      encoded=response.read();body=gzip.decompress(encoded)
     digest=hashlib.sha256(body.replace(base.encode(),b'<ORIGIN>')).hexdigest()
     if case in expected:assert digest==expected[case]
     else:expected[case]=digest
     assert body.count(b'data-message-id="')==40
     command=['taskset','-c','4-7',str(root/'.cache/linux/loadgen'),'http','--base',base,'--cookie',cookie,'--path',path,'--conc','16','--gzip','1']
     subprocess.run(command+['--duration','1'],env=environment,check=True,stdout=subprocess.DEVNULL)
     sample=json.loads(subprocess.check_output(command+['--duration','10'],env=environment))
     assert sample['errors']==0 and set(sample['statuses'])=={'200'} and sample['avg_bytes']==len(encoded)
     report['samples'].append({'rep':repetition,'variant':variant,'case':case,'body_sha256':digest,'encoded_bytes':len(encoded),'result':sample})
     print(variant,case,sample['rps'],sample['latency'],flush=True)
     (logs/'raw.json').write_text(json.dumps(report,indent=2)+'\n')
for path in logs.glob('*'):
 if path.is_file():
  with path.open('rb') as source,gzip.open(out/(path.name+'.gz'),'wb',compresslevel=9) as destination:shutil.copyfileobj(source,destination)
shutil.rmtree(logs)
