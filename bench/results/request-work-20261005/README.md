# Remaining per-request preparation

This pass removes three avoidable operations without caching request results:

- Private/non-cacheable HTTP responses no longer clone headers that the public response
  cache immediately discards. Cacheable public headers are still captured at the same point,
  after cookie stripping and before downstream middleware can change them.
- Browser checks and page rendering share one lazily parsed, request-owned User-Agent.
  A single context value also owns host/origin metadata. Authentication/session refresh still
  precede browser rejection; derived contexts share metadata without detached work.
- The bounded 40-message reference query allocates its result slice once, only after its
  first row. Empty results remain nil; query text, scanning and cancellation are unchanged.

A separate candidate gives room shells an explicit rendering input instead of the general
page model. The same input drives both template execution and the cache key, so this is not
an independent hand-maintained validator list. Nested user/room/account values remain complete;
new top-level room-template dependencies must enter the rendering input as well as its key.
Cursor/message bytes remain independent immutable parts. No templates or response bytes changed.

## Application measurements

Same Apple M3 Max Docker Desktop Linux/arm64 VM as previous iterations. Three rotating
repetitions, public HTTP/1.1, 16 clients, five-second samples following two-second warmups.
Identical fresh seed data, active rooms targeting 20 commits/sec, access logging enabled.
Server VM CPUs 0–3/GOMAXPROCS=4; load generator CPUs 4–7. No agent, build or test workloads
ran during timing. VM affinity does not pin physical host cores.

| Encoding/workload | Direct records | Request work | Plus bounded shell | CPU µs/request, records → final |
|---|---:|---:|---:|---:|
| Identity room | 13,334 | 14,611 | 15,319 | 160 → 143 |
| Identity active room | 18,254 | 19,404 | 19,552 | 142 → 130 |
| Identity history | 21,480 | 22,129 | 22,535 | 128 → 124 |
| Identity search | 15,888 | 15,426 | 15,638 | 210 → 208 |
| Identity static CSS | 104,884 | 104,170 | 103,670 | 17.5 → 17.6 |
| Gzip room | 19,039 | 20,797 | 22,152 | 146 → 134 |
| Gzip active room | 26,548 | 27,692 | 27,950 | 131 → 123 |
| Gzip history | 29,797 | 31,009 | 31,659 | 123 → 118 |
| Gzip search | 15,395 | 15,730 | 15,547 | 210 → 209 |
| Gzip static CSS | 104,368 | 102,918 | 103,541 | 17.6 → 17.6 |

Values are median requests/sec from [identity samples](final-0/raw.json) and
[gzip samples](final-1/raw.json). Final room improvement is 15% identity and 16% gzip over
the direct-record candidate in this sequence. Bounded shell input adds roughly 5%/7% room
throughput over the request-work candidate, with smaller CPU reductions. Search/static
results are effectively unchanged; overlapping ranges limit claims for the smaller changes.
Do not combine these ratios with separately timed historical runs.

Median post-workload Pss, MiB:

| Encoding | Direct records | Request work | Plus bounded shell |
|---|---:|---:|---:|
| Identity | 121.8 | 109.0 | 123.9 |
| Gzip | 116.7 | 115.6 | 116.5 |

These are whole-process observations after five workloads, including caches and GC headroom,
not isolated object sizes. Reduced temporary allocation does not establish a resident-memory
reduction. In particular, the identity shell sequence has higher Pss than request work alone.
Fragment/gzip cache budgets are unchanged. Initial room output stays 374,036 plain/21,289 gzip bytes.

All 18 final runs and 90 HTTP samples had zero errors. All 2,497 active-room posts persisted
and matched FTS counts. Complete gzip responses matched identity without cursor masking;
initial candidate semantics matched and final active content was verified.

[Mac](check-mac.txt) and [Linux](check-linux.txt) vet/race gates passed, including the
WebSocket fork. Tests cover private fresh cookies/headers, alternating mobile/desktop platform
HTML, pagination equivalence against full records, timestamp precision/formats/BLOBs,
external commits and cancellation. Expanded room-shell byte comparisons cover actual
Account.JoinCode, logo version, user bio/avatar version, direct-room type, VAPID and platforms.
An independent read-only review found no concrete regressions or material contract gaps.

## Rejected JSON batching

The reference query is still row-based. A bounded SQLite JSON aggregate was tried to reduce
CGo/cancellable row steps, retaining the request context and existing timestamp validation.
It did not demonstrate a consistent application gain: room throughput medians were
18,196 → 18,513, active-room 23,579 → 22,165. Median CPU/request was 150.1 → 147.9 µs for
rooms and 137.4 → 142.6 µs for active rooms, with substantial VM drift across repetitions.
That is insufficient evidence for the additional SQL/JSON allocation and decoding machinery.
The candidate was removed, not shipped.

[Rejected raw samples](rejected-json-batch/raw.json), metadata and lossless logs are retained
with its [source patch](rejected-json-batch/candidate.patch) and binary hash in the manifest.
These probe runs are separate from the final result tables.

## Remaining costs and scope

The separate [final profile](profiles/cumulative.txt) puts message reference reads at 23.2%
cumulative CPU, room-shell preparation at 3.6%, timestamp scanning at 3.2%, and User-Agent
parsing at 1.0%. Public header cloning is gone from private responses. Profiles are excluded
from throughput tables. Most logging cost remains the actual file write; disabling, sampling
or asynchronously buffering it would change the retained operational contract.

Fresh session/authorization reads are retained, as is SQLite cancellation. No query snapshot,
authentication TTL, response-result cache, SQLite API replacement or WebSocket change was
introduced. No browser/media/Cable/TLS or external-integration throughput was remeasured;
existing gates remain the evidence where available. The sidebar layout gap is unchanged.
Container networking was disabled except loopback. These are same-VM Go-to-Go results,
not native AMD or Go/Rust comparisons.

## Reproduce

[harness.json](harness.json) records the three production commits, exact binary hashes,
source staging, rejected candidate and shared environment. Build the direct-record source,
request-work source and bounded-shell source with generated assets, Go 1.27.1, CGO and
`-trimpath -buildvcs=false -tags sqlite_fts5`. Use the pinned seed/load-generator/image
from the [initial environment manifest](../prepared-gzip-20261005/harness.json).

```sh
bench/application --rust-root reference --seed .cache/parity-seeds/default \
  --loadgen .cache/linux/loadgen \
  --candidate records=/path/to/records --candidate request=/path/to/request \
  --candidate shell=/path/to/shell --apps records request shell \
  --listener public --gzip 1 \
  --routes room_show active_room messages_page search static_css \
  --concurrency 16 --reps 3 --seconds 5 --server-cpus 0-3 --loadgen-cpus 4-7 \
  --cable-clients --upload-reps 0 --out bench/results/my-request-run
```

Repeat with `--gzip 0` and a separate output directory. Profile and rejected-candidate runs
are retained separately.
