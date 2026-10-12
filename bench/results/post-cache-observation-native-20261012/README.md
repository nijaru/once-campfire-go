# Write cache-observation guard — native rejection, 2026-10-12

**The guard landed at `513485a` is reverted after failing its native
non-regression controls.** Its [local CPU/write reduction](../post-cache-observation-20261012/README.md)
is real evidence for that local workload, not a justification for keeping the
change despite these native results. No correctness defect or missing delivery is
observed, and no host-noise cause is asserted.

## Identities and bounded decision

The Fedora i9 desktop was rebooted and available for this run. Guard application
source `513485a8921606266dc812d83c310cd46f42e076` differs from baseline production
`8351224b56c9a3c6784ed5d7510857212133ac4c` only by the method guard. The new image
`9a646a3c1f7df42d436d0c6bd69649a178f6a37933ba4bd56978d88d02d93b6c`
reuses the verified production image, replacing that one file before a serialized
Go 1.27.1 CGO/FTS5 build. Both source generations and upstream use identical
Go build settings, four runtime workers/readers, three jobs and default GC.

| Binary | SHA-256 |
|---|---|
| Guard `/go-guard` | `77b66b029d9e3a0f6211e3afaab9e2ad25d3d6b8e710e97e00ec3bc8722c7e2e` |
| Baseline `/go-candidate` | `8a1409c09aa79e188010fb04dfa0c63b66e4f2748de734a117f6ae5db49b0e29` |
| Upstream `/go-upstream`, `c4ab53d` | `70cca09f7fa45eac8e383a2160990ac1ad9a1bdd9c6b95f47dc666b60b987b54` |

Seed SHA is `691c52bfdc30b38bd8867e93b0914613df20e03d4cd7b00796ab68e0a5d26c9a`;
shared verification is `ec02debb3e44f40b8778a6a42db3c4515b2f9eaf`. Server CPUs
are `0,2,4,6`, clients `8,10,12,14`, runner `16,18`; affinity is not exclusive host,
physical-core or thermal isolation. Fixture push/webhook endpoints point at closed
localhost ports. No task-owned build, test, agent or profiler overlaps the
server-uninstrumented controls. The client still performs common ACK recording;
its overhead neutrality is not established.

The declared budget is three interleaved fixed-rate POST pairs, one POST-capacity
pair and one 64-independently-authenticated-person Cable pair. The non-regression
limits are no greater than 5% median fixed POST p99 increase, capacity/Cable
throughput loss, or paced all-client Cable p99 increase; all writes and delivery
must pass. There is no repeat-to-win extension.

## Complete, uninstrumented controls

Fixed POST uses gzip, 3,000 offered writes/s, 64 workers, two warmup plus eight
timed seconds, complete decoded Turbo responses and no dropped/expired/unscheduled
arrival. Orders are baseline/guard, guard/baseline, baseline/guard.

| Pair | Baseline p99 (ms) | Guard p99 (ms) | Guard change |
|---|---:|---:|---:|
| 1 | 0.892 | 0.947 | +6.2% |
| 2 | 0.978 | 0.959 | −1.9% |
| 3 | 0.899 | 0.991 | +10.2% |
| Median | **0.899** | **0.959** | **+6.7% — fails** |

Every second arm has higher p99 in this small fixed cohort. That limits causal
interpretation, but does not turn the failed predeclared gate into a pass or
justify automatically attributing it to system noise. There is no additional
sweep. All **180,000** warm/timed writes pass exact persistence audits.

POST capacity uses 16 clients and two warmup plus eight timed seconds:

| One pair | Baseline | Guard | Guard change |
|---|---:|---:|---:|
| Delivered POST/s | 9792.1 | 9865.2 | +0.7% |
| POST p99 (ms) | 6.235 | 6.159 | −1.2% |

All **196,026** warm/timed writes pass. This one pair is a non-regression control,
not enough for a general capacity-gain claim.

Cable uses 64 distinct authenticated people, six subscriptions each, 300 paced
messages at 100 ms intervals, then five saturated seconds/four posters. The
corrected pair runs guard first, baseline second.

| One corrected pair | Baseline | Guard | Guard change |
|---|---:|---:|---:|
| Complete delivered messages/s | 3083.3 | 3122.9 | +1.3% |
| Paced all-client p99 (ms) | 6.591 | 6.971 | **+5.8% — fails** |
| Saturated POST p99 (ms) | 3.371 | 3.245 | −3.7% |

All **31,640** messages pass exact ACK/body/creator/room/FTS checks and complete
all-client delivery. SQLite integrity is OK, the seed's one FK orphan is unchanged,
and presence returns to zero. One pair and 300 paced messages per arm are limited
tail evidence, not proof of a universal regression. They fail this experiment's
bounded gate nonetheless.

## Separate CPU attribution, not native capacity evidence

One production-hook CPU pair profiles guard/upstream at 3,000 offered POST/s,
64 workers, two warmup plus twenty timed seconds, excluding authentication/setup
through server restart. All **132,000** writes pass complete response and exact
persistence/FTS/integrity audits.

Guard samples 34.34 CPU seconds versus upstream 33.98, or approximately
520.3 versus 514.8 µs/completed write. Total CPU is effectively flat at this sample
size. Guard finite gzip preparation samples **10.41 s**; upstream gzip closure
samples **4.57 s**. Both decoded/wire response populations pass the same client,
and average wire bytes are 1957 in each timed phase. This attribution identifies
compression as a substantial owner worth investigating; it does not establish
that CPU difference as the cause of the native tail gap or justify changing
compression level/decoded content blindly. Functions can move between stacks
when control flow changes; percentages alone are not a work reduction.

These are guard/upstream profiles, not a baseline/guard native CPU comparison.
Profiled p99 values are instrumented and are not promoted as new current/upstream
ranking evidence. The previous [three-pair upstream snapshot](../current-upstream-native-20261011/README.md)
still measures source `8351224`, now restored by this rejection.

## Failures, counts and delivery

The first Cable adapter changed its setup executable but not its restart
executable. Its arm labeled guard actually timed baseline `8351224`, producing
**15,657 audited baseline messages**, not guard-control evidence. The baseline
restart then selected nonexistent `/go-baseline` and failed before benchmark
posting. Both failures and the mislabeled arm are retained separately; assertions
are unchanged. The corrected adapter records the SHA of each actual executable
and runs one fresh pair. Excluded writes are not counted as a successful guard
cohort.

Pprof extraction first failed at the SELinux mount boundary; a following command
used Bash assignment syntax in the remote Fish shell and failed before analysis.
Corrected commands read the existing profiles with label-disabled mounts. Neither
failure exercised the application or added workload samples.

Valid cohorts total **539,666 distinct audited writes**: 132,000 attribution,
180,000 fixed controls, 196,026 capacity and 31,640 Cable. The separately audited
mislabeled baseline population is excluded from that total.

The guard is removed, restoring the exact `8351224` production source. `bin/check`
passes on the reversion. Its previously passing cache/query race, four shared
Go/Rust HTTP flows and independent read-only review remain source-specific
correctness evidence for the rejected guard, not a performance promotion.
The manifest retains profiles, exact adapters, raw ACK populations, samples,
metadata, failure logs and gates; secrets, cookies, databases and runtime request
logs are excluded. No PR is opened or prepared. Further work must target a fresh,
demonstrated mechanism rather than repeat this closed control.
