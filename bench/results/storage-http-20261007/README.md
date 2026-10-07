# Upload and download controller contracts

Fixed four observed differences from pinned Rust:

- Missing blob files no longer return a century-long immutable 404. Failed file opens and invalid
  file objects disable caching; successful proxies retain immutable caching.
- Representation proxies honor attachment disposition but ignore Range, returning the complete
  representation or a conditional 304. They do not advertise byte ranges.
- Blob proxy ranges use inline disposition unless the MIME type forces attachment. Disk ranges,
  streaming, multipart framing and HEAD behavior remain separate policies.
- Direct-upload allocation accepts URL-encoded scalar blob fields as well as JSON.

The shared file reader now receives an explicit controller policy instead of inferring it from
the request path. Avatar/logo files retain their existing ordinary `net/http` content policy.
Removed two unused upload/attachment helpers that committed blobs separately from their records;
active callers already use transaction-owned staging.

Go regressions fail on the original handlers for each defect, including representation range
bytes and conditional requests. The missing-file check extends the existing JSON direct-upload
regression rather than retaining a redundant standalone test. Range tests validate complete
multipart parts, termination and content length through the real HTTP server.

`bin/check-parity` now consumes binary responses and retains response headers, actor cookie jars
and same-origin headers. Two new workflows check JSON/form allocation, integrity rejection with
no file left behind, successful PUT and public download, stale-session rejection, disk/blob
ranges and HEAD/304/416 behavior, and seeded representation disposition/full-body/304 behavior.
Both new workflows passed in Rust before the fix and failed in Go. Both implementations now pass
all eight workflows. Native and Linux CGO vet/race checks and the WebSocket-fork gates pass.

No application timing was run. These probes do not establish multipart attachment lifecycle,
arbitrary form-metadata structures, cross-port multipart byte equality, live Cable behavior or
crash durability. The earlier media-vector mismatch is unchanged and recorded separately in
`../storage-staging-20261007/`.
