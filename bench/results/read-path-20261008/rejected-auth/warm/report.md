# Campfire application benchmark

Release binaries; application HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 8.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Opt-in message forms require complete document layouts and seed-grounded editor/attachment/ordered boost controls; selected applications also require exact decoded form bytes after normalization of only their listener origin. These are full-document GETs, not Turbo-Frame click measurements. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | go-before | 13,964 (9,285–15,072) | 0.885 | 1.936 | 5.079 | 76.3 |
| room_show | 16 | go | 16,713 (15,566–16,867) | 0.711 | 1.565 | 4.911 | 62.3 |
| messages_page | 16 | go-before | 15,347 (15,112–15,425) | 0.819 | 1.618 | 4.947 | 69.0 |
| messages_page | 16 | go | 17,223 (12,737–17,416) | 0.706 | 1.482 | 4.879 | 61.1 |
| sidebar | 16 | go-before | 17,249 (13,967–17,661) | 0.714 | 1.375 | 4.939 | 60.5 |
| sidebar | 16 | go | 15,981 (14,899–19,860) | 0.640 | 2.025 | 5.459 | 65.9 |
| search | 16 | go-before | 17,450 (17,220–17,719) | 0.724 | 1.337 | 4.655 | 60.4 |
| search | 16 | go | 18,527 (15,853–19,707) | 0.633 | 1.327 | 4.695 | 57.0 |
| post_message | 16 | go-before | 2,692 (2,488–2,708) | 4.475 | 12.735 | 24.031 | 569.2 |
| post_message | 16 | go | 2,038 (1,947–2,717) | 5.583 | 16.575 | 32.687 | 695.9 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| go-before | not validated | — | 63.6 | 41.7 | 143.5 | 143.5 | 36.4 |
| go | not validated | — | 58.4 | 41.4 | 142.1 | 142.1 | 36.4 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the application HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 147,283 acknowledged HTTP posts and 0 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | go-before | 374,036 | 21290 | gzip |
| messages_page | go-before | 342,444 | 12819 | gzip |
| sidebar | go-before | 29,684 | 5816 | gzip |
| search | go-before | 135,497 | 9463 | gzip |
| room_show | go | 374,036 | 21290 | gzip |
| messages_page | go | 342,444 | 12819 | gzip |
| sidebar | go | 29,684 | 5816 | gzip |
| search | go | 135,497 | 9463 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
