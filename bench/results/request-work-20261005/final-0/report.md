# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | records | 13,334 (12,589–13,363) | 0.751 | 2.713 | 6.979 | 159.5 |
| room_show | 16 | request | 14,611 (14,583–14,907) | 0.718 | 2.159 | 6.571 | 147.0 |
| room_show | 16 | shell | 15,319 (15,137–15,426) | 0.695 | 2.014 | 6.263 | 142.5 |
| active_room | 16 | records | 18,254 (17,277–18,701) | 0.673 | 1.428 | 4.735 | 141.8 |
| active_room | 16 | request | 19,404 (19,078–19,574) | 0.649 | 1.319 | 4.407 | 134.4 |
| active_room | 16 | shell | 19,552 (19,172–20,256) | 0.636 | 1.301 | 4.307 | 130.3 |
| messages_page | 16 | records | 21,480 (20,960–21,495) | 0.578 | 1.167 | 4.319 | 128.3 |
| messages_page | 16 | request | 22,129 (21,887–22,919) | 0.558 | 1.131 | 4.147 | 121.5 |
| messages_page | 16 | shell | 22,535 (21,173–23,116) | 0.559 | 1.106 | 4.089 | 123.5 |
| search | 16 | records | 15,888 (13,990–15,929) | 0.831 | 1.782 | 4.131 | 209.7 |
| search | 16 | request | 15,426 (14,880–16,017) | 0.847 | 1.834 | 4.399 | 209.0 |
| search | 16 | shell | 15,638 (14,714–15,838) | 0.837 | 1.814 | 4.323 | 207.6 |
| static_css | 16 | records | 104,884 (104,132–105,710) | 0.063 | 0.362 | 0.896 | 17.5 |
| static_css | 16 | request | 104,170 (103,480–105,019) | 0.063 | 0.366 | 0.906 | 17.6 |
| static_css | 16 | shell | 103,670 (102,313–106,316) | 0.063 | 0.369 | 0.919 | 17.6 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| records | not validated | — | 62.4 | 37.9 | 121.8 | 121.8 | 36.3 |
| request | not validated | — | 62.4 | 38.3 | 109.0 | 109.0 | 36.3 |
| shell | not validated | — | 35.0 | 38.0 | 123.9 | 123.9 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and identity encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

9 completed application runs; 0 acknowledged HTTP posts and 1,244 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | records | 374,036 | 374036 | identity |
| active_room | records | 374,036 | 374036 | identity |
| messages_page | records | 342,444 | 342444 | identity |
| search | records | 135,496 | 135496 | identity |
| static_css | records | 1,218 | 1218 | identity |
| room_show | request | 374,036 | 374036 | identity |
| active_room | request | 374,036 | 374036 | identity |
| messages_page | request | 342,444 | 342444 | identity |
| search | request | 135,496 | 135496 | identity |
| static_css | request | 1,218 | 1218 | identity |
| room_show | shell | 374,036 | 374036 | identity |
| active_room | shell | 374,036 | 374036 | identity |
| messages_page | shell | 342,444 | 342444 | identity |
| search | shell | 135,496 | 135496 | identity |
| static_css | shell | 1,218 | 1218 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
