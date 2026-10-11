import json
import pathlib
import statistics
import sys

root = pathlib.Path(sys.argv[1]) if len(sys.argv) > 1 else pathlib.Path(__file__).parent / 'capture'
results = []
for result in json.loads((root / 'results.json').read_text()):
    n = result['round']
    rows = json.loads((root / f'current-{n}.proc.json').read_text())
    events = {}
    for line in (root / f'current-{n}.phases.log').read_text().splitlines():
        if line.startswith('PHASE '):
            _, phase, stamp = line.split()
            events[phase] = int(stamp) * 1_000_000
    begin = events['saturated']
    end = events['saturated_posted']
    a = min(rows, key=lambda row: abs(row['unix_ns'] - begin))
    b = min(rows, key=lambda row: abs(row['unix_ns'] - end))
    wall = (b['monotonic_ns'] - a['monotonic_ns']) / 1e9
    output = {'round': n, 'delivered_per_second': result['cable']['throughput']['delivered_msgs_per_sec'],
              'posted': result['cable']['throughput']['posted'], 'post': result['cable']['throughput']['post'],
              'wall_seconds': wall, 'boundary_error_ms': [(a['unix_ns']-begin)/1e6, (b['unix_ns']-end)/1e6],
              'sample_count': sum(begin <= row['unix_ns'] <= end for row in rows)}
    for owner in ['server', 'client']:
        old = a[owner]
        new = b[owner]
        common = old['threads'].keys() & new['threads'].keys()
        stats = {'cpu_seconds': (new['cpu_ticks']-old['cpu_ticks']) / 100,
                 'cpu_cores': (new['cpu_ticks']-old['cpu_ticks']) / 100 / wall,
                 'common_threads': len(common), 'new_threads': len(new['threads'].keys() - old['threads'].keys()),
                 'missing_threads': len(old['threads'].keys() - new['threads'].keys()),
                 'cpu_max': old['cpu_max'], 'runnable_wait_seconds_common_threads': 0.0,
                 'context_switches': {'voluntary': 0, 'involuntary': 0}}
        for tid in common:
            ot, nt = old['threads'][tid], new['threads'][tid]
            if ot['runnable_wait_ns'] is not None and nt['runnable_wait_ns'] is not None:
                stats['runnable_wait_seconds_common_threads'] += (nt['runnable_wait_ns']-ot['runnable_wait_ns']) / 1e9
            for kind, key in [('voluntary', 'voluntary_ctxt_switches'), ('involuntary', 'nonvoluntary_ctxt_switches')]:
                stats['context_switches'][kind] += nt['switches'][key] - ot['switches'][key]
        if old['cgroup_cpu'] and new['cgroup_cpu']:
            stats['cgroup_cpu_delta'] = {key: new['cgroup_cpu'][key]-value for key,value in old['cgroup_cpu'].items()}
        output[owner] = stats
    for name, cpus in [('server_cpus', range(4)), ('client_cpus', range(4,8))]:
        total = busy = stolen = 0
        for cpu in cpus:
            aa, bb = a['cpu'][f'cpu{cpu}'], b['cpu'][f'cpu{cpu}']
            delta = [bv-av for av,bv in zip(aa,bb)]
            # guest and guest_nice are already included in user/nice.
            total += sum(delta[:8])
            busy += sum(delta[:3]) + sum(delta[5:8])
            stolen += delta[7]
        output[name] = {'busy_fraction': busy/total, 'steal_fraction': stolen/total}
    overhead = [row['observation_ns']/1e6 for row in rows if begin <= row['unix_ns'] <= end]
    output['observation_ms'] = {'median': statistics.median(overhead), 'max': max(overhead)}
    results.append(output)
(root / 'phase-analysis.json').write_text(json.dumps(results, indent=2)+'\n')
for row in results:
    print(f"{row['round']}: delivered={row['delivered_per_second']} server_cores={row['server']['cpu_cores']:.2f} client_cores={row['client']['cpu_cores']:.2f} runnable_wait_server={row['server']['runnable_wait_seconds_common_threads']:.3f}s runnable_wait_client={row['client']['runnable_wait_seconds_common_threads']:.3f}s post_p99={row['post']['p99_ms']}")
