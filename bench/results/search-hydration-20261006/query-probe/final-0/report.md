# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| search | 1 | baseline | 3,165 (3,114–3,328) | 0.215 | 0.302 | 3.499 | 240.8 |
| search | 1 | hit | 3,601 (3,436–3,627) | 0.185 | 0.259 | 3.305 | 207.1 |
| search | 1 | rust | 5,795 (5,493–5,881) | 0.171 | 0.190 | 0.210 | 153.6 |
| search | 16 | baseline | 10,250 (9,996–10,439) | 0.999 | 3.123 | 9.223 | 239.4 |
| search | 16 | hit | 11,895 (11,488–11,971) | 0.805 | 2.651 | 9.015 | 212.4 |
| search | 16 | rust | 24,597 (24,369–24,649) | 0.591 | 0.982 | 1.399 | 114.4 |
| search | 64 | baseline | 9,776 (9,578–9,851) | 5.583 | 11.287 | 17.375 | 243.6 |
| search | 64 | hit | 10,562 (10,240–10,812) | 5.079 | 10.663 | 17.263 | 219.0 |
| search | 64 | rust | 29,203 (29,130–29,417) | 2.071 | 2.999 | 4.595 | 104.7 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| baseline | not validated | — | 37.7 | 38.5 | 59.3 | 59.3 | 36.3 |
| hit | not validated | — | 59.0 | 39.0 | 59.3 | 59.3 | 36.3 |
| rust | not validated | — | 45.6 | 41.9 | 57.5 | 57.5 | 35.2 |

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
| search | baseline | 135,497 | 135497 | identity |
| search | hit | 135,497 | 135497 | identity |
| search | rust | 149,625 | 149625 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
