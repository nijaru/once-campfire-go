# Search references and shared message hydration

Two application-owned changes reduce work without caching search results or authorization:

- `afd7c08` queries fresh matching IDs, room IDs and versions, then retains an existing immutable message-list Part. On a miss it runs the original scoped full search, selecting matching bodies and rows in one SQL statement. The displayed count belongs to those actual results. Disabled fragment retention uses that full query directly. The unused search-controller room-list query is gone; the fresh return-room lookup remains.
- `76f6c3c` hydrates uncached views in their existing output slice. Its request-local room/creator maps are shared across misses instead of rebuilt per message. Cached items remain untouched and ordered; there is no cross-request metadata cache.

The first search design would have hydrated reference misses later with an unscoped ID read. Review caught that an intervening edit could render a body that no longer matched the query. That design was corrected before timing. The shipped miss path preserves the original query/body snapshot; hits retain the actual Part, not a promise to retrieve or hydrate it later.

## Environment and scope

Apple M3 Max, macOS 26.6.2, OrbStack Linux/aarch64 VM, 16 VM CPUs. Same toolchain image and pinned Rust `64f8635` as the [matched boundary comparison](../rust-go-boundary-20261006/README.md): Go 1.27.1, Rust 1.98.0, CGO/FTS5, normal release builds, no PGO or private SQLite driver patch. Go and Rust retain their normal SQLite versions (3.53.4 and 3.53.2). Native `ssh desktop` confirmation remains blocked by connection timeouts.

Every run starts from an identical parity seed copy. Public HTTP/1.1, logging enabled, four application workers on VM CPUs 0–3, loadgen on 4–7, three rotating repetitions, two-second warmups and five-second samples. Applications run sequentially; no builds, tests or child work runs during timing. Output stays in container `/tmp` until timing finishes, then logs are compressed losslessly and copied out, avoiding multi-gigabyte host-visible log creation during sampling. This does not eliminate VM/UI scheduling noise.

These are application measurements, not general language comparisons. Rust returns different, larger HTML: initial search is Go 135,497 / Rust 149,625 plain bytes (9,462 / 9,767 gzip), with the same 13 message IDs and displayed count. Cross-language checks compare result contracts, not unconditional raw HTML equivalence. Within each application, decoded gzip and identity bytes are compared exactly, without legacy cursor masking.

## Fresh-reference search trial

`query-probe/` and `query-edges/` compare count-corrected baseline `eb7adad`, the snapshot-safe reference prototype, and Rust. The prototype still performed the extra reference query with fragment retention disabled; production `afd7c08` bypasses that unnecessary lookup. Do not treat these prototype cold numbers as shipped behavior or multiply gains from this trial by the later hydration comparison.

Median requests/sec and CPU µs/request, 16 clients:

| Encoding / workload | Baseline req/s | References req/s | Rust req/s | Baseline CPU | References CPU | Rust CPU |
|---|---:|---:|---:|---:|---:|---:|
| Identity, unchanged search | 10,250 | 11,895 | 24,597 | 239.42 | 212.35 | 114.39 |
| Gzip, unchanged search | 10,512 | 11,880 | 24,956 | 226.91 | 208.37 | 111.86 |
| Identity, active search | 7,193 | 8,954 | 9,981 | 375.11 | 283.46 | 263.55 |
| Gzip, active search | 8,356 | 12,606 | 14,118 | 340.10 | 247.79 | 197.59 |

Active search posts matching messages at a target 20/sec while reading; final results cap at the query's 100-message limit. This short single-concurrency workload is **not** the longer 1/16/64-client population used below. In the prototype's fragment-disabled trial, throughput changed 831→807 identity and 843→785 gzip; the avoidable second query was removed rather than accepting that regression.

## Final hydration comparison

`final/` compares production query build `afd7c08`, hydration build `76f6c3c`, and the same Rust binary. Default-retention runs exercise search and active search at 1/16/64 clients. Fragment-disabled runs exercise full room, history and search rendering at 16 clients; completed compression memoization remains enabled, so this is **not** an entirely cold application or compressor benchmark.

Median 16-client results:

| Encoding / workload | Query req/s | Hydration req/s | Rust req/s | Query CPU | Hydration CPU | Rust CPU |
|---|---:|---:|---:|---:|---:|---:|
| Identity, unchanged search | 11,495 | 11,852 | 24,267 | 217.92 | 216.96 | 118.23 |
| Gzip, unchanged search | 11,947 | 12,395 | 24,593 | 210.90 | 198.62 | 113.17 |
| Identity, active search | 7,330 | 7,060 | 7,111 | 343.70 | 348.28 | 370.55 |
| Gzip, active search | 9,714 | 9,808 | 10,111 | 293.28 | 288.20 | 283.56 |
| Identity, no fragments, room | 279 | 304 | 2,060 | 9,530.25 | 8,500.65 | 1,284.18 |
| Gzip, no fragments, room | 323 | 352 | 1,402 | 8,906.05 | 7,705.01 | 2,406.59 |
| Identity, no fragments, history | 302 | 303 | 2,200 | 8,844.12 | 8,097.83 | 1,227.42 |
| Gzip, no fragments, history | 347 | 367 | 1,479 | 8,287.51 | 7,091.94 | 2,331.67 |
| Identity, no fragments, search | 677 | 774 | 5,558 | 3,266.10 | 3,038.26 | 507.75 |
| Gzip, no fragments, search | 812 | 867 | 4,221 | 3,058.91 | 2,603.40 | 846.39 |

The useful hydration result is lower fragment-disabled CPU (about 7–15%), not a universal warm throughput gain. Warm active identity is about 4% lower and gzip essentially unchanged. Fragment-disabled p99 latency did not improve: room 123.90→124.35 ms identity / 109.95→113.34 gzip, history 128.83→140.03 / 99.78→111.62, search 67.46→72.26 / 57.92→63.04. We retain this small in-place cleanup for its measured CPU reduction and coherent ownership, **not** as a latency improvement. Raw ranges remain available; three VM samples do not establish a production tail distribution.

In the final matched default run Rust is about **1.98×** faster on unchanged gzip search. At the 100-result active-search stage Go reaches 9,808 versus 10,111 req/s, but Rust p99 is 2.94 versus Go 7.27 ms. It is misleading to call that throughput proximity overall parity. Native confirmation, the room/write gap, and cold-template cost remain open concerns.

Whole-process PSS after the final default two-route, three-concurrency workload is query/hydration/Rust **132.2/125.0/66.6 MiB identity** and **152.6/161.3/70.2 MiB gzip**. Fragment-disabled three-route scope is **58.0/55.3/95.8** and **58.1/59.7/102.8 MiB**. Those scopes have different populations; do not compare their totals as a cache allocation measurement. Hydration's gzip PSS increase remains visible despite lower CPU; no universal memory reduction is claimed.

## Rejected allocation shape and remaining costs

`rejected-batch/` retains the first hydration candidate, which gathered misses into a second records/output slice and merged their positions. Its gains were inconsistent, with repeated worse cold tails. `extra-views.patch` reproduces its production change against `afd7c08`; the final candidate fills the original views in place instead. The rejected trial's attractive active-gzip median was driven by baseline outliers and is not evidence for a broad gain.

A separate fragment-disabled profile of `afd7c08`, excluded from tables, attributes 77.4% cumulative CPU to message-item hydration and 47.5% to template execution (overlapping). Room/user reads are about 2.15%/2.63%. Sharing those reads cannot close the whole Rust gap. Template execution, rich-text processing, SQLite/CGO row handling, fresh session observations, retained allocations and logging remain costs; removing correctness boundaries is not an optimization.

## Verification and artifacts

Across all retained experiments: **414 HTTP samples, zero loadgen errors; 14,598 active posts persisted and matched FTS**. Every initial search has 13 unique rendered IDs and badge 13; every final active search has 100 rendered IDs and badge 100. Warmup/sample bodies are complete, and log output remains enabled. The new harness `active_search` workload and `--fragment-cache-mb 0` make those checks reproducible.

Real-DB regressions cover membership freshness, top-100 ordering including tied timestamps, sanitized/literal terms, timestamp storage formats/errors and cancellation. The snapshot regression edits after reference selection, exercises disabled/oversized/retained cache paths, and renders a captured Part after deletion and eviction. Mixed cached/uncached full records and references produce the same ordered bytes as per-message rendering. Existing full-response validators, refresh, ownership, authorization and compressed-body tests remain intact. Full native `bin/check`, Linux vet/race and websocket race gates passed; logs are in `checks/`.

`manifest.json` pins sources, binaries, image, checks and totals. `application` snapshots the harness. `query-probe/`, `query-edges/`, `rejected-batch/` and `final/` retain raw JSON, metadata, representative body captures, losslessly compressed logs and generated reports. `profiles/` is separate, non-timed evidence. A representative final command is:

```sh
bench/application --rust-root reference --seed .cache/parity-seeds/default \
  --loadgen .cache/linux/loadgen \
  --candidate query=.cache/linux/final-search-query \
  --candidate hydrate=.cache/linux/final-message-hydrate \
  --candidate rust=.cache/rust-linux-target/release/campfire \
  --apps query hydrate rust --listener public --gzip 1 \
  --routes search active_search --concurrency 1 16 64 --reps 3 --seconds 5 \
  --server-cpus 0-3 --loadgen-cpus 4-7 --cable-clients --upload-reps 0 \
  --out /tmp/search-hydration
```

Use a fresh output directory per encoding/scope. The no-fragment suite substitutes `--fragment-cache-mb 0 --routes room_show messages_page search --concurrency 16`. Run only after builds/checks finish, compress logs after all timing, then copy results out.
