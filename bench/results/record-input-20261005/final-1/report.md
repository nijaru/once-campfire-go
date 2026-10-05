# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | keys | 16,198 (15,688–16,430) | 0.623 | 1.742 | 6.787 | 156.3 |
| room_show | 16 | records | 17,907 (17,694–18,278) | 0.572 | 1.543 | 6.255 | 147.2 |
| active_room | 16 | keys | 24,348 (23,575–25,332) | 0.511 | 1.080 | 3.473 | 137.1 |
| active_room | 16 | records | 25,362 (25,070–25,940) | 0.497 | 1.005 | 3.327 | 132.8 |
| messages_page | 16 | keys | 28,483 (28,084–28,648) | 0.471 | 0.931 | 2.016 | 127.0 |
| messages_page | 16 | records | 28,979 (28,712–29,730) | 0.463 | 0.904 | 2.013 | 123.7 |
| search | 16 | keys | 15,001 (14,914–15,525) | 0.850 | 1.920 | 4.559 | 214.0 |
| search | 16 | records | 15,403 (15,397–15,460) | 0.851 | 1.880 | 4.395 | 210.4 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| keys | not validated | — | 63.0 | 38.2 | 114.5 | 114.5 | 36.3 |
| records | not validated | — | 62.0 | 37.8 | 114.9 | 114.9 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 0 acknowledged HTTP posts and 835 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | keys | 374,036 | 21289 | gzip |
| active_room | keys | 374,036 | 21289 | gzip |
| messages_page | keys | 342,444 | 12819 | gzip |
| search | keys | 135,497 | 9462 | gzip |
| room_show | records | 374,036 | 21288 | gzip |
| active_room | records | 374,036 | 21288 | gzip |
| messages_page | records | 342,444 | 12819 | gzip |
| search | records | 135,496 | 9463 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
