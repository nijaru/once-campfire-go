# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| post_message | 16 | base | 2,275 (2,268–2,284) | 5.259 | 15.023 | 28.127 | 675.1 |
| post_message | 16 | index | 2,284 (2,265–2,316) | 5.243 | 14.991 | 27.871 | 670.6 |
| post_message | 16 | dense | 2,165 (2,123–2,262) | 5.443 | 15.751 | 30.799 | 715.5 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|
| 1,000 | False | base | 299.1 (297.7–306.1) | 299,116 | 49.759 | 76.415 | not measured |
| 1,000 | False | index | 316.5 (309.3–317.0) | 316,455 | 46.303 | 74.495 | not measured |
| 1,000 | False | dense | 318.4 (307.8–320.7) | 318,392 | 50.431 | 78.015 | not measured |
| 1,000 | True | base | 224.0 (220.5–224.9) | 223,982 | 56.255 | 83.135 | 595.9 |
| 1,000 | True | index | 226.6 (212.5–233.2) | 226,578 | 62.175 | 94.271 | 603.1 |
| 1,000 | True | dense | 227.8 (223.9–231.4) | 227,790 | 58.431 | 84.287 | 606.0 |
| 10,000 | False | base | 26.7 (25.8–26.8) | 267,250 | 248.447 | 469.759 | not measured |
| 10,000 | False | index | 34.0 (30.9–34.4) | 340,497 | 342.527 | 656.383 | not measured |
| 10,000 | False | dense | 33.8 (33.3–37.0) | 337,858 | 293.887 | 399.871 | not measured |
| 10,000 | True | base | 25.3 (22.6–25.4) | 252,699 | 382.463 | 470.271 | 671.0 |
| 10,000 | True | index | 26.0 (25.9–26.4) | 260,290 | 617.983 | 743.935 | 691.5 |
| 10,000 | True | dense | 26.0 (25.8–26.5) | 260,115 | 512.767 | 631.295 | 690.8 |

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| base | not validated | — | 58.8 | 38.7 | 134.9 | 938.6 | 36.3 |
| index | not validated | — | 57.6 | 38.6 | 141.1 | 956.1 | 36.3 |
| dense | not validated | — | 57.7 | 39.0 | 138.9 | 938.2 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

9 completed application runs; 140,244 acknowledged HTTP posts and 0 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Incomplete throughput deliveries: 0.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
