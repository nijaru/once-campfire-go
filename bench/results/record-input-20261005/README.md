# Direct message-record input

Controllers now pass database messages directly to `render`. Recorded room/history/search
responses no longer allocate message views just to copy their embedded records back into
another slice. Other templates still build their hydrated views in the renderer. Show-message
and boost-index controllers also stop preparing views that the renderer immediately discarded.
All render callers use the new input; direct template execution still receives prepared views.

Authorization, database query definitions, request cancellation, cache identities, validators
and templates are unchanged. The renderer consumes its private record input before assembling
page output. Records are not retained in the shell key or cached as a query snapshot.

## Measurements

Three alternating repetitions, 16 clients, five-second samples after two-second warmups.
Public HTTP/1.1 with access logging enabled, identical fresh seed data and active rooms
at a target 20 commits/sec. Same Apple M3 Max Docker Desktop Linux/arm64 VM as the previous
iterations: server CPUs 0–3/GOMAXPROCS=4, load generator CPUs 4–7. No compilation, tests or
agent workloads ran during timing. VM affinity does not pin physical Apple cores.

| Encoding/workload | Binary-key build | Direct records | Ratio | CPU µs/request, keys → records |
|---|---:|---:|---:|---:|
| Identity room | 11,836 | 13,235 | 1.12× | 169 → 161 |
| Identity active room | 17,507 | 17,942 | 1.02× | 151 → 145 |
| Identity history | 20,082 | 21,416 | 1.07× | 137 → 133 |
| Identity search | 15,466 | 15,557 | 1.01× | 216 → 212 |
| Gzip room | 16,198 | 17,907 | 1.11× | 156 → 147 |
| Gzip active room | 24,348 | 25,362 | 1.04× | 137 → 133 |
| Gzip history | 28,483 | 28,979 | 1.02× | 127 → 124 |
| Gzip search | 15,001 | 15,403 | 1.03× | 214 → 210 |

Values are medians from [identity raw samples](final-0/raw.json) and
[gzip raw samples](final-1/raw.json). The room improvement is clearer than the small changes
elsewhere; overlapping ranges and VM variability limit broader claims. Do not multiply
these ratios by those from earlier experiments. Initial room output remains 374,036 plain
bytes and 21,289 gzip bytes.

Post-workload Pss was 114.4 → 120.8 MiB for identity and 114.5 → 114.9 MiB for gzip.
Eliminating transient slices did not produce a measured resident-memory reduction; the
identity increase is disclosed rather than inferred away. These are whole-process totals
for four workloads, including caches and GC headroom. Cache budgets are unchanged.

All 12 final runs and 48 HTTP samples had zero errors. All 1,661 active posts persisted and
matched FTS counts. Complete decoded gzip responses matched identity without masking;
initial semantics matched across candidates and final active content was verified.
[Mac](check-mac.txt) and [Linux](check-linux.txt) vet/race gates passed, including the
WebSocket fork. A focused HTTP test checks show/edit/boost-index/new-boost content and
fresh message edits through real authenticated routes; existing room-cursor, conditional,
compressed-body and revocation tests remain intact.

The separate [CPU profile](profiles/cumulative.txt) now attributes 23.7% cumulative CPU
to message reference reads and 4.5% to timestamp scanning. Profile timing is excluded from
the tables. This does not establish that replacing SQLite APIs or weakening cancellation
would be safe or beneficial.

No Cable, browser, media, static/post throughput, TLS or external integrations were remeasured.
Their code paths were unchanged; existing gates remain the evidence where available. Container
networking was disabled except loopback. The pre-existing sidebar layout gap is unchanged.
These are same-VM Go-to-Go measurements, not native AMD or Go/Rust comparisons.

## Reproduce

[harness.json](harness.json) identifies production sources, binary hashes, exact source
staging and the shared environment manifest. Build the binary-key commit `12f5e1c` and
the direct-record production commit with generated assets, Go 1.27.1, CGO and
`-trimpath -buildvcs=false -tags sqlite_fts5`. Use the pinned seed/load-generator/Linux image
from the [initial environment manifest](../prepared-gzip-20261005/harness.json).

```sh
bench/application --rust-root reference --seed .cache/parity-seeds/default \
  --loadgen .cache/linux/loadgen \
  --candidate keys=/path/to/keys --candidate records=/path/to/records \
  --apps keys records --listener public --gzip 1 \
  --routes room_show active_room messages_page search \
  --concurrency 16 --reps 3 --seconds 5 --server-cpus 0-3 --loadgen-cpus 4-7 \
  --cable-clients --upload-reps 0 --out bench/results/my-record-run
```

Repeat with `--gzip 0` and a different output directory. Profile artifacts are retained
separately from final throughput samples.
