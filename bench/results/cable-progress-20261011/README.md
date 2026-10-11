# Cable low-throughput progress capture

The first instrumented throughput attempt delivers **583.1 messages/s**. Capture
stops there, as declared; the other five attempts are not used. The phase is limited
by completed POSTs, not a growing fanout/socket backlog. Its underlying cause is
still unresolved. No production change is proposed or landed from these counters.
Instrumentation overhead has not been isolated, so this is **not proof that the
historical pristine collapse has the same cause**.

## Workload and instrumentation

Production source is `ccfa838`, archived at documentation HEAD `2a935e7`. The private
[server patch](server.patch.gz) adds stage counters and per-socket enqueue/dequeue/
write counters without changing authorization, queues, framing, timers, deadlines
or cancellation. Socket counters are separate per connection, not one shared
per-frame counter. A separate loopback-only HTTP endpoint exports snapshots and
joins at shutdown; production routes do not expose it.

The [private client patch](client.patch.gz) retains the shared generator's response
validator and delivery accounting. It adds numeric acknowledged IDs, POST progress
and a joined 200 ms sampler. It briefly scans the existing receipt map under its
existing lock; no socket I/O occurs while that lock is held. The final ID map is
used for persisted-write validation, not per-frame logging.

The unchanged shared shape is 64 distinct fixture people/sessions, six subscriptions
per client, 100 paced posts at 100 ms intervals, then four closed-loop posters for
five seconds and all-client draining. Refresh remains 50 seconds. Server CPUs are
0–3, clients 4–7 and observer 8–9 in the Linux ARM64 VM. Go 1.27.1, Rust 1.98.1,
libvips 8.16.1, CGO/FTS5 and GOGC=100 are retained. Affinity is not isolation;
unrelated activity and physical/VM scheduling remain uncontrolled. No task-owned
build, test, agent or profiler runs during capture. No host dependency/configuration
change, provider write, queue expansion or GC tuning is involved.

## Observed phase

| Contract or measurement | Result |
|---|---|
| Ready / failed clients | 64 / 0 |
| Paced successful / all-client-complete messages | 100 / 100 |
| Throughput attempted / successful / all-client-complete | 2921 / 2921 / 2921 |
| POST / delivered rate | 583.1 / 583.1 messages/s |
| POST mean / p99 | 6.84 / 19.375 ms |
| All-client throughput delivery p99 | 24.639 ms |
| Paced all-client p99 | 6.043 ms |
| Drain extension | 0.00 s at the reported precision |
| Queue overflows / socket-write errors / auth rejections | 0 / 0 / 0 |
| Exact acknowledged persisted writes, including paced posts | 3021 |
| Presence connections after shutdown | 0 |

All 26 client progress samples have matching server observations. All-complete
progress advances in every approximately 200 ms interval, at **375.1–953.3/s**;
this is sustained low progress, not an observed stop-the-world plateau. Sampled
POST-success minus all-complete counts range from **0 to 7**. Socket enqueue minus
dequeue counts peak at **320 across 64 sockets**, only **7 on any socket** against
the unchanged 256-frame queue; at most two socket writes are sampled unfinished.
These observations do not exclude stalls between samples.

Command, effects, handler, publication and fresh authorization counts all continue
advancing. Sampled started-minus-completed maxima are 4 commands/handlers, 2 effects/
publications/authorizations and 1 notification/preparation. Low delivery closely
tracks low closed-loop POST completion; the sampled transport does not accumulate
an expanding queue or increasingly incomplete deliveries. Four posters and the
reported mean POST time are consistent with the observed completed-request rate.
That identifies a boundary to investigate, **not the cause inside or outside it**.

The inner observed window records approximately 1.63 server and 0.78 client CPU
cores. Server cgroup CPU/quota observations are unavailable, not zero. The client
cgroup includes the observer/runner and reports no added throttling; that does not
clear the server or host. No CPU-capacity or host-interference explanation follows.

## Counter and verification limits

Starts, completions and duration sums are independently read, not linearized.
Residuals remain signed. Queue differences are sampled counter differences, not
atomic queue-length measurements. Component duration sums overlap concurrent work
and can straddle bucket boundaries; they are neither wall-clock stalls nor stage
p99s. HTTP-handler return excludes the final net/http socket flush. There are no
individually correlated client/handler/flush records, so these data cannot assign
unobserved time to SQL, scheduling, network or client parsing.

Maximum client-to-observer lag is 21.25 ms; `/proc` observation takes at most 5.79 ms
and counter loading 0.066 ms, excluding JSON serialization/transfer. Added atomic
operations, clocks, JSON snapshots, receipt-map scans and acknowledged-ID collection
can perturb the workload. No matched pristine cost control is run after the first
sub-1,000/s phase: the stop rule is honored rather than collecting another baseline.
This remains a verification gap, not an overhead estimate or a general capacity
comparison. Earlier pristine low-rate samples and healthy profiles are retained at
[the scheduler record](../cable-scheduler-20261010/README.md) and
[the native/cancellation record](../native-cancellation-20261011/README.md).

Private Go vet, affected application/web races and full Cable/front race suites
pass. Rust fmt, strict Clippy, all 20 existing tests and locked release build pass.
No production source is modified; production `bin/check`, broad functional/browser
and media suites are not redundantly rerun for this diagnostic-only delivery.

The complete shared Cable validator passes unchanged. Offline SQL validates every
acknowledged ID against its exact sequence/body, room, creator and FTS body, plus
population count, unique markers and SQLite integrity. The seed already has one
message-to-user foreign-key orphan; the exact same violation remains, with **zero
added violations**. Whole-fixture foreign-key cleanliness is not claimed.

## Retained setup failures and decision

The first runner incorrectly limited its container to observer CPUs 8–9, preventing
its child from selecting client CPUs 4–7. This fails at login, before any Cable
throughput attempt. The corrected container permits 4–9 while pinning the observer
to 8–9; nested application affinity remains 0–3.

After the measured run, an overescaped anchored regex in the new SQL audit matches
no bodies. The raw sample and database remain unchanged. Offline recovery proves
the faulty regex rejects that same population and the corrected exact matcher
accepts all 3021 identities. A subsequent whole-fixture foreign-key assertion exposes
the pre-existing seed orphan; recovery compares exact seed/run violations instead
of pretending the fixture was clean. Both failures, original metadata and recovery
are retained. **No workload is rerun to obtain the pass.**

The bounded capture is closed. Evidence does not justify a wire, cancellation,
queue, authorization or scheduler rewrite. A new investigation would need bounded,
individually correlated HTTP/POST timing and an explicit instrumentation-cost
control; these observations alone cannot select a production fix. POST tails, the
adverse native capacity sample and stable Cable performance remain open readiness
gates. No PR-readiness or universal-correctness claim follows.

The [manifest](manifest.json) covers sanitized counters, samples, patches, adapters,
analysis and gates. Runtime logs, raw ACK maps, databases, fixture/session files,
environment values, cookies, signed paths and binaries are excluded. Raw sample
and database hashes are retained in recovery provenance.
