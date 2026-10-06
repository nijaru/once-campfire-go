# Search layouts and remaining database candidates

`9a1ba87` retains the immutable search layout around the freshly selected message Part. One typed input owns both template execution and cache identity, including the full nested user/account, query, count, recent searches, return room and request/layout dependencies. Fresh authorization and SQL observations still happen before lookup. Room and search share a small `templateShell` payload owner; payload bytes are charged once plus slice/Part metadata. Search with disabled fragment retention keeps the original template path.

The final matched six-route trial puts warm search throughput close to Rust and active search ahead, **not** the whole application at parity. Rust still leads room and writes, usually tails, and process memory. The normal reader statement owner remains unchanged: replacing it with the driver's cache was not worth the measured tail regressions. No index or checkpoint policy changed.

## Conditions and verification

Same Apple M3 Max / macOS 26.6.2 / OrbStack Linux ARM64 VM and toolchain image as [the previous phase](../search-hydration-20261006/README.md). Go 1.27.1, Rust 1.98.0; normal dependencies retain SQLite 3.53.4/3.53.2. Four application CPUs 0–3, client CPUs 4–7, public HTTP/1.1, logging enabled, identical fresh parity seeds, three sequential rotating repetitions, two-second warmups, five-second samples, 16 HTTP clients. All review/build/test work finished before timing; live logs stayed in container `/tmp`, were compressed at level 9 after timing, then copied to the host. No synthetic PGO or private SQLite fusion is enabled.

Full native `bin/check`, Linux vet/race for all packages and the WebSocket fork passed. Tests compare cached shells against complete uncached template bytes across changed dependencies, preserve the old three-Part wire validator, and exercise disabled/oversized retention, payload accounting and captured Parts after eviction. An authenticated HTTP test verifies unchanged 304s, newly recorded recent searches, changed return-room cookies and revoked return-room membership. Existing search snapshot/count/auth and response/header/compression tests remain intact. Native desktop SSH still times out; there is no native performance confirmation.

Final suites retain **120 HTTP samples, zero errors, 326,024 HTTP posts plus 5,006 concurrent active posts persisted and FTS-verified**. Every initial search has 13 unique IDs and displayed count 13; final active searches have 100 unique IDs and count 100. Every encoding check compares complete decoded bytes, without cursor masking. Representative responses remain different across implementations: Go search 135,497 plain bytes / Rust 149,625; final active search about 879 KB / 983 KB. Cross-language checks validate the result contract, not equal HTML work.

## Final six-route comparison

Median requests/sec, identity / gzip, from `final/default-{0,1}`:

| Workload | Baseline Go `76f6c3c` | Search-shell Go | Pinned Rust |
|---|---:|---:|---:|
| Room | 15,298 / 22,431 | 16,100 / 22,040 | 19,244 / 28,307 |
| Active room | 20,542 / 29,528 | 20,775 / 29,829 | 16,205 / 26,853 |
| History | 22,650 / 32,522 | 22,957 / 32,858 | 20,484 / 29,319 |
| Search | 19,255 / 19,757 | 26,225 / 26,561 | 26,484 / 26,734 |
| Active search | 9,008 / 13,082 | 10,814 / 16,591 | 9,660 / 14,083 |
| Post | 2,322 / 2,272 | 2,329 / 2,271 | 3,212 / 3,183 |

Median CPU µs/request and p99 ms, gzip:

| Workload | Baseline CPU | Shell CPU | Rust CPU | Baseline p99 | Shell p99 | Rust p99 |
|---|---:|---:|---:|---:|---:|---:|
| Room | 129.78 | 126.12 | 100.41 | 5.223 | 5.355 | 1.886 |
| Active room | 115.39 | 113.49 | 106.42 | 2.491 | 2.399 | 1.973 |
| History | 111.59 | 110.87 | 92.58 | 1.578 | 1.565 | 1.793 |
| Search | 171.71 | 111.96 | 105.02 | 3.627 | 2.065 | 1.278 |
| Active search | 246.08 | 183.32 | 198.79 | 5.843 | 4.447 | 2.321 |
| Post | 666.14 | 668.60 | 765.22 | 28.431 | 28.751 | 11.239 |

Search CPU drops about 35% and throughput rises 34% in this scope; active-search CPU falls about 25% and throughput rises 27%. Identity gains are comparable. The room/control rows are broadly unchanged, not additional claimed speedups. Search's similar throughput does not erase Rust's lower p99, larger returned HTML, or the dependence on warm cached representations. At 16 clients in this VM, Rust is still about 1.28× faster on gzip room and 1.40× on posts.

Whole-process PSS after this **six-route scope** is baseline/shell/Rust **134.0/132.5/111.5 MiB identity**, **152.9/156.2/133.5 MiB gzip**. Search-shell retention slightly increases gzip PSS. This is neither a cache-only allocation measurement nor comparable to the previous two-route PSS totals. Five-second, three-repeat VM observations are not a production tail distribution or evidence of native/universal parity.

`final/disabled-{0,1}` uses zero fragment retention and only search. Baseline/shell median req/s is **855/866 identity**, **814/847 gzip**; p99 **59.90/62.59**, **64.58/66.11 ms**. The shell optimization is bypassed entirely; these mixed CPU/tail observations establish no cold speedup.

## Earlier shell trial and failed runs

`prototype/` retains the earlier search-only comparison and its exact patch. With default retention, gzip unchanged search was 11,845→20,700 req/s, active 12,500→15,950; corresponding CPU 209.37→128.10 and 245.32→185.94 µs/request. Its fragment-disabled path unnecessarily prepared an unretained shell and regressed 3–6% throughput; that path was removed before production.

The final six-route scope warms room/history before search and has a different cache/write population. Its baseline is therefore not numerically interchangeable with the search-only trial or older 1/16/64-client runs. Do not multiply historical gains.

Two initial final-suite attempts aborted because the harness counted **all** earlier `bench active` FTS rows when validating the second active workload. The failure was a validation bug, not missing application writes (`139` new saved versus `277` accumulated indexed). `8d49527` records the initial indexed count and verifies each workload's delta, retaining the same acknowledged-write equality. This reproduces the original failure and passes both active workloads in the final real application runs. `aborted/` retains the second attempt and traceback, excluded from all tables/totals; the first attempt's container was removed before its temporary logs were recovered. The final runner copies bounded evidence even on failure. No failed response was timed as successful application work.

The first Linux gate missed the staged source's relative reference-fixture symlink. Its environment failure log is retained; corrected final gates passed. These failures are not silently replaced with successful logs.

## Reader statement-cache replacement: rejected

`rejected-reader/` replaces the shared first-256 `sql.Stmt` map with direct `*sql.DB` reads and normal go-sqlite3 `_stmt_cache_size=64` per reader connection. It removes the old owner instead of layering another cache. Writer capacity 64, reader pool limits, read-only/query-only options, cancellation and durability are unchanged. The patch applies against `76f6c3c`; dependencies/module caches were not modified.

Default-retention, five-route gzip medians:

| Workload | Baseline req/s | Driver-cache req/s | Baseline CPU | Driver CPU | Baseline p99 | Driver p99 |
|---|---:|---:|---:|---:|---:|---:|
| Room | 22,390 | 23,314 | 129.47 | 124.46 | 5.047 | 5.539 |
| Active room | 29,929 | 31,052 | 115.60 | 109.15 | 2.251 | 3.033 |
| History | 32,485 | 32,803 | 113.02 | 108.67 | 1.556 | 2.579 |
| Search | 19,787 | 19,721 | 170.69 | 168.81 | 3.709 | 4.467 |
| Post | 2,253 | 2,248 | 671.39 | 673.30 | 28.191 | 28.335 |

Identity room/active/history throughput rises about 2–4%, search/posts do not. Memory is essentially unchanged (gzip baseline/driver **146.1/146.2 MiB** in this five-route scope). Small read CPU savings do not justify repeated worse tails or claim a write improvement; the driver has different LRU/reset/eager-first-step costs rather than magically removing preparation work. Retain the existing statement owner.

All **60** samples had zero errors; **192,205** HTTP posts and **1,661** active posts persisted and matched FTS. Full Linux application gates passed. Native/Linux database probes also passed churn beyond 64 statement shapes, bindings/value ownership, changed column counts after schema reprepare, committed data visibility, cancellation during row iteration and subsequent connection reuse. The private probe is archived as `.go.txt`, not added to the production suite.

## Covering index and checkpoint decisions

A source/planner review and copied-seed probe investigated `(room_id,created_at,id,updated_at)` replacing the existing `(room_id,created_at)` index. `index_probe.py` / `index_probe.json` are reproducible structural evidence, using native Python SQLite **3.53.4**, not HTTP/Go timing. On the 169-message seed (largest room 131), the reference projection becomes covering, while a full-record projection still needs table access. Explicit ID before updated_at preserves tied-timestamp traversal; selected references and tied IDs match. No pagination semantics are changed by the probe.

The index grows **12→16 KiB**, payload **5,239→9,464 bytes**. For 200 isolated message+room touch transactions, WAL frames grow **405→614 (+52%)**. Autocheckpoint was disabled **only in disposable probe DBs** to isolate amplification. Edits and boosts touch message versions; paying that additional index maintenance to avoid a narrow table lookup is not justified here. No HTTP gain is claimed or measured, and the schema remains unchanged. A future read-heavy deployment could evaluate this trade-off with its own representative writes/storage requirements; adoption would need deliberate transactional replacement of the Go-owned index, not `IF NOT EXISTS` retaining an obsolete definition or a second layered index.

Write tails alone do not establish that autocheckpoint is the bottleneck. Available profiles do not correlate commit queues, checkpoint duration and WAL growth. Moving checkpoints off the writer, disabling autocheckpoint, weakening synchronization or accepting unbounded WAL growth is therefore **not justified**. Current checkpoint/durability/reader/shutdown behavior stays intact.

## Stopping assessment and reproduction

The evaluated straightforward changes have reached diminishing returns, not an absolute performance ceiling. More cold-template/rich-text work, SQLite boundary work, Cable/high-client memory work, or checkpoint instrumentation may expose further gains, but would need new representative evidence and larger scoped changes. The driver fusion, reader-cache, JSON batching, timestamp parser and default-PGO trials did not earn retention. Native confirmation and fresh matched Ruby/Elixir/Cable measurements remain absent; published AMD ratios cannot be transferred to this VM or multiplied by historical fork gains. The known sidebar layout mismatch remains excluded.

`manifest.json` pins production source, exact staged source/binary hashes, harness, checks and totals. `application` is the corrected final harness; prototype/reader trials used its pre-fix version from `76f6c3c`, where only one active workload was selected. Raw JSON, response captures, metadata, compressed logs, rejected patches/probes and generated reports remain in their named directories. Both experimental patches apply against `76f6c3c` with `git apply --unidiff-zero`. Representative final command inside the matched toolchain container:

```sh
bench/application --rust-root reference --seed .cache/parity-seeds/default \
  --loadgen .cache/linux/loadgen \
  --candidate base=.cache/linux/final-message-hydrate \
  --candidate shell=.cache/linux/final-search-shell \
  --candidate rust=.cache/rust-linux-target/release/campfire \
  --apps base shell rust --listener public --gzip 1 \
  --routes room_show active_room messages_page search active_search post_message \
  --concurrency 16 --reps 3 --seconds 5 --server-cpus 0-3 --loadgen-cpus 4-7 \
  --cable-clients --upload-reps 0 --out /tmp/search-shell-final
```

Use fresh output per encoding/scope; build/test/review first, compress logs only after timing finishes, then copy results out. The final disabled scope selects only `search`, baseline/shell, and `--fragment-cache-mb 0`.
