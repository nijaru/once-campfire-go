# Structured direct-upload metadata

Direct uploads now consume structured request parameters. URL-encoded forms preserve nested
hashes, scalar arrays, ordered arrays of hashes, empty/bare values, numeric hash keys and
repeated-key behavior. Query parameters replace the corresponding top-level body parameter,
as in pinned Rust. The ordered parser is scoped to the direct-upload route; existing flat form
consumers are unchanged.

JSON retains scalar types and number precision. Direct-upload metadata follows Action Dispatch's
`deep_munge`: null array elements are removed recursively, while null hash values remain. This
also fixes the previously raw-body handler's missed query precedence. The handler no longer
serializes every field to JSON just to decode its scalar value again.

The parser rejects type conflicts, invalid encoding and excessive depth at the request boundary.
It retains the server's body limit, allocation limit and authentication. A valid form above Go's
independent default 10 MiB parsing limit is covered: already-bounded input is not reread through
an unwrapped body with a smaller hidden cap. Depth, allocation/body limits and rejected requests'
no-allocation invariant have focused HTTP regressions.

Before: the metadata roundtrip fails on original Go, while the final external workflow passes
Rust. After: both ports pass all eleven isolated workflows. Response and stored metadata are
checked independently, including typed JSON, query precedence, ordered nested forms and rejected
conflicts/depth. Native/Linux CGO vet/race and WebSocket-fork gates pass.

The initial JSON probe incorrectly expected null array elements to survive Rust's normalization;
source and observed Rust output established the corrected contract. Earlier exploratory output
is not passing evidence. The final fail-before report uses that corrected workflow.

This does not claim compatibility for malformed Rack bracket keys or multipart metadata. No
application timing was run, and the unrelated exact video-vector failure remains documented in
`../media-environment-20261007/`.
