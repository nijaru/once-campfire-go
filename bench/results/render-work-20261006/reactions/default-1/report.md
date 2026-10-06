# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 alternating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | dsa | 21,984 (21,797–22,593) | 0.494 | 1.242 | 5.271 | 130.0 |
| room_show | 16 | reaction | 22,180 (21,708–22,352) | 0.498 | 1.210 | 5.303 | 124.8 |
| search | 16 | dsa | 21,279 (20,375–21,348) | 0.575 | 1.307 | 4.235 | 128.2 |
| search | 16 | reaction | 20,576 (20,302–20,906) | 0.589 | 1.334 | 4.543 | 125.8 |
| active_search | 16 | dsa | 16,116 (15,996–16,253) | 0.762 | 1.817 | 4.687 | 188.1 |
| active_search | 16 | reaction | 16,065 (15,945–16,088) | 0.769 | 1.857 | 4.735 | 187.5 |
| post_message | 16 | dsa | 2,335 (2,274–2,347) | 5.091 | 14.503 | 28.415 | 651.1 |
| post_message | 16 | reaction | 2,356 (2,345–2,391) | 5.047 | 14.407 | 27.247 | 638.5 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| dsa | not validated | — | 35.0 | 38.0 | 143.2 | 143.2 | 36.3 |
| reaction | not validated | — | 62.9 | 37.7 | 149.2 | 149.2 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

6 completed application runs; 98,298 acknowledged HTTP posts and 0 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | dsa | 374,036 | 21289 | gzip |
| search | dsa | 135,497 | 9462 | gzip |
| active_search | dsa | 135,497 | 9462 | gzip |
| room_show | reaction | 374,036 | 21289 | gzip |
| search | reaction | 135,497 | 9462 | gzip |
| active_search | reaction | 135,497 | 9462 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
