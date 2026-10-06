# Rich-text and render work — 2026-10-06

Three source-backed changes are retained, evaluated in order: DSA, relevant library assessment, then micro-optimizations. No production dependency, schema, durability, authorization or cancellation policy changed.

| Candidate | Commit | Change |
|---|---|---|
| Baseline | `33e21b6` | Previously delivered dense Cable routing; executable source unchanged by report HEAD `9e7a314` |
| Rich-text DSA | `1fe5ece` | Consume Display's owned tree after plain rendering; skip impossible URL/email matching and tag indexes; sanitize identical URL text/destination once |
| Reaction DSA | `270d415` | Prepare eight immutable inner reaction-form bodies using the existing `html/template`; leave message IDs/client IDs in contextually escaped outer forms |
| Final micros | `3b692dc` | Reuse fixed escaping replacers and model-purpose regexes; avoid old-body plaintext immediately overwritten by a supplied replacement body |

All HTML serialize/reparse normalization stages remain. `Process` retains the original attachment tree for verified recipients even when filtering removes an attachment-bearing subtree. URL gates are necessary conditions of the existing regex, not new recognition rules; the email gate runs after URL rewriting. User/host context, resolver calls, filtered errors, attachment fallback and transactional FTS semantics remain intact. The edit-specific discarded parse is removed but was not separately timed.

## Isolated cold-path trials

Fragment retention disabled; room, history and search execute the real full render path for every request. Values are medians of three rotating repetitions at 16 clients, five seconds per sample. Each row compares adjacent candidates **within its own matched trial**, not across phases.

| Phase | Identity room/history/search requests/s | Gzip room/history/search requests/s |
|---|---|---|
| Rich-text DSA | 333.7 → 374.4 / 344.8 → 391.4 / 851.0 → 909.5 | 356.2 → 381.0 / 377.2 → 390.8 / 848.4 → 933.0 |
| Reaction bodies | 370.9 → 396.8 / 384.4 → 401.2 / 881.0 → 947.4 | 374.9 → 420.9 / 386.6 → 433.1 / 888.9 → 965.8 |
| Fixed primitives | 414.6 → 478.1 / 421.5 → 491.8 / 950.8 → 1,148.6 | 414.0 → 462.5 / 422.2 → 486.0 / 914.8 → 1,088.4 |

Rich-text DSA reduces read-route CPU/request about 3–23%; reactions another 11–19%; micros about 3–11% in their respective trials. Read-route p99 improves in all these comparisons. For the last micro phase, identity room/history/search p99 is 99.647 → 91.519 / 98.111 → 82.943 / 55.327 → 48.735 ms; gzip 98.303 → 89.471 / 101.567 → 88.127 / 57.119 → 47.871 ms.

Do not multiply these separate gains into a combined application claim. The phases have different control observations. Memory is mixed: reaction trials' end PSS increases about 4.2 MiB identity and 2.2 MiB gzip; the micro-only cold trial increases 2.4 MiB identity and decreases 0.8 MiB gzip. These are different workload populations—reaction trials include writes; micro-only cold trials do not—and their PSS values cannot be compared across phases. Prepared reaction strings themselves add only fixed startup-owned markup, not a request/authorization cache.

## Final six-route comparison

Baseline / final Go / pinned Rust, normal fragment retention, both encodings, three repetitions, 16 clients. These are direct matched observations, not extrapolations from the cold phases.

| Route | Identity requests/s: baseline / final / Rust | Gzip requests/s: baseline / final / Rust |
|---|---:|---:|
| Room | 15,434 / 15,922 / 20,664 | 22,353 / 21,390 / 28,306 |
| Active room | 21,408 / 21,201 / 17,169 | 29,423 / 29,237 / 26,512 |
| History | 23,700 / 23,871 / 21,391 | 31,919 / 32,082 / 29,027 |
| Search | 26,407 / 26,045 / 26,380 | 26,361 / 26,061 / 26,561 |
| Active search | 11,091 / 11,115 / 9,917 | 16,326 / 16,437 / 13,938 |
| Posts | 2,333 / 2,483 / 3,204 | 2,293 / 2,410 / 3,188 |

Posts improve about 6% identity and 5% gzip. CPU/request is 636.0 → 585.8 µs identity, 669.9 → 625.8 µs gzip; p99 27.631 → 26.399 and 27.519 → 26.591 ms. Rust remains faster at writes and has substantially lower write tails.

Warm read controls are broadly unchanged or mixed, not broadly faster. The short gzip room result regresses about 4%, despite essentially unchanged CPU/request. A named follow-up, three alternating **30-second** room samples per candidate, measures baseline 22,204 versus final 22,085 requests/s (−0.5%), CPU 128.53 → 127.15 µs/request, p99 5.279 → 5.259 ms. That resolves a repeatable large warm-room regression concern, not a room-speedup claim.

The six-route gzip p99 values (baseline / final / Rust, ms) are:

| Route | p99 |
|---|---|
| Room | 5.239 / 5.395 / 1.894 |
| Active room | 2.433 / 2.471 / 2.024 |
| History | 1.700 / 1.699 / 1.751 |
| Search | 1.982 / 2.129 / 1.287 |
| Active search | 4.511 / 4.567 / 2.331 |
| Posts | 27.519 / 26.591 / 10.815 |

End-of-six-route PSS MiB is 133.8 / 136.4 / 111.0 identity and 155.8 / 155.5 / 133.1 gzip. Rust generally retains lower tails and memory. Go remains faster in warm active-room/history/active-search throughput here, near Rust's search throughput, and slower in room/writes. This is not universal or native parity.

## Fresh profiles and libraries

`profile-dsa/` profiles 20 seconds of fragment-disabled gzip search after both DSA changes. URL regexp work no longer appears among the major nodes. `html/template` is 52.53% cumulative, rich-text processing 11.28%, required HTML parsing 3.89%, fixed replacer construction 3.79%, and avatar SignedID generation 3.56%. Cumulative costs overlap. This supports reusing fixed primitives rather than adding a dependency or repeatedly computing the same static reaction content.

`profile-final/` repeats that diagnostic workload after the micros: template execution 58.12%, rich-text processing 6.97%, HTML parsing 4.70%. Fixed replacer construction and SignedID generation no longer appear in the top-100 cumulative listing. Relative shares are not standalone speedup measurements; these profiled runs are excluded from timing medians.

No high-performance library is adopted:

- Alternative HTML parsers would have to preserve malformed-fragment normalization and the existing sanitizer/oracle contract for a relatively small remaining share.
- URL matching no longer merits an alternate regex engine; model-purpose recompilation was fixed at its source using the same regex implementation.
- JSON/SHA libraries are not material hotspots in these profiles. The prior SQLite boundary/reader/index, gzip-library and compiler-PGO trials are not repeated without new evidence.
- Compiled-template libraries are plausible **architecture experiments**, not byte/context-compatible drop-ins. [Quicktemplate's primary README](https://github.com/valyala/quicktemplate) describes generated Go, HTML-escaped output and separate URL/JSON tags. [templ's security documentation](https://templ.guide/security/injection-attacks/) specifies its own class, URL and style policies. Both require template/escaping migration and new build ownership. Their synthetic claims do not establish a gain for these complete application responses. Primary documentation checked 2026-10-06; neither library was benchmarked or installed.

Repeated timestamp helpers and a broader edit-view specialization remain possible follow-ups, but are not demonstrated material residual hotspots. The current pass stops at diminishing returns among evaluated candidates, not an absolute ceiling on Go/template performance. Native matched confirmation is higher-value than another ungrounded dependency swap.

## Verification and artifacts

Across unprofiled suites: **66 process runs, 270 HTTP samples, zero errors, 740,741 HTTP posts and 5,829 active posts**; all **746,570 acknowledged writes** persisted and matched FTS. The final six-route pair alone has 108 samples, 331,757 HTTP posts and 5,001 active posts. `validation-summary.json` retains per-suite totals.

Full decoded responses are checked against the validated snapshot for each process run; every Go candidate retains the same initial byte lengths and message IDs. Search starts with 13 results and ends with 100 under active writes. Initial Go room/history/search is 374,036 / 342,444 / 135,497 bytes; Rust is 416,139 / 383,844 / 149,625. Active search ends around 879 KB Go / 983 KB Rust. Ports/request context vary across processes; this is not cross-language byte equality or a benchmark of an incomplete bare frame.

All 658 rich-text oracle cases match all six fields. Focused coverage extends Display's filtered/error results, verified recipients after foreign-subtree removal, mixed resolver failures, changing user/host context, URL/email boundaries and entity rewriting. Reaction tests compare the original dynamic loop's bytes and contextual escaping, including hostile client IDs and A/B/A reuse. Existing Rails vectors and body/attachment/rollback/FTS tests protect the micros.

Native `bin/check`, Linux vet/race and the WebSocket fork gates pass for each source candidate. Read-only safety reviews finish before timing; no agents, builds or tests run during timing. `checks/` retains logs. One Linux gate observed an apparent truncated EOF in a just-copied bind-mounted oracle test file; native tests and host hashes were correct, the unchanged-source retry passed, and the failed log is retained. The exact infrastructure cause was not established. Native Fedora SSH still times out; its latest log is included.

`manifest.json` records exact commits, staged-source hashes, executable hashes, toolchain image and seed. `richtext/`, `reactions/`, `micro/` and `confirmation/` contain raw JSON, commands, environment metadata and compressed server/load-generator logs. The exact harness is `application.py`; profiles and gate logs are separate. Binaries and immutable staged trees remain in the recorded local `.cache/` locations.

## Environment and reproduction

Apple M3 Max, **OrbStack Linux ARM64 VM**, 16 vCPUs; common Go 1.27.1/Rust 1.98.0 toolchain image. Go SQLite 3.53.4, Rust SQLite 3.53.2—not identical SQLite runtimes. Server GOMAXPROCS/workers 4, CPUs 0–3; load generator CPUs 4–7. Public HTTP/1.1, logging enabled, common parity seed, two-second warmup, sequential rotating repetitions. Logs remain in container `/tmp` until timing finishes, then gzip level 9 and copy-out. No host settings were changed.

Representative final command, repeated for gzip 0/1:

```sh
python3 bench/application --rust-root reference \
  --seed .cache/parity-seeds/default --loadgen .cache/linux/loadgen \
  --candidate base=.cache/linux/final-cable-dense \
  --candidate micro=.cache/linux/final-render-micro \
  --candidate rust=.cache/rust-linux-target/release/campfire \
  --apps base micro rust --listener public --gzip 1 \
  --routes room_show active_room messages_page search active_search post_message \
  --concurrency 16 --reps 3 --seconds 5 \
  --server-cpus 0-3 --loadgen-cpus 4-7 --cable-clients --upload-reps 0 \
  --out /tmp/render-final
```

Cold trials add `--fragment-cache-mb 0` and use the adjacent candidate pairs/routes retained in metadata. The empty `--cable-clients` disables Cable clients; these HTTP suites do **not** remeasure Cable. Earlier [Cable routing results](../cable-routing-20261006/README.md) still disclose worse saturated p99 despite better throughput/paced delivery.

Five-second samples, three repetitions, one small seed and ARM virtualization do not establish production tail behavior, native Intel/AMD ratios, larger-data scaling, TLS/media/browser performance or fresh Ruby/Elixir parity. No upstream issue, PR, message or write was made. Native Fedora testing should rebuild matched candidates and use distinct physical P-cores, longer samples and larger/churning data.
