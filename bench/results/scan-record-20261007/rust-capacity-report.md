# Campfire application benchmark

Release binaries; application HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 10.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Opt-in message forms require complete document layouts and seed-grounded editor/attachment/ordered boost controls; selected applications also require exact decoded form bytes after normalization of only their listener origin. These are full-document GETs, not Turbo-Frame click measurements. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | rust | 26,546 (21,047–29,505) | 0.496 | 1.001 | 2.167 | 101.2 |
| room_show | 16 | go | 25,394 (9,028–25,784) | 0.437 | 1.047 | 4.627 | 114.9 |
| messages_page | 16 | rust | 30,817 (21,109–30,982) | 0.445 | 0.812 | 1.827 | 86.1 |
| messages_page | 16 | go | 22,223 (8,553–23,852) | 0.475 | 1.225 | 5.375 | 126.2 |
| search | 16 | rust | 25,950 (22,242–26,366) | 0.574 | 0.878 | 1.232 | 104.8 |
| search | 16 | go | 21,604 (8,661–22,143) | 0.579 | 1.259 | 3.969 | 122.4 |
| post_message | 16 | rust | 2,634 (1,539–3,153) | 5.847 | 8.099 | 12.999 | 918.9 |
| post_message | 16 | go | 2,699 (1,375–2,714) | 4.407 | 12.703 | 24.383 | 546.5 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| rust | not validated | — | 44.0 | 41.8 | 120.1 | 120.1 | 35.2 |
| go | not validated | — | 66.2 | 41.6 | 140.8 | 140.8 | 36.4 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](rust-capacity-raw.json.gz) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](rust-capacity-metadata.json.gz) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the application HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 170,832 acknowledged HTTP posts and 0 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | rust | 416,139 | 24235 | gzip |
| messages_page | rust | 383,844 | 16158 | gzip |
| search | rust | 149,625 | 9767 | gzip |
| room_show | go | 374,036 | 21290 | gzip |
| messages_page | go | 342,444 | 12819 | gzip |
| search | go | 135,497 | 9463 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
