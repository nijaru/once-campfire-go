# Compression receipt/representation probe — local, 2026-10-12

**The matched receipt probe does not support a payload or finite-buffer change
as the explanation for the larger native compression samples.** No implementation
change or application performance gain is promoted.

The [native attribution](../post-cache-observation-native-20261012/README.md)
samples 10.41 s in fork gzip preparation versus 4.57 s in upstream gzip closure,
with nearly flat total CPU. Source inspection finds the same pooled level-6 gzip
writer in both applications. Fork finite output compresses before emission;
upstream closes its streaming writer afterward. Stack placement differs, but
that alone cannot explain absolute compression work.

## Bounded probe and observations

One sequential local capture pair uses eight deterministic posts per arm, matched
seed `691c52bfdc30b38bd8867e93b0914613df20e03d4cd7b00796ab68e0a5d26c9a`, four
workers/readers, three jobs, default GC and Go 1.27.1 darwin/arm64. Binaries match
the previously verified `8351224`/`c4ab53d` binary hashes in the
[local attribution record](../post-cache-observation-20261012/README.md).
Authentication occurs before captures. Requests explicitly accept gzip and complete
Turbo receipts; all **16** successful posts pass exact ID/body/client-ID/creator/
room/FTS, integrity and unchanged seed-FK checks. Each arm uses a disposable SQLite
backup and storage copy; production data is not modified.

- Every decoded receipt is 8,363 bytes; compressed members are 1,913–1,916 bytes.
- Corresponding decoded bodies differ only in creation/update timestamps and the
  localhost origin port. Normalizing those fields makes all eight pairs identical.
- All **16 gzip members match byte-for-byte** with standalone level-6 recompression
  of their unmodified captured body, including headers and trailer.

A standalone replay runs three interleaved pairs, 10,000 bodies per arm/mode
(120,000 compression operations, **not additional persisted writes**). It compares
new finite output buffers with a discard streaming sink, using the same writer
reset/close policy. OS process CPU is measured with `getrusage` within each replay
phase; it is not a sampled-profile percentage.

| Median local process CPU/body | Current payloads | Upstream payloads |
|---|---:|---:|
| Finite memory output | 31.38 µs | 31.40 µs |
| Streaming discard sink | 30.79 µs | 30.75 µs |

Payload-set compression work is effectively equal. Finite buffering adds roughly
2% in this isolated comparison—far smaller than the native sampled-owner ratio.
These findings close this probe's payload/representation premise; they do not
prove the native profile unbiased or explain the native POST tail gap. A native
absolute-work measurement is still needed before selecting a compression change.

## Limits and retained failure

This is a small matched-body probe, not a native application benchmark. Its
`compression-probe-N` client IDs are deterministic, shorter and less entropic than
the fixed-load client's random `rate-<26-character token>-N` IDs. Creation times,
sequence widths and population also differ from a 66,000-write profile. Replay is
single-goroutine, uses captured bodies rather than application rendering and
excludes network emission; a discard sink is not the real upstream writer.
Neither the near-equality nor the 2% buffer delta establishes general native cost.
Do not use them to claim improved latency/capacity, pick a gzip level blindly, or
relax complete output, authorization, cache budgets or joined lifecycles.

The first capture attempt asserts a guessed `id="message_N"` DOM marker and fails
after one POST. The existing shared-compatible ACK parser establishes the actual
contract, `data-message-id="N"`; correcting the probe preserves every persistence
and receipt assertion. That incomplete attempt is excluded from the 16 audited
captures; its error is retained. It is not a product defect or a successful extra
comparison. No application source or permanent test changes result.

The manifest hashes exact capture/replay programs, unmodified captured bodies and
gzip members, differences, identities, raw replay samples and the failed assertion.
Runtime cookies, databases and application request logs are excluded. No PR is
prepared or opened.
