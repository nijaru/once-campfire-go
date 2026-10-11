# Correlated native POST traces — 2026-10-11

This diagnostic matched **36,000 complete client responses** to server tasks across two current/upstream pairs. All **48,000 timed and warmup writes** passed exact population, body-index, unique client-ID, creator/room and FTS audits; SQLite integrity passed. Production is unchanged. The traces narrow where time was spent but do not establish a general cause or fix for the POST-tail deficit. The Cable collapse was not exercised.

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

A separate stack-span parser attributes every timed request's Waiting state. Its summed spans exactly match the original task-state parser for all 36,000 requests. In the selected tails, current SQL `ExecContext` waits within message creation average **88.36/90.34 µs**, versus upstream **31.92/40.64 µs**. Current response-version observation averages **19.50/21.04 µs** in Waiting, versus upstream's **0.96 µs** in the second sample. Upstream's second-sample writer-connection acquisition wait averages **464.60 µs**.

These spans include operation execution, child completion and wakeup; they do not isolate SQLite CPU, disk I/O or scheduler delay. Source inspection shows both creation paths execute the same five write statements for this attachment-free workload, in different orders; current also selects writer-current receipt inputs and prepares search content outside the writer. Similar transaction-hold durations therefore do not support removing cancellation, authorization or snapshot contracts. The prior observer bypass and read-BEGIN controls are not reclassified as successful.

Current tail requests overlap STW pauses more often (53/46 of 90 versus upstream 37/32), but mean overlap is only **22.84/21.08 µs** versus **16.91/16.57 µs**. Mark assists also vary. These are observed overlaps, not proof of the stable tail deficit's cause; they do not justify GC tuning or artificial heap retention.

## Retained evidence

The record includes sanitized samples, client/correlated CSVs, completed-task intervals, wait-stack summaries, parsers, exact private patches, build/gate logs and hashes. The four raw traces (approximately 28–34 MiB each), runtime logs, credentials, signed paths, databases and binaries remain private. Full repeated waiting-stack dumps are omitted; per-request correlations and aggregate stack attribution are retained. Earlier unfavorable uninstrumented results remain in the [pacing/deadline record](../pacing-deadlines-20261010/README.md), and rejected Cable controls remain in the [native cancellation record](../native-cancellation-20261011/README.md).
