# Current-upstream comparison and reader-plan caching — 2026-10-08

The retained changes use connection-local read plans and batch fresh generation
observations. The first replaces the global `readPool` preparation wrapper with the
existing SQLite driver's bounded, connection-local statement cache. Ordinary reads
and scoped read transactions now use the same cache; the wrapper and its global
prepare lock are removed. The single writer, bounded readers, read-only snapshots,
fresh authorization and foreign-commit validation remain intact.
Statements and bindings are reused, not query results. No dependency or schema changes.

## Current Go/Rust baseline

These are fresh measurements of this fork at `a2cd697` against current Rust `9872c1d`,
not the upstream README's x86 figures. Upstream Go `9b0ef06` is incorporated. The
shared verification runner is current `main` at `e244051`; revisions record the
measured inputs, not a frozen acceptance target.

| Route | Fork Go median req/s | Current Rust median req/s | Go / Rust |
|---|---:|---:|---:|
| Room | 13,898 | 40,287 | 35% |
| Messages | 13,941 | 38,552 | 36% |
| Sidebar | 15,122 | 45,502 | 33% |
| Search | 15,037 | 44,627 | 34% |
| Post | 4,031 | 5,570 | 72% |

The baseline validated 5,561,296 timed responses and audited 270,692 exact
acknowledged writes, including warmup, with zero errors or invalid responses.
Rust's newer completed-response cache makes the historical pinned-Rust comparison
unsuitable for describing the current gap.

## Why change reader preparation

Separate whole-application CPU profiles cover warm reads, reads with a paced HQ
writer, and message posts. The posting profile attributes 15.3% cumulative CPU to
statement preparation; message-display preparation accounts for 23.1%. Read transactions
bypass the old wrapper, so every post prepares its display queries again. These
percentages overlap other stacks and are not additive projected gains.

The read cache holds at most 256 idle plans per connection, including transactional
queries. The writer retains its existing 64-plan cache and the observer owns one
explicitly prepared statement. At four readers this allows 1,089 retained plans;
statement counts and cache payload budgets are not total-memory limits. The broader
resource-budget work remains unfinished.

## Matched before/after measurements

Each population runs serially in three alternating rounds, with 16 clients, gzip,
two seconds of excluded warmup and eight seconds per timed sample. Both binaries
use identical fixture snapshots, public listeners, four guest server CPUs and four
separate guest client CPUs. Each app has four readers, three workers per job queue,
64 MiB completed-response and 32 MiB fragment budgets. The current shared runner's
independent response contracts and exact acknowledged-write/FTS audits are unchanged.
`go-before.patch` adds only a second Go label/source directory for the comparison.

| Warm route | Before median req/s | Reader cache median req/s | Change |
|---|---:|---:|---:|
| Room | 13,141 | 11,516 | −12.4% |
| Messages | 13,212 | 12,934 | −2.1% |
| Sidebar | 13,933 | 14,306 | +2.7% |
| Search | 13,771 | 14,493 | +5.2% |
| Post | 2,969 | 4,154 | +39.9% |

Posts improve in two of three pairs; the third regresses 3,817→3,506. A separate
three-round POST-only control improves every pair and raises the median
2,328→3,955 (+69.9%). Its before range is 2,127–2,711 and after range 2,894–4,129.
The two controls support a posting benefit, not a precise universal percentage.

| Reads with ten HQ writes/sec | Before median req/s | Reader cache median req/s | Change |
|---|---:|---:|---:|
| Room | 11,050 | 11,308 | +2.3% |
| Messages | 13,212 | 11,719 | −11.3% |
| Sidebar | 16,577 | 14,947 | −9.8% |
| Search | 13,701 | 15,713 | +14.7% |

Mixed writes invalidate global epochs but do not target the displayed Watercooler
room. They do not measure same-room fanout or cache capacity under many viewers.
Warm comparisons validated 2,679,603 timed responses and audited 217,569 writes.
Mixed comparisons validated 2,437,369 timed reads, 1,915 timed writes and 2,395
acknowledged writes including warmup. The POST-only control audited 182,774 writes.
Every response and exact database/FTS audit passed. Raw samples retain all ranges,
tails, failed-pair regressions and host load.

This is a simpler preparation owner with a measured posting benefit, not an
across-the-board throughput improvement. Read regressions are retained. Host activity
and VM variance are substantial; for example, candidate warm-room runs span
8,326–13,087 req/s. No final candidate-versus-Rust capacity claim is inferred by
combining different populations. The large read gap remains unresolved.

## Batched fresh generation observations

The second cohort starts at the reader-cache commit `2f8b305`. One observer worker
owns its existing connection and statement. It batches only callers registered
before the physical SQLite query starts; arrivals during query/Scan/Close require
the next observation. It adds no TTL, local-only epoch, cached authorization or
in-flight singleflight reuse. Both existing cache-generation gates remain live.

At most 256 result channels queue, with at most 64 in a batch. Request contexts are
not retained. Cancellation abandons one waiter without cancelling another caller's
observation. Whole-close joins admission, then the worker, drains the queue and
closes SQL resources. The constructor's partial-open cleanup joins this owner too.

The following current three-build comparison uses the same controls as above.
Every warm-read pair improves; earlier two-round screening also improved every
read pair. Rust's raw second round is much slower and is retained, not discarded.

| Warm route | Reader cache median req/s | Batched observer median req/s | Change | Current Rust median req/s |
|---|---:|---:|---:|---:|
| Room | 12,667 | 21,409 | +69.0% | 32,885 |
| Messages | 12,435 | 22,241 | +78.9% | 30,771 |
| Sidebar | 13,572 | 22,444 | +65.4% | 33,843 |
| Search | 12,573 | 23,490 | +86.8% | 37,633 |
| Post | 3,943 | 3,853 | −2.3% | 4,748 |

The observer change improves reads, not posting. Post pairs fall in two rounds and
rise in one. Current Go is 62–72% of Rust's warm-read medians and 81% of its posting
median **in this population**. Do not combine those ratios with the earlier baseline
as an isolated language or end-to-end speedup attribution.

| Reads with ten HQ writes/sec | Reader cache median req/s | Batched observer median req/s | Change | Current Rust median req/s |
|---|---:|---:|---:|---:|
| Room | 6,762 | 19,881 | +194.0% | 22,131 |
| Messages | 11,225 | 22,613 | +101.4% | 25,756 |
| Sidebar | 10,234 | 27,775 | +171.4% | 33,950 |
| Search | 11,925 | 25,264 | +111.9% | 34,923 |

Every mixed read pair improves. These unusually large ratios include severe
slowdowns in the before population; they are not general capacity multipliers.
The mixed Go/Rust medians are 72–90%, with substantial Rust variation. No parity
claim follows. Warm comparisons validate 6,421,338 timed responses and audit
344,173 acknowledged writes; mixed comparisons validate 5,667,403 timed reads,
2,867 timed writes and 3,587 acknowledged writes including warmup. Errors and
invalid responses are zero, and every exact persistence/FTS audit passes.
Peak generator CPU is 55.1% warm and 49.8% mixed, with 400% available.

`coalescing.patch` and `coalescing-metadata.json` identify the candidate and binary
hashes. Current upstream Go, Rust and verification heads were refreshed after
measurement and remained unchanged. Coalesced warm CPU profiles are diagnostic
follow-up evidence, not a latency or allocation comparison. A scoped independent
read-only review found no P1/P2 blocker; it did not rerun tests or benchmarks.

## Evidence and verification

Compressed JSON bundles retain every raw warmup/timed sample, per-round results,
image identities, source identities, audits and summary from each named population.
The numbered rounds alternate forward/reverse app order; the new three-build
population keeps the candidate in the middle, rather than rotating all positions.
`metadata.json` adds binary hashes, toolchains and limitations; `reader-cache.patch`
identifies the measured candidate on `a2cd697`. CPU profiles and cumulative listings
are diagnostic baseline and follow-up evidence. Their temporary runner stops containers gracefully
before removal so opt-in profile files flush; their rates are not capacity results.
The first mixed profile was overwritten by the next disposable runtime and was
rerun separately; `profile-mixed-retained.json.gz` corresponds to the retained files.

Native `bin/check`, full database/application/web races and five focused repetitions
of scoped-read, foreign-writer, generation, search and private-cache checks passed.
The fresh native build passed current shared Chromium and public HTTP/write audits.
Chromium uses the disclosed persisted-echo synchronization correction, with its exact
message-count assertion unchanged. Current HTTP/search contracts are unmodified.
The real public listener also passed the forwarding-disabled scheme-alias check.
The observer cohort additionally passed Linux vet/races, ten race repetitions of cancellation,
concurrent foreign commits and joined closure, full database/application/web/runtime
races, `bin/check`, fresh builds, Chromium and current public HTTP/write audits.
A first real-driver contention fixture timed out inside `synctest`, because its
SQL mutex wait was not durably blocked; it was corrected to bounded real-time
waits. Admission cancellation retains its deterministic `synctest` check.
Linux verification also exposed a pre-existing room test comparing a nanosecond
creation receipt with a microsecond persisted timestamp. It failed on the unchanged
`2f8b305` implementation too. `0fd5df9` captures the stored room before rejected
updates and retains the exact whole-record and membership comparisons. No production
timestamp behavior or assertion was weakened. One temporary source-restore compile
reported an incomplete file; matching host/container hashes and a clean rerun passed.

Measurements use ARM64 Linux in OrbStack on macOS, with guest affinity rather than
isolated physical cores. Data is on VM tmpfs. Exact committed transactions are
verified, not crash durability or sustained NVMe throughput. Peak client CPU was
52.3%, 29.7% and 34.9% in baseline, warm and mixed populations respectively, where
400% is available. One viewer does not exercise multi-viewer eviction pressure;
throughput-run p99 is not matched-offered-load latency. Native Intel is unavailable.
The CLI's forced-exit path remains unexercised. No media algorithms changed, and
unchanged media acceptance was not repeated for this query-only change.
