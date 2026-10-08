# Rejected background-checkpoint prototype

Parent: `cf7afd6`. No checkpoint/dependency changes were retained in production.

The pinned Rust database already replaces SQLite's default writer-side automatic
checkpoint with a WAL commit callback: background PASSIVE after each additional
1,000 frames, coordinated writer RESTART at 10,000 frames. Its source documents
larger transient WAL size, lower writer stalls, and unchanged WAL/NORMAL policy.
Go still uses the native 1,000-page automatic checkpoint. See
`reference/crates/db/src/database.rs`, especially the introduction and lines 337–411.

The public Go driver has no WAL-hook API. This private prototype added a narrow
hook and native checkpoint method to an isolated copy of go-sqlite3 v1.14.52
(SQLite 3.53.4), not the production dependency or module cache. A per-database
connector configured every writer connection; one owned checkpoint connection,
worker and capacity-one wake channel handled PASSIVE. A shared mutex excluded
background checkpoints during writer RESTART. Shutdown canceled/joined the worker.
The callback returned success even on maintenance failure, since SQLite has already
committed. Request transactions and cancellation remained on database/sql.

Unlike Rust's external writer owner, the prototype ran native RESTART inside the
WAL callback, before database/sql released the physical connection. This avoids a
writer-wrapper/caller migration. SQLite explicitly permits checkpointing from the
WAL callback after the write lock is released. It must use the native API, not
reenter SQL through database/sql's locked driver. The 10,000-frame setting is a
restart threshold, not a strict cap under held readers, external writers or large
transactions; RESTART does not truncate physical WAL allocation.

Primary sources checked 2026-10-08:
- https://sqlite.org/c3ref/wal_hook.html
- https://sqlite.org/wal.html
- The installed SQLite 3.53.4 C API declarations and driver callback lifecycle.

Linux vet/full race tests pass with the proper pinned reference mount. A temporary
real-write probe crosses both thresholds and verifies copied frames and all 180
rows; five race repetitions pass. The first test attempt had missing reference
fixtures. The first benchmark attempt completed only its baseline, then rejected
the prototype's source identity because its fixture symlink was not a valid Git
submodule; `comparison.log.gz` preserves that setup failure. Neither attempt enters
the comparator. Tests/builds did not run during timing.

Three alternating shared-harness repetitions, public listener, eight timed seconds
plus excluded two-second warmup, gzip/sixteen clients, identical seed, server guest
CPUs 0–3/client 4–7:

| Route | Before req/s median | Prototype req/s median | Change | Before/prototype p99 ms |
|---|---:|---:|---:|---:|
| Room control | 13,692 | 13,594 | −0.7% | 5.335 / 5.351 |
| Sidebar control | 15,184 | 14,999 | −1.2% | 5.107 / 5.107 |
| Post | 2,749 | 2,828 | +2.9% | 22.399 / 22.271 |

One post pair regresses 2,759→2,566; the other two improve about 3%. All sidebar
pairs regress. All measured responses and exact acknowledged ID/body/room/FTS
writes pass; raw ranges, audits and source identities are retained. The small,
inconsistent write gain and unchanged tails do not justify a driver fork and
another concurrency/resource subsystem on the available ARM64 VM. The source
remains an experiment, not a supported application option. Native desktop results
are unavailable; no claim is made about other storage devices.

The earlier 16% cumulative commit CPU attribution did not isolate checkpoint
work. This experiment does not establish that checkpointing explains the Go/Rust
gap. Keep that limitation rather than treating a profile label as a promised win.
