# Message write preparation

The combined candidate improves median message POST throughput by **7.6%** against
`32a9f03` in a matched two-build comparison. A fresh upstream comparison still
shows a **10.3% deficit**. These improvements reduce the earlier regression; they
do not establish PR readiness, Rust parity or an optimal implementation.

## Changes and protection

Submitted canonical bodies are parsed before acquiring the sole writer. Plaintext
without mentions is also prepared there. Mention names and blank-body attachment
fallback remain transaction-current; omitted update bodies are still selected and
derived inside the mutation. Search preparation has its own small module rather
than extending the command module's SQL/effect responsibilities.

Creation now selects current room, author and membership permission in one writer
query instead of three. Trusted webhook replies retain their explicit no-membership
check. The overwritten room timestamp is no longer loaded just to discard it.

Provider payloads use Go 1.27's standard JSON v2 encoder directly, removing the
marshal/parse/re-encode round trip. Push retains non-HTML escaping; webhook payloads
retain HTML escaping. Both retain literal U+2028/U+2029, escaped backslashes,
invalid-UTF-8 replacement, struct field order, nullable HTML and full int64 values.
Cookie and verifier encoding is unchanged. The installed Go 1.27.1 API documentation
confirms these options; no experimental flag or dependency was added.

A focused concurrent regression queues creation behind a profile/room edit while
all readers are held. Its committed FTS mention and receipt use the writer's new
names, not earlier preparation. Existing omitted-body, attachment and rollback
checks remain. Exact provider-payload assertions cover escaping, nil/empty HTML and
integer precision; these assertions also passed against the former encoder before
its replacement. Existing transport tests did not protect payload construction.

Native `bin/check`, Linux vet and affected database/application/web/integration
races pass. Current shared Chromium workflows and public HTTP/write contracts pass,
including the four exact acknowledged-write/FTS audits. The shared browser's local
persisted-echo wait remains disclosed; its exact-three assertion is unchanged.

## Complete POST comparisons

Each row is a separate population: three sequential forward/reverse rounds,
eight timed seconds plus two excluded warmup seconds, sixteen clients, one viewer,
four runtime workers/readers and three workers per job queue. Server affinity is
0–3, client affinity 4–7; gzip and production cache defaults are retained. The
three-build order keeps the candidate in the middle. No builds, tests, profilers
or agents overlapped timing.

| Candidate scope | Branch control req/s | Candidate req/s | Median change |
|---|---:|---:|---:|
| Search preparation only | 4,687.2 | 4,806.9 | +2.6% |
| Plus writer selection | 4,662.4 | 4,933.9 | +5.8% |
| Plus direct provider encoding | 4,741.4 | 5,103.7 | +7.6% |

Search-only is slower in one pair; both later candidates improve all three pairs.
The rows do not isolate each added component's causal gain: their controls and
host conditions differ. Every individual sample, including the slower pair, is
retained. Candidate/control sources, patches and binary hashes are in metadata.

The final matched upstream population measures **5,697.6 / 5,113.2 / 6,014.3 req/s**
for upstream Go / candidate / Rust. Candidate samples are 5,098.9–5,164.8; upstream
samples are 5,691.4–5,724.5. The remaining deficit is consistent, not one outlier.
This population compares the entire branch against upstream, not just this cohort.

Across the four populations, 1,098,305 timed responses pass and 1,366,545 writes,
including warmup, pass exact ID/body/room/FTS and integrity audits. Every error and
invalid-response count is zero. Peak generator CPU is 34.6% of a 400% budget.

## Provenance and limits

Measured candidate source is `d202a7f` plus the retained final patch; the same code
is committed with this evidence. Branch control is `32a9f03`, upstream control
`c4ab53d`, current Rust source `6dae2fd`. All three upstream heads were refreshed
and unchanged before measurement. Rust's unchanged binary build remains `9872c1d`;
shared source is `ec02deb`, with unchanged loadgen built at `e244051`.

ARM64 OrbStack/macOS guest affinity is not physical isolation. VM-host tmpfs does
not establish NVMe throughput or crash durability. Push/webhook URLs are rewritten
to loopback port 9; Go push delivery then exits at endpoint validation. This measures
payload preparation and admission, not provider encryption/network throughput.
One viewer is not distinct-user capacity; throughput-run tails are not matched-rate
latency. Read routes and mixed churn are not remeasured here because their query and
invalidation paths are unchanged. The previous read evidence remains at its actual
revision. Matched-rate latency, distinct-user Cable fanout and further POST
contention diagnosis remain unfinished.
