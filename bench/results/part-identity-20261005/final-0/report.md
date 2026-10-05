# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | prepared | 8,735 (8,549–8,777) | 0.970 | 5.171 | 11.119 | 228.8 |
| room_show | 16 | parts | 8,749 (8,624–8,761) | 0.962 | 5.179 | 11.271 | 226.7 |
| active_room | 16 | prepared | 15,935 (15,713–16,364) | 0.705 | 1.777 | 5.851 | 186.6 |
| active_room | 16 | parts | 16,154 (15,977–16,267) | 0.688 | 1.729 | 5.835 | 186.6 |
| messages_page | 16 | prepared | 21,154 (21,023–21,696) | 0.566 | 1.194 | 4.575 | 137.2 |
| messages_page | 16 | parts | 20,887 (20,749–21,234) | 0.573 | 1.216 | 4.595 | 138.0 |
| sidebar | 16 | prepared | 16,301 (15,911–16,501) | 0.834 | 1.624 | 3.763 | 149.9 |
| sidebar | 16 | parts | 15,968 (15,743–17,013) | 0.840 | 1.680 | 4.019 | 150.6 |
| search | 16 | prepared | 15,363 (15,333–15,466) | 0.854 | 1.797 | 4.387 | 214.4 |
| search | 16 | parts | 15,054 (14,719–15,417) | 0.861 | 1.844 | 4.771 | 214.3 |
| static_css | 16 | prepared | 102,716 (100,319–102,874) | 0.063 | 0.371 | 0.943 | 17.7 |
| static_css | 16 | parts | 102,293 (100,650–102,464) | 0.063 | 0.374 | 0.953 | 17.7 |
| post_message | 16 | prepared | 2,310 (2,308–2,348) | 5.103 | 14.743 | 27.919 | 642.1 |
| post_message | 16 | parts | 2,329 (2,320–2,331) | 5.107 | 14.719 | 27.695 | 637.1 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| prepared | not validated | — | 63.5 | 38.0 | 131.7 | 131.7 | 36.3 |
| parts | not validated | — | 60.3 | 38.0 | 133.6 | 133.6 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and identity encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 97,465 acknowledged HTTP posts and 834 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | prepared | 374,036 | 374036 | identity |
| active_room | prepared | 374,036 | 374036 | identity |
| messages_page | prepared | 342,444 | 342444 | identity |
| sidebar | prepared | 9,462 | 9462 | identity |
| search | prepared | 135,497 | 135497 | identity |
| static_css | prepared | 1,218 | 1218 | identity |
| room_show | parts | 374,036 | 374036 | identity |
| active_room | parts | 374,036 | 374036 | identity |
| messages_page | parts | 342,444 | 342444 | identity |
| sidebar | parts | 9,462 | 9462 | identity |
| search | parts | 135,497 | 135497 | identity |
| static_css | parts | 1,218 | 1218 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
