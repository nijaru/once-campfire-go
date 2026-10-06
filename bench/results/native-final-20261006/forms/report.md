# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 30.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0,2,4,6; load generator on CPUs 8,10,12,14. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Opt-in message forms require complete document layouts and seed-grounded editor/attachment/ordered boost controls; selected applications also require exact decoded form bytes after normalization of only their listener origin. These are full-document GETs, not Turbo-Frame click measurements. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| boosts_index | 16 | base | 13,103 (13,041–13,105) | 0.979 | 2.453 | 4.031 | 289.8 |
| boosts_index | 16 | final | 16,056 (15,997–16,115) | 0.767 | 2.049 | 3.533 | 238.8 |
| boosts_index | 16 | rust | 28,184 (28,120–28,187) | 0.553 | 0.777 | 0.990 | 135.1 |
| message_edit | 16 | base | 18,350 (18,346–18,368) | 0.705 | 1.693 | 2.887 | 206.0 |
| message_edit | 16 | final | 22,649 (22,610–22,687) | 0.531 | 1.457 | 2.591 | 168.1 |
| message_edit | 16 | rust | 30,131 (30,111–30,299) | 0.519 | 0.712 | 0.892 | 126.0 |
| message_edit_attachment | 16 | base | 16,600 (16,595–16,620) | 0.813 | 1.799 | 3.031 | 225.1 |
| message_edit_attachment | 16 | final | 20,148 (20,122–20,157) | 0.610 | 1.625 | 2.767 | 186.6 |
| message_edit_attachment | 16 | rust | 31,508 (31,321–31,712) | 0.497 | 0.679 | 0.852 | 119.5 |
| new_boost | 16 | base | 17,765 (17,742–17,807) | 0.735 | 1.743 | 2.979 | 211.9 |
| new_boost | 16 | final | 28,469 (28,442–28,523) | 0.408 | 1.141 | 2.257 | 134.2 |
| new_boost | 16 | rust | 28,884 (28,878–29,018) | 0.539 | 0.757 | 0.964 | 131.7 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| base | not validated | — | 30.5 | 40.1 | 54.7 | 54.7 | 37.8 |
| final | not validated | — | 26.0 | 40.0 | 53.8 | 53.8 | 37.8 |
| rust | not validated | — | 37.1 | 44.3 | 51.9 | 51.9 | 35.6 |

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
| boosts_index | base | 29,566 | 5144 | gzip |
| message_edit | base | 23,461 | 4759 | gzip |
| message_edit_attachment | base | 22,755 | 4477 | gzip |
| new_boost | base | 21,938 | 4450 | gzip |
| boosts_index | final | 29,566 | 5144 | gzip |
| message_edit | final | 23,461 | 4759 | gzip |
| message_edit_attachment | final | 22,755 | 4477 | gzip |
| new_boost | final | 21,938 | 4450 | gzip |
| boosts_index | rust | 30,246 | 5203 | gzip |
| message_edit | rust | 23,657 | 4781 | gzip |
| message_edit_attachment | rust | 22,899 | 4481 | gzip |
| new_boost | rust | 22,242 | 4494 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
