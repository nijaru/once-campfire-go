import csv,json,math,pathlib,re,subprocess,sys
header=re.compile(r'^M=(-?\d+) P=(-?\d+) G=(-?\d+) (\w+) Time=(\d+) (.*)$')
def field(text,key):
 m=re.search(r'\b'+key+r'=(?:"([^"]*)"|([^ ]+))',text)
 return (m.group(1) if m.group(1) is not None else m.group(2)) if m else None
def stats(values):
 v=sorted(values)
 if not v:return {'n':0}
 return {'n':len(v),'mean_us':sum(v)/len(v)/1000,'p50_us':v[math.ceil(.5*len(v))-1]/1000,'p99_us':v[math.ceil(.99*len(v))-1]/1000,'max_us':v[-1]/1000}
for name in sys.argv[1:]:
 path=pathlib.Path(name);tasks={};spans={};regions={};pauses=[]
 proc=subprocess.Popen(['go','tool','trace','-d=parsed',str(path)],stdout=subprocess.PIPE,text=True,bufsize=1024*1024)
 for line in proc.stdout:
  m=header.match(line)
  if not m:continue
  kind=m[4];ns=int(m[5]);text=m[6];gid=int(m[3])
  if kind=='TaskBegin' and field(text,'Type')=='http.message.create':
   tasks[int(field(text,'ID'))]={'start':ns,'stages':{},'span':0}
  elif kind=='TaskEnd' and int(field(text,'ID')) in tasks:
   tasks[int(field(text,'ID'))]['end']=ns
  elif kind=='Log' and field(text,'Category').startswith('sqlite.begin.'):
   tid=int(field(text,'Task'))
   if tid not in tasks:continue
   cat=field(text,'Category');marker=field(text,'Message')
   stages=tasks[tid]['stages'].setdefault(cat,{})
   if marker in stages:raise ValueError(f'duplicate stage {tid} {cat} {marker}')
   stages[marker]=ns
  elif kind in ('RegionBegin','RegionEnd') and field(text,'Type')=='display.snapshot.begin':
   tid=int(field(text,'Task'))
   if kind=='RegionBegin':regions[(gid,tid)]=ns
   elif tid in tasks:tasks[tid]['span']+=ns-regions.pop((gid,tid))
  elif kind in ('RangeBegin','RangeEnd') and field(text,'Name').startswith('stop-the-world'):
   key=(field(text,'Name'),field(text,'Scope'))
   if kind=='RangeBegin':spans[key]=ns
   else:pauses.append((spans.pop(key),ns))
 if proc.wait():raise RuntimeError('trace parse failed')
 complete=[(tid,t) for tid,t in tasks.items() if 'end' in t]
 complete.sort(key=lambda pair:pair[1]['start'])
 rows=[]
 for tid,t in complete[-9000:]:
  read=t['stages']['sqlite.begin.BEGIN']
  writer=t['stages']['sqlite.begin.BEGIN IMMEDIATE']
  duration=t['end']-t['start']
  row={'task':tid,'start_ns':t['start'],'duration_ns':duration,'display_begin_ns':t['span'],'driver_begin_ns':read['exit']-read['enter'],'read_spawn_ns':read['childstart']-read['parentbefore'],'read_exec_ns':read['childend']-read['childstart'],'read_resume_ns':read['parentresume']-read['childend'],'writer_spawn_ns':writer['childstart']-writer['parentbefore'],'writer_exec_ns':writer['childend']-writer['childstart'],'writer_resume_ns':writer['parentresume']-writer['childend'],'request_stw_ns':sum(max(0,min(t['end'],end)-max(t['start'],start)) for start,end in pauses)}
  row['read_scheduling_ns']=row['read_spawn_ns']+row['read_resume_ns']
  rows.append(row)
 csvpath=path.with_suffix('.begin.csv')
 with csvpath.open('w') as out:
  writer=csv.DictWriter(out,fieldnames=list(rows[0]));writer.writeheader();writer.writerows(rows)
 tail=sorted(rows,key=lambda r:r['duration_ns'])[-90:]
 keys=['duration_ns','display_begin_ns','driver_begin_ns','read_spawn_ns','read_exec_ns','read_resume_ns','read_scheduling_ns','writer_spawn_ns','writer_exec_ns','writer_resume_ns','request_stw_ns']
 summary={'completed_tasks':len(complete),'analyzed_last_tasks':len(rows),'scope':'last 9000 creation tasks: timed phase after 3000 warmup; handler duration excludes final net/http flush and client consumption','all':{k:stats([r[k] for r in rows]) for k in keys},'slowest_1pct_requests':{k:stats([r[k] for r in tail]) for k in keys},'mean_begin_scheduling_fraction_slowest_1pct':sum(r['read_scheduling_ns']/r['duration_ns'] for r in tail)/len(tail),'max_begin_scheduling_fraction_slowest_1pct':max(r['read_scheduling_ns']/r['duration_ns'] for r in tail),'slowest_10':list(reversed(tail[-10:]))}
 path.with_suffix('.begin-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
 print(path.name,json.dumps(summary),flush=True)
