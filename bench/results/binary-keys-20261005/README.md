# Binary message-version keys

Message and message-list cache keys now encode fixed-width ID, Unix seconds and truncated
microsecond fields instead of formatting a UTC timestamp for each record. Seconds and
fractions stay separate to avoid UnixNano/UnixMicro range overflow. List keys concatenate
20-byte records in order, with separate item/list namespaces. This preserves the existing
UTC microsecond identity, not a new query cache or invalidation policy. No database queries,
response bytes, authorization, cancellation or wire-validator construction changed.

The [previous room profile](../static-shell-20261005/profiles/cumulative.txt) attributed
5.0% cumulative CPU to message-key construction. The [new profile](profiles/cumulative.txt)
puts the whole message-list lookup/render path at about 0.5%, versus 5.6% previously;
allocation is 8.1% cumulative versus 13.5%. These separate profiles are diagnostic samples,
not throughput observations or precise attribution of all allocator savings.

## Application measurements

Three alternating repetitions, 16 clients, five-second samples after two-second warmups.
Public HTTP/1.1, identical fresh seed data, active rooms targeting 20 posts/sec. Same Apple
M3 Max OrbStack Linux/arm64 VM as the previous iterations: server CPUs 0–3 with
GOMAXPROCS=4, load generator CPUs 4–7. Access logging stays enabled. No compilation,
tests or other agent work ran during timing. VM affinity does not pin physical Apple cores.

| Encoding/workload | Static shell | Binary keys | Ratio | CPU µs/request, shell → keys |
|---|---:|---:|---:|---:|
| Identity room | 11,193 | 11,935 | 1.07× | 175 → 164 |
| Identity active room | 18,450 | 18,715 | 1.01× | 149 → 141 |
| Identity history | 21,195 | 21,819 | 1.03× | 137 → 131 |
| Identity search | 15,422 | 15,354 | 1.00× | 210 → 214 |
| Gzip room | 15,215 | 16,769 | 1.10× | 165 → 154 |
| Gzip active room | 23,760 | 25,449 | 1.07× | 142 → 133 |
| Gzip history | 27,448 | 28,308 | 1.03× | 133 → 125 |
| Gzip search | 15,432 | 15,604 | 1.01× | 211 → 210 |

Values are medians from [identity samples](final-0/raw.json) and [gzip samples](final-1/raw.json).
Search is effectively unchanged; small history/identity-active differences should not be
read as a broad application speedup. These ratios compare pinned candidates in this
experiment only, not accumulated improvements across separate runs.

Post-workload Pss was 107.6 → 106.2 MiB for identity and 119.8 → 115.9 MiB for gzip.
This four-route scope differs from previous two/seven-route sequences. It measures the
whole process, including GC headroom, not the exact storage saved by shorter keys. Fragment
and gzip cache budgets remain unchanged. Initial room bytes remain 374,036 plain/21,289 gzip.

## Checks and limits

All 12 final runs and 48 HTTP samples had zero errors. All 1,661 active-room posts persisted
and matched FTS counts. Complete decoded gzip bodies matched identity without cursor masking;
initial response semantics matched across candidates, and final active content was verified.

[Mac](check-mac.txt) and [Linux](check-linux.txt) vet/race gates passed, including the WebSocket
fork. Focused key tests compare against database.Stamp equivalence across time zones,
submicrosecond truncation, microsecond mutations, pre-epoch/zero/far-future dates and ID
changes. They also protect list ordering, record boundaries and namespace separation.
Existing real rendering, edited-message, revocation, cookie, conditional and compression
checks remain intact. Binary keys are internal and are not HTTP validators.

No Cable, browser, media, static/post throughput, TLS or external integrations were remeasured;
those paths were unchanged and covered by the existing gates where available. Container
networking was disabled except loopback. The known sidebar layout gap is unchanged. These
are VM Go-to-Go results, not native AMD measurements or Go/Rust parity comparisons.

## Reproduce

[harness.json](harness.json) records the production commits, exact binary hashes, source
staging and shared environment manifest. Build `048e729` and the binary-key production
commit with their generated assets, Go 1.27.1, CGO and
`-trimpath -buildvcs=false -tags sqlite_fts5`. Use the pinned seed/load-generator/Linux image
from the [initial environment manifest](../prepared-gzip-20261005/harness.json).

```sh
bench/application --rust-root reference --seed .cache/parity-seeds/default \
  --loadgen .cache/linux/loadgen \
  --candidate shell=/path/to/shell --candidate keys=/path/to/keys \
  --apps shell keys --listener public --gzip 1 \
  --routes room_show active_room messages_page search \
  --concurrency 16 --reps 3 --seconds 5 --server-cpus 0-3 --loadgen-cpus 4-7 \
  --cable-clients --upload-reps 0 --out bench/results/my-key-run
```

Repeat with `--gzip 0` and a separate output directory. Profile runs are retained separately
and excluded from the table.
