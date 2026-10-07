# Immutable sidebar response parts

Sidebar responses now combine an owned frame with owned surrounding document/frame-layout parts.
The existing byte-bounded fragment store retains both independently. Fresh session, membership,
participant, placeholder, account, profile, flash and request data still precede selection. No
query-result cache, detached context, schema change or new caching framework was introduced.

The same bounded layout input drives template execution and identity. Only the frame insertion
marker is substituted during preparation. All payloads own their bytes beyond pooled-buffer reuse,
eviction and disabled-cache operation. Stable part identities also let the existing compression
memo reuse a completed representation without hashing and rendering the surrounding document again.

## Before/after

Three sequential interleaved repetitions compared the unchanged `218374b` application binary with
the candidate on identical seed snapshots, gzip and guest affinity: four server CPUs (0–3), client
CPUs 4–7, 16 concurrent clients, two-second warmups and ten-second samples. Both are ARM64 release
binaries in the same OrbStack environment on an Apple M3 Max. No builds, tests or agents ran during
timing. Guest CPUs are not dedicated physical cores.

| Full sidebar document | Before | After |
|---|---:|---:|
| Median requests/second | 11,990 | 14,805 |
| Median CPU µs/request | 219.06 | 158.58 |
| Median saturated p99 ms | 6.835 | 5.727 |

Median throughput rises 23.5%, CPU/request falls 27.6%, and saturated p99 falls 16.2%. The candidate
has higher throughput in every paired repetition, but the VM remains noisy. This is not a
matched-offered-rate latency claim or proof of Rust-level performance. Before ranged from 11,540
to 14,182 requests/second; after from 14,691 to 15,200.

Every complete decoded response matches the before implementation after normalizing only its
known listener origin. Message/room contracts and response encoding remain validated; no legacy
cursor mask was used. Raw samples and complete source/binary/seed/tool identities are archived.
A separate diagnostic CPU profile is not included in the comparison medians.

## Correctness and cleanup

A focused template comparison exercises cold/hit, document/frame and disabled-cache paths, including
profile escaping, role, logo, CSS, flash, permissions and stream changes. Real HTTP tests continue
to verify fresh layout data and frame/document structure. Their old private cache-entry-count
assertion was replaced with an observable assertion that layout-only changes leave the returned
frame unchanged; the store now legitimately contains surrounding-byte entries too.

Native CGO `bin/check`, Linux CGO vet/full race tests with `sqlite_fts5`, the WebSocket-fork race
suite and all 15 paired workflows pass. The obsolete `page.SidebarHTML` controller field and
HTML-only sidebar preparation path are removed. No browser sweep was needed.
