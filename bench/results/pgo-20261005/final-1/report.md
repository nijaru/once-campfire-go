# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | owned | 22,191 (21,461–22,216) | 0.511 | 1.212 | 5.283 | 130.5 |
| room_show | 16 | pgo | 20,400 (20,018–21,075) | 0.515 | 1.356 | 5.647 | 132.5 |
| active_room | 16 | owned | 28,425 (28,390–28,894) | 0.469 | 0.870 | 2.413 | 121.6 |
| active_room | 16 | pgo | 28,211 (27,942–28,397) | 0.467 | 0.901 | 2.609 | 121.0 |
| messages_page | 16 | owned | 31,020 (30,733–31,638) | 0.446 | 0.825 | 1.686 | 118.0 |
| messages_page | 16 | pgo | 31,670 (30,544–31,899) | 0.434 | 0.813 | 1.664 | 115.9 |
| search | 16 | owned | 16,066 (16,008–16,189) | 0.825 | 1.797 | 4.143 | 204.6 |
| search | 16 | pgo | 15,952 (15,929–15,998) | 0.809 | 1.862 | 4.219 | 202.0 |
| static_css | 16 | owned | 102,790 (102,115–103,035) | 0.063 | 0.374 | 0.936 | 17.7 |
| static_css | 16 | pgo | 102,886 (102,732–107,668) | 0.061 | 0.379 | 0.950 | 17.6 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| owned | not validated | — | 34.7 | 37.6 | 123.8 | 123.8 | 36.3 |
| pgo | not validated | — | 62.1 | 38.4 | 124.1 | 124.1 | 36.7 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 0 acknowledged HTTP posts and 834 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | owned | 374,036 | 21289 | gzip |
| active_room | owned | 374,036 | 21289 | gzip |
| messages_page | owned | 342,444 | 12819 | gzip |
| search | owned | 135,496 | 9463 | gzip |
| static_css | owned | 1,218 | 654 | gzip |
| room_show | pgo | 374,036 | 21289 | gzip |
| active_room | pgo | 374,036 | 21289 | gzip |
| messages_page | pgo | 342,444 | 12819 | gzip |
| search | pgo | 135,496 | 9463 | gzip |
| static_css | pgo | 1,218 | 654 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
