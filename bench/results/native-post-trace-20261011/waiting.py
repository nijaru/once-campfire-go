import collections
import bisect
import csv
import json
from pathlib import Path
import re
import subprocess

root=Path(__file__).parent/'capture'
header=re.compile(r'^M=(-?\d+) P=(-?\d+) G=(-?\d+) (\w+) Time=(\d+) (.*)$')
def field(text,key):
    match=re.search(r'\b'+key+r'=(?:"([^"]*)"|([^ ]+))',text)
    return (match.group(1) if match.group(1) is not None else match.group(2)) if match else None

def events(stream):
    item=None
    for line in stream:
        match=header.match(line)
        if match:
            if item is not None: yield item
            item=[match,[]]
        elif item is not None: item[1].append(line.rstrip())
    if item is not None: yield item

summary={}
for stem in ['go-before-1','go-1','go-2','go-before-2']:
    tasks=json.loads((root/(stem+'.requests.json')).read_text())
    timed={t['request_id']:t for t in tasks if 'request_id' in t}
    by_g=collections.defaultdict(list)
    for ident,t in timed.items(): by_g[t['gid']].append((t['begin'],t['end'],ident))
    for intervals in by_g.values(): intervals.sort()
    starts={gid:[r[0] for r in intervals] for gid,intervals in by_g.items()}
    spans=[]; waiting={}
    proc=subprocess.Popen(['go','tool','trace','-d=parsed',str(root/(stem+'.trace'))],stdout=subprocess.PIPE,text=True,bufsize=1024*1024)
    for match,stack in events(proc.stdout):
        if match[4]!='StateTransition': continue
        text=match[6]
        if field(text,'GoID') is None: continue
        gid=int(field(text,'GoID')); stamp=int(match[5])
        old,new=re.search(r' (\w+)->(\w+)',text).groups()
        if old=='Waiting' and gid in waiting:
            began,reason,frames=waiting.pop(gid)
            intervals=by_g[gid]
            index=max(0,bisect.bisect_right(starts[gid],began)-1)
            while index<len(intervals) and intervals[index][0]<stamp:
                start,end,ident=intervals[index]
                duration=max(0,min(stamp,end)-max(began,start))
                if duration:
                    spans.append({'request_id':ident,'duration_ns':duration,'reason':reason,'stack':frames})
                index+=1
        if new=='Waiting' and gid in by_g:
            frames=[]
            for line in stack:
                if line=='Stack=': break
                if ' @ ' in line: frames.append(line.strip().split(' @ ',1)[0])
            waiting[gid]=(stamp,field(text,'Reason'),frames)
    if proc.wait()!=0: raise RuntimeError('trace parser failed')
    per_request=collections.defaultdict(lambda:collections.defaultdict(int))
    for span in spans:
        # Keep all transition frames; labels use library boundaries without
        # interpreting a wait as physical disk, CPU starvation or wall-time stall.
        frames=span['stack']; key=' | '.join(f for f in frames if not f.startswith('runtime.'))
        per_request[span['request_id']][key]+=span['duration_ns']
    clients=list(csv.DictReader((root/(stem+'-correlated.csv')).open()))
    slowest={c['request_id'] for c in sorted(clients,key=lambda c:int(c['client_service_ns']))[-90:]}
    def describe(ids):
        totals=collections.Counter()
        for ident in ids: totals.update(per_request[ident])
        return [{'stack':key,'total_ms':value/1e6,'mean_us_per_selected_request':value/1000/len(ids)} for key,value in totals.most_common()]
    summary[stem]={'all':describe(list(timed)), 'slowest_client_service_one_percent':describe(slowest),'requests':len(timed),'wait_spans':len(spans)}
    (root/(stem+'-waits.json')).write_text(json.dumps(spans,separators=(',',':'))+'\n')
    print(stem, json.dumps(summary[stem]['slowest_client_service_one_percent'][:6]))
(root/'waiting-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
