# Owned JSON byte copies — 2026-10-11

**Rejected; production unchanged.** Removing two redundant Go byte copies reduced
sampled Go allocation bytes/message by **9.62%**, with effectively flat sampled
CPU. Initial complete-delivery controls were mixed but near unchanged. The sole
longer paced follow-up lost materially: all-client p99 **+7.33%**, saturated
delivery rate **−12.21%**. The cause is unclassified; the allocation reduction does
not justify dismissing that result or promoting the combined candidate.

Application base is `f671137`, production-identical to `ccfa838`; shared
verification is `ec02deb`. Desktop SSH timed out. All measurements below are
Linux ARM64 in the Mac's Docker VM, not native or general-capacity evidence.

## SQL access-path discriminator

Existing CPU profiles identify publication authorization as about 12–13% of
sampled CPU. A bounded probe used the application's Go driver/bundled SQLite
**3.53.4**, real schema, 1/64 distinct sessions and scoped/unscoped authority.
`IN (SELECT value FROM json_each(?))` already uses the unique token index and user
primary key; scoped membership uses the covering room/user index. A direct join
from the JSON array removes the list subquery but retains those access paths.

Three one-second repetitions per isolated case used cancellable contexts and
complete row/map consumption. At 64 sessions, median elapsed query cost was
**70.906→64.410 µs unscoped (−9.16%)** and **75.960→69.133 µs scoped (−8.99%)**;
allocation counts were unchanged. One-session medians also decreased. These are
isolated elapsed costs, not SQLite CPU or application gains. The probe order was
batched rather than interleaved, and only `GOMAXPROCS=4` was fixed. Identity and
duplicate-token results matched. The savings were too small to justify application
captures from this premise; **the direct-join query was not promoted or combined
with the byte-copy candidate**.

## Allocation capture and candidate

One current allocation capture used the existing 64-distinct-person Cable
workload: six subscriptions each, 100 paced messages at 100 ms, four saturated
posters for five seconds, presence refreshes and complete all-client draining.
Setup/logins/stream scraping run in an unprofiled process; the server restarts on
the same disposable database before the workload. A private shutdown-only hook
writes cumulative `allocs` samples after joined shutdown and two final GCs.
`MemProfileRate` remains its default 524,288 bytes. Startup, connection setup,
posting, disconnect/shutdown and audit-idle time remain included; fixed costs are
not subtracted. The hook and forced GCs are not production changes or throughput
instrumentation.

Current sampled Go allocations were about **202.6 KiB/message**. Two visible
copies accounted for roughly 95 MiB in `NewPreparedMessage` and 63 MiB in
`string(raw)` authorization arguments. The private combined candidate:

- Binds the freshly marshaled authorization JSON as `[]byte`, using
  `json_each(CAST(? AS TEXT))` to preserve JSON text semantics rather than rely on
  SQLite's legacy blob parsing. `IN`, session/active-status/membership predicates,
  ordinary rows, statement snapshot and context cancellation stay unchanged.
- Adds an explicitly ownership-consuming prepared-message constructor. The
  existing borrowed constructor still clones. All three Hub callers transfer
  fresh `json.Marshal` output without later mutation or buffer reuse. Compression,
  masking/context-takeover fallbacks, write hooks/deadlines and queues are unchanged.

Budget: one equivalent candidate allocation capture; require ≥6% lower sampled
Go bytes/message, then two interleaved CPU pairs without a consistent material
CPU regression, followed by bounded uninstrumented controls.

| Capture | Audited messages | Sampled allocated bytes | Bytes/message | Objects/message |
|---|---:|---:|---:|---:|
| Current | 9,440 | 1,958,184,405 | 207,434.79 | 2175.58 |
| Candidate | 10,252 | 1,921,961,230 | 187,471.83 | 2125.03 |

Sampled Go bytes/message fell **9.62%**, objects/message **2.32%**. This does not
measure native SQLite allocations or establish latency/capacity improvement.

CPU controls used the same private hook in both images, with allocation emission
and forced GCs disabled. Order: current/candidate, candidate/current. CPU profiling
starts after the authentication restart and flushes on graceful shutdown.

| Pair | Current CPU ms/message | Candidate CPU ms/message | Change |
|---|---:|---:|---:|
| 1 | 1.0916 | 1.0947 | +0.28% |
| 2 | 1.1043 | 1.0887 | −1.42% |

Median normalized CPU change is **−0.57%**, effectively unchanged in these coarse
samples. No task-owned build/test overlapped captures. Affinity and default
sampling do not establish an error bound or isolate native execution costs.

## Uninstrumented complete-delivery controls

Both arms use production binaries without the private allocation hook. Go 1.27.1,
CGO/FTS5, `-trimpath`, `-buildvcs=false`, four runtime workers, three job workers and
`GOGC=100` match. Server CPUs 0–3, client CPUs 4–7, runner CPUs 8–9; affinity is not
VM isolation. No task-owned agents, builds, tests or profilers overlapped timing.
The common minimal client ACK/timestamp recorder remains; its overhead is not
isolated. These are not relabeled historical pristine or native comparisons.

Three alternating 64-person pairs use the original workload:

| Pair/order | Current delivered/s | Candidate delivered/s | Current paced all-client p99 ms | Candidate paced p99 ms | Current saturated POST p99 ms | Candidate POST p99 ms |
|---|---:|---:|---:|---:|---:|---:|
| 1, current/candidate | 2010.5 | 2028.1 | 14.879 | 13.983 | 8.215 | 8.247 |
| 2, candidate/current | 2023.1 | 2052.4 | 11.495 | 11.495 | 8.351 | 8.399 |
| 3, current/candidate | 1986.6 | 2038.4 | 13.055 | 14.423 | 8.231 | 8.143 |

Median delivery rate is +1.39%, saturated POST p99 +0.19%, but paced all-client
p99 +7.11%, with mixed individual results. Those thin 100-message tails motivated
the one longer follow-up below, not a general performance claim.

Two alternating shared-cookie/64-socket pairs retain the same fixture (64 added
people remain unused by these sockets). All sockets subscribe to the creator's
six channels/streams; audits check the creator's presence teardown.

| Pair/order | Current delivered/s | Candidate delivered/s | Current paced p99 ms | Candidate paced p99 ms | Current POST p99 ms | Candidate POST p99 ms |
|---|---:|---:|---:|---:|---:|---:|
| 1, current/candidate | 2318.2 | 2375.8 | 13.487 | 11.455 | 7.471 | 7.211 |
| 2, candidate/current | 2321.6 | 2219.7 | 11.911 | 11.415 | 7.071 | 7.515 |

Shared-cookie medians: delivery −0.95%, paced p99 −9.95%, saturated POST p99
+1.27%. Individual losses remain; these controls do not prove equivalence.

### Failed shared-cookie setup: retained, not a pass

The first baseline attempt omitted `--streams` when removing `--sessions`, leaving
only three subscriptions and no signed room-message streams. All **17,454** POSTs
were acknowledged, but room-message delivery was zero. The unchanged shared
validator rejected it before persistence/presence audits; the candidate never ran.
Raw records, settings, runner and error log are retained separately. Adding the
scraped stream arguments restored six subscriptions; the two valid pairs above
were rerun without changing assertions. This invalid offered population is not
counted as audited successful work or attributed to a production delivery defect.

### Sole longer paced follow-up: reject

Candidate-first, 64 distinct people, **1,000 paced messages** at the same 100 ms;
the saturated phase remains five seconds. The intent was to discriminate the
mixed thin paced tails, with no further sweep if a material regression remained.

| Population | Current | Candidate | Change |
|---|---:|---:|---:|
| Paced all-client p99 | 12.887 ms | 13.831 ms | +7.33% |
| Paced POST p99 | 12.015 ms | 12.631 ms | +5.13% |
| Saturated delivered rate | 1924.0/s | 1689.1/s | −12.21% |
| Saturated POST p99 | 8.447 ms | 9.135 ms | +8.14% |
| Saturated all-client p99 | 11.631 ms | 13.007 ms | +11.83% |

The follow-up did not resolve the non-regression gap in the candidate's favor.
No same-phase causal attribution explains these losses; host noise, changed
native copying, GC or scheduling are possibilities, not findings. This single VM
pair does not prove a universal regression or identify which of the two changes
caused one. It does block promoting **this combined candidate on this evidence**.
The candidate and all capture/follow-up budgets are closed. Reopening requires
changed evidence that accounts for the full-workload trade-off, not merely fewer
Go heap bytes or repeated copies of these controls. Native desktop availability
and historical rankings are not new global readiness gates.

## Correctness, retained evidence and limits

All **191,106** writes across the valid allocation, CPU and delivery controls
passed the unchanged shared Cable validator and independent exact ACK-ID/body/
creator/room/marker/FTS/population/integrity audits. Every acknowledged message
reached every client, presence returned to zero, integrity was `ok` and the seed's
one FK orphan was unchanged. The invalid setup's 17,454 writes are separate.

Private candidate `bin/check`, Linux affected vet/database/application/Cable races
and WebSocket-fork races passed. Expanded borrowed/owned prepared-frame tests
retain mutation protection and exercise mixed compressed/plain concurrent readers
with interleaved ordinary messages. Escaped/NUL token keys, duplicates and empty
batches pass against original and byte-bound source; a separate temporary probe
also checks large IDs and already-cancelled-context/next-read behavior. Independent
read-only review found no concrete ownership/authorization defect; that is not a
whole-program proof. No shared browser/media rerun followed rejection, and no
permanent test, constructor, binding change or diagnostic hook lands in production.

[manifest.json](manifest.json) hashes [artifacts](artifacts/): raw allocation/CPU
profiles, listings, exact client ACK/timestamp records, audited summaries, capture
metadata/scripts, rejected source/test patch, private hook, SQL probe and gate
logs. Binary hashes and image IDs are in metadata; binaries, fixture databases
and session files are not committed. Existing private fixture/helper dependencies
are the same as the [earlier CPU record](../authorization-materialization-20261011/README.md).
Go allocation samples do not include native malloc work, and CPU CGO frames do
not split interpreter/allocator/synchronization/boundary costs. Profiles can be
reused to choose a different supported owner; none explains the historical
low-rate phenomenon or establishes an application-wide performance improvement.
