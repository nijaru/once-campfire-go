# Native WebSocket frame-encoding control — 2026-10-11

**Closed without promotion.** Both private prototypes preserve decoded JSON and
retain the existing copying prepared-frame constructor. Neither met the declared
allocation criterion. Production remains unchanged; no uninstrumented timing
comparison was run and no latency, capacity or historical-slowdown fix is claimed.

## Question and budget

After the desktop became available, capture one current native CPU/allocation
profile to confirm publication materialization owners. Test transport-local
`json.Encoder.SetEscapeHTML(false)`: Action Cable parses these frames as JSON, not
as HTML script text, so expanding markup `<`, `>` and `&` is unnecessary. This
control does not alter renderer sanitization or authorization.

The criterion was at least 10% fewer sampled Go allocation bytes per audited
message, without a material CPU/message regression, before three uninstrumented
native delivery pairs. The first prototype missed that criterion and allocated
more objects. One explicit refinement reused the encoder/buffer within each
publication and connection writer; prepared messages still copy before reuse.
That sole refinement also missed the criterion. The budget is closed, with no
follow-up timing sweep, pool, ownership API, database binding, runtime, queue or
cancellation change.

## Capture and scope

- Current source `71f2637`; production source remains `ccfa838`. Shared verification
  client/contracts checkpoint `ec02debb3e44f40b8778a6a42db3c4515b2f9eaf`.
- Native Fedora desktop, rootless Podman, isolated **Go 1.27.1 linux/amd64** image.
  Server affinity `0,2,4,6`, client `8,10,12,14`, runner `16,18`; affinity is not
  exclusive host reservation or proof of physical-core isolation.
- Identical disposable seed copies and settings: 64 distinct authenticated people,
  100 paced messages at 100 ms, then five seconds/four saturated posters, plus
  complete delivery and persisted-write audits. Setup/login process stopped before
  profiling restart. Default Go memory sampling and `GOGC=100` are unchanged.
- CPU recording ends during graceful shutdown. A **private-only** allocation hook
  then forces collection twice and writes the cumulative `allocs` profile after
  joined shutdown. It is absent from production and unsuitable for timing controls.
- CPU and allocation captures include restart, idle/drain and shutdown work, not
  solely publication. Allocation excludes native SQLite/C and kernel allocations.
  One capture per version supports attribution, not a precision/performance claim.

## Results

| Version | Audited messages | Sampled Go bytes/message | Sampled Go objects/message | CPU ms/message |
|---|---:|---:|---:|---:|
| Current | 15,856 | 204,580.91 | 2,078.28 | 0.95106 |
| First encoder prototype | 16,015 | 185,924.17 | 2,221.40 | 0.93662 |
| Sole encoder-reuse refinement | 16,062 | 195,385.53 | 2,194.12 | 0.94135 |

Relative to current, the first prototype reduces sampled bytes/message **9.12%**
and CPU/message **1.52%**. The refinement reduces bytes/message only **4.49%** and
CPU/message **1.02%**. Both allocate more objects/message. These are sampled,
complete-process figures normalized by exact audited writes, not isolated helper
costs, wire-byte measurements or uninstrumented application performance.

The current native allocation profile attributes approximately **55.45%** of Go
bytes cumulatively to `Hub.publish` and **16.78%** to JSON quoting. Socket writes
account for approximately **42.37%** of sampled CPU, versus **11.07%** for fresh
session authorization. This confirms distinct materialization and transport work;
it does not identify the cause of any historical adverse phase. The rejected
owned-copy and packed-authorization controls remain closed.

All **47,933** acknowledged writes passed shared exact ACK-ID/body/creator/room,
FTS, presence and foreign-key audits, with complete all-client delivery. Profiled
delivery rates were 3,150.1 / 3,182.0 / 3,190.9 messages/s respectively; do not use
these instrumented rates as promotion or capacity evidence.

## Correctness and evidence

Both prototypes passed private `bin/check`. Native affected `go vet` and race tests
passed. Existing real WebSocket room/alias authorization and compression coverage
ran; a focused codec probe preserved decoded identifier, markup, nested raw JSON,
large integer, literal escape, control/Unicode and invalid UTF-8 content, and
rejected non-finite numbers. The probe and production changes were not landed.
No shared browser/media rerun was required after rejection; no new native timing
or unrelated upstream comparison was performed.

[`manifest.json`](manifest.json) hashes 41 retained artifacts: native raw profiles,
exact client ACK records/audits, runner/settings metadata, source patches, private
allocation hook, build/check logs and analysis. Text is gzip-compressed; profiles
retain their native pprof format. The seed and measured source inputs are hashed
in metadata/manifest. Build extraction warnings concern macOS archive attributes,
not workload execution. Profiles can be inspected with `go tool pprof -top`.

The next decision is to inspect a **different demonstrated owner** using these
profiles, not to reopen frame-copy, HTML-escaping or room-preflight controls on
benchmark ranking alone.
