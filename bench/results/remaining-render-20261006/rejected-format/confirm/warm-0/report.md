# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Opt-in message forms require complete document layouts and seed-grounded editor/attachment/ordered boost controls; selected applications also require exact decoded form bytes after normalization of only their listener origin. These are full-document GETs, not Turbo-Frame click measurements. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | base | 15,838 (15,829–16,268) | 0.642 | 2.022 | 6.039 | 136.8 |
| room_show | 16 | format | 16,295 (16,260–16,356) | 0.638 | 1.936 | 5.895 | 132.8 |
| room_show | 16 | rust | 21,240 (19,839–21,413) | 0.676 | 1.143 | 2.473 | 128.2 |
| active_room | 16 | base | 21,655 (21,191–21,879) | 0.595 | 1.124 | 4.053 | 120.3 |
| active_room | 16 | format | 21,815 (21,740–21,904) | 0.587 | 1.131 | 4.007 | 120.5 |
| active_room | 16 | rust | 17,452 (17,416–17,480) | 0.806 | 1.498 | 2.819 | 131.8 |
| messages_page | 16 | base | 23,831 (23,595–23,849) | 0.527 | 1.005 | 4.059 | 115.8 |
| messages_page | 16 | format | 23,914 (23,906–24,292) | 0.524 | 0.998 | 4.079 | 114.6 |
| messages_page | 16 | rust | 22,041 (21,945–22,051) | 0.638 | 1.128 | 2.401 | 117.2 |
| search | 16 | base | 26,715 (26,375–26,738) | 0.523 | 0.947 | 2.047 | 112.4 |
| search | 16 | format | 26,637 (26,276–27,271) | 0.523 | 0.966 | 2.007 | 113.1 |
| search | 16 | rust | 26,574 (26,450–26,719) | 0.545 | 0.916 | 1.341 | 108.4 |
| active_search | 16 | base | 11,476 (11,433–11,627) | 1.067 | 2.561 | 5.603 | 216.1 |
| active_search | 16 | format | 11,494 (11,316–11,515) | 1.075 | 2.585 | 5.703 | 215.3 |
| active_search | 16 | rust | 10,382 (10,323–10,402) | 1.403 | 2.381 | 3.663 | 259.7 |
| post_message | 16 | base | 2,479 (2,456–2,480) | 4.799 | 13.935 | 25.791 | 583.9 |
| post_message | 16 | format | 2,500 (2,474–2,522) | 4.791 | 13.727 | 25.391 | 575.3 |
| post_message | 16 | rust | 3,278 (3,263–3,282) | 4.595 | 6.275 | 11.447 | 723.2 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| base | not validated | — | 63.0 | 37.9 | 134.4 | 134.4 | 36.3 |
| format | not validated | — | 64.6 | 37.7 | 134.8 | 134.8 | 36.3 |
| rust | not validated | — | 72.2 | 41.2 | 110.2 | 110.2 | 35.2 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and identity encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

9 completed application runs; 172,230 acknowledged HTTP posts and 1,249 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | base | 374,036 | 374036 | identity |
| active_room | base | 374,036 | 374036 | identity |
| messages_page | base | 342,444 | 342444 | identity |
| search | base | 135,497 | 135497 | identity |
| active_search | base | 135,497 | 135497 | identity |
| room_show | format | 374,036 | 374036 | identity |
| active_room | format | 374,036 | 374036 | identity |
| messages_page | format | 342,444 | 342444 | identity |
| search | format | 135,497 | 135497 | identity |
| active_search | format | 135,497 | 135497 | identity |
| room_show | rust | 416,139 | 416139 | identity |
| active_room | rust | 416,139 | 416139 | identity |
| messages_page | rust | 383,844 | 383844 | identity |
| search | rust | 149,625 | 149625 | identity |
| active_search | rust | 149,625 | 149625 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
