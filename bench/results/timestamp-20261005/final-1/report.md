# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | owned | 21,233 (21,066–22,032) | 0.507 | 1.282 | 5.431 | 128.9 |
| room_show | 16 | timestamp | 21,678 (20,626–22,101) | 0.516 | 1.268 | 5.147 | 136.1 |
| active_room | 16 | owned | 28,534 (28,487–28,988) | 0.464 | 0.870 | 2.345 | 121.1 |
| active_room | 16 | timestamp | 29,041 (28,597–29,490) | 0.456 | 0.862 | 2.469 | 117.8 |
| messages_page | 16 | owned | 31,458 (30,387–31,534) | 0.439 | 0.829 | 1.623 | 117.5 |
| messages_page | 16 | timestamp | 30,577 (30,167–31,398) | 0.443 | 0.866 | 1.836 | 119.0 |
| search | 16 | owned | 16,200 (16,062–16,568) | 0.820 | 1.773 | 3.989 | 205.3 |
| search | 16 | timestamp | 15,719 (15,535–15,762) | 0.835 | 1.867 | 4.159 | 208.0 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| owned | not validated | — | 36.1 | 37.5 | 122.0 | 122.0 | 36.3 |
| timestamp | not validated | — | 64.9 | 37.8 | 120.2 | 120.2 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 0 acknowledged HTTP posts and 830 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | owned | 374,036 | 21289 | gzip |
| active_room | owned | 374,036 | 21289 | gzip |
| messages_page | owned | 342,444 | 12819 | gzip |
| search | owned | 135,496 | 9463 | gzip |
| room_show | timestamp | 374,036 | 21289 | gzip |
| active_room | timestamp | 374,036 | 21289 | gzip |
| messages_page | timestamp | 342,444 | 12819 | gzip |
| search | timestamp | 135,496 | 9463 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
