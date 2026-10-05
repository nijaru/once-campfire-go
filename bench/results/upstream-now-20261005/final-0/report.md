# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | upstream | 8,153 (7,915–8,231) | 1.090 | 5.107 | 10.351 | 224.7 |
| room_show | 16 | fork | 15,230 (14,717–15,362) | 0.721 | 2.011 | 6.183 | 146.8 |
| active_room | 16 | upstream | 14,437 (13,954–14,724) | 0.823 | 2.004 | 5.643 | 193.5 |
| active_room | 16 | fork | 18,913 (18,781–20,130) | 0.677 | 1.391 | 4.009 | 136.7 |
| messages_page | 16 | upstream | 19,585 (19,293–19,714) | 0.632 | 1.324 | 4.211 | 142.1 |
| messages_page | 16 | fork | 21,720 (21,354–22,087) | 0.588 | 1.170 | 3.745 | 128.1 |
| sidebar | 16 | upstream | 17,276 (15,314–17,765) | 0.792 | 1.469 | 3.743 | 154.2 |
| sidebar | 16 | fork | 16,558 (15,876–17,131) | 0.821 | 1.556 | 3.915 | 153.0 |
| search | 16 | upstream | 14,765 (14,752–14,991) | 0.885 | 1.897 | 4.623 | 219.4 |
| search | 16 | fork | 15,338 (14,887–15,362) | 0.853 | 1.847 | 4.515 | 212.3 |
| static_css | 16 | upstream | 102,366 (100,192–103,126) | 0.066 | 0.372 | 0.934 | 17.9 |
| static_css | 16 | fork | 103,832 (102,548–104,573) | 0.064 | 0.370 | 0.915 | 17.7 |
| post_message | 16 | upstream | 2,301 (2,273–2,304) | 5.131 | 14.959 | 28.367 | 669.8 |
| post_message | 16 | fork | 2,327 (2,316–2,360) | 5.099 | 14.711 | 27.391 | 655.1 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| upstream | not validated | — | 57.1 | 38.0 | 134.0 | 134.0 | 36.2 |
| fork | not validated | — | 60.2 | 37.8 | 132.4 | 132.4 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and identity encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 96,823 acknowledged HTTP posts and 832 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | upstream | 374,036 | 374036 | identity |
| active_room | upstream | 374,036 | 374036 | identity |
| messages_page | upstream | 342,444 | 342444 | identity |
| sidebar | upstream | 9,462 | 9462 | identity |
| search | upstream | 135,497 | 135497 | identity |
| static_css | upstream | 1,218 | 1218 | identity |
| room_show | fork | 374,036 | 374036 | identity |
| active_room | fork | 374,036 | 374036 | identity |
| messages_page | fork | 342,444 | 342444 | identity |
| sidebar | fork | 9,462 | 9462 | identity |
| search | fork | 135,496 | 135496 | identity |
| static_css | fork | 1,218 | 1218 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
