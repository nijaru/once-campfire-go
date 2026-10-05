# Single-owner message-list payloads

Production `fd6517b` removes the duplicate HTML string retained beside each message-list
byte Part. Lists now allocate an exact-sized, unpooled byte buffer and transfer ownership
to the Part before cache admission. Individual message HTML and the three room-shell Parts
remain unchanged. The private list API returns a Part directly; the obsolete HTML fallback
and namespace-based payload construction are removed.

The LRU charges the payload once, plus key/base metadata. Its default 32 MiB budget,
quarter-budget admission threshold, first-value retention and 75% pruning policy remain.
Removing duplicate storage lets more list entries fit within that budget; it does not promise
that the process will use less memory. Disabled/oversized caches return the same prepared
Part without retaining it, and eviction cannot mutate outstanding responses.

## Measurements

Same Linux/arm64 OrbStack VM, fresh seed, public HTTP/1.1, access logging, 16 clients,
server VM CPUs 0–3/GOMAXPROCS=4 and client CPUs 4–7. Three rotating repetitions per encoding,
five-second samples after two-second warmups. Baseline is the preceding bounded-shell build
`9e5c349`, not the original application. No tests, builds or agent work ran during timing. Earlier reports mislabeled the VM as
Docker Desktop; their wording is corrected to OrbStack, matching the retained raw host
manifests. No historical measurements were changed.

| Encoding/route | Baseline req/s | Owned req/s | CPU µs/request, baseline → owned |
|---|---:|---:|---:|
| Identity room | 15,124 | 14,687 | 145.6 → 147.7 |
| Identity active room | 19,742 | 19,107 | 131.4 → 132.5 |
| Identity history | 22,306 | 21,467 | 122.7 → 124.3 |
| Identity search | 15,882 | 15,326 | 208.8 → 210.5 |
| Gzip room | 21,782 | 21,465 | 133.7 → 134.4 |
| Gzip active room | 27,148 | 28,708 | 122.7 → 120.7 |
| Gzip history | 31,181 | 31,074 | 118.1 → 117.5 |
| Gzip search | 15,955 | 15,957 | 205.9 → 206.4 |

The initial identity sequence was slower across all four routes. A separate ten-second
identity check reversed that direction, so neither sequence establishes a broad speedup
or a consistent regression:

| Route | Baseline req/s | Owned req/s | CPU µs/request, baseline → owned |
|---|---:|---:|---:|
| Room | 14,644 | 14,877 | 153.0 → 147.7 |
| Active room | 18,767 | 19,579 | 134.7 → 131.7 |
| History | 20,903 | 21,994 | 127.4 → 122.4 |
| Search | 14,830 | 15,276 | 216.3 → 211.0 |

Post-four-route median whole-process Pss, MiB:

| Sequence | Baseline | Owned |
|---|---:|---:|
| Five-second identity | 123.6 | 112.8 |
| Five-second gzip | 116.4 | 123.2 |
| Ten-second identity | 107.9 | 118.1 |

Resident memory did not reliably decrease. These observations include GC headroom and the
changed number of retained lists. The concrete gain is one retained payload instead of two,
plus simpler ownership/admission; the change is retained on that basis, not a throughput claim.
Do not combine these ratios with historical experiments or compare Pss across workload scopes.

All 18 runs/72 HTTP samples had zero errors; all 3,078 active posts persisted and matched FTS.
Complete decoded gzip/identity checks passed without cursor masking. Initial rooms remain
374,036 plain/21,289 gzip bytes. Native and Linux vet/race gates passed, including the
WebSocket fork. Tests cover admission at its boundary, disabled/oversized caches, single
payload accounting, eviction with live Parts, edited messages, exact bytes and unchanged
wire validators/conditional responses.

## Reproduce and limits

[harness.json](harness.json) pins production, baseline, exact binary hashes and environment.
Raw results and lossless logs are in [identity](final-0/raw.json), [gzip](final-1/raw.json)
and the longer [identity check](identity-check/raw.json). Native/Linux checks are retained.

```sh
bench/application --rust-root reference --seed .cache/parity-seeds/default \
  --loadgen .cache/linux/loadgen \
  --candidate baseline=/path/to/bounded-shell --candidate owned=/path/to/owned \
  --apps baseline owned --listener public --gzip 1 \
  --routes room_show active_room messages_page search --concurrency 16 \
  --reps 3 --seconds 5 --server-cpus 0-3 --loadgen-cpus 4-7 \
  --cable-clients --upload-reps 0 --out bench/results/my-ownership-run
```

Repeat with `--gzip 0`; the longer check used `--seconds 10 --apps owned baseline`.
These are VM Go-to-Go measurements, not native AMD or Go/Rust results. No browser, Cable,
media, TLS or external-integration throughput was remeasured. Authorization, fresh queries,
cookies, cancellation, gzip bounds and logging behavior are unchanged.
