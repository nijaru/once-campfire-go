# Cable routing index — 2026-10-06

Dense subscriber buckets replace the scan of every connected client's subscriptions for each publication. Candidate selection is O(matching subscriptions), with a contiguous snapshot copy; insertion and swap-removal are O(1). Fresh session, active-user and room-membership checks still run for each publication. Queues, timeouts, prepared frames, compression and disconnect behavior are unchanged.

Retained production: `33e21b6`. The harness's new Cable persistence/FTS checks are `c6b570a`. Baseline executable source is `9a1ba87`, with report-only HEAD `7026459`. An initial map-of-recipient-sets implementation was not retained. Its source, gates and trials remain here.

## Longer confirmation

Three rotating repetitions, 10,000 clients, 15-second saturated samples and the load generator's separate 30-message paced phase. Values are medians. Throughput counts messages received by **all** connected clients, not individual frames.

| Encoding | Baseline complete messages/s | Dense complete messages/s | Paced all-client p99, baseline → dense | Saturated all-client p99, baseline → dense |
|---|---:|---:|---:|---:|
| Uncompressed | 26.3 | 36.6 | 121.279 → 109.055 ms | 441.599 → 637.439 ms |
| No-context-takeover deflate | 25.6 | 26.4 | 163.455 → 132.991 ms | 541.695 → 741.375 ms |

Uncompressed saturated throughput improves 39%; compressed throughput improves only 3%. Paced latency improves, **but saturated p99 worsens 44% and 37% respectively**. This is a throughput/tail trade-off, not a universal latency improvement. There is no new admission throttle or artificial delay to make the saturated result look better.

The preceding matched five-second trial also includes 1,000 clients and the rejected map implementation. Baseline → dense complete messages/s: 299.1 → 318.4 uncompressed, 224.0 → 227.8 deflated at 1,000 clients; 26.7 → 33.8 and 25.3 → 26.0 at 10,000. Its tails are mixed too. See `dense-trial/matched/raw.json`; the first map-only trial is separate and not combined into these medians.

Ten-second HTTP post samples preceding the longer Cable workload are essentially unchanged: 2,241 → 2,259 requests/s, 675.5 → 677.7 CPU µs/request, p99 28.831 → 28.367 ms. End-of-workload PSS is 956.7 → 949.6 MiB. Those PSS snapshots follow HTTP writes and both 10,000-client Cable modes; they are **not** steady-state live-client memory measurements or evidence of a reliable memory reduction.

## Profiles and library assessment

`current-profiles/` holds fresh, separate warm search, writes, fragment-disabled search and 1,000-client compressed Cable profiles of the baseline. They identify:

- Warm search: fresh SQLite references/row stepping, about 30% flat cgo attribution.
- Writes: commit about 34% cumulative; this does not establish checkpoint causality or justify weaker durability.
- Cold search: templates about 50%, rich-text processing about 22%, with overlapping cumulative stacks.
- Cable: socket writes about 33%, publication about 9%, per-frame timeout construction about 11%.

`cable-profiles/` reruns the same profiled 1,000-client compressed workload for baseline and dense. Publication falls from 8.45% cumulative (3.19 sampled CPU seconds) to 5.07% (1.83 seconds). Socket syscalls and timeout bookkeeping remain dominant. These profiled runs are diagnostic, excluded from timing medians.

No new WebSocket or compression library is justified by this profile. Prepared messages already share compression, and the inherited buffered writer combines a frame's header and compressed payload into one underlying write when they fit its buffer. Vectored header/payload output would not remove that call under that condition. Batching multiple queued frames could change ordering, cancellation, backpressure and latency; queue accumulation and actual frame sizes would need measurement before such a trial.

## Verification

Across the three unprofiled suites: **21 process runs, 21 HTTP samples, 72 Cable samples, zero HTTP errors, 390,431 HTTP posts and 50,884 Cable posts**. Every Cable client connected successfully; all 2,160 paced messages and 48,724 saturated messages reached every client. All acknowledged posts persisted and matched FTS; the harness checks workload-local deltas. Counts are in `validation-summary.json`.

Native `bin/check` and Linux vet/race gates, including the local WebSocket fork, passed before timing. Logs are in `checks/`. Protocol tests cover repeated identifiers, distinct aliases on one socket, selective unsubscribe, resubscribe, stream delivery, fresh membership revocation and disconnect cleanup. The routing test checks inverse positions, authorization-scope changes and cleared spare capacity. Read-only safety reviews finished before timing; no agents, builds or tests ran during timing.

## Environment and reproduction

Matched public HTTP/1.1 on the Apple M3 Max's **OrbStack Linux ARM64 VM**, 16 vCPUs. Server GOMAXPROCS/workers 4 on CPUs 0–3; load generator CPUs 4–7. Logging remains enabled. Identical parity seed, application listener, validation and candidate rotation; Go 1.27.1, normal SQLite 3.53.4. Image, executable hashes and complete staged-source hashes are in `manifest.json`; each suite's metadata records commands and environment. Logs stayed in container `/tmp`, were gzip-compressed after timing and then copied out.

Build baseline and production using CGO, `-trimpath -buildvcs=false -tags sqlite_fts5`; then, in the common toolchain container:

```sh
python3 bench/application --rust-root reference \
  --seed .cache/parity-seeds/default --loadgen .cache/linux/loadgen \
  --candidate base=.cache/linux/final-search-shell \
  --candidate dense=.cache/linux/final-cable-dense --apps base dense \
  --listener public --gzip 1 --routes post_message --concurrency 16 \
  --reps 3 --seconds 10 --server-cpus 0-3 --loadgen-cpus 4-7 \
  --cable-clients 10000 --deflate 0 1 --cable-seconds 15 \
  --upload-reps 0 --out /tmp/cable-confirm
```

Retained `application.py` is the exact harness used. The smaller trials use five-second HTTP/Cable samples and both 1,000/10,000 client counts. No Rust/Ruby/Elixir or native Fedora Cable comparison was run. VM scheduling, fixed four-core client load, one seed, one shared authenticated user and short samples constrain generalization. The saturated load is closed-loop with four HTTP posters; it is not a fixed offered-rate comparison. The throughput gain must not be multiplied by historical HTTP gains or transferred to published AMD results.
