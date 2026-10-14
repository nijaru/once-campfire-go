# Live critical path and publication handoff — 2026-10-14

**Cooperative handoff rejected; production remains `8351224`.** A current-source
live trace joins every acknowledged POST and all 891,200 created-message server
frame lifecycles. Most enqueue-to-write delay finds the writer runnable, not doing
socket I/O. Yielding after an actual publication burst gives only **+0.94%** median
delivery rate; paced all-client p99 is **2.32% worse**, missing the benefit gate.
No scheduler/writer replacement, queue change or production code lands.

## One bounded live diagnostic

Unlike the earlier zero-socket [POST trace](../post-critical-path-deflate-20261014/README.md),
this capture uses the existing shared Cable shape: 64 distinct authenticated people,
six subscriptions each, 100 paced messages at 100 ms, then five seconds/four
saturated posters and complete draining. The maintained shared client and its exact
ACK/timestamp recorder are unchanged (`464b57…`). Source is exact `8351224` plus a
private execution-trace overlay; no CPU/allocation profile, perf or periodic
sampler runs. Setup/login occurs in an untraced process before restart.

Go 1.27.1/Linux amd64, CGO/FTS5, four runtime workers/readers, three job workers,
standard codec and default GC remain. Physical Fedora/i9 desktop, server CPUs
`0,2,4,6`, client `8,10,12,14`, runner `16,18`; no exclusive SMT/thermal isolation
claim. Native focused Cable/front/web races, vet and WebSocket-fork races/vet pass;
local database/application/web/front/Cable tests and vet pass. Archive extraction
warnings concern macOS attributes, not failed tests. The old embedded asset,
template and PWA trees are retained; this is not an asset-cleanup experiment.

The declared budget is **one capture**, with no repeat or Rust arm. All **13,925
writes** pass exact ACK-ID/body/creator/room/FTS, population, integrity, unchanged
seed FK and zero-presence audits. Every created message joins exactly 64 distinct
successful enqueue/write lifecycles. All HTTP task, functional-region, frame-write
and queued-writer state partitions reconstruct their durations. The additional
queue-state analysis reuses the same trace, not another workload.

| Individually correlated mean | All | Independently selected slowest 1% |
|---|---:|---:|
| Client POST complete-wire service, µs | 1461.37 | 5096.58 |
| HTTP task, µs | 1079.46 | 3676.43 |
| Writer acquisition, µs | 425.19 | 772.35 |
| Writer hold, µs | 87.16 | 241.99 |
| Two publication authorization regions, µs | 183.55 | 696.98 |
| Publication regions including authorization, µs | 257.60 | 1232.17 |
| Server frame enqueue-to-write completion, µs | 428.29 | 3634.90 |
| Server frame queue interval, µs | 420.29 | 3577.24 |
| Writer runnable during queue interval, µs | 394.21 | 2922.14 |
| Writer waiting during queue interval, µs | 14.17 | 587.74 |
| Frame write interval, µs | 8.00 | 57.66 |

HTTP tails select 140 individual client services; frame tails select 8,912
individual server lifecycles separately. These populations are not joined tail
percentiles. Region durations overlap and cannot be added into an exclusive CPU
budget. Writer runnable means ready in Go's trace, not a particular kernel or
scheduler cause; running is not physical CPU and waiting can include child work.
Queue intervals can include other queued/control frames and reader work.

Instrumentation adds trace events and temporary labels and is not assumed neutral.
Its observed 2763.9 delivered/s and 7.083 ms paced p99 are diagnostic, not new
rankings or proof of the uninstrumented Rust gap's cause. Shared Cable POSTs do not
request gzip; response preparation here is not comparable to the gzip fixed-POST
owner. Actual client receive timestamps are not added: a completed server write
is not all-client delivery. The exact client service boundary uses complete wire
before unchanged semantic validation. **5,791 signed HTTP boundary residuals** are
negative and retained, not clamped; socket completion can precede the handler's
final logging/task end. No marginal-percentile subtraction follows.

Joined process CPU is 15.762685 s (user 9.800491/system 5.962194), including startup,
connections, paced/throughput work and shutdown, not isolated publication CPU.
Raw execution trace remains private with its hash. The runner omitted image/seed
hashes from this phase's initial metadata; a postcapture inspection confirms the
established image `619c76…` and seed `691c52…`. It is not relabeled as a before/after
immutability check. The subsequent controls record binary/image identities in
cohort metadata.

## Uninstrumented cooperative-handoff control

The distinct private candidate records whether existing `sendFrame` admits any
frame, then calls `runtime.Gosched()` once after that publication burst. It does
not yield for no admitted frames. Existing fresh SQL authority, frame bytes and
ownership, per-connection order, deadlines, cancellation, joined shutdown and
256-frame queue remain unchanged. This tests ready-writer scheduling, not the
rejected contiguous-frame, lifetime-cancellation, encoding or authorization paths.
Native affected Cable/application/web races and vet pass. Baseline `8a1409c…` is
byte-identical to production; candidate is `6e148427…`.

Before timing, the gate requires **at least 8% median delivery-rate improvement OR
10% median paced all-client p99 reduction**. The other measure and saturated POST
p99 may not regress more than 5%. The finite budget is three alternating pairs,
300 paced messages at 100 ms plus the unchanged five-second saturated phase.
Only a pass permits three fixed-POST and three capacity pairs; they do not run.
No hooks, profilers, process sampling, task builds/tests or agents overlap timing.

| Pair/order | Current delivery/s | Candidate delivery/s | Current paced p99, ms | Candidate paced p99, ms | Current saturated POST p99, ms | Candidate POST p99, ms |
|---|---:|---:|---:|---:|---:|---:|
| 1, current/candidate | 3056.6 | 3064.9 | 6.711 | 6.723 | 3.271 | 3.233 |
| 2, candidate/current | 3061.9 | 3090.8 | 6.755 | 6.927 | 3.231 | 3.215 |
| 3, current/candidate | 3064.9 | 3102.9 | 6.223 | 6.867 | 3.301 | 3.153 |
| Median | 3061.9 | 3090.8 | 6.711 | 6.867 | 3.271 | 3.215 |

Delivery rises **0.94%**, paced p99 worsens **2.32%** (loses every pair), and
saturated POST p99 improves **1.71%**. The third paced loss is 10.35% and remains
valid; the preregistered gate uses medians, not an invented individual-run limit.
Non-regression passes but meaningful benefit fails. Three short rounds do not
establish precise confidence/noise bounds. No refinement, second capture, wait
policy, timer/deadline change or further timing controls follow. The fact that
queue delay is substantial does not make this handoff a useful repair.

All **94,035 control writes**, and **107,960 total writes** across both phases,
pass the unchanged shared all-client validator and exact persistence/FTS/integrity/
FK/presence audits. No dropped, duplicate or missing created-message deliveries
are observed. Earlier correctness/shared checks on unchanged production remain
valid; no permanent tests, dependency or code change is promoted, so a new broad
browser/media acceptance run is unnecessary. Fixed-POST and Rust live-delivery
optimization remain unfinished.

## Evidence

[`manifest.json`](manifest.json) retains budgets, complete overlays, adapters,
checks, all individual control results and ACK timestamps, complete HTTP
correlations and a compact full frame population. `frames-columnar.json.gz`
defines its columns and trace origin once; each row retains sequence/user/writer,
enqueue offset, queue/write durations and both state partitions. Write begin/end
are reconstructed from enqueue plus those durations. No rows are discarded.

Raw runtime trace and expanded frame dictionaries stay private with hashes;
credentials/cookies, database/storage, private keys, binaries and runtime logs are
not published. Numeric benchmark identities/partitions do not contain session
credentials. Code paths, not benchmark ranks or this trace alone, must ground any
subsequent candidate. This closes the cooperative-handoff hypothesis, not all
possible publication/scheduling improvements.
