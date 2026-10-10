# Pacing and connection-owned deadlines — 2026-10-10

The POST-tail deficit remains. Reusing Cable write deadlines removes measured allocation churn, but these results do **not** establish the cause or resolution of the earlier 593.4-delivered/s collapse. A later cohort reproduced a low 800.2-delivered/s run on the previous application; upstream and Rust also had unfavorable runs. All such samples are retained.

Landed source: `90e4295` adds explicit active pacing, `75116aa` gives each Cable writer one reusable 30-second deadline, and `e26bdf9` avoids attachment-free plain-text cloning and literal-fragment parsing. `production.patch.gz` contains the exact combined changes from `7aa66f2`; `provenance.json` identifies builds. No authorization cache, larger publication queue, GC-policy change or heap-budget increase was used. The legacy timer mode remains the default.

## Pacing

Go 1.27's Linux epoll path rounds positive sub-millisecond waits to milliseconds. At 3,000 requests/s, precise configured deadlines therefore did not mean evenly spaced socket writes. Relative nanosleep and spin-only probes failed to correct clustering. Spinning and yielding after successful enqueueing reduced sub-100µs write gaps to about 1–3%, with median gaps near 333µs in the recorded ingress probes. These are socket-write-completion callbacks, not packet-arrival timestamps.

The supported `--pacing active` mode preserves absolute deadlines, bounded workers/queue, drops, cancellation and accounting. It consumes roughly one generator CPU; native runs used about 109–110% aggregate generator CPU. Allocate client CPUs separately and identify the mode in comparisons. Correcting the generator is not an application win, and the older timer workloads remain evidence.

## Uninstrumented comparisons

Three sequential, alternating orders; shared seed, complete encoded responses, real authentication, GOGC=100, four server/client CPUs, three job workers. HTTP used active pacing at 3,000/s, concurrency bound 64, two-second warmup and eight-second measurement. Every scheduled request succeeded; exact persisted body populations, unique client IDs, creator/room, normalized FTS and database integrity passed independent audits. HTTP timing does not capture acknowledgment IDs or validate each response's semantic content; shared route contracts run before timing.

### Fixed-rate POST p99, milliseconds

| Application / source | Pair 1 | Pair 2 | Pair 3 |
|---|---:|---:|---:|
| ARM64 VM upstream | 4.960 | 4.104 | 4.128 |
| ARM64 VM current, before richtext changes | 5.456 | 4.816 | 4.776 |
| ARM64 VM Rust | 2.936 | 2.724 | 2.944 |
| Native AMD64 upstream | 0.883 | 0.808 | 0.820 |
| Native AMD64 current, complete landed code | 0.946 | 0.958 | 0.931 |

Current lost all three upstream pairs on both machines. Native medians are 0.946 versus 0.820 ms, a **15.4% deficit**. Native mean latencies were much closer, approximately 0.46–0.48 ms. Native results contain 144,000 timed and 180,000 audited benchmark writes; the VM three-application cohort contains 216,000 timed and 270,000 audited writes. Do not compare absolute rates or latencies across architectures as a controlled application comparison.

### Cable delivered messages/s and paced all-client p99

64 independently authenticated fixture people; 100 paced posts at 100ms intervals, then five seconds with four saturated posters. Shared delivery checks, exact unique persisted markers/FTS, integrity and zero persisted presence after disconnect passed.

| Machine / application | Delivered/s, all three samples | Paced p99 ms, all three samples |
|---|---|---|
| ARM64 VM upstream | 833.6 / 1655.3 / 1676.9 | 10.255 / 13.119 / 17.615 |
| ARM64 VM previous `5ca5d3d` | 800.2 / 2628.7 / 2579.0 | 8.903 / 10.671 / 11.959 |
| ARM64 VM current | 2347.8 / 2598.7 / 2377.7 | 13.119 / 9.983 / 8.011 |
| ARM64 VM Rust | 1444.8 / 1229.9 / 2377.7 | 8.663 / 8.943 / 7.851 |
| Native AMD64 upstream | 2220.8 / 2379.7 / 2390.6 | 4.447 / 5.547 / 5.903 |
| Native AMD64 current | 3368.4 / 3396.0 / 3365.2 | 5.315 / 5.507 / 6.499 |

The VM current median is **7.8% below** the previous application's median, despite avoiding its first low run. Native current throughput wins all three upstream pairs, while paced latency loses two. There is no native Rust application comparison. VM and native cohorts audited 114,015 and 86,235 writes respectively.

The VM is OrbStack ARM64 with server CPUs 0–3 and client 4–7. Native Linux is Fedora 44, kernel 7.2.9, i9-13900KF. Native binaries use Go **1.27.1**, CGO, sqlite_fts5, trimpath and libvips 8.16.1 in a Debian toolchain container, not the host's Go 1.27.2. Rootless Podman uses host networking; server `taskset` pins 0,2,4,6 and client 8,10,12,14 to different physical P cores. Neither arrangement reserves CPUs or excludes background work/SMT sibling interference. Rootless cpuset delegation and SELinux initially blocked setup; per-process affinity and a container-local label-disable option resolved them without changing host configuration. Desktop SSH later timed out, preventing the proposed native observer control.

## Diagnostic evidence and rejected controls

These runs are instrumented, not capacity claims. Private patches and bounded profiles/samples are retained; full traces, runtime databases, credentials and large request dumps remain private.

- Saturated-phase Cable profiles exclude login and paced sleeps. Four baseline runs delivered 2667.0–2700.8/s without reproducing the collapse. Reader waits were negligible and heartbeat queries short in those good runs; that does not characterize the bad runs.
- Reusing the deadline reduced sampled phase allocation profiles from about **3728 MB to 2794 MB**, and sampled GC cycles from 96–100 to 74. Timeout contexts, cancellation registration and timers were substantial baseline costs. Fresh session authorization and bounded queues remain unchanged. The diagnostic binary used an inline version of the same watchdog, not the final extracted writer type.
- A private connection-owned cancellation bridge additionally removes WebSocket's per-frame `context.AfterFunc`. Sampled allocation per estimated message fell from about **208 KB to 188 KB**, and GC cycles to 62–63, but delivered/s was 2784.0/2769.8 versus the earlier watchdog's 2877.0/2845.2. Those unmatched instrumented samples do not justify a throughput or collapse claim. The bridge is **not landed**; interrupted socket I/O and lock waits require its explicit close bridge, not merely a noncancelable context.
- Native whole-run CPU profiles vary substantially between pairs: compression and cgo attribution do not establish a stable cause. GC logs are incomplete in some rootless runtime captures; missing lines must not be treated as absent cycles.
- An upstream fragment-retention control compared its default 32MB admission with zero admission. Default scheduled p99 was 4.320/5.592 ms versus 7.760/73.472 ms without retention; a default warmup outlier of 19.072 ms is retained too. This does not justify retaining unused fragments, increasing heap budgets or changing GOGC.
- Removing general presentation-wrapper reparsing failed the 658-case richtext oracle and was reverted. A later, oracle-safe single-text wrapper shortcut had native p99 0.953/0.973/0.946 versus 0.929/0.972/0.974 ms before it; it was also reverted because benefit was unclear. These experiments are not part of the landed source.
- Message creation still performs an initial generation-observer round trip despite using unscoped receipt presentation. A private bypass preserved read observation and authentication gates. VM p99 with observation was 5.000/5.208/5.384 versus bypass 5.712/4.896/5.440 ms. All 180,000 writes passed audits; bypass lost two pairs and its median. This is redundant work, but not an established explanation or fix. No bypass or observation-ownership change is landed.

## Verification

`bin/check` passed before all code commits. Native and Linux full races/vet, vendored WebSocket races, Cable/client repeated races, the 658-case richtext oracle and six affected shared Go/Rust workflows passed. New real-socket coverage checks idle deadline disarming, blocked-write expiration and terminal cancellation after expiration. Parser-tree tests compare the literal shortcut directly with the HTML parser.

The **pristine shared browser smoke failed** its immediate exact-three-message assertion on both current and pre-change native builds. The flow waits for the other tab's Cable delivery but can inspect the sender before its own asynchronous HTTP rendering finishes. A private observation-only copy waited for the three expected messages on each tab before retaining the original exact-count and XSS assertions; the complete flow then passed. This is a harness timing limitation, not a passing pristine gate. Raw debug DOM logs are excluded because they contain signed fixture tokens.

The requested POST-tail and Cable-collapse outcomes remain unfinished. These records support the specific pacing correction and allocation reductions, not universal parity, external-provider acceptance, forced-exit durability or readiness to open a PR.
