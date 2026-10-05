# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 16 | upstream | 8,155 (7,945–8,284) | 1.021 | 5.455 | 11.191 | 220.7 |
| room_show | 16 | cursor | 7,788 (7,704–8,324) | 1.040 | 5.663 | 11.655 | 228.7 |
| room_show | 16 | prepared | 8,314 (8,260–8,335) | 1.026 | 5.403 | 10.983 | 229.6 |
| active_room | 16 | upstream | 14,824 (14,394–14,855) | 0.771 | 1.929 | 6.067 | 186.4 |
| active_room | 16 | cursor | 14,594 (13,999–15,178) | 0.759 | 1.993 | 6.127 | 192.4 |
| active_room | 16 | prepared | 14,502 (14,477–14,867) | 0.786 | 1.960 | 6.167 | 189.2 |
| messages_page | 16 | upstream | 19,961 (19,617–19,971) | 0.606 | 1.280 | 4.707 | 137.9 |
| messages_page | 16 | cursor | 19,984 (19,484–20,549) | 0.600 | 1.269 | 4.715 | 139.8 |
| messages_page | 16 | prepared | 19,972 (19,374–20,066) | 0.611 | 1.284 | 4.691 | 139.0 |
| sidebar | 16 | upstream | 16,292 (15,647–16,342) | 0.835 | 1.626 | 3.753 | 150.9 |
| sidebar | 16 | cursor | 16,236 (15,405–16,716) | 0.835 | 1.627 | 3.757 | 151.1 |
| sidebar | 16 | prepared | 15,788 (15,284–16,865) | 0.852 | 1.682 | 3.843 | 153.0 |
| search | 16 | upstream | 15,552 (14,912–15,868) | 0.842 | 1.801 | 4.247 | 214.0 |
| search | 16 | cursor | 15,114 (15,010–15,764) | 0.858 | 1.860 | 4.651 | 214.5 |
| search | 16 | prepared | 15,175 (15,048–15,411) | 0.857 | 1.831 | 4.599 | 216.8 |
| static_css | 16 | upstream | 103,345 (99,706–105,439) | 0.062 | 0.374 | 0.928 | 17.6 |
| static_css | 16 | cursor | 101,377 (100,599–104,071) | 0.062 | 0.382 | 0.965 | 18.0 |
| static_css | 16 | prepared | 102,603 (98,243–103,818) | 0.064 | 0.374 | 0.937 | 17.8 |
| post_message | 16 | upstream | 2,313 (2,307–2,345) | 5.147 | 14.839 | 27.855 | 646.4 |
| post_message | 16 | cursor | 2,329 (2,316–2,366) | 5.083 | 14.863 | 27.839 | 642.4 |
| post_message | 16 | prepared | 2,326 (2,234–2,360) | 5.087 | 14.775 | 27.695 | 646.0 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| upstream | not validated | — | 64.0 | 38.6 | 131.9 | 131.9 | 36.2 |
| cursor | not validated | — | 57.9 | 38.6 | 136.5 | 136.5 | 36.2 |
| prepared | not validated | — | 65.9 | 38.8 | 133.1 | 133.1 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and identity encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

9 completed application runs; 145,918 acknowledged HTTP posts and 1,244 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | upstream | 374,036 | 374036 | identity |
| active_room | upstream | 374,036 | 374036 | identity |
| messages_page | upstream | 342,444 | 342444 | identity |
| sidebar | upstream | 9,462 | 9462 | identity |
| search | upstream | 135,497 | 135497 | identity |
| static_css | upstream | 1,218 | 1218 | identity |
| room_show | cursor | 374,036 | 374036 | identity |
| active_room | cursor | 374,036 | 374036 | identity |
| messages_page | cursor | 342,444 | 342444 | identity |
| sidebar | cursor | 9,462 | 9462 | identity |
| search | cursor | 135,497 | 135497 | identity |
| static_css | cursor | 1,218 | 1218 | identity |
| room_show | prepared | 374,036 | 374036 | identity |
| active_room | prepared | 374,036 | 374036 | identity |
| messages_page | prepared | 342,444 | 342444 | identity |
| sidebar | prepared | 9,462 | 9462 | identity |
| search | prepared | 135,497 | 135497 | identity |
| static_css | prepared | 1,218 | 1218 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
