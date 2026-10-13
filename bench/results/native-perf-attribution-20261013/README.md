# Independent native CPU attribution — 2026-10-13

Linux `perf` gives a current-fork capture consistent with joined process CPU and
the counted gzip-Close elapsed bound. The upstream capture loses 30 samples and
**fails the declared zero-loss comparative gate**. No comparative component-CPU
claim, POST-tail explanation or application performance gain follows this pair.

| Work, 66,000 audited writes/arm | Fork | Upstream Go |
|---|---:|---:|
| Joined process CPU, s | 32.752 | 33.917 |
| Sum of sampled CPU-clock periods, s | 32.956 | 34.095 |
| Lost samples | 0 | 30 |
| Close-chain sampled CPU, s | 5.411 | 3.310 |
| Counted Close elapsed sum, s | 5.509 | 5.250 |

Sample periods are nanoseconds from `PERF_COUNT_SW_CPU_CLOCK`, not cycles converted
to seconds. Whole-process estimates are within 1% of joined CPU. The fork's Close
estimate does not exceed those calls' summed elapsed execution-CPU upper bound,
unlike the [earlier Go-profiler capture](../native-compression-attribution-20261013/README.md).
Consistency with an upper bound does not prove unbiased component attribution or
equal compression CPU. Cgo caller chains are incomplete; inclusive owner totals
omit native work that cannot be unwound back to Go. In particular, the lower
upstream Close-chain figure is not evidence of an equivalent absolute CPU saving.

## Question and finite capture

The predeclared question was whether an independent profiler could identify
substantial current-source work without the earlier Go SIGPROF owner inference.
The budget was one sequential fork/upstream pair: two warm seconds plus twenty
timed seconds at 3,000 offered/s, 64 workers, active queue 8,192 and gzip.
Acceptance required zero lost records, usable Go/C symbols, sampled CPU within
20% of joined process CPU and Close-chain CPU no more than 1.2 times its counted
elapsed bound. Comparative owners needed at least one CPU second. Those gates
remain unchanged; upstream loss rejects the comparative inference.

Sources are `8351224` and `c4ab53d`, with the exact, identical counted-Close overlays
already archived in the earlier record. Stock gzip level, pools, types and error
behavior are unchanged. `GO_CPU_PROFILE` and allocation profiling are disabled.
Only Linux `perf` records application CPU, at the declared 997 Hz using
`cpu-clock` and `dwarf,16384` call graphs. There is no frequency or rate sweep.

The physical Fedora i9-13900KF host, seed, affinity, workers, client and complete
receipt audits match the [current-source baseline](../current-three-way-20261013/README.md).
Authentication and shared route preparation precede a fresh server restart.
`perf` attaches to the verified Go child of the container supervisor, with existing
thread identities recorded and inheritance enabled. The supervisor forwards
shutdown and joins the child with `wait4`; both application exits are clean.

All **132,000 writes** have canonical receipt/ID/body/client/creator/room/FTS audits.
Integrity passes and the seed FK population is unchanged. Counter observations are
exactly 66,000 scoped Close calls per arm. Capture windows include warmup and audit
idle intervals; joined CPU also includes startup/shutdown outside the attachment
window. Instrumented latency values are retained but are not performance controls.

## Harness failures and evidence limits

Two earlier fork captures, each with 66,000 receipts, are excluded because the
private adapter incorrectly required normal `perf` exit on its own SIGINT—first
zero, then a shell-style 130. Attached `perf` flushes data and re-raises SIGINT;
Python exposes `-SIGINT`. An isolated attached-process probe confirms that contract
before the corrected pair. These partial captures lack final joined accounting
and integrity/shutdown results. Their receipts, source variants and identities are
retained, not counted as valid performance populations or application defects.

[`manifest.json`](manifest.json) retains the declared capture, counts, reports,
period-weighted named-frame samples, canonical receipts, attachment identities,
supervisor and failed harness variants. Raw `perf.data` remains in the private
desktop workspace because DWARF captures stack memory that can contain credentials.
Its hashes are retained. Public sample exports contain periods and named frames,
not raw stack memory; unresolved frame addresses are removed. Symbol files resolve
the exact container executables through a private `symfs`. No host configuration,
profiler-frequency tuning or application dependency change is made.

The usable current-source ordinary-operation evidence motivates testing association
round-trip consolidation, not inferring an upstream regression from percentages.
The [projection experiment](../display-projection-20261013/README.md) subsequently
misses its whole-application CPU gate and is rejected. Further optimization needs
correlated critical-path waits/scheduling rather than another join-shape control.
