# Fork `8351224` versus upstream Go — native, 2026-10-11

**Fork source `8351224` versus upstream `c4ab53d`: higher POST capacity and
live-delivery throughput, but higher fixed-load POST p99.** This is the latest native
upstream comparison, not a relabeling of older Go/Rust or VM measurements. The
[local POST cache-observation guard](../post-cache-observation-20261012/README.md)
is not included and is now [reverted after its native controls](../post-cache-observation-native-20261012/README.md).
Production again matches the application source measured here.

Three sequential interleaved pairs per workload use the Fedora desktop, matched
fixtures/settings, complete responses and applicable exact write/delivery audits.
Every pair has the same direction for the headline metrics below. No unusually
slow phase occurs in this cohort; all **830,687 warm/timed writes** pass audits.

| Native run medians | Current fork | Upstream Go | Fork change |
|---|---:|---:|---:|
| Fixed-load POST p99, ms | 0.954 | 0.810 | **+17.8%**, worse |
| POST capacity, requests/s | 9,861.3 | 9,182.0 | **+7.4%** |
| Capacity-run POST p99, ms | 6.091 | 6.543 | −6.9% |
| Cable delivered messages/s | 3,099.7 | 2,368.5 | **+30.9%** |
| Cable paced all-client p99, ms | 6.391 | 7.315 | **−12.6%** |
| Cable saturated POST p99, ms | 3.291 | 4.089 | −19.5% |

Capacity-run and fixed-load tails answer different questions. Neither substitutes
for the other. Medians are medians of run statistics, not pooled percentiles.
This is not a claim about every route, fixture, population or external provider.
Rust and sparse-search comparisons are not rerun in this cohort.

## Sources and measurement budget

GitHub upstream `main` was checked before and after the comparison and remains
`c4ab53d00b0a19a55962e03d18057ca6c3da9437`, unchanged from the historical native
comparison. Its latest commit updates KLOC documentation. Current production is
`8351224b56c9a3c6784ed5d7510857212133ac4c`.

Accepted, source-identified binaries are reused, not old experimental candidates:

| Binary | SHA-256 |
|---|---|
| Current `8351224` | `8a1409c09aa79e188010fb04dfa0c63b66e4f2748de734a117f6ae5db49b0e29` |
| Upstream `c4ab53d` | `70cca09f7fa45eac8e383a2160990ac1ad9a1bdd9c6b95f47dc666b60b987b54` |

Their build/source provenance remains in the linked manifests for
[publication-local maps](../publication-session-map-20261011/manifest.json) and the
[earlier native build](../native-rust-search-20261011/manifest.json). Both run in the
same isolated rootless image, Go **1.27.1 linux/amd64**, CGO/FTS5, `-trimpath` and
`-buildvcs=false`, with no private server allocation/profiling hook enabled.

The question is current status against unchanged upstream, not an implementation
experiment or a retrospective slowdown diagnosis. The finite budget is three
native pairs for each existing workload, order upstream/current, current/upstream,
upstream/current. No follow-up sweep is pending. No task-owned build, test, agent,
profiler or server process-counter observer overlaps timing.

## Workloads and environment

Native Fedora x86_64 on the i9-13900KF. Server affinity `0,2,4,6`, client
`8,10,12,14`, runner `16,18`; affinity is not exclusive host reservation, SMT or
thermal isolation. A preflight process listing was mostly quiet, but does not prove
absence of background work during every request. Default application cache budgets
are not equalized. Four runtime workers/readers, three job workers, gzip and
`GOGC=100`; no host configuration, dependency or production tuning changes.

Fresh seed copies use SHA-256
`691c52bfdc30b38bd8867e93b0914613df20e03d4cd7b00796ab68e0a5d26c9a`.
The shared fixture/tool checkpoint is
`ec02debb3e44f40b8778a6a42db3c4515b2f9eaf`.

- Fixed POST: 3,000 offered/s, 64-worker bound, active pacing, two warmup seconds
  then eight timed seconds. All offered requests are scheduled and complete; the
  common private client validates decoded Turbo responses and retains exact
  ACK-ID/sequence records for independent persistence/creator/room/FTS audits.
- POST capacity: two-second/four-client warmup, eight timed seconds at 16 clients;
  unchanged shared semantic-response and exact ACK-ID/body checks. Raw capacity
  receipt files are verified before runtime cleanup but not retained; result/audit
  summaries are retained.
- Cable: 64 independently verified authenticated fixture people, 300 paced
  messages at 100 ms, five saturated seconds/four posters, complete all-client
  delivery and exact ACK-ID/body/creator/room/FTS/presence/integrity audits. The
  seed's one FK orphan remains unchanged, with no new violation.

The common fixed client and minimal Cable ACK/timestamp recorder differ from older
client populations; overhead neutrality is not established. These are matched
server-uninstrumented current controls, not directly comparable historical capacity
or instrumentation-cost estimates. HTTP seed-grounded route preflights pass on
both versions. The raw inherited HTTP metadata still includes an obsolete
“not semantic content” clause; the measured private fixed client does validate
decoded Turbo responses, as its source/build provenance and ACK records establish.

## All individual runs

| Workload | Upstream rounds 1 / 2 / 3 | Current rounds 1 / 2 / 3 |
|---|---|---|
| Fixed POST p99, ms | 0.810 / 0.811 / 0.808 | 0.955 / 0.926 / 0.954 |
| POST capacity, requests/s | 9,188.5 / 9,140.6 / 9,182.0 | 9,861.3 / 9,866.7 / 9,822.9 |
| Cable delivered messages/s | 2,379.2 / 2,368.5 / 2,363.3 | 3,092.4 / 3,099.7 / 3,107.7 |
| Cable paced all-client p99, ms | 7.315 / 7.659 / 7.203 | 6.367 / 6.419 / 6.391 |

Audit counts are **180,000 fixed POST**, **566,807 capacity POST** and **83,880
Cable writes**, including warmup/paced writes. Zero invalid/dropped offered requests
or missing acknowledged delivery. This is workload-specific correctness evidence,
not a whole-program audit or live-provider acceptance.

## Scope of older measurements

Earlier local/VM runs overlapped other project builds and workloads. Those contended
samples are archived, excluded from current performance decisions and not unfinished
reproduction tasks. The older Go/Rust comparison also uses a different Go source,
`811ae04`, not `8351224`. Raw historical tables remain intact.

The fixed-load POST disadvantage here is independently observed in every matched
native pair at `8351224`. Diagnose current costs on their own evidence rather than
reopening old local timing.

## Evidence and next decision

[`manifest.json`](manifest.json) hashes 75 retained artifacts, measured source trees
and accepted binary identities. Compressed text includes all raw fixed/Cable ACK
records, individual samples/audits, exact adapters, metadata, logs and summaries.
Runtime secrets, session files and databases are excluded. No production code
changes or new correctness defect result from this comparison.

Snapshot `8351224` is correctness-checked, with measured bulk-POST/live-delivery
advantages and a **0.144 ms fixed-load POST p99 gap** in this workload.
Further work should investigate current native creation costs while preserving its
contracts, rather than repeat closed controls or treat historical variability as a
standalone readiness gate.
