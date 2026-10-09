# Cookie verification allocation

Cookie verification rebuilt the base64 alphabet translator for each outer and
inner payload. This change shares its immutable lookup table across cookies,
signed identifiers and stream names, preserving standard, URL-safe and mixed
alphabets. Verification also writes the expected lowercase SHA1 signature into
fixed buffers rather than allocating temporary hex strings. Constant-time HMAC
comparison, Rails envelopes, expiry/purpose checks and legacy JSON remain unchanged.
No identity, authorization decision, decoded cookie or HMAC state is cached.

## Source reconciliation and checks

The baseline is `94c2118`, the non-rewriting merge of upstream `c4ab53d`. Rust is
`6dae2fd` and shared verification `ec02deb`. The Go/Rust advances change README
source-size reporting only. Shared verification adds a template-code extractor,
its tests, a check entry and reporting documentation; application contracts,
browser, load generator and benchmark inputs are unchanged. Full diffs were
inspected and executable-input equivalence asserted. The existing Rust binary
built at `9872c1d` and load generator built at `e244051` are intentionally reused;
metadata distinguishes their build revisions from current source revisions.
VM copies include the new shared files, preserving only the previous temporal
baseline adapter and diagnostic graceful-stop adapter.

The full updated shared check passes natively, including the extractor tests.
The first attempt used a toolchain image without Ruby; the successful native
suite supersedes that environment failure, not a failed product assertion.
Five repeated Rails vector races, full native `bin/check`, fresh native/Linux
builds, Linux vet and Rails/web races, browser flows and actual public checks pass.
Public checks verify twenty responses, seven preflights, four exact acknowledged
ID/body/room/FTS records and forwarding-disabled authority. The browser retains
only its disclosed persisted-echo wait before the unchanged exact-three assertion.
Existing protocol vectors cover signing, encryption, expiry, purpose, tampering,
legacy payloads and identifier/stream formats; no extra permanent test is added
for moving an immutable translator or using equivalent signature buffers.

## Isolated evidence

A temporary Go 1.27.1 Apple M3 Max probe verifies one Rails-generated session cookie.
Five before repetitions use 2,096–2,773 ns/op, 2,644 B/op and 28 allocations; five
after repetitions use 1,684–1,733 ns/op, 1,778 B/op and 20 allocations. The slow
before sample remains in the record. This is fixed-input verification evidence,
not complete-application allocation or capacity evidence. The probe is archived
as `probe_test.go.txt`, not added to the maintained suite.

## Complete application

Three sequential interleaved rounds use eight timed seconds plus two excluded
warmup seconds, sixteen clients and one viewer. Four runtime workers/readers,
three workers per job queue, server CPUs 0–3/client CPUs 4–7. Completed/fragment
budgets are 64/32 MiB plus inherited gzip/public 32/64 MiB. Forward/reverse order
keeps the candidate in the middle. No builds, tests, profiles or agents overlap
capacity timing. Responses and exact acknowledged writes use shared contracts.

| Route | Previous median req/s | Candidate median req/s | Measured Rust median req/s |
|---|---:|---:|---:|
| Room | 21,827 | 27,588 | 40,722 |
| Messages | 26,420 | 26,020 | 36,605 |
| Sidebar | 22,676 | 27,855 | 43,518 |
| Search | 21,917 | 30,916 | 43,829 |
| Post | 3,982 | 4,445 | 5,826 |

Room/search improve every pair. Messages improve only the second pair; sidebar
and posting regress first and improve the remaining pairs. The previous build
slows markedly in rounds two/three: room falls from 28,831 to 21,827/20,149;
sidebar from 33,201 to 15,190/22,676; search from 31,765 to 16,207/21,917. Candidate
messages also slow to 19,679 in round two. All samples remain. The resulting
+26%/+23%/+41% room/sidebar/search median ratios are **not isolated causal gains**
from the small cookie change. The message median is 1.5% lower. No general
capacity or Rust-parity claim is supported by this noisy comparison.

The cohort validates 9,217,968 timed responses and audits 418,392 exact writes
including warmup. Errors/invalid responses are zero; all SQLite/FTS audits pass.
Peak generator CPU is 50.7% out of 400% available, not proof that all generators
are unconstrained. Mixed invalidation runs are not repeated: freshness, database
work and coding are unchanged; the suite includes complete uncached writes.

A separate two-round five-second warm diagnostic records 49.05 CPU seconds over
31.52 wall seconds. Cookie verification uses 2.56 seconds (5.22%), including
1.91 seconds of envelope unpacking. Fresh observation and allocation remain
material. Percentages from separate throughput populations are not normalized
per-response reductions. Both profiles and every raw sample are retained.

These ARM64 OrbStack/macOS runs use guest affinity, not physical isolation.
Tmpfs verifies transactions/FTS, not NVMe or crash durability. One viewer does
not establish multi-user/fanout capacity; throughput tails are not matched-load
latency. Native Intel is unavailable, forced CLI exit remains unexercised and
unchanged media checks are not rerun. Resource design and broader application
boundaries remain unfinished; global optimality is not claimed.
