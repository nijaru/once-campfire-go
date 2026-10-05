# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | records | 19,039 (19,026–19,497) | 0.564 | 1.454 | 5.791 | 145.6 |
| room_show | 16 | request | 20,797 (20,566–21,379) | 0.524 | 1.341 | 5.383 | 136.9 |
| room_show | 16 | shell | 22,152 (22,081–22,574) | 0.501 | 1.240 | 4.791 | 134.0 |
| active_room | 16 | records | 26,548 (25,481–26,598) | 0.490 | 0.958 | 3.019 | 130.8 |
| active_room | 16 | request | 27,692 (26,706–27,699) | 0.475 | 0.901 | 2.971 | 125.2 |
| active_room | 16 | shell | 27,950 (27,878–28,124) | 0.466 | 0.890 | 2.919 | 123.3 |
| messages_page | 16 | records | 29,797 (29,283–30,234) | 0.461 | 0.876 | 1.717 | 123.5 |
| messages_page | 16 | request | 31,009 (30,774–31,205) | 0.445 | 0.829 | 1.687 | 118.9 |
| messages_page | 16 | shell | 31,659 (31,087–31,728) | 0.441 | 0.818 | 1.557 | 117.9 |
| search | 16 | records | 15,395 (15,351–15,541) | 0.851 | 1.865 | 4.399 | 210.5 |
| search | 16 | request | 15,730 (15,504–15,811) | 0.827 | 1.869 | 4.263 | 208.0 |
| search | 16 | shell | 15,547 (15,492–15,730) | 0.829 | 1.889 | 4.519 | 209.1 |
| static_css | 16 | records | 104,368 (102,363–106,383) | 0.063 | 0.364 | 0.906 | 17.6 |
| static_css | 16 | request | 102,918 (100,604–104,901) | 0.062 | 0.375 | 0.932 | 17.6 |
| static_css | 16 | shell | 103,541 (102,682–104,767) | 0.063 | 0.366 | 0.907 | 17.6 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| records | not validated | — | 31.1 | 37.8 | 116.7 | 116.7 | 36.3 |
| request | not validated | — | 35.9 | 38.2 | 115.6 | 115.6 | 36.3 |
| shell | not validated | — | 61.7 | 38.3 | 116.5 | 116.5 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

9 completed application runs; 0 acknowledged HTTP posts and 1,253 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | records | 374,036 | 21289 | gzip |
| active_room | records | 374,036 | 21289 | gzip |
| messages_page | records | 342,444 | 12819 | gzip |
| search | records | 135,496 | 9463 | gzip |
| static_css | records | 1,218 | 654 | gzip |
| room_show | request | 374,036 | 21288 | gzip |
| active_room | request | 374,036 | 21288 | gzip |
| messages_page | request | 342,444 | 12819 | gzip |
| search | request | 135,496 | 9463 | gzip |
| static_css | request | 1,218 | 654 | gzip |
| room_show | shell | 374,036 | 21289 | gzip |
| active_room | shell | 374,036 | 21289 | gzip |
| messages_page | shell | 342,444 | 12819 | gzip |
| search | shell | 135,496 | 9463 | gzip |
| static_css | shell | 1,218 | 654 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
