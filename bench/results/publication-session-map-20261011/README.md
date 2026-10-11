# Publication-local session map — 2026-10-11

**Promoted as a modest allocation reduction, not a throughput or latency gain.**
One publication-local token/user map now replaces the temporary token set and
separately allocated authorization-result map. Fresh SQL authority, decoded frames,
copying prepared-frame ownership, bounded queues and joined cancellation remain
unchanged. No schema, storage, cookie or frontend contract changes.

## Hypothesis and budget

The retained [native current profile](../native-frame-encoding-20261011/README.md)
at production `ccfa838` samples approximately 207 MiB in token-set insertion and
122 MiB allocating authorization-result maps. These are separate materializations
of the same token keys. Reusing one publication-owned map can remove redundant
allocation without another query, authorization cache or serialized-result path.

`DB.AuthorizeSessions` resets every supplied permission before querying, writes
actual authorized user IDs from the existing one-statement snapshot, and resets
all permissions on any returned error, including terminal row errors. Unknown,
revoked and inactive sessions remain zero. Publication still rejects identity
mismatches and fails closed on query errors. The API's only caller and existing
revocation test are migrated; the old result-returning method is removed.

The finite budget reused the unchanged current native baseline and allowed one
candidate CPU/allocation capture. The criterion was at least 3% fewer sampled Go
bytes/audited message, no increase in objects and no more than 5% sampled CPU
regression before application controls. It passed. Controls were three interleaved
64-person pairs and one candidate-first shared-cookie/64-socket pair, with no
further sweep or follow-up. The declared non-regression criterion rejected material
unexplained changes over 5%; it is not a claim of a statistical error bound.

This is not the rejected map-capacity sizing, ordinal/packed authorization,
owned-byte-copy or unescaped-frame encoder control. Those remain closed.

## Native allocation and CPU capture

| Version | Audited messages | Sampled Go bytes/message | Sampled Go objects/message | CPU ms/message |
|---|---:|---:|---:|---:|
| Current, retained native capture | 15,856 | 204,580.91 | 2,078.28 | 0.95106 |
| Single-map candidate | 15,666 | 195,701.71 | 2,058.21 | 0.95110 |
| Change | — | **−4.34%** | **−0.97%** | **+0.005%**, effectively flat |

The reduction is about 8,879 sampled Go bytes/message. Profiles use default memory
sampling and `GOGC=100`; they exclude native SQLite/C and kernel allocation. One
capture per version is attribution evidence, not a precise universal effect size.
Both include restart, connection, paced/saturated posting, drain and shutdown,
with setup/login in a separate unprofiled process. The private shutdown allocation
hook and final GCs are absent from production and timing binaries. CPU flushing
precedes those GCs. The CPU-hook behavior itself is unchanged.

## Complete-workload controls

Native Fedora/rootless Podman with isolated Go **1.27.1 linux/amd64**, CGO and FTS5.
Source before is `63c0197`, identical in production to `ccfa838`. Server CPUs
`0,2,4,6`, client `8,10,12,14`, runner `16,18`; affinity is not exclusive host or
physical-core isolation. Seed SHA-256 is
`691c52bfdc30b38bd8867e93b0914613df20e03d4cd7b00796ab68e0a5d26c9a`.

All controls use identical fixture copies, complete response consumption, 300
paced messages at 100 ms, five saturated seconds/four posters and exact write/
delivery audits. Order is current/candidate, candidate/current, current/candidate;
the shared-cookie pair is candidate/current. No task-owned agent, check, build or
profiler overlaps timing. Server binaries contain no private allocation hook.
The common client retains minimal post-completion ACK-ID/timestamp recording;
its overhead neutrality is not established. These are matched server-uninstrumented
controls, not directly interchangeable with historical generator populations.

| Three 64-person pair medians | Current | Candidate | Change |
|---|---:|---:|---:|
| Delivered messages/s | 3,073.5 | 3,084.5 | +0.36% |
| Saturated POST p99, ms | 3.337 | 3.321 | −0.48% |
| Saturated all-client p99, ms | 5.679 | 5.491 | −3.31% |
| Paced all-client p99, ms | 6.579 | 6.547 | −0.49% |
| Paced POST p99, ms | 5.095 | 5.207 | **+2.20%** |

Individual delivered-rate changes are −0.21%, +0.53%, +0.45%; paced POST p99
changes are 0%, +2.20%, +4.06%. Keep these small adverse observations: a smaller
heap profile is not a general latency gain. Medians are medians of run statistics,
not pooled percentiles.

The single shared-cookie control gives 3,474.3→3,462.0 delivered/s (**−0.35%**),
saturated POST p99 3.119→3.065 ms, paced POST p99 4.587→4.395 ms and paced all-client
p99 6.111→5.979 ms. It exercises deduplication of many sockets sharing one token;
one pair is a non-regression control, not a capacity estimate.

All **145,173 new capture/control writes** pass unchanged shared complete delivery
and exact ACK-ID/body/creator/room/FTS/integrity/presence audits. The seed's one FK
orphan remains unchanged, with no new violation. The reused 15,856-write baseline
is recorded separately and not counted again.

## Correctness and limits

- Private and integrated `bin/check` pass; native affected database/Cable/application/
  web vet and races pass. The first integrated formatting gate found one extra EOF
  newline; formatting corrected it before the successful gate, with no assertion
  change.
- The migrated real-database test retains current-session, membership and inactive-
  user checks and adds stale-map/unknown-token reset plus pre-cancellation clearing.
  It does not deterministically inject failure after partial row population; static
  review checked the named-error cleanup for that path.
- Independent focused read-only review finds no concrete ownership, authority or
  migration defect. It does not independently reproduce executed checks.
- Four existing functional workflows—message lifecycle, pagination, room membership
  and authentication—pass for candidate Go and the accepted pinned Rust binary.
  Initial setup failed before app requests because a read-only seed mount prevented
  SQLite opening WAL sidecars. An identical hashed disposable writable seed copy
  fixed setup; the failure log is retained and assertions are unchanged.
- The unmodified shared Playwright 1.63.0/Chromium smoke passes natively, including
  two-tab live delivery, duplicate suppression, markup safety and room/session flows.
  Browser source hash matches the shared checkpoint. Unaffected media vectors were
  not rerun. Earlier intermittent browser-count uncertainty is not erased.

Authorization remains a per-query snapshot, not atomic with queue delivery—unchanged
from production. No historical slowdown cause/fix, Rust-wide advantage, new queue
budget or global readiness claim follows.

## Evidence

[`manifest.json`](manifest.json) hashes 54 retained artifacts and measured base/
candidate source inputs. Artifacts include the exact patch, raw candidate profiles,
ACK records and audits, matched runner metadata, all individual control results,
analysis, builds/checks, browser result and functional setup failure/pass. Functional
reports omit request headers/cookies; runtime secrets/databases are not published.
Text is gzip-compressed; pprof files retain their native format. The reused baseline
profiles and private allocation-hook patch remain in the linked native-frame record.
