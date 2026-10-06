# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Opt-in message forms require complete document layouts and seed-grounded editor/attachment/ordered boost controls; selected applications also require exact decoded form bytes after normalization of only their listener origin. These are full-document GETs, not Turbo-Frame click measurements. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| boosts_index | 16 | base | 9,967 (9,834–10,307) | 1.070 | 3.035 | 9.351 | 299.9 |
| boosts_index | 16 | forms | 12,457 (12,289–12,486) | 0.848 | 2.499 | 7.999 | 260.8 |
| boosts_index | 16 | rust | 27,498 (26,947–27,566) | 0.489 | 0.926 | 2.099 | 103.3 |
| message_edit | 16 | base | 12,966 (12,694–13,122) | 0.825 | 2.283 | 7.739 | 220.3 |
| message_edit | 16 | forms | 17,362 (16,174–17,704) | 0.626 | 1.731 | 5.915 | 183.6 |
| message_edit | 16 | rust | 28,638 (28,504–29,648) | 0.464 | 0.909 | 1.934 | 97.1 |
| message_edit_attachment | 16 | base | 11,091 (10,750–11,125) | 0.924 | 2.777 | 9.575 | 231.9 |
| message_edit_attachment | 16 | forms | 13,915 (13,419–14,229) | 0.681 | 2.169 | 8.439 | 195.8 |
| message_edit_attachment | 16 | rust | 29,706 (29,383–30,568) | 0.447 | 0.871 | 2.041 | 89.7 |
| new_boost | 16 | base | 13,286 (12,932–13,606) | 0.836 | 2.151 | 6.819 | 217.9 |
| new_boost | 16 | forms | 22,767 (21,493–23,174) | 0.461 | 1.316 | 4.767 | 148.8 |
| new_boost | 16 | rust | 28,438 (28,230–28,449) | 0.475 | 0.905 | 1.957 | 99.2 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| base | not validated | — | 32.3 | 37.5 | 52.6 | 52.6 | 36.3 |
| forms | not validated | — | 54.0 | 37.8 | 52.8 | 52.8 | 36.3 |
| rust | not validated | — | 38.3 | 41.3 | 48.7 | 48.7 | 35.2 |

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
| boosts_index | base | 29,566 | 5144 | gzip |
| message_edit | base | 23,461 | 4759 | gzip |
| message_edit_attachment | base | 22,755 | 4477 | gzip |
| new_boost | base | 21,938 | 4450 | gzip |
| boosts_index | forms | 29,566 | 5144 | gzip |
| message_edit | forms | 23,461 | 4759 | gzip |
| message_edit_attachment | forms | 22,755 | 4477 | gzip |
| new_boost | forms | 21,938 | 4450 | gzip |
| boosts_index | rust | 30,246 | 5203 | gzip |
| message_edit | rust | 23,657 | 4781 | gzip |
| message_edit_attachment | rust | 22,899 | 4481 | gzip |
| new_boost | rust | 22,242 | 4494 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
