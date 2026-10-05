# Immutable room-shell parts

Room shells now retain three immutable, prehashed static parts. Each request inserts its
queried refresh cursor and current message-list part between them, without searching,
copying and hashing the surrounding HTML again. The cache key still includes the entire
remaining page: authorization, account/user/room state, origin, platform, flash and future
page fields remain fresh dependencies. No query snapshots, SQLite changes or template edits.

The preceding [part-identity profile](../part-identity-20261005/profiles/cumulative.txt)
attributed 13.7% of room CPU to shell assembly, including 9.7% to string replacement.
This iteration removes that work from cache hits.

## Results

Three alternating repetitions, public HTTP/1.1, 16 clients, five-second samples after
two-second warmups, identical fresh canonical seeds. Same Apple M3 Max OrbStack
Linux/arm64 VM, server CPUs 0–3/GOMAXPROCS=4 and load generator CPUs 4–7. Access logging
remains enabled. No builds, tests or agent work ran during timing. VM CPU affinity does not
control host physical-core scheduling; these are not native AMD or Rust comparisons.

| Encoding/workload | Previous parts | Static shell parts | Ratio | CPU µs/request, previous → shell |
|---|---:|---:|---:|---:|
| Identity room | 8,760 | 11,106 | 1.27× | 228 → 174 |
| Identity active room | 15,647 | 18,400 | 1.18× | 188 → 148 |
| Gzip room | 10,023 | 15,545 | 1.55× | 217 → 165 |
| Gzip active room | 18,752 | 23,986 | 1.28× | 178 → 141 |

These are medians from [identity raw samples](final-0/raw.json) and
[gzip raw samples](final-1/raw.json), not ratios formed by multiplying earlier experiments.
The room remains 374,036 plain bytes; the initial gzip response remains 21,289 bytes.
Active rooms target 20 posts/sec and contain newly committed messages after timing.

Room-only post-workload Pss increased: identity 114.1 → 123.4 MiB, gzip 124.4 → 129.1 MiB.
This includes GC headroom and all application caches, not isolated shell retention. It must
not be compared directly to the earlier seven-route memory totals. Static shell allocations
are charged to the existing bounded fragment LRU, including their additional metadata;
compression retains its existing 32 MiB budget. The observed memory increase is a trade-off,
not a claimed memory optimization.

All 12 final runs and 24 HTTP samples had zero errors. All 1,660 active-room posts persisted
and matched FTS counts. Full encoded responses matched identity bytes without masking.
The room-shell test compares complete cached output with uncached html/template output,
including cursor/messages, user/role, room, flash/errors, styles, origin, frame, invitation,
stream and cache bypass. [Mac](check-mac.txt) and [Linux](check-linux.txt) vet/race gates pass.

The [new CPU profile](profiles/cumulative.txt) reduces room assembly to 5.1% cumulative;
string replacement is gone. Message reference reads now account for 21.3%, message-key
construction 5.0%, and allocation/SQLite row conversion are visible costs. Profile samples
are separate and excluded from the result table. Room validators use the new five-part
identity, so their opaque values change; unchanged bytes remain conditional and mutations
invalidate them. Other recorded responses retain their existing construction.

Only room and active-room throughput were remeasured. No Cable, browser, media, TLS/ACME
or external-integration throughput claim is made; tests cover other response paths. Current
sidebar output still has the documented pre-existing layout gap. The container had no network
access except loopback. Cache misses are exercised by active writes and cache-bypass tests,
not represented as an every-read miss workload.

## Reproduce

[harness.json](harness.json) records sources `53e3601`/`048e729`, binary hashes, build flags
and the shared environment manifest. Build each source with its generated assets using
Go 1.27.1, CGO and `-trimpath -buildvcs=false -tags sqlite_fts5`. Run each encoding separately:

```sh
bench/application --rust-root reference --seed .cache/parity-seeds/default \
  --loadgen .cache/linux/loadgen \
  --candidate parts=/path/to/parts --candidate shell=/path/to/shell \
  --apps parts shell --listener public --gzip 1 --routes room_show active_room \
  --concurrency 16 --reps 3 --seconds 5 --server-cpus 0-3 --loadgen-cpus 4-7 \
  --cable-clients --upload-reps 0 --out bench/results/my-shell-run
```

Repeat with `--gzip 0` and a new output directory. Use the pinned seed/load-generator/image
from the [initial environment manifest](../prepared-gzip-20261005/harness.json).
