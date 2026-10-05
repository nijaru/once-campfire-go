# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | shell | 15,215 (14,742–15,462) | 0.650 | 1.887 | 7.487 | 165.3 |
| room_show | 16 | keys | 16,769 (16,628–16,945) | 0.615 | 1.648 | 6.523 | 154.1 |
| active_room | 16 | shell | 23,760 (23,662–24,662) | 0.523 | 1.099 | 3.587 | 142.1 |
| active_room | 16 | keys | 25,449 (25,376–25,496) | 0.496 | 1.016 | 3.341 | 133.1 |
| messages_page | 16 | shell | 27,448 (25,898–27,562) | 0.492 | 1.005 | 2.077 | 132.8 |
| messages_page | 16 | keys | 28,308 (28,214–28,822) | 0.467 | 0.948 | 2.135 | 125.2 |
| search | 16 | shell | 15,432 (15,266–15,937) | 0.842 | 1.909 | 4.559 | 211.4 |
| search | 16 | keys | 15,604 (15,556–15,606) | 0.846 | 1.843 | 4.211 | 209.6 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| shell | not validated | — | 35.7 | 37.8 | 119.8 | 119.8 | 36.3 |
| keys | not validated | — | 66.1 | 37.7 | 115.9 | 115.9 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 0 acknowledged HTTP posts and 833 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | shell | 374,036 | 21289 | gzip |
| active_room | shell | 374,036 | 21289 | gzip |
| messages_page | shell | 342,444 | 12819 | gzip |
| search | shell | 135,497 | 9462 | gzip |
| room_show | keys | 374,036 | 21289 | gzip |
| active_room | keys | 374,036 | 21289 | gzip |
| messages_page | keys | 342,444 | 12819 | gzip |
| search | keys | 135,497 | 9462 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
