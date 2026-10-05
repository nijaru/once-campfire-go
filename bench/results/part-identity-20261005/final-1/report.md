# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | prepared | 8,172 (8,099–8,653) | 1.315 | 3.619 | 11.671 | 361.0 |
| room_show | 16 | parts | 10,165 (9,686–10,362) | 0.920 | 3.843 | 10.223 | 218.8 |
| active_room | 16 | prepared | 11,818 (11,738–11,880) | 0.895 | 2.729 | 5.627 | 315.0 |
| active_room | 16 | parts | 18,884 (18,019–19,007) | 0.616 | 1.585 | 4.707 | 179.2 |
| messages_page | 16 | prepared | 15,009 (14,972–15,214) | 0.750 | 2.103 | 4.073 | 257.0 |
| messages_page | 16 | parts | 27,320 (26,115–28,051) | 0.482 | 1.004 | 2.040 | 132.9 |
| sidebar | 16 | prepared | 15,432 (15,150–18,948) | 0.890 | 1.715 | 3.723 | 152.6 |
| sidebar | 16 | parts | 17,160 (15,858–17,754) | 0.797 | 1.493 | 3.711 | 148.6 |
| search | 16 | prepared | 13,943 (13,652–14,199) | 0.925 | 2.233 | 4.455 | 257.3 |
| search | 16 | parts | 15,391 (15,073–15,394) | 0.841 | 1.908 | 4.723 | 211.2 |
| static_css | 16 | prepared | 103,830 (102,132–105,278) | 0.061 | 0.372 | 0.926 | 17.5 |
| static_css | 16 | parts | 102,026 (100,914–104,378) | 0.063 | 0.369 | 0.950 | 17.6 |
| post_message | 16 | prepared | 2,263 (2,255–2,272) | 5.247 | 15.151 | 28.799 | 679.1 |
| post_message | 16 | parts | 2,287 (2,266–2,312) | 5.259 | 14.823 | 28.223 | 667.4 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| prepared | not validated | — | 58.1 | 38.0 | 149.2 | 149.2 | 36.3 |
| parts | not validated | — | 63.1 | 37.9 | 150.0 | 150.0 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 95,513 acknowledged HTTP posts and 836 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | prepared | 374,036 | 21289 | gzip |
| active_room | prepared | 374,036 | 21289 | gzip |
| messages_page | prepared | 342,444 | 12819 | gzip |
| sidebar | prepared | 9,462 | 2249 | gzip |
| search | prepared | 135,497 | 9462 | gzip |
| static_css | prepared | 1,218 | 654 | gzip |
| room_show | parts | 374,036 | 21289 | gzip |
| active_room | parts | 374,036 | 21289 | gzip |
| messages_page | parts | 342,444 | 12819 | gzip |
| sidebar | parts | 9,462 | 2249 | gzip |
| search | parts | 135,497 | 9462 | gzip |
| static_css | parts | 1,218 | 654 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
