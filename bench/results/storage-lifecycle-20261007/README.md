# Multipart attachment and avatar lifecycle contracts

Both Go and pinned Rust pass ten isolated HTTP workflows, including two new storage lifecycles:

- Multipart image-message creation with the body omitted: exact uploaded file/checksum, analyzed
  dimensions from the pinned vector, a tracked thumbnail, filename FTS, and no invented rich-text
  row. Replacing it with video preserves the omitted body, changes fresh presentation/search,
  analyzes video metadata, and serves a preview. Detaching removes filename search and eventually
  purges each unshared original/preview/variant blob and file. The final file inventory and blob
  count equal the initial state; deleting the remaining message removes its FTS entry.
- Avatar null assignment retains the attachment; invalid assignment fails without changing bio
  or avatar. Multipart replacement with a different image purges the previous graph and changes
  conditional download bytes. Empty assignment purges the new graph and renders initials.
  HEAD omits the body, and a revoked session with a retained cookie and validator redirects to
  login rather than receiving a cached avatar or 304.

No production fix was needed. Storage workflows now have one owner in `bin/parity_storage.py`;
reports hash both checker files and retain failure tracebacks. All previous workflows still pass.
Required Go gate evidence is reused from `../storage-http-20261007/`: production inputs and the
application binary are unchanged. Ruff and Python compilation checks pass for the changed tools.

Initial exploratory assertions were corrected from independent evidence: video uses `poster`,
not an image `src`; the seed user's name is Jason, not Jason Fried; identical replacement bytes
can legitimately retain a content validator. The pinned Rust initials HEAD validator differs
from GET, so this flow validates HEAD's body and content type, not cross-method validator identity.

These are HTTP/persisted-state contracts, not browser, live Cable, arbitrary form metadata or
crash-durability proof. Background completion waits are bounded to ten seconds. No timing ran.
