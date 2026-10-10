# Fixed-rate HTTP and distinct-user Cable

Application source matches the measured materialization candidate; rebased commit
`e756cff` changes only history and evidence retention. Upstream is c4ab53d. The same
production binaries, four workers/readers, server CPUs 0–3, client CPUs 4–7, gzip,
VM-host tmpfs and loopback provider destinations are used as in the preceding
capacity cohort. Three sequential forward/reverse rounds keep the candidate in
the middle. No builds, tests, profilers or other agents run during timing.

## Fixed offered rates

The existing `bench/httprate` schedules arrivals independently of completions.
Reported scheduled latency includes client queueing and generator lateness; service
latency is measured separately. Each route uses two excluded warmup seconds and
eight timed seconds, one viewer, a 64-request concurrency bound, 10,000 reads/s or
3,000 POSTs/s. These loads are below all three builds' measured capacity.

| Scheduled p99, median of three runs (ms) | Upstream | Candidate | Rust |
|---|---:|---:|---:|
| Room | 5.792 | 5.704 | 6.112 |
| Messages | 5.664 | 5.536 | 5.840 |
| Sidebar | 5.984 | 5.944 | 6.032 |
| Search | 5.304 | 5.456 | 6.272 |
| POST | 7.368 | 8.080 | 5.416 |

Candidate POST p99 is higher in all three pairs: 7.768/8.080/8.880 ms versus
7.272/7.368/7.424 ms upstream. Its median service p99 is 6.536 ms versus 5.528 ms.
Generator lateness p99 is substantial: 4.256/4.160 ms for candidate/upstream POST,
and roughly 3.6–3.8 ms for reads. Do not subtract quantiles to infer application
latency, treat these measurements as isolated server time, or claim a tiny read
p99 difference is a reliable improvement. Candidate median POST p50 is 1.416 ms,
versus upstream's 1.178 ms; the tail difference warrants further investigation.

The population completes 3,096,000 timed and 774,000 warmup responses with no lost
arrivals, dropped/expired requests, transport errors or unsuccessful statuses.
All 270,000 writes, including warmup, are audited for the complete body sequence,
unique client identity, creator, room and whitespace-normalized FTS content; database
integrity passes. Peak generator CPU is 51.1% of the available 400%.

Shared seed-grounded route contracts run before timing and four shared per-response
checks follow each read sample. The timed rate generator consumes complete wire
bodies and checks transport/status/nonempty responses, **not semantic content or
returned acknowledgment IDs**. Its success counts must not be described as millions
of shared-contract-validated responses. Independent persisted-state audits and the
preceding exact HTTP acknowledgment contracts complement, not erase, this limit.

An initial incomplete run stopped on the temporary audit script's reserved SQL
alias `indexed`; it is not included in these paired results. The corrected run uses
fresh outputs. No application behavior or oracle was changed to accommodate it.

## Sixty-four distinct people

Each sample adds 64 active member users and room memberships to the same seed.
Their sessions are created by real login requests; SQL independently verifies 64
distinct authenticated user IDs. Each user's own page supplies its three signed
stream subscriptions, plus presence, unread rooms and heartbeat. The shared
load generator uses one session per client, 50-second staggered presence refreshes,
30 paced messages at 200 ms intervals, then four closed-loop posters for five
seconds. WebSocket compression is not negotiated in this population.

| Median of three runs | Upstream | Candidate | Rust |
|---|---:|---:|---:|
| Complete all-client messages/s | 1,641.0 | 1,809.2 | 3,066.4 |
| Paced all-client p99 (ms) | 14.111 | 15.735 | 11.223 |

All 98,343 marked writes are uniquely persisted in the correct room and indexed;
every posted message reaches all 64 clients. All subscriptions confirm without
failures, unread notices are observed, and presence counters return to zero after
clients disconnect. Candidate throughput improves in all three upstream pairs.
Its paced all-client p99 is higher in each pair, with samples 24.655/15.207/15.735
ms versus 14.111/14.631/12.831 ms. Only 30 paced messages per run make these tail
estimates noisy; retain the outlying first sample rather than presenting it as
solved or discarding it. Peak generator CPU is 64.1% of 400%.

Fixture preparation initially hit the real per-IP login limiter. The completed
cohort logs fixture users in from distinct loopback addresses; the limiter remains
enabled. The failed setup has no timed sample and is excluded from the population.

Phase RSS/HWM/thread snapshots are retained. Candidate sampled RSS after saturated
drain is 132–133 MiB, upstream 123–136 MiB and Rust 135–146 MiB. These are phase
snapshots, not peak aggregate memory, arbitrary-user scaling or deployment bounds.
Normal presence release is observed; queue, cancellation and shutdown contracts
remain protected by the existing race suite, not proven globally by this load run.

## Scope

These cohorts do not establish PR readiness. POST and paced-fanout tail differences
remain visible. Affinity does not isolate physical cores on the shared ARM64
OrbStack/macOS host; tmpfs audits do not measure NVMe or crash durability. Loopback
push destinations bypass Go provider encryption/network work. Sixty-four people
are not an unlimited-user bound. Source/image/binary hashes, settings, sanitized
raw results and local orchestration adapters are retained; no cookies, session
files, environment secrets or runtime databases are included.
