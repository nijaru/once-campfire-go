# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| search | 1 | query | 3,533 (3,503–3,538) | 0.191 | 0.264 | 3.295 | 213.9 |
| search | 1 | hydrate | 3,641 (3,592–3,700) | 0.190 | 0.239 | 3.207 | 206.5 |
| search | 1 | rust | 5,642 (5,541–5,650) | 0.176 | 0.192 | 0.213 | 153.1 |
| search | 16 | query | 11,947 (11,676–12,210) | 0.812 | 2.611 | 8.911 | 210.9 |
| search | 16 | hydrate | 12,394 (11,747–12,554) | 0.794 | 2.571 | 8.751 | 198.6 |
| search | 16 | rust | 24,592 (24,278–24,648) | 0.599 | 0.956 | 1.352 | 113.2 |
| search | 64 | query | 10,093 (10,074–10,431) | 5.371 | 10.975 | 17.759 | 218.5 |
| search | 64 | hydrate | 10,409 (10,273–11,034) | 5.203 | 10.751 | 16.783 | 211.4 |
| search | 64 | rust | 28,375 (27,967–28,576) | 2.109 | 3.013 | 4.483 | 103.8 |
| active_search | 1 | query | 3,455 (3,412–3,458) | 0.253 | 0.287 | 1.904 | 262.2 |
| active_search | 1 | hydrate | 3,457 (3,405–3,518) | 0.254 | 0.284 | 1.841 | 263.2 |
| active_search | 1 | rust | 3,945 (3,899–3,967) | 0.253 | 0.286 | 0.424 | 230.6 |
| active_search | 16 | query | 9,714 (9,696–9,874) | 1.171 | 3.441 | 7.535 | 293.3 |
| active_search | 16 | hydrate | 9,808 (9,643–9,903) | 1.182 | 3.423 | 7.271 | 288.2 |
| active_search | 16 | rust | 10,111 (9,896–10,128) | 1.505 | 2.107 | 2.935 | 283.6 |
| active_search | 64 | query | 6,983 (6,925–7,085) | 7.883 | 16.447 | 27.247 | 367.6 |
| active_search | 64 | hydrate | 6,969 (6,862–7,024) | 7.863 | 16.367 | 26.927 | 369.1 |
| active_search | 64 | rust | 7,689 (7,654–7,775) | 8.199 | 9.679 | 11.927 | 355.4 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| query | not validated | — | 59.5 | 38.0 | 152.6 | 152.6 | 36.3 |
| hydrate | not validated | — | 61.4 | 37.8 | 161.3 | 161.3 | 36.3 |
| rust | not validated | — | 72.1 | 41.3 | 70.2 | 70.2 | 35.2 |

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
| search | query | 135,497 | 9462 | gzip |
| active_search | query | 135,497 | 9462 | gzip |
| search | hydrate | 135,497 | 9462 | gzip |
| active_search | hydrate | 135,497 | 9462 | gzip |
| search | rust | 149,625 | 9766 | gzip |
| active_search | rust | 149,625 | 9766 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
