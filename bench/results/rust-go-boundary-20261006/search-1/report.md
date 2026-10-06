# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| search | 1 | rust | 5,559 (5,445–5,614) | 0.179 | 0.194 | 0.210 | 154.0 |
| search | 1 | go | 3,239 (3,172–3,282) | 0.213 | 0.274 | 3.465 | 241.4 |
| search | 1 | fused | 3,215 (3,148–3,316) | 0.210 | 0.293 | 3.387 | 241.4 |
| search | 16 | rust | 24,016 (23,903–24,443) | 0.607 | 0.994 | 1.445 | 117.5 |
| search | 16 | go | 10,326 (10,169–10,533) | 1.014 | 3.093 | 8.807 | 234.7 |
| search | 16 | fused | 10,172 (9,218–10,460) | 1.027 | 3.191 | 9.079 | 245.7 |
| search | 64 | rust | 28,641 (28,006–28,656) | 2.085 | 3.057 | 4.603 | 106.9 |
| search | 64 | go | 9,355 (9,048–9,456) | 6.031 | 11.511 | 17.823 | 243.6 |
| search | 64 | fused | 9,143 (8,975–9,823) | 6.043 | 12.007 | 18.815 | 250.6 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| rust | not validated | — | 42.2 | 42.0 | 57.1 | 57.1 | 35.2 |
| go | not validated | — | 51.9 | 38.9 | 59.4 | 59.4 | 36.3 |
| fused | not validated | — | 34.1 | 38.6 | 58.9 | 58.9 | 36.3 |

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
| search | rust | 149,625 | 9767 | gzip |
| search | go | 135,497 | 9462 | gzip |
| search | fused | 135,497 | 9462 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
