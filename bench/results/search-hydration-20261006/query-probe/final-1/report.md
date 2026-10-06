# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| search | 1 | baseline | 3,267 (3,245–3,330) | 0.210 | 0.276 | 3.327 | 233.7 |
| search | 1 | hit | 3,569 (3,522–3,621) | 0.190 | 0.262 | 3.263 | 211.5 |
| search | 1 | rust | 5,733 (5,647–5,758) | 0.175 | 0.190 | 0.206 | 148.3 |
| search | 16 | baseline | 10,512 (10,509–10,834) | 0.975 | 3.039 | 8.623 | 226.9 |
| search | 16 | hit | 11,880 (11,484–11,898) | 0.803 | 2.737 | 9.223 | 208.4 |
| search | 16 | rust | 24,956 (24,848–24,980) | 0.588 | 0.943 | 1.340 | 111.9 |
| search | 64 | baseline | 9,384 (9,362–9,708) | 5.947 | 11.503 | 17.951 | 234.7 |
| search | 64 | hit | 10,334 (10,330–10,625) | 5.235 | 10.775 | 17.503 | 215.7 |
| search | 64 | rust | 28,862 (28,774–29,304) | 2.071 | 2.959 | 4.395 | 101.3 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| baseline | not validated | — | 64.7 | 38.1 | 59.7 | 59.7 | 36.3 |
| hit | not validated | — | 61.8 | 38.1 | 58.3 | 58.3 | 36.3 |
| rust | not validated | — | 45.0 | 41.3 | 57.1 | 57.1 | 35.2 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

9 completed application runs; 0 acknowledged HTTP posts and 0 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| search | baseline | 135,497 | 9462 | gzip |
| search | hit | 135,497 | 9462 | gzip |
| search | rust | 149,625 | 9766 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
