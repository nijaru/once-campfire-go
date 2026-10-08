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

- For application-boundary refactors, follow [the design and execution plan](plans/architecture.md).
  Settle ownership and contracts before profiling; complete affected caller migrations
  and remove superseded paths. Then profile the complete application and use representative
  measurements to resolve remaining query, allocation, cache and contention choices.
- Use the shared harness or `bench/application`: matched data, complete responses,
  persisted-write audits, settings and sequential interleaved repetitions. Keep raw
  evidence and limitations in `bench/results/`; inspect relevant rejected experiments
  before repeating them. No tests/builds running alongside timing comparisons.
- Benchmark changed workloads as well as warm hits when invalidation or database work
  changes. Do not present a cache-hit benchmark as general application capacity.
- Keep PR changes scoped and reviewable. Do not open PRs, post comments, publish or
  release without explicit permission.
