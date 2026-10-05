# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | go-before | 8,422 (7,990–8,721) | 0.977 | 5.231 | 10.871 | 228.9 |
| room_show | 16 | go | 8,382 (8,126–9,302) | 0.997 | 5.039 | 10.759 | 226.8 |
| active_room | 16 | go-before | 15,488 (14,798–16,656) | 0.720 | 1.805 | 5.703 | 188.6 |
| active_room | 16 | go | 16,126 (15,294–16,220) | 0.693 | 1.725 | 5.831 | 186.9 |
| messages_page | 16 | go-before | 20,658 (20,098–21,708) | 0.572 | 1.232 | 4.463 | 137.9 |
| messages_page | 16 | go | 21,081 (19,694–21,393) | 0.575 | 1.210 | 4.459 | 137.1 |
| sidebar | 16 | go-before | 15,749 (14,285–17,516) | 0.866 | 1.687 | 3.797 | 150.2 |
| sidebar | 16 | go | 15,917 (15,754–16,609) | 0.861 | 1.665 | 3.791 | 151.2 |
| search | 16 | go-before | 14,944 (14,080–15,455) | 0.871 | 1.894 | 4.527 | 214.6 |
| search | 16 | go | 15,280 (14,944–15,482) | 0.864 | 1.810 | 4.607 | 215.5 |
| static_css | 16 | go-before | 104,553 (103,752–105,254) | 0.062 | 0.363 | 0.901 | 17.3 |
| static_css | 16 | go | 101,372 (100,619–103,516) | 0.062 | 0.382 | 0.958 | 17.8 |
| post_message | 16 | go-before | 2,435 (2,413–2,453) | 4.879 | 14.119 | 26.927 | 549.5 |
| post_message | 16 | go | 2,483 (2,479–2,497) | 4.795 | 13.863 | 26.095 | 532.6 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| go-before | not validated | — | 62.7 | 38.5 | 131.3 | 131.3 | 36.3 |
| go | not validated | — | 64.3 | 38.9 | 135.6 | 135.6 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and identity encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 103,189 acknowledged HTTP posts and 832 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | go-before | 374,036 | — | not recorded |
| active_room | go-before | 374,036 | — | not recorded |
| messages_page | go-before | 342,444 | — | not recorded |
| sidebar | go-before | 9,462 | — | not recorded |
| search | go-before | 135,497 | — | not recorded |
| static_css | go-before | 1,218 | — | not recorded |
| room_show | go | 374,036 | — | not recorded |
| active_room | go | 374,036 | — | not recorded |
| messages_page | go | 342,444 | — | not recorded |
| sidebar | go | 9,462 | — | not recorded |
| search | go | 135,497 | — | not recorded |
| static_css | go | 1,218 | — | not recorded |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
