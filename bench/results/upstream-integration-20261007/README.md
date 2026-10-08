# Upstream integration verification — 2026-10-07

The integration starts at upstream `352706f` and replays the 16 local commits after
`bc1ac23`. Our merged PR #9 was already present in that base; it is not replayed.
Unmerged PR #10 is not included. The published `fix/functional-parity` branch is
unchanged. `metadata.json` identifies the replay head, affected source bytes and
verified binaries before the integration correction commit.

## Contracts retained

Search keeps the pinned Rust/Rails `created_at DESC LIMIT 100` selection and
presentation ordering, rather than upstream's insertion-ID search. The existing
schema is retained: upstream's additional `updated_at` index is not installed.
The removed upstream search tests exercised its superseded ID-order/probing
algorithm; existing chronological search, reachability and reference/full-result
comparisons remain. Authentication, cursor scope, nullable involvement, partial
updates, atomic storage staging, owned response parts and narrow sidebar hydration
remain covered by their existing tests and paired workflows.

Compatible upstream changes include completed-response caching, nested cache
namespaces for database generation and origin, rendered-body pagination validators,
sidebar refresh/editor preservation and complete session-transfer forms.

## Cache defects reproduced and corrected

- Overlapping `PRAGMA data_version` statements on the pinned connection reuse an
  older SQLite read transaction. The diagnostic held a stepped row across a foreign
  commit: generation stayed **3** after the commit and advanced to **4** only after
  closing the old row. The permanent concurrent-reader/foreign-writer regression
  failed in all 20 before-fix runs. `ResponseVersion` now serializes the entire
  query/scan/close lifecycle with cancellation-aware admission. Both that regression
  and queued cancellation passed 20 race repetitions. The temporary diagnostic
  is not retained as a permanent test.
- The completed-response key omitted `Content-Type`, although XHR format negotiation
  uses it when `Accept` is absent. Warming HTML then requesting JSON returned cached
  **200**, rather than the required **406**. The regression fails before including
  `Content-Type` in the key and passes 20 race repetitions afterward.
- A cache-level regression warms closed-room search and sidebar representations,
  revokes membership through an independent SQLite writer and checks that neither
  representation exposes the formerly reachable content.

## Verification

- Native `CGO_ENABLED=1 bin/check`: formatting, asset generation, vet, full race
  suite and the WebSocket fork race suite passed.
- ARM64 Linux: vet, full race suite and WebSocket fork race suite passed; the
  candidate was built afterward with `sqlite_fts5` in the existing toolchain image.
- All **15 paired workflows / 30 Go–Rust runs** passed with isolated seeded SQLite
  and storage snapshots. The final report and logs are retained compressed.
- Focused Chromium tests use the actual served Turbo, Stimulus and sidebar-controller
  bytes, a real unfinished HTTP response and isolated Cable callbacks. They verify
  waiting before reconnect reloads, preservation of unfinished input and selected
  recipients, deliberate cancellation and unsubscribe on frame removal. The old
  binary fails the unfinished-frame check; the integrated binary passes repeated
  runs. The existing two-await cancellation fixture also passes.

**The full application browser smoke is not green.** The integrated binary timed
out waiting for a live edited message; other attempts timed out on initial live
messages. The prior native binary also timed out on initial live messages, with
view-transition timeout diagnostics. This does not establish an environment cause
or rule out an application defect. Original assertions are unchanged; the failures
remain archived. Focused sidebar verification is not full browser or actual Cable
reconnect acceptance.

No media encoder, pinned source/vector, version, durability setting or benchmark
mask changed. These checks do not prove universal correctness or Rust performance
parity. Historical benchmark archives remain historical; this record has no timing
comparison for the integrated source.
