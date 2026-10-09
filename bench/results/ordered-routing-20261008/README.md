# Ordered HTTP routing

This cohort replaces the two-pass controller/private-mux dispatcher with one
ordered contract match and a prepared handler binding. Controller captures are
decoded once, callers use their canonical names, and dispatch leaves the request
URL intact. A method/first-segment index excludes impossible contract matches
without changing order within a group. Ordinary canonical paths avoid normalization
allocations; escaped or repeated-slash paths retain their existing normalization.

The previous production build is `98c6914`. `candidate.patch` identifies the
measured source and tests; `metadata.json` records source revisions and binary
hashes. Current Go upstream, Rust and shared verification were refreshed after
measurement and remained unchanged. The original pinned route vectors still
protect recognition order, defaults, formats and escaped parameters.

## Correctness and ownership

The old private regex router, all per-module path registrations, and the bot path
regex are removed. Prepared bindings retain browser checks, authentication, storage
capabilities, reference 404/500 actions, mailbox responses and native navigation.
Cable remains a separate exact-path protocol endpoint: unsupported methods and
encoded path aliases are rejected before authentication, as before. Bot cookie
writes remain subject to Fetch Metadata; only actual bot-key authentication exempts
writes. Decoded slashes and the reserved nested-message key are still rejected
before authority checks.

A demonstrated bot bug is deliberately fixed: a key beginning with `boosts` used to
select a boost action on a message route. The focused regression fails on the
unchanged previous source with a 404, then creates and retrieves the message on
the new dispatcher. This behavior change is also recorded in the main README.

The existing route vectors, focused five-repeat races, ten-repeat protocol rejection
checks, full `bin/check`, Linux vet/races, and fresh native/Linux builds pass. Current
Chromium flows pass with the previously disclosed persisted-echo wait before the
unchanged exact-three-message assertion. The real public listener passes current
unmodified HTTP/SQL oracles, seven preflights, twenty responses, four exact
acknowledged-write/FTS audits and the forwarding-disabled scheme-alias probe.
An independent read-only source review found no P1/P2 blocker. Its identified
protocol-rejection coverage gap was addressed before measurement; the reviewer
ran no tests or benchmarks.

Early tests caught two mistakes in the new candidate index: optional formats were
included in its definition key, and formatted static paths were not assigned to
their base segment. Both were corrected before successful gates or measurement;
no reference expectations or test assertions changed.

## Matched application measurements

The current shared harness checks complete responses and exact acknowledged writes
against independent SQLite expectations. The only temporal-comparison adapter adds
`go-before` as a Go label; its patch is included here. Every timed response and
write audit passes, with zero errors or invalid responses.

Settings: three sequential interleaved rounds, eight timed seconds and two excluded
warmup seconds, sixteen clients and one viewer; four runtime workers/readers and
three workers per job queue. Server CPUs are 0–3 and client CPUs 4–7. Completed
response and fragment budgets are 64/32 MiB, in addition to inherited public/gzip
budgets. Three-build app order alternates forward/reverse; the candidate remains
in the middle rather than rotating through every position. No builds, tests,
reviewers or CPU profiling overlap timed comparisons.

| Warm route | Previous median req/s | Candidate median req/s | Change | Current Rust median req/s |
|---|---:|---:|---:|---:|
| Room | 23,347 | 25,548 | +9.4% | 38,822 |
| Messages | 25,310 | 25,453 | +0.6% | 39,188 |
| Sidebar | 26,703 | 28,153 | +5.4% | 42,822 |
| Search | 26,916 | 27,982 | +4.0% | 41,707 |
| Post | 4,569 | 4,484 | −1.9% | 5,577 |

The candidate improves two of three room pairs, two message pairs, and every
sidebar/search pair. The first previous room round slows to 16,897 req/s while
later rounds reach 23,347–25,550; the last candidate room pair regresses. Posting
has a large candidate second-round slowdown to 2,659 req/s, versus 4,484–4,533 in
its other rounds. Rust's last posting round also slows substantially. These samples
remain in the raw evidence. The result is not an across-the-board speedup claim.
Go's warm-read medians are 65–67% of current Rust's, and its posting median is 80%.

| Reads with ten paced HQ writes/sec | Previous median req/s | Candidate median req/s | Change | Current Rust median req/s |
|---|---:|---:|---:|---:|
| Room | 29,744 | 31,682 | +6.5% | 36,633 |
| Messages | 29,893 | 31,672 | +6.0% | 36,793 |
| Sidebar | 35,958 | 37,430 | +4.1% | 42,012 |
| Search | 35,506 | 36,004 | +1.4% | 41,457 |

Every mixed room/message/sidebar pair improves; search regresses in the second
pair. The candidate's mixed-read medians are 86–89% of Rust's in this population.
Mixed Headquarters writes do not write to the displayed Watercooler room and do
not establish same-room fanout capacity. Rates differ substantially between warm
and mixed runs on this VM; do not combine cohorts into an isolated speedup factor.

Warm comparisons validate 9,220,855 timed responses and audit 394,824 writes
including warmup. Mixed comparisons validate 10,143,445 timed reads, 2,878 timed
writes and 3,598 acknowledged writes including warmup. Peak generator CPU is
51.1% warm and 59.4% mixed, with 400% available. This is not proof of an
unconstrained generator under every workload.

## Profiles and limitations

The diagnostic warm profile uses two five-second read rounds and graceful process
shutdown to flush the existing opt-in CPU profile. Both profiles were copied before
another run could overwrite the runtime directory. At 53.19 CPU seconds over
31.57 wall seconds, recognition accounts for 0.77 seconds (1.45%) and regex
submatches for 0.34 seconds (0.64%). The predecessor's separate diagnostic profile
recorded recognition at 3.57%. Different request counts and CPU parallelism mean
these percentages are not a normalized per-request cost comparison. The removed
private mux has no remaining production path. Network writes, SQLite observation,
authentication and allocation remain substantial costs.

`warm.json.gz` and `mixed.json.gz` contain all round summaries, raw samples,
contracts and checks. `profile-warm.json.gz`, profiles/listings and gate logs retain
diagnostic and verification evidence. Runtime databases and write-audit files were
consumed by the exact oracle before the harness's normal cleanup; per-round audit
results and counts are retained.

These are ARM64 OrbStack VM measurements on macOS, not upstream's published x86
machine. Guest affinity is not physical core isolation, and host variance is
substantial. Tmpfs data validates transactions/FTS, not NVMe throughput or crash
durability. One viewer with cache-friendly settings does not establish general
multi-user capacity or eviction behavior. Throughput-run tails are not matched-load
latency; no allocation snapshot was taken. Native Intel remains unavailable and
the CLI's thirty-second forced-exit path remains unexercised. Media algorithms are
unchanged and unrelated media inventories were not rerun.

This completes ordered controller dispatch, not the whole architecture plan.
Completed-representation coding/emission ownership and aggregate resource work
remain unfinished; the current Rust performance gap is not closed.
