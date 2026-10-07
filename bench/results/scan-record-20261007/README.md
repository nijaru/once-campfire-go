# Reuse reference-query scan records

`MessagePageReferences` and `SearchReferences` now reuse one scan record per query. Each successful
row is still copied into the owned result slice. Previously, taking addresses of a new `Message`
inside each iteration made that record escape to the heap. SQL, result ordering, timestamp
validation, cancellation, authorization and cache identities are unchanged. No driver or schema
changes, JSON aggregation or retained query results were introduced.

A temporary native Go allocation probe used real SQLite queries with 40 page results and 100
search results. Its before functions reproduce the previous loops; its after functions call the
production methods. The probe source and output are archived rather than added to the permanent
suite. These are diagnostic allocations, not HTTP capacity measurements:

| Query | Before B/op | After B/op | Before allocs/op | After allocs/op |
|---|---:|---:|---:|---:|
| Page, 40 rows | 12,360 | 7,464 | 130 | 92 |
| Search, 100 rows | 50,736 | 38,064 | 320 | 221 |

The existing pagination and search tests already compare distinct IDs, rooms, timestamps and
order with full-record queries. They also exercise malformed timestamps, BLOB timestamps,
committed updates, empty results, missing/foreign cursors, revoked search membership and cancelled
queries. Both passed 20 race repetitions; no duplicate permanent tests were added.

## Complete application comparison

Three sequential interleaved repetitions compared the verified `5f12ecd` binary with the candidate.
Both ran on the same ARM64 OrbStack Linux VM, using identical seed snapshots, gzip, four server
workers, 16 clients, guest server CPUs 0–3/client CPUs 4–7, excluded two-second warmups and ten-second
samples. Complete decoded response bytes matched after normalizing only known listener origins.
No builds, tests or agents ran during timing.

| Route | Before req/s | After req/s | Before CPU µs/request | After CPU µs/request |
|---|---:|---:|---:|---:|
| Room | 23,290 | 25,384 | 118.41 | 114.36 |
| Messages | 22,607 | 23,482 | 128.35 | 125.49 |
| Search | 21,457 | 21,526 | 122.83 | 122.41 |
| Post control | 2,709 | 2,677 | 544.90 | 548.92 |

Room throughput improves 9.0% and CPU/request falls 3.4%; every room pair improves, but the third
pair gains only 1.0%. Messages improve 3.9% by medians, but one candidate sample drops to 19,421
req/s with higher CPU/request, so the evidence does not establish a consistent history gain.
Search is effectively unchanged. Post code is unchanged; its before samples include a 2,073 req/s
outlier. Keep those observations rather than attributing all variation to this two-loop change.
Room saturated p99 falls 5.227 → 4.571 ms. Capacity tails are not matched-offered-rate latency.

## Fresh Rust comparison

A separate three-repetition Go/Rust comparison uses the same workload and environment. Median
capacity in requests/second:

| Route | Rust | Go |
|---|---:|---:|
| Room | 26,546 | 25,394 |
| Messages | 30,817 | 22,223 |
| Search | 25,950 | 21,604 |
| Post | 2,634 | 2,699 |

The write medians differ from the preceding comparison where Rust was ahead. Neither small-seed
VM write population establishes a robust Go advantage. Rust still leads read capacity; **the
Rust-performance goal remains unfinished**. Sidebar production is unchanged, so its preceding
measurement was not repeated.

At 500 independently offered requests/second and a 64-request concurrency bound, median service
p99 (ms) is Rust/Go: room 3.484/5.448, messages 4.872/5.856, search 5.904/5.032, posts 7.288/17.856.
There are no dropped, expired or unscheduled requests. Scheduled p99 also includes generator
lateness: Rust search's second sample has 18.656 ms lateness p99; Go's third post sample has
51.264 ms lateness p99. Go's second post sample also has a real 25.760 ms service p99, not explained
by its 4.496 ms generator lateness. This change does not demonstrate a write-latency improvement.

Across before/after, fresh capacity and matched-rate comparisons: 18 runs, 72 timed samples,
8,523,328 complete responses, zero HTTP errors. All 418,683 acknowledged posts including warmups
match committed messages and FTS rows. Source/binary/tool/seed identities, raw measurements,
contracts, gate logs and the preceding mixed-route diagnostic profile are archived.

Native CGO `bin/check`, Linux CGO vet/full race tests with `sqlite_fts5`, WebSocket-fork race tests
and all 15 paired workflows pass. These checks preserve fresh authorization, body/cache ownership
and persistence. They do not establish universal malformed-input or browser parity.

Guest affinity does not establish dedicated physical cores. `/proc` CPU accounting is coarse,
VM variance remains substantial, and native allocation results cannot be combined with VM HTTP
rates. No native Intel, browser, Ruby/Elixir, TLS or large-seed comparison was made in this pass.
