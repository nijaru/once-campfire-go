# Room conversion and membership revocation

Extended `bin/check-parity` with a room lifecycle against the same isolated seed in Go and Rust:
create an open room, create a searchable message, warm member room/search/sidebar reads, reject
a non-owner rename, convert to closed with duplicate/nonexistent selected IDs, verify revocation,
reopen and restore active memberships, reject direct-room conversion, then delete and verify
message/membership/FTS cleanup. Revoked message creation returns the existing deleted-room
composer without storing a message; it is not a successful write.

The Rust workflow passed before the change. Go returned 404 for a revoked user's namespaced room
PATCH instead of redirecting home with the inaccessible-room alert. The update handler now uses
the same lookup-failure policy as room display/edit/delete. Authorization remains membership
scoped, including for administrators, and denied updates cannot mutate the room.

The new Go regression fails on the original handler and passes after the fix. Both implementations
now pass all six HTTP workflows. Native and Linux CGO vet/race checks and the WebSocket-fork gates
pass; raw reports, regression failures and gate logs are retained here. Formatting follows the
repository's Go tooling.

This verifies fresh HTTP authorization, persisted membership and search state, not live socket
disconnection or a concurrent revoke/write race. No timing was performed: the production change
only affects lookup failures and makes no successful-request performance claim.
