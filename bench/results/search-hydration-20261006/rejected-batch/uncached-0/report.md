# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | query | 323 (322–336) | 44.703 | 72.575 | 102.271 | 8886.9 |
| room_show | 16 | batch | 342 (338–364) | 41.887 | 74.943 | 121.215 | 8327.5 |
| room_show | 16 | rust | 2,402 (2,390–2,424) | 6.251 | 10.447 | 15.023 | 1187.7 |
| messages_page | 16 | query | 341 (318–343) | 42.463 | 70.015 | 99.775 | 8250.0 |
| messages_page | 16 | batch | 368 (368–372) | 38.047 | 70.527 | 109.439 | 7466.2 |
| messages_page | 16 | rust | 2,461 (2,459–2,467) | 6.135 | 10.023 | 14.343 | 1155.8 |
| search | 16 | query | 827 (818–844) | 15.711 | 34.399 | 56.511 | 3070.9 |
| search | 16 | batch | 829 (682–868) | 15.407 | 37.055 | 62.943 | 2880.5 |
| search | 16 | rust | 6,323 (6,300–6,326) | 2.351 | 3.601 | 5.543 | 472.7 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| query | not validated | — | 61.9 | 38.5 | 58.0 | 58.0 | 36.3 |
| batch | not validated | — | 58.5 | 38.2 | 57.7 | 57.7 | 36.3 |
| rust | not validated | — | 39.9 | 41.9 | 90.3 | 90.3 | 35.2 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and identity encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

9 completed application runs; 0 acknowledged HTTP posts and 0 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | query | 374,036 | 374036 | identity |
| messages_page | query | 342,444 | 342444 | identity |
| search | query | 135,497 | 135497 | identity |
| room_show | batch | 374,036 | 374036 | identity |
| messages_page | batch | 342,444 | 342444 | identity |
| search | batch | 135,497 | 135497 | identity |
| room_show | rust | 416,139 | 416139 | identity |
| messages_page | rust | 383,844 | 383844 | identity |
| search | rust | 149,625 | 149625 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
