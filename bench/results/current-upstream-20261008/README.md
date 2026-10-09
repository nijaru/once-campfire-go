# Current-upstream comparison and reader-plan caching — 2026-10-08

The retained change replaces the global `readPool` preparation wrapper with the
existing SQLite driver's bounded, connection-local statement cache. Ordinary reads
and scoped read transactions now use the same cache; the wrapper and its global
prepare lock are removed. The single writer, bounded readers, read-only snapshots,
fresh authorization and serialized foreign-commit observer remain unchanged.
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

## Evidence and verification

Compressed JSON bundles retain every raw warmup/timed sample, per-round results,
image identities, source identities, audits and summary from each named population.
`metadata.json` adds binary hashes, toolchains and limitations; `reader-cache.patch`
identifies the measured candidate on `a2cd697`. CPU profiles and cumulative listings
are diagnostic baseline evidence. Their temporary runner stops containers gracefully
before removal so opt-in profile files flush; their rates are not capacity results.
The first mixed profile was overwritten by the next disposable runtime and was
rerun separately; `profile-mixed-retained.json.gz` corresponds to the retained files.

Native `bin/check`, full database/application/web races and five focused repetitions
of scoped-read, foreign-writer, generation, search and private-cache checks passed.
The fresh native build passed current shared Chromium and public HTTP/write audits.
Chromium uses the disclosed persisted-echo synchronization correction, with its exact
message-count assertion unchanged. Current HTTP/search contracts are unmodified.
The real public listener also passed the forwarding-disabled scheme-alias check.

Measurements use ARM64 Linux in OrbStack on macOS, with guest affinity rather than
isolated physical cores. Data is on VM tmpfs. Exact committed transactions are
verified, not crash durability or sustained NVMe throughput. Peak client CPU was
52.3%, 29.7% and 34.9% in baseline, warm and mixed populations respectively, where
400% is available. One viewer does not exercise multi-viewer eviction pressure;
throughput-run p99 is not matched-offered-load latency. Native Intel is unavailable.
The CLI's forced-exit path remains unexercised. No media algorithms changed, and
unchanged media acceptance was not repeated for this query-only change.
