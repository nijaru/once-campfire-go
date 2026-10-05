# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 10.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | owned | 14,877 (13,292–14,879) | 0.726 | 2.119 | 6.383 | 147.7 |
| room_show | 16 | baseline | 14,644 (14,124–15,087) | 0.728 | 2.065 | 6.415 | 153.0 |
| active_room | 16 | owned | 19,579 (19,327–19,701) | 0.636 | 1.323 | 4.239 | 131.7 |
| active_room | 16 | baseline | 18,767 (17,535–18,961) | 0.660 | 1.395 | 4.419 | 134.7 |
| messages_page | 16 | owned | 21,994 (21,928–22,265) | 0.570 | 1.124 | 4.059 | 122.4 |
| messages_page | 16 | baseline | 20,903 (20,052–21,277) | 0.589 | 1.219 | 4.271 | 127.4 |
| search | 16 | owned | 15,276 (15,048–15,654) | 0.864 | 1.813 | 4.355 | 211.0 |
| search | 16 | baseline | 14,830 (14,683–14,842) | 0.879 | 1.927 | 4.659 | 216.3 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| owned | not validated | — | 61.9 | 38.0 | 118.1 | 118.1 | 36.3 |
| baseline | not validated | — | 61.7 | 37.9 | 107.9 | 107.9 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and identity encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 0 acknowledged HTTP posts and 1,416 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | owned | 374,036 | 374036 | identity |
| active_room | owned | 374,036 | 374036 | identity |
| messages_page | owned | 342,444 | 342444 | identity |
| search | owned | 135,496 | 135496 | identity |
| room_show | baseline | 374,036 | 374036 | identity |
| active_room | baseline | 374,036 | 374036 | identity |
| messages_page | baseline | 342,444 | 342444 | identity |
| search | baseline | 135,496 | 135496 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
