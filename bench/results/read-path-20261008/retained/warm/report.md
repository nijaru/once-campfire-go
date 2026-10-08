# Campfire application benchmark

Release binaries; application HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 8.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Opt-in message forms require complete document layouts and seed-grounded editor/attachment/ordered boost controls; selected applications also require exact decoded form bytes after normalization of only their listener origin. These are full-document GETs, not Turbo-Frame click measurements. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | go-before | 14,378 (14,134–14,697) | 0.834 | 1.861 | 5.283 | 71.9 |
| room_show | 16 | go | 15,687 (13,191–15,793) | 0.751 | 1.671 | 5.147 | 66.0 |
| messages_page | 16 | go-before | 14,804 (14,647–15,176) | 0.821 | 1.748 | 5.127 | 70.7 |
| messages_page | 16 | go | 16,189 (16,116–16,258) | 0.741 | 1.569 | 4.967 | 65.0 |
| sidebar | 16 | go-before | 16,848 (16,756–16,971) | 0.718 | 1.498 | 4.903 | 62.1 |
| sidebar | 16 | go | 18,075 (18,025–18,134) | 0.657 | 1.351 | 4.751 | 57.9 |
| search | 16 | go-before | 16,813 (16,798–16,876) | 0.728 | 1.470 | 4.819 | 62.7 |
| search | 16 | go | 17,740 (15,440–17,966) | 0.678 | 1.365 | 4.887 | 59.2 |
| post_message | 16 | go-before | 2,705 (2,666–2,706) | 4.439 | 12.519 | 23.791 | 562.3 |
| post_message | 16 | go | 2,693 (2,413–2,725) | 4.483 | 12.663 | 24.159 | 556.2 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| go-before | not validated | — | 61.4 | 41.9 | 139.1 | 139.1 | 36.4 |
| go | not validated | — | 64.5 | 42.5 | 141.5 | 141.5 | 36.4 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the application HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 156,824 acknowledged HTTP posts and 0 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

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
