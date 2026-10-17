# Writer statements without the driver goroutine — local, 2026-10-17

**A CPU/allocation reduction on message creation; no native latency or capacity
claim.** `mattn/go-sqlite3` takes a goroutine-plus-select path for every statement
when the context has a Done channel. A current-head execution trace showed the
slowest one percent of POSTs waiting in `create.unread`: 7.7 ms of waiting against
2.6 us of work, with `writer.acquire` at 400 us. The delay was the spawned statement
goroutine, not SQLite or room work.

The candidate runs the short `createMessage` statements with a non-cancelable
context and checks the caller's context between statements and before commit.
`database/sql` still rolls a canceled request back; the regression asserts that
nothing commits. Other commands keep their existing statement cancellation.

## Local Mac complete-workload pair

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

CPU per timed write: 499.4/438.7 us baseline, 468.0/425.5 us candidate.

A local create-only benchmark (three 1,000-iteration runs) measured
**92.3 -> 88.5 us/op (-4.1%)**, **5.34 -> 4.17 KB/op**, and
**122 -> 100 allocs/op**.

## Limits and verification

The Mac was shared with other work and had no CPU affinity. The first p99 samples
were contaminated (baseline 53.6 ms, candidate 364.5 ms); the clean later pair was
1.752 vs 1.708 ms. **No p99, capacity, or native Fedora claim is made.** A matched
native pair remains pending.

`bin/check` passes. Local shared Go/Rust `message_lifecycle` and `pagination`
workflows pass 4/4. Raw per-run samples and the temporary create probe are not
retained; the compact numbers are in [summary.json](summary.json).
