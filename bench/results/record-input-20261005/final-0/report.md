# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | keys | 11,836 (11,392–12,048) | 0.809 | 3.421 | 7.767 | 168.9 |
| room_show | 16 | records | 13,235 (13,046–13,533) | 0.754 | 2.689 | 6.951 | 160.7 |
| active_room | 16 | keys | 17,507 (16,904–17,991) | 0.687 | 1.556 | 4.935 | 150.6 |
| active_room | 16 | records | 17,942 (16,923–18,314) | 0.684 | 1.481 | 4.663 | 145.4 |
| messages_page | 16 | keys | 20,082 (19,452–20,760) | 0.612 | 1.289 | 4.419 | 137.0 |
| messages_page | 16 | records | 21,416 (20,974–21,788) | 0.581 | 1.176 | 4.211 | 133.3 |
| search | 16 | keys | 15,466 (14,734–15,507) | 0.866 | 1.812 | 4.279 | 215.5 |
| search | 16 | records | 15,556 (15,394–15,644) | 0.850 | 1.826 | 4.327 | 212.5 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| keys | not validated | — | 62.4 | 37.8 | 114.4 | 114.4 | 36.3 |
| records | not validated | — | 33.0 | 38.4 | 120.8 | 120.8 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and identity encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 0 acknowledged HTTP posts and 826 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | keys | 374,036 | 374036 | identity |
| active_room | keys | 374,036 | 374036 | identity |
| messages_page | keys | 342,444 | 342444 | identity |
| search | keys | 135,497 | 135497 | identity |
| room_show | records | 374,036 | 374036 | identity |
| active_room | records | 374,036 | 374036 | identity |
| messages_page | records | 342,444 | 342444 | identity |
| search | records | 135,496 | 135496 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
