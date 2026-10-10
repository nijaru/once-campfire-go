# PR readiness comparison: 32a9f03

The branch is **not PR-ready** on this evidence. Warm reads improve against current
Go upstream, but message POST throughput regresses consistently. Matched-rate
latency and distinct-user Cable fanout remain unmeasured in this population.

## Sources and method

Candidate: `32a9f0384489783bb9d1a75b375c394f574544b6`. Go control:
`c4ab53d00b0a19a55962e03d18057ca6c3da9437` (`upstream/main`). Both are fresh
Go 1.27.1 linux/arm64 CGO/FTS5 trimpath builds. The control's asset generator uses
its own overrides and the same pinned Rust/Rails reference inputs; its missing
worktree submodule was not substituted with current upstream assets.

Rust source is `6dae2fd3c1a2dce61ca11beb731bcce2b42b67ad`; its existing production
binary was built at `9872c1d9541551c0889a60e3d49a14488bd1e403`. Later Rust changes
only affect source-size reporting, not application/build inputs. Shared verification
is `ec02debb3e44f40b8778a6a42db3c4515b2f9eaf`; executable loadgen inputs are
unchanged from its build revision `e2440515d9093b8304da3d7662d8526f3234b047`.

The shared harness's response/write/FTS assertions are unchanged. A local adapter
adds the `go-before` application name and maps it to the Go runtime and exact
control source. The adapter patch is retained. A first preflight failed because
Ruby ran in US-ASCII; UTF-8 preflights then passed for all three applications. This
was an environment failure, not an application-contract failure.

Warm and mixed populations each use three sequential forward/reverse rounds,
eight timed seconds plus two excluded warmup seconds, sixteen read clients, four
runtime workers/readers, and three workers per job queue. Server affinity is 0–3;
client affinity is 4–7. Production cache defaults are retained. No builds, tests,
profiling or review agents ran during either timing population. The three-build
order keeps the candidate in the middle; it does not balance every order effect.
The seed is copied to VM-host tmpfs; push/webhook destinations are rewritten to
loopback port 9 before starting an application.

## Results

| Warm route | Go upstream req/s | Candidate req/s | Change | Rust req/s |
|---|---:|---:|---:|---:|
| Room | 28,207 | 33,012 | +17.0% | 43,966 |
| Messages | 28,473 | 34,212 | +20.2% | 42,740 |
| Sidebar | 32,406 | 38,865 | +19.9% | 48,289 |
| Search | 32,424 | 37,299 | +15.0% | 47,523 |
| POST | 5,653 | 4,653 | **−17.7%** | 5,939 |

Candidate POST samples are 4,646–4,669 req/s; control samples are 5,646–5,671.
The regression is not explained by one outlier. All individual samples remain in
the raw evidence. These measurements cover the entire branch against upstream,
not the causal gain or cost of the latest boost migration.

| Reads with ten paced writes/sec | Go upstream req/s | Candidate req/s | Change | Rust req/s |
|---|---:|---:|---:|---:|
| Room | 35,578 | 39,978 | +12.4% | 41,472 |
| Messages | 35,748 | 41,920 | +17.3% | 40,579 |
| Sidebar | 43,908 | 48,477 | +10.4% | 47,250 |
| Search | 43,649 | 47,033 | +7.8% | 46,541 |

Mixed writes target HQ while room/message reads target Watercooler. This exercises
global cache invalidation, not same-room fanout. Do not compare absolute warm and
mixed rates as a causal effect of writes: they are separate populations on a
shared host with different background activity.

Warm runs validate 11,160,642 timed responses and audit 483,167 writes including
warmup. Mixed runs validate 12,282,887 timed reads, 2,880 timed writes and 3,600
writes including warmup. Every response error/invalid count is zero, and all exact
acknowledged-write, body, room, FTS and integrity audits pass. Peak generator CPU
is 56.1% and 68.5%, out of 400% available; this is not a universal generator limit.

## Diagnostic and verification

A separate two-round, five-second POST diagnostic profiles both Go builds, using
graceful shutdown to flush profiles. Its rates are not capacity evidence. All four
profiles were recovered from the actual runtime PID directory and copied before
another diagnostic could reuse it. Notification preparation, message display
assembly and gzip completion are material candidate costs; cumulative CPU shares
are not normalized per-response reductions or an isolated cause of the regression.
The diagnostic validates 105,855 timed responses and audits 147,668 writes.

At candidate source, native `bin/check`, Linux vet and affected database/application/
web/runtime races pass. The shared Chromium workflows and real public HTTP/write
contracts pass. The bot-author regression fails behaviorally on `b301e13` and
passes on the candidate: a profile commit after authentication no longer leaves
bot boost JSON with the old name, role or avatar version. Static independent review
identified that gap and an obsolete permission mode; both were resolved. Static
review is not independent test evidence.

## Limits

ARM64 OrbStack on macOS; guest affinity is not physical core isolation. Native
system activity is not eliminated. Tmpfs audits do not measure NVMe or crash-safe
durability. One read viewer is not multi-user capacity. Throughput-run latency tails
are not matched-offered-rate latency. Intel and forced CLI exit remain untested;
unchanged media was not rerun. There is no Rust-parity, global-optimality or zero-bug
claim, and the POST regression must be investigated before presenting the branch
as a high-performance PR.
