# Correlated native POST traces — 2026-10-11

This diagnostic matched **36,000 complete client responses** to server tasks across two current/upstream pairs. All **48,000 timed and warmup writes** passed exact population, body-index, unique client-ID, creator/room and FTS audits; SQLite integrity passed. Production is unchanged. A later follow-up attributes much of the apparent membership-update cost to FTS5 savepoint work, but moving FTS last fails the matched full-workload landing criterion. This is not a general cause or fix for the POST-tail deficit. The Cable collapse was not exercised.

## Workload and instrumentation

The application is `811ae04` from checkout `4e8d7b3`; upstream is `c4ab53d`. Both private builds use Go 1.27.1, CGO, `sqlite_fts5`, `-trimpath` and `-buildvcs=false` in the accepted native pre-wrapper container toolchain. Upstream generated assets were restored from that image's original upstream build. Binary identities and raw-trace hashes are in `provenance.json`; exact diagnostic patches and the Dockerfile are retained. No diagnostic hooks were applied to production.

The native server used CPUs `0,2,4,6`, client `8,10,12,14`, and Ruby driver `16,18`. Separate physical cores do not exclude SMT sibling or background activity. `GOGC=100`, four server workers, gzip, 3000 offered POSTs/s, bounded 64-worker/8192-slot admission, active pacing, one warmup second and three timed seconds were used. The order was upstream→current, then current→upstream. These shorter **instrumented diagnostics** are not replacements for earlier eight-second uninstrumented measurements. Active pacing consumes roughly one generator CPU.

Private server hooks create a task around the public POST handler and regions around writer acquisition/transaction hold. The task includes authorization, command, postcommit effects, presentation and compression, but not the HTTP server's final socket flush. A private generator adds a numeric request header and records existing arrival/start/finish timestamps and encoded body length into a preallocated array. It still consumes the complete wire body and preserves the original scheduler, cancellation and population checks. Headers, task/region tracing and the client array perturb execution. Native generator races and both server builds' affected-package vet passed before capture; `bin/check` passed before delivery.

## Complete-client p99

| Pair | Upstream latency p99 ms | Current latency p99 ms | Upstream service p99 ms | Current service p99 ms |
| --- | ---: | ---: | ---: | ---: |
| 1 | 0.810 | 0.828 | 0.803 | 0.826 |
| 2 | 0.804 | 0.875 | 0.797 | 0.873 |

Current loses both p99 pairs. Generator lateness p99 was 0.001 ms in each timed sample. This does not make the instrumentation transparent or prove that the client/network contributed no delay.

## Per-request correlation

Every timed client ID has exactly one completed server task and writer regions. No marginal percentiles were subtracted. For each request, signed intervals connect client start, server ingress, handler completion and complete client consumption. The intervals reconstruct that request's service duration; any negative boundaries are retained, not clamped. Before-ingress time includes request construction, connection acquisition, transmission and server scheduling. After-handler time is a residual interval, not a pure network measurement.

The following means select each sample's **slowest 90 client-service requests** (one percent), not the same requests across implementations. They describe tail populations, not an additive decomposition of p99. Upstream has the largest outliers despite its lower p99.

| Sample | Client service µs | Handler µs | Before ingress µs | After handler µs | Writer acquire µs | Writer hold µs | Handler Waiting µs |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Upstream 1 | 1180.20 | 893.66 | 219.06 | 67.48 | 194.73 | 139.30 | 189.93 |
| Current 1 | 1150.97 | 914.53 | 191.35 | 45.09 | 199.17 | 154.64 | 281.19 |
| Current 2 | 1211.21 | 988.30 | 178.87 | 44.03 | 180.14 | 152.69 | 291.43 |
| Upstream 2 | 1536.05 | 1318.20 | 169.20 | 48.64 | 539.94 | 167.26 | 515.14 |

A separate stack-span parser attributes every timed request's Waiting state. Its sums match at the reported precision. The follow-up found four small truncations at generation-boundary `Waiting→Waiting` snapshots (1–256 ns/request), corrected by ignoring unchanged-state transitions. Those snapshots are not wakes; this was a diagnostic-parser issue, not a production defect. In the selected tails, current SQL `ExecContext` waits within message creation average **88.36/90.34 µs**, versus upstream **31.92/40.64 µs**. Current response-version observation averages **19.50/21.04 µs** in Waiting, versus upstream's **0.96 µs** in the second sample. Upstream's second-sample writer-connection acquisition wait averages **464.60 µs**.

These spans include operation execution, child completion and wakeup; they do not isolate SQLite CPU, disk I/O or scheduler delay. Source inspection shows both creation paths execute the same five write statements for this attachment-free workload, in different orders; current also selects writer-current receipt inputs and prepares search content outside the writer. Similar transaction-hold durations therefore do not support removing cancellation, authorization or snapshot contracts. The prior observer bypass and read-BEGIN controls are not reclassified as successful.

Current tail requests overlap STW pauses more often (53/46 of 90 versus upstream 37/32), but mean overlap is only **22.84/21.08 µs** versus **16.91/16.57 µs**. Mark assists also vary. These are observed overlaps, not proof of the stable tail deficit's cause; they do not justify GC tuning or artificial heap retention.

## Follow-up: child execution and hidden FTS5 work

The retained events link each execution child to its request goroutine. The new
[`children.py`](followup/children.py.gz) splits child states and request-wait overlaps,
retains non-semaphore waits and uncovered intervals, and exactly matches an
independent interval-clipping calculation on the same parsed stream. No new hooks
or timing runs were needed for this first attribution.

For each sample's slowest 90 client-service requests, current creation waits
88.36/90.35 µs contain **79.40/81.06 µs** overlapping child `Syscall` state;
child runnable overlap is only **4.67/3.99 µs**. Source-site/ordinal attribution
locates **58.17/59.71 µs** of child C-call time in the membership UPDATE, versus
upstream **5.27/6.96 µs**. These elapsed C-call spans still include native execution
and possible descheduling—not SQL CPU alone.

A separate private Linux diagnostic wraps only that SQLite step with monotonic
and `CLOCK_THREAD_CPUTIME_ID` readings plus statement counters. Two forward/reverse
pairs repeat the existing one-second warmup/three-second 3,000/s workload; all
**48,000 additional message writes** pass exact population/body/creator/room/FTS
and integrity checks. Current source is `ccfa838`; upstream remains `c4ab53d`.
Both builds carry the same private clock/status patch and tracing overhead.

| Sample | Mean membership step wall µs | Mean thread CPU µs | Outer VM steps | Full scans | Changed rows |
|---|---:|---:|---:|---:|---:|
| Upstream 1 | 5.51 | 4.92 | 135 | 0 | 5 |
| Current 1 | 19.69 | 18.82 | 135 | 0 | 5 |
| Current 2 | 19.47 | 18.89 | 135 | 0 | 5 |
| Upstream 2 | 5.41 | 4.88 | 135 | 0 | 5 |

The mean CPU gap is not principally off-CPU waiting in these instrumented
populations. Equal outer statement counters do **not** mean equal internal work:
SQLite's multi-row UPDATE opens a statement savepoint; callbacks run on all active
virtual-table transactions. FTS5's `fts5SavepointMethod` flushes pending terms via
`sqlite3Fts5FlushToDisk`, including internal shadow-table writes and possible merges.
Those internal statements are outside the UPDATE's VM-step count. Current inserts
FTS before updating memberships; upstream does memberships first. Its single-row
room update does not take the same multi-write savepoint path. Independent review
confirmed this source chain and found no surviving writer cursor to explain it.

An isolated 1,000-operation probe holds seed and writer-current authorization
constant while changing write order. Current order averages **9.05 µs** membership
CPU versus **2.14 µs** upstream order; the actual production command averages
**10.05 µs**. Every operation still reports 135 VM steps, zero full scans and five
changes. This supports the ordering mechanism, not a whole-application benefit:
FTS still must be flushed at commit.

### FTS-last control: rejected

The candidate moves creation's FTS insertion after all ordinary-table writes,
retaining the five writes, writer-current inputs, canonical text and atomic
rollback. `bin/check`, native Linux affected-package races (three repeats), vet
and builds pass. An initial native gate lacked the unchanged reference image
fixture; the setup failure and corrected run are retained. No assertions changed.

Three uninstrumented pairs use the existing full-body fixed-rate POST workload:
3,000/s, 64 workers, bounded 8,192 queue, active pacing, gzip, two warmup seconds
and eight timed seconds. Baseline is `ccfa838`; candidate source is the retained
[`patch`](followup/fts-last-candidate.patch.gz). All **180,000 timed/warmup message
writes** pass exact population/body/identity/FTS and integrity audits.

| Pair | Baseline p99 ms | FTS-last p99 ms |
|---|---:|---:|
| 1 | 0.912 | 0.973 |
| 2 | 1.002 | 0.884 |
| 3 | 0.923 | 0.978 |

The candidate loses two pairs; median p99 is **0.923→0.973 ms (+5.4%)**. Reordering
is rejected and reverted. Capacity/Cable sweeps were not run for a candidate that
already failed its primary metric. Do not infer saved total SQL work or a general
scheduling fix from the isolated membership improvement. This closes SQL-order
promotion as a supported remedy under these inputs; it does not close POST tails.

## Retained evidence

The record includes sanitized samples, client/correlated CSVs, completed-task intervals, wait-stack summaries, parsers, exact private patches, build/gate logs and hashes. The [follow-up manifest](followup/manifest.json) adds sanitized CPU/request measurements, summaries, adapters, private probe/patch sources and gate logs. The original four raw traces (approximately 28–34 MiB each), additional CPU/order traces, repeated child spans, runtime logs, credentials, signed paths, databases and binaries remain private. Full repeated waiting-stack dumps are omitted; per-request correlations and aggregate stack attribution are retained. Earlier unfavorable uninstrumented results remain in the [pacing/deadline record](../pacing-deadlines-20261010/README.md), and rejected Cable controls remain in the [native cancellation record](../native-cancellation-20261011/README.md).
