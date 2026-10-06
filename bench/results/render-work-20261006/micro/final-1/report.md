# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | base | 22,353 (22,051–22,586) | 0.490 | 1.219 | 5.239 | 129.5 |
| room_show | 16 | micro | 21,390 (21,377–22,473) | 0.495 | 1.279 | 5.395 | 129.2 |
| room_show | 16 | rust | 28,306 (28,075–28,306) | 0.489 | 0.892 | 1.894 | 100.1 |
| active_room | 16 | base | 29,423 (29,259–29,840) | 0.446 | 0.836 | 2.433 | 115.4 |
| active_room | 16 | micro | 29,236 (29,110–29,391) | 0.450 | 0.850 | 2.471 | 114.8 |
| active_room | 16 | rust | 26,512 (26,218–26,757) | 0.513 | 0.964 | 2.024 | 107.4 |
| messages_page | 16 | base | 31,919 (31,454–32,562) | 0.427 | 0.793 | 1.700 | 112.3 |
| messages_page | 16 | micro | 32,082 (31,620–32,111) | 0.430 | 0.775 | 1.699 | 112.3 |
| messages_page | 16 | rust | 29,026 (28,958–29,118) | 0.476 | 0.862 | 1.751 | 94.0 |
| search | 16 | base | 26,361 (26,330–28,350) | 0.531 | 0.967 | 1.982 | 112.9 |
| search | 16 | micro | 26,061 (26,052–26,498) | 0.532 | 0.961 | 2.129 | 113.3 |
| search | 16 | rust | 26,561 (26,303–26,672) | 0.549 | 0.900 | 1.287 | 107.5 |
| active_search | 16 | base | 16,326 (15,821–16,671) | 0.755 | 1.804 | 4.511 | 188.0 |
| active_search | 16 | micro | 16,436 (16,132–16,493) | 0.757 | 1.773 | 4.567 | 186.7 |
| active_search | 16 | rust | 13,938 (13,844–14,035) | 1.081 | 1.646 | 2.331 | 199.5 |
| post_message | 16 | base | 2,293 (2,241–2,293) | 5.215 | 14.935 | 27.519 | 669.9 |
| post_message | 16 | micro | 2,410 (2,389–2,413) | 4.911 | 14.359 | 26.591 | 625.8 |
| post_message | 16 | rust | 3,188 (3,179–3,242) | 4.747 | 6.487 | 10.815 | 758.2 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| base | not validated | — | 63.3 | 37.9 | 155.8 | 155.8 | 36.3 |
| micro | not validated | — | 59.0 | 38.0 | 155.5 | 155.5 | 36.3 |
| rust | not validated | — | 41.4 | 40.8 | 133.1 | 133.1 | 35.2 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

9 completed application runs; 164,861 acknowledged HTTP posts and 1,248 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | base | 374,036 | 21289 | gzip |
| active_room | base | 374,036 | 21289 | gzip |
| messages_page | base | 342,444 | 12819 | gzip |
| search | base | 135,497 | 9462 | gzip |
| active_search | base | 135,497 | 9462 | gzip |
| room_show | micro | 374,036 | 21289 | gzip |
| active_room | micro | 374,036 | 21289 | gzip |
| messages_page | micro | 342,444 | 12819 | gzip |
| search | micro | 135,497 | 9462 | gzip |
| active_search | micro | 135,497 | 9462 | gzip |
| room_show | rust | 416,139 | 24234 | gzip |
| active_room | rust | 416,139 | 24234 | gzip |
| messages_page | rust | 383,844 | 16158 | gzip |
| search | rust | 149,625 | 9766 | gzip |
| active_search | rust | 149,625 | 9766 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
