# Consolidate one-to-one message associations — rejected, 2026-10-13

Combining receipt room, author and attachment reads removes two SQL round trips
per ordinary display but produces no meaningful whole-application CPU reduction.
The native diagnostic pair passes exact outcome audits and **misses the declared
3% CPU gate**. The prototype is rejected; no timing-control cohort or production
change follows.

| Native complete POST diagnostic | Current fork | Projection |
|---|---:|---:|
| Audited writes | 66,000 | 66,000 |
| Joined process CPU, s | 34.520649 | 34.509049 |
| CPU, µs/completed write | 523.040 | 522.864 |

The change is **−0.034%**, effectively flat. One diagnostic pair cannot establish
a small CPU gain or a latency improvement. The failed gate is not refined after
seeing the result and no extra favorable-run attempt is used.

## Mechanism and preserved behavior

Application source is `8351224`. The hypothesis follows the current-fork
[independent native CPU capture](../native-perf-attribution-20261013/README.md),
which exposes display materialization in ordinary operation but has incomplete
cgo caller chains. It does not reopen the earlier rejected `IN`-to-JSON-ID-join
rewrite: this candidate consolidates association round trips instead of changing
the ID-list plan for the same separate queries.

The prototype projects caller-owned `[message_id,room_id,creator_id]` tuples.
It joins one room, an optional author and at most one attachment blob. The correlated
attachment selector keeps the lowest valid attachment edge, skipping dangling blob
edges; a missing author suppresses attachment and boost display. Boosts and direct
participants remain independent collections, so joined rows cannot multiply them.
Mention users remain a separate, mention-only projection. All reads stay inside
the original short snapshot; no connection survives materialization. No message
record is reloaded, and authorization, writer state, caches, queues, deadlines and
resource joins are unchanged.

Existing database, presentation and application tests pass. A focused private case
checks caller-owned/unpersisted records, duplicate identities, missing authors and
missing rooms—risks introduced by the new outer-join projection that the existing
collection/lowest-attachment test did not cover. These tests are prototype evidence,
not permanent test-suite additions. The initial archive lacks read-only reference
fixtures; supplying the pinned schema/vector paths corrects setup without skipping
assertions. Nullable author timestamps are parsed only when an author exists.

The native build tests the projection, restores the exact baseline file, then
builds each arm serially with Go 1.27.1, CGO, `sqlite_fts5`, `-trimpath` and no VCS
embedding. The resulting baseline binary is **byte-identical** to the current
production binary:
`8a1409c09aa79e188010fb04dfa0c63b66e4f2748de734a117f6ae5db49b0e29`.
Candidate SHA-256 is
`676ecf514bc26004815994b0c88804c78153a2e66aa8cee6b4542d116d5d8aa5`.
An initial missing build-cache mount fails before compilation; the existing owned
cache path corrects it. No dependency, toolchain, host or network setting changes.

## Declared gates and complete outcomes

The work budget is one current/prototype process-CPU pair, two warm seconds plus
twenty timed seconds at 3,000 offered/s and 64 workers: **132,000 writes** total.
The physical desktop, seed, worker/affinity settings and maintained receipt client
match the [three-way baseline](../current-three-way-20261013/README.md). There are
no CPU profilers or private server hooks. Setup/login/shared preflight precede a
fresh server restart, and a supervisor collects joined OS CPU at clean shutdown.
The diagnostic's latency values do not establish application performance gains.

Only a work-gate pass would authorize the predeclared three interleaved pairs each
of fixed POST, capacity and 64-person Cable: at least 8% median fixed-POST p99
improvement, no greater than 5% median loss in capacity, complete-delivery throughput
or paced all-client p99, and all correctness gates. **Those controls do not run**,
because the CPU gate fails.

All offered messages complete with valid gzip/Turbo/client/sequence receipts.
Canonical ACK IDs match exact persisted body, client, creator, room and FTS values.
Both databases pass integrity and retain the unchanged seed FK population; both
servers exit cleanly. No production defect is established.

[`manifest.json`](manifest.json) retains both raw receipt populations, audits,
joined CPU, preflights, the declared gates, source diff/final files, focused tests
and failed/corrected setup logs. `bin/check` passes on the unchanged application
before evidence delivery. The result closes round-trip consolidation as this
performance candidate. The next question is where individual POST critical-path
waits and scheduling—not assumed extra SQL or compression CPU—account for the gap.
