# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | shell | 11,193 (10,627–11,591) | 0.765 | 4.001 | 8.303 | 174.7 |
| room_show | 16 | keys | 11,935 (11,900–12,085) | 0.745 | 3.627 | 7.923 | 163.8 |
| active_room | 16 | shell | 18,450 (18,431–18,540) | 0.644 | 1.403 | 4.931 | 148.8 |
| active_room | 16 | keys | 18,715 (18,472–19,202) | 0.640 | 1.386 | 4.807 | 141.5 |
| messages_page | 16 | shell | 21,195 (20,718–21,785) | 0.572 | 1.208 | 4.523 | 137.1 |
| messages_page | 16 | keys | 21,819 (21,551–22,255) | 0.559 | 1.129 | 4.519 | 130.8 |
| search | 16 | shell | 15,422 (14,990–15,946) | 0.858 | 1.804 | 4.363 | 210.2 |
| search | 16 | keys | 15,354 (15,157–15,714) | 0.846 | 1.860 | 4.291 | 214.1 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| shell | not validated | — | 35.7 | 38.8 | 107.6 | 107.6 | 36.3 |
| keys | not validated | — | 32.7 | 38.4 | 106.2 | 106.2 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and identity encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 0 acknowledged HTTP posts and 828 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | shell | 374,036 | 374036 | identity |
| active_room | shell | 374,036 | 374036 | identity |
| messages_page | shell | 342,444 | 342444 | identity |
| search | shell | 135,497 | 135497 | identity |
| room_show | keys | 374,036 | 374036 | identity |
| active_room | keys | 374,036 | 374036 | identity |
| messages_page | keys | 342,444 | 342444 | identity |
| search | keys | 135,497 | 135497 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
