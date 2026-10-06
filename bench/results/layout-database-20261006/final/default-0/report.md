# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | base | 15,298 (15,118–15,528) | 0.702 | 2.011 | 5.999 | 140.6 |
| room_show | 16 | shell | 16,100 (15,564–16,143) | 0.680 | 1.936 | 5.863 | 141.1 |
| room_show | 16 | rust | 19,244 (19,113–19,323) | 0.721 | 1.349 | 2.855 | 128.8 |
| active_room | 16 | base | 20,542 (20,442–20,756) | 0.617 | 1.237 | 3.859 | 127.1 |
| active_room | 16 | shell | 20,775 (20,400–20,869) | 0.613 | 1.272 | 3.789 | 126.2 |
| active_room | 16 | rust | 16,205 (16,000–16,456) | 0.860 | 1.629 | 3.009 | 136.5 |
| messages_page | 16 | base | 22,650 (22,136–23,026) | 0.548 | 1.123 | 3.807 | 119.7 |
| messages_page | 16 | shell | 22,957 (22,891–23,389) | 0.545 | 1.096 | 3.769 | 118.8 |
| messages_page | 16 | rust | 20,484 (20,260–20,862) | 0.678 | 1.276 | 2.459 | 119.4 |
| search | 16 | base | 19,255 (19,135–19,615) | 0.667 | 1.473 | 3.833 | 174.6 |
| search | 16 | shell | 26,225 (25,789–26,345) | 0.525 | 0.992 | 2.193 | 115.8 |
| search | 16 | rust | 26,484 (26,226–26,626) | 0.546 | 0.922 | 1.351 | 108.7 |
| active_search | 16 | base | 9,008 (8,946–9,504) | 1.383 | 3.369 | 6.847 | 289.6 |
| active_search | 16 | shell | 10,814 (10,796–10,962) | 1.120 | 2.769 | 5.903 | 224.0 |
| active_search | 16 | rust | 9,660 (9,638–9,772) | 1.464 | 2.653 | 4.191 | 270.8 |
| post_message | 16 | base | 2,322 (2,306–2,346) | 5.123 | 14.871 | 28.367 | 642.3 |
| post_message | 16 | shell | 2,329 (2,300–2,360) | 5.071 | 14.751 | 27.967 | 639.0 |
| post_message | 16 | rust | 3,212 (3,184–3,231) | 4.731 | 6.427 | 11.175 | 728.8 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| base | not validated | — | 57.9 | 37.4 | 134.0 | 134.0 | 36.3 |
| shell | not validated | — | 55.7 | 37.9 | 132.5 | 132.5 | 36.3 |
| rust | not validated | — | 70.7 | 41.3 | 111.5 | 111.5 | 35.2 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and identity encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

9 completed application runs; 164,486 acknowledged HTTP posts and 1,248 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | base | 374,036 | 374036 | identity |
| active_room | base | 374,036 | 374036 | identity |
| messages_page | base | 342,444 | 342444 | identity |
| search | base | 135,497 | 135497 | identity |
| active_search | base | 135,497 | 135497 | identity |
| room_show | shell | 374,036 | 374036 | identity |
| active_room | shell | 374,036 | 374036 | identity |
| messages_page | shell | 342,444 | 342444 | identity |
| search | shell | 135,497 | 135497 | identity |
| active_search | shell | 135,497 | 135497 | identity |
| room_show | rust | 416,139 | 416139 | identity |
| active_room | rust | 416,139 | 416139 | identity |
| messages_page | rust | 383,844 | 383844 | identity |
| search | rust | 149,625 | 149625 | identity |
| active_search | rust | 149,625 | 149625 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
