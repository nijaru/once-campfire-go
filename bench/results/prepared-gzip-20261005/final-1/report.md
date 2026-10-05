# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | upstream | 4,145 (4,088–4,186) | 2.959 | 7.355 | 13.111 | 930.6 |
| room_show | 16 | cursor | 4,123 (4,044–4,192) | 2.897 | 7.491 | 13.175 | 937.4 |
| room_show | 16 | prepared | 8,077 (8,068–8,238) | 1.313 | 3.691 | 12.455 | 361.8 |
| active_room | 16 | upstream | 4,415 (4,393–4,439) | 2.629 | 6.807 | 12.103 | 883.8 |
| active_room | 16 | cursor | 4,391 (4,388–4,453) | 2.653 | 6.807 | 12.439 | 880.6 |
| active_room | 16 | prepared | 11,665 (11,513–11,822) | 0.904 | 2.763 | 5.767 | 316.9 |
| messages_page | 16 | upstream | 5,334 (5,297–5,355) | 2.207 | 5.775 | 8.919 | 739.3 |
| messages_page | 16 | cursor | 5,386 (5,375–5,387) | 2.207 | 5.755 | 8.719 | 737.4 |
| messages_page | 16 | prepared | 14,965 (14,814–15,053) | 0.753 | 2.177 | 4.055 | 258.1 |
| sidebar | 16 | upstream | 15,430 (14,335–15,883) | 0.879 | 1.700 | 4.155 | 182.6 |
| sidebar | 16 | cursor | 15,412 (15,115–15,553) | 0.880 | 1.714 | 4.171 | 182.7 |
| sidebar | 16 | prepared | 15,030 (14,802–15,901) | 0.914 | 1.769 | 3.969 | 152.9 |
| search | 16 | upstream | 7,922 (7,832–7,942) | 1.411 | 4.243 | 7.051 | 481.3 |
| search | 16 | cursor | 7,925 (7,912–7,965) | 1.402 | 4.195 | 7.355 | 479.6 |
| search | 16 | prepared | 13,642 (13,315–13,775) | 0.932 | 2.271 | 4.855 | 258.7 |
| static_css | 16 | upstream | 103,063 (103,042–103,662) | 0.062 | 0.368 | 0.928 | 17.6 |
| static_css | 16 | cursor | 106,183 (102,994–106,273) | 0.062 | 0.359 | 0.900 | 17.4 |
| static_css | 16 | prepared | 103,147 (101,471–105,688) | 0.063 | 0.374 | 0.938 | 17.6 |
| post_message | 16 | upstream | 2,229 (2,217–2,276) | 5.291 | 15.359 | 28.735 | 688.0 |
| post_message | 16 | cursor | 2,251 (2,231–2,264) | 5.279 | 15.055 | 28.927 | 686.4 |
| post_message | 16 | prepared | 2,280 (2,229–2,291) | 5.287 | 14.855 | 28.207 | 681.4 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|
| 100 | False | upstream | 1,530.4 (1,396.3–1,531.4) | 153,044 | 9.551 | 12.631 | not measured |
| 100 | False | cursor | 1,522.6 (1,483.4–1,528.1) | 152,263 | 9.639 | 13.279 | not measured |
| 100 | False | prepared | 1,528.6 (1,519.5–1,530.8) | 152,863 | 9.279 | 12.591 | not measured |
| 100 | True | upstream | 1,389.1 (1,384.5–1,405.9) | 138,913 | 10.215 | 12.799 | 370.0 |
| 100 | True | cursor | 1,364.5 (1,359.7–1,375.7) | 136,446 | 10.463 | 13.383 | 363.3 |
| 100 | True | prepared | 1,398.8 (1,376.3–1,423.2) | 139,883 | 10.223 | 12.839 | 372.5 |

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| upstream | not validated | — | 53.1 | 38.1 | 138.7 | 145.3 | 36.2 |
| cursor | not validated | — | 58.0 | 38.0 | 141.5 | 144.4 | 36.2 |
| prepared | not validated | — | 62.9 | 38.1 | 146.7 | 150.0 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

9 completed application runs; 141,725 acknowledged HTTP posts and 1,246 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Incomplete throughput deliveries: 0.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | upstream | 374,036 | 21288 | gzip |
| active_room | upstream | 374,036 | 21288 | gzip |
| messages_page | upstream | 342,444 | 12819 | gzip |
| sidebar | upstream | 9,462 | 2249 | gzip |
| search | upstream | 135,497 | 9462 | gzip |
| static_css | upstream | 1,218 | 654 | gzip |
| room_show | cursor | 374,036 | 21289 | gzip |
| active_room | cursor | 374,036 | 21289 | gzip |
| messages_page | cursor | 342,444 | 12819 | gzip |
| sidebar | cursor | 9,462 | 2249 | gzip |
| search | cursor | 135,497 | 9462 | gzip |
| static_css | cursor | 1,218 | 654 | gzip |
| room_show | prepared | 374,036 | 21289 | gzip |
| active_room | prepared | 374,036 | 21289 | gzip |
| messages_page | prepared | 342,444 | 12819 | gzip |
| sidebar | prepared | 9,462 | 2249 | gzip |
| search | prepared | 135,497 | 9462 | gzip |
| static_css | prepared | 1,218 | 654 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
