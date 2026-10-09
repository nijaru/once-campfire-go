# Constant-space encoding negotiation

The private selector previously allocated lists, wildcard expansions, a rejection
map and a stable sort, then ignored entries after the first sixteen. A wildcard
followed by fifteen unknown codings and `gzip;q=0` produced gzip despite its
explicit exclusion. The regression fails on `99d1cf8` and passes on this candidate.
[RFC 9110 §12.5.3](https://www.rfc-editor.org/rfc/rfc9110.html#section-12.5.3)
excludes explicitly named codings from a wildcard and defines qvalue zero as not
acceptable. Current Rust inherits the sixteen-entry Rack limit. Scanning the
complete field is a deliberate Go correction, documented in the root README.

`internal/front/negotiation.go` now owns both selectors. Private selection retains
only gzip, identity and the first wildcard, with existing duplicate rejection,
quality parsing, preference ordering and implicit-identity rules. Public selection
retains the first gzip/zstd declaration and its separate quality, zstd tie and HEAD
policy. Both use constant state instead of allocating input-sized token lists.
Unknown private codings need no quality parsing; no decoded request is retained.
Streaming thresholds, jitter, validators, coding bytes, admission and authority
are unchanged. This is not completion of aggregate output-memory/resource work.

## Verification

Native front/web races repeat three times. Full native `bin/check`, fresh native
and Linux builds, Linux vet/front/web races, shared browser flows and actual
public HTTP/write checks pass. The public probe primes an identity room response,
then confirms identical uncoded bytes for a late gzip exclusion and 406 when both
supported representations are excluded. It also runs twenty shared responses,
seven preflights, four exact acknowledged ID/body/room/FTS audits and the existing
forwarding-disabled authority probe. No shared assertions are weakened; the
browser retains only its disclosed persisted-echo wait before exact-three.

A temporary deterministic probe compares 20,000 short fields against the original
private/public selectors, including wildcard, duplicate, quality and tie cases.
All match. Its common `gzip, deflate, br` benchmem input records five original
repetitions at 225–235 ns/op, 568 B/op and nine allocations; the replacement uses
46–74 ns/op, zero bytes and zero allocations. The slower candidate sample remains.
This is isolated evidence, not a complete-application allocation claim. The probe
is archived as `probe_test.go.txt`; only the demonstrated late-exclusion HTTP
regression is added to the maintained suite.

## Complete-application control

Two builds alternate sequentially across three rounds with eight timed seconds
and two excluded warmup seconds, sixteen clients and one viewer. Four runtime
workers/readers, three workers per job queue; server CPUs 0–3/client CPUs 4–7.
Completed/fragment budgets are 64/32 MiB plus inherited gzip/public 32/64 MiB.
No builds, tests, profiles or agents overlap timing. The independent shared
complete-response and exact SQLite/FTS write oracles are unchanged.

| Route | Previous median req/s | Candidate median req/s | Change |
|---|---:|---:|---:|
| Room | 32,309 | 31,693 | −1.9% |
| Messages | 33,498 | 32,844 | −2.0% |
| Sidebar | 36,493 | 36,984 | +1.3% |
| Search | 35,123 | 36,646 | +4.3% |
| Post | 4,605 | 4,665 | +1.3% |

Room/messages improve first and regress the remaining pairs. Sidebar/search
improve every pair; posting regresses second and improves first/third. All samples
remain. The mixed direction does not establish a general capacity gain; the
contract fix and bounded allocation improvement stand independently of throughput.
No fresh Rust comparison is claimed. Mixed invalidation is not repeated because
observation/database work is unchanged; the suite includes complete uncached writes.

The control validates 6,781,120 timed responses and audits 276,894 exact writes
including warmup, with zero errors/invalid responses and all FTS audits passing.
Peak generator CPU is 56.2% out of 400% available, not universal generator headroom.
The current shared source is `ec02deb`; its additions since `e244051` are source-size
reporting/tools/checks only. The unchanged load generator is intentionally reused,
with its original build revision and hash recorded in metadata.

A separate two-round five-second warm diagnostic records 53.35 CPU seconds over
31.56 wall seconds. Fresh observation uses 5.69 seconds, authentication SQL 5.61,
allocation 4.07 and cookie verification 2.80. Negotiation is no longer a material
node in the retained cumulative listing. These shares are not normalized
per-response reductions. Both profiles, raw results and gate logs are archived.

ARM64 OrbStack affinity is not physical isolation. Tmpfs verifies transactions
and FTS, not NVMe/crash durability. One viewer does not establish multi-user/fanout
capacity; throughput-run tails are not matched-offered-rate latency. Native Intel
is unavailable, forced CLI exit remains unexercised, and unchanged media is not
rerun. Resource/presentation work remains unfinished; optimality is not claimed.
