# Skip unused write-request cache observations — local, 2026-10-12

**Historical local experiment: landed at `513485a`, then reverted after
[native non-regression failures](../post-cache-observation-native-20261012/README.md).**
The measured guard avoids acquiring an unused write-request cache generation.
The guard preserves every GET/HEAD observation, fresh session authentication,
writer-current permission check, cache budget and postcommit obligation. Mutation
query scopes retain generation zero, which disables persisted-fragment reuse and
admission; receipt rendering already bypasses that cache.

Three matched, server-uninstrumented local pairs show **3.0% lower median process
CPU per completed write**, with the same direction in every pair. POST p99 is mixed
and its median is **1.3% worse**; no latency or capacity improvement is claimed.
This is local CPU/work evidence, **not validation of the native POST p99 gap** or a
new upstream/Rust ranking. Subsequent native validation rejects the guard.

## Sources and mechanism

The application baseline is `8351224`; later `737625f` changes documentation only.
The measured candidate is that application source plus the exact guard patch in
[`manifest.json`](manifest.json). The prototype and initially integrated
`internal/web/response_cache.go` at `513485a` match byte-for-byte. Profiles use the production
`GO_CPU_PROFILE` hook, without a private hook or runtime tuning.

`beginResponseCache` used to call `DB.ResponseVersion` before ruling out non-GET/HEAD
methods. Every POST therefore submitted an observer request and waited for its
SQL observation despite being ineligible for private-response reuse. Its generation
was also unused by created-receipt rendering. The change moves the existing method
guard before that call and removes the redundant later guard. It neither caches nor
skips an authorization query.

Authentication and route preparation occur before a server restart, so bcrypt and
setup are excluded from the posting profiles and process-CPU controls. Stable
fixture secrets and persisted sessions preserve the posting cookie across restart.
The profiles include startup, profiled warmup, receipt-audit idle intervals and
joined shutdown; counters include all process CPU after that restart, not only the
last timed interval.

## Initial attribution and rejected query shape

One local current/upstream pair posts 66,000 messages per arm, at 3,000 offered/s,
two warmup seconds and twenty timed seconds. Fork sampled CPU is 24.18 seconds
versus upstream's 28.52; this does not explain a native tail gap. Message-display
hydration accounts for 1.69 seconds and the version observer for 0.54 seconds of the
fork capture. Percentages alone are not the discriminator.

A private direct-JSON-join candidate replaces display-hydration `IN` subqueries
with joins to the same unique ID sets. SQLite 3.53.4 plans eliminate list subqueries;
the isolated probe is faster. But one complete application pair gives display CPU
**1.77 versus 1.65 seconds**, total **24.75 versus 24.60 seconds**, and effectively
flat/slightly worse process CPU. It fails the declared display/whole-application
gate and is rejected. No SQL changes or extra production paths are promoted.
Those 132,000 writes pass exact audits and remain separate evidence.

## Guard profile and bounded control

One guard/current capture pair has **66,000 audited writes per arm**:

| Local profiled work | Baseline | Guard | Change |
|---|---:|---:|---:|
| Total sampled CPU, s | 23.96 | 23.55 | −1.7% |
| OS process CPU, s | 28.123 | 27.519 | −2.1% |
| Version-observer sampled CPU, s | 0.51 | absent | unused work removed |

The initial **2% sampled-CPU target is missed**; it is not recorded as a pass.
Independent OS counters nevertheless show a reduction and the profiled mechanism
is gone. Before the planned uninstrumented controls, the discriminator is refined
to actual process CPU per completed write: three interleaved pairs, require at least
2% median reduction, no more than 5% median POST p99 regression and no invalid or
missing offered write. No more profiling pair or repeat-to-win sweep is used.

Controls use the existing fixed POST workload: **3,000 offered/s**, **64 workers**,
active pacing, gzip, **two warmup/eight timed seconds**, four runtime workers/readers,
three job workers and default `GOGC=100`. Order is baseline/guard, guard/baseline,
baseline/guard. Each fresh arm posts 30,000 exact-audited writes.

| Local control | Baseline rounds 1 / 2 / 3 | Guard rounds 1 / 2 / 3 |
|---|---|---|
| Process CPU, µs/write | 416.387 / 417.066 / 416.963 | 407.604 / 402.615 / 404.418 |
| Timed POST p99, ms | 1.590 / 1.590 / 1.592 | 1.574 / 1.610 / 1.678 |

Median process CPU is **416.963 versus 404.418 µs/write** (−3.0%). Median p99 is
**1.590 versus 1.610 ms** (+1.3%). The third guard p99 is 5.4% worse individually;
it is retained, not hidden behind the CPU result. All fixed arrivals are scheduled,
complete with status 200 and pass decoded Turbo/ACK/persistence checks.

The controls pass the refined, declared CPU/non-regression discriminator, not the
original sampled-CPU threshold. The landing claim is less process work plus a
simpler write-request path—not lower tails, higher throughput or native parity.

## Environment and limits

Native Mac **darwin/arm64**, Go **1.27.1**, CGO/FTS5 and SQLite **3.53.4**. No CPU
affinity or exclusive host isolation is available here. Unrelated applications and
the pre-existing OrbStack container are left untouched. No task-owned builds,
tests, agents or profilers overlap the uninstrumented controls. Request logging
and all production defaults remain active.

The seed SHA-256 is
`691c52bfdc30b38bd8867e93b0914613df20e03d4cd7b00796ab68e0a5d26c9a`;
shared verification is `ec02debb3e44f40b8778a6a42db3c4515b2f9eaf`.
The common private offered-rate client validates decoded Turbo responses and retains
exact ACK-ID/sequence records. The shared seed-grounded preflight checks room,
pagination, sidebar, search, avatar, CSS, health and write-room response contracts
before profiling/timing.

Every run independently checks exact persisted populations, unique client identities,
ACK IDs, body sequences, creator, room and FTS body; SQLite integrity passes and the
seed's one FK orphan remains unchanged. There are **312,000 writes** for the guard
profile/control cohorts, **132,000** for initial attribution and **132,000** for the
rejected display joins: **576,000 distinct audited writes** total. These are POST
workloads without live viewers, not a fresh fanout/capacity comparison.

Earlier contended local/VM timing is not used for this decision or made an unfinished
reproduction goal. Likewise, this local result does not update the
[native `8351224`/upstream comparison](../current-upstream-native-20261011/README.md).
It needs matched desktop controls before any native performance claim.

## Correctness and review

- Integrated `bin/check` passes formatting, regenerated assets, vet, all races and
  the local WebSocket fork's suite.
- Affected cache/query races pass in the prototype, including GET/HEAD handling,
  foreign changes, fragments with response caching disabled, and mutation/auth checks.
- Existing Go/Rust HTTP flows `message_lifecycle`, `pagination`, `room_membership`
  and `authentication` pass in the local Linux/arm64 container. Go binary and tool
  identities are retained in the report. Cached Rust binary
  `250dfeae44c9bdbf478c15ae6d8cc4e55222d81c862162bec18c27ca7ed7f3dd`
  was built at `9872c1d`; comparison with pinned `6dae2fd` shows only README changes,
  no application-source difference.
- Independent read-only review finds no concrete correctness, compatibility or
  authorization defect: method overrides cannot produce GET/HEAD, fragment generation
  zero disables reuse, receipt rendering bypasses fragments, and all live permission
  checks remain intact.

No production test scaffolding or new compatibility behavior is introduced.
Existing observable cache/mutation/foreign-write tests protect the affected
contracts; a test of the observer's private call count would add no independent
protection. No broad media/browser rerun is claimed for this server-side guard.

Three harness errors are retained: optional CSRF-token handling failed before the
first app profile; a renamed result-print key failed after a valid control, whose
unchanged audited row is reused; parity output-directory setup failed before app
requests. Each is corrected without weakening response or persistence assertions.
They are not product failures or extra successful workload counts.

## Retained evidence and next decision

The manifest hashes the guard/rejected patches, exact adapters, profiles, per-phase
raw ACK records, SQL probe, OS CPU/latency samples, gates and compatibility reports.
Runtime secrets, cookies, databases and application request logs are excluded.

The local counters remain valid for these samples; both the direct-join premise and
the guard are now closed. Native fixed POST median p99 +6.7% and paced all-client
Cable p99 +5.8% miss the declared non-regression limits; the guard is reverted.
The [native record](../post-cache-observation-native-20261012/README.md) owns those
controls, failures and source identities. No historical slowdown or native tail gap
is claimed repaired, and no PR is opened or prepared.
