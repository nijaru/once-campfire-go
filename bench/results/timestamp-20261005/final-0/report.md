# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | owned | 15,088 (14,324–15,196) | 0.712 | 2.053 | 6.355 | 143.6 |
| room_show | 16 | timestamp | 14,869 (14,671–15,116) | 0.713 | 2.147 | 6.411 | 144.1 |
| active_room | 16 | owned | 19,711 (19,623–19,787) | 0.646 | 1.313 | 4.191 | 131.9 |
| active_room | 16 | timestamp | 20,107 (19,980–20,211) | 0.624 | 1.262 | 4.295 | 125.4 |
| messages_page | 16 | owned | 21,818 (21,789–22,222) | 0.564 | 1.123 | 4.183 | 122.8 |
| messages_page | 16 | timestamp | 21,721 (21,436–22,749) | 0.574 | 1.134 | 4.167 | 122.7 |
| search | 16 | owned | 15,494 (15,390–16,050) | 0.840 | 1.836 | 4.207 | 211.1 |
| search | 16 | timestamp | 15,743 (15,677–15,963) | 0.829 | 1.794 | 4.211 | 207.9 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| owned | not validated | — | 31.5 | 38.3 | 111.4 | 111.4 | 36.3 |
| timestamp | not validated | — | 63.4 | 38.4 | 111.9 | 111.9 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and identity encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 0 acknowledged HTTP posts and 830 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | owned | 374,036 | 374036 | identity |
| active_room | owned | 374,036 | 374036 | identity |
| messages_page | owned | 342,444 | 342444 | identity |
| search | owned | 135,496 | 135496 | identity |
| room_show | timestamp | 374,036 | 374036 | identity |
| active_room | timestamp | 374,036 | 374036 | identity |
| messages_page | timestamp | 342,444 | 342444 | identity |
| search | timestamp | 135,496 | 135496 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
