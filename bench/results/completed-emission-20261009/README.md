# Completed private response emission

This cohort gives finite application output one private coding/emission owner in
`internal/front/response.go`. It establishes validators, conditional responses,
coding, Vary and selected-body length before writing headers. Streaming files and
upgraded sockets remain separate. Public zstd/gzip selection, jitter, cache policy,
cookie stripping and forwarding policy are unchanged.

The previous production build is `134f4dd`; `candidate.patch` identifies the
measured code and tests. Source revisions and before/candidate binary hashes are
in `metadata.json`. Current upstream Go, Rust and shared verification were fetched
before comparison and remained unchanged.

## Contracts and verification

The private response cache no longer creates gzip members or joins the identity
body before compression. It retains a copied completed representation after the
encoder selects its bytes, with fresh generation checks and live authority/cookies
unchanged. Its budget and admission cost now measure selected encoded bytes.
The exact-Part encoding memo and HTTP response cache remain distinct, separately
bounded stores; a miss can retain copies in both. Retained budgets are not aggregate
active-memory bounds.

The old optional `WriteBody` encoder path is removed. Ordinary and recorded-body
validators retain their existing distinct constructions. Conditional evaluation
still gives entity tags precedence over modification dates. Completed emission
bypasses streaming compression and cannot append an empty gzip member. No-store
still permits coding but not retention; no-transform prevents coding. Existing
encodings, empty completed bodies and no-body statuses keep their exceptions.

Two demonstrated defects are fixed:

- Successful avatar/logo files no longer pass through the finite body buffer.
  Previously, a default logo GET had length 405,666 and a body-derived validator,
  while HEAD had length zero and no validator. Files now use their file-controller
  length/validator policy. Failures before selecting a file remain ordinary finite
  error responses; session-cookie commitment remains live.
- Uncached completed gzip now has its selected byte length, including HEAD.
  Conditional coding/Vary metadata no longer depends on response-cache warmth.

Both regressions fail on the unchanged previous source; original failures are
archived. Existing authority, foreign-write, flash, variant, no-store/no-transform,
range, cookie and gzip-lifecycle tests pass. Redundant old private call-sequence
assertions in gzip tests were replaced with the new observable completed-response
contract; round trips, exact-Part reuse, no-retention, HEAD and single-member
protection remain.

Native repeated races, full `bin/check`, Linux vet/races and fresh native/Linux
builds pass. Current Chromium flows pass with the previously disclosed persisted
wait before the unchanged exact-three assertion. The real public listener passes
seven preflights, twenty responses, four exact ID/body/room/FTS write audits and the
forwarding-disabled scheme-alias probe. Independent read-only contract and resulting
source reviews ran no tests or timings; the latter found no P1/P2 blocker.

The first full gate also caught a delivery error in the previous evidence archive:
`ordered-routing-20261008/Dockerfile.go` was treated as Go source by `go vet ./...`.
It is renamed `Dockerfile`; subsequent full gates pass. No application behavior or
verification assertion was changed to resolve that failure.

## Matched measurements

The current shared harness validates full responses and exact acknowledged writes
with independent SQLite expectations. Its only temporal adapter allows the
`go-before` label; the patch is included. All recorded errors and invalid responses
are zero, and every exact-write/FTS audit passes.

Settings: three sequential interleaved rounds, eight timed seconds plus two excluded
warmup seconds, sixteen clients and one viewer; four runtime workers/readers and
three workers per job queue. Server CPUs are 0–3 and client CPUs 4–7. Completed
responses/fragments use 64/32 MiB, except the uncached population disables the
completed-response cache. Inherited gzip/public cache budgets remain 32/64 MiB.
Three-build order alternates forward/reverse but keeps the candidate in the middle;
the two-build uncached comparison alternates both positions. No tests, builds,
profilers or agents overlap timing.

| Warm route | Previous median req/s | Candidate median req/s | Change | Current Rust median req/s |
|---|---:|---:|---:|---:|
| Room | 29,760 | 30,239 | +1.6% | 42,402 |
| Messages | 31,380 | 31,917 | +1.7% | 40,795 |
| Sidebar | 34,794 | 35,610 | +2.3% | 46,878 |
| Search | 33,205 | 33,073 | −0.4% | 45,968 |
| Post | 4,588 | 4,517 | −1.6% | 5,704 |

Room/messages improve two pairs, sidebar every pair, search two pairs and post two
pairs. Small medians and per-pair regressions are retained, not presented as an
across-the-board speedup. The Go warm-read medians are 71–78% of current Rust,
and posting is 79% in this population.

| Reads with ten paced HQ writes/sec | Previous median req/s | Candidate median req/s | Change | Current Rust median req/s |
|---|---:|---:|---:|---:|
| Room | 29,485 | 32,183 | +9.2% | 37,824 |
| Messages | 26,976 | 34,139 | +26.6% | 35,899 |
| Sidebar | 33,003 | 40,511 | +22.8% | 42,533 |
| Search | 32,697 | 33,741 | +3.2% | 41,352 |

These differences are not isolated causal gains. The previous last room/message
round collapses to 2,647/2,361 req/s, versus 26,976–37,115 in earlier runs. The
candidate's last room/sidebar rounds also slow substantially. Mixed sidebar
regresses in two pairs, messages/search in one each, despite positive medians.
All slow rounds remain in the evidence. HQ writes do not target the displayed
Watercooler room or establish same-room fanout capacity. Different cohorts on this
VM have materially different throughput and variance.

| Completed-response cache disabled | Previous median req/s | Candidate median req/s | Change |
|---|---:|---:|---:|
| Room | 15,905 | 16,121 | +1.4% |
| Messages | 17,027 | 17,026 | approximately unchanged |
| Sidebar | 13,594 | 14,441 | +6.2% |
| Search | 16,447 | 16,022 | −2.6% |
| Post | 4,442 | 4,426 | −0.4% |

The uncached population still reuses fragments and exact-byte gzip memo entries;
it is not cold rendering. Room/sidebar/messages improve two pairs, post one;
search regresses in every pair. Rust is not rerun in this two-build control. These
regressions are retained alongside the correctness/ownership improvement.

Warm runs validate 10,820,309 timed responses and audit 438,798 writes including
warmup. Mixed runs validate 9,655,005 timed reads, 2,873 timed writes and 3,593 exact
acknowledged writes including warmup. Uncached runs validate 3,310,314 timed
responses and audit 266,184 writes including warmup. Peak generator CPU is
54.5%, 64.2% and 37.9%, respectively, out of 400% available. Those values do not
prove the generator is unconstrained under every workload.

## Profiles and remaining work

A separate two-round five-second mixed diagnostic flushes the CLI CPU profile via
graceful shutdown; both profiles were copied before the runtime directory could be
reused. It records 53.79 CPU seconds over 31.95 wall seconds. Completed preparation
uses 0.94 seconds (1.75%); emission's 17.12% includes network IO, not pure wrapper
overhead. Allocation, fresh SQLite observation, authentication, cookie verification
and JSON serialization remain substantial. No matched-rate latency or allocation
snapshot is claimed; the diagnostic does not reproduce every slow timing round.

Raw summaries, samples, contracts and audits are retained in `warm.json.gz`,
`mixed.json.gz`, `uncached.json.gz` and `profile-mixed.json.gz`. Consumed runtime
SQLite databases and ack files receive the harness's normal cleanup; their verified
results/counts remain. Profiles, original failures and gate logs are also included.

This is ARM64 OrbStack on macOS, not upstream's published x86 machine. Guest affinity
is not physical core isolation. Tmpfs verifies transactions/FTS, not NVMe throughput
or crash durability. One viewer does not establish multi-user/eviction/fanout
capacity. Throughput-run tails are not matched-offered-rate latency. Native Intel
is unavailable and the CLI forced-exit path remains unexercised. Media processing
is unchanged and unrelated media inventories were not rerun.

Public streaming compression and aggregate active/retained resource work remain
separate unfinished goals. Materializing uncached completed gzip also retains an
additional coded body until emission; that is not an aggregate memory bound.
The resulting implementation is simpler, but neither global optimality nor Rust
performance parity is established.
