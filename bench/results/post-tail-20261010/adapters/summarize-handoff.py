"""Extract bounded writer-handoff evidence from the second private GC-off pair.

Input files are analyze-trace.py outputs, omitted from the public archive.
Only those two traces were reparsed with region_spans before this extraction.
"""
import csv
import gzip
import io
import json
import math
from pathlib import Path
import sys

source, output = map(Path, sys.argv[1:])
output.mkdir(parents=True, exist_ok=True)
results = {}
for name in ('go-2', 'go-before-2'):
    tasks = sorted(json.loads((source/(name+'.requests.json')).read_text()), key=lambda t:t['begin'])
    assert len(tasks) == 12000
    tasks = tasks[3000:]  # Separate warmup generator completed before timed generator.
    text = io.StringIO()
    writer = csv.writer(text)
    writer.writerow(['task_begin_ns','task_end_ns','acquire_begin_ns','acquire_end_ns','hold_begin_ns','hold_end_ns'])
    for task in tasks:
        assert task['end']-task['begin'] == sum(task['states'].values())
        acq, hold = (task['region_spans'][k] for k in ('writer.acquire','writer.hold'))
        assert acq[1]-acq[0] == task['regions']['writer.acquire']
        assert hold[1]-hold[0] == task['regions']['writer.hold']
        writer.writerow([task['begin'],task['end'],*acq,*hold])
    with (output/(name+'.writer.csv.gz')).open('wb') as file:
        with gzip.GzipFile(filename='', mode='wb', fileobj=file, mtime=0) as compressed:
            compressed.write(text.getvalue().encode())
    ordered = sorted(tasks, key=lambda t:t['region_spans']['writer.hold'][0])
    gaps = []
    for previous, task in zip(ordered, ordered[1:]):
        ended = previous['region_spans']['writer.hold'][1]
        if task['region_spans']['writer.acquire'][0] < ended:
            gaps.append(task['region_spans']['writer.hold'][0]-ended)
    assert all(gap >= 0 for gap in gaps)
    gaps.sort()
    results[name] = {'timed_tasks':len(tasks), 'already_acquiring_cases':len(gaps),
        'mean_us':sum(gaps)/len(gaps)/1000,
        'p50_us':gaps[math.ceil(.5*len(gaps))-1]/1000,
        'p99_us':gaps[math.ceil(.99*len(gaps))-1]/1000}
(output/'handoff-summary.json').write_text(json.dumps(results,indent=2)+'\n')
print(json.dumps(results,indent=2))
