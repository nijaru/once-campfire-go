# Room refresh correctness and gzip reuse — 2026-10-05

The fork matched latest upstream Go, `504428a`, at the start of this work. This pass
fixes the room refresh cursor and reuses completed gzip bodies. It does not add a
framework, change SQLite, replace templates, or implement generation-based page caching.
The final upstream recheck still found `504428a` on main, but unmerged optimization
[PR #4](https://github.com/basecamp/once-campfire-go/pull/4) and
[PR #5](https://github.com/basecamp/once-campfire-go/pull/5) now exist. PR #5 opened during
this work. Their proposed changes are not included in this comparison.

## Results

Three rotating repetitions, 16 clients, five-second samples after two-second warmups.
All candidates use the public HTTP listener, identical disposable Rails parity seeds,
Go 1.27.1, four application workers, and fixed server/client CPU sets.

**These are Docker Desktop Linux/arm64 VM measurements on an Apple M3 Max.** Affinity
pins VM vCPUs, not physical Apple cores. Host scheduling remains uncontrolled. Public
request logging is enabled, including its bind-mounted log writes. These numbers are
not comparable to the published native AMD Go/Rust table and establish no Rust ratio.

Median gzip requests/sec:

| Workload | Upstream | + Room cursor fix | + Gzip reuse | Reuse / upstream |
|---|---:|---:|---:|---:|
| Unchanged room | 4,145 | 4,123 | 8,077 | 1.95× |
| Active room, target 20 commits/sec | 4,415 | 4,391 | 11,665 | 2.64× |
| Message history | 5,334 | 5,386 | 14,965 | 2.81× |
| Search | 7,922 | 7,925 | 13,642 | 1.72× |
| Sidebar, upstream bare frame | 15,430 | 15,412 | 15,030 | 0.97× |
| Post message | 2,229 | 2,251 | 2,280 | 1.02× |
| Static CSS | 103,063 | 106,183 | 103,147 | 1.00× |

The cursor fix alone does not improve dynamic compression. It makes room bodies stable;
compression reuse then removes repeated gzip work. Room CPU fell from 931 to 362 µs/request,
and message-history CPU from 739 to 258 µs/request. Room p50/p90 fell from 2.96/7.36 to
1.31/3.69 ms, but p99 improved only from 13.11 to 12.46 ms. History p99 fell from 8.92 to
4.06 ms. Sidebar CPU fell from 183 to 153 µs/request without a throughput improvement:
not every route benefits. The report includes all ranges and percentiles, not just wins.

Median identity requests/sec:

| Workload | Upstream | + Room cursor fix | + Gzip reuse |
|---|---:|---:|---:|
| Unchanged room | 8,155 | 7,788 | 8,314 |
| Active room | 14,824 | 14,594 | 14,502 |
| Message history | 19,961 | 19,984 | 19,972 |
| Search | 15,552 | 15,114 | 15,175 |
| Sidebar, upstream bare frame | 16,292 | 16,236 | 15,788 |
| Post message | 2,313 | 2,329 | 2,326 |
| Static CSS | 103,345 | 101,377 | 102,603 |

Identity throughput changed by roughly −3% to +2%, rather than tracking the gzip gains.
After the gzip HTTP sequence, median process Pss was 138.7 MiB upstream and 146.7 MiB
with reuse. After identity HTTP it was 131.9 and 133.1 MiB. These are sequence/high-water
measurements, not the cache's maximum resident footprint.

[Identity report](final-0/report.md), [gzip report](final-1/report.md),
[build/source manifest](harness.json), and raw JSON retain sizes, hashes, CPU, latency,
memory, errors, settings and source references. No builds or tests ran during final timing.

## Correctness and cache ownership

Rails defines `Time::DATE_FORMATS[:epoch]` as epoch milliseconds and uses the queried
room's `updated_at` in `messages_helper.rb`. Go previously used the render-time clock.
The regression failed on upstream: its cursor advanced past a message committed between
query and render. The fix retains the queried room timestamp, including millisecond
truncation, and verifies the late message arrives through the real refresh endpoint.
Repeated unchanged room rendering also retains identical bytes and ETags.

The integrated server has two compression layers. Its inner `Deflate` middleware already
compresses gzip requests before the outer public compressor sees them. The new memo sits
there, not in the public HTTP response cache.

Completed GET bodies hand their existing byte parts to that compressor. A SHA-256 of
actual body bytes plus gzip mtime identifies immutable gzip bytes; a weak/resource ETag
is deliberately insufficient. Authorization, page queries, rendering/fragment lookup,
conditional handling and cookies still run on every request. No headers, cookies or plain
bodies are retained. Changing the response changes its compression identity automatically.

The per-handler LRU accounts for compressed backing capacity plus entry overhead within
32 MiB. Bodies over 1 MiB, compressed entries over 256 KiB, and entries exceeding one eighth
of the cache budget are not retained. No-store, write responses and streaming paths bypass
reuse. Up to 16 in-flight keys share compression outside the cache lock; saturation permits
independent compression instead of blocking unrelated keys. This is not a bound on total
process heap, pooled writers, or responses currently being sent.

Tests cover exact decoding, chunk-boundary-independent identity, mtime, concurrent reuse,
eviction/large-entry bypass, distinct bodies sharing a weak ETag, per-request cookies,
no-store/write admission, streaming/empty/HEAD lifecycles, edits and revoked membership.
No frontend, authorization, durability, encoding-negotiation or jitter policy was changed.

## Cache misses and compressor trial

The separate `bench/compression` module compares stdlib and klauspost/compress v1.19.0
without adding a production dependency. Its input is the full validated 374,036-byte
room response in [compressor-room.html.gz](compressor-room.html.gz), not a partial renderer
or synthetic repeated string. Both implementations decode exactly to the input and produce
21,289-byte members at level 6.

Five single-core repetitions, alternating implementation order:

| Compressor operation | stdlib µs/op | klauspost µs/op |
|---|---:|---:|
| Reuse/reset writer; compress every time | 662.8 | 524.3 |
| Construct writer; compress every time | 684.7 | 546.9 |

For this body, klauspost's reset path takes about 21% less time. That is much smaller than
avoiding compression, and this trial does not establish its behavior across small writes,
sidebars or other compression levels. The application retains stdlib pending broader miss
workloads; only the benchmark module depends on klauspost.

The production memo microbenchmark measured a median 134.8 µs and zero allocations for a
warm hit **including the full-body hash**; a fresh-cache first request with an already
warmed writer pool took 787.6 µs, 88,453 bytes and 20 allocations. Compression alone took 654.0 µs, 65,396 bytes and ten
allocations. These exclude application work and networking. [Raw samples](compressors.txt)
include every repetition, new/reset operations, allocation counts and output sizes.

The focused [upstream profile](profiles/upstream-top.txt) puts 60% of CPU in the level-6
encoder alone. In the [reuse profile](profiles/prepared-top.txt), SHA-256 becomes 40% of CPU
and compression drops out of the leading costs. The next justified target is authoritative
representation/part identities, not a template or SQLite-driver rewrite. These profiles
precede GET-only admission; their measured GET path is unchanged.

## Workload validation and limits

- 18 completed final application runs, 126 HTTP samples, zero HTTP errors.
- 287,643 acknowledged HTTP posts and 2,490 active-room posts verified in messages and FTS.
  Active-room readers run while an independent writer targets 20 commits/sec; actual counts
  are recorded. This is not an every-read cache-miss scenario. Final room responses contain
  newly committed content; cold misses are measured separately above.
- Encoded responses decode to full identity responses. The original room cursor is the sole
  field masked for this comparison because of its known per-request timestamp defect.
  Initial message/room IDs and static bytes match across candidates. The sidebar is the
  upstream Go bare-frame response, not a complete Rails/Rust sidebar page: Go's sidebar
  template omits the layout that the reference's sidebar extends. Its rows measure the same
  existing Go response across candidates; they must not be used for full-application or
  cross-port sidebar comparisons. PR #4 proposes a layout fix. The unchanged room is
  374,036 bytes plain and about 21.3 KB gzip; history is 342,444 bytes plain and 12,819 gzip.
- 18 Cable samples, 100 clients, with and without permessage-deflate: all clients subscribed;
  all 130,332 saturated and 540 paced messages reached every client. Fanout stayed near
  1,530 complete messages/sec uncompressed and 1,400 compressed. No Cable optimization or
  high-client-count claim is made.
- The final container has networking disabled except loopback. External integration delivery,
  TLS/ACME, media throughput, browsers, Rust interoperability and pinned media byte goldens
  were not remeasured. The full Go vet/race gates and upstream WebSocket race tests pass on
  both [Mac](check-mac.txt) and [Linux](check-linux.txt). The first Linux attempt lacked the
  zstd CLI; its [environment failure](check-linux-env-failure.txt) is retained, and the rerun
  passed after installing it. A read-only fresh-context review found no lifecycle defects.
- `preflight-*` includes initial harness runs. Identity completed, gzip was interrupted; tests
  and builds overlapped those runs. They are retained for provenance and excluded from every
  table above. Profile samples are also excluded from final throughput tables.

## Reproduce

Use Linux with Go 1.27.1, CGO, libvips/libzstd development packages, the zstd CLI, ffmpeg,
Python, taskset and Rust for the pinned load generator. [harness.json](harness.json) records
the Docker recipe, exact source commits, binary hashes and compressor commands.

Generate the canonical seed without changing reference sources:

```sh
git submodule update --init --recursive
export PARITY_SEED_DIR="$PWD/.cache/parity-seeds"
reference/parity/bin/reference build
reference/parity/bin/seed build default
python3 bin/build-assets
CARGO_TARGET_DIR="$PWD/.cache/loadgen" cargo build --locked --release \
  --manifest-path reference/bench/loadgen/Cargo.toml
```

Build each manifest source commit in a separate directory with generated assets copied into
`assets/generated`, using the recorded build flags. Then run each encoding sequentially:

```sh
bench/application --rust-root reference --seed .cache/parity-seeds/default \
  --loadgen .cache/loadgen/release/loadgen \
  --candidate upstream=/path/to/upstream --candidate cursor=/path/to/cursor \
  --candidate prepared=/path/to/prepared --apps upstream cursor prepared \
  --listener public --gzip 1 \
  --routes room_show active_room messages_page sidebar search static_css post_message \
  --concurrency 16 --reps 3 --seconds 5 --server-cpus 0-3 --loadgen-cpus 4-7 \
  --cable-clients 100 --deflate 0 1 --cable-seconds 5 --upload-reps 0 \
  --out bench/results/my-gzip-run
```

Repeat with `--gzip 0 --cable-clients` and a new output directory for identity. Raw access
and load-generator logs are losslessly gzip-compressed after timing; the application logger
remains enabled during measurement. Different hosts, libraries and physical CPU topology
will produce different absolute numbers.
