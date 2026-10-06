# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | query | 339 (284–352) | 41.727 | 71.359 | 108.031 | 8328.7 |
| room_show | 16 | batch | 342 (287–345) | 41.215 | 75.583 | 117.183 | 8485.2 |
| room_show | 16 | rust | 1,409 (1,287–1,417) | 10.743 | 17.375 | 24.991 | 2389.5 |
| messages_page | 16 | query | 356 (343–374) | 38.783 | 68.799 | 108.927 | 7537.2 |
| messages_page | 16 | batch | 359 (358–363) | 39.071 | 72.639 | 113.535 | 7898.8 |
| messages_page | 16 | rust | 1,470 (1,394–1,472) | 10.303 | 16.639 | 23.903 | 2329.8 |
| search | 16 | query | 892 (883–895) | 14.399 | 32.079 | 54.943 | 2721.2 |
| search | 16 | batch | 856 (651–879) | 15.055 | 35.327 | 62.879 | 2914.8 |
| search | 16 | rust | 4,118 (3,740–4,243) | 3.615 | 6.019 | 8.831 | 847.8 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| query | not validated | — | 62.6 | 38.0 | 61.0 | 61.0 | 36.3 |
| batch | not validated | — | 59.1 | 37.7 | 57.5 | 57.5 | 36.3 |
| rust | not validated | — | 35.6 | 40.8 | 100.8 | 100.8 | 35.2 |

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
| room_show | query | 374,036 | 21288 | gzip |
| messages_page | query | 342,444 | 12818 | gzip |
| search | query | 135,497 | 9463 | gzip |
| room_show | batch | 374,036 | 21289 | gzip |
| messages_page | batch | 342,444 | 12819 | gzip |
| search | batch | 135,497 | 9461 | gzip |
| room_show | rust | 416,139 | 24235 | gzip |
| messages_page | rust | 383,844 | 16158 | gzip |
| search | rust | 149,625 | 9767 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
