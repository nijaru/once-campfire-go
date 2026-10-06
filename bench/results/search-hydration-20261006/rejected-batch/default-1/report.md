# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| search | 1 | query | 3,586 (2,837–3,587) | 0.190 | 0.265 | 3.213 | 213.5 |
| search | 1 | batch | 3,511 (3,273–3,657) | 0.192 | 0.267 | 3.351 | 214.7 |
| search | 1 | rust | 5,724 (4,555–5,899) | 0.173 | 0.188 | 0.252 | 149.2 |
| search | 16 | query | 11,540 (10,167–11,577) | 0.852 | 2.833 | 9.151 | 212.6 |
| search | 16 | batch | 11,843 (10,666–12,100) | 0.811 | 2.693 | 9.039 | 212.0 |
| search | 16 | rust | 20,983 (18,609–24,260) | 0.719 | 1.114 | 1.623 | 126.6 |
| search | 64 | query | 9,991 (8,113–10,366) | 5.507 | 11.143 | 17.823 | 218.4 |
| search | 64 | batch | 10,419 (9,736–11,106) | 5.151 | 10.839 | 16.895 | 223.4 |
| search | 64 | rust | 27,915 (27,371–27,980) | 2.139 | 3.087 | 4.575 | 104.2 |
| active_search | 1 | query | 3,088 (2,770–3,545) | 0.246 | 0.314 | 2.537 | 287.6 |
| active_search | 1 | batch | 3,392 (2,730–3,492) | 0.256 | 0.293 | 1.992 | 268.3 |
| active_search | 1 | rust | 3,939 (3,831–3,945) | 0.255 | 0.288 | 0.396 | 233.6 |
| active_search | 16 | query | 6,684 (5,662–9,767) | 1.693 | 5.179 | 11.095 | 354.9 |
| active_search | 16 | batch | 9,803 (9,320–10,192) | 1.162 | 3.481 | 7.451 | 297.1 |
| active_search | 16 | rust | 9,832 (9,598–10,018) | 1.540 | 2.193 | 3.037 | 282.5 |
| active_search | 64 | query | 6,101 (4,667–6,867) | 8.807 | 18.927 | 32.623 | 410.2 |
| active_search | 64 | batch | 7,005 (6,234–7,014) | 7.963 | 16.431 | 26.063 | 368.6 |
| active_search | 64 | rust | 7,674 (7,619–7,756) | 8.239 | 9.679 | 12.031 | 353.5 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| query | not validated | — | 59.2 | 37.6 | 148.7 | 148.7 | 36.3 |
| batch | not validated | — | 57.7 | 38.0 | 151.4 | 151.4 | 36.3 |
| rust | not validated | — | 66.0 | 40.8 | 69.6 | 69.6 | 35.2 |

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
| search | batch | 135,497 | 9462 | gzip |
| active_search | batch | 135,497 | 9462 | gzip |
| search | rust | 149,625 | 9766 | gzip |
| active_search | rust | 149,625 | 9766 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
