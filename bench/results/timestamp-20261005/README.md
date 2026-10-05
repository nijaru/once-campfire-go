# Rejected canonical timestamp fast path

A strict parser for `Stamp`'s fixed UTC `YYYY-MM-DD HH:MM:SS.ffffff` format was tried on
production `fd6517b`. It checked ASCII digits, delimiters, field ranges and month-day
normalization; other formats and malformed input retained the existing `time.Parse` path.
The candidate is **not shipped**.

A native scalar probe measured approximately 41 ns for canonical scanning versus 106 ns
for direct stdlib parsing, both without allocations. That probe parses one timestamp only;
it does not represent application throughput, SQLite row acquisition or cancellation.
Differential fuzzing against the existing stdlib layouts passed 1,456,755 executions.
Native/Linux vet/race gates also passed for the trial.

## Application results

Same Linux/arm64 OrbStack VM and four-route methodology as the
[ownership comparison](../ownership-20261005/README.md): three rotating repetitions,
public HTTP/1.1, fresh seed, 16 clients, five-second samples, two-second warmups, logging
and active commits enabled. No test/build/agent work ran during timing.

| Encoding/route | Owned baseline req/s | Trial req/s | CPU µs/request, baseline → trial |
|---|---:|---:|---:|
| Identity room | 15,088 | 14,869 | 143.6 → 144.2 |
| Identity active room | 19,711 | 20,107 | 131.9 → 125.4 |
| Identity history | 21,818 | 21,721 | 122.8 → 122.7 |
| Identity search | 15,494 | 15,743 | 211.1 → 207.9 |
| Gzip room | 21,233 | 21,678 | 128.9 → 136.1 |
| Gzip active room | 28,534 | 29,041 | 121.1 → 117.8 |
| Gzip history | 31,458 | 30,577 | 117.5 → 119.0 |
| Gzip search | 16,200 | 15,719 | 205.3 → 208.0 |

Whole-process post-four-route median Pss, MiB: identity 111.4 → 111.9; gzip 122.0 → 120.2.
All 12 runs/48 HTTP samples had zero errors; all 1,660 active posts persisted and matched FTS.
Complete response validation passed without cursor masking.

The scalar gain did not establish a consistent application benefit. Active rooms showed
small gains, while room/history/search results varied by encoding, with overlapping sample
ranges. Timestamp scanning was only 3.2% cumulative CPU in the preceding room profile.
Maintaining a second calendar/parser validation path is not justified by this evidence.
Production therefore keeps the stdlib parser and its full existing input/error contract.

## Evidence and reproduction

[harness.json](harness.json) pins the baseline, trial binary, source staging and environment.
The [source patch](candidate.patch), [differential fuzz test](validation_test.go.txt),
[scalar probe source](timestamp_probe_test.go.txt), compressed fuzz/probe logs, check logs,
[identity samples](final-0/raw.json) and [gzip samples](final-1/raw.json) are retained.
Validation/probe files are artifacts, not additions to the permanent test suite.

To reproduce the candidate, apply the zero-context patch to `fd6517b` with
`git apply --unidiff-zero candidate.patch`, install the validation artifact as
`internal/database/timestamp_test.go`, and build with the manifest's pinned Go/toolchain
flags and generated assets. Run the ownership report's application command with named
`owned`/`timestamp` candidates, `--apps owned timestamp`, and separate identity/gzip outputs.
The source patch is intentionally absent from production.
