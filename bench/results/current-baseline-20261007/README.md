# Current Go/Rust server baseline

Go source `218374b` is compared directly with pinned Rust `64f8635`. This is a baseline after the
correctness fixes, not evidence of a new optimization or a reproduction of the historical
multi-language table. Ruby and Elixir were not measured.

Both release binaries ran sequentially in the same ARM64 Linux environment under OrbStack on an
Apple M3 Max. Each application had four guest CPUs (0–3); the client had CPUs 4–7. These are guest
CPU affinities, not dedicated physical cores. Three alternating repetitions used identical seed
snapshots, gzip, two-second warmups and ten-second samples. No builds, tests, agents or asset
generation ran during timing.

## Capacity

Median completed requests/second at 16 concurrent clients:

| Full application response | Rust | Go |
|---|---:|---:|
| Room | 29,271 | 23,665 |
| Messages page | 29,812 | 19,341 |
| Sidebar document | 27,183 | 11,505 |
| Search | 24,758 | 17,274 |
| Post message | 2,189 | 2,704 |

See [capacity/report.md](capacity/report.md) for ranges, latency, CPU, memory and response sizes.
Go's sidebar gap is substantial in this environment. The write median is not a robust Go-win
claim: Rust ranged from 1,449 to 3,252 requests/second and Go from 1,037 to 2,716. Several third-run
results slowed materially. Historical native Intel results are a different population.

## Matched load

At 500 offered requests/second, with a 64-request concurrency bound, all scheduled requests
completed. There were no dropped, expired or unscheduled requests. Median run-level p99 values:

| Route | App | Service latency | Scheduled latency | Generator lateness |
|---|---|---:|---:|---:|
| Sidebar | Rust | 5.408 ms | 13.024 ms | 7.928 ms |
| Sidebar | Go | 6.392 ms | 9.440 ms | 4.592 ms |
| Post message | Rust | 13.056 ms | 18.496 ms | 4.568 ms |
| Post message | Go | 12.592 ms | 14.720 ms | 4.728 ms |

These columns are separate distributions; their percentiles cannot be added. One Rust sidebar
sample had scheduled p99 316.416 ms, including generator lateness p99 307.712 ms, while service
p99 was 7.064 ms. It is retained, not mistaken for an application-only regression. Variance and
millisecond generator lateness prevent a strong tail-latency winner claim here.

## Validation and profiling

Twelve comparison runs produced 42 timed HTTP samples and **5,485,629 completed timed responses,
zero HTTP errors**. All **214,537 acknowledged posts**, including write warmups, were verified in
both messages and FTS. Full bodies were consumed; decoded message/room contracts, response sizes,
encoding, source/binary/seed hashes and load observations are in the archived raw data and metadata.
The required application tests separately protect sidebar document/frame completeness and fresh
layout/session/authorization behavior. No legacy-cursor mask was used.

A separate Go-only sidebar CPU profile is diagnostic, not part of the comparison medians. In that
sample, `sidebarRooms` accounted for about 20% of sampled CPU cumulatively and outer `render` about
24%; SQLite row iteration and CGO were prominent. The next performance investigation should focus
on sidebar data hydration and document rendering, preserving fresh authorization and external
SQLite visibility. This record does not justify another cache, SQL rewrite or relaxed cancellation
without a measured before/after improvement.

The profile, exact commands, logs, raw samples and metadata are retained here. This small-seed,
HTTP/1.1 application-listener run does not establish large-installation performance, native Intel
performance, TLS throughput, media speed, or broad language rankings.
