# Remaining render candidates — 2026-10-06

The timestamp/avatar preparation experiment was **reverted**. Only two unchanged constant regexes are retained in production `7afad06`; benchmark `63ad07a` extends exact decoded response comparison beyond forms. Baseline is `96cb2df`.

## Rejected preparation trial

The private candidate formatted repeated creation timestamps once with template-local variables, and prepared creator title/avatar once per fresh hydration batch rather than retaining full users and signing per message. It introduced no cross-request cache. The former uncached message template remained the independent byte/escaping reference, including zero/offset/wide timestamps and hostile strings. Fresh creator A/B/A metadata and captured-value ownership were checked. The two constant regex hoists were also present in this combined trial.

The initial fragment-disabled room/history/search pair used identity and gzip, 16 clients, three alternating five-second samples. All six route/encoding pairs reduced median CPU/request about 3–5%, but search throughput fell about 2–3%. Longer alternating 30-second search samples did not resolve the inconsistency:

| Search, no fragments | Requests/s baseline → candidate | CPU µs/request | p99 ms |
|---|---:|---:|---:|
| Identity | 1,135.8 → 1,094.0 | 1,813.06 → 1,955.98 | 47.903 → 48.543 |
| Gzip | 1,107.2 → 1,129.6 | 2,020.51 → 1,888.86 | 49.343 → 47.647 |

Identity end PSS rose 58.7 → 61.7 MiB; gzip fell 60.8 → 59.0. Six-route warm/write controls were mixed and broadly unchanged; gzip posts fell 2,442.2 → 2,390.5 requests/s, with p99 26.159 → 26.735 ms. These results do not identify which preparation change caused the differences. The modest, inconsistent application benefit did not justify retaining the extra representation and template state. The complete prototype source/executable hashes, patch, private tests and measurements remain in `manifest.json` and `rejected-format/`. Production creator/timestamp rendering was restored.

## Retained fixed regexes

`externalURL` and `UnverifiedUserGID` now compile their unchanged patterns once. No algorithm, matching semantics, released interface or dependency changes. Existing rich-text oracle/signing coverage and full native/Linux race gates pass.

Private probe executables were built before timing, then run sequentially on the same server affinity, with alternating order and three two-second repetitions. Median costs:

| Boundary | Before → after ns/op | Bytes/op | Allocations/op |
|---|---:|---:|---:|
| External HTTP URL validation | 1,085 → 270.3 | 960 → 177 | 12 → 2 |
| Legacy unverified GID envelope | 5,824 → 1,426 | 5,199 → 1,033 | 42 → 13 |

These are deliberately exercised scalar paths, **not application speedups**. The separate complete-response cold room/history/search and warm room/search/post trials are mixed, with no consistent overall HTTP gain. For example, warm gzip room improves 22,188.0 → 23,206.8 requests/s, while identity falls 16,147.1 → 15,647.5. We retain the tiny allocation-removing cleanup, not a broad throughput claim. Probe sources and raw logs are in `regex/`; none are added to the permanent production test suite.

## Verification and stopping assessment

**66 application runs, 228 HTTP samples, zero errors; 542,101 HTTP posts and 5,001 active posts verified in messages and FTS.** This total includes the rejected trial and independent regex trials, not the diagnostic profile. Cold Go pairs require identical complete decoded response bytes after normalizing only the known listener origins; each process also requires exact identity/decoded-gzip equivalence. No nonce, cursor or content masking is enabled.

The final 20-second fragment-disabled gzip search profile attributes 57.74% cumulative CPU to `html/template.ExecuteTemplate`, 7.75% to rich-text processing and 4.96% to parsing. Shares overlap and are not isolated savings. `profile/` retains the profile and measured executable hash and is excluded from timing medians.

Native `bin/check`, Linux vet/race, all rich-text oracle cases and the WebSocket fork pass before production commits. Read-only review found no production blockers; its two test gaps were corrected before timing. No agents, builds or tests run during application timing. All bulk logs remain in container `/tmp` until measurement ends, then are gzip level 9 compressed and copied out. Source/executable hashes are recorded in `manifest.json`; exact staged sources/binaries remain at their listed local paths.

The measurement environment is the Apple M3 Max's **OrbStack Linux ARM64 VM**: Go 1.27.1, Rust 1.98.0, four server workers/CPUs 0–3, load generator CPUs 4–7, identical seed, public HTTP/1.1 and logging enabled. Workload populations differ between the cold, search-only and warm suites; do not compare their PSS or multiply phase gains.

The evaluated DSA, library and micro-optimization candidates have reached diminishing returns under the compatibility constraints. Generated-template/parser migrations are not byte/context-compatible drop-ins; rejected database/compression/PGO experiments have no new evidence warranting repetition. We stop here rather than claim an absolute performance ceiling. Desktop access subsequently returned; [the native follow-up](../native-final-20261006/README.md) separately tests the retained application on physical P-cores. Ruby/Elixir, native Cable, identity/native concurrency sweeps and larger/churning datasets remain outside that follow-up.
