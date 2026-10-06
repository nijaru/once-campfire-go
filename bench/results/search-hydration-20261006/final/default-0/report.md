# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| search | 1 | query | 3,486 (3,453–3,523) | 0.198 | 0.274 | 3.185 | 220.9 |
| search | 1 | hydrate | 3,458 (3,436–3,519) | 0.199 | 0.274 | 3.311 | 221.5 |
| search | 1 | rust | 5,685 (5,616–5,764) | 0.174 | 0.193 | 0.222 | 156.5 |
| search | 16 | query | 11,495 (11,406–11,776) | 0.850 | 2.809 | 8.887 | 217.9 |
| search | 16 | hydrate | 11,852 (11,241–11,870) | 0.833 | 2.719 | 8.911 | 217.0 |
| search | 16 | rust | 24,267 (23,949–24,424) | 0.598 | 0.995 | 1.440 | 118.2 |
| search | 64 | query | 10,167 (9,700–10,665) | 5.359 | 10.943 | 16.975 | 224.6 |
| search | 64 | hydrate | 10,808 (10,291–10,829) | 4.951 | 10.463 | 16.911 | 224.1 |
| search | 64 | rust | 27,994 (27,634–28,107) | 2.171 | 3.163 | 4.899 | 110.1 |
| active_search | 1 | query | 2,531 (2,252–2,655) | 0.366 | 0.441 | 1.318 | 296.8 |
| active_search | 1 | hydrate | 2,646 (2,478–2,656) | 0.347 | 0.422 | 1.360 | 293.3 |
| active_search | 1 | rust | 2,800 (2,690–2,808) | 0.346 | 0.425 | 0.542 | 293.6 |
| active_search | 16 | query | 7,330 (7,297–7,340) | 1.746 | 4.079 | 8.311 | 343.7 |
| active_search | 16 | hydrate | 7,060 (6,881–7,166) | 1.785 | 4.323 | 8.831 | 348.3 |
| active_search | 16 | rust | 7,111 (7,052–7,162) | 2.105 | 3.325 | 5.091 | 370.5 |
| active_search | 64 | query | 5,817 (5,721–5,832) | 9.799 | 18.863 | 30.031 | 442.4 |
| active_search | 64 | hydrate | 5,824 (5,642–6,045) | 9.799 | 18.687 | 29.615 | 445.3 |
| active_search | 64 | rust | 6,565 (5,655–6,566) | 9.567 | 12.911 | 17.055 | 458.0 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| query | not validated | — | 31.9 | 37.7 | 132.2 | 132.2 | 36.3 |
| hydrate | not validated | — | 56.8 | 38.0 | 125.0 | 125.0 | 36.3 |
| rust | not validated | — | 64.2 | 41.3 | 66.6 | 66.6 | 35.2 |

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
| search | query | 135,497 | 135497 | identity |
| active_search | query | 135,497 | 135497 | identity |
| search | hydrate | 135,497 | 135497 | identity |
| active_search | hydrate | 135,497 | 135497 | identity |
| search | rust | 149,625 | 149625 | identity |
| active_search | rust | 149,625 | 149625 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
