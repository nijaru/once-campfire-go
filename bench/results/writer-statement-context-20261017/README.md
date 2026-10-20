# Writer statements without the driver goroutine — rejected, 2026-10-17

This was a CPU/allocation experiment. A local Mac pair showed a reduction, but
the native Fedora gate found no improvement and a small CPU regression. The code
was reverted in `9d9e614`; this record is retained as a closed experiment.

## Why it was tried

`mattn/go-sqlite3` takes a goroutine-plus-select path for every statement when
the context has a Done channel. A current-head execution trace showed the slowest
one percent of POSTs waiting in `create.unread`: 7.7 ms of waiting against 2.6 us
of work, with `writer.acquire` at 400 us. The delay was the spawned statement
goroutine, not SQLite or room work.

The candidate ran the short `createMessage` statements with a non-cancelable
context and checked the caller's context between statements and before commit.
`database/sql` still rolled a canceled request back; the regression asserted that
nothing commits.

## Local Mac pair

Apple M3 Max/macOS arm64, Go 1.27.1, shared parity seed, four runs in
baseline/candidate/candidate/baseline order. Each run is 3,000 offered POSTs/s,
64 workers, an 8,192-slot queue, two warmup and twenty timed seconds. All
**240,000 timed writes** pass exact ID/body/client/creator/room/FTS audits;
SQLite integrity and the unchanged seed FK population pass.

| Run | Baseline CPU, s | Candidate CPU, s |
|---|---:|---:|
| Pair 1 | 29.965 | 28.079 |
| Pair 2 | 26.324 | 25.527 |
| Paired delta | — | **-6.29% / -3.03%** |

CPU per timed write: 499.4/438.7 us baseline, 468.0/425.5 us candidate. A local
create-only benchmark measured **92.3 -> 88.5 us/op (-4.1%)**,
**5.34 -> 4.17 KB/op**, and **122 -> 100 allocs/op**.

The Mac was shared and had no CPU affinity. The first p99 samples were
contaminated (baseline 53.6 ms, candidate 364.5 ms); the clean later pair was
1.752 vs 1.708 ms. That pair is not native latency evidence.

## Native Fedora gate — decisive

Fedora i9-13900KF, 32 CPUs, Go 1.27.1, matching parity seed, candidate built in
the publication image. Server CPUs `0,2,4,6`; client CPUs `10,12,14,16`. Four
runs in baseline/candidate/candidate/baseline order at 3,000 offered POSTs/s,
64 workers, two warmup and twenty timed seconds. The host was idle; each run
passes 60,000 timed exact ID/body/client/creator/room/FTS audits, SQLite
integrity and the unchanged seed FK population.

| Run | Baseline CPU, s | Candidate CPU, s |
|---|---:|---:|
| Pair 1 | 34.873 | 35.154 |
| Pair 2 | 34.844 | 35.087 |
| Paired delta | — | **+0.81%** |

| | Baseline | Candidate |
|---|---:|---:|
| p99, ms | 0.968 / 0.987 | 0.961 / 0.991 |
| mean, ms | 0.498 / 0.493 | 0.499 / 0.498 |

The candidate is slightly worse in CPU and mean latency; p99 shows no material
change. The local Mac reduction did not transfer to the native target. **The
optimization was rejected and reverted** to the `5741508` statement-context
behavior; its cancellation regression test was removed with it.

## Verification and limits

The reverted source is byte-identical to `5741508` in the affected files.
`bin/check` passes on the reverted head. The earlier local shared Go/Rust
`message_lifecycle` and `pagination` workflows passed on the candidate; they are
not evidence for the native latency question. Compact numbers are in
[summary.json](summary.json), [native-summary.json](native-summary.json), and
[native-runs.json](native-runs.json). Raw Mac samples and the temporary create
probe are not retained.
