# Shared reaction-form representation screen — 2026-10-14

**Rejected at the representation gate; no implementation or native timing.**
Consolidating eight quick-reaction forms into one form reduces raw captured HTML
by **20.06%**, but the modeled escaped WebSocket envelope by only **19.37%**.
The preregistered screen requires at least 20% for both. It is not relaxed.
Production remains `8351224`; no UI change, browser compatibility claim or new
performance result follows.

The source repeats each message's form action and Turbo target eight times.
A single form can instead carry eight named `boost[content]` submit buttons,
retaining the existing `quick-boosts` grid class, labels, emoji and ninth new-boost
link. This is a representation hypothesis, not the earlier prepared literal
reaction bodies, JSON escaping, compression or cooperative-handoff controls.

One deterministic transform reuses the **256 native receipts** from the
[encoder experiment](../post-critical-path-deflate-20261014/README.md). It removes
repeated form/hidden-input wrappers and moves each hidden value to its submit
button. Independent HTML parsing compares all eight button attributes (apart
from the intended field/value migration) and the new-link attributes. It does
not prove browser submission, layout or Turbo behavior; those checks would have
preceded timing only if the representation screen passed.

| Aggregate candidate/current size | Ratio | Reduction |
|---|---:|---:|
| Raw HTML | 0.799390 | 20.061% |
| Modeled Go-style escaped JSON envelope | 0.806309 | 19.369% |
| Python zlib level-6 gzip | 0.978032 | 2.197% |

The envelope uses Go's HTML escaping and a constant identifier, not actual socket
receipts. Real identifier overhead would further dilute the reduction. Python
zlib is a preliminary size screen, not Go 1.27.1 compressor or CPU evidence;
no individual body grows in that screen. Native standard-Go gzip validation is
not run after the JSON-size gate fails. No new corpus, writes, codec/rate sweep,
code/dependency change, browser run or conditional performance control occurs.
The first script failed its assertion before saving its summary; the reporting
repair saves the same population and the failed gate, without changing criteria.

[`manifest.json`](manifest.json) retains the declared budget, transform, all
per-body sizes, derived bodies, failed report and corrected failed screen. Existing
native source identities and receipt provenance remain in the encoder record.
These are **zero new audited writes**, not another 256-write population.
This closes this bounded form-consolidation control; meaningful fixed-POST and
Rust live-delivery optimization remain unfinished.
