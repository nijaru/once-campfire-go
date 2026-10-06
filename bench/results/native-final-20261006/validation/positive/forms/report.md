# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

1 rotating repetitions; 0.1-second HTTP samples after 2-second warmups. Four application workers on CPUs 0,2,4,6; load generator on CPUs 8,10,12,14. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Opt-in message forms require complete document layouts and seed-grounded editor/attachment/ordered boost controls; selected applications also require exact decoded form bytes after normalization of only their listener origin. These are full-document GETs, not Turbo-Frame click measurements. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| boosts_index | 1 | base | 2,704 (2,704–2,704) | 0.303 | 0.608 | 0.792 | 442.8 |
| boosts_index | 1 | final | 2,955 (2,955–2,955) | 0.260 | 0.598 | 0.856 | 370.4 |
| boosts_index | 1 | rust | 7,856 (7,856–7,856) | 0.124 | 0.133 | 0.150 | 114.4 |
| message_edit | 1 | base | 3,650 (3,650–3,650) | 0.223 | 0.460 | 0.860 | 327.0 |
| message_edit | 1 | final | 4,761 (4,761–4,761) | 0.179 | 0.307 | 0.395 | 251.6 |
| message_edit | 1 | rust | 8,566 (8,566–8,566) | 0.114 | 0.122 | 0.139 | 116.7 |
| message_edit_attachment | 1 | base | 3,594 (3,594–3,594) | 0.218 | 0.467 | 0.650 | 360.1 |
| message_edit_attachment | 1 | final | 4,449 (4,449–4,449) | 0.185 | 0.354 | 0.546 | 269.1 |
| message_edit_attachment | 1 | rust | 9,762 (9,762–9,762) | 0.100 | 0.105 | 0.127 | 102.2 |
| new_boost | 1 | base | 3,571 (3,571–3,571) | 0.227 | 0.446 | 0.772 | 334.3 |
| new_boost | 1 | final | 4,736 (4,736–4,736) | 0.166 | 0.361 | 0.551 | 231.6 |
| new_boost | 1 | rust | 8,056 (8,056–8,056) | 0.121 | 0.129 | 0.156 | 124.1 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| base | not validated | — | 36.3 | 40.0 | 50.3 | 50.3 | 37.8 |
| final | not validated | — | 51.9 | 40.1 | 50.3 | 50.3 | 37.8 |
| rust | not validated | — | 65.1 | 44.3 | 48.6 | 48.6 | 35.6 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

3 completed application runs; 0 acknowledged HTTP posts and 0 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| boosts_index | base | 29,566 | 5144 | gzip |
| message_edit | base | 23,461 | 4759 | gzip |
| message_edit_attachment | base | 22,755 | 4477 | gzip |
| new_boost | base | 21,938 | 4450 | gzip |
| boosts_index | final | 29,566 | 5144 | gzip |
| message_edit | final | 23,461 | 4759 | gzip |
| message_edit_attachment | final | 22,755 | 4477 | gzip |
| new_boost | final | 21,938 | 4450 | gzip |
| boosts_index | rust | 30,246 | 5203 | gzip |
| message_edit | rust | 23,657 | 4781 | gzip |
| message_edit_attachment | rust | 22,899 | 4481 | gzip |
| new_boost | rust | 22,242 | 4494 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
