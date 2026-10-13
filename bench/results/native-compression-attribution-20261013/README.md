# Native compression attribution — 2026-10-13

**The sampled gzip-owner difference persists, but does not establish extra
compression CPU.** Within the explicitly observed current-source calls, pprof
attributes 16.17 CPU seconds to `gzip.Writer.Close`, while those calls total only
5.6835 seconds of elapsed time. That estimate exceeds the synchronous work's
elapsed-time upper bound by 2.85×. Joined process CPU is essentially unchanged.
No compression change or further sweep follows.

This closes the agreed one-pair attribution experiment, not the fixed-load POST
latency disadvantage or the cause of the sampling discrepancy. Production remains
`8351224`; the rejected cache-observation guard is not used.

## Sources, workload and discriminator

- Current application `8351224b56c9a3c6784ed5d7510857212133ac4c`; upstream
  `c4ab53d00b0a19a55962e03d18057ca6c3da9437`, still public HEAD when checked before
  capture. Each is exported into an isolated tree; the desktop checkout is unused.
- Fedora i9-13900KF, native Linux/amd64, cached image
  `619c76d7a213cad0af1c4588af504714751295801728c63df1041b9c93608fe3`.
  Go 1.27.1, CGO, `sqlite_fts5`, `-trimpath`, `GOAMD64=v1`, no PGO;
  four server workers/readers, three job workers, stock gzip level 6 and default GC.
- Seed SHA `691c52bfdc30b38bd8867e93b0914613df20e03d4cd7b00796ab68e0a5d26c9a`;
  existing shared HTTP preflight precedes the profiled restart. Push/webhook URLs
  point to closed loopback endpoints. Each arm uses a fresh SQLite backup/storage
  copy, including the one existing FK orphan.
- Exactly one sequential current/upstream pair: 3,000 offered POSTs/s, 64 workers,
  bounded active queue 8,192; two warmup plus twenty timed seconds per arm.
  Profiles and counters include both phases plus startup/shutdown, not login or
  preflight. Server affinity is `0,2,4,6`, client `8,10,12,14`, runner `16,18`.
  Affinity is not exclusive core/SMT/thermal isolation.
- The same maintained receipt-audited client from `4a35851` is rebuilt for both
  arms: SHA `4f8848e2a52927660d6aa031fa38c16b7f7e7cd500cf1c8aee7e3f245b73b1f9`.
  Its validation/retention overhead is not assumed neutral or equated with older
  private-client cohorts.
- Predeclared discriminator: compare the same gzip Close/Write symbols in absolute
  sampled CPU/write, requiring at least one sampled second per owner/arm. A 10%
  difference is material attribution only, not a latency or promotion gate.
  Corroborate with call counts, wire sizes and joined process CPU. No extension.

The private overlay adds identical explicit `observeGzipClose(*gzip.Writer)`
wrappers to the two front receipt-compression paths, atomic call/elapsed totals
and one shutdown report. Writer types, pools, levels, errors, authorization,
queues, caches and resource lifecycles stay unchanged. These hooks do **not**
count every process-wide gzip Close: public coding and upstream cache paths also
contain unwrapped callers. The conclusion below concerns the observed stacks.

Private tests verify unchanged level-6 bytes, preserved Close errors and exact
counts. Both local front race suites and native focused hook/client tests pass;
both native binaries build serially. The supervisor forwards SIGTERM, joins the
server with `wait4`, and records its OS user/system CPU. Summed Close elapsed time
is never labeled CPU. Instrumentation may perturb scheduling.

## Complete outcomes and absolute work

All **132,000 writes** pass canonical ACK-ID/body/client/creator/room/FTS audits:
6,000 warmup and 60,000 timed writes per arm. Every response is complete and valid;
there are no dropped/unscheduled arrivals, invalid receipts or HTTP/network errors.
Each arm records **66,000 observed Close calls**, exits cleanly, passes SQLite
integrity, and preserves the seed FK population. All task-owned builds/tests and
the independent interpretation review run outside capture.

| Whole capture, 66,000 writes/arm | Current | Upstream |
|---|---:|---:|
| Joined process CPU | 33.914193 s | 34.165181 s |
| Process CPU/write | 513.85 µs | 517.65 µs |
| Total sampled CPU | 33.82 s | 34.12 s |
| Sampled `observeGzipClose` cumulative CPU | 16.18 s | 2.77 s |
| Sampled gzip Close CPU within observed stacks | 16.17 s | 2.76 s |
| Sampled gzip Close CPU/write | 245.00 µs | 41.82 µs |
| Summed observed Close elapsed | 5.683528271 s | 4.971836873 s |
| Observed Close elapsed/call | 86.11 µs | 75.33 µs |
| Timed mean wire bytes/response | 1,957 | 1,955 |

Process CPU differs −0.73%, while the sampled Close owner differs nearly sixfold.
Both Close owners exceed the sampling threshold. Write has only 0.27/0.02 sampled
seconds, below the threshold; no comparative Write conclusion is supported.

A focus filter on `front.observeGzipClose` retains the same 16.17/2.76 gzip Close
seconds, excluding the unwrapped callers. The monotonic counter brackets each
synchronous call; summing durations remains an upper bound on those calls'
execution CPU even when calls overlap across goroutines. Current's sampled Close
estimate cannot represent actual 16.17-second execution inside 5.6835 summed
elapsed seconds. Near-agreement between total sampled and process CPU does not
validate the attribution among owners.

Go's matching installed/native source uses a [100 Hz CPU profile](https://github.com/golang/go/blob/go1.27.1/src/runtime/pprof/pprof.go)
and [Linux thread-CPU timers with randomized initial offsets](https://github.com/golang/go/blob/go1.27.1/src/runtime/os_linux.go).
The source hashes are retained. These implementation facts do not establish
aliasing, delayed signal delivery, a specific Go/kernel bug or a repair. The
elapsed difference likewise does not establish the component CPU difference.
Independent read-only review confirms this scoped interpretation and its limits.

Raw instrumented latencies remain visible: warmup p99 12.224/4.656 ms and timed
p99 0.895/0.818 ms, current/upstream. They are diagnostic, not a new ranking or a
substitute for the [three-pair uninstrumented comparison](../current-upstream-native-20261011/README.md).
Its fixed-load POST disadvantage remains disclosed.

## Decision and retained evidence

The old sampled-owner-based compression-optimization premise is closed: this
pair does not establish extra compression CPU worth changing production policy.
It does not prove equal component CPU, exclude future compression improvements,
or explain POST tails. A new candidate needs different representative work
attribution and declared correctness/benefit/non-regression gates. Do not revive
the rejected level-3 control or switch to streaming to reduce a profile share.
No additional rate, profiler or gzip-level sweep follows this finite experiment.

The manifest retains **39 gzip artifacts, 1,168,362 bytes**: raw profiles,
focused/cumulative reports, ACK populations and audit outcomes, source/seed/binary
identities, final private patches and source snapshots, build/test output,
runner/client/audit sources and toolchain source hashes. Cookies, credentials,
databases, storage, TLS fixture keys and runtime request logs are excluded.
There is no new production implementation change, latency/capacity improvement
claim, universal-parity claim or PR action.
