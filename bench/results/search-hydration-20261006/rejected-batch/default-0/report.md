# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| search | 1 | query | 3,505 (3,406–3,649) | 0.199 | 0.276 | 3.103 | 222.0 |
| search | 1 | batch | 3,571 (3,546–3,632) | 0.197 | 0.272 | 3.061 | 218.5 |
| search | 1 | rust | 5,766 (5,730–5,829) | 0.173 | 0.191 | 0.211 | 153.6 |
| search | 16 | query | 11,855 (11,636–12,001) | 0.827 | 2.737 | 8.519 | 215.1 |
| search | 16 | batch | 11,616 (10,990–12,239) | 0.824 | 2.827 | 9.287 | 214.1 |
| search | 16 | rust | 24,066 (23,345–24,242) | 0.600 | 1.012 | 1.462 | 116.7 |
| search | 64 | query | 10,652 (10,434–10,732) | 5.175 | 10.551 | 17.119 | 224.2 |
| search | 64 | batch | 10,845 (10,378–10,895) | 5.151 | 10.167 | 16.263 | 222.5 |
| search | 64 | rust | 28,373 (27,387–29,427) | 2.099 | 3.131 | 4.839 | 108.7 |
| active_search | 1 | query | 2,631 (2,476–2,820) | 0.337 | 0.415 | 1.330 | 296.4 |
| active_search | 1 | batch | 2,616 (2,543–2,622) | 0.356 | 0.423 | 1.300 | 294.1 |
| active_search | 1 | rust | 2,859 (2,855–3,162) | 0.341 | 0.402 | 0.498 | 290.3 |
| active_search | 16 | query | 7,606 (7,388–7,797) | 1.666 | 4.055 | 8.079 | 343.2 |
| active_search | 16 | batch | 7,499 (7,449–7,836) | 1.688 | 3.953 | 8.351 | 344.8 |
| active_search | 16 | rust | 7,383 (7,285–7,665) | 2.019 | 3.125 | 5.003 | 367.2 |
| active_search | 64 | query | 5,739 (5,532–5,884) | 9.935 | 19.007 | 29.919 | 446.4 |
| active_search | 64 | batch | 5,922 (5,867–5,965) | 9.599 | 18.575 | 28.767 | 434.9 |
| active_search | 64 | rust | 6,707 (6,587–6,723) | 9.327 | 12.519 | 16.511 | 447.0 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| query | not validated | — | 56.3 | 37.8 | 127.7 | 127.7 | 36.3 |
| batch | not validated | — | 57.4 | 37.9 | 125.9 | 125.9 | 36.3 |
| rust | not validated | — | 66.4 | 40.8 | 65.2 | 65.2 | 35.2 |

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
| search | batch | 135,497 | 135497 | identity |
| active_search | batch | 135,497 | 135497 | identity |
| search | rust | 149,625 | 149625 | identity |
| active_search | rust | 149,625 | 149625 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
