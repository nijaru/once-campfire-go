# Current fork versus upstream — 2026-10-05

Fresh head-to-head measurements, not ratios assembled from earlier experiments.
Upstream main was `8d2f7f2`; it differs from the retained `504428a` baseline binary's
source only in README.md. The fork's production source is `fd6517b` (documentation HEAD
at measurement: `2e262f5`). Neither binary uses PGO. Unmerged upstream PRs are not included.

## Throughput

Median requests/sec from three rotating repetitions per encoding:

| Route | Upstream gzip | Fork gzip | Ratio | Upstream identity | Fork identity | Ratio |
|---|---:|---:|---:|---:|---:|---:|
| Room | 4,051 | 19,345 | 4.78× | 8,153 | 15,230 | 1.87× |
| Active room | 4,352 | 27,558 | 6.33× | 14,437 | 18,913 | 1.31× |
| History | 5,307 | 29,679 | 5.59× | 19,585 | 21,720 | 1.11× |
| Search | 7,760 | 15,608 | 2.01× | 14,765 | 15,338 | 1.04× |
| Sidebar | 14,087 | 16,753 | 1.19× | 17,276 | 16,558 | 0.96× |
| Post message | 2,202 | 2,226 | 1.01× | 2,301 | 2,327 | 1.01× |
| Static CSS | 104,840 | 104,183 | 0.99× | 102,366 | 103,832 | 1.01× |

Small differences should not be treated as established improvements. Gzip gains are
concentrated in completed HTML GETs, especially room/history bodies. Writes and static
responses are essentially unchanged; identity sidebar is slightly lower in this sequence.
Both Go sidebar implementations still return the same bare-frame workload; these numbers
establish no equivalence with Rust's full sidebar layout.

Gzip room median p99 latency fell from 13.72 to 5.40 ms; history from 9.06 to 2.13 ms.
Room CPU/request fell from 945.7 to 138.4 µs, history from 746.8 to 120.1 µs. Identity
room CPU/request fell from 224.7 to 146.8 µs. These are measured within this comparison.

Median post-seven-route whole-process Pss: identity 134.0 → 132.4 MiB; gzip 141.4 →
147.7 MiB. Gzip's extra approximately 6.2 MiB is an observed trade-off, not an isolated
cache allocation measurement. Totals include GC headroom and the write workload and
must not be compared directly with earlier four-/five-route memory figures.

## Method and checks

OrbStack Linux/arm64 VM on an Apple M3 Max, Go 1.27.1, identical fresh parity seed copies,
public HTTP/1.1 and access logging enabled. Six runs per encoding, 16 clients, five-second
samples after two-second warmups. Server VM CPUs 0–3/GOMAXPROCS=4; client CPUs 4–7.
Active rooms target 20 HTTP commits/sec. No builds, tests or agent workloads ran during
timing. Affinity pins VM vCPUs, not physical host cores. Container networking was disabled
except loopback.

All 12 runs/84 HTTP samples had zero errors. All 189,975 acknowledged message posts and
1,668 active-room posts persisted and matched FTS. Complete decoded response comparison
used the explicit legacy-cursor mask: upstream's known render-time room refresh cursor is
excluded, while the remaining bytes are compared. This exception is necessary for the
pre-fix upstream binary; it is not unconditional full-body parity. The fork's queried-room
cursor correctness and fractional-millisecond truncation remain covered by the permanent
regression suite. No production source changed for this comparison; existing native/Linux
vet/race gates from the final optimization delivery remain applicable.

No native AMD, Rust, Cable, media, browser, TLS or external-integration performance is
claimed. Earlier optimization ratios are not multiplied into this report.

## Reproduce

[harness.json](harness.json) pins source/binary hashes, the upstream documentation-only
difference, environment and validation exception. Full [identity](final-0/raw.json) and
[gzip](final-1/raw.json) samples, metadata, captured rooms and compressed logs are retained.

```sh
bench/application --rust-root reference --seed .cache/parity-seeds/default \
  --loadgen .cache/linux/loadgen \
  --candidate upstream=/path/to/upstream --candidate fork=/path/to/fork \
  --apps upstream fork --listener public --gzip 1 --mask-legacy-room-cursor \
  --routes room_show active_room messages_page search sidebar post_message static_css \
  --concurrency 16 --reps 3 --seconds 5 --server-cpus 0-3 --loadgen-cpus 4-7 \
  --cable-clients --upload-reps 0 --out bench/results/my-upstream-run
```

Repeat with `--gzip 0` and a separate output directory. Build source and generated assets
with the manifest's Go flags, CGO and SQLite FTS5; use the pinned environment/seed/loadgen.
