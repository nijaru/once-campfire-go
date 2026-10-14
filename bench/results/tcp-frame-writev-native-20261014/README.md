# Single-frame TCP gather-write: Linux control rejected

No production change. The prototype increases median delivery 15.6%, but fails the paced all-client p99 nonregression gate. The three-pair native budget is closed; conditional fixed POST/capacity controls and integration were not run.

## Finite matched control

Physical Fedora desktop, Linux 7.2.10-200.fc44.x86_64, i9-13900KF/32 logical CPUs, Go 1.27.1, `CGO_ENABLED=1`, `sqlite_fts5`, default GC, four workers/readers and three background jobs. Server affinity was 0,2,4,6; maintained client 8,10,12,14; runner 16,18. This is not exclusive-core or thermal isolation. The kernel changed since the historical Linux competitor cohort; neither those numbers nor Mac comparisons enter this control.

Exact production `8351224` and the final private prototype used identical assets, seed, runtime image and client. Baseline binary SHA256 `8a1409c09aa79e188010fb04dfa0c63b66e4f2748de734a117f6ae5db49b0e29` was asserted before posting; prototype `2cf28cbb8aec696ef58be2caa2eb592ac538e924e5204a78bfd654c9d53e9124`. Source, mechanism and terminal-error regression are retained in the [Mac before/after record](../tcp-frame-writev-local-20261014/README.md). No batching, scheduler, cache, GC, buffer/queue enlargement or socket-policy change.

Exactly six runs, 2026-10-14 18:13:46–18:18:52 UTC, in orders current/candidate, candidate/current, current/candidate. Each had 64 distinct authenticated people with six subscriptions, 300 paced messages at 100ms and five saturated seconds with four posters. Authentication ran on a separate setup server; the measured server restarted and waited three seconds after readiness. No task-owned builds, tests, agents, profilers or process samplers overlapped capture. Maintained receipt/delivery instrumentation is common to both arms, not assumed neutral.

| Metric | Current rounds | Candidate rounds | Median change |
|---|---|---|---:|
| Delivered messages/s | 3065.4 / 2929.7 / 2908.5 | 3386.1 / 3390.5 / 3074.6 | +15.58% |
| Paced all-client p99, ms | 6.527 / 1.816 / 1.777 | 6.243 / 6.487 / 1.870 | +243.78% |
| Saturated POST p99, ms | 3.389 / 3.551 / 3.545 | 3.013 / 3.023 / 3.353 | −14.72% |

The preregistered benefit criterion was at least 8% median delivery gain **or** 10% paced-p99 reduction, with at most 5% regression in the other metric and saturated POST p99. Delivery passes; paced p99 fails. There is a sharp mid-cohort shift to lower paced latency, including both variants in the final pair. Consequently, the large median difference is not a confident code-effect estimate. It does not permit discarding valid runs, changing the gate or repeating the cohort. Even the median paired paced change is +5.23%, outside the declared limit.

All 95,603 writes passed exact receipt IDs, canonical markers/body, creator/room, persisted population, FTS, integrity and unchanged seed FK audits. All 64 people completed delivery without client errors; presence returned to zero and every measured server joined with exit 0. Native application Cable/application/web races and vet, plus the fork's race suite and amd64 vet, passed before timing. This does not make the inherited ARM64 fork-vet errors green, nor replace a full integration check.

The host Ruby shim initially failed to parse its updated mise configuration before any posting. Capture used the existing Ruby 4.0.7 binary and system Python 3.14.8 through a command-local PATH; no host configuration was changed. Task-owned fixture copies and disposable build-container zstd installation supplied required existing tests, without skips. Preserved archive metadata warnings do not indicate altered embeds; the baseline binary matches exactly.

## Decision and evidence

Reject promotion under the original budget. The [Mac gain](../tcp-frame-writev-local-20261014/README.md) remains platform-specific, and the [separate Mac competitor ranking](../tcp-frame-writev-headtohead-mac-20261014/README.md) remains a different cohort. These Linux pairs establish neither current Go/Rust rankings nor the cause or repair of the historical competitor deficits. Production remains exact `8351224`; optimization goals remain unfinished.

`manifest.json` retains sanitized receipt populations, numeric ACK identities, summaries, criteria, all adverse samples, finite conclusion, native build/check log, source/binary/client identities and host evidence with raw/compressed hashes. Per-request timing traces are omitted from public receipt files; their private source hashes remain. Those traces, runtime logs, sessions/cookies, databases, binaries and source archives stay private.
