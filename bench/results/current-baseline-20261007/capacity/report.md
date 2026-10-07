# Campfire application benchmark

Release binaries; application HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 10.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Opt-in message forms require complete document layouts and seed-grounded editor/attachment/ordered boost controls; selected applications also require exact decoded form bytes after normalization of only their listener origin. These are full-document GETs, not Turbo-Frame click measurements. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | rust | 29,271 (26,295–29,495) | 0.462 | 0.869 | 2.024 | 93.4 |
| room_show | 16 | go | 23,665 (14,194–24,224) | 0.465 | 1.161 | 4.543 | 120.8 |
| messages_page | 16 | rust | 29,812 (26,232–30,801) | 0.451 | 0.850 | 1.944 | 87.2 |
| messages_page | 16 | go | 19,341 (18,197–20,511) | 0.557 | 1.483 | 5.143 | 138.0 |
| sidebar | 16 | rust | 27,183 (26,960–29,474) | 0.489 | 0.984 | 2.059 | 99.2 |
| sidebar | 16 | go | 11,505 (9,559–11,936) | 1.044 | 2.595 | 6.651 | 215.2 |
| search | 16 | rust | 24,758 (22,382–25,862) | 0.599 | 0.929 | 1.309 | 108.8 |
| search | 16 | go | 17,274 (9,242–20,984) | 0.701 | 1.689 | 4.651 | 143.1 |
| post_message | 16 | rust | 2,189 (1,449–3,252) | 5.619 | 13.519 | 25.231 | 1041.2 |
| post_message | 16 | go | 2,704 (1,037–2,716) | 4.423 | 12.719 | 24.239 | 553.1 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| rust | not validated | — | 68.7 | 42.0 | 125.6 | 125.6 | 35.2 |
| go | not validated | — | 59.2 | 41.2 | 140.2 | 140.2 | 36.4 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json.gz) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json.gz) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the application HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 160,887 acknowledged HTTP posts and 0 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | rust | 416,139 | 24235 | gzip |
| messages_page | rust | 383,844 | 16158 | gzip |
| sidebar | rust | 30,763 | 5910 | gzip |
| search | rust | 149,625 | 9766 | gzip |
| room_show | go | 374,036 | 21289 | gzip |
| messages_page | go | 342,444 | 12819 | gzip |
| sidebar | go | 29,684 | 5815 | gzip |
| search | go | 135,497 | 9463 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
