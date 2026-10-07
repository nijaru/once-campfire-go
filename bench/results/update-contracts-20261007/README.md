# Room and account update contracts

Local comparisons use the pinned Rust/Rails sources, not a fresh upstream fetch. No PR or other
upstream action was taken. No performance measurement is claimed for these correctness changes.

Room updates now preserve omitted names, distinguish SQL NULL from empty text, and avoid changing
room freshness for an unchanged name/type. The current name is resolved inside the write
transaction, so an earlier web read cannot overwrite a newer committed name. Membership revision
still runs for valid closed-room updates. Missing/empty required room roots are rejected before
creation, conversion or membership mutation; previously an empty closed-room update revoked all
members. Structured parsing also handles POST `_method` forms and top-level query precedence.
The existing ordered form tree is shared with direct uploads; unrelated forms retain flat
parameters. Creation now preserves nullable room names in the same schema.

Account restriction casting recognizes Rails' case-sensitive false values and stores blank as
JSON null. Paired checks verify the resulting member permission, not just the stored setting.
CSS saves preserve omitted styles, return to the CSS editor, and show the success notice. The
Chromium smoke now waits for the completed same-URL Turbo render and checks that destination and
notice; its old `/account/edit` expectation contradicted both pinned controllers and the Rails
controller test. The failed old expectation is retained separately.

## Evidence

The original room/account HTTP tests fail on the old handlers. Isolated paired reports show Rust
passing while the earlier Go binaries fail on omitted room names, missing roots, method override
query precedence, and false account settings. The final report contains thirteen workflows and
26 passing implementation runs, with binary/tool/SQLite-backup identities. Checks include fresh
room/sidebar reads, nullable names, rejected creation without allocation, unchanged memberships
on invalid requests, permission changes and CSS persistence/navigation/flash.

Native CGO `bin/check`, Linux CGO vet/race with `sqlite_fts5`, the WebSocket-fork race suite, and
current-binary Chromium smoke pass. Targeted room/account and competing-store variant tests pass
twenty race-enabled repetitions. The existing variant test now uses separate stores and SQLite
connections, so in-memory coalescing cannot hide database uniqueness or loser-file cleanup.

Independent source review of cancellation, staging ownership, external database visibility and
the changed controllers found the required-root and method-override defects; both were reproduced
and fixed. This review and these checks are bounded evidence, not proof of every input or schedule.
Strict malformed Rack compatibility, broad cross-browser parity and crash durability remain
outside these workflows. At this checkpoint the required exact video-vector mismatch remained
unresolved in `../media-environment-20261007/`. The subsequent
[verification](../acceptance-corrections-20261007/README.md) identifies the architecture difference
and passes the unchanged goldens; no encoding flags or expectations were changed to mask it.
