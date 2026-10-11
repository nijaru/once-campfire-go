# Rust-to-Go implementation record

Port: `once-campfire-go`. Pinned Rust: `64f86353021145b63849fb1cd93adeb08f3b8dbb`.
Pinned Rails: `90b330024dec3e757c79b6a7e6568f93da8e3148`.

This records the original port and its 2026-10-02/03 verification, not current gate
results or an optimal-design claim. The maintained [application design](architecture.md)
replaces the completed migration checklist; current checks follow [AGENTS.md](../AGENTS.md).

The full application is implemented and independently runnable. The implementation uses the standard
library for HTTP/routing, templates, SQL access and process lifecycle, with focused libraries for
SQLite, WebSockets, bcrypt/ACME, HTTP/2 and HTML parsing. It preserves the original frontend and
existing database/storage/cookie formats. The Rust and Rails references are unchanged.

## Coverage

| Area | Implementation and verification |
|---|---|
| Persisted formats | Rails cookies, encryption, signing, IDs/SGIDs, Turbo stream names, Active Storage formats, Ruby coercions/escaping and JSON; golden vectors and cross-binary upgrade check |
| Database | Original schema/migrations, single-writer WAL/read pools, accounts/users/sessions, all room types/memberships, messages/FTS, boosts, bots, push subscriptions, lifecycle callbacks and transactional uploads |
| HTTP | 177 ordered route contracts, authentication/CSRF, setup/join, session transfers, management, browser checks, formats, conditional responses, error pages, Active Storage and ranges |
| Frontend | Original assets/importmap; translated templates for chat, search, management, bots, sessions, PWA and error states; original Playwright screen inventory plus end-to-end workflow tests |
| Rich text | Sanitizing, attachments/mentions, autolink, plain text, editing, malformed HTML limits; all six outputs across 658 reference cases |
| Media | Original signed URLs, layout, checksums, variants, previews, metadata, MIME detection; pinned-toolchain byte goldens and upload-to-real-thumbnail comparison |
| Realtime | Action Cable subscriptions, heartbeats, message/boost/Turbo streams, presence, typing, read/unread, room and session authorization/revocation, bounded queues and compression |
| Integrations/jobs | Bounded queues, Web Push/VAPID, webhook behavior, Open Graph, cleanup; reference vectors and local delivery tests |
| Operations | Production non-root image, HTTP/2/TLS/ACME, H2C, gzip/zstd/jitter, response cache, graceful shutdown, live backups and offline ONCE restore |
| Performance | Identical seed, CPU affinity, alternating repetitions, validated HTTP/write/Cable/media contracts; raw records under bench/results/ |

[route-coverage.json](route-coverage.json) records implemented routes and the deliberately preserved
missing-action error contracts from the reference. A route entry is not a blanket parity claim.
[validation.md](validation.md) records those checkpoints' checks and screen-layer results. Known differences
are listed in [README.md](../README.md), including strict DOM/network failures; no additional parity
masks or allowlists suppress them.

## Historical measurements

The original protocol and populations belong to their dated records under
[`bench/results/`](../bench/results/), not to a mandatory acceptance matrix for
later changes. Use the shared verification harness or [`bench/application`](../bench/application)
and the [current measurement rules](../AGENTS.md#performance-and-delivery).

Health-only measurements and early application preflights are not full-application
capacity evidence. The initial upload selector fetched an avatar; those timings
remain invalid. Later media measurements fetch actual representation URLs and
verify image bytes/hashes. Direct-listener results do not establish public
TLS/compression throughput, and CPU affinity does not establish physical isolation.
