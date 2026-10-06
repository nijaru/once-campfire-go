# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | query | 323 (308–330) | 45.087 | 73.535 | 109.951 | 8906.0 |
| room_show | 16 | hydrate | 352 (330–360) | 39.231 | 77.567 | 113.343 | 7705.0 |
| room_show | 16 | rust | 1,402 (1,394–1,437) | 10.751 | 17.599 | 25.167 | 2406.6 |
| messages_page | 16 | query | 347 (345–359) | 40.927 | 67.903 | 99.775 | 8287.5 |
| messages_page | 16 | hydrate | 367 (359–386) | 38.111 | 71.743 | 111.615 | 7091.9 |
| messages_page | 16 | rust | 1,479 (1,465–1,479) | 10.263 | 16.383 | 23.839 | 2331.7 |
| search | 16 | query | 812 (754–825) | 16.071 | 34.367 | 57.919 | 3058.9 |
| search | 16 | hydrate | 867 (837–887) | 14.447 | 36.415 | 63.039 | 2603.4 |
| search | 16 | rust | 4,221 (4,138–4,222) | 3.537 | 5.807 | 8.679 | 846.4 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| query | not validated | — | 61.1 | 37.7 | 58.1 | 58.1 | 36.3 |
| hydrate | not validated | — | 61.8 | 38.1 | 59.7 | 59.7 | 36.3 |
| rust | not validated | — | 73.0 | 41.3 | 102.8 | 102.8 | 35.2 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

9 completed application runs; 0 acknowledged HTTP posts and 0 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | query | 374,036 | 21289 | gzip |
| messages_page | query | 342,444 | 12819 | gzip |
| search | query | 135,497 | 9462 | gzip |
| room_show | hydrate | 374,036 | 21288 | gzip |
| messages_page | hydrate | 342,444 | 12819 | gzip |
| search | hydrate | 135,497 | 9462 | gzip |
| room_show | rust | 416,139 | 24235 | gzip |
| messages_page | rust | 383,844 | 16158 | gzip |
| search | rust | 149,625 | 9766 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
