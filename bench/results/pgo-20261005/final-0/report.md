# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | owned | 14,535 (14,534–14,973) | 0.720 | 2.197 | 6.363 | 145.0 |
| room_show | 16 | pgo | 14,650 (14,550–14,823) | 0.709 | 2.169 | 6.631 | 141.3 |
| active_room | 16 | owned | 20,062 (19,120–20,185) | 0.632 | 1.272 | 4.291 | 130.4 |
| active_room | 16 | pgo | 19,827 (19,126–20,343) | 0.625 | 1.297 | 4.291 | 126.9 |
| messages_page | 16 | owned | 20,856 (20,803–21,790) | 0.588 | 1.219 | 4.207 | 125.7 |
| messages_page | 16 | pgo | 21,639 (21,342–22,236) | 0.571 | 1.142 | 4.227 | 121.8 |
| search | 16 | owned | 15,410 (15,221–15,469) | 0.845 | 1.831 | 4.319 | 211.1 |
| search | 16 | pgo | 15,834 (15,550–16,034) | 0.813 | 1.819 | 4.563 | 201.8 |
| static_css | 16 | owned | 102,927 (102,892–103,667) | 0.063 | 0.370 | 0.924 | 17.8 |
| static_css | 16 | pgo | 102,998 (102,350–103,606) | 0.064 | 0.374 | 0.947 | 17.7 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| owned | not validated | — | 62.0 | 38.1 | 111.9 | 111.9 | 36.3 |
| pgo | not validated | — | 63.7 | 37.7 | 112.4 | 112.4 | 36.7 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and identity encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 0 acknowledged HTTP posts and 828 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | owned | 374,036 | 374036 | identity |
| active_room | owned | 374,036 | 374036 | identity |
| messages_page | owned | 342,444 | 342444 | identity |
| search | owned | 135,496 | 135496 | identity |
| static_css | owned | 1,218 | 1218 | identity |
| room_show | pgo | 374,036 | 374036 | identity |
| active_room | pgo | 374,036 | 374036 | identity |
| messages_page | pgo | 342,444 | 342444 | identity |
| search | pgo | 135,496 | 135496 | identity |
| static_css | pgo | 1,218 | 1218 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
