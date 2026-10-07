# Transactional derivative ownership

Cancellation at the SQLite attachment-link boundary previously left an unattached derivative
blob and file. The new regression reproduces this for both an image variant and a video preview
using real media processing and SQLite. A test-local SQLite function cancels the request as the
link is inserted; no production hook or detached-context cleanup is used.

Derivatives now remain staged through metadata analysis. The blob, variant record (when needed)
and attachment commit in one transaction. Files are retained only after commit; cancelled,
failed or losing transactions discard their staged files without needing database cleanup.
Preview linking also verifies that the source blob still exists inside the transaction, because
its polymorphic attachment has no foreign key to the source record. Existing blob analysis still
updates metadata and touches attached records transactionally.

The regression fails on `912e443` with two blobs/files instead of one. It passes ten race-enabled
repetitions after the fix. Concurrent variant reuse also checks the complete graph and absence
of losing files. Native and Linux CGO vet/race checks and the WebSocket-fork gates pass.

An additional Linux `media_vectors` check fails on **both** the unchanged baseline and candidate:
`alpha-centuri-preview_image.jpg` is 11,790 bytes with MD5 `Ckn6FYF3wfdTi3GYIHXgbg==`, while the
stored vector is 11,788 bytes with MD5 `untJ1VKBeYvbPA/dvRXh9A==`. The output is identical before
and after this change; the environment mismatch was unresolved at this checkpoint. The subsequent
[verification](../acceptance-corrections-20261007/README.md) reproduces the unchanged golden with
AMD64 FFmpeg and passes both original and separate ARM64 Rails comparisons. No expected bytes were changed. The initial baseline attempt lacked the reference-fixture symlink and was
repeated with it in place; the retained baseline log is the complete repeat.

No performance measurements or crash-durability claims were made. This removes an unnecessary
derivative metadata transaction but does not establish a throughput improvement.
