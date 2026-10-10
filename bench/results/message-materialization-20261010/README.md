# Message materialization and allocation

At unchanged default `GOGC=100`, the candidate improves complete POST throughput
**7.3% against b1cad80**. Its matched upstream median deficit is **1.8%**, versus
17.7% in the earlier readiness cohort. Cache-disabled room/message/search reads
improve 25–41%. These results do not establish PR readiness: matched-rate latency
and distinct-user Cable fanout are still unfinished.

## Source changes

The HTML tokenizer uses a reader's known remaining byte length to avoid allocating
4 KiB for small in-memory fragments. An extra byte accommodates the EOF probe.
Unknown-size and large readers retain the original initial capacity; normal growth
and input limits remain unchanged. New observable tests compare string, byte and
buffer readers with a bounded streaming reader, including malformed/raw content,
UTF-8, 4 KiB boundaries and a buffer filled after tokenizer construction.

`MessageQueries.Views` already returns complete rendered fragments. Creation
publication and room refresh now consume those fragments directly instead of
printing them through another template and copying the whole message again.
Presentation remains the markup owner; failure handling, postcommit processing
results and independent notification/publication obligations are unchanged.

Native `bin/check`, affected native/Linux races and Linux vet pass. The richtext
oracle corpus, existing room-refresh checks, shared Chromium workflows and public
HTTP/write contracts pass. The browser still uses the disclosed persisted-echo
wait with its unchanged exact-three assertion. All real-public acknowledged-write,
FTS, scheme-alias, late-coding and involvement checks pass. No media algorithm,
reference file, schema, cookie or authority policy changed.

## Diagnosis before changing source

Temporary worktrees instrument b1cad80 and c4ab53d, never the production checkout.
Execution traces add two writer regions: acquisition includes pool waiting and
BEGIN; hold covers the transaction body through commit/rollback. Region counts
match acknowledged writes exactly, with no unmatched pairs. Both builds show
roughly 68–69 µs mean holds and 1.8 ms mean acquisition at this load. This does not
isolate the normal-build regression as a slower transaction body.

Trace collection takes about 45 seconds. Offline decoding unexpectedly takes
42 minutes; that time is analysis overhead, not application timing. CPU/blocking
profiles, raw region lines and summaries are retained here. Full traces remain
local, not in Git; their checksums are recorded in metadata. Region statistics can
be recomputed from the retained records, but regenerating blocking profiles needs
the full traces. Trace instrumentation perturbs scheduling: its throughput is
diagnostic only.

Memory diagnostics at default GC show the pre-change candidate completing about
440 collections and 930 ms of aggregate pauses, versus upstream's 150 collections
and 300 ms. It creates fewer objects but similar total allocated bytes. A separate
matched GOGC=200 diagnostic largely removes the throughput gap; neither the default
nor a memory limit is changed on that evidence.

Allocation profiles identify fixed tokenizer buffers and redundant materialization
as substantial costs. After the source changes, diagnostic allocation volume is
about **96 KB/write and 1,133 allocations/write**, versus upstream's **149 KB and
1,504**. These ratios include startup, warmup and shutdown, not an isolated operation
microbenchmark. Candidate aggregate pauses fall to 698–724 ms; upstream is 327–330
ms. End-of-run Go heap-in-use is 21–25 MiB versus 69–79 MiB, not peak RSS or an
aggregate resource bound. GC remains material; no optimality claim follows.

## Unprofiled capacity comparisons

Three sequential alternating rounds, eight timed seconds plus two excluded warmup
seconds, sixteen clients, one viewer, four runtime workers/readers, three workers
per job queue, gzip, server affinity 0–3 and client 4–7. No builds, tests, profilers
or agents overlap these populations. Two-build order alternates both positions;
the three-build order retains the candidate in the middle.

| POST population | Control req/s | Candidate req/s | Median change |
|---|---:|---:|---:|
| Tokenizer sizing only vs b1cad80 | 5,131.1 | 5,417.5 | +5.6% |
| Tokenizer sizing only vs upstream | 5,637.8 | 5,391.1 | −4.4% |
| Final candidate vs b1cad80 | 5,210.1 | 5,589.7 | +7.3% |
| Final candidate vs upstream | 5,723.2 | 5,618.0 | −1.8% |

The final candidate improves all three branch-control pairs. It trails upstream
in two pairs and leads in one; every slower sample is retained. Current Rust's
final matched median is 5,993.3 req/s. These rows are separate populations, not
isolated causal gains or a comparison against upstream's published x86 numbers.

| Both Go response and fragment caches disabled | b1cad80 req/s | Candidate req/s | Change |
|---|---:|---:|---:|
| Room | 942.0 | 1,216.2 | +29.1% |
| Messages | 1,061.1 | 1,499.5 | +41.3% |
| Search | 2,034.3 | 2,549.5 | +25.3% |

All nine uncached pairs improve. The exact-byte gzip memo remains enabled, so this
is repeated parsing/rendering, not cold compression or total cache removal. These
settings differ from normal warm-read evidence and must not be conflated with it.
Rust is not rerun in this two-build population.

The five unprofiled capacity populations validate 1,561,371 timed responses and
audit 1,658,733 writes including warmup. All errors and invalid responses are zero;
exact acknowledged ID/body/room/FTS and integrity audits pass. Diagnostic counts,
raw samples, code patches, source/image/binary hashes and settings remain separate
in metadata and archives.

## Limits and provenance

Source base is b1cad80 plus the retained final patch, committed with this evidence.
Branch control's production binary was built from d202a7f plus the exact patch that
became b1cad80. Upstream c4ab53d, Rust source 6dae2fd and shared source ec02deb were
refreshed and unchanged before unprofiled measurement. Rust's unchanged binary is
built at 9872c1d; loadgen's unchanged binary is built at e244051. Instrumented binary
sources and hashes are distinct from production capacity inputs.

The branch was rebased onto c4ab53d and the four oversized trace archives removed
from history. Application source is unchanged. Metadata maps measured revisions
to their rebased equivalents; binaries and measurements keep their original build
identities, rather than being relabeled as new builds.

ARM64 OrbStack/macOS guest affinity is not physical isolation. Shared host activity
is not eliminated. VM-host tmpfs audits do not measure NVMe or crash-safe durability.
Loopback rewritten push destinations bypass Go provider encryption/network work.
One viewer does not establish many-user capacity; saturation tails are not fixed-rate
latency. No GOGC tuning, stronger delivery guarantee, Rust parity or zero-bug claim
is made. The small remaining POST difference and broader latency/fanout/resource
questions remain visible rather than being called solved by a passing test suite.
