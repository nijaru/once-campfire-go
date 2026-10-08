# Cohesive application refactor

Design baseline: `38ab6c7`, reviewed 2026-10-08. Reference pins: Rust
`64f86353021145b63849fb1cd93adeb08f3b8dbb`, Rails
`90b330024dec3e757c79b6a7e6568f93da8e3148`.

This is the accepted direction and execution plan for restructuring the Go port.
It supersedes incremental, profile-led optimization as the approach to this work.
The [implementation record](go-conversion.md) and [validation record](validation.md)
remain historical evidence, not requirements or proof that the current design is
optimal. Review this document when changing application boundaries. Update decisions
in place when implementation evidence requires reconciliation.

## Goal and limits

Establish one owner for each invariant, observation, resource and effect. Remove
cross-layer orchestration, redundant data preparation and obsolete paths. Use
concrete Go components, explicit SQL and the standard library; no web framework,
ORM, DI container, generic repository layer, event bus or universal cache framework.

Preserve existing installations: SQLite schema/data, storage keys/layout, cookies,
signatures, routes and frontend. Compatible additions are possible but must serve
the resulting query architecture. Do not change references or media algorithms.
Keep declared safety differences and document any new intentional differences in
[README.md](../README.md#known-differences).

Design and reconcile the affected architecture before further profiling. Verify
correctness during implementation. Profile the complete application only after the
structural migrations and their contracts are coherent. Principles establish
ownership and algorithmic shape; they do not prove optimal pool sizes, indexes,
cache admission or cross-language performance. Rust parity is not established.

## Current problems, grounded in source

The whole-application review followed HTTP ingress through commands, persistence,
presentation, publication, jobs and shutdown. These findings are source-derived;
the concurrent/error cases still need focused regression reproduction.

- `internal/web/server.go::New` constructs storage, Cable and workers, and installs
  database callbacks. `internal/database/database.go::DB` can initiate purge,
  content-removal and reconnect effects through those callbacks. This is a runtime
  ownership cycle despite the absence of an import cycle.
- Browser creation, bot creation and webhook replies reconstruct related
  commit/media/render/notification flows independently in `web/server.go`,
  `web/bots.go`, `web/webhooks.go` and `web/message_attributes.go`.
- `web/message_attributes.go::updateMessageAttributes` derives omitted-body search
  text from an earlier read. `database/messages.go::UpdateMessageWithUpload` can
  leave a concurrently changed body intact while overwriting FTS with that stale
  text. Atomic writes alone do not establish the derivation invariant.
- Room mutation authority is checked in HTTP before the mutation transaction.
  Message mutations already put permission checks inside their transaction.
- Some command paths commit, then require another read or media operation before
  returning success. Later failure can obscure the committed state and suppress
  independent effects.
- `web/messages.go::hydrateMessageViews` loads boosts and attachments per missing
  message and full users for display. References use partially populated `Message`
  values and a `CreatorID == 0` sentinel. Rendering resolves mentions through DB
  callbacks; infrastructure errors are discarded by several rich-text wrappers.
- `database/read_pool.go` holds a global cache mutex during `PrepareContext`, which
  can wait for a pooled connection. Admission retains the first 256 SQL shapes.
- Completion and coding policy are divided among response buffering, recorded parts,
  private-cache gzip, application compression and public compression.
- Purge removes graph rows before unlinking, but loses descendant traversal and file
  identity on an unlink error. Queue rejection and timeout shutdown have different
  consequences for notifications, cleanup and still-running work, yet their owners
  do not consistently distinguish them.

Retain sound foundations: one immediate-transaction writer, bounded WAL readers,
the pinned foreign-commit observer, staged files, immutable parts, guarded template
escaping, indexed publication recipients and publication-local authorization.

## Responsibilities and dependency direction

### Process runtime

The composition root parses immutable configuration, opens resources, constructs
services/adapters and unwinds partial construction. HTTP construction must not
install behavior on a database object or secretly create background services.

Coordinate listener, socket, command, job and media lifetimes. Close persistence
last. Do not replace one oversized `Server` with an oversized application object;
keep operation families cohesive within the application package.

### HTTP adapter

`internal/web` owns protocol adaptation: normalized request facts, ordered routing,
cookies/browser state, forgery checks, bounded decoding, format selection and
mapping operation results to HTTP. It does not own SQL, business effect sequencing,
worker construction or hidden data preparation during rendering.

Use one ordered route declaration for handler and policy: authentication mode,
decoder, formats, completed/file/socket behavior and public-cache eligibility.
Preserve declaration order, escaped paths, optional formats, method override,
query precedence, null/absent distinctions and existing error contracts. This is
not permission to substitute `ServeMux` semantics for the released route contract.

Normalize peer/proxy address, authority, scheme and original target once. Consumers
must not independently reinterpret forwarding headers. Preserve configured proxy
behavior; untrusted-header stripping must cover every alias consumed downstream.

### Application services

A small `internal/application` package owns command/query orchestration and
post-commit obligations. Browser HTTP, bot HTTP and workers invoke the same
operation families. It may coordinate a presenter for publication; transaction
logic itself has no templates, HTTP writers or external transport work.

Use concrete dependencies. Introduce interfaces only for real protocol boundaries,
not to wrap every database operation. Trusted webhook replies and administrative
cleanup remain explicit operations rather than boolean authorization bypasses.

### Database

`internal/database` owns explicit SQL, persisted graphs and transactional invariant
checks. Expose typed commands/query results, not public raw pools to HTTP callers.
Return operation-specific results instead of invoking callbacks into other owners.

Keep one local writer and read-only pooling. Use short, selective read snapshots
where membership, pagination and associations require one observation; never hold
a connection while rendering, encoding, transmitting or doing media/network work.
The separate serialized `PRAGMA data_version` observer remains authoritative for
local and foreign commits. Its token is connection-lifetime state, not a persisted
revision interchangeable between connections.

### Presentation and content

Extract `internal/presentation` for templates, message/shell assembly, signed
presentation URLs and fragment reuse. Its inputs are complete owned display data,
explicit host/origin and immutable configuration. It must not query a database or
retain an HTTP request. Keep template ownership of markup and existing escaping
safeguards; do not generalize manual interpolation.

`internal/richtext` remains a transformation library. Supply resolved mention data
rather than database callbacks. Separate deliberate malformed-content fallbacks
from DB/cancellation failures. Parsing/canonicalization, search derivation and
presentation are different consumers; do not calculate every output for every use.

### Files, execution and transports

Storage owns bytes, staging, media computation and derivative coordination.
Database owns blob/attachment/variant rows. Application coordinates their commit
boundary. Jobs own scheduling; application owns typed task meaning. Cable owns
connection admission, recipient selection, live publication authorization and
transport backpressure. Integration clients own network/provider protocols.

Database does not import application, web, jobs, storage or Cable. Presentation
does not load DB data. Web does not construct application dependencies. Remove the
existing reverse callbacks and cross-layer staged lifecycle interface as callers
migrate; do not leave compatibility paths for these internal interfaces.

## Operation contracts

### Commands and commit

1. Decode typed input and prepare owned content/files outside the writer.
2. Start a short writer transaction; load current actor/permission and authoritative
   mutation inputs there. Authority linearizes at that transaction observation,
   not at the last response byte. Do not invent stronger guarantees for in-flight
   revocation.
3. Commit record, associations, canonical content, FTS and timestamps together.
4. Return an owned receipt plus explicit effect inputs: committed identities,
   revoked users or detached blob candidates, as appropriate to that operation.
5. Transfer staged-file ownership immediately on successful commit.
6. Initiate post-commit effects independently of HTTP rendering/cancellation.
7. Render/respond from the receipt or a separately identified follow-up query.

Precommit failure, postcommit processing failure and response-delivery failure must
remain distinguishable. Do not report a follow-up read failure as evidence that
mutation did not commit, and do not emit a fake successful Turbo response. Preserve
endpoint-specific responses initially while making commit phase explicit; settle
any changed recovery/status behavior before changing that endpoint's contract.

A submitted canonical body and its plaintext are one owned derivation. For omitted
body updates, earlier caller plaintext is not authoritative. Derive from current
transaction state, including attachment fallback and resolved mention inputs. Move
pure preprocessing outside the writer where possible; any optimistic preparation
must validate exact inputs, not merely timestamps, before committing. Do not put
full presentation, file processing or external IO under the writer.

### Scoped, cache-aware queries

Use distinct reference, persisted-record and display types. Select authorized
records/references in a bounded query. Identify reusable immutable fragments, then
batch only missing messages, author displays, rooms/participants, boosts,
attachments and mention targets. Assemble in the established order.

The application query coordinator can consult presentation cache identities during
preparation; the renderer cannot fetch data. If a logical query needs a read
snapshot, finish all required SQL and materialize data before releasing it; render
outside that lifetime. Cache generation checks bracket the relevant observations.
A changed observation disables unsafe reuse/admission rather than relabeling older
data with a newer generation.

Query count should scale with relationship types/batches, not message count. Do
not join independent collections into a row-multiplying query. Preserve the
lowest-attachment-ID selection and search's matching membership/body observation;
a search miss must not become unscoped ID hydration.

Resolve pagination anchors once. Preserve `created_at` ordering and existing tied
boundary behavior; adding an `(created_at,id)` cursor is a separate compatibility
change. Use sets for membership differences and restrict direct-room candidates
through indexed participant memberships rather than scanning all direct rooms.

### Representation completion and caching

Define a concrete completed representation: status, format, immutable parts,
representation headers and cache/validator policy. Live cookies/security headers
remain request-owned. One emitter owns coding, `Vary`, conditional GET/HEAD,
selected-body length and final output. Distinguish preparation success from
successful socket delivery.

Keep files and upgraded sockets separate from completed-body retention. Finite
Turbo markup is completed output despite its stream format name.

Retain distinct policies:

- Public HTTP reuse: declared route/capability policy, freshness and variants.
- Private response reuse: live identity/access and unchanged foreign-aware epoch.
- Fragments/shells: complete display dependencies and immutable parts.
- Encoding memo: exact byte/part identity and coding parameters, not weak ETags.

Do not cache authentication or replace foreign observation with TTLs/application
counters. Keep conservative invalidation until a finer mechanism preserves foreign
writes. Use one encoding owner; remove superseded gzip paths after migration.
Define retained-cache budgets separately from active work/memory/disk admission.

### Post-commit work and cleanup

Use small owned task inputs, usually identities, not captured requests, contexts,
staged handles or subscription credentials. Make admission outcomes explicit.
Retain independent queues and existing in-process durability; no outbox or
exactly-once delivery requirement is introduced.

Do not make notification admission depend on successful message rendering. Critical
cleanup must retain an owned continuation when child/background admission fails;
best-effort notifications may have an explicit rejection policy. Do not block a
worker indefinitely trying to admit dependent work to its own saturated queue.

Graph deletion returns removed file identity and descendant candidates. Preserve
the frontier before unlink; continue independent cleanup and retain actionable
failed keys after rows disappear. Check current incoming references at each
destructive step. Do not add speculative orphan sweeps.

Retain eager media processing initially, its algorithms and vectors. Heavy compute
stays outside transactions; derivative winners are linked atomically and losers
are discarded. Queued analysis reloads by blob ID rather than capturing request
staging state. Live delivery checks have an explicit admission point; they cannot
retract traffic already sent.

### Shutdown

Stop listener/socket intake, drain active commands and socket/presence cleanup,
then drain accepted jobs/continuations. On grace expiry cancel remaining work and
account for completion before releasing resources. `net/http.Shutdown` does not
join upgraded sockets; cancellation alone does not establish a join.

Native libvips work is not preemptible through the current API. Preserve a bounded
CLI shutdown policy without promising impossible graceful joins: if owned work
cannot finish, forced process exit is distinct from successful resource teardown.
Do not close shared DB resources under surviving work and call it graceful shutdown.

## Compatibility decisions not to hide in refactoring

- Pinned Rust/Rails public avatar caching can serve warm signed avatar responses
  without another application authentication check; Rust also strips live cookies
  on public-cache misses. This is inherited behavior, not a confirmed Go-only bug.
  Preserve it initially and make route policy explicit. Tightening that contract
  needs a deliberate documented security decision, not silent cache relocation.
- Preserve existing proxy configuration while consolidating trust ownership.
  Match reference handling where applicable; new trusted-proxy defaults are not
  implicitly authorized by package extraction.
- Retain trusted queued-webhook reply semantics, eager media timing, in-process job
  durability, timestamp ties and file-controller protocol exceptions initially.
- Do not silently change push recipient timing or already-admitted Cable-frame
  revocation semantics. Define the supported live-check boundary and reconcile
  any stronger policy before implementation of that change.

These boundaries do not block internal ownership migrations. Resolve the relevant
behavior choice before its migration cohort, using source and actual requirements.

## Execution plan

Execute coherent caller migrations, not isolated optimizations. Each completed
cohort removes its superseded production paths. Necessary temporary migration
boundaries are not permanent compatibility shims.

1. **Command/commit boundary.** Reproduce stale omitted-body FTS and committed-result
   failures with focused tests. Introduce authoritative message/room command inputs
   and receipts, transactional permission checks, and explicit effect results.
   Migrate browser, bot and webhook message callers together; remove stale derivation
   and mandatory postcommit reconstruction. Start here.
2. **Application and persisted graph ownership.** Extract shared orchestration;
   migrate rooms/accounts/attachments and cleanup results. Move graph SQL to database,
   file handles to storage coordination, and remove DB-installed effect callbacks
   and transaction-capable staged interfaces once all callers migrate.
3. **Query/presentation boundary.** Introduce explicit references and narrow display
   inputs, cache-aware batch assemblers and batch-resolved mentions. Extract pure
   presentation and remove per-item loads, DB callbacks and the broad page/model
   sentinel paths. Preserve search and snapshot invariants.
4. **Typed effects and runtime lifecycle.** Migrate task submissions/continuations,
   admission outcomes and shared message effects. Move composition out of HTTP;
   complete socket/worker/resource shutdown ownership and constructor unwind.
5. **Ingress, routes and response ownership.** Consolidate normalization/dispatch,
   typed decoding and completed representation emission. Migrate coding and cache
   policies without changing inherited public/capability contracts accidentally.
6. **Resulting query/resource design.** Remove redundant preparation layers: prefer
   the supported bounded connection-local driver cache over the global read wrapper,
   with one strategy per pool. Review actual resulting access paths and compatible
   indexes, reader/page-cache budgets and active-work admission. Do not stack cache
   mechanisms or revive rejected driver/checkpoint experiments without new rationale.
7. **Verification, then profiling.** Run the shared practical checks below. Once the
   design is implemented coherently, take fresh complete-application profiles for
   warm, changed and write workloads; evaluate remaining choices with matched,
   sequential measurements. Publish raw evidence and limitations, not optimality
   or parity claims unsupported by results.

## Verification and delivery

Use focused observable regressions for demonstrated failures, then `bin/check` and
applicable shared HTTP/write contracts and browser flows. Preserve original failing
assertions and distinguish product defects from environment/oracle failures. Do
not invent an exhaustive acceptance matrix or rerun unrelated media inventories.
Run media vectors when processing changes; graph/lifecycle migrations use their
existing integration coverage and focused failure cases.

Exercise the real public listener when validating protocol/security/coding changes.
Preserve fresh sessions/access, foreign commits, committed-write/FTS audits,
cookies/flash, conditional/HEAD/ranges and cancellation in affected cohorts. No
builds/tests/agents during timed comparisons. Reuse valid unchanged evidence.

Commit coherent agent-owned chunks and synchronize only `origin/perf/campfire`.
Do not bump versions, publish, open PRs, post comments or push upstream without
explicit permission.

## Primary sources and refresh

Review current source whenever these owners change. Revisit external semantics
when Go, SQLite or the protocol contract changes:

- [Go connection ownership](https://go.dev/doc/database/manage-connections)
- [Go transactions](https://go.dev/doc/database/execute-transactions)
- [SQLite isolation](https://sqlite.org/isolation.html) and
  [query planning](https://sqlite.org/queryplanner.html)
- [HTTP semantics](https://httpwg.org/specs/rfc9110.html) and
  [HTTP caching](https://httpwg.org/specs/rfc9111.html)
- `go doc net/http.Server.Shutdown` for upgraded-connection ownership.

These sources justify boundaries, not a universal architecture or projected speedup.
