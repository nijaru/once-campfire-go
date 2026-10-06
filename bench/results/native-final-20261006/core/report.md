# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 30.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0,2,4,6; load generator on CPUs 8,10,12,14. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Opt-in message forms require complete document layouts and seed-grounded editor/attachment/ordered boost controls; selected applications also require exact decoded form bytes after normalization of only their listener origin. These are full-document GETs, not Turbo-Frame click measurements. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | final | 32,348 (32,336–32,391) | 0.414 | 0.891 | 1.532 | 116.7 |
| room_show | 16 | rust | 29,693 (29,578–29,749) | 0.524 | 0.740 | 0.947 | 129.1 |
| active_room | 16 | final | 37,429 (37,265–37,460) | 0.374 | 0.709 | 1.522 | 102.0 |
| active_room | 16 | rust | 28,057 (27,953–28,063) | 0.550 | 0.789 | 1.064 | 136.6 |
| messages_page | 16 | final | 39,671 (39,606–39,696) | 0.351 | 0.725 | 1.213 | 97.3 |
| messages_page | 16 | rust | 31,869 (31,441–31,965) | 0.491 | 0.669 | 0.836 | 119.5 |
| search | 16 | final | 31,090 (30,950–31,226) | 0.464 | 0.846 | 1.452 | 121.2 |
| search | 16 | rust | 25,459 (25,435–25,553) | 0.584 | 0.899 | 1.264 | 128.0 |
| active_search | 16 | final | 12,267 (12,246–12,624) | 1.146 | 2.221 | 5.667 | 278.5 |
| active_search | 16 | rust | 8,517 (8,334–8,519) | 1.759 | 2.975 | 4.093 | 424.8 |
| post_message | 16 | final | 3,718 (3,495–4,550) | 2.391 | 10.735 | 21.135 | 613.0 |
| post_message | 16 | rust | 5,670 (5,500–5,802) | 2.455 | 3.705 | 11.543 | 501.4 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| final | not validated | — | 51.2 | 39.8 | 210.3 | 210.3 | 37.8 |
| rust | not validated | — | 38.4 | 44.1 | 150.2 | 150.2 | 35.6 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 913,923 acknowledged HTTP posts and 3,840 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | final | 374,036 | 21289 | gzip |
| active_room | final | 374,036 | 21289 | gzip |
| messages_page | final | 342,444 | 12819 | gzip |
| search | final | 135,497 | 9462 | gzip |
| active_search | final | 135,497 | 9462 | gzip |
| room_show | rust | 416,139 | 24235 | gzip |
| active_room | rust | 416,139 | 24235 | gzip |
| messages_page | rust | 383,844 | 16158 | gzip |
| search | rust | 149,625 | 9766 | gzip |
| active_search | rust | 149,625 | 9766 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
