# Application design and compatibility boundaries

Reconciled against `70d6329` (application source `811ae04`). Review these contracts
when changing their owners. This document describes the current boundaries and
constraints to preserve, not a checklist of unfinished migrations or proof of
performance parity. The [port implementation record](go-conversion.md) and
[validation record](validation.md) are dated historical evidence.

## Ownership

- `cmd/campfire/runtime.go` constructs services and owns resource lifetimes.
  `internal/web.New` receives prepared dependencies; it does not open persistence,
  install database effect callbacks or start workers.
- `internal/web` adapts HTTP input, sessions, forgery checks, ordered routes and
  operation results. Business SQL belongs in `internal/database`, not handlers.
- `internal/application` coordinates cohesive command/query families, owned
  receipts, display preparation and independent postcommit obligations. Browser,
  bot and trusted queued replies share the relevant operation owners.
- `internal/database` owns SQL, persisted graphs and transaction-current invariant
  checks. Commands return receipts; persistence does not call into transports,
  storage, workers or HTTP. The driver owns bounded connection-local statement
  caches; there is no global read-statement wrapper to migrate.
- `internal/presentation` renders prepared display data. `internal/richtext`
  transforms owned content with resolved mentions, rather than DB callbacks.
- `internal/storage` owns files, staging, media computation and purge
  continuations. Database owns blob/attachment/variant rows; application coordinates
  their commit boundary.
- `internal/jobs` owns bounded scheduling and admission. Application task types own
  meaning. `internal/cable` owns socket lifetimes, indexed recipient candidates,
  fresh authorization and transport backpressure. Integration clients own provider
  protocols.

Use concrete Go dependencies and one owner per invariant. A package extraction is
not a reason to introduce an ORM, generic repository, DI container, event bus,
replacement scheduler or universal cache. Complete affected caller migrations and
remove superseded paths rather than retaining internal compatibility shims.

## Commit and effect boundaries

Message commands in `database/message_commands.go` and application `Messages`
separate failed mutations from committed results with processing errors:

1. Decode typed input and prepare owned content/staged files outside the writer.
2. Acquire the single immediate-transaction writer. Select current permission and
   authoritative receipt inputs there. Authority linearizes at this observation,
   not at response delivery or a previous HTTP check.
3. Commit records, associations, canonical body, FTS and timestamps together.
4. Return owned identities, display inputs and detached-file candidates. Transfer
   staged-file ownership immediately after successful commit.
5. Run independent postcommit obligations with replacement bounds; then present
   and deliver output. Processing or delivery failure does not undo the commit.

For omitted-body updates, derive search text from writer-current content,
attachment fallback and mention inputs. Caller-prepared plaintext is not
independently authoritative. Pure preparation can move outside the writer only
when the exact committed inputs remain valid; timestamps alone do not establish
that invariant. Do not render, process media or perform external IO under the
writer. Trusted queued webhook replies remain an explicit operation, not a general
boolean authorization bypass exposed to callers.

Posting disconnect must not abandon committed work. Attachment processing stays
joined to the command with `context.WithoutCancel`; `Store.ProcessAttachment`
owns a two-minute budget covering admission and SQL as well as subprocess work.
Analysis and preview children retain their separate 30/60-second limits. Native
libvips calls are not preemptible through this API, so the context is not a hard
wall-clock bound on native execution.

`MessageNotifications.Created` gives unread publication and push preparation
separate five-second budgets. One timeout must not consume the other's budget.
Notifications and webhook admission precede fallible created-message presentation.
Each effect retains its own supported authorization/admission boundary; combining
work must not move that boundary silently.

## Query algorithms and observations

The local writer, bounded read-only WAL pool and separately pinned
`PRAGMA data_version` observer have different jobs. The observer covers foreign
as well as local commits; its token is connection-lifetime state, not a portable
persisted revision. Statement caches reuse plans, never results or permission.
Do not replace fresh authentication or foreign observation with a TTL or local
application counter.

Use selective read snapshots when membership, pagination and associations need
one observation. Materialize SQL results before releasing the snapshot, then
render, encode and transmit outside it. Cache generation checks bracket the
observations they protect. If the generation changes, disable unsafe reuse or
admission rather than relabeling old data with a new generation.

`MessageQueries` uses explicit references, prepared display data and immutable
fragments. Batch missing messages and relationship types—authors, boosts,
attachments and mention targets—rather than querying per message. Do not join
independent collections into a row-multiplying query. Preserve:

- Lowest-attachment-ID selection and owned attachment display transformations.
- Search's matching membership/body observation; a miss must not become unscoped
  ID hydration.
- `created_at` pagination and existing tied boundaries. An `(created_at,id)` cursor
  is a separate compatibility change, not an incidental index optimization.
- Set-based membership differences and indexed direct-room candidate selection.

These constraints leave query shape, compatible indexes and selective batching
open to improvement. Measure actual plans, rows visited and complete workloads;
package boundaries and cache-hit throughput do not establish algorithmic quality.

## Publication and bounded work

Cable's indexes select candidates, not authorization. `Hub.PublishStreams` batches
same-event streams while retaining fresh session and membership checks local to
publication. Release the hub lock before database work. The per-connection writer
owns frame order, cancellation and its reusable write deadline; nonblocking bounded
queue admission cancels overflowing connections rather than retaining more work.
Do not increase queues or bypass live checks to improve a benchmark.

Notification tasks normally retain small owned identities. Push is the deliberate
exception: its selected subscription credentials and prepared payload are the
admitted provider operation. Refetching in the worker would change recipient timing;
an identity-to-payload registry would merely hide the same retained snapshot.
Admission distinguishes accepted, full and closed queues. Best-effort rejection
must remain visible; critical cleanup retains its continuation. A worker must not
block indefinitely admitting dependent work to its own saturated queue.

`Cleanup.Detached` registers purge roots before job admission. Storage purge
continuations retain graph-removal receipts, failed keys and the descendant
frontier before unlinking. Each destructive graph step rechecks incoming
references. Independent descendants continue after another unlink fails, and
shutdown retries retained work. These continuations are process-owned, not
crash-durable; this design does not introduce an outbox, exactly-once delivery or
speculative orphan sweeps.

Media computation remains outside SQL transactions. Derivative winners link
atomically and losers are discarded. Queued analysis reloads by blob ID rather
than retaining request staging state. Preserve existing media algorithms and
reference vectors when changing ownership.

## HTTP representations and caches

`front.PrepareResponse` and `front.EmitResponse` own finite application response
completion: validators, selected coding/body length, conditional GET/HEAD and
output. Borrowed request-buffer parts are valid through synchronous emission;
retained caches must own copies. Live cookies and security headers remain outside
retained representations. Preparation success is not successful socket delivery.

Files, upgraded sockets and streaming output keep their protocol-specific paths.
Private completed-response coding and public compression have distinct policies;
do not merge them merely to reduce the number of types. Finite Turbo markup is
completed output despite its stream format name.

Keep these cache purposes distinct:

- Public HTTP reuse: declared route/capability policy, freshness and variants.
- Private response reuse: live identity/access and an unchanged foreign-aware epoch.
- Fragments/shells: complete display dependencies and immutable parts.
- Encoding memoization: exact byte/part identity and coding parameters, not weak ETags.

Retained cache budgets are not active-work or disk-admission budgets. An optimization
must reduce work rather than conceal it with more retained heap or headroom.

Routing and policy share the ordered route declarations. Preserve escaped paths,
optional formats, method override, query precedence, absent/null distinctions and
endpoint errors. Normalize proxy/peer address, authority, scheme and original
target once; untrusted-header stripping covers every downstream alias. Keep the
configured trust policy rather than silently substituting new proxy defaults or
`ServeMux` semantics.

## Shutdown and compatibility

Runtime stops intake, closes and joins upgraded sockets/presence cleanup, joins
admitted handlers, drains or cancels and joins jobs, retries purge continuations,
and closes persistence last. `net/http.Shutdown` does not join upgraded sockets;
cancellation is not a join. The CLI's forced-exit deadline handles native work
that cannot finish. Forced exit is not successful resource teardown or proof of
crash durability; shared resources must not close under surviving work.

Keep existing SQLite data/upgrades, storage keys/layout, signatures, cookies,
routes and frontend usable. Compatible query/index changes are allowed. Record
intentional behavior changes in [README.md](../README.md#known-differences).
Inherited public avatar caching can serve warm signed responses without another
application authentication check; the pinned Rust also strips live cookies on
public-cache misses. This is not an established Go-only bug. Do not silently
change that policy, trusted webhook reply semantics, eager media timing, push
recipient selection, timestamp ties or already-admitted Cable-frame revocation
semantics. Stronger guarantees need a deliberate contract decision, not a refactor
disguised as a bug fix.

## Verification and design changes

Fix demonstrated contract violations before optimizing. Use focused regressions,
`bin/check` and applicable shared verification as described in [AGENTS.md](../AGENTS.md).
Exercise the public listener for protocol/coding changes and media vectors for
processing changes. Historical validation does not replace affected current checks.

Use representative profiles to choose optimization work after correctness gates
are satisfied. Start with CPU/allocation costs; when elapsed time is not explained
by CPU, inspect blocking, scheduler behavior, SQL access paths and request boundaries.
Prefer simpler algorithms, query shape, batching, ownership and resource lifetimes
before micro-optimization complexity. Compare absolute work per completed operation,
not just a smaller percentage in a changed profile; lower CPU alone does not prove
better latency or throughput.

Prefer quiet native hardware when available. Local/VM profiles can identify work in
that environment, but keep their timing and causal claims separate from native
results. A candidate needs a concrete hypothesis, bounded discriminator, preserved
contracts and matched complete-workload benefit. Use uninstrumented end-to-end
checks to validate profile-guided changes, retaining adverse populations and current
source identities when comparing Go upstream and Rust.

Distinguish delivery failures from timing variability. The retained
[low-throughput Cable workflow observations](../bench/results/cable-progress-20261011/README.md)
do not establish a transport defect: the instrumented slow phase delivered every
acknowledged message without an expanding socket queue,
and its lower delivered rate followed slower POST completion. Its cause remains
unclassified. Ordinary-operation profiles can support useful improvements without
explaining that historical phase. Do not claim it fixed, blame system noise without
evidence, or make its reproduction or historical benchmark parity a standalone
delivery gate. Diagnose and fix any demonstrated correctness or supported-workload
regression; report unresolved measurement limits separately.

External semantics to revisit when dependencies or contracts change:
[Go connection ownership](https://go.dev/doc/database/manage-connections),
[transactions](https://go.dev/doc/database/execute-transactions),
[SQLite isolation](https://sqlite.org/isolation.html),
[query planning](https://sqlite.org/queryplanner.html),
[HTTP semantics](https://httpwg.org/specs/rfc9110.html) and
[caching](https://httpwg.org/specs/rfc9111.html).
