# TCP gather-write prototype versus upstream Go and Rust — Mac, 2026-10-14

The optimized fork leads upstream Go in this Mac cohort but still trails Rust in fixed-rate POST p99 and live delivery. **27 balanced runs audit 969,903 performance writes**, plus 300 functional-preflight writes. All receipts, persisted messages, FTS, delivery and shutdown checks pass. No source was promoted.

This answers a different question from the [three-pair before/after screen](../tcp-frame-writev-local-20261014/README.md). That screen established a local prototype benefit over the current fork; these runs compare the prototype directly with fresh upstream competitors. None of its earlier measurements enters the medians here.

## Results

Medians of three rounds per application:

| Metric | Optimized fork | Upstream Go | Rust |
|---|---:|---:|---:|
| Fixed POST p99, ms, 3000/s | **1.556** | 1.712 | **1.380** |
| POST capacity, requests/s | **7346.2** | 6977.8 | 5321.7 |
| Capacity POST p99, ms | 8.495 | 8.927 | 4.583 |
| Live delivered messages/s | **2714.5** | 1053.4 | **2996.0** |
| Paced all-client p99, ms | 5.959 | 6.499 | 5.791 |
| Saturated live POST p99, ms | 3.459 | 6.223 | 2.501 |

The fork's fixed POST p99 is 9.1% lower than Go and **12.8% higher than Rust**. Capacity is 5.3% higher than Go and 38.0% higher than Rust; capacity latency is measured at unequal achieved loads, not an equal-load latency win. Live delivery is 157.7% higher than Go but **9.4% lower than Rust**. Paced p99 is 8.3% lower than Go and 2.9% higher than Rust.

All rounds, including adverse latency outcomes:

| Application / round | Fixed p99 ms | Capacity/s | Capacity p99 ms | Delivery/s | Paced p99 ms |
|---|---:|---:|---:|---:|---:|
| Optimized / 1 | 1.640 | 7346.2 | 8.527 | 2725.4 | 8.503 |
| Optimized / 2 | 1.556 | 7456.1 | 8.423 | 2714.5 | 5.495 |
| Optimized / 3 | 1.556 | 7325.6 | 8.495 | 2706.9 | 5.959 |
| Go / 1 | 1.668 | 6962.2 | 8.927 | 1064.5 | 5.903 |
| Go / 2 | 1.740 | 7120.6 | 8.775 | 1053.4 | 6.879 |
| Go / 3 | 1.712 | 6977.8 | 8.959 | 1028.6 | 6.499 |
| Rust / 1 | 1.324 | 5247.2 | 4.835 | 3009.6 | 5.791 |
| Rust / 2 | 1.380 | 5387.6 | 4.407 | 2991.7 | 5.919 |
| Rust / 3 | 1.698 | 5321.7 | 4.583 | 2996.0 | 5.187 |

The gather-write path changes Cable frame emission, **not ordinary fixed POST work**. A different fixed POST ranking on this Mac does not show that the candidate repaired the earlier Linux deficit. Do not mix these values with the [Linux three-way baseline](../current-three-way-20261013/README.md), extrapolate across platforms or use three rounds as confidence bounds.

## Sources and method

Both upstream HEADs were freshly checked and unchanged: Go `c4ab53d00b0a19a55962e03d18057ca6c3da9437`, Rust `6dae2fd3c1a2dce61ca11beb731bcce2b42b67ad`. Go was archived from the exact Git object; Rust came from its exact commit tarball with GitHub tree/commit metadata. Each retains its pinned assets: Go's Rust reference is `64f8635`, and Rust's Rails reference is `90b3300`. Existing `reference/` checkouts were not edited or fetched.

The prototype is exact production `8351224` plus the single-frame fork patch retained in the earlier screen. Its Mac binary is unchanged (`1b66894b…`); upstream Go is `d081aded…`, Rust `b797463a…`. Full identities and archive hashes are retained. Go uses 1.27.1, CGO, FTS5, trimpath and buildvcs=false. Rust uses installed 1.99.0, locked dependencies and the unchanged upstream release profile (fat LTO, one codegen unit, default allocator). Both link the installed libvips 8.18.7; no media-processing workload is ranked.

Physical Mac15,9, 16 logical CPUs, 128 GiB, macOS 26.6.2/arm64. Processes use default macOS scheduling, **no affinity isolation**. Normal UI/other-repository activity continued; system snapshots are taken outside timing, not assumed neutral. No task-owned builds, tests, reviews, profilers or process samplers overlap the runs. Results are local physical-hardware evidence, not an exclusive-host or Linux comparison.

Each workload follows the same three sequential Latin orders: Go → optimized → Rust; Rust → Go → optimized; optimized → Rust → Go. Every arm gets a fresh SQLite/storage seed (`691c52…`), closed loopback push/webhook providers, four server workers/readers, three jobs and unchanged GC/allocator defaults. Supported loopback forwarding supplies synthetic fixture login IPs, without host aliases or network changes. Setup uses a separate server; the workload server restarts, waits for readiness and pauses three seconds.

- **Fixed POST:** maintained exact-receipt Mac client (`3e12379d…`), gzip, 3000/s, 64 workers, active pacing, queue 8192, two-second warmup and eight timed seconds. Every offered request completes: 270,000 writes.
- **Capacity POST:** unchanged trace-enabled shared client (`85cf2ed4…`, source `fc9cf488…`), gzip, two-second/four-client warmup and eight timed seconds/16 clients. Route contracts and exact ACK/persistence audits pass: 595,715 writes.
- **Cable:** the same shared client, 64 distinct sessions/people, six subscriptions each, 300 paced messages at 100 ms, then five seconds/four saturated posters. POSTs do not request gzip. All clients receive every created message; unread frames and final presence cleanup also pass: 104,188 writes.

Full-wire client timing boundaries are unchanged. Joined OS CPU comes from `wait4`, not instrumentation, and includes startup, idle, connection setup, warmup, both phases and shutdown, excluding the separate setup server. It is retained for inspection, not isolated component attribution; throughput populations differ.

## Verification, failures and remaining work

Before timing, each built application passes the existing shared HTTP preflight and 100 exact fixed-POST receipt/persistence/FTS writes, integrity and unchanged seed-FK checks. Maintained HTTP-client races pass. Earlier candidate fork/Cable/application/web races, deadline tests and the final terminal-error ownership review remain applicable because its source/binary did not change. The inherited ARM64 fork-vet assembly error remains disclosed in the earlier record.

Two setup failures are retained: an offline Rust build lacked `bcrypt`, so the same locked build downloaded its dependencies; a private script named `http.py` shadowed Python's standard `http.client` and failed before login/warmup/posting. Renaming it fixes the harness import. That attempt contributes **zero writes and zero performance runs**. There were no repeated valid captures, alternate source variants or budget extensions. A reporting-only missing-key error after capture did not change data or trigger another run.

**Production remains unchanged.** Desktop SSH still times out after this cohort; no host/network repair was attempted. Linux benefit/non-regression controls, fresh `bin/check`, applicable shared HTTP/browser flows and integration review are still required before landing. Rust's remaining fixed/live lead is not a completed optimization target.

`manifest.json` verifies compressed/raw SHA256, sizes and artifacts: all 27 runs, raw ACKs, CPU, preflight, source identities, preregistration, scripts, build/check/setup logs and sanitized host summaries. Runtime application logs, full process rosters, source archives, binaries, cookies, fixture environment and databases remain private. Temporary databases/session files were removed after audit.
