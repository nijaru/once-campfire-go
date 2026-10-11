# Cable scheduler and contiguous-frame investigation — 2026-10-10

The Cable collapse remains unexplained. A pristine-generator, uninstrumented baseline run reproduced **814.0 delivered messages/s** on production source `811ae04`. The external scheduler and CPU-profile cohorts did not capture that severe slowdown. A contiguous prepared-frame prototype reduced underlying transport writes in focused tests, but its matched application comparison regressed: **2528.0 versus 2547.15 delivered/s** by median, and **13.167 versus 10.667 ms** paced all-client p99. It is **not landed**.

Production was unchanged throughout this investigation. This is a current-versus-private-prototype record, not a new upstream Go/Rust comparison, a POST-tail result, or evidence that the collapse is fixed.

## Workload and controls

OrbStack Linux ARM64, Go 1.27.1, CGO, sqlite_fts5, trimpath, libvips 8.16.1, GOGC=100 and three job workers. The same SQLite/storage seed and pristine shared-verification load generator were used throughout. Each run authenticated 64 distinct fixture people, established their normal subscriptions, posted 100 messages at 100 ms intervals, then used four closed-loop posters for five seconds. Every run passed the shared delivery check and independent persisted-count, exact unique-marker, FTS, integrity and zero-presence-after-disconnect audits: **341,719 audited writes across 28 application runs**.

Server CPUs were 0–3 and client CPUs 4–7. The Ruby driver and external observer ran on 8–9. These affinities do not reserve physical host CPUs or eliminate background activity. The driver container permitted CPUs 0–9 so its client subprocess could switch to 4–7; initially restricting the container itself to 8–9 caused a pre-workload affinity setup failure, retained in `capture-setup-affinity.log`.

No task-owned build, test, agent or profiler ran alongside the uninstrumented comparisons. Diagnostic hooks remained private. Exact patches, binary/source hashes, adapters, bounded profiles and all samples are retained. Runtime databases, sessions, environment secrets and binaries are excluded. Individual JSON samples omit credential-bearing fields; duplicate raw JSON is not copied.

## External phase measurements

`capture/` contains eight production/pristine-generator attempts with external `/proc` observation every 200 ms. The declared stop rule was eight attempts maximum, stopping after the first rate below 1000/s. None crossed that threshold:

| Attempt | Delivered/s |
|---|---:|
| 1 | 2641.8 |
| 2 | 2744.1 |
| 3 | 2685.3 |
| 4 | 2562.3 |
| 5 | 2527.6 |
| 6 | 2396.3 |
| 7 | 2560.5 |
| 8 | 1655.9 |

The slower eighth attempt used about **2.27 server cores and 1.41 client cores** during posting, versus 2.32–2.46 and 1.27–1.33 in the preceding attempts. Server cgroup CPU quotas were unlimited, with no recorded throttling; kernel CPU counters reported no steal time. Aggregate runnable wait on client threads present at both boundaries was **2.93 seconds**, versus 5.37–5.71 seconds previously. These caller/thread sums are not wall-clock stall durations. New server threads are listed separately, and disappearing threads are not converted into negative deltas.

Those observations do not explain the 1656/s sample by quota throttling or increased aggregate client runnable wait. They do **not** clear scheduling, host interference or the generator generally: the severe 814/614/s behavior was not captured here. Boundary snapshots were within about 1 ms of phase markers. Median snapshot collection cost was 0.82–1.53 ms, with a maximum of 10.20 ms. The scheduler enablement control was unavailable, while per-thread `schedstat` counters were exposed and nonzero. `*.proc.json.gz` retains the bounded snapshots; `phase-analysis.json` and `analyze.py` retain the calculation.

## Phase profiles and queued-burst probe

A second, explicitly instrumented cohort used current source plus only private signal-driven CPU profiling and before/after memory counters. It declared six attempts maximum and a stop threshold below 1800/s, to investigate the moderate slowdown rather than claim full-collapse capture. Rates were **2541.3 / 2224.6 / 2468.6 / 2549.1 / 2646.7 / 2604.5**; no attempt crossed the threshold. Allocation was about **203.6 KB per saturated post**, with 57–68 GC cycles and 40–54 ms aggregate stop-the-world pause time per phase.

The profiles attribute roughly four CPU seconds per five-second phase to socket-write paths, making transport work a concrete optimization hypothesis. They still describe healthier runs, not the severe collapse. CPU profiling and memory snapshots perturb execution; profile intervals start slightly after the generator's phase marker.

Two further instrumented runs counted already-queued frames at each dequeue, merging writer-local counters only during teardown. About **61–63%** of dequeues had another frame waiting, but approximately half the payloads exceeded 4096 bytes. This challenged simple deferred-flush batching: large payloads force buffered writes anyway, while draining more frames would add in-flight headroom and affect control-frame/deadline scheduling. No batching API or larger queue was implemented. `queue-capture/` retains the distributions and both audited runs.

## Rejected contiguous prepared-frame prototype

The private prototype owns one immutable header-plus-payload allocation per prepared representation. Plain and compressed wire forms preserve the existing client/context-takeover fallback, masking, frame boundaries, cancellation hooks, locks, final flush and close handling. It changes no publication authorization, queue or timeout policy.

A real-socket test decoded binary messages at payload sizes 0, 125, 126, 4095, 4096, 65535 and 65536. Baseline failed at 4095 bytes with two underlying `net.Conn.Write` calls; the prototype used one at every tested size. These are transport method calls, not packet or syscall counts. Additional focused checks cover masked-client and context-takeover fallback, interleaved ordinary writes, partial transport failures and no replay after an error. Native and Linux fork races and existing Cable deadline/shutdown races passed; independent source review found no concrete correctness defect. Those checks do not establish universal protocol correctness or a performance win.

The first three matched pairs improved saturated throughput, but paced p99 worsened in all three. A second, reverse-order three-pair cohort was declared to resolve that latency concern. Both cohorts and every unfavorable sample are retained:

| Pair | Order | Current delivered/s | Prototype delivered/s | Current paced p99 ms | Prototype paced p99 ms |
|---|---|---:|---:|---:|---:|
| 1 | current → prototype | 2634.0 | 2703.3 | 11.895 | 15.311 |
| 2 | prototype → current | 1939.2 | 2443.7 | 10.647 | 13.183 |
| 3 | current → prototype | **814.0** | 2426.5 | 5.979 | 6.019 |
| 4 | prototype → current | 2460.3 | **1487.9** | 10.495 | 13.159 |
| 5 | current → prototype | 2658.0 | 2774.2 | 12.727 | 13.175 |
| 6 | prototype → current | 2646.4 | 2612.3 | 10.687 | 11.887 |

Across all six pairs the prototype median throughput was **0.75% worse**, while paced p99 was **23.4% worse** and lost every pair. The 814/s baseline and 1488/s prototype samples remain valid audited observations, not discarded outliers. Reducing isolated writes did not produce a useful application improvement, so the prototype and its tests remain private evidence rather than production changes.

Desktop SSH still timed out, so there is no new native comparison. The existing POST-tail disadvantage and intermittent Cable collapse remain open. See the [earlier pacing/deadline investigation](../pacing-deadlines-20261010/README.md) for prior comparisons and rejected controls; this record does not relabel them as results for `811ae04`.
