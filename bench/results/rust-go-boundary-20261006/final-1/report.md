# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; gzip encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses gzip encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 1 | rust | 6,360 (6,321–6,374) | 0.156 | 0.165 | 0.178 | 137.4 |
| room_show | 1 | go | 5,391 (5,388–5,474) | 0.139 | 0.176 | 1.755 | 145.5 |
| room_show | 1 | fused | 5,569 (5,509–5,600) | 0.138 | 0.166 | 1.628 | 141.4 |
| room_show | 16 | rust | 26,254 (26,148–26,620) | 0.523 | 0.959 | 1.924 | 105.4 |
| room_show | 16 | go | 20,887 (20,051–21,244) | 0.522 | 1.312 | 5.507 | 136.6 |
| room_show | 16 | fused | 21,031 (19,255–21,332) | 0.519 | 1.302 | 5.471 | 130.2 |
| room_show | 64 | rust | 26,147 (25,987–26,512) | 2.267 | 3.695 | 5.915 | 107.1 |
| room_show | 64 | go | 20,060 (19,646–20,280) | 2.559 | 6.471 | 10.911 | 137.9 |
| room_show | 64 | fused | 19,750 (16,853–20,634) | 2.575 | 6.739 | 11.383 | 137.9 |
| active_room | 1 | rust | 6,056 (6,029–6,165) | 0.161 | 0.173 | 0.196 | 147.3 |
| active_room | 1 | go | 6,492 (6,462–6,497) | 0.139 | 0.154 | 0.219 | 134.8 |
| active_room | 1 | fused | 6,560 (6,540–6,576) | 0.138 | 0.153 | 0.216 | 133.2 |
| active_room | 16 | rust | 24,359 (23,950–24,760) | 0.562 | 1.059 | 2.157 | 113.6 |
| active_room | 16 | go | 28,251 (28,229–28,380) | 0.474 | 0.887 | 2.349 | 123.0 |
| active_room | 16 | fused | 28,312 (28,084–28,716) | 0.465 | 0.876 | 2.471 | 120.1 |
| active_room | 64 | rust | 24,342 (23,762–24,728) | 2.431 | 3.995 | 6.327 | 116.7 |
| active_room | 64 | go | 27,396 (24,553–27,553) | 2.053 | 3.693 | 7.391 | 128.2 |
| active_room | 64 | fused | 28,338 (27,092–28,379) | 2.003 | 3.589 | 7.063 | 126.0 |
| messages_page | 1 | rust | 6,391 (6,384–6,485) | 0.156 | 0.167 | 0.181 | 138.0 |
| messages_page | 1 | go | 6,674 (6,315–6,679) | 0.141 | 0.154 | 0.183 | 129.1 |
| messages_page | 1 | fused | 6,749 (6,718–6,780) | 0.139 | 0.152 | 0.179 | 127.1 |
| messages_page | 16 | rust | 25,987 (25,854–26,399) | 0.527 | 0.988 | 2.022 | 100.7 |
| messages_page | 16 | go | 30,893 (21,592–31,599) | 0.443 | 0.849 | 1.782 | 118.5 |
| messages_page | 16 | fused | 31,494 (30,819–31,512) | 0.441 | 0.815 | 1.666 | 117.3 |
| messages_page | 64 | rust | 26,351 (25,020–26,908) | 2.255 | 3.657 | 5.655 | 101.2 |
| messages_page | 64 | go | 29,885 (26,152–30,354) | 1.908 | 3.375 | 6.831 | 121.7 |
| messages_page | 64 | fused | 29,324 (29,270–29,372) | 1.924 | 3.439 | 7.467 | 120.1 |
| search | 1 | rust | 5,440 (5,431–5,484) | 0.181 | 0.196 | 0.213 | 160.5 |
| search | 1 | go | 4,444 (4,417–4,453) | 0.208 | 0.229 | 0.316 | 204.3 |
| search | 1 | fused | 4,516 (4,514–4,516) | 0.205 | 0.225 | 0.296 | 200.2 |
| search | 16 | rust | 24,413 (24,320–24,694) | 0.597 | 0.978 | 1.423 | 117.5 |
| search | 16 | go | 15,273 (14,444–15,758) | 0.857 | 1.971 | 4.683 | 210.4 |
| search | 16 | fused | 15,765 (11,058–15,926) | 0.841 | 1.843 | 4.351 | 207.8 |
| search | 64 | rust | 28,830 (28,148–28,846) | 2.063 | 3.091 | 4.611 | 107.3 |
| search | 64 | go | 14,630 (14,405–14,743) | 3.957 | 6.807 | 11.855 | 213.7 |
| search | 64 | fused | 14,797 (13,207–15,363) | 3.855 | 6.819 | 12.103 | 213.9 |
| post_message | 1 | rust | 1,731 (1,533–1,739) | 0.544 | 0.628 | 1.742 | 663.7 |
| post_message | 1 | go | 1,557 (1,518–1,566) | 0.551 | 0.668 | 3.279 | 680.7 |
| post_message | 1 | fused | 1,565 (1,564–1,565) | 0.548 | 0.664 | 3.245 | 674.7 |
| post_message | 16 | rust | 3,136 (3,116–3,162) | 4.839 | 6.555 | 10.647 | 785.5 |
| post_message | 16 | go | 2,268 (2,244–2,270) | 5.219 | 15.151 | 29.135 | 682.7 |
| post_message | 16 | fused | 2,284 (2,283–2,292) | 5.203 | 14.999 | 28.383 | 679.0 |
| post_message | 64 | rust | 3,138 (3,125–3,160) | 20.255 | 22.959 | 27.855 | 794.1 |
| post_message | 64 | go | 2,177 (2,053–2,197) | 20.767 | 66.303 | 132.991 | 693.9 |
| post_message | 64 | fused | 2,208 (2,199–2,222) | 20.607 | 64.927 | 128.831 | 680.1 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| rust | not validated | — | 72.6 | 41.7 | 139.1 | 139.1 | 35.2 |
| go | not validated | — | 63.7 | 38.5 | 166.1 | 166.1 | 36.3 |
| fused | not validated | — | 58.9 | 38.6 | 162.8 | 162.8 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and gzip encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

9 completed application runs; 343,511 acknowledged HTTP posts and 3,033 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | rust | 416,139 | 24235 | gzip |
| active_room | rust | 416,139 | 24235 | gzip |
| messages_page | rust | 383,844 | 16158 | gzip |
| search | rust | 149,625 | 9767 | gzip |
| room_show | go | 374,036 | 21289 | gzip |
| active_room | go | 374,036 | 21289 | gzip |
| messages_page | go | 342,444 | 12819 | gzip |
| search | go | 135,496 | 9463 | gzip |
| room_show | fused | 374,036 | 21289 | gzip |
| active_room | fused | 374,036 | 21289 | gzip |
| messages_page | fused | 342,444 | 12819 | gzip |
| search | fused | 135,496 | 9463 | gzip |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
