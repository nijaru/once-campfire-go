# Stream-publication profiling

These are CPU diagnostics, not the final pre-PR benchmark. They use the existing
production `GO_CPU_PROFILE` hook and shared load generators. Containers stop
gracefully before profiles are copied. No builds, tests or agents overlap profiling.
Go 1.27.1, four runtime workers/readers, three job workers, default GOGC 100,
server CPUs 0–3 and client CPUs 4–7 match the preceding cohorts.

## POST diagnosis

Two forward/reverse pairs hold arrivals at 3,000 POSTs/s: two warmup seconds and
20 timed seconds, one viewer and a 64-request concurrency bound. Upstream is
`c4ab53d`; the candidate is presence-fixed `d7e8382`, before stream batching.
All 264,000 writes, including warmup, pass exact persisted body/identity/creator/
room/FTS and integrity audits; 240,000 timed responses complete without errors or
lost arrivals. Timed responses are transport/status/nonempty-body checked, not
semantic acknowledgment-ID validated. Shared route contracts run before timing.

| CPU samples, including preparation and shutdown (s) | Upstream | Candidate |
|---|---:|---:|
| Run 1 | 20.90 | 19.51 |
| Run 2 | 20.72 | 19.36 |
| Background marking, run 1 | 0.66 | 0.94 |
| Background marking, run 2 | 0.67 | 0.72 |

The candidate uses less total sampled CPU but retains higher POST service tails:
p99 5.640/5.800 ms versus upstream 4.656/4.888 ms. GC work is somewhat higher;
these profiles do not establish that it causes the tail difference. CPU sampling
is not a waiting-time or pause-distribution measurement. Generator lateness remains
substantial. GC defaults are unchanged; no heap padding or memory-budget increase
is used to hide the difference.

## Cable hotspot and change

One diagnostic before and one after use 64 independently authenticated people,
30 paced messages at 200 ms intervals, then four posters for five seconds. Each
person has their own page-derived subscriptions. The shared Cable checker verifies
complete delivery. Exact unique message markers, room, FTS, integrity and zero
presence after disconnect are audited: 8,959 writes before and 13,828 after.

The original unread loop creates a separate publication per user. Each repeats
session authorization, payload encoding and frame preparation. Batching one event's
selected streams preserves fresh authorization and subscription identifiers while
sharing frames with identical identifiers. Closed-room sidebar events with the
same payload also batch; direct-room viewer-specific payloads remain separate.
There is no authorization cache or change to queue limits.

| Cumulative sampled CPU (s) | Before | After |
|---|---:|---:|
| `Hub.publish` | 5.56 | 2.16 |
| `AuthorizedSessions` | 4.59 | 1.63 |
| `MessageNotifications.Created` | 5.34 | 1.77 |

These functions use less sampled CPU despite more completed writes. Whole profiles
include fixture setup: roughly half their CPU is the 65 real bcrypt logins. The
closed-loop populations differ, so neither total CPU per write nor the observed
1,784.8→2,757.5 messages/s establishes a general capacity gain. Paced all-client p99
is 17.743→12.199 ms in these single profiled samples; this is not a reliable final
tail comparison. Formal comparisons belong before the PR, not after every tweak.

## Verification and provenance

The focused real-WebSocket test protects duplicate-stream suppression, distinct-user
routing and fresh session revocation without timing-based no-delivery assertions.
Existing alias, membership-revocation, compression, backpressure and presence tests
remain. Native races repeat the Cable/application suites ten times; native
`bin/check`, Linux vet/affected races/build and the shared Chromium flow pass.
A separate read-only review finds no actionable regression; it does not independently
reproduce these executed checks.

Before source is `d7e8382`. After source is that revision plus `candidate.patch`
(SHA-256 `ef93ed052cc660851ad268677188e9087571c8c0ddb4c7f5a07f0027f34814fa`).
Binary hashes, image/seed/loadgen metadata, raw profiles, diagnostic results and
orchestration adapters are retained. Upstream and shared-verification main revisions
were refreshed and unchanged. Cookies, session files, secrets and runtime databases
are excluded. The shared ARM64 OrbStack/macOS host, guest affinity, tmpfs storage
and loopback provider destinations retain the preceding environmental limits.
