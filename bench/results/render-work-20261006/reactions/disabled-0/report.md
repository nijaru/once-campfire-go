# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | dsa | 371 (366–377) | 37.887 | 70.719 | 109.311 | 7138.2 |
| room_show | 16 | reaction | 397 (381–410) | 36.383 | 65.855 | 105.343 | 6340.0 |
| messages_page | 16 | dsa | 384 (365–395) | 36.159 | 68.479 | 109.119 | 6928.9 |
| messages_page | 16 | reaction | 401 (397–430) | 34.495 | 66.751 | 103.551 | 6102.8 |
| search | 16 | dsa | 881 (879–895) | 14.375 | 34.335 | 58.783 | 2583.6 |
| search | 16 | reaction | 947 (938–968) | 13.311 | 32.895 | 55.679 | 2293.6 |
| post_message | 16 | dsa | 1,750 (1,523–1,781) | 6.955 | 19.519 | 36.255 | 754.3 |
| post_message | 16 | reaction | 1,809 (1,807–1,828) | 6.751 | 18.687 | 35.871 | 725.7 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| dsa | not validated | — | 59.4 | 38.5 | 55.1 | 55.1 | 36.3 |
| reaction | not validated | — | 52.1 | 38.5 | 59.3 | 59.3 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and identity encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 73,075 acknowledged HTTP posts and 0 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | dsa | 374,036 | 374036 | identity |
| messages_page | dsa | 342,444 | 342444 | identity |
| search | dsa | 135,497 | 135497 | identity |
| room_show | reaction | 374,036 | 374036 | identity |
| messages_page | reaction | 342,444 | 342444 | identity |
| search | reaction | 135,497 | 135497 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
