# Campfire application benchmark

Release binaries; application HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 10.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Opt-in message forms require complete document layouts and seed-grounded editor/attachment/ordered boost controls; selected applications also require exact decoded form bytes after normalization of only their listener origin. These are full-document GETs, not Turbo-Frame click measurements. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Fixed-rate arrivals are independent of completions. Latency starts at the scheduled arrival, including generator lateness and client queueing; service latency alone would omit these delays. Quantiles are upper bucket bounds (1 µs resolution below 1 ms, <0.2% above). Successful-response percentiles must be read with error/drop/expiry counters in raw samples. Samples with any such failures are rejected and retained separately. Throughput includes drain time when completion exceeds the arrival window. Generator lateness and queue delay disclose client-side limitations; these are not server-only latency measurements.

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Max clients | Offered req/s | App | Successful req/s (range) | Scheduled p50 ms | p95 ms | p99 ms | Generator p99 ms | Queue p99 ms | Service p99 ms | CPU µs/request |
|---|---:|---:|---|---:|---:|---:|---:|---:|---:|---:|---:|
| sidebar | 64 | 500 | rust | 500 (500–500) | 2.320 | 6.200 | 9.040 | 5.016 | 0.536 | 5.520 | 612.0 |
| sidebar | 64 | 500 | go | 500 (500–500) | 2.256 | 6.480 | 9.984 | 5.032 | 0.422 | 6.896 | 802.0 |
| post_message | 64 | 500 | rust | 500 (500–500) | 2.688 | 7.944 | 11.312 | 4.632 | 0.396 | 8.160 | 1748.0 |
| post_message | 64 | 500 | go | 500 (500–500) | 2.696 | 8.816 | 12.592 | 4.832 | 0.410 | 9.968 | 1550.0 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| rust | not validated | — | 50.4 | 41.9 | 110.1 | 110.1 | 35.2 |
| go | not validated | — | 67.5 | 40.2 | 129.3 | 129.3 | 36.4 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](rust-rate-raw.json.gz) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](rust-rate-metadata.json.gz) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the application HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 63,590 acknowledged HTTP posts and 0 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| sidebar | rust | 30,763 | 5910 | gzip |
| sidebar | go | 29,684 | 5815 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
