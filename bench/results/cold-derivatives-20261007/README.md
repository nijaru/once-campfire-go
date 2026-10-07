# Concurrent cold derivatives

Cold requests for the same image variant or video preview now share one request-owned operation.
The leader uses its original context. A cancelled waiter leaves independently; cancelling the
leader stops its work, and live waiters retry under their own contexts. No detached context,
background processing goroutine or shared mutable Blob is introduced. Waiters read their own
committed records. Database uniqueness and transaction-owned staging remain authoritative,
including across Store instances or processes. Bookkeeping admits at most sixteen cold keys;
extra keys retain the original bounded media processing path. Warm lookups retain the original
fast path.

## Assessment and results

The original four-request cold batches consumed roughly three times the single-image CPU and
four times the single-video CPU, despite producing one committed graph. An initial whole-process
CPU profile found media-library and checksum work; it also includes login/setup and cannot
attribute ffmpeg child CPU. `/proc` accounting below includes the server and reaped children.

Final measurements used existing ARM64 OrbStack Linux tools, four application workers, guest
server CPUs 0–3 and client CPUs 4–7, identity encoding and identical SQLite backups/storage.
Three sequential interleaved repetitions alternate baseline/candidate ordering. Each run warms
libraries on separate sources, then measures fresh-source cold batches and complete warm reads.
No builds, tests or agents ran during timing.

Median batch completion and coarse server/child CPU milliseconds:

| Workload | Baseline time | Candidate time | Baseline CPU | Candidate CPU |
| --- | ---: | ---: | ---: | ---: |
| Image, one cold request | 30.16 | 28.48 | 50 | 50 |
| Image, four identical cold requests | 68.27 | 31.86 | 150 | 40 |
| Video, one cold request | 57.40 | 56.31 | 70 | 50 |
| Video, four identical cold requests | 85.25 | 57.43 | 240 | 60 |
| Image, 128 warm reads / four clients | 36.70 | 32.00 | 30 | 25 |
| Video, 128 warm reads / four clients | 28.38 | 23.80 | 20 | 20 |

All 3,132 timed responses have complete validated lengths, MIME types and identical bytes across
implementations/repetitions. Every source/derivative file matches its committed checksum and
size; exact graph/blob and new-file counts exclude loser orphans. Forty-eight cold/warm samples
are retained, with binary/tool/seed identities and the exact compressed assessment script.

This is a short batch assessment, not application capacity, offered-rate latency, steady state,
physical-core or native-platform evidence. CPU counters resolve only 10 ms. Warm batches are
client-sensitive and noisy: an earlier unchanged-production comparison showed approximately
5% slower median warm image batches, and one candidate cold-video batch took approximately
86 ms. Those populations are retained separately, not pooled into the final table. The repeated
cold sharing benefit is clear; no broad warm or whole-application speedup is claimed.

## Verification

A compact `testing/synctest` table checks shared success, cancelled leader, cancelled waiter and
both cancelled, including live-waiter retry and complete flight removal. These tests plus the
real SQLite/image/video cancellation-orphan and concurrent variant tests pass twenty race-enabled
repetitions. Native/Linux CGO vet/race and WebSocket-fork gates pass. Both ports pass all eleven
HTTP/persisted-state workflows. Production bytes used by the final measurements match the
recorded source hashes.

The optional exact video-preview vector failure remains unchanged and documented in
`../media-environment-20261007/`; neither ffmpeg flags nor vector assertions changed.
