# Final verification and comparisons, 2026-10-10

Application source: `32a08ecc0953528bfee7cef16b1354843a8228b0`. Go upstream:
`c4ab53d`; Rust source: `6dae2fd`; shared verification: `ec02deb`.
[Provenance](provenance.json) records binary hashes, source trees and build settings.
The only uncommitted change during measurements was README restore-path guidance.

## Correctness and operational checks

- Native `bin/check`: formatting, generated assets, vet, full races and local
  WebSocket fork tests passed. Final Linux vet, full races and fork races passed.
- All **15 functional workflows passed for both Go and Rust**, including messages,
  membership, direct participants/pings, account settings, uploads, downloads and
  avatar replacement. The [report](contracts/functional-report.json) retains the
  request/status traces, with signed URL tokens redacted. These are HTTP/state
  contracts, not exact HTML or JavaScript parity.
- The **unmodified shared Chromium smoke flow passed** on a clean `ec02deb`
  worktree. Earlier browser evidence used a local persisted-echo synchronization;
  neither that modification nor altered assertions were used for this final run.
- Shared harness Rust/Ruby self-tests passed. Its sidebar reload and editor
  preservation tests passed against the served Go controller override. Initial
  probes mistakenly targeted the unserved Rails vendor source; those failures
  were not application failures. The override matches current Rust's controller.
- Dedicated Turbo response cancellation checks passed; unrelated delegate errors
  still propagate. Local Pebble ACME challenge, certificate issuance, HTTPS and
  cached restart with the CA offline passed.
- The production image built and passed non-root startup, public-listener setup,
  live SQLite backup, offline ONCE restore and application restart checks.
- Linux upgrade checks passed: Rust-issued login accepted by Go, Go-issued login
  accepted by Rust, and a Go-written message/FTS entry rendered and searched by
  Rust. The original seed remained unchanged. The unchanged upgrade tool ran
  through the workspace adapter described in provenance.
- The public listener passed the existing response contracts, exact acknowledged
  write/FTS audits, involvement transitions and settings restoration. This small
  probe was verification, not a capacity measurement.

The functional run exposed missing record validators on streamed avatar/logo
responses. `32a08ec` restores record-version ETags before choosing uploaded,
stock or initials representations. The focused regression fails the original
production code, then passes ten race-enabled repetitions after the fix. It
covers changed records, attachment removal and conditional GET/HEAD responses.
The initial asset check also needed an explicit seed path; the corrected check
passed before final functional verification.

Two independent source-only reviews found no additional actionable regressions
in authorization/receipts, scoped queries, foreign-write cache invalidation,
route parameters, provider ownership, webhook admission, staging/purge ownership
or runtime teardown. They reviewed `71f7107`; the small media fix was subsequently
reviewed and exercised by the parent. Static review is not executed coverage.

## Final matched comparisons

Production candidate image versus the previously built Go upstream and Rust
images. Sequential, interleaved forward/reverse/forward rounds; ARM64 OrbStack;
server CPUs 0–3, client CPUs 4–7; default GOGC 100; four Go processors/readers and
three job workers. Fixtures were made writable by UID 1000 before startup so the
non-root production image could use the same mounted storage. No task-owned
builds, tests or profilers ran alongside timing.

| Workload, median of three runs | Go upstream | Candidate | Rust |
| --- | ---: | ---: | ---: |
| Contract-validated POST, req/s | 2569.5 | 2592.2 | 2778.6 |
| Fixed 3000 POST/s, scheduled p99 ms | 6.912 | 8.416 | 4.928 |
| Fixed 3000 POST/s, service p99 ms | 5.512 | 7.024 | 3.048 |
| 64-person Cable, delivered messages/s | 1383.6 | 2292.6 | 2684.7 |
| Cable paced all-client p99 ms | 11.919 | 9.663 | 7.799 |

- **POST capacity:** sixteen clients, two-second warmup, eight-second timing.
  All **189,206 timed responses** passed shared semantic response validation;
  **235,587 writes including warmup** passed acknowledged-ID/body/FTS/integrity
  audits. Candidate median was +0.9%; it led two pairs and trailed one. This is
  not convincing evidence of a general POST capacity improvement.
- **Fixed POST:** bound 64, two-second warmup, eight-second timing. All **216,000
  timed responses** and **270,000 writes including warmup** completed without
  status/transport errors, loss, dropped or expired requests. Exact persisted
  bodies, unique client identities, creator/room and normalized FTS were audited.
  Timed `httprate` checks full wire bodies/status/nonempty, not semantic response
  content or HTTP acknowledgment IDs. Candidate scheduled p99 samples were
  **8.144/8.416/125.696 ms**, versus **6.912/6.304/25.696 ms** upstream. The large
  third-round outliers and candidate queue peak of 365 are retained, not averaged
  away. Generator lateness is included in scheduled latency; quantiles cannot be
  subtracted to isolate it. The POST tail disadvantage remains.
- **Cable:** 64 distinct people and authenticated sessions, six page-derived
  subscriptions each; 100 paced messages at 100 ms intervals, then four posters
  for five seconds. Every client confirmed subscriptions and received every
  expected message; **88,218 writes** passed unique-body/FTS-presence/integrity
  audits, and persisted presence was zero after disconnect. Candidate throughput
  improved all three pairs (+65.7% by median), and paced p99 improved all three
  (-18.9% by median). This supports the profiled stream-publication batching
  improvement for this cohort, not unlimited-user scaling or universal tails.

POST capacity is much lower for **all three implementations** than earlier
archives. Guest load rose across the capacity rounds; unrelated native test
activity was observed after timing. Host activity and thermal state were not
controlled, and virtual CPU affinity is not physical isolation. These runs do
not establish the cause of the common slowdown or the fixed-rate outliers. Do
not compare their absolute rates directly with earlier sessions or erase the
[earlier POST deficit](../pr-readiness-20261010/README.md). The
[profile archive](../stream-publication-20261010/README.md) provides separate CPU
evidence; these comparisons are not profiles.

## Remaining limits

This is **not a claim that every upstream verification passes**. Strict HTML,
network and pixel inventories have documented differences and were not repeated
here. Full media-output parity was not rerun for the validator-only change.
Live external Web Push/provider acceptance remains unverified; loopback fixture
destinations are not provider delivery evidence. Jobs and purge continuations
remain process-owned: forced exit can lose them, and a thirty-second forced exit
is not successful teardown. Default-layout backup/restore passed; the inherited
ONCE restore hook does not honor custom storage roots/database filenames, which
require matching manual restore handling. No release, deployment or PR was made.

## Evidence

`post-capacity/`, `fixed-post/` and `distinct-fanout/` contain sanitized raw rows
and metadata. `logs/` contains gates, operational checks, regression results and
timing output. `adapters/` retains the private cohort/comparison/public-probe
scripts; they use the pinned shared helpers and local fixture workspace. The
comparison adapter adds the historical Go image and fixture ownership setup;
assertions were not weakened. Runtime databases, cookies, session files, fixture
environment secrets and write-audit JSONL files are intentionally excluded.
