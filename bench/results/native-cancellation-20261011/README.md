# Native cancellation and phase-boundary controls — 2026-10-11

The connection-owned cancellation bridge was **not promoted**. After its cancellation races were corrected, it improved native Cable throughput by 4.0%, but VM paced p99 was 51.6% worse at the median. Production remains unchanged. Neither the intermittent Cable collapse nor the earlier POST-tail deficit is resolved.

All 36 application runs passed the shared Cable checks and persisted-write audits: **528,178 writes**, exact unique markers, FTS inclusion, SQLite integrity and zero fixture presence after disconnect. These checks cover the exercised workload, not universal correctness.

## Sources and controls

The production application is `811ae04`, from checkout `66895c8`. Upstream Go is `c4ab53d`; the generator is the unchanged shared verification load generator. Binary, seed, image and source identities are retained in cohort metadata and `provenance.json`. No Rust application or HTTP fixed-rate POST comparison was run in this record.

The desktop was available again. A fresh Linux amd64 build used Go 1.27.1, CGO, `sqlite_fts5`, `-trimpath`, `-buildvcs=false` and libvips 8.16.1 inside an isolated rootless container. The build base was the accepted pre-wrapper image, not the latest image containing the rejected literal-wrapper experiment. Cable/cmd vet and races and the WebSocket fork races passed before measurements. No host dependencies or configuration were changed.

Native server CPUs were `0,2,4,6`; generator CPUs were `8,10,12,14`; the Ruby driver used `16,18`. These are separate physical cores on the i9-13900KF, but SMT siblings and background activity were not excluded. VM server/client/driver affinity was `0-3`/`4-7`/`8-9`, not physical isolation. `GOGC=100`, four server workers, 64 separately authenticated fixture people, 100 paced messages at 100 ms, and five saturated seconds with four closed-loop posters were retained. Authorization, queue limits and complete response/frame consumption were unchanged.

## Phase-boundary observations

`native-cable-boundary/` and `vm-boundary-capture/` read external process CPU, per-thread scheduler counters, cgroup counters and memory only at generator phase markers. There was no periodic sampler and no application or generator profiling hook. These are instrumented diagnostics, not clean timing comparisons.

| Native pair | Upstream delivered/s | Current delivered/s | Upstream paced p99 ms | Current paced p99 ms |
| --- | ---: | ---: | ---: | ---: |
| 1 | 2385.9 | 3359.6 | 6.071 | 4.955 |
| 2 | 2380.8 | 3359.8 | 5.963 | 5.667 |
| 3 | 2380.7 | 3417.4 | 5.663 | 6.059 |

Current used approximately 2.79–2.84 process CPU cores during saturated posting, versus upstream's 3.38; the native generator used about 2.09–2.14 versus 1.52. This is aggregate phase CPU, including background/control work—not a per-request causal attribution. No native collapse occurred.

The VM diagnostic had a declared maximum of 12 attempts and would stop after the first delivered rate below 1000/s. All 12 completed above that threshold: **2663.0, 2666.0, 2582.5, 2621.4, 2653.4, 2645.0, 2751.8, 2663.1, 2624.4, 2569.8, 2649.1, 2636.0**. Thus it did not characterize the earlier 614/814 collapses. Less frequent observation is not evidence that observers caused or eliminated them.

The maximum boundary timestamp error was 8.37 ms native and 9.83 ms VM; maximum snapshot duration was 3.29 ms and 14.89 ms. One VM client process snapshot was unavailable; its phase CPU is explicitly missing, not zero. Scheduler waits sum common threads, omit new/disappearing thread contributions and are not wall-clock stalls. Client cgroup activity cannot be attributed solely to the generator. `phase-analysis.json` retains these limits and counters; bounded snapshots are stored as `.proc.json.gz`.

## Private cancellation bridge

The prototype detached per-frame I/O contexts and installed one connection-lifetime cancellation callback. The existing reusable write deadline remained. This had previously reduced allocations, but had lacked a matched comparison.

The initial native comparison in `native-bridge-comparison/` favored the prototype in all three throughput and paced-p99 pairs. It was nevertheless rejected for correctness. Independent review found that `Conn.CloseNow` merely waits when graceful `Conn.Close` has already claimed closure. Without the existing per-write cancellation hook, a backpressured data writer could then survive parent cancellation until graceful close's five-second timeout.

The private correction makes immediate close interrupt the transport before joining, while preserving physical-close errors. A real blocked pipe reproduced the graceful-close overlap; injected transport-close failure reproduced a discarded error. A real TCP test gates cleanup after physical closure and proves that the writer joins its callback; removing that join fails the test. The original failures and passing native/Linux Cable and fork race gates are retained. Source-only reviews complemented those checks; neither establishes universal correctness or hard wall-clock cleanup bounds.

The corrected candidate was measured separately; unsafe-candidate measurements were not relabeled. `unsafe-bridge.patch.gz` and `corrected-bridge.patch.gz` retain exact private source changes, including tests. Neither patch was applied to production. The fork changes were needed for the candidate's lifetime hook and were not promoted independently.

## Corrected candidate: matched, uninstrumented pairs

Orders were current→bridge, bridge→current, current→bridge. No task-owned builds, tests, agents or profilers ran during timing. There were no process-counter observations or profiling hooks in these cohorts. All adverse samples are retained.

| Environment / pair | Current delivered/s | Bridge delivered/s | Current paced p99 ms | Bridge paced p99 ms |
| --- | ---: | ---: | ---: | ---: |
| Native 1 | 3330.4 | 3452.4 | 6.507 | 6.275 |
| Native 2 | 3321.0 | 3463.4 | 6.027 | 5.399 |
| Native 3 | 3355.0 | 3526.7 | 5.439 | 5.983 |
| VM 1 | 2583.9 | 2617.0 | 12.063 | 10.367 |
| VM 2 | 2191.9 | 2430.5 | 7.583 | 15.615 |
| VM 3 | 2518.5 | 2564.3 | 9.143 | 13.863 |

Native median throughput rose **3330.4→3463.4 (+4.0%)**; paced p99 was nearly unchanged at **6.027→5.983 ms**, losing one pair. VM median throughput rose **2518.5→2564.3 (+1.8%)**, but paced p99 rose **9.143→13.863 ms (+51.6%)**, losing two pairs. These are medians of sample statistics, not percentiles of a pooled population. The latency trade-off does not justify promotion; its cause remains unresolved. Neither candidate nor baseline collapsed severely in this cohort.

Adapters, build/gate logs, regression failures, exact patches and checksums are retained. Compressed logs preserve their original bytes, including harmless archive-extraction warnings. Binaries, runtime databases, environment values, session files and duplicate raw responses remain private. Earlier adverse runs and rejected controls remain in the [scheduler follow-up](../cable-scheduler-20261010/README.md) and [pacing/deadline record](../pacing-deadlines-20261010/README.md).
