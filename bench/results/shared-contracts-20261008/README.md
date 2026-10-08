# Shared contracts and session reads — 2026-10-08

Branch: `perf/campfire`. Baseline: `943597b`. The new candidate combines live user
and session-activity reads in `AuthenticateSession`, without caching authorization.
It retains the conditional hourly UPDATE and cookie-refresh election. Other
`SessionUser` callers are unchanged. Session operations now have one source owner;
the obsolete separate `RefreshSession` path is removed.

## Shared verification

Used `basecamp/once-campfire-verification` at
`7b2dbc7e856fb2cbf2d95833a22110beab026eec`: its contract preparation, production
load generator, per-response `route-contract-v1` checks, and exact acknowledged
ID/body/room/FTS audit. All eight initial routes passed for Go and Rust. The first
comparison also timed all eight, including real avatar, CSS and health responses.
Warmups and timed writes are audited. Across the five populations there were
29,511,047 valid timed responses, zero errors/invalid responses, and 519,794 verified
writes including warmups and paced writers. Populations are not pooled for performance.

Two harness corrections are explicit, not Go behavior changes:

- The shared browser's immediate message-count assertion can sample the sender's
  optimistic preview before its persisted echo. The prior
  [wait patch and diagnostic](../read-path-20261008/README.md) preserve the exact
  count assertion after waiting for persisted elements. The full flow passes.
- Shared `contracts.rb` expects search insertion-ID order. The pinned Rails
  `Message.ordered`/`Message::Searchable` and Rust `search_reachable` instead order by
  `created_at`. The original preflight fails on the real fixture: ID `933434483`
  sorts differently by creation time. `contracts.rb.patch` corrects that independent
  SQL oracle to chronological order; both implementations then pass. No HTML,
  expected ID, status or write failure is masked.

`compare.rb.patch` merely adds a `go-before` alias for interleaving two actual Go
builds, with separate source identity and image records. Validators are unchanged
by that alias. These changes are local to the private harness checkout; no upstream
PR/comment or push was made. The CLI wrapper does not silently patch checkouts.

The initial tools container used the C locale, causing Ruby's US-ASCII decoding to
reject valid UTF-8 HTML. `LANG=C.UTF-8` fixes the runner environment; the original
failure log remains here. This was not an application response defect.

Native `bin/check`, Linux vet/full race tests, the common browser flow and a focused
session regression pass. The regression checks full fresh user fields, no ordinary
authentication write, one refresher under concurrent hourly requests, and denial of
a banned user. Independent source review found no additional session defects;
that review was static, not an exhaustive correctness proof.

## Session before/after

Three sequential alternating rounds, eight timed seconds after two excluded warmup
seconds, gzip, sixteen clients, identical isolated seed snapshots. Server guest CPUs
0–3, client CPUs 4–7. Both builds run the complete public listener, not a stub or
internal helper. Binary/source hashes are in `metadata.json`.

| Route | Before median req/s | Candidate median req/s | Change |
|---|---:|---:|---:|
| Room | 12,824 | 13,344 | +4.1% |
| Messages | 13,199 | 13,684 | +3.7% |
| Sidebar | 14,197 | 14,960 | +5.4% |
| Search | 14,176 | 14,857 | +4.8% |
| Post | 2,717 | 2,744 | +1.0% |

Every warm throughput pair improves. POST is a control, not a substantial write-win
claim. Message p99 slightly worsens (5.299→5.359 ms); other warm median tails are
flat or better. Under the shared different-room writer at at most ten messages/sec,
read medians improve 6.1–6.9%. One candidate search pair regresses (15,685→13,974),
and one baseline room pair is unusually slow (9,479). Raw ranges and tails remain
in `session-mixed.json.gz`; do not hide these behind medians.

The prior generic-row-scan query-fusion experiment was rejected after unstable
results. This version uses a direct typed scan and leaves the existing user-row
helper untouched. Fresh, stable shared warm comparisons provide the new evidence
for retaining it; its protection and earlier rejection have not been erased.

## Fresh final Go/Rust comparison

Same shared validators, topology, seed and three alternating rounds after the
session change. These are capacity medians, not latency at a matched offered rate.

| Route | Go req/s | Rust req/s |
|---|---:|---:|
| Room | 13,517 | 20,782 |
| Messages | 13,843 | 22,655 |
| Sidebar | 14,860 | 20,711 |
| Search | 15,024 | 21,704 |
| Post | 2,734 | 3,182 |

Rust remains ahead in every route. Improvements do not establish parity or optimal
performance. Baseline warm and mixed Rust comparisons are separately retained;
no new Rails or Elixir performance claim is made.

## Reproduction and limits

The unmodified shared `bin/benchmark` entry point runs in an ARM64 Linux tools
container with the Docker socket, mounted canonical host paths, Ruby 3.3/SQLite CLI,
and the shared release load generator (built with Rust 1.98.0). Application images
wrap the verified release binaries in the same existing comparison-toolchain base;
they are not the projects' published production images. Go has `GOMAXPROCS=4`;
Rust has four reader/job/runtime workers. The runner's advertised 64 MiB response-cache
budget is not a fact about the pinned Rust binary: it has a 32 MiB fragment store
but no corresponding completed-response cache setting. Go retains its 64 MiB
completed-response cache and existing bounded nested caches. This is a defaults
comparison, not an equal-memory-cache comparison. Exact images/topology, source snapshots,
seed hashes, generator CPU and load samples are retained in each compressed record.

`comparison` is the pre-session eight-route Go/Rust population; `mixed` is its
separate shared churn profile; `session-*` compare old/new Go; `final-comparison` is
the final five-route Go/Rust population. The diagnostic profile is excluded from
these comparisons: it attributes 16.0% cumulative CPU to SQLite transaction commits,
4.8% to user reads and 3.0% to the now-removed separate activity read. Percentages
overlap and are not additive projected gains.

Guest CPU affinity is not dedicated physical-core isolation. Workstation/VM load
can vary; no fresh native Intel, TLS, Cable, zstd, large-seed or media-processing
benchmark is claimed. The seed is the existing pinned Rails parity seed, not a newly
generated shared fixture. HTML/write correctness here is the common contract,
not an invented universal UI or crash-equivalence matrix.
