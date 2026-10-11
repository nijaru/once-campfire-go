# Publication authorization materialization — 2026-10-11

**Rejected; production unchanged.** Packing the live authorization result into a
single SQLite JSON object did not produce a meaningful CPU reduction. The two
current profiles remain useful ordinary-operation evidence; neither they nor the
candidate profiles diagnose or repair the historical low-throughput observation.

The native desktop was unavailable (SSH timed out). These are Linux ARM64 profiles
in the Mac's Docker VM, not native capacity measurements. Application base
`3039e71` has production code identical to `ccfa838`; shared verification is
`ec02deb`. The candidate changed only `DB.AuthorizedSessions` in a private copy.

## Workload and capture

The existing complete Cable workload uses 64 distinct authenticated people, six
subscriptions each, 100 paced messages at 100 ms and four saturated posters for
five seconds. It retains presence refreshes, complete HTTP consumption, all-client
draining and exact acknowledged message identities. Go 1.27.1, CGO, FTS5,
`-trimpath`, `-buildvcs=false`, four runtime workers, three job workers and
`GOGC=100` match both arms. Server CPUs are 0–3, client CPUs 4–7 and runner CPUs
8–9; affinity does not isolate VM placement or background activity.

Fixture setup, logins and stream scraping run with profiling disabled. The server
then restarts on the same disposable database with `GO_CPU_PROFILE` enabled;
existing signed cookies/streams remain usable. Graceful shutdown flushes the
profile. Captures still include startup/readiness, WebSocket connection setup,
paced and saturated posting, refreshes, disconnect/shutdown and idle audit time.
They are not isolated saturated-phase profiles. CPU/message below divides total
sampled server CPU by all audited paced/saturated messages; fixed overhead is not
subtracted. SQL audits run in the runner after the workload, outside the server
process.

Budget: two current captures to identify a cost, then at most two equivalent
candidate captures. Order was **current, current, candidate, candidate**, not an
interleaved timing comparison. No task-owned build or test overlapped captures.
The common client records minimal per-request timestamps and ACK IDs; client
probe overhead and CPU-profiler overhead are not isolated.

## Cost and candidate

Current samples place `AuthorizedSessions` at **1.50/1.36 s**, about 12–13% of
sampled server CPU. Its `Rows.Next` subtree accounts for 0.95 s in the first
capture, much of it SQLite stepping through CGO. This identifies an owner, not a
measurement of boundary-crossing overhead separately from SQLite execution.
Socket writes are larger: the prepared-frame writer accounts for 4.96/4.70 s,
about 41–42%. Those observations do not reopen the rejected contiguous-frame or
cancellation-bridge controls from the same premise.

Hypothesis: returning `json_group_object(s.token,s.user_id)` and decoding one
`map[string]int64` avoids enough repeated Go/CGO result materialization to reduce
complete-workload CPU. The query retains its live session/active-user predicates,
room membership predicate, stable `json_each` input, single-statement snapshot and
cancellable SQL. It adds no authority cache, schema change or queue budget.

Before candidate capture, the criterion was at least **20% lower authorization
CPU/message** and **3% lower total CPU/message**, followed by a small uninstrumented
complete-delivery control only if the profile reduction earned it.

| Capture | Audited messages | Sampled CPU s | Authorization CPU s | Total CPU ms/message | Authorization CPU ms/message | Instrumented delivered/s | Saturated POST p99 ms |
|---|---:|---:|---:|---:|---:|---:|---:|
| Current 1 | 11,694 | 11.73 | 1.50 | 1.0031 | 0.1283 | 2318.1 | 6.695 |
| Current 2 | 11,473 | 11.55 | 1.36 | 1.0067 | 0.1185 | 2273.7 | 7.123 |
| Candidate 1 | 11,723 | 11.35 | 1.42 | 0.9682 | 0.1211 | 2322.1 | 7.131 |
| Candidate 2 | 10,957 | 11.65 | 1.24 | 1.0632 | 0.1132 | 2170.5 | 7.679 |

Medians of the per-run normalized CPU figures give **−5.07% authorization CPU**
and **+1.08% total CPU**. Authorization barely moved; total CPU did not show a
consistent reduction. These are coarse sampled observations from sequential VM
captures, not precise causal effect estimates. The lower-rate candidate capture
is retained; its cause is not assigned to host noise or the SQL change.

The profile gate failed. The candidate is rejected and its budget is closed;
there was no uninstrumented follow-up or production implementation/test change.
Fewer returned rows alone is not sufficient justification. A new candidate needs
more specific evidence about query access paths, SQLite execution or another
substantial owner, rather than another result-serialization rearrangement from
this premise.

## Correctness and limits

All **45,847** paced/saturated writes passed the unchanged shared Cable validator
and independent exact ACK-ID/body/creator/room/unique-marker/FTS/population audits.
Every acknowledged message reached every client; there were no failed POSTs,
missing deliveries or new FK violations. Integrity was `ok`, the seed's one FK
orphan was unchanged and presence returned to zero.

The private candidate passed `bin/check` on macOS and Linux ARM64 vet/database,
application and Cable race checks. Existing session-revocation, active-status and
membership tests passed unchanged. A temporary compatibility probe for empty and
duplicate inputs, escaped token keys, integer IDs above 2^53 and an already-
cancelled context followed by a successful read passed against both original and
candidate source. It is retained as
experiment evidence, not added to the production suite. No browser/media rerun was
needed or performed after rejection; existing production evidence is unchanged.
There is no new general-capacity, correctness-completeness or slowdown-fix claim.

## Retained evidence

[manifest.json](manifest.json) records source-input and artifact hashes. The
[artifacts](artifacts/) contain four raw CPU profiles, flat/cumulative listings,
raw client records with exact ACK IDs/timestamps, audited summaries, metadata,
capture scripts, the rejected query patch and private gate/probe logs. Binary
SHA-256 values and matching build settings are in the metadata; binaries and
fixture databases are not committed. The scripts reuse the existing private
fixture/helper paths shown in the [HTTP-boundary record](../http-boundary-room-post-20261011/README.md),
not a new benchmark harness.

From this directory, function-level CPU totals can be inspected without a binary:

```sh
go tool pprof -top artifacts/current-1.cpu.pprof
go tool pprof -top -cum artifacts/candidate-1.cpu.pprof
```

Source/assembly inspection requires the matching recorded executable. Go profiles
attribute native SQLite work to CGO frames rather than separating the SQLite
interpreter, allocation, synchronization and boundary costs.
