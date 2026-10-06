# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Opt-in message forms require complete document layouts and seed-grounded editor/attachment/ordered boost controls; selected applications also require exact decoded form bytes after normalization of only their listener origin. These are full-document GETs, not Turbo-Frame click measurements. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| boosts_index | 16 | base | 10,035 (9,917–10,354) | 1.059 | 2.965 | 9.031 | 297.5 |
| boosts_index | 16 | forms | 11,471 (10,926–12,275) | 0.842 | 2.719 | 9.575 | 259.9 |
| boosts_index | 16 | rust | 26,775 (26,378–28,659) | 0.497 | 0.948 | 2.153 | 104.8 |
| message_edit | 16 | base | 12,941 (12,534–13,266) | 0.811 | 2.251 | 7.963 | 219.8 |
| message_edit | 16 | forms | 16,332 (15,715–16,751) | 0.627 | 1.840 | 7.143 | 182.8 |
| message_edit | 16 | rust | 29,600 (28,662–29,815) | 0.451 | 0.865 | 1.985 | 95.3 |
| message_edit_attachment | 16 | base | 11,126 (10,499–11,304) | 0.911 | 2.811 | 9.303 | 226.5 |
| message_edit_attachment | 16 | forms | 13,608 (12,898–14,555) | 0.675 | 2.233 | 9.295 | 195.2 |
| message_edit_attachment | 16 | rust | 29,939 (29,753–31,538) | 0.438 | 0.859 | 1.965 | 88.5 |
| new_boost | 16 | base | 13,785 (8,167–14,354) | 0.820 | 2.051 | 6.823 | 217.2 |
| new_boost | 16 | forms | 21,069 (20,065–22,335) | 0.468 | 1.350 | 5.955 | 149.2 |
| new_boost | 16 | rust | 27,915 (27,746–28,087) | 0.476 | 0.919 | 2.119 | 99.4 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| base | not validated | — | 58.5 | 37.8 | 53.4 | 53.4 | 36.3 |
| forms | not validated | — | 52.8 | 38.1 | 52.6 | 52.6 | 36.3 |
| rust | not validated | — | 67.3 | 41.4 | 48.6 | 48.6 | 35.2 |

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
| boosts_index | base | 29,566 | 29566 | identity |
| message_edit | base | 23,461 | 23461 | identity |
| message_edit_attachment | base | 22,755 | 22755 | identity |
| new_boost | base | 21,938 | 21938 | identity |
| boosts_index | forms | 29,566 | 29566 | identity |
| message_edit | forms | 23,461 | 23461 | identity |
| message_edit_attachment | forms | 22,755 | 22755 | identity |
| new_boost | forms | 21,938 | 21938 | identity |
| boosts_index | rust | 30,246 | 30246 | identity |
| message_edit | rust | 23,657 | 23657 | identity |
| message_edit_attachment | rust | 22,899 | 22899 | identity |
| new_boost | rust | 22,242 | 22242 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
