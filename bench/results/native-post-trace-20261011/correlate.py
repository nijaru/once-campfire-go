import csv
import json
import math
from pathlib import Path
import statistics

root=Path(__file__).parent/'capture'
summary={}
for stem in ['go-before-1','go-1','go-2','go-before-2']:
    tasks=json.loads((root/(stem+'.requests.json')).read_text())
    clients=list(csv.DictReader((root/(stem+'-client.csv')).open()))
    timed={t['request_id']:t for t in tasks if 'request_id' in t}
    assert len(clients)==len(timed)==9000,(stem,len(clients),len(timed))
    assert len(set(c['request_id'] for c in clients))==9000
    joined=[]
    for client in clients:
        task=timed[client['request_id']]
        start,finish=int(client['started_unix_ns']),int(client['finished_unix_ns'])
        ingress=task['wall_begin']
        row={'request_id':client['request_id'],'client_service_ns':finish-start,'client_latency_ns':finish-int(client['scheduled_unix_ns']),
             'server_handler_ns':task['duration'],'before_ingress_ns':ingress-start,'after_handler_ns':finish-ingress-task['duration'],
             'writer_acquire_ns':task['regions'].get('writer.acquire',0),'writer_hold_ns':task['regions'].get('writer.hold',0),
             'server_running_ns':task['states'].get('Running',0),'server_runnable_ns':task['states'].get('Runnable',0),
             'server_waiting_ns':task['states'].get('Waiting',0),'server_syscall_ns':task['states'].get('Syscall',0),
             'server_stw_overlap_ns':task['stw_overlap'],'server_assist_ns':task['assist'],'wire_bytes':int(client['wire_bytes'])}
        assert abs(row['client_service_ns']-(row['before_ingress_ns']+row['server_handler_ns']+row['after_handler_ns']))<=1
        joined.append(row)
    def describe(rows):
        keys=[k for k in rows[0] if k not in ['request_id','wire_bytes']]
        out={}
        for key in keys:
            values=sorted(r[key]/1000 for r in rows)
            out[key.removesuffix('_ns')+'_us']={'mean':statistics.mean(values),'p50':values[len(values)//2],'p99':values[math.ceil(len(values)*.99)-1],'max':max(values),'min':min(values)}
        out['count']=len(rows)
        out['mean_wire_bytes']=statistics.mean(r['wire_bytes'] for r in rows)
        out['negative_boundary_intervals']=sum(r['before_ingress_ns']<0 or r['after_handler_ns']<0 for r in rows)
        out['requests_overlapping_stw']=sum(r['server_stw_overlap_ns']>0 for r in rows)
        out['requests_assisting_gc']=sum(r['server_assist_ns']>0 for r in rows)
        return out
    tail=sorted(joined,key=lambda r:r['client_service_ns'])[-90:]
    summary[stem]={'all':describe(joined),'slowest_client_service_one_percent':describe(tail),'missing_task_ids':0,'missing_writer_regions':sum('writer.acquire' not in t['regions'] or 'writer.hold' not in t['regions'] for t in timed.values())}
    with (root/(stem+'-correlated.csv')).open('w') as f:
        writer=csv.DictWriter(f,fieldnames=list(joined[0]));writer.writeheader();writer.writerows(joined)
    s=summary[stem]['slowest_client_service_one_percent']
    print(stem,'tailmeanus',{k:round(s[k]['mean'],2) for k in ['client_service_us','server_handler_us','before_ingress_us','after_handler_us','writer_acquire_us','writer_hold_us','server_runnable_us','server_waiting_us','server_stw_overlap_us','server_assist_us']},'GCoverlaps',s['requests_overlapping_stw'],'assists',s['requests_assisting_gc'],'missingwriter',summary[stem]['missing_writer_regions'])
(root/'correlation-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
