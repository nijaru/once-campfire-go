# Later-PR investigation: SQLite accounting and rich-text serialization

No production change. This record tests one small PR #10 mechanism and the functional cost of PR #11's `SQLITE_DEFAULT_MEMSTATUS=0`; it is not an application-performance comparison.

## SQLite: a real trade-off, no measured application benefit

Disabling memory statistics avoids SQLite's global allocator-accounting mutex. It also disables global memory-use/high-water counters and soft/hard heap-limit enforcement. It does **not** disable SQLite thread safety, transactions, FTS, cancellation or the existing connection-local page-cache setting.

A native C probe against Go's pinned SQLite 3.53.4 confirms that an allocation above a configured hard limit fails by default but succeeds with the flag. The disabled library still accepts and reports the limit setting, so checking that setting alone is insufficient. `sqlite3_db_status(SQLITE_DBSTATUS_CACHE_USED)` still reports 18,048 bytes in both probes. This is a feature test, not a timing benchmark; the C libraries are not application binaries.

The application currently calls neither the affected global status APIs nor heap-limit APIs. Its connection-local cache and prepared-statement bounds remain. The flag changes an initialization default, not database format or durability, and can be reversed. Nevertheless, retain the current default until a complete-workload benefit justifies losing these diagnostics and SQLite-specific limits. No flag was added to builds, checks or containers.

Other ports, inspected on 2026-10-14:

- **Express `bb70a7d`:** its default Dockerfile pins Node 24.21.0, whose bundled SQLite build explicitly sets `SQLITE_DEFAULT_MEMSTATUS=0`; the port uses `node:sqlite`. A local Node 26.11.1 probe also reports the flag, but is not a runtime test of the pinned container. Its optional Bun path was not certified.
- **Rust `6dae2fd`:** its actual freshly built `libsqlite3-sys` archive uses SQLite 3.53.2 with accounting enabled. The same probe reports nonzero memory and enforces the hard limit. Its `SQLITE_ENABLE_MEMORY_MANAGEMENT` build flag is separate from `SQLITE_DEFAULT_MEMSTATUS`.
- **Elixir `58964cf`:** locked exqlite 0.42.0's Makefile and NIF initialization do not disable memory status. They configure allocator callbacks, not this bookkeeping flag. This is source inspection, not a runtime probe of the port or its precompiled NIF.
- **Rails `edbc779`:** locked sqlite3 gem 2.9.6's packaged-library flags and extension entry point contain no such disabling setting. Platform/system-library and precompiled-gem options can differ; no Rails runtime verdict is claimed.
- **C `873a51f`:** the inspected SQLite build/probe flags and application startup/database files contain no disabling setting. No C application binary was built or measured.

Sources: SQLite's [compile option](https://sqlite.org/compile.html#default_memstatus), [configuration contract](https://sqlite.org/c3ref/c_config_covering_index_scan.html#sqliteconfigmemstatus) and [heap-limit contract](https://sqlite.org/c3ref/hard_heap_limit64.html); the artifact source ledger identifies inspected dependency files and upstream URLs. An initial probe compile accidentally selected the SDK's older SQLite header; explicit pinned headers corrected it. The Rust archive linker also warned about its newer deployment target. Final probes ran successfully; neither setup issue supplied timing evidence.

## PR #10: direct rich-text escaping is low-yield

Compared current application `8351224` with PR #10 head `029c943`, not upstream main. Only `serializeTo` changed privately: `textEscaper.WriteString(builder, text)` replaces `builder.WriteString(escapeText(text))`. The immutable standard-library replacer and call-owned builder remain; no arena, pooling, borrowing, attribute change or JSON-escaping change.

Both source variants pass existing rich-text oracle/document/display and presentation race tests. A task-private byte/UTF-8 boundary test compares direct writes with the standalone escaper. After builds/tests ended, exactly three alternating pairs measured complete `Prepare`, `Content` and `Display` operations on plain text, escaped text and one retained 908-character seed body, with Go 1.27.1, Mac ARM64, `GOMAXPROCS=4`, `GOGC=100` and 500ms per case. Background host activity was not isolated.

| Case | Baseline → candidate bytes/op | Objects/op | Median elapsed change |
|---|---:|---:|---:|
| Plain | 4,200 → 3,992 (−5.0%) | 93 → 77 | −7.1% |
| Escaped | 11,384 → 10,152 (−10.8%) | 206 → 184 | −2.9% |
| Seed | 102,836 → 88,609 (−13.8%) | 1,488 → 1,061 | −1.5% |

All cases miss the preregistered 15% complete-rich-text byte-saving gate. The retained native live profile attributed only about 0.34% of whole-process Go allocation bytes to `escapeText`, and native `serializeTo` inclusive CPU was 0.57%. Those sampled figures are not exact recoverable savings or a hard bound, but they provide no material whole-application premise. Close as low-yield: no complete-POST profile, capacity/latency cohort, refinement or integration follows. These helper timings do not establish application CPU, capacity or tail-latency improvement. There are zero new application writes.

The review also closes child traversal snapshots as smaller low-yield work and webhook keepalive as unproven in representative workloads. Current prepared-document/typed-command ownership already covers PR #10's combined derivations. Permission memoization, retained verified-cookie caches, broader HTTP/database-writer machinery, batching, GC tuning and durability changes were not selected.

## Evidence

`manifest.json` owns compressed/raw hashes and sizes for the finite probe's criteria, all six benchmark outputs, conclusions, focused checks, private prototype diff/test and SQLite probe/source ledger. Seed body, application binaries, source archives and unrelated runtime data remain private. Linux application tests were not run in this investigation; desktop access returned only after the Mac probe. Production remains exact `8351224`.
