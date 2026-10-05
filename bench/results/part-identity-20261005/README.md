# Strong part identities: gzip iteration

Completed response parts now own their bytes and a constructor-computed SHA-256 digest.
The compression memo hashes ordered lengths and digests instead of scanning the complete
body again. Message-list parts retain their existing digest; ordinary buffered responses
share the digest already needed for their validator. Authentication, authorization, queries,
rendering, cookies and headers still run on every request. No SQLite or template changes.

Compared with the previous prepared-gzip build, gzip throughput improved 24% for unchanged
rooms, 60% for active rooms and 82% for history. Identity results stayed within about ±2%.
These measurements use the same Apple M3 Max OrbStack Linux/arm64 VM as the
[previous experiment](../prepared-gzip-20261005/README.md), not native AMD or Rust comparisons.

## Application results

Median requests/sec over three alternating repetitions, 16 clients, five-second samples
following two-second warmups. Public HTTP/1.1; server VM CPUs 0–3 with GOMAXPROCS=4,
load generator CPUs 4–7. Default public access logging remains enabled. No builds, tests
or other agent work ran during timing. VM affinity does not pin physical Apple cores.

| Gzip workload | Previous memo | Part identities | Ratio | CPU µs/request, previous → parts |
|---|---:|---:|---:|---:|
| Room | 8,172 | 10,165 | 1.24× | 361 → 219 |
| Active room, target 20 posts/sec | 11,818 | 18,884 | 1.60× | 315 → 179 |
| History | 15,009 | 27,320 | 1.82× | 257 → 133 |
| Sidebar, existing bare frame | 15,432 | 17,160 | 1.11× | 153 → 149 |
| Search | 13,943 | 15,391 | 1.10× | 257 → 211 |
| Post message | 2,263 | 2,287 | 1.01× | 679 → 667 |
| Static CSS | 103,830 | 102,026 | 0.98× | 18 → 18 |

The sidebar ranges overlap substantially; its modest CPU reduction is clearer than its
throughput result. It remains a bare-frame response, not Rails/Rust full-layout parity.
Post/static differences are small and not evidence of a useful optimization there.

| Identity workload | Previous memo | Part identities |
|---|---:|---:|
| Room | 8,735 | 8,749 |
| Active room | 15,935 | 16,154 |
| History | 21,154 | 20,887 |
| Sidebar, existing bare frame | 16,301 | 15,968 |
| Search | 15,363 | 15,054 |
| Post message | 2,310 | 2,329 |
| Static CSS | 102,716 | 102,293 |

Median post-HTTP Pss was 149.2 → 150.0 MiB in the gzip sequence and 131.7 → 133.6 MiB
in identity. This is whole-process memory after these workloads, not isolated cache usage.
The gzip memo still has its 32 MiB accounted LRU budget, admission limits and 16-flight bound.

[Identity raw samples](final-0/raw.json) and [gzip raw samples](final-1/raw.json) retain ranges,
latencies, CPU/request, body sizes, validation counters and resource measurements. Each
sequence includes its metadata, full initial room bodies and losslessly compressed logs.

## Component costs and profile

On the captured 374,036-byte full room, the [microbenchmark](microbench.txt) measured:

- Cached-part identity plus warm memo lookup: median **71.8 ns**, zero allocations. This does
  **not** read, render, write or deliver 374 KB; no application-throughput claim follows.
- Constructing a new full-body part plus a warm lookup: **122.2 µs**, zero allocations.
- Fresh cache plus full part construction, warmed writer pool: **798.8 µs**.
- Compression alone with already prepared parts: **658.3 µs**. Output remains 21,289 bytes.

The prior warm memo measured 134.8 µs including its full-body hash. Avoiding that scan is
useful only when the renderer already owns a valid immutable part; fresh bytes still require
hashing. A different decomposition of identical bytes may now occupy a separate memo entry.
This sacrifices cross-partition deduplication, not content correctness or memory bounds.

The separate [room CPU profile](profiles/top.txt) puts SHA-256 at 6.7% flat CPU, versus
roughly 40% in the previous prepared build. [Cumulative costs](profiles/cumulative.txt)
include room-shell assembly at 13.7% and string replacement at 9.7%. Cached static shell
parts are the next source-grounded opportunity; speculative query snapshots are not required.
Profile timing is excluded from the result tables.

## Validation and scope

All 12 final application runs and 84 HTTP samples completed with zero HTTP errors.
192,978 acknowledged HTTP posts and 1,670 active-room posts persisted and matched FTS
counts. Encoded responses matched complete identity bytes **without cursor masking**;
initial semantic contracts matched across candidates. Active-room final content was checked.

Mac and Linux [vet/race gates](check-linux.txt) pass, including the WebSocket fork; the
[Mac log](check-mac.txt) includes the asset gate. Tests cover changed bytes in every part
under the same weak ETag, preserved wire validators, edited messages, revoked membership,
per-request headers, streaming, HEAD, bounds and ordinary pooled-buffer A/B/A reuse.
An independent read-only review found no introduced lifecycle or authorization errors.

No Cable code changed; the previous complete-delivery evidence remains relevant, but Cable,
browsers, media goldens, Rust interoperability, TLS/ACME and external integrations were not
remeasured. Networking was disabled except container loopback. Sidebar ID checks do not
prove complete layout parity. The benchmark does not model every read as a cache miss;
active rooms and full fresh-part component costs are separate evidence.

## Reproduce

[harness.json](harness.json) records source commits, binary hashes, build flags and the
previous environment manifest. Build `0902c01` and `53e3601` in separate source directories
with generated assets, using Go 1.27.1, CGO and `-trimpath -buildvcs=false -tags sqlite_fts5`.
Use the same canonical seed, pinned load generator and Linux image as the previous report.
Run encodings sequentially, without tests or compilation alongside the samples:

```sh
bench/application --rust-root reference --seed .cache/parity-seeds/default \
  --loadgen .cache/linux/loadgen \
  --candidate prepared=/path/to/previous --candidate parts=/path/to/parts \
  --apps prepared parts --listener public --gzip 1 \
  --routes room_show active_room messages_page sidebar search static_css post_message \
  --concurrency 16 --reps 3 --seconds 5 --server-cpus 0-3 --loadgen-cpus 4-7 \
  --cable-clients --upload-reps 0 --out bench/results/my-part-run
```

Repeat with `--gzip 0` and a different output directory. Cursor masking is now explicit
`--mask-legacy-room-cursor`, needed only when comparing the original pre-fix binary.
