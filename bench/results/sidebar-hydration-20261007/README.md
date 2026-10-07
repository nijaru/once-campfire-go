# Narrow sidebar hydration

The sidebar now loads retained direct participants in one scoped query rather than one full-user
query per visible direct room. An owned `RoomParticipant` carries only ID, name and avatar version;
credentials, profile text and authorization state stay in `User`. Direct settings still use full
retained `RoomMembers`. Other direct-room titles and broadcasts use the same narrow projection;
a broadcast reads display participants once while selecting active recipients separately.

Placeholder selection is one query with a materialized distinct exclusion set instead of moving
IDs into Go and issuing a second full-user query. The set includes invisible and NULL-involvement
directs and inactive/banned participants. The viewer is excluded separately and still consumes an
additional slot even when already in the set: `max(0,19-distinct_count)`. Active bots remain eligible.
Existing directs are not capped.

The batch is scoped to IDs from the retained visible-room read, not a second visibility test.
A read-only review caught that rechecking visibility could render a concurrently hidden direct as
a self-ping; that draft was corrected before authoritative builds and timing. Room entry and Cable
publication retain their independent fresh authorization. No result cache, schema/driver change,
detached context or permissive timestamp scanner was introduced.

## Before/after

Three sequential interleaved repetitions compared `6c3dde3` with the candidate on identical seed
snapshots, gzip, 16 concurrent clients, guest server CPUs 0–3/client CPUs 4–7, excluded two-second
warmups and ten-second samples. Both release binaries used the same ARM64 OrbStack environment on
an Apple M3 Max. No builds, tests or agents ran during timing. Every complete decoded sidebar body
matches the before implementation after normalizing only its known listener origin.

| Sidebar | Before | After |
|---|---:|---:|
| Median requests/second | 15,588 | 18,243 |
| Median CPU µs/request | 159.13 | 140.20 |
| Median saturated p99 ms | 5.771 | 5.475 |

Throughput improves 17.0%, CPU/request falls 11.9%, and saturated p99 falls 5.1%. Every pair has
higher candidate throughput, but one gains only 2.0%; VM variation remains material. This is not a
matched-offered-rate latency claim. The separate diagnostic profile is excluded from these medians.

## Fresh Rust comparison

A separate alternating Go/Rust comparison used the same fixture, workers, affinity, gzip, warmups,
three repetitions and complete-response/write validation. Median capacity in requests/second:

| Route | Rust | Go |
|---|---:|---:|
| Room | 29,570 | 23,748 |
| Messages | 31,213 | 22,694 |
| Sidebar | 29,174 | 17,166 |
| Search | 26,120 | 21,019 |
| Post | 3,235 | 2,718 |

**Rust-level capacity has not been reached.** This run's writes are less variable than the previous
baseline: Rust 3,210–3,237, Go 2,683–2,724. It does not invalidate the earlier noisy measurements or
establish native Intel behavior. Guest affinity is not dedicated physical-core isolation.

At an independent offered rate of 500 requests/second, with a 64-request concurrency bound, neither
application drops, expires or leaves requests unscheduled. Median service p99 (ms): sidebar Rust
5.520/Go 6.896; posts Rust 8.160/Go 9.968. Scheduled p99 includes generator lateness: Rust sidebar's
20.704 ms second sample coincides with 12.512 ms generator-lateness p99 and 7.072 ms service p99.
Do not interpret that scheduled outlier as application service latency.

Across before/after, fresh capacity and matched-rate comparisons: 18 runs, 48 timed samples,
7,228,361 complete responses, zero HTTP errors. Posts including warmups are checked against both
committed messages and FTS rows. Raw data, source/binary/seed/tool identities, contracts, gate logs
and diagnostic profile are archived. The comparison log also retains two setup-only CLI failures;
neither produced timed samples. No browser sweeps or Ruby/Elixir comparisons were made.

## Verification

Native CGO `bin/check`, Linux CGO vet/full race tests with `sqlite_fts5`, the WebSocket-fork race suite
and all 15 paired workflows pass. Focused database/web tests pass 20 race repetitions. Coverage
includes multiple direct batches, more than four participants, inactive/banned members, self-pings,
NULL/invisible membership visibility, placeholder count boundaries, active bots, equal activity
timestamps, independently committed participant changes, cancellation and a single-reader pool.
Full template/cache freshness tests remain in place. Schema and source vectors are unchanged.
