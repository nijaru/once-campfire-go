# Single-frame TCP gather writes — local Mac comparison, 2026-10-14

The private prototype passed the declared local Cable and CPU-efficiency gates. Across three alternating pairs, median delivered throughput rose **18.4%**, from **2336.7 to 2766.3 messages/s**, while joined server CPU per audited write fell **16.1%**. All **78,377 writes** passed the existing delivery, receipt, persistence, FTS, integrity, unchanged-seed-FK and presence-cleanup checks.

**Production is unchanged from `8351224`.** This is a physical Mac comparison, not a Linux result, an upstream Go/Rust ranking, or permission to promote the prototype. Desktop SSH timed out before this phase; no Linux build or performance comparison ran.

## Mechanism and scope

A large complete unmasked server data frame normally fills the 4 KiB buffered writer with its header and payload prefix, then writes the remaining payload separately. The prototype encodes the same header and uses `net.Buffers.WriteTo` to gather that header and the existing payload on a direct `*net.TCPConn`, when buffered output is empty and the complete frame exceeds the current buffer capacity. It avoids that prefix copy and split emission without constructing another combined prepared representation.

The existing frame mutex, cancellation callback, per-frame deadline, frame order, compression and 256-frame queue remain in place. Controls, streaming/continuations, masked clients, TLS, wrapped transports and small frames retain buffered writes. Two segment descriptors and header scratch belong to the connection; message references are cleared after success or error. No larger I/O buffer, batch delay, queue, cache, packet setting or GC tuning was added.

This differs from the [rejected contiguous-frame control](../cable-scheduler-20261010/README.md): it does not allocate and copy a combined prepared header/payload representation. Go's TCP vector-write implementation grounds the proposed work reduction; this experiment did **not** count syscalls or packets. The complete comparison measures the candidate, not an isolated attribution to one helper.

The current Campfire listeners and standard HTTP hijacker bind the returned writer to the returned connection. An arbitrary custom hijacker with separate write interception could invalidate that assumption; no fork-wide custom-writer compatibility claim follows.

Multi-frame deferred-flush batching was rejected before implementation or timing: one batch deadline would change per-frame policy, and a batch-wide lock/drain would change control opportunities. This candidate sends one frame per operation and leaves full menu HTML intact.

## Matched local workload

Physical Mac15,9, 16 logical CPUs, 128 GiB RAM; macOS arm64; Go 1.27.1, CGO, sqlite_fts5, trimpath, buildvcs=false; libvips 8.18.7. Both binaries use the same production archive, embedded assets and dependencies, differing only in the retained fork patch. The unchanged trace-enabled shared client was built for the Mac with locked dependencies and Rust 1.99.0.

Each arm starts from the same `691c52…` SQLite/storage seed, authenticates 64 distinct people through real HTTP, establishes six subscriptions per person, then posts 300 paced messages at 100 ms and runs four saturated posters for five seconds. Setup/login uses a separate server; the measured server restarts before WebSocket setup, with a three-second readiness pause. Closed loopback push/webhook providers, GOMAXPROCS=4, four database readers, three jobs and default GOGC=100 are matched. Supported loopback forwarding supplies distinct synthetic fixture IPs during login; no host aliases or network configuration changed.

Docker was unavailable, so these are native Mac processes, not VM runs. macOS default scheduling supplies **no CPU affinity isolation**. Normal activity in other repositories and desktop services continued. Before/after snapshots retain aggregate observations and private-raw hashes, not an exclusive-host or background-neutrality claim. All task-owned builds, tests and reviews finished before timing; no profiler or process sampler overlapped it. The final source review ran afterward.

| Round / order | Current delivered/s | Prototype delivered/s | Current paced p99 ms | Prototype paced p99 ms | Current CPU µs/write | Prototype CPU µs/write |
|---|---:|---:|---:|---:|---:|---:|
| 1 / current → prototype | 2336.7 | 2762.7 | 9.847 | 8.415 | 1740.668 | 1458.183 |
| 2 / prototype → current | 2324.9 | 2770.5 | 8.583 | 6.807 | 1720.086 | 1459.049 |
| 3 / current → prototype | 2349.3 | 2766.3 | 7.907 | 7.903 | 1738.064 | 1473.742 |

| Median metric | Current | Prototype | Change |
|---|---:|---:|---:|
| Delivered messages/s | 2336.7 | 2766.3 | **+18.4%** |
| Paced all-client p99 ms | 8.583 | 7.903 | −7.9% |
| Saturated POST p99 ms | 3.831 | 3.421 | −10.7% |
| Joined CPU µs/audited write | 1738.064 | 1459.049 | **−16.1%** |

The declared benefit gate required ≥8% median delivery improvement **or** ≥10% paced p99 reduction, with ≤5% regression in the other metric and saturated POST p99. Delivery passes; the paced benefit threshold alone does not. All non-regression gates and the separate ≥5% CPU/write reduction gate pass. Each pair improves delivery and CPU/write; the third paced result is nearly flat. Three pairs demonstrate observed repeatability, not quantified noise or confidence bounds. No repeat, refinement or sweep followed.

CPU comes from `os.wait4` after joined shutdown, without server instrumentation. It includes startup, the readiness pause, WebSocket setup, both posting phases, cleanup and shutdown, but excludes the separate login server. Total CPU remains near 20.5–20.9 seconds while the prototype completes more writes. CPU/write therefore measures whole-workload efficiency at unequal achieved loads, not an isolated fixed-population component saving. Nothing here establishes why the current Linux/Rust ranking gap exists or how much this change would repair it.

## Correctness and remaining gates

Before timing, full local fork races and Cable/application/web races and vet passed. Real TCP tests check independent RFC6455 bytes at short/extended/buffer boundaries, standard DEFLATE decoding, interleaved ordinary writes, blocked-write cancellation and descriptor release. Existing masked/shared-compression and Cable deadline-lifecycle checks also passed.

A focused probe caught a defect in the **initial private prototype**, not production: after a TCP write failed, clearing its transport deadline allowed another frame because the direct path bypassed bufio's sticky error. The probe passes on original `8351224`, fails on the unfixed prototype and passes after the connection owns terminal frame-I/O error state under the existing mutex. The final read-only review found no blocker in that fix. It did not rerun tests or establish broader protocol correctness.

Full fork vet reports the same pre-existing ARM64 assembly argument-name errors on original and candidate; it is not claimed green. An initial web test run lacked reference fixtures; correcting the task-owned reference link allowed the full rerun to pass without skipped assertions. Retained logs distinguish those setup failures from product defects.

Linux controls, fresh production `bin/check`, applicable shared HTTP/browser verification and integration review remain outstanding. No production code, dependency, frontend, version or permanent test changed in this delivery.

## Evidence

`manifest.json` lists every compressed/raw SHA256, size and gzip artifact. It retains all six raw client samples and canonical numeric ACK traces, results, joined CPU, identities, preregistrations, source patch, scripts, checks and final review. No failed population was discarded. Runtime logs, full system rosters, binaries and source archives stay private; temporary databases/session files were removed after audit. Credentials, cookies and fixture environment are not published.
