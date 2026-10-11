# once-campfire-go

Go port of Campfire. `reference/` pins Rust; `reference/reference/` pins Rails.
Use their source and the shared Campfire verification contracts to resolve behavior,
not old plans, benchmark tables, or assumptions about frameworks.

## Compatibility and implementation

- Keep existing installations usable: SQLite data, storage layout, signed/encrypted
  cookies and the frontend. Compatible query/index/schema improvements are allowed;
  preserve existing data and upgrade paths. Record deliberate behavior changes in
  README.md rather than requiring bug-for-bug or byte-for-byte equivalence.
- Do not edit `reference/`. Frontend overrides belong in `assets/overrides/`.
- Prefer idiomatic Go and the standard library. No web framework, ORM, dependency
  injection framework or JavaScript build framework. Small libraries for protocols
  and algorithms missing from the standard library are appropriate.
- Refactor the affected subsystem when needed: remove superseded paths and redundant
  work rather than layering caches or wrappers around a weak design. Preserve real
  authorization, data-integrity, cancellation and resource-lifecycle contracts.

## Verification

- Run `bin/check` before committing code. It owns formatting, asset generation,
  `go vet`, the race suite and the local WebSocket fork's tests. Builds require
  `CGO_ENABLED=1` and `-tags sqlite_fts5`; use the Go version in `mise.toml` when absent
  from PATH. Serialize asset generation and builds so embedded assets are complete.
- Use the same shared verification as Rails, Elixir and Rust:
  https://github.com/basecamp/once-campfire-verification. Run affected existing tests
  and its applicable HTTP/write contracts or browser flows. Implementation-specific
  tests complement those checks; do not invent a larger acceptance matrix.
- Add focused regressions for demonstrated defects. Reuse existing vectors and seeds.
  Run media-output checks when changing media processing, using the reference's
  toolchain/architecture; they are not a prerequisite for unrelated changes.
- Do not turn whitespace, implementation details, every browser/platform combination,
  or hypothetical crash guarantees into requirements. Distinguish actual compatibility
  failures from documented differences and test/environment defects. Report failed
  or unavailable checks without weakening assertions or claiming universal correctness.

## Performance and delivery

- Fix demonstrated correctness failures before optimizing. Use the current
  [application design contracts](plans/architecture.md), not the completed migration
  checklist or historical validation records, to preserve ownership and compatibility.
- Profile representative complete workloads before choosing optimizations. Use CPU
  and allocation evidence for expensive work; inspect blocking, scheduling and SQL
  when wall time is not explained by CPU. Improve demonstrated algorithms, query
  shape, batching and resource lifetimes before adding micro-optimization complexity.
  Complete caller migrations and remove superseded paths.
- Prefer quiet native hardware for profiling and matched performance comparisons
  when available. Keep local/VM diagnostics scoped to that environment; variable
  throughput alone is not a diagnosed Cable defect or proof of system noise.
- Before an experiment, name the hypothesis, discriminator, bounded run budget and
  landing/rejection criteria. End with a fix or a closed hypothesis that changes the
  next decision; another diagnostic archive is not a resolved performance goal.
  Inspect rejected controls before repeating them, and require new evidence to reopen
  one. Profiles of ordinary operation can justify improvements to that work, but
  cannot establish the cause or repair of an uncaptured slowdown. Historical ranking
  gaps and unexplained VM variability are not independent delivery gates.
- Use the shared harness or `bench/application`: matched data, complete responses,
  persisted-write audits, settings and sequential interleaved repetitions. Keep raw
  evidence and limitations in `bench/results/`, including adverse runs and failed
  offered populations. No task-owned agents, tests, builds or profilers during
  uninstrumented timing. Correlate individual requests rather than subtracting
  marginal percentiles; distinguish aggregate waits from wall-clock stalls.
- Benchmark changed workloads as well as warm hits when invalidation or database work
  changes. Do not present a cache-hit benchmark as general application capacity.
- Compare current source identities before claiming improvement over upstream Go or
  Rust; old snapshots and instrumented diagnostics are not current baseline proof.
  Do not buy gains with weaker authorization, cancellation, larger queues/cache
  budgets, artificial heap retention or production GC tuning.
- Prepare a PR only after affected correctness gates and matched performance evidence
  support the result; explain remaining gaps and trade-offs rather than claiming
  global optimality. Keep changes scoped. Do not open PRs, post comments, publish or
  release without explicit permission.
