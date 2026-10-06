# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | base | 22,431 (21,117–22,436) | 0.491 | 1.215 | 5.223 | 129.8 |
| room_show | 16 | shell | 22,040 (21,331–22,327) | 0.493 | 1.224 | 5.355 | 126.1 |
| room_show | 16 | rust | 28,307 (28,281–28,560) | 0.485 | 0.881 | 1.886 | 100.4 |
| active_room | 16 | base | 29,528 (28,782–29,568) | 0.448 | 0.829 | 2.491 | 115.4 |
| active_room | 16 | shell | 29,829 (29,772–30,166) | 0.441 | 0.816 | 2.399 | 113.5 |
| active_room | 16 | rust | 26,853 (26,609–26,886) | 0.508 | 0.953 | 1.973 | 106.4 |
| messages_page | 16 | base | 32,522 (31,689–32,848) | 0.426 | 0.763 | 1.578 | 111.6 |
| messages_page | 16 | shell | 32,858 (32,071–33,718) | 0.421 | 0.761 | 1.565 | 110.9 |
| messages_page | 16 | rust | 29,319 (28,988–29,706) | 0.471 | 0.851 | 1.793 | 92.6 |
| search | 16 | base | 19,757 (19,587–19,811) | 0.657 | 1.456 | 3.627 | 171.7 |
| search | 16 | shell | 26,560 (26,541–27,034) | 0.525 | 0.953 | 2.065 | 112.0 |
| search | 16 | rust | 26,734 (26,564–26,896) | 0.546 | 0.890 | 1.278 | 105.0 |
| active_search | 16 | base | 13,082 (12,229–13,142) | 0.890 | 2.443 | 5.843 | 246.1 |
| active_search | 16 | shell | 16,590 (16,526–16,657) | 0.743 | 1.805 | 4.447 | 183.3 |
| active_search | 16 | rust | 14,083 (13,902–14,100) | 1.073 | 1.627 | 2.321 | 198.8 |
| post_message | 16 | base | 2,272 (2,261–2,316) | 5.227 | 15.119 | 28.431 | 666.1 |
| post_message | 16 | shell | 2,271 (2,254–2,293) | 5.175 | 15.159 | 28.751 | 668.6 |
| post_message | 16 | rust | 3,183 (3,144–3,198) | 4.759 | 6.451 | 11.239 | 765.2 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| base | not validated | — | 61.6 | 37.9 | 152.9 | 152.9 | 36.3 |
| shell | not validated | — | 60.2 | 38.2 | 156.2 | 156.2 | 36.3 |
| rust | not validated | — | 50.4 | 41.3 | 133.5 | 133.5 | 35.2 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

9 completed application runs; 161,538 acknowledged HTTP posts and 1,250 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | base | 374,036 | 21289 | gzip |
| active_room | base | 374,036 | 21289 | gzip |
| messages_page | base | 342,444 | 12819 | gzip |
| search | base | 135,497 | 9462 | gzip |
| active_search | base | 135,497 | 9462 | gzip |
| room_show | shell | 374,036 | 21289 | gzip |
| active_room | shell | 374,036 | 21289 | gzip |
| messages_page | shell | 342,444 | 12819 | gzip |
| search | shell | 135,497 | 9462 | gzip |
| active_search | shell | 135,497 | 9462 | gzip |
| room_show | rust | 416,139 | 24235 | gzip |
| active_room | rust | 416,139 | 24235 | gzip |
| messages_page | rust | 383,844 | 16158 | gzip |
| search | rust | 149,625 | 9766 | gzip |
| active_search | rust | 149,625 | 9766 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
