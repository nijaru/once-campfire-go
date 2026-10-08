# Read-path work and shared verification — 2026-10-08

The retained candidate starts at `170149e`, the verified upstream-based integration.
It reuses route recognition within one request and gives the SQLite generation
observer one prepared statement and one resource owner. Authentication, hourly
session refresh, room access checks and foreign-commit observation remain live.
There is no TTL shortcut or cache-based authorization.

## Evidence and retained changes

A separate seven-workload profile attributed 9.36% cumulative CPU to generation
reads and 6.24% to route recognition. These percentages overlap other call stacks
and are not additive projected speedups. The profile is diagnostic, not a comparator.

`retained/warm/` contains three alternating repetitions of all five HTTP routes:
16 clients, gzip, eight measured seconds after two excluded warmup seconds, identical
seed snapshots, server guest CPUs 0–3 and client CPUs 4–7. Before/after complete
HTML matches after normalizing only the listener origin. Every acknowledged write
is checked in messages and FTS. There were zero HTTP errors.

| Route | Before median req/s | Retained median req/s | Change | Before/after CPU µs/request |
|---|---:|---:|---:|---:|
| Room | 14,378 | 15,687 | +9.1% | 71.9 / 66.0 |
| Messages | 14,804 | 16,189 | +9.4% | 70.7 / 65.0 |
| Sidebar | 16,848 | 18,075 | +7.3% | 62.1 / 57.9 |
| Search | 16,813 | 17,740 | +5.5% | 62.7 / 59.2 |
| Post | 2,705 | 2,693 | −0.4% | 562.3 / 556.2 |

Room's last candidate repetition regressed (13,191 versus 14,378); search's second
regressed (15,440 versus 16,876). Messages and sidebar improved in every pair.
Posts did not improve, and candidate post p99 was 24.159 versus 23.791 ms.
Do not extrapolate median read gains to writes or matched-offered-rate latency.

`retained/churn/` separately changes the room/search workload during reads at up to
20 committed messages/second. Room median rises 13,750→15,285 (+11.2%), search
11,563→12,218 (+5.7%); every throughput pair improves. This is the existing active
workload, not the shared harness's different-room, ten-write/second mixed profile.
Warm and changed workloads are not pooled. Final bodies and committed messages/FTS
are checked; all active writes passed. CPU includes the paced writer.

These are ARM64 Linux VM measurements, not dedicated physical-core or native Intel
results. `/proc` CPU has 10 ms granularity. Raw ranges, tails and host load remain
visible; client decoding and VM variance can limit capacity. No fresh Rails,
Elixir, Rust, TLS, Cable or media performance comparison is claimed.

The native generation microprobe improves elapsed time but increases
allocated bytes per call from about 840 to 1,033, with 19 allocations in both.
The before timings varied strongly; it is not an independent HTTP-win claim.

## Rejected authentication-query experiment

`rejected-auth/` retains source and all measurements for combining current-user and
session-activity reads. It passed package/race tests and preserved the conditional
hourly refresh, but the final warm population worsened sidebar and write medians;
posts fell 2,692→2,038 req/s with worse tails. Its changed-workload reads improved.
An isolated sidebar/post comparison was highly unstable (post rates 922–2,496
across both binaries), so it did not establish a dependable overall benefit.
The experiment, its generic row-scan extension and its new test are not production.
`SessionUser` and `RefreshSession` retain their original behavior. Reconsider only
with representative, stable evidence rather than selecting favorable read samples.

## Common verification, not a larger acceptance matrix

The shared harness is `basecamp/once-campfire-verification` at
`7b2dbc7e856fb2cbf2d95833a22110beab026eec`. Its controller regressions pass against
the served override. Its full browser flow passes on the integration and retained
read-path candidate without a Go-specific mask or changed assertion. The local
browser command now launches that flow, rather than maintaining a divergent copy.
The redundant local sidebar fixture was removed: shared controller tests cover
loading/reconnect/cleanup, and shared browser flows exercise actual input-node,
recipient and deliberate-navigation preservation. The unique unfinished-body
Turbo cancellation regression remains an explicit `--turbo-cancellation` check.

The shared flow also has a premature count assertion: after another tab sees the
third broadcast, the sender can still have an optimistic preview without
`data-message-id`. A diagnostic captured two persisted elements plus that preview;
the subsequent page dump displayed all three messages. Unmodified runs both pass
and fail at that immediate assertion. `shared-browser-wait.patch` waits for at
least three persisted elements before retaining the existing exact-three assertion.
It applies to every implementation and neither masks nor accepts duplicate/missing
messages. The corrected flow passes. The patch is prepared locally, not submitted;
`bin/check-browser.mjs` does not silently apply it. `verification.json` distinguishes
revision and modified script hashes. Raw failures and the diagnostic remain archived.

Native `bin/check` and Linux vet/race checks passed on the retained production
changes. Final checks and the corrected shared browser flow passed after rejecting
authentication changes; the rebuilt Linux binary matches the measured retained
binary's SHA-256 exactly. A focused independent source review found no additional
production defects (static review, not an exhaustive proof). The retained populations
include 4,500,110 timed HTTP responses, zero errors and 159,177 committed posts
(including warmups and paced writes), verified in messages and FTS. No media, schema, dependency or frontend production bytes changed.
This is a scoped PR candidate, not proof of perfect correctness or optimal performance.
