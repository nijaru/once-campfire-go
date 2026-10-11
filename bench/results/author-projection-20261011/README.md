# Author projection on display misses: rejected

Production remains unchanged. Removing the redundant creator-name JOIN on display
misses is mechanically sound, but this comparison does not establish a meaningful
complete-workload gain sufficient to justify the extra projection/scanner path.
The candidate is reverted; POST tails and Cable collapse are not resolved.

## Candidate and contracts

The [exact private patch](candidate.patch.gz) makes display-cache misses hydrate
message bodies without joining users for names. The existing same-snapshot display
batch still supplies authors, biographies and versions. Full `Records` callers keep
creator names. Reference scope, body/version observation, ordering, generation
checks, release before presentation, cancellation and missing-author placeholders
remain intact. The scanner also pre-sizes its record slice; measurements cannot
isolate JOIN removal from that allocation change.

Baseline production is `ccfa838` (documentation HEAD `3f6bf02`); candidate sources
are that baseline plus the patch. Native binaries use Go 1.27.1, CGO/FTS5 and the
accepted Linux/libvips image. [Hashes](binary-hashes.txt.gz), image identities,
fixture/load-generator hashes, settings and orders are in the metadata. Neither
rejected wire nor cancellation-bridge changes are included.

## Complete HTTP comparison

Three alternating pairs use eight clients, two warmup seconds and eight timed
seconds, gzip, GOGC=100 and matched data. Server CPU affinity is 0,2,4,6; client
8,10,12,14; driver 16,18. No task-owned agents, builds, tests or profilers run during
timing. Affinity is not host isolation; thermals, SMT siblings and other activity
remain incompletely controlled. No adverse sample is discarded or blamed on them.

Changed workloads disable response and fragment caches: 40-message pagination,
seeded dense `coffee` search and the existing 100-accessible-message sparse search
fixture. The warm control repeats pagination with the ordinary 32 MB budgets; it
is not general application capacity. Independent SQL IDs, complete response bodies,
DOM identities and message content are validated using the shared contracts.

| Workload | Throughput change, pairs 1 / 2 / 3 | p99 change, pairs 1 / 2 / 3 |
|---|---|---|
| Uncached page40 | +2.25% / +25.22% / +15.16% | −4.81% / −34.94% / −24.09% |
| Sparse search | +1.25% / +17.62% / +0.79% | −2.07% / −38.48% / −0.92% |
| Dense search | +0.36% / +10.98% / +0.30% | **+1.14%** / −11.89% / **+1.47%** |
| Warm page40 | +4.91% / +5.58% / **−0.49%** | −5.03% / −10.17% / **+0.65%** |

The page direction is initially favorable, but the largest apparent gains depend
on slower baseline observations. The unchanged warm path also varies materially;
that is a caution about effect-size attribution, not a cold-path error bound.
Independent review found no code defect but judged meaningful benefit unresolved.

One final, longer candidate-first pair—the only inconclusive follow-up—runs page40
for two warmup and 30 timed seconds with otherwise unchanged settings. Baseline /
candidate throughput is **2832.7 / 2861.2 responses/s (+1.01%)**; p99 is
**8.067 / 8.111 ms (+0.55%, worse)**. This does not support the broad median gain
suggested by the short cohort. The modest throughput difference and flat/slightly
worse tail do not earn the additional projection path. Rejected; no further timing.

All **4,492,439** warmup/timed responses pass complete identity/content validation,
with zero request errors or invalid responses. These are read workloads, not POST
write audits. [Initial results](comparison-results.json.gz) and [final results](comparison-final-results.json.gz)
retain every sample, including adverse baselines and warm/dense losses.

## Correctness gates and setup limitations

The exact candidate passes local `bin/check`, native Linux affected database,
application and web races three times, vet and builds. Expanded existing snapshot
coverage checks both projections, full-record creator names and forbidden
references. All eight applicable Go/Rust HTTP flows pass: message lifecycle,
pagination, room membership and authentication. The unmodified shared Chromium
smoke flow passes, including messaging, search, profile updates and direct pings.
No permanent tests are added for this unlanded candidate.

The first adapter used a full-search document contract for a pagination fragment
and stopped during warmup. It was corrected to the shared `messages_page` kind;
no assertion was weakened. That setup failure is retained, but its discarded
in-memory sample is unavailable and not counted as valid comparison evidence.
Desktop SSH was initially unavailable; a successful Tailscale connectivity probe
restored the path without host configuration or dependency changes.

The [manifest](manifest.json) covers sanitized measurements, adapters, source patch
and gate logs. Cookies, environment values, signed paths, private fixtures,
databases, runtime/server logs and binaries are excluded. Media/provider/forced-exit
checks were not rerun for this read-only projection candidate.
