# Compiler-only PGO trial and stopping point

A PGO build of production `fd6517b` was compared with the same source built normally.
No default profile or build-policy change is shipped. The mixed benchmark profile is
retained for reproduction, not presented as representative production traffic.

The profile combines equal-duration identity/gzip application runs spanning room, active
room, history and search. Training runs are separate from the result tables. Go's
[PGO documentation](https://go.dev/doc/pgo), checked 2026-10-05, recommends representative
production profiles and explains both profile merging and explicit `-pgo` builds.

## Measurements

Same Linux/arm64 OrbStack VM, fresh seeds, public HTTP/1.1, access logging, 16 clients,
server VM CPUs 0–3/GOMAXPROCS=4 and client CPUs 4–7. Three rotating repetitions per encoding,
five-second samples after two-second warmups. No tests, builds or agent work ran during timing.

| Encoding/route | Normal req/s | PGO req/s | CPU µs/request, normal → PGO |
|---|---:|---:|---:|
| Identity room | 14,535 | 14,650 | 145.0 → 141.3 |
| Identity active room | 20,062 | 19,827 | 130.4 → 126.9 |
| Identity history | 20,856 | 21,639 | 125.7 → 121.8 |
| Identity search | 15,410 | 15,834 | 211.1 → 201.8 |
| Identity static CSS | 102,927 | 102,998 | 17.8 → 17.7 |
| Gzip room | 22,191 | 20,400 | 130.5 → 132.5 |
| Gzip active room | 28,425 | 28,211 | 121.7 → 121.0 |
| Gzip history | 31,020 | 31,670 | 118.0 → 115.9 |
| Gzip search | 16,066 | 15,952 | 204.6 → 202.0 |
| Gzip static CSS | 102,790 | 102,886 | 17.7 → 17.6 |

Whole-process post-five-route median Pss, MiB: identity 111.9 → 112.4; gzip 123.8 → 124.1.
All 12 evaluation runs/60 HTTP samples had zero errors; all 1,662 active posts persisted
and matched FTS. Complete gzip/identity checks passed without cursor masking. Linux vet
and race tests with the profile passed. Normal-source native/Linux gates are retained in
the [ownership report](../ownership-20261005/README.md).

PGO lowered CPU for several routes but did not produce a broad, consistent throughput
improvement. Gzip room throughput was lower in this sequence, and the profile cannot
establish benefits for unprofiled production workloads. The evidence does not justify
making this synthetic profile a default build input. Operators with actual production
profiles can use Go's existing explicit `-pgo` option; no project wrapper is necessary.

## Why this optimization pass stops here

The remaining measured costs are mainly fresh SQLite/CGo row acquisition, authorization
observations, socket output, GC and access-log file writes. Reasonable repeated-work and
ownership candidates have been implemented or evaluated:

- Prepared gzip, immutable Part identities, static room shells, binary version keys,
  direct record input and request-owned preparation are retained.
- The final [single-owner list cleanup](../ownership-20261005/README.md) removes duplicate
  retained payloads without claiming a whole-process memory or broad throughput reduction.
- SQLite JSON batching, a second canonical timestamp parser and this PGO default were not
  justified by their application results or maintenance/representativeness costs.

Further snapshot/generation caches would need complete mutation, external-write and
permission invalidation, rather than assuming that a room GET can skip fresh observations.
Replacing cancellable SQLite stepping, combining session checks, suppressing logging or
making logging lossy/asynchronous would alter retained contracts. Manual room-key encoding
would replace a shared rendering/key input with another dependency list for a small remaining
JSON cost (about 3% cumulative CPU). None is justified by the current evidence.

This is diminishing returns for the evaluated workload and reasonable changes, not an
absolute performance ceiling. VM noise limits small comparisons; physical desktop access
was unavailable. No native AMD, Go/Rust, full-browser, Cable/media/TLS or external-integration
throughput claim follows. The pre-existing sidebar layout gap is unchanged.

## Reproduce

[harness.json](harness.json) records exact source/binary/profile hashes and flags. Training
profiles/logs are in `training-0` and `training-1`; evaluation samples are
[identity](final-0/raw.json) and [gzip](final-1/raw.json). The merged input is
[combined.pprof](combined.pprof).

```sh
go tool pprof -proto training-0/owned-1.pprof training-1/owned-1.pprof > combined.pprof
CGO_ENABLED=1 go build -trimpath -buildvcs=false -tags sqlite_fts5 \
  -pgo=/absolute/path/to/combined.pprof -o campfire-pgo ./cmd/campfire
```

Compare normal and PGO binaries using the ownership report's application command with
`--apps owned pgo`, adding `static_css` to its routes and separate identity/gzip outputs.
The profile applies only when explicitly requested; ordinary builds remain unchanged.
