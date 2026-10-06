# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| search | 1 | rust | 5,904 (5,860–5,978) | 0.169 | 0.188 | 0.211 | 158.9 |
| search | 1 | go | 3,210 (3,112–3,366) | 0.212 | 0.300 | 3.485 | 249.8 |
| search | 1 | fused | 3,153 (3,063–3,246) | 0.215 | 0.301 | 3.581 | 249.9 |
| search | 16 | rust | 23,672 (23,665–24,102) | 0.608 | 1.021 | 1.525 | 122.5 |
| search | 16 | go | 10,153 (10,063–10,720) | 1.052 | 3.153 | 8.751 | 248.7 |
| search | 16 | fused | 10,271 (10,227–10,463) | 1.008 | 3.109 | 9.095 | 249.0 |
| search | 64 | rust | 28,346 (27,662–28,525) | 2.133 | 3.177 | 4.967 | 111.0 |
| search | 64 | go | 9,280 (9,088–9,318) | 6.063 | 11.895 | 18.335 | 254.4 |
| search | 64 | fused | 9,335 (9,125–9,545) | 5.915 | 11.591 | 18.015 | 253.2 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| rust | not validated | — | 50.5 | 42.1 | 55.0 | 55.0 | 35.2 |
| go | not validated | — | 64.3 | 38.5 | 58.5 | 58.5 | 36.3 |
| fused | not validated | — | 62.9 | 38.2 | 59.1 | 59.1 | 36.3 |

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
| search | rust | 149,625 | 149625 | identity |
| search | go | 135,497 | 135497 | identity |
| search | fused | 135,497 | 135497 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
