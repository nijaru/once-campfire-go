# Local compression policy, attribution and receipt-audit fix — 2026-10-12

**Level 3 is rejected: 19.4% less compression CPU costs 9.2% more wire bytes.**
Production compression and application source remain unchanged at `8351224`.
The useful code change at `4a35851` fixes the maintained benchmark client and
harness: malformed POST receipts no longer count as successes, and fixed POSTs
receive exact acknowledgement/persistence audits.

## One algorithm candidate, closed locally

The installed Go 1.27.1 source gives a concrete algorithm distinction: level 6
uses two match tables with full indexing and repeat checks; level 3 uses one
medium table with two candidates per entry. This is not a blind level sweep.
The declared preliminary gate is at least **25% lower compression CPU**, at most
**3% wire growth**, and identical decoded bytes. Application timing is allowed
only if this preliminary gate passes.

The [matched receipt corpus](../compression-receipts-20261012/README.md) supplies
16 unmodified bodies. Another 16 bodies replace the short deterministic client
IDs with the fixed-rate client's 26-character base32 namespace shape and five-digit
sequences. These are derived compression inputs, not new persisted messages or
exact replicas of the previous native traffic population. All 32 inputs roundtrip
at both levels. Three interleaved pairs use 10,000 bodies/arm, one compression
goroutine, pooled writer reset/close semantics, default GC and Go 1.27.1
**darwin/arm64**. CPU is OS process `getrusage` within each phase, not a sampled
profile percentage.

| Compression CPU/body | Level 6 | Level 3 |
|---|---:|---:|
| Pair 1 | 32.0955 µs | 25.8615 µs |
| Pair 2 | 32.0791 µs | 25.8290 µs |
| Pair 3 | 32.0157 µs | 25.9795 µs |
| Median | **32.0791 µs** | **25.8615 µs** |

Aggregate wire growth is **9.1676%**, maximum individual growth **9.4517%**.
Both benefit/size gates fail. There is no application prototype, additional level
comparison, or native latency/capacity claim. No compression-cache, pool budget,
queue, output or authorization change is made.

## Local sampled-owner calibration

Existing local profiles sample very little gzip work despite complete gzip
responses. One current-source arrival perturbation pair compares 3,000 and
3,001 offered requests/s, 64 workers, two warmup plus eight timed seconds, equal
client namespaces per phase, four server workers/readers and three jobs.
Setup/login/shared preflight precede the profiled restart. All **60,010 writes**
pass exact ACK-ID/body/creator/room/FTS, unique client-ID, integrity and unchanged
seed-FK audits. This is diagnosis, not an uninstrumented ranking.

The discriminator is more than 20% sampled gzip CPU/write variation while joined
OS process CPU/write stays within 5%. It is not met: each capture samples **0.01 s**
in gzip preparation. Process CPU/write is **429.6665 versus 429.3178 µs** (−0.08%).
This small arrival perturbation does not support the aliasing hypothesis; it also
does not prove sampled-owner attribution reliable. No extra rate sweep follows.

A separate private build adds atomic gzip call/elapsed counters and one shutdown
report, changing no application behavior. One 3,000/s, two-plus-eight-second
capture passes the same audits for **30,000 writes**. The counter observes exactly
**30,000 finite gzip calls**, **1.312057 s summed elapsed work**, or **43.7352 µs/call**;
the CPU profile contains no gzip samples. Joined process CPU is 13.422911 s and
sampled CPU 11.61 s. **Elapsed work is not CPU:** preemption, scheduling, GC and
pool/buffer work can contribute. The private hook may also perturb scheduling.
It is not production scaffolding or a timing control.

One standalone level-6 calibration performs 60,000 compressions on the same
captured/derived corpus: OS CPU 1.924319 s, wall 1.909490 s, sampled CPU 1.73 s,
including 1.63 s under the compressor and 1.60 s under gzip closure. Together these
observations justify using explicit work counts and process-CPU controls rather
than treating a negligible local sampled owner as absent work. They do **not**
identify a Go profiler bug or establish the cause of the native owner/tail gap.
Linux/amd64 profiling still needs its own matched absolute-work evidence.

The algorithm and standalone timed loops total **120,000 compression operations**,
not database writes. Diagnostic captures total **90,010 distinct audited writes**.
Seed SHA is `691c52bfdc30b38bd8867e93b0914613df20e03d4cd7b00796ab68e0a5d26c9a`;
shared verification is `ec02debb3e44f40b8778a6a42db3c4515b2f9eaf`.

## Completed benchmark-correctness fix

The maintained `bench/httprate` previously drained raw bodies and counted every
nonempty HTTP 200 as successful. The new regression fails on the original client
for incomplete Turbo output, corrupt gzip, wrong sequence, wrong client identity
and wrong content type. Valid identity/gzip receipts pass. Existing coverage did
not test receipt contents or gzip integrity.

POST mode now decodes and checks receipts before success accounting, emits exact
`[sequence, message_id]` acknowledgements and a client-ID prefix, and reports
invalid receipts separately. GET mode retains its constant-memory body drain.
Wire-completion timestamps remain unchanged; validation cost affects generator
work and is not assumed neutral. POST acknowledgement retention scales with the
completed population, as in the previously used private audit clients.

`bench/application` keeps receipt files before auditing and matches fixed POSTs
to exact SQLite bodies, client IDs, creators, rooms and FTS entries—not just row
counts. Independent review found that the prototype derived sequences from stored
text and could accept a consistently zero-padded body/client ID (`00`) for ACK
sequence `0`. Its focused regression fails before correction. The final audit
derives canonical expected body/client values from the acknowledgement instead.
Receipt completion order is immaterial; duplicate/misassigned IDs or matching
counts with corrupted persisted data fail. Resume identity includes the
new audit helper. Rebuild `bench/httprate` for this protocol; no compatibility shim
accepts unaudited old-client samples. Historical private-client cohorts retain
their recorded identities and are not relabeled as new-client runs.

Verification:

- Original-source receipt regression failure retained; focused Go tests and real
  SQLite acknowledgement tests pass. `bin/check` adds the Python tests and client
  formatting to its existing vet/race/assets/fork gates; the full check passes.
- Real CLI probes against Go `8351224` and cached Rust run identity/gzip, eight
  posts per encoding/app: **32 receipt/persistence-audited writes**, before the
  canonical-sequence review correction. Cached Rust source differs from pinned
  `6dae2fd` only in README, as recorded in the local guard evidence.
- The actual `bench/application` Linux/arm64 entry point runs Go/current and Rust,
  one fixed 50/s trial per app. The first integration, before the review correction,
  audits 100 timed writes and **11,563** total shared warmup/timed writes.
- After the canonical fix, fresh `bin/check` and the real entry point pass again:
  **100 timed writes** receive the final exact audit; unchanged shared capacity
  warmups plus those writes total **11,734**, passing existing aggregate
  persistence/FTS checks. Both runs are retained. Short functional timings here
  are not application performance evidence.

Two initial Rust probe attempts read SQLite from the macOS side of the VM mount
while the writer ran in Linux; they saw no new rows despite validated HTTP ACKs.
Their **16 HTTP-only receipts** are retained and excluded from the exact-audited
cohort. The corrected adapter performs live SQLite reads/audits in the writer's
Linux filesystem; both encodings pass without relaxing assertions. Already valid
Go probes are reused. This is a harness/filesystem-boundary failure, not a claimed
Rust persistence defect. The real Linux application-harness run independently
passes on its native shared filesystem.

## Evidence and next decision

The manifest retains 75 verified gzip artifacts (747,817 bytes): algorithms, raw
samples, profiles, the private hook patch, exact adapters/client sources, ACK
populations, failed probes and gates. Runtime
secrets, cookies, databases and application request logs are excluded.
The level-3 and arrival-perturbation premises are closed. Use the maintained
receipt-audited client for subsequent comparisons with matched client hashes.
Before a different compression policy, obtain native absolute-work evidence; do
not revive this level control or change behavior merely to reduce a profile share.
No native gain, global completion, or PR preparation/publication is claimed.
