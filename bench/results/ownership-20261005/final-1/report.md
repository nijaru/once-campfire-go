# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | baseline | 21,782 (21,386–22,050) | 0.506 | 1.284 | 5.171 | 133.7 |
| room_show | 16 | owned | 21,465 (21,216–22,048) | 0.512 | 1.290 | 5.267 | 134.4 |
| active_room | 16 | baseline | 27,148 (26,747–28,124) | 0.472 | 0.909 | 3.183 | 122.7 |
| active_room | 16 | owned | 28,708 (28,310–29,171) | 0.465 | 0.873 | 2.401 | 120.6 |
| messages_page | 16 | baseline | 31,181 (30,066–31,355) | 0.445 | 0.832 | 1.699 | 118.1 |
| messages_page | 16 | owned | 31,074 (30,991–31,679) | 0.439 | 0.831 | 1.700 | 117.5 |
| search | 16 | baseline | 15,955 (15,782–16,036) | 0.824 | 1.837 | 4.163 | 205.9 |
| search | 16 | owned | 15,957 (15,625–16,063) | 0.816 | 1.821 | 4.219 | 206.4 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| baseline | not validated | — | 33.8 | 38.3 | 116.4 | 116.4 | 36.3 |
| owned | not validated | — | 61.6 | 38.4 | 123.2 | 123.2 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 0 acknowledged HTTP posts and 834 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | baseline | 374,036 | 21289 | gzip |
| active_room | baseline | 374,036 | 21289 | gzip |
| messages_page | baseline | 342,444 | 12819 | gzip |
| search | baseline | 135,496 | 9463 | gzip |
| room_show | owned | 374,036 | 21289 | gzip |
| active_room | owned | 374,036 | 21289 | gzip |
| messages_page | owned | 342,444 | 12819 | gzip |
| search | owned | 135,496 | 9463 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
