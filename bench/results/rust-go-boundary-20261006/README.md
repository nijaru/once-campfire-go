# Matched Go/Rust HTTP comparison and SQLite crossing trial

Measured 2026-10-06 on the same OrbStack Linux/arm64 VM on an Apple M3 Max.
**These are not native desktop or published AMD measurements.** Native Intel desktop
SSH/Tailscale access remained unavailable. The user approved VM trials instead.

Go already has higher warm active-room/history throughput in this workload. Rust
still has better room/search/write throughput, generally lower tail latency, and
lower observed process memory. Higher Go throughput does not necessarily mean
lower CPU/request. The SQLite crossing candidate is **not shipped**: small CPU
savings and inconsistent throughput do not justify maintaining a driver fork.

## Sources, builds, and workload

- Initial Go: `fd6517b` production; initial harness `95160b3`.
- Corrected search Go and harness: `eb7adad`. The correction is described below.
- Rust: pinned `64f86353021145b63849fb1cd93adeb08f3b8dbb`; Rails
  `90b330024dec3e757c79b6a7e6568f93da8e3148`. No reference changes.
- Go 1.27.1, CGO, `sqlite_fts5`, trimpath, buildvcs=false, no PGO.
  Rust 1.98.0, locked release, fat LTO/one codegen unit, no target-native flags.
- Common container media toolchain: libvips 8.16.1. Each application's pinned
  SQLite remains intact: Go 3.53.4, Rust 3.53.2. This is not a controlled comparison
  of identical SQLite builds, allocators, or language runtimes.
- Same disposable parity seeds; public HTTP/1.1; identity/gzip; logging enabled;
  four application workers; server VM CPUs 0–3, client 4–7. VM affinity does not
  pin physical Apple cores. No builds, tests, or child work during timing.
- Concurrency 1/16/64, three rotating sequential repetitions, five-second samples,
  two-second warmups, active writes targeted at 20/sec. No Cable/uploads/media/TLS.
- Sidebar is excluded: Go's bare frame is not the reference's complete layout.
  Complete decoded gzip/identity bodies match **within** each app without cursor
  masking; selected message IDs agree across apps. HTML is not byte-identical
  across languages and strict overall behavior/HTML parity is not established.

Raw bodies differ in size: initial room Go 374,036 plain/21,289 gzip bytes, Rust
416,139 plain/24,235 gzip bytes; corrected search Go 135,497/9,462, Rust
149,625/9,767. These are full responses from the implementations, not equivalently
sized synthetic payloads.

## Search regression and accepted evidence

Inspection found a bug introduced by direct record input (`88ada08`): the search
badge still used `len .Messages`, but cached rendering no longer populated that
view slice. Messages rendered correctly while the badge showed zero. The initial
harness checked selected message IDs, not this badge; **its initial Go search
cells are not accepted as correct search comparisons** and remain raw evidence.

`eb7adad` supplies the fresh result count explicitly. Its authenticated HTTP test
failed before the fix (0 instead of 2), then passed for misses/hits/deletion/empty
results. Native `bin/check` and Linux vet/race, including WebSocket tests, passed.
The harness now checks the badge against rendered IDs; an old-binary probe fails
before timing. All corrected search runs show the expected 13 results.

Search was rerun separately with all three apps and both encodings. Its table
cells come from `search-*`, not `final-*`; memory and setup scopes differ. Do not
attribute the difference between those separately warmed suites to the count fix.
A separate old/fixed search probe and profiles do not reproduce such a causal
regression; their raw data includes a large old-binary first-run outlier.

The initial suite has 270 timing samples; corrected search adds 54. Tables select
216 valid non-search initial samples plus 54 corrected search samples. All load
samples have zero HTTP errors. Initial active writes (6,059) and HTTP posts
(702,634, including warmups) were persisted and FTS-indexed without loss.

## Median public gzip results, 16 clients

| Route | Go req/s | Rust req/s | Go CPU µs/request | Rust CPU µs/request | Go p99 ms | Rust p99 ms |
|---|---:|---:|---:|---:|---:|---:|
| Room | 20,887 | 26,254 | 136.60 | 105.40 | 5.507 | 1.924 |
| Active room | 28,251 | 24,359 | 123.03 | 113.63 | 2.349 | 2.157 |
| History | 30,893 | 25,987 | 118.54 | 100.73 | 1.782 | 2.022 |
| Search, corrected/separate suite | 10,326 | 24,016 | 234.73 | 117.46 | 8.807 | 1.445 |
| Post | 2,268 | 3,136 | 682.75 | 785.50 | 29.135 | 10.647 |

The write row illustrates why throughput and CPU alone are insufficient: Rust
serves more posts with lower tails despite greater measured process CPU/request.
This is a reason to profile writer scheduling/checkpoints, not proof of their cause.

## Median identity results, 16 clients

| Route | Go req/s | Rust req/s | Go CPU µs/request | Rust CPU µs/request | Go p99 ms | Rust p99 ms |
|---|---:|---:|---:|---:|---:|---:|
| Room | 15,002 | 18,902 | 147.95 | 135.47 | 6.055 | 2.797 |
| Active room | 20,325 | 15,629 | 132.09 | 143.62 | 4.131 | 3.041 |
| History | 22,574 | 19,803 | 123.90 | 127.42 | 4.067 | 2.557 |
| Search, corrected/separate suite | 10,153 | 23,672 | 248.72 | 122.53 | 8.751 | 1.525 |
| Post | 2,300 | 3,191 | 649.59 | 748.04 | 27.951 | 11.191 |

Concurrency 1 and 64 samples, ranges, and memory are in each generated report/raw
file. At 64 gzip clients, room throughput is 20,060 Go/26,147 Rust, corrected
search 9,355/28,641, and posts 2,177/3,138. Post p99 is 132.99/27.86 ms.

After the initial five-route, three-concurrency suite, median process Pss is
135.4/116.8 MiB Go/Rust identity and 166.1/139.1 gzip. These include caches, heaps,
SQLite, queues and differently sized acknowledged-write populations; they are not
cache-size or per-connection figures. Do not compare them with earlier smaller
workload scopes or the separate search-only measurements.

## Rejected fused SQLite boundary

`rejected-fusion/driver.patch` applies to a private copy of go-sqlite3 v1.14.52.
It combines ordinary step and column extraction into one CGO call, retaining the
C-owned row buffer and Go copies of borrowed text/BLOB values. Approximately 81
step/extraction crossings for 40 ordinary rows become 41; cached eager first-row
metadata/capture remains separate. No SQL, query freshness, timestamp conversion,
or database/sql public API changes. No JSON aggregation or query/result cache.

Review found no concrete introduced defect. Driver vet/race passed with default
and FTS5 tags; additional mixed-value/lifetime, cached reprepare growth/shrinkage,
subsequent-row cancellation/reuse, subsequent-row error/reuse and zero-column/
no-Next cases passed. The unlock-notify race suite passed with vet disabled;
its existing goroutine `t.Fatal` vet failure was reproduced on the original.
Integrated application Linux vet/race and WebSocket race passed. The production
module cache, go.mod and dependency remain unmodified.

At 16 gzip clients, initial Go/fused medians were room 20,887/21,031 (+0.7%),
active 28,251/28,312 (+0.2%), history 30,893/31,494 (+1.9%), posts 2,268/2,284
(+0.7%). Room CPU fell 136.60→130.18 µs; other reductions were smaller. Identity
room/active throughput worsened, and 64-client outcomes were mixed. Corrected
search was 10,326/10,172 gzip and 10,153/10,271 identity, with no reliable CPU win.
Those outcomes do not warrant a permanent fork or patch-at-build system.

## Remaining application-owned work

Separate corrected search profiling attributes about 23% cumulative CPU to
HTML template execution, 19% to full search retrieval/scanning, and 7% to a room
list queried by the controller but not used by the search/layout templates.
Percentages overlap; they are not additive predicted savings. Narrow fresh search
references and removing unused preparation deserve experiments before more
SQLite internals. Typed immutable search-shell reuse would require the same key/
render dependency discipline as rooms, with fresh query/count/recent-search data.

Covering pagination indexes and connection-local read statement ownership remain
unmeasured hypotheses. Background checkpoints need evidence of checkpoint-related
write stalls, bounded WAL growth, unchanged durability, and shutdown verification.
Session observations, cancellation and access logging must not be weakened.

## Limits and reproduction

VM scheduling and short samples introduce substantial variance: retain all ranges,
including the initial gzip history/Go 21,592–31,599 and fused-search 11,058–15,926
req/s ranges. Post-run host inspection found active Spotlight workers and desktop
UI activity, without swap; this does not identify their exact timing or quantify
interference. No native-capacity or universal Go/Rust ratio is established.
There are no fresh Elixir/Ruby measurements and no multiplication of published
AMD figures by this fork's VM speedups. Native confirmation remains blocked.

`manifest.json` records binary hashes, exact staged Go source/assets, pinned
sources/module hashes, patch/tests, seed, toolchain image and lockfile hashes.
`toolchain.Dockerfile` records the Rust image used to extend the existing Go bench
image. `application-before-count` and `application` preserve harness versions;
metadata in each directory records arguments/environment. Logs are losslessly
compressed only after timing. Profiles/probes are excluded from result tables.

The common command shape is:

```sh
bench/application --rust-root reference --seed .cache/parity-seeds/default \
  --loadgen .cache/linux/loadgen \
  --candidate rust=.cache/rust-linux-target/release/campfire \
  --candidate go=.cache/linux/final-ownership \
  --candidate fused=.cache/linux/final-fused --apps rust go fused \
  --listener public --gzip 1 --routes room_show active_room messages_page search post_message \
  --concurrency 1 16 64 --reps 3 --seconds 5 \
  --server-cpus 0-3 --loadgen-cpus 4-7 --cable-clients --upload-reps 0 --out OUT
```

Run once per encoding. For accepted search, use the corrected harness,
`final-search-count`/`final-fused-count` Go binaries and `--routes search`.
These paths are local build artifacts, not distributed binaries. The isolated
replacement is experimental only; reapply the recorded patch to the pinned driver
in a separate source tree, select it with a private modfile, and reproduce the
recorded build flags. No PRs, upstream commits or pushes were made.
