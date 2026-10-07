# Video-preview vector environment investigation

The optional exact media-vector failure is reproducible outside Go and Rust. The two existing
local ARM64 images, `campfire-compare-toolchain` and `campfire-reference`, run the same pinned
preview command on the same fixture and both produce the identical 11,790-byte JPEG:

- Actual MD5: `Ckn6FYF3wfdTi3GYIHXgbg==`
- Stored vector: 11,788 bytes, MD5 `untJ1VKBeYvbPA/dvRXh9A==`

Full version/configuration, executable hash, linked libraries and image identities are retained
in the compressed logs. Go (`internal/storage/preview.go`) and Rust
(`reference/crates/storage/src/process.rs`, `content_types.rs`) use the same ffmpeg arguments.
Neither reencodes the generated JPEG during staging or analysis. Both Rails-vector generation
and the optional tests compare ffmpeg's first version line, not the complete build/runtime.
The vector's original image digest and architecture were not recorded.

Decoding actual and expected JPEGs with the same tool produces 172,800 RGB bytes each. Of those,
21,812 channel values differ, with a maximum difference of 5. This is not merely a metadata or
filename difference. A single-thread diagnostic reproduces the default output. Disabling CPU
features (with a single encoder thread) produces a third output, 11,785 bytes, not the vector.
These flags are diagnostics only; no application flags, vectors or assertions were changed.

The cause of the original vector's exact encoding remains unresolved. Current Rails and port
runtime agreement rules out a Go staging regression; it does not establish equivalence to the
unrecorded generator environment. The optional `media_vectors` test still fails, and earlier
failures remain recorded in `../storage-staging-20261007/`. This investigation used existing
images with network access disabled; it made no performance or cross-architecture claim.
