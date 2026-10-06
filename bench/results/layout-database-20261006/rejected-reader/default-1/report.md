# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | base | 22,390 (22,214–22,408) | 0.489 | 1.208 | 5.047 | 129.5 |
| room_show | 16 | driver | 23,314 (23,084–23,883) | 0.429 | 1.236 | 5.539 | 124.5 |
| active_room | 16 | base | 29,928 (29,644–30,092) | 0.448 | 0.809 | 2.251 | 115.6 |
| active_room | 16 | driver | 31,052 (30,654–31,190) | 0.368 | 0.933 | 3.033 | 109.2 |
| messages_page | 16 | base | 32,485 (32,033–33,624) | 0.427 | 0.752 | 1.556 | 113.0 |
| messages_page | 16 | driver | 32,803 (32,440–33,068) | 0.360 | 0.879 | 2.579 | 108.7 |
| search | 16 | base | 19,787 (19,600–19,995) | 0.657 | 1.448 | 3.709 | 170.7 |
| search | 16 | driver | 19,721 (19,167–19,961) | 0.553 | 1.512 | 4.467 | 168.8 |
| post_message | 16 | base | 2,253 (2,209–2,296) | 5.279 | 15.231 | 28.191 | 671.4 |
| post_message | 16 | driver | 2,248 (2,243–2,298) | 5.271 | 15.167 | 28.335 | 673.3 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| base | not validated | — | 29.7 | 37.9 | 146.1 | 146.1 | 36.3 |
| driver | not validated | — | 57.1 | 37.9 | 146.2 | 146.2 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 95,029 acknowledged HTTP posts and 831 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | base | 374,036 | 21289 | gzip |
| active_room | base | 374,036 | 21289 | gzip |
| messages_page | base | 342,444 | 12819 | gzip |
| search | base | 135,497 | 9462 | gzip |
| room_show | driver | 374,036 | 21289 | gzip |
| active_room | driver | 374,036 | 21289 | gzip |
| messages_page | driver | 342,444 | 12819 | gzip |
| search | driver | 135,497 | 9462 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
