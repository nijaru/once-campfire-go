# SQLite memory accounting: application CPU gate failed

No production change. Disabling SQLite memory accounting saves only 0.10% joined application CPU in one matched native fixed-POST pair, below the preregistered 5% gate. The flag remains off; no conditional fixed/capacity cohorts, refinement or integration follow.

The [functional investigation](../pr10-sqlite-review-20261014/README.md) verified a real cost: disabling accounting makes global memory counters and SQLite soft/hard heap limits nonoperational, even though the hard-limit setter still accepts and reports a limit. Per-connection cache accounting remains. Our app uses none of those global APIs; Express's pinned Node build disables accounting, while Rust keeps it. Neither source facts nor another port's choice established a benefit for this workload.

## One preregistered native pair

Exact application `8351224` with original WebSocket fork; no gather-write code. Only the candidate build added `CGO_CFLAGS="-O2 -g -DSQLITE_DEFAULT_MEMSTATUS=0"`. Baseline native binary SHA256 `8a1409c09aa79e188010fb04dfa0c63b66e4f2748de734a117f6ae5db49b0e29` matched byte-for-byte before posting; candidate `821faa8a7d3f83eca8df94674cb341f13b9009aca10be919f4bcb59c7df8e4b3`. Driver-linked probes confirmed SQLite 3.53.4 and the expected compile-option values 0/1. These probes alone do not certify every SQLite feature; existing affected database/application/web race tests and vet also passed.

Physical i9-13900KF/32 logical CPU desktop, Fedora Linux 7.2.10-200.fc44.x86_64, Go 1.27.1/CGO/FTS5, four workers/readers, three jobs and default GC. Server CPUs 0,2,4,6; client 8,10,12,14; runner 16,18. No exclusive hardware, thermal isolation or precise confidence/noise bound is claimed. Normal host work was not prohibited. No task-owned builds, tests, agents, profilers or process samplers overlapped timed posting.

Exactly current then candidate, each 3000 POST/s, 64 workers, gzip, queue 8192, two warmup plus twenty measured seconds. Maintained receipt client, seed, assets and runtime image matched. Authentication/preflight used a separate setup server. A joined `wait4` supervisor counted whole-server user+system CPU from restart through readiness (three-second wait), warmup, posting, audits and shutdown. There were no Go CPU/allocation profiles. Complete-wire latency precedes maintained validation; that validation's overhead is not assumed neutral.

| Result | Accounting enabled | Disabled | Change |
|---|---:|---:|---:|
| Joined process CPU, s | 34.513757 | 34.478288 | −0.1028% |
| CPU per 66,000 writes, µs | 522.936 | 522.398 | −0.1028% |
| Timed POST p99, ms | 0.938 | 0.935 | −0.3198% |

User CPU fell 28.354772→25.392676s while system CPU increased 6.158985→9.085612s. Total CPU is the gate; this redistribution is not a component-saving claim or a diagnosis. One pair with almost flat total CPU does not establish equivalence or a tiny real improvement.

All 132,000 workload writes passed exact receipt-to-persistence, canonical body/client identity, creator/room, FTS, integrity, unchanged seed FK and joined exit-0 checks. Existing shared HTTP route preflight passed both arms; its setup writes are separate from the workload count. The first test invocation lacked the required read-only Rust schema fixture; copying it into this task-owned source corrected setup, with the failed log retained and no skipped assertions. No host tool or network configuration changed.

## Decision

The finite CPU screen fails. There is no demonstrated material application gain to justify losing accounting/heap-limit capability, so retain existing build defaults. No deployment, flag adoption or fresh competitor ranking follows. This changes the next decision: do not import PR #11's flag on the strength of its allocator mechanism or Express's default. Reopening needs genuinely new representative workload evidence, not a repeat of this control.

`manifest.json` owns criteria, build/checks and compile-option probes, capture/supervisor/audit helpers, both raw receipt populations, joined CPU, full summaries and hashes. Private source archives, binaries, runtime logs, databases, credentials and full CPU rosters are excluded. Production remains exact `8351224`; the broader competitor-performance goal remains unfinished.
