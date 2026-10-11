# HTTP boundaries and redundant room read — 2026-10-11

**No production change.** The six HTTP-boundary controls captured healthy Cable
phases, not the historical collapse. A separate candidate removing the redundant
room read from attachment-free browser POSTs lost all three fixed-rate p99 pairs
and was rejected. Neither result closes the POST-tail or Cable performance goals.

Current application source is `perf/campfire@86e0a1a`, byte-identical to `ccfa838`.
Shared verification remains `ec02deb`. All probes, client extensions and candidate
code were confined to detached private copies; no permanent tests were added.

## HTTP-boundary control

Hypothesis: per-request client completion, handler entry/return and connection
writes can distinguish application work, final HTTP output and client-side delay
in a slow Cable phase. Budget: three alternating current/traced-server pairs,
with the same minimal client recorder in both arms. This isolates server-probe
cost, **not** client overhead or equivalence to historical pristine failures.
There were no periodic counter/receipt-map scans or process samplers.

Linux ARM64 VM, Go 1.27.1/CGO/FTS5/`-trimpath`/`-buildvcs=false`, Rust 1.98.1,
libvips 8.16.1; server CPUs 0–3, clients 4–7, runner 8–9; `GOGC=100`.
The existing 64-distinct-person fixture has six subscriptions each, 100 paced
messages at 100 ms and four saturated posters for five seconds, with presence
refreshes and complete all-client draining. Queue and authorization policies are
unchanged. Both arms record exact acknowledged message identities.

| Round | Current delivered/s | Traced delivered/s | Current POST p99 ms | Traced POST p99 ms |
|---|---:|---:|---:|---:|
| 1 | 2727.4 | 2811.8 | 5.123 | 4.771 |
| 2 | 2770.1 | 2699.8 | 5.111 | 4.983 |
| 3 | 2769.8 | 2827.8 | 5.175 | 4.731 |

No phase fell below 1,000/s. All **83,667** warm/timed POSTs passed the unchanged
shared Cable validator and exact ACK-ID/body/creator/room/FTS/population/integrity
audits. Presence returned to zero; the seed's one FK orphan was unchanged.

All **42,010** traced warm/timed requests joined exactly once by sequence, with
nonzero completed socket writes and no write errors. Saturated mean total POST
spans were 1.410–1.477 ms: handler means 0.759–0.801 ms, socket `Write` duration
sums 0.010–0.012 ms and last-write-to-client-body means 0.364–0.372 ms. In each
run's individually selected slowest 1%, handler work accounted for 3.974–4.331 ms
of mean total spans 5.857–6.497 ms. These are correlated request means, **not**
subtracted marginal percentiles or additive p99s.

Final HTTP output was not the dominant delay in these healthy phases. This does
not explain an uncaptured bad phase or establish client/scheduler/host causality.
The probe wraps `net.Conn`; its complete-workload effect and the common client's
recording work remain measurement limitations. Cross-process realtime clocks
share one VM; signed negative ordering residuals are retained. Affinity is not
isolation. There is no overhead-neutrality or general-capacity claim.

## Attachment-free POST candidate: rejected

The browser currently reads `DB.Room` before its creation command independently
checks writer-current room access and selects receipt inputs. The candidate
removed only the attachment-free preliminary read. Uploads retained their early
access check before staging IO; the transaction, FTS, notifications, publication,
queues and cancellation stayed unchanged. Command access denial used the existing
200 room-unavailable composer response, including revocation detected by the
writer that previously returned 403. This behavior change was not promoted.

Hypothesis: removing one redundant read-pool lookup improves fixed POST tails.
Budget: three alternating 3,000/s pairs; require at least two wins and ≥5% median
p99 improvement before capacity testing. Fresh current/candidate binaries used
Go 1.27.1, native Fedora x86_64 and the existing isolated toolchain. Server CPUs
0,2,4,6; clients 8,10,12,14; runner 16,18; gzip, four runtime workers, three job
workers and `GOGC=100`. Two warmup seconds and eight timed seconds per arm, 64
client workers, unchanged active pacing and 8,192 pending-arrival slots. No
application probe, build, test, agent or profiler ran during timing.

| Round | Order | Current p99 ms | Candidate p99 ms |
|---|---|---:|---:|
| 1 | Current → candidate | 0.925 | 0.950 |
| 2 | Candidate → current | 0.989 | 0.995 |
| 3 | Current → candidate | 0.934 | 0.979 |

All three pairs lose; median p99 is **0.934 → 0.979 ms (+4.82%, worse)**. All
**180,000** offered warm/timed writes passed complete decoded Turbo-response
checks and independent exact ACK-ID/body/creator/room/unique-client-ID/FTS audits.
No errors, dropped or unscheduled arrivals; integrity and unchanged-seed FK checks
passed. The new private client retains complete bodies and validates/decodes after
its completion timestamp, before its next worker request. It is common to both
arms; these are not relabeled historical fixed-rate measurements.

The candidate is rejected and its budget is closed. No capacity, Cable or broad
HTTP/browser comparison followed primary failure. Removing the preliminary room
read alone is not a supported tail improvement; reopening needs a changed premise.

## Gates and retained evidence

Current `bin/check` passes. Focused read-only creation/publication review found no
additional contract violation. The HTTP probe passes front/Cable race tests, vet
and a real-listener completion check; its client passes Rust fmt, strict Clippy,
all 20 existing tests and locked release build. The room candidate passes local
`bin/check`, native database/application/web races, vet and release builds. Its
missing/inaccessible-room body/upload test also passes the original source.

Before timing, probe setup checks exposed a detached reference-link failure and a
private TLS-wrapper type assertion; both were corrected. The initial unqualified
Rust Clippy invocation lacked its component; final accepted gates/build explicitly
used 1.98.1 in an isolated container. These were probe/environment failures, not
production defects. Unchanged production did not receive another broad shared
HTTP/browser/media sweep; existing source-specific shared evidence is not expanded
by these private controls. No readiness or zero-bugs claim follows.

[`manifest.json`](manifest.json) lists 23 compressed artifacts with hashes:
patches, runners, metadata, complete sanitized measurement/audit records, correlated
summaries and gate logs. Private raw ACK maps, per-request timestamps, databases,
cookies, session files and environment values are excluded. Private working
copies retain their original `20261012` directory labels despite the UTC capture
date above. The existing Cable progress record remains separate historical evidence.
