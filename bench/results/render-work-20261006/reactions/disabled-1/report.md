# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | dsa | 375 (320–392) | 37.599 | 71.551 | 105.471 | 7101.1 |
| room_show | 16 | reaction | 421 (404–431) | 32.303 | 63.999 | 97.407 | 5781.2 |
| messages_page | 16 | dsa | 387 (385–395) | 35.487 | 67.199 | 106.879 | 6556.9 |
| messages_page | 16 | reaction | 433 (415–438) | 31.231 | 62.943 | 99.967 | 5714.3 |
| search | 16 | dsa | 889 (844–920) | 14.111 | 34.655 | 58.751 | 2516.2 |
| search | 16 | reaction | 966 (966–1,000) | 12.735 | 31.743 | 56.671 | 2166.5 |
| post_message | 16 | dsa | 1,978 (1,961–1,995) | 6.155 | 17.263 | 32.559 | 733.2 |
| post_message | 16 | reaction | 1,981 (1,932–2,002) | 6.147 | 17.231 | 33.023 | 729.8 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| dsa | not validated | — | 58.1 | 38.0 | 64.8 | 64.8 | 36.3 |
| reaction | not validated | — | 32.5 | 37.7 | 67.0 | 67.0 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 82,441 acknowledged HTTP posts and 0 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | dsa | 374,036 | 21289 | gzip |
| messages_page | dsa | 342,444 | 12819 | gzip |
| search | dsa | 135,497 | 9462 | gzip |
| room_show | reaction | 374,036 | 21289 | gzip |
| messages_page | reaction | 342,444 | 12819 | gzip |
| search | reaction | 135,497 | 9462 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
