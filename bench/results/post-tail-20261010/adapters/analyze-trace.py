import collections, json, math, pathlib, re, subprocess, sys, time
root=pathlib.Path(__file__).parent
header=re.compile(r'^M=(-?\d+) P=(-?\d+) G=(-?\d+) (\w+) Time=(\d+) (.*)$')
def field(text,key):
 m=re.search(r'\b'+key+r'=(?:"([^"]*)"|([^ ]+))',text)
 return (m.group(1) if m.group(1) is not None else m.group(2)) if m else None
def stats(values):
 values=sorted(values)
 if not values:return {'count':0}
 q=lambda p:values[min(len(values)-1,math.ceil(p*len(values))-1)]/1e6
 return {'count':len(values),'sum_ms':sum(values)/1e6,'mean_ms':sum(values)/len(values)/1e6,'p50_ms':q(.5),'p95_ms':q(.95),'p99_ms':q(.99),'max_ms':values[-1]/1e6}
def analyze(path):
 start=time.monotonic();tasks={};active={};states={};regions={};ranges={};pauses=[];gc=collections.defaultdict(list)
 events=0
 proc=subprocess.Popen(['go','tool','trace','-d=parsed',str(path)],stdout=subprocess.PIPE,text=True,bufsize=1024*1024)
 def account(gid,when):
  if gid in active and gid in states:
   state,since=states[gid];tasks[active[gid]]['states'][state]+=when-since
  if gid in states:states[gid]=(states[gid][0],when)
 for line in proc.stdout:
  m=header.match(line)
  if not m:continue
  events+=1;gid=int(m[3]);kind=m[4];when=int(m[5]);text=m[6]
  if kind=='StateTransition' and field(text,'GoID') is not None:
   g=int(field(text,'GoID'));account(g,when)
   state=re.search(r' (\w+)->(\w+)',text)[2];states[g]=(state,when)
  elif kind=='TaskBegin' and field(text,'Type')=='http.message.create':
   ident=int(field(text,'ID'));tasks[ident]={'begin':when,'gid':gid,'regions':{},'region_spans':{},'states':collections.defaultdict(int),'assist':0}
   active[gid]=ident;states[gid]=('Running',when)
  elif kind=='Log' and field(text,'Category')=='benchmark.ingress':
   ident=int(field(text,'Task'));message=field(text,'Message')
   if ident in tasks and message:
    request_id,wall=message.split(',');tasks[ident]['request_id']=request_id;tasks[ident]['wall_begin']=int(wall)
  elif kind=='TaskEnd':
   ident=int(field(text,'ID'))
   if ident in tasks:
    task=tasks[ident];g=task['gid'];account(g,when);task['end']=when;active.pop(g,None)
  elif kind in ('RegionBegin','RegionEnd'):
   ident=int(field(text,'Task'));name=field(text,'Type');key=(gid,ident,name)
   if kind=='RegionBegin':regions[key]=when
   elif key in regions:
    began=regions.pop(key);elapsed=when-began
    if ident in tasks:
     tasks[ident]['regions'][name]=tasks[ident]['regions'].get(name,0)+elapsed
     tasks[ident]['region_spans'][name]=[began,when]
  elif kind in ('RangeBegin','RangeEnd'):
   name=field(text,'Name');scope=field(text,'Scope');key=(name,scope)
   if kind=='RangeBegin':ranges[key]=when
   elif key in ranges:
    began=ranges.pop(key);gc[name].append(when-began)
    if name.startswith('stop-the-world'):pauses.append((began,when))
    if name=='GC mark assist' and scope.startswith('Goroutine('):
     g=int(re.search(r'\d+',scope)[0])
     if g in active:tasks[active[g]]['assist']+=when-began
 if proc.wait()!=0:raise RuntimeError('trace parser failed')
 rows=[]
 for task in tasks.values():
  if 'end' not in task:continue
  task['duration']=task['end']-task['begin'];task['stw_overlap']=sum(max(0,min(task['end'],end)-max(task['begin'],begin)) for begin,end in pauses)
  rows.append(task)
 result={'parse_seconds':time.monotonic()-start,'events':events,'requests':stats([t['duration'] for t in rows]),'writer':{name:stats([t['regions'].get(name,0) for t in rows]) for name in ('writer.acquire','writer.hold')},'request_states':{state:stats([t['states'].get(state,0) for t in rows]) for state in ('Running','Runnable','Waiting','Syscall')},'gc_ranges':{name:stats(vals) for name,vals in gc.items() if 'GC' in name or 'stop-the-world' in name},'request_assist':stats([t['assist'] for t in rows]),'request_stw_overlap':stats([t['stw_overlap'] for t in rows])}
 # Keep per-request intervals for correlation without dumping the enormous parsed stream.
 target=path.with_suffix('.requests.json');target.write_text(json.dumps(rows,separators=(',',':')))
 return result
all_results={}
source=pathlib.Path(sys.argv[1]) if len(sys.argv)>1 else root/'fixed-traces'
summary=root/('trace-summary.json' if source.name=='fixed-traces' else 'trace-summary-'+source.name+'.json')
paths=[source] if source.is_file() else sorted(source.glob('*.trace'))
for path in paths:
 print('Analyzing',path.name,flush=True);all_results[path.name]=analyze(path)
 summary.write_text(json.dumps(all_results,indent=2)+'\n')
 print(json.dumps(all_results[path.name]),flush=True)
