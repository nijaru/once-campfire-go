# POST critical paths and a rejected encoder — 2026-10-14

**No production change.** The native trace pair completes 60,000 exact-audited
writes and correlates every timed response. Response preparation is substantial
ordinary work, but a faster standalone encoder does not reduce complete-application
CPU. The substitution is rejected before latency, capacity or Cable controls.
Fixed POST and Rust live-delivery optimization remain unfinished.

## Correlated current-source requests

Current production is `8351224`; upstream Go remains `c4ab53d`. These are clean
private archives, not the older `811ae04` trace builds. Native Fedora/i9 hardware,
Go1.27.1/CGO/FTS5, four server workers, three jobs, default GC, gzip and affinities
match the [current baseline](../current-three-way-20261013/README.md). The desktop
had rebooted after an access outage. Affinity does not exclude SMT, thermal or
background interference.

One current→upstream pair uses 3,000 offered POSTs/s, 64 workers, bounded 8,192-slot
admission, two warmup seconds and eight timed seconds per arm. Identical private
execution tracing adds request tasks, writer acquire/hold regions and functional
phase regions. The common maintained receipt client adds an identity header and
preallocated timestamps; complete wire consumption still precedes semantic
validation. Instrumentation is not overhead-neutral or latency ranking evidence.
No CPU/allocation profiler, perf or overlapping task builds/tests/agents run.

All **60,000 writes** pass exact ACK/body/client/creator/room/FTS, integrity,
unchanged-seed FK and clean-shutdown checks. Every **48,000 timed response** joins
one completed task and writer regions. State intervals reconstruct each task and
region exactly. Multiple same-named regions are retained, not overwritten.
Cross-process wall clocks correlate ingress with client start/finish; signed
residuals remain, including negative boundaries. Marginal p99s are not subtracted.

Each arm independently selects its slowest 240 client-service requests (1%).
These are **means of selected requests**, not an additive p99 decomposition:

| Selected-tail mean, µs | Current | Upstream Go |
|---|---:|---:|
| Complete client service | 1085.69 | 1119.68 |
| Handler task | 838.93 | 884.06 |
| Before ingress | 140.51 | 192.95 |
| After handler | 106.25 | 42.67 |
| Writer acquisition | 130.72 | 156.03 |
| Writer hold | 148.98 | 139.03 |
| Notification preparation/admission | 87.55 | 123.69 |
| Display database work | 64.25 | 98.45 |
| Handler Waiting state | 223.55 | 164.58 |
| Handler Runnable state | 111.82 | 148.72 |

Current `response.prepare` averages **178.27µs**, or **21.25%** of selected handler
time; `response.finish` includes it and averages191.84µs. Upstream encodes at a
different middleware boundary, so its11.09µs all-request `response.finish` cannot
be compared as compression cost. Trace Running state is not physical thread CPU;
Waiting includes child work and wakeup. Notification/display means are lower in
current, not evidence of an extra postcommit bottleneck. Writer waits are substantial
in both; this does not reopen the failed FTS-order, room-preflight or cache-guard
controls. Preparation is a justified work-reduction target, not an explanation of
the uninstrumented competitor deficit.

The private parser reuses the earlier trace format. Its first legacy-fixture check
incorrectly requires IDs on3,000 old unlabelled warmup tasks; the correction keeps
new captures strict while explicitly accepting that old fixture's warmup layout.
All12,000 legacy state/region partitions and9,000 old correlations then pass.
An empty `GO` environment assignment caused a separate analysis invocation failure,
not a trace defect. Parsing the full pair exceeded one shell timeout after current
output completed; upstream alone completed in175s. No application capture repeats.

## Encoder control

Valid [current CPU-clock attribution](../native-perf-attribution-20261013/README.md)
and the preparation spans support examining ordinary compression work. They do
not establish extra compression CPU versus upstream. A separate finite experiment
compares Go1.27.1 standard gzip with `github.com/klauspost/compress/gzip` **v1.20.3**,
explicit level6, unchanged finite buffering, headers, writer lifecycle and default
GC. No level sweep, cache/queue increase or authorization change is made.

The preregistered description incorrectly calls current Go level6 chain matching:
**Go1.27.1 also uses a specialized fast encoder**. The actual pinned implementations
and binaries are unchanged. This is a measured implementation comparison, not proof
that only the library has specialized matching. Primary source:
[tagged deflate implementation](https://github.com/klauspost/compress/blob/v1.20.3/flate/deflate.go).

One unchanged native production capture yields **256 actual POST receipts**,
using the maintained client's headers, canonical bodies and random-prefix shape,
one worker and256 offered posts in one second. Every receipt and persisted write
passes exact audits; every stock gzip member is reproduced byte-for-byte. Three
sequential interleaved standalone pairs each compress10,000 bodies per arm from
this corpus. Independent standard-gzip decoding is exact for both encoders.

| Standalone medians | Standard | Library |
|---|---:|---:|
| OS CPU per body, µs | 37.2389 | 23.8432 |
| Aggregate corpus wire bytes | 500688 | 500688 |
| Retained encoder state | about1.09MB | about1.09MB |

The **36.0% CPU reduction**, zero aggregate/individual wire growth and exact decoding
pass the declared25%/3%/5% work gates. This single-core,256-body replay is not the
complete frontend/media distribution or application capacity evidence.

### Application promotion gate: failed

The private candidate changes only three front compressor imports and its module
requirement. It preserves pool/cache budgets, finite bodies, public streaming,
headers, cancellation and teardown. Standard-library decoders remain in tests and
outbound integration compression is unchanged. Existing local front/web tests pass;
native front/web races and vet pass with reference fixtures and the independent
zstd test decoder present.

One native complete-workload pair uses the unchanged maintained client **`4f8848…`**,
3,000/s,64 workers,2+20seconds. No profiler or timing hook runs; a joined supervisor
reports complete-process `wait4` CPU. All **132,000 writes** pass canonical receipts,
persistence/FTS, integrity, unchanged FK and clean shutdown.

| Joined complete-process CPU | Standard baseline | Candidate |
|---|---:|---:|
| Seconds for66,000 writes | 33.686408 | 34.365532 |

Observed CPU increases **2.02%**, rather than the required5% reduction. The work gate
fails: **no fixed/capacity/Cable controls, production dependency change or gate
refinement follows**. A single diagnostic pair does not isolate the cause of the
increase. Standalone savings do not establish application benefit.

The baseline is byte-identical to the prior production binary **`8a1409c…`**;
candidate **`9ca09cd…`** differs only in the private compressor substitution and
module metadata. Restoring byte identity required the original embedded trees,
including existing AppleDouble files. `COPYFILE_DISABLE` correctly stopped adding
new resource forks, but initially omitted those already embedded in the prior
baseline. Clean-assets and partial-restoration builds were rejected **before posting**.
Both final arms use identical embedded data; this is not an asset cleanup claim.

Other retained setup failures are missing zstd in the disposable build container,
missing module checksum, omitted reference vectors/image fixture and an existing
client symlink. They are corrected without skipping checks. Test-only zstd installs
occur only in disposable containers; the runtime image, host and network settings
are unchanged. The candidate's native application budget is closed.

## Evidence and limits

[Manifest](manifest.json) retains exact private overlays/adapters, budgets,
client/helper/module identities, all ACK sidecars,48,000 correlated rows, replay
outcomes,256 decoded/wire corpus bodies and failed/corrected setup logs. In total,
**192,256 audited writes** and **60,000 standalone compressions** are retained.
Runtime traces remain private; their hashes are public. Credentials, cookies,
request logs, databases, storage, TLS private keys, binaries and the private
Tailscale snapshot are excluded.

Production remains exact8351224; its prior `bin/check` evidence remains valid because
no application source/configuration changed. Rejected private code does not require
new shared browser/media acceptance runs. These diagnostics neither overturn the
current three-way ranking nor establish a tail cause. Another candidate needs a
source-grounded complete-workload mechanism, not another standalone encoder score.
