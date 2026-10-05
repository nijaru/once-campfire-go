# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | baseline | 15,124 (14,942–15,162) | 0.723 | 2.000 | 6.143 | 145.6 |
| room_show | 16 | owned | 14,687 (14,504–14,700) | 0.749 | 2.133 | 6.351 | 147.7 |
| active_room | 16 | baseline | 19,742 (19,661–19,947) | 0.636 | 1.306 | 4.279 | 131.4 |
| active_room | 16 | owned | 19,107 (18,760–19,695) | 0.658 | 1.372 | 4.251 | 132.5 |
| messages_page | 16 | baseline | 22,306 (21,630–22,564) | 0.571 | 1.116 | 4.039 | 122.7 |
| messages_page | 16 | owned | 21,467 (21,351–21,932) | 0.571 | 1.173 | 4.147 | 124.3 |
| search | 16 | baseline | 15,882 (15,746–15,970) | 0.832 | 1.798 | 4.175 | 208.8 |
| search | 16 | owned | 15,326 (15,143–15,485) | 0.851 | 1.848 | 4.411 | 210.4 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| baseline | not validated | — | 41.4 | 38.7 | 123.6 | 123.6 | 36.3 |
| owned | not validated | — | 54.7 | 38.1 | 112.8 | 112.8 | 36.3 |

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
| room_show | baseline | 374,036 | 374036 | identity |
| active_room | baseline | 374,036 | 374036 | identity |
| messages_page | baseline | 342,444 | 342444 | identity |
| search | baseline | 135,496 | 135496 | identity |
| room_show | owned | 374,036 | 374036 | identity |
| active_room | owned | 374,036 | 374036 | identity |
| messages_page | owned | 342,444 | 342444 | identity |
| search | owned | 135,496 | 135496 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
