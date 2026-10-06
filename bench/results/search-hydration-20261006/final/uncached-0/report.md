# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | query | 279 (274–299) | 51.295 | 86.783 | 123.903 | 9530.2 |
| room_show | 16 | hydrate | 304 (302–331) | 46.111 | 84.927 | 124.351 | 8500.7 |
| room_show | 16 | rust | 2,060 (2,044–2,169) | 7.199 | 12.463 | 18.079 | 1284.2 |
| messages_page | 16 | query | 302 (301–315) | 47.167 | 80.255 | 128.831 | 8844.1 |
| messages_page | 16 | hydrate | 303 (292–305) | 46.943 | 88.767 | 140.031 | 8097.8 |
| messages_page | 16 | rust | 2,200 (2,187–2,217) | 6.779 | 11.447 | 16.639 | 1227.4 |
| search | 16 | query | 677 (672–722) | 20.655 | 39.999 | 67.455 | 3266.1 |
| search | 16 | hydrate | 774 (632–797) | 16.703 | 38.879 | 72.255 | 3038.3 |
| search | 16 | rust | 5,558 (5,363–5,559) | 2.681 | 4.139 | 6.683 | 507.7 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| query | not validated | — | 54.1 | 38.4 | 58.0 | 58.0 | 36.3 |
| hydrate | not validated | — | 57.1 | 38.3 | 55.3 | 55.3 | 36.3 |
| rust | not validated | — | 66.9 | 41.9 | 95.8 | 95.8 | 35.2 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and identity encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

9 completed application runs; 0 acknowledged HTTP posts and 0 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | query | 374,036 | 374036 | identity |
| messages_page | query | 342,444 | 342444 | identity |
| search | query | 135,497 | 135497 | identity |
| room_show | hydrate | 374,036 | 374036 | identity |
| messages_page | hydrate | 342,444 | 342444 | identity |
| search | hydrate | 135,497 | 135497 | identity |
| room_show | rust | 416,139 | 416139 | identity |
| messages_page | rust | 383,844 | 383844 | identity |
| search | rust | 149,625 | 149625 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
