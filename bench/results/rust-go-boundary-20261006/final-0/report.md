# Campfire application benchmark

Release binaries; public HTTP/1.1 listener; identity encoding; identical seed and CPU affinity. Media use installed native libraries. Host background load recorded.

3 rotating repetitions; 5.0-second HTTP samples after 2-second warmups. Four application workers on CPUs 0-3; load generator on CPUs 4-7. HTTP uses identity encoding. Every run starts with a separate copy of the same seed.

HTTP response contracts compare message IDs, sidebar room IDs, avatar bytes and CSS bytes. Write checks require every successful request to persist and enter the FTS index. Cable checks require every client to subscribe and every message to arrive at every client. Upload checks fetch the actual representation and compare its bytes across applications.

## HTTP

Rates show median (minimum–maximum); latency is the median of each run’s percentile.

| Workload | Clients | App | Requests/s (range) | p50 ms | p90 ms | p99 ms | CPU µs/request |
|---|---:|---|---:|---:|---:|---:|---:|
| room_show | 1 | rust | 5,595 (5,545–5,598) | 0.175 | 0.197 | 0.226 | 164.7 |
| room_show | 1 | go | 5,003 (4,954–5,104) | 0.158 | 0.201 | 1.327 | 159.1 |
| room_show | 1 | fused | 4,970 (4,724–5,108) | 0.157 | 0.200 | 1.467 | 159.8 |
| room_show | 16 | rust | 18,902 (18,376–19,200) | 0.740 | 1.329 | 2.797 | 135.5 |
| room_show | 16 | go | 15,002 (14,571–15,690) | 0.710 | 2.095 | 6.055 | 147.9 |
| room_show | 16 | fused | 14,827 (14,810–15,661) | 0.688 | 2.159 | 6.427 | 145.3 |
| room_show | 64 | rust | 19,946 (19,728–20,034) | 2.873 | 5.135 | 8.143 | 144.6 |
| room_show | 64 | go | 15,578 (15,514–15,754) | 3.267 | 7.867 | 13.071 | 160.9 |
| room_show | 64 | fused | 15,944 (15,912–16,252) | 3.221 | 7.691 | 12.647 | 161.1 |
| active_room | 1 | rust | 5,266 (5,087–5,275) | 0.184 | 0.208 | 0.247 | 173.6 |
| active_room | 1 | go | 5,874 (5,780–5,928) | 0.156 | 0.179 | 0.240 | 146.1 |
| active_room | 1 | fused | 5,777 (5,691–5,939) | 0.160 | 0.184 | 0.265 | 146.4 |
| active_room | 16 | rust | 15,629 (15,432–15,802) | 0.898 | 1.675 | 3.041 | 143.6 |
| active_room | 16 | go | 20,325 (19,814–20,450) | 0.626 | 1.246 | 4.131 | 132.1 |
| active_room | 16 | fused | 19,941 (19,738–20,391) | 0.632 | 1.280 | 4.247 | 130.5 |
| active_room | 64 | rust | 16,765 (16,642–16,828) | 3.507 | 6.027 | 9.327 | 153.6 |
| active_room | 64 | go | 20,643 (20,595–20,907) | 2.671 | 5.131 | 10.031 | 149.5 |
| active_room | 64 | fused | 20,676 (20,285–20,934) | 2.637 | 5.203 | 10.167 | 146.6 |
| messages_page | 1 | rust | 5,194 (5,100–5,376) | 0.189 | 0.209 | 0.228 | 167.1 |
| messages_page | 1 | go | 6,036 (5,932–6,037) | 0.156 | 0.175 | 0.216 | 139.8 |
| messages_page | 1 | fused | 6,051 (6,007–6,082) | 0.155 | 0.176 | 0.209 | 138.4 |
| messages_page | 16 | rust | 19,803 (19,498–20,014) | 0.702 | 1.286 | 2.557 | 127.4 |
| messages_page | 16 | go | 22,574 (22,189–22,849) | 0.554 | 1.100 | 4.067 | 123.9 |
| messages_page | 16 | fused | 22,579 (22,012–22,644) | 0.564 | 1.082 | 4.167 | 122.2 |
| messages_page | 64 | rust | 20,274 (20,104–20,925) | 2.825 | 5.099 | 7.947 | 134.1 |
| messages_page | 64 | go | 23,271 (23,094–23,337) | 2.361 | 4.563 | 9.167 | 136.0 |
| messages_page | 64 | fused | 22,866 (22,102–23,047) | 2.359 | 4.723 | 9.511 | 136.5 |
| search | 1 | rust | 5,685 (5,604–5,803) | 0.173 | 0.193 | 0.219 | 165.0 |
| search | 1 | go | 4,344 (4,323–4,360) | 0.208 | 0.236 | 0.326 | 207.8 |
| search | 1 | fused | 4,436 (4,348–4,438) | 0.206 | 0.233 | 0.298 | 204.6 |
| search | 16 | rust | 23,767 (23,712–24,164) | 0.607 | 1.022 | 1.527 | 121.4 |
| search | 16 | go | 15,011 (14,975–15,292) | 0.854 | 1.887 | 4.415 | 215.2 |
| search | 16 | fused | 16,027 (15,134–16,194) | 0.832 | 1.764 | 3.911 | 209.1 |
| search | 64 | rust | 28,503 (28,125–28,583) | 2.105 | 3.179 | 4.887 | 111.0 |
| search | 64 | go | 14,346 (14,137–14,434) | 3.943 | 7.283 | 12.311 | 215.1 |
| search | 64 | fused | 15,160 (14,580–15,336) | 3.751 | 6.627 | 11.967 | 212.9 |
| post_message | 1 | rust | 1,917 (1,912–1,930) | 0.488 | 0.567 | 1.704 | 610.4 |
| post_message | 1 | go | 1,651 (1,649–1,678) | 0.511 | 0.622 | 3.209 | 639.3 |
| post_message | 1 | fused | 1,656 (1,656–1,672) | 0.507 | 0.618 | 3.249 | 640.0 |
| post_message | 16 | rust | 3,191 (3,182–3,233) | 4.767 | 6.403 | 11.191 | 748.0 |
| post_message | 16 | go | 2,300 (2,298–2,360) | 5.175 | 14.815 | 27.951 | 649.6 |
| post_message | 16 | fused | 2,319 (2,302–2,322) | 5.123 | 14.871 | 27.695 | 648.4 |
| post_message | 64 | rust | 3,155 (3,150–3,231) | 20.079 | 22.911 | 28.271 | 760.6 |
| post_message | 64 | go | 2,264 (2,257–2,292) | 20.031 | 62.751 | 123.263 | 648.6 |
| post_message | 64 | fused | 2,297 (2,258–2,307) | 19.807 | 63.039 | 122.559 | 645.2 |

## Action Cable

Throughput counts posted messages delivered to **all** clients; one such message produces one frame per client. Latency includes delivery, rather than just accepting a post. The reference load generator reports p90 rather than p95.

| Clients | Deflate | App | Messages/s (range) | Frames/s | Per-client p99 ms | All-clients p99 ms | Wire MB/s |
|---:|---|---|---:|---:|---:|---:|---:|

## Media and resources

| App | Upload + thumbnail median ms | Thumbnail bytes | Startup median ms | Idle Pss MiB | After HTTP Pss MiB | Final Pss MiB | Binary MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| rust | not validated | — | 73.7 | 41.7 | 116.8 | 116.8 | 35.2 |
| go | not validated | — | 61.9 | 38.6 | 135.4 | 135.4 | 36.3 |
| fused | not validated | — | 60.3 | 38.3 | 134.5 | 134.5 | 36.3 |

Startup includes application initialization, measured to a successful health request at 25 ms polling intervals. Final memory follows the complete workload sequence and includes allocator high-water effects; it is not a per-client memory measurement. Each media sample creates a new blob and fetches its generated thumbnail.

## Reproduction and limits

- [Raw samples](raw.json) include statuses, errors, latency distributions, delivery counters, byte hashes and memory snapshots. [Metadata](metadata.json) records source/binary/seed hashes, CPU details, affinity and load averages.
- These are local workstation measurements, sequential within the harness. Background host activity is recorded, not eliminated. They are not a language-wide performance claim.
- This comparison uses the public HTTP listener and identity encoding. It does not measure TLS/ACME or zstd throughput. Active-room CPU includes the concurrent mutation writer.
- Sidebar room-ID checks do not validate application/Turbo-frame layout completeness; verify that scope separately before any cross-port sidebar comparison.
- Screen-level HTML and network comparisons remain stricter than the functional response contracts used here. Benchmark validation is not a declaration of complete byte-for-byte UI parity.

## Validation totals

9 completed application runs; 359,123 acknowledged HTTP posts and 3,026 active-room posts verified in messages and FTS; 0 validated thumbnail samples. HTTP errors: 0. Cable throughput was not measured.

Initial full response sizes from the first repetition, before workload timing. Message/room ID and byte contracts are recorded in raw.json. Benchmark contracts do not establish browser or strict HTML parity.

| Response | App | Plain bytes | Encoded bytes | Encoding |
|---|---|---:|---:|---|
| room_show | rust | 416,139 | 416139 | identity |
| active_room | rust | 416,139 | 416139 | identity |
| messages_page | rust | 383,844 | 383844 | identity |
| search | rust | 149,625 | 149625 | identity |
| room_show | go | 374,036 | 374036 | identity |
| active_room | go | 374,036 | 374036 | identity |
| messages_page | go | 342,444 | 342444 | identity |
| search | go | 135,496 | 135496 | identity |
| room_show | fused | 374,036 | 374036 | identity |
| active_room | fused | 374,036 | 374036 | identity |
| messages_page | fused | 342,444 | 342444 | identity |
| search | fused | 135,496 | 135496 | identity |

Application logging is retained. Settings, native libraries, source/binary hashes and toolchains are recorded in metadata.json. Container media byte goldens require their separately pinned toolchain.
