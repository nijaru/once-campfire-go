# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Opt-in message forms require complete document layouts and seed-grounded editor/attachment/ordered boost controls; selected applications also require exact decoded form bytes after normalization of only their listener origin. These are full-document GETs, not Turbo-Frame click measurements. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | base | 22,578 (22,243–22,906) | 0.484 | 1.200 | 5.267 | 127.3 |
| room_show | 16 | format | 22,765 (22,611–23,547) | 0.481 | 1.192 | 5.219 | 128.0 |
| room_show | 16 | rust | 28,436 (28,356–28,756) | 0.484 | 0.880 | 1.893 | 100.3 |
| active_room | 16 | base | 30,071 (29,753–30,393) | 0.441 | 0.806 | 2.445 | 113.2 |
| active_room | 16 | format | 30,393 (29,933–30,890) | 0.439 | 0.800 | 2.281 | 112.9 |
| active_room | 16 | rust | 26,820 (26,473–27,091) | 0.508 | 0.955 | 2.028 | 105.6 |
| messages_page | 16 | base | 32,715 (32,223–32,939) | 0.424 | 0.759 | 1.676 | 111.0 |
| messages_page | 16 | format | 32,461 (32,179–33,192) | 0.422 | 0.758 | 1.683 | 111.1 |
| messages_page | 16 | rust | 29,031 (28,875–29,438) | 0.473 | 0.863 | 1.816 | 93.0 |
| search | 16 | base | 27,192 (26,856–27,460) | 0.519 | 0.931 | 2.030 | 111.3 |
| search | 16 | format | 26,422 (26,389–26,936) | 0.532 | 0.954 | 1.926 | 112.3 |
| search | 16 | rust | 26,570 (26,551–26,740) | 0.548 | 0.896 | 1.277 | 105.9 |
| active_search | 16 | base | 16,748 (16,664–16,916) | 0.735 | 1.770 | 4.495 | 182.4 |
| active_search | 16 | format | 16,620 (16,600–16,886) | 0.753 | 1.766 | 4.363 | 183.5 |
| active_search | 16 | rust | 14,070 (14,014–14,197) | 1.072 | 1.624 | 2.299 | 197.5 |
| post_message | 16 | base | 2,442 (2,437–2,465) | 4.847 | 14.023 | 26.159 | 612.8 |
| post_message | 16 | format | 2,390 (2,384–2,426) | 4.971 | 14.231 | 26.735 | 616.8 |
| post_message | 16 | rust | 3,209 (3,172–3,256) | 4.719 | 6.407 | 11.879 | 753.8 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| base | not validated | — | 63.2 | 37.4 | 156.6 | 156.6 | 36.3 |
| format | not validated | — | 62.5 | 37.9 | 155.5 | 155.5 | 36.3 |
| rust | not validated | — | 42.8 | 41.3 | 134.3 | 134.3 | 35.2 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

9 completed application runs; 168,613 acknowledged HTTP posts and 1,251 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | base | 374,036 | 21289 | gzip |
| active_room | base | 374,036 | 21289 | gzip |
| messages_page | base | 342,444 | 12819 | gzip |
| search | base | 135,497 | 9462 | gzip |
| active_search | base | 135,497 | 9462 | gzip |
| room_show | format | 374,036 | 21289 | gzip |
| active_room | format | 374,036 | 21289 | gzip |
| messages_page | format | 342,444 | 12819 | gzip |
| search | format | 135,497 | 9462 | gzip |
| active_search | format | 135,497 | 9462 | gzip |
| room_show | rust | 416,139 | 24235 | gzip |
| active_room | rust | 416,139 | 24235 | gzip |
| messages_page | rust | 383,844 | 16158 | gzip |
| search | rust | 149,625 | 9766 | gzip |
| active_search | rust | 149,625 | 9766 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
