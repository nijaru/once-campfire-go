# Restore verification and POST tail investigation — 2026-10-10

The presentation cleanup at `5ca5d3d` removes unused author mention-context work and
renders created-message append streams without an intermediate HTML copy. It
reduces measured allocation by 16–17% against the preceding branch implementation.
The final uninstrumented POST capacity median is 6.2% above Go upstream, with all
three paired runs improving. Fixed-rate POST tails still trail upstream in two
pairs, and Rust retains the highest capacity median. These results do not establish
uniformly better latency or general application capacity.

## Final uninstrumented comparisons

All three implementations use matched seed data, four runtime threads/readers,
three job workers, gzip and sequential forward/reverse rounds. Both Go versions
use `GOGC=100`. Server affinity is `0-3`, generator affinity `4-7`; this is OrbStack
virtual-CPU affinity, **not physical CPU isolation**. Host activity and thermals
are uncontrolled. No task-owned builds, tests, agents or profilers ran alongside
these timing samples. The candidate binary was rebuilt from clean committed
`5ca5d3d`; `logs/final-build.log` retains its Go build metadata.

### POST capacity: 16 clients, 2-second warmup, 8-second timing

| Implementation | Round 1 req/s | Round 2 | Round 3 | Median |
|---|---:|---:|---:|---:|
| Go upstream | 2284.9 | 2538.0 | 2509.8 | 2509.8 |
| Go branch | 2666.0 | 2571.0 | 2675.8 | 2666.0 |
| Rust | 2898.0 | 2911.9 | 2892.4 | 2898.0 |

The branch's paired uplifts are 16.7%, 1.3% and 6.6%; **6.2% compares the capacity
medians**, not the uplift in every pair or the median of paired uplifts. Rust's median is 8.7% higher than the branch's.
There are 191,724 semantically validated timed responses and 238,801 audited
benchmark writes including warmup. Shared route checks validate complete timed
responses; independent audits verify acknowledged IDs, exact bodies, room,
creator, rich text, normalized FTS and database integrity.

### Fixed 3,000 POST/s: bound 64, 2-second warmup, 8-second timing

| Implementation | Scheduled p99 ms, rounds 1 / 2 / 3 | Median | Service p99 ms, rounds 1 / 2 / 3 |
|---|---|---:|---|
| Go upstream | 8.160 / 6.792 / 7.504 | 7.504 | 6.600 / 4.728 / 5.832 |
| Go branch | 7.888 / 8.496 / 8.832 | 8.496 | 6.064 / 7.016 / 7.264 |
| Rust | 5.136 / 5.456 / 5.128 | 5.136 | 2.708 / 3.356 / 2.716 |

The branch loses two upstream pairs. The scheduled p99 median is 13.2% higher.
`httprate` consumes full encoded bodies and checks status, nonempty output and
transport completion; it does **not** validate each timed acknowledgement body or
capture acknowledgement IDs. Exact postrun body/creator/room/client-identity/FTS
and integrity audits cover 270,000 writes, including 216,000 timed responses.
No offered arrivals were lost or failed. Raw queue, generator lateness and maximum
latency values remain in `fixed-post/results.json`; quantiles must not be subtracted
to estimate individual latency components.

### Cable: 64 distinct people, four posters

Each person owns a separate authenticated session and six subscriptions. The
shared harness sends 100 paced messages at 100 ms intervals, then four posters
run for five seconds. Subscription confirmations, complete delivery, exact marked
writes, FTS, integrity and presence returning to zero pass for 97,311 writes.

| Implementation | Delivered messages/s, rounds 1 / 2 / 3 | Median | Paced all-client p99 ms, rounds 1 / 2 / 3 |
|---|---|---:|---|
| Go upstream | 1747.7 / 1697.7 / 1562.4 | 1697.7 | 16.767 / 14.607 / 13.671 |
| Go branch | 2663.6 / 2717.3 / **593.4** | 2663.6 | 12.583 / 13.903 / 6.707 |
| Rust | 2926.6 / 2557.3 / 2808.9 | 2808.9 | 9.783 / 12.007 / 5.831 |

The branch wins two upstream throughput pairs and all three paced-p99 pairs.
Its median throughput uplift is 56.9%, but the third throughput collapse is a
valid, audited result whose cause was not isolated. It is retained, not discarded
as noise. Rust has the higher throughput median and lower paced p99 in every pair.
Phase memory observations are not global peak-RSS bounds.

The first Cable launch failed because the tools container could not see the
server's host PID under `/proc`. Adding `--pid host` fixed the measurement setup;
`logs/final-comparisons.log` retains that failure. No result from that incomplete
launch is counted.

## Diagnostic evidence, not capacity proof

Nine diagnostic phases use the existing fixed-rate client and unchanged state
audits. Each phase has two forward/reverse pairs at 3,000 POST/s, bound 64 and the
same affinity. Trace phases use one second of warmup and three seconds of timing;
allocation/CPU phases use one and ten seconds. They audit 868,000 benchmark writes
in total. Profiles and MemStats include startup, preflight, warmup, timing, audits
and shutdown; bytes/write are whole-lifetime totals divided by benchmark writes,
not exact per-operation allocation benchmarks.

- **Original allocation comparison:** branch 95,787 / 95,543 bytes/write versus
  upstream 149,137 / 148,782. The branch nevertheless collects more often:
  300 / 294 collections versus 135 / 133. Upstream retains substantially more
  message-fragment cache memory; admitting unscoped receipt fragments into that
  cache is not an acceptable substitute for reducing work.
- **Mention-only cleanup:** 89,086 / 88,890 bytes/write against original branch
  94,406 / 93,808. Profiled service p99 worsens in both pairs. No tail improvement
  is attributed to that change alone.
- **Combined cleanup:** 79,304 / 78,748 bytes/write against original branch
  95,145 / 93,861; allocations/write fall from about 1,115 to 1,034. Collections
  fall from 289 / 283 to 247 / 246. Profiled service p99 improves from
  7.600 / 7.400 ms to 6.736 / 7.072 ms. These are allocation-instrumented comparisons
  against the original branch, **not upstream** (`go-before` labels that control).
- **Original execution traces:** writer hold durations are similar, while the
  branch spends more time in `BeginTx`, waiting and runnable states. Request/STW
  overlap is higher. `writer.acquire` includes `BeginTx`, not only queue handoff;
  trace Running state and STW overlap are not CPU attribution.
- **Finite GC-off diagnostic:** the acquisition gap remains with no collections.
  GC is therefore not its sole cause. Production remains `GOGC=100`; neither a
  larger heap/GC budget nor GC-off is adopted.
- **Already-queued writer handoffs:** the second GC-off pair does not show slower
  branch handoff. Mean previous-hold-end to next-hold-begin is approximately
  96 µs for the branch versus 113 µs upstream. A new writer queue is not justified
  by those measurements.
- **One client CPU:** restricting the existing generator to one CPU does not
  eliminate the gap. Its unfavorable results remain in `diagnostics/fixed-client-one`.
- **Compression CPU/allocation:** original branch `flate.NewWriter` cumulative
  allocation is 134–146 MiB versus upstream 47–60 MiB; gzip Close consumes
  1.47–1.58 seconds versus 1.29–1.36. The branch holds borrowed plaintext through
  completed compression/emission and returns separately owned gzip bytes. The
  larger `responseBuffer.finish` CPU mainly reflects compression moving into that
  owner, not an independent sixfold regression. Workspace churn is worth noting,
  but these profiles do not establish it as the remaining tail cause. Returning
  a pooled output slice after releasing its buffer would violate ownership.

### Correlated client writes and handler starts

Private instrumentation adds an ID header and records the existing client's
`httptrace.WroteRequest` callback time. The server records the same ID and wall
clock at handler entry. All 9,000 timed requests per run correlate uniquely;
12,000 including warmup have complete task/state accounting. The two processes
share a Linux wall clock, while trace durations use trace timestamps.

| Timed run | Client write gaps <100 µs | Handler-start gaps <100 µs |
|---|---:|---:|
| Branch 1 | 43.7% | 43.9% |
| Branch 2 | 42.7% | 43.5% |
| Upstream 1 | 19.2% | 17.9% |
| Upstream 2 | 20.3% | 19.0% |

Bursting is already present at client writes; it is not introduced only by writer
acquisition. The existing scheduler catches up overdue arrivals without another
wait. This does **not** isolate why client pacing differs or exonerate the server:
loopback response feedback, scheduling and host activity remain coupled.
`WroteRequest` records socket-write callback completion, not packet arrival or
wire transit. A handler could start before the callback on another CPU.

A separate **identity-encoding diagnostic**, with gzip disabled only for its timed
POSTs, makes both versions' client writes bursty (41–45% of gaps below 100 µs).
The branch-minus-upstream **mean handler-duration difference** shrinks to
125 / 26 µs; this is not an inter-arrival gap or a quantile subtraction. Mean
inter-arrival spacing remains about 333 µs. Branch service p99 still loses both
pairs: 6.272 / 6.600 ms versus 5.056 / 6.200. Removing compression therefore does
not eliminate the disadvantage; different wire sizes prevent treating this as a
pure compression-workspace experiment. No gzip policy change is adopted.

Further writer, heap-budget or buffer-ownership changes would outrun this evidence.
The measured allocation cleanup is retained; residual POST-tail and Cable-outlier
causes are not claimed to be solved.

## Correctness and restore verification

For `5ca5d3d`, native `bin/check`, native build, affected package races, pristine
shared Chromium smoke, Linux vet/full race suite, vendored WebSocket races and six
Go/Rust HTTP/state workflows pass. The workflows are message lifecycle, direct
ping, pagination, sidebar, direct upload and attachment lifecycle. Exact contextual
HTML/template comparisons cover the direct append path, boosted and unsupported
layout fallbacks. The mention projection regression fails on original source.
The first Linux gate command used nonexistent `internal/ws`; `linux-gates.log`
records that setup error and `linux-finish.log` records the corrected fork tests.

The restore fix at `ae21881` passes native gates, focused races, Linux command/
database races and actual production non-root setup/live-backup/offline-restore/
restart for default paths, storage/environment override, database override,
combined overrides and `?`/`#` paths. The unescaped staging-URI regression was
independently reproduced and fixed. Regression cases cover invalid/empty snapshots,
cancellation, source immutability, destination preservation and sidecar cleanup.
The earlier archive's custom-path restore limitation is resolved by this fix;
restore remains an offline operation, not a concurrent/crash-durability guarantee.

Static source and evidence reviews found no additional actionable correctness or
parser issue. These reviews are not execution proofs. Live provider acceptance,
universal frontend/media parity and forced-exit continuation durability remain
outside the exercised contracts described in earlier archives.

## Files and provenance

`post-capacity/`, `fixed-post/` and `distinct-fanout/` retain all uninstrumented
samples and audits. `diagnostics/` retains every diagnostic pair, memory records,
CPU/allocation/scheduler profiles and derived summaries. Correlation CSVs are
compressed with deterministic gzip; they contain IDs and timings, not bodies or
credentials. General trace summaries include warmup; ingress summaries select
only timed requests. Two compact GC-off writer CSVs also retain the intervals
behind the already-queued handoff comparison. Full trace and full per-request JSON dumps remain private;
their trace sizes and hashes are listed in `provenance.json`.

`adapters/` and `instrumentation/` retain the private scripts, original client,
patches and Dockerfiles. Patches use deterministic gzip to preserve their exact
context whitespace; `provenance.json` records their uncompressed hashes. They preserve their original scratch-path assumptions:
reproduction requires the shared harness helpers, seed and a locally supplied
fixture environment at those paths. No environment file or credentials are
included. These are inspectable experiment inputs, not supported production APIs.
Use `analyze-trace.py` with Go 1.27's `go tool trace -d=parsed`; use
`summarize-ingress.py <phase-directory>` after parsing the ingress traces.

`provenance.json` distinguishes committed final builds, pre-commit equivalent
source and instrumented builds. `go-production` was built from `ae21881` plus the
presentation patch and is not relabeled as a clean `5ca5d3d` build. Runtime databases,
full traces, cookies, session files and private environment files are omitted.
