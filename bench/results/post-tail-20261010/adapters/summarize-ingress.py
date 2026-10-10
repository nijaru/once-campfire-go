"""Correlate private HTTP WroteRequest timestamps with handler-entry wall time.

This is diagnostic, not a packet capture: WroteRequest is completion of a client
write to its socket. Server entry can precede that callback on another CPU.
Both processes use the same Linux wall clock; trace durations remain monotonic.
"""
import csv
import json
import math
import sys
from pathlib import Path

ROOT = Path(__file__).parent

def summary(values):
    values = sorted(values)
    if not values:
        return {"n": 0}
    def q(p):
        return values[min(len(values)-1, math.ceil(p*len(values))-1)] / 1000
    return {"n": len(values), "mean_us": sum(values)/len(values)/1000,
            "p50_us": q(.5), "p95_us": q(.95), "p99_us": q(.99),
            "min_us": values[0]/1000, "max_us": values[-1]/1000}

def gaps(values):
    ordered = sorted(values)
    delta = [b-a for a,b in zip(ordered, ordered[1:])]
    return summary(delta) | {"under_100us_fraction": sum(v < 100_000 for v in delta)/len(delta),
                             "over_1ms": sum(v > 1_000_000 for v in delta)}

source = Path(sys.argv[1]) if len(sys.argv) > 1 else ROOT / 'fixed-ingress'
results = {}
for path in sorted(source.glob("*.requests.json")):
    name = path.name.removesuffix(".requests.json")
    tasks = json.loads(path.read_text())
    assert len(tasks) == 12000
    assert all(t['end']-t['begin'] == sum(t['states'].values()) for t in tasks)
    by_id = {t['request_id']: t for t in tasks}
    assert len(by_id) == len(tasks)
    with (path.parent / (name + "-3.send.csv")).open() as f:
        sends = list(csv.DictReader(f))
    assert len(sends) == 9000
    assert all(int(s['wrote_request_ns']) > 0 for s in sends)
    rows = [(s, by_id[s['id']]) for s in sends]
    delay = [t['wall_begin']-int(s['wrote_request_ns']) for s,t in rows]
    results[name] = {
        "timed_requests": len(rows),
        "client_write_gaps": gaps([int(s['wrote_request_ns']) for s,t in rows]),
        "handler_start_gaps": gaps([t['wall_begin'] for s,t in rows]),
        "write_complete_to_handler_start": summary(delay),
        "handler_before_write_callback": sum(d < 0 for d in delay),
        "request_duration": summary([t['duration'] for s,t in rows]),
        "writer_acquire": summary([t['regions']['writer.acquire'] for s,t in rows]),
        "writer_hold": summary([t['regions']['writer.hold'] for s,t in rows]),
    }
    # Compact, independently re-analyzable evidence; no cookies, request/response
    # bodies, full trace events, or runtime databases.
    with (path.parent / (name + ".correlation.csv")).open('w') as f:
        w = csv.writer(f)
        w.writerow(['id','scheduled_wall_ns','enqueued_wall_ns','worker_wall_ns','wrote_wall_ns',
                    'handler_wall_ns','task_begin_ns','task_end_ns','acquire_ns','hold_ns'])
        for s,t in rows:
            w.writerow([s['id'],s['scheduled_ns'],s['enqueued_ns'],s['worker_start_ns'],s['wrote_request_ns'],
                        t['wall_begin'],t['begin'],t['end'],t['regions']['writer.acquire'],t['regions']['writer.hold']])
(ROOT / ('ingress-summary-' + source.name + '.json')).write_text(json.dumps(results, indent=2)+'\n')
print(json.dumps(results, indent=2))
