# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | upstream | 4,051 (4,024–4,106) | 2.907 | 7.719 | 13.719 | 945.7 |
| room_show | 16 | fork | 19,345 (19,209–20,975) | 0.551 | 1.422 | 5.403 | 138.4 |
| active_room | 16 | upstream | 4,352 (4,335–4,390) | 2.681 | 7.279 | 12.255 | 896.4 |
| active_room | 16 | fork | 27,558 (24,172–27,711) | 0.477 | 0.933 | 2.703 | 123.9 |
| messages_page | 16 | upstream | 5,307 (5,305–5,313) | 2.237 | 5.835 | 9.055 | 746.8 |
| messages_page | 16 | fork | 29,679 (29,622–30,365) | 0.451 | 0.866 | 2.125 | 120.1 |
| sidebar | 16 | upstream | 14,087 (13,810–15,164) | 0.960 | 1.888 | 4.423 | 189.8 |
| sidebar | 16 | fork | 16,753 (16,489–17,859) | 0.815 | 1.517 | 3.843 | 148.7 |
| search | 16 | upstream | 7,760 (7,728–7,766) | 1.427 | 4.327 | 7.455 | 487.9 |
| search | 16 | fork | 15,608 (15,407–15,680) | 0.842 | 1.869 | 4.287 | 211.1 |
| static_css | 16 | upstream | 104,840 (104,434–106,142) | 0.063 | 0.360 | 0.902 | 17.4 |
| static_css | 16 | fork | 104,183 (102,104–105,714) | 0.064 | 0.367 | 0.926 | 17.7 |
| post_message | 16 | upstream | 2,202 (2,199–2,302) | 5.419 | 15.407 | 29.503 | 701.3 |
| post_message | 16 | fork | 2,226 (2,069–2,286) | 5.379 | 15.335 | 28.495 | 691.9 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| upstream | not validated | — | 57.6 | 38.1 | 141.4 | 141.4 | 36.2 |
| fork | not validated | — | 56.2 | 38.2 | 147.7 | 147.7 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 93,152 acknowledged HTTP posts and 836 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | upstream | 374,036 | 21288 | gzip |
| active_room | upstream | 374,036 | 21288 | gzip |
| messages_page | upstream | 342,444 | 12819 | gzip |
| sidebar | upstream | 9,462 | 2249 | gzip |
| search | upstream | 135,497 | 9462 | gzip |
| static_css | upstream | 1,218 | 654 | gzip |
| room_show | fork | 374,036 | 21289 | gzip |
| active_room | fork | 374,036 | 21289 | gzip |
| messages_page | fork | 342,444 | 12819 | gzip |
| sidebar | fork | 9,462 | 2249 | gzip |
| search | fork | 135,496 | 9463 | gzip |
| static_css | fork | 1,218 | 654 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
