# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | base | 15,434 (15,290–15,445) | 0.664 | 2.095 | 6.183 | 140.4 |
| room_show | 16 | micro | 15,922 (15,520–16,179) | 0.639 | 1.977 | 6.031 | 138.1 |
| room_show | 16 | rust | 20,664 (20,637–21,024) | 0.682 | 1.199 | 2.683 | 127.2 |
| active_room | 16 | base | 21,408 (21,275–21,675) | 0.586 | 1.174 | 4.081 | 125.4 |
| active_room | 16 | micro | 21,201 (20,806–21,524) | 0.592 | 1.181 | 4.143 | 124.3 |
| active_room | 16 | rust | 17,169 (16,904–17,207) | 0.819 | 1.523 | 2.833 | 133.7 |
| messages_page | 16 | base | 23,700 (23,658–23,986) | 0.523 | 1.038 | 4.107 | 118.9 |
| messages_page | 16 | micro | 23,871 (23,239–24,482) | 0.529 | 1.032 | 3.903 | 118.2 |
| messages_page | 16 | rust | 21,391 (16,760–22,015) | 0.649 | 1.207 | 2.601 | 118.4 |
| search | 16 | base | 26,407 (26,155–26,471) | 0.524 | 0.972 | 2.199 | 113.9 |
| search | 16 | micro | 26,045 (25,774–27,126) | 0.532 | 0.976 | 2.149 | 114.5 |
| search | 16 | rust | 26,380 (25,000–26,397) | 0.549 | 0.925 | 1.353 | 109.0 |
| active_search | 16 | base | 11,091 (11,053–11,115) | 1.113 | 2.703 | 5.767 | 222.9 |
| active_search | 16 | micro | 11,115 (10,980–11,146) | 1.102 | 2.693 | 5.935 | 221.4 |
| active_search | 16 | rust | 9,917 (9,404–9,957) | 1.463 | 2.491 | 3.917 | 267.4 |
| post_message | 16 | base | 2,333 (2,259–2,345) | 5.091 | 14.807 | 27.631 | 636.0 |
| post_message | 16 | micro | 2,483 (2,455–2,508) | 4.759 | 13.855 | 26.399 | 585.7 |
| post_message | 16 | rust | 3,204 (3,132–3,301) | 4.699 | 6.407 | 11.959 | 736.6 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| base | not validated | — | 54.6 | 37.9 | 133.8 | 133.8 | 36.3 |
| micro | not validated | — | 55.4 | 37.9 | 136.4 | 136.4 | 36.3 |
| rust | not validated | — | 44.8 | 41.3 | 111.0 | 111.0 | 35.2 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and identity encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

9 completed application runs; 166,896 acknowledged HTTP posts and 1,247 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | base | 374,036 | 374036 | identity |
| active_room | base | 374,036 | 374036 | identity |
| messages_page | base | 342,444 | 342444 | identity |
| search | base | 135,497 | 135497 | identity |
| active_search | base | 135,497 | 135497 | identity |
| room_show | micro | 374,036 | 374036 | identity |
| active_room | micro | 374,036 | 374036 | identity |
| messages_page | micro | 342,444 | 342444 | identity |
| search | micro | 135,497 | 135497 | identity |
| active_search | micro | 135,497 | 135497 | identity |
| room_show | rust | 416,139 | 416139 | identity |
| active_room | rust | 416,139 | 416139 | identity |
| messages_page | rust | 383,844 | 383844 | identity |
| search | rust | 149,625 | 149625 | identity |
| active_search | rust | 149,625 | 149625 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
