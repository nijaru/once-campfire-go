# Native Rust baseline and scoped search — 2026-10-11

The first current native Rust application comparison closes a measurement gap.
A separate structural search change removes the global probe and restart: one
membership-scoped, descending FTS iterator query
selects the newest 100 accessible IDs in the existing read snapshot. Full sparse
search throughput improved **25.4%** and capacity-run p99 improved **12.7%** across
three matched pairs; the seeded dense control was essentially unchanged.

## Source and environment

The native baseline checkout is `9990759`, with application source unchanged from
`811ae04`; upstream Go is `c4ab53d`, Rust `6dae2fd`, and shared verification `ec02deb`.
The search candidate is that Go baseline plus the exact
[`search-candidate.patch.gz`](adapters/search-candidate.patch.gz). It changes no
schema, storage format, authorization policy, cache budget or writer behavior.

Fedora native x86_64 on the i9-13900KF, isolated rootless containers, Go 1.27.1/CGO/
`sqlite_fts5`/`-trimpath`/`-buildvcs=false`, Rust 1.98.1 locked release build with its
workspace profile, and libvips 8.16.1. The fresh Rust build extends the accepted
pre-wrapper toolchain, not the image containing the rejected literal wrapper.
Accepted Go binaries are reused with their unchanged source identity.

Server CPUs were `0,2,4,6`, clients `8,10,12,14`, and driver `16,18`. Affinity does
not control SMT sibling activity, background work or thermals. Four runtime
workers/readers, three job workers, gzip, `GOGC=100` and the same seed are used.
Baseline caches use each application's defaults; this is not an equal-cache-memory
comparison. Active fixed-rate pacing costs roughly one generator CPU. No task-owned
build, test, agent, profiler or process-counter observer ran during timing.

Binary SHA-256:

| Binary | SHA-256 |
|---|---|
| Current baseline Go | `6446e567b3f3486fb9616e8fe84f155fd4db718798b733d01c143b98e95505e7` |
| Upstream Go | `70cca09f7fa45eac8e383a2160990ac1ad9a1bdd9c6b95f47dc666b60b987b54` |
| Rust | `37599be35e60db3532c7887b7ac28e5fcbc576cc0d77c80de3edabc10a78fed4` |
| Search candidate Go | `165e7570739f61c5ea657985109ac703c2714b2df7fd828a3cc0eb73844f67b2` |

## Native baseline: all three rounds

Orders were upstream→current→Rust, Rust→current→upstream, current→upstream→Rust.
Each workload starts from a fresh fixture copy. POST capacity uses two warmup
seconds at four clients and eight timed seconds at 16. Fixed-rate POST uses
3,000 offered/s, 64 workers, 8,192 queue slots, active pacing, two warmup seconds
and eight timed seconds. Capacity responses use shared semantic contracts and
acknowledged-ID/body audits; fixed-rate timing consumes complete wire bodies and
independently audits exact populations/bodies, unique client identities,
creator/room and normalized FTS. Fixed-rate response IDs are not captured.

| Implementation | POST capacity req/s, rounds 1/2/3 | Fixed-rate POST p99 ms, rounds 1/2/3 |
|---|---|---|
| Upstream Go | 9144.7 / 9111.5 / 8506.5 | 0.777 / 0.850 / 0.798 |
| Current Go | 9808.0 / 9745.8 / **4819.5** | 0.943 / 0.910 / 1.006 |
| Rust | 7068.1 / 7077.5 / 7104.2 | 0.626 / 0.621 / 0.619 |

Current median POST capacity is 9745.8/s versus Rust 7077.5/s, but the third Go run
loses substantially. It remains a valid audited adverse sample, not a discarded
outlier or an explained failure. This does not establish a stable capacity advantage.
At matched load, current p99 loses every pair: median **0.943 versus 0.798 ms**
upstream (+18.2%) and **0.621 ms** Rust (+51.9%). Capacity-run p99 is not substituted
for matched-load latency.

Cable uses the existing shared 64-person workflow: 100 paced messages at 100ms
intervals, followed by four saturated posters for five seconds. Ordinary clients
exercise real subscriptions and presence; authentication and queue limits remain
unchanged. Every sample passes complete all-client delivery, exact unique-marker/
persisted-count/FTS/integrity and zero-presence-after-disconnect audits.

| Implementation | Delivered/s, rounds 1/2/3 | Paced all-client p99 ms, rounds 1/2/3 |
|---|---|---|
| Upstream Go | 1989.7 / 2383.2 / 2383.5 | 4.555 / 6.499 / 5.971 |
| Current Go | 3362.0 / 3355.1 / 3372.1 | 4.451 / 6.619 / 5.571 |
| Rust | 4225.2 / 4314.9 / 4316.2 | 4.099 / 5.423 / 5.419 |

Current median delivery is **22.1% below Rust**, while exceeding upstream Go by
41.1%. Paced latency varies; the small median Go/Rust gap is not a general latency
conclusion. No severe low-rate Cable phase occurred in these native samples. This
does not explain or erase the prior VM timing observations. All **1,144,332
timed/warmup writes** across the 27 baseline runs passed their applicable audits.

## Search design discriminator

The small shared seed has 169 messages and 13 coffee matches; it never exercises
the old 1,000-row fallback. The sparse workload therefore follows the existing
`TestSearchFindsSparseMembershipBehindInaccessibleMatches` pattern, rather than
claiming a common-seed gain: 100 older accessible messages and 1,100 newer matches,
with 50 accessible among the newest 1,000. Membership and content are unchanged
between candidate and baseline.

An independent SQLite 3.53.1 probe matched exact IDs for zero/50 accessible prefix
cases. Continuing below the probe cutoff reduced total VM steps by about 22–24%,
but retained two queries and Go filtering. A single scoped FTS query reduced
steps **34144→10267 / 34494→9867** and removed the temporary sort. These diagnostic
VM counts are not Go CPU or application performance measurements. A separate
probe of the actual Go-bundled SQLite **3.53.4** confirms descending FTS
`INDEX 192:M0`, primary-key/message and covering-membership lookups, with no sort
in that observed plan. The simpler single query supersedes the continuation proposal.

The change removes `searchProbeLimit` and its custom filtered-row scanner. The
existing owned-reference scanner, selection/body snapshot, newest-ID ordering,
reversal, foreign-aware generation checks and error propagation remain intact.
Inaccessible timestamps are filtered in SQL before materialization.

## Search: matched full-response comparison

Three baseline/candidate pairs reverse the second pair. Eight clients, two-second
four-client warmup, eight timed seconds, gzip and four server workers are used.
Completed-response and fragment caches are disabled equally to exercise real
query/materialization work; no cache-hit or general application capacity gain is
claimed. The dense control is the existing seeded coffee query. Every complete
response uses shared validation with independently selected ordered message IDs,
content and DOM identities: **374,965 validated warm/timed responses** total.

| Pair | Sparse baseline/candidate req/s | Sparse baseline/candidate p99 ms | Dense baseline/candidate req/s | Dense baseline/candidate p99 ms |
|---|---|---|---|---|
| 1 | 1009.1 / 1265.1 | 17.103 / 14.735 | 5203.8 / 5203.4 | 4.979 / 4.975 |
| 2 | 1012.4 / 1260.0 | 16.671 / 15.039 | 5176.9 / 5094.3 | 5.095 / 4.963 |
| 3 | 1007.4 / 1267.7 | 17.055 / 14.887 | 5163.0 / 5181.4 | 4.907 / 4.999 |

Sparse medians are **1009.1→1265.1/s (+25.4%)** and
**17.055→14.887 ms (−12.7%)**, winning all three pairs. Dense medians are
**5176.9→5181.4/s (+0.09%)** and **4.979→4.975 ms (−0.08%)**; individual losses
remain visible. These are medians of sample statistics, not pooled percentiles.

## Verification and retained evidence

- Expanded sparse coverage preserves the original fewer-than-100 case and adds
  interspersed accessible results exceeding the cap, exact ordered IDs/bodies and
  inaccessible malformed timestamps. The tests also pass original source: this is
  an optimization, not a newly claimed correctness defect.
- `bin/check`, focused native/macOS and Linux x86_64 database/web races (three
  repeats), and the existing snapshot/ordering/permission oracles pass.
- A private `sqlite_trace` probe gates entry to real SQLite execution, cancels,
  joins the search and verifies single-reader reuse; three native races pass.
  The optional hook and probe are not production code or permanent tests.
- Four applicable Go/Rust HTTP workflows—message lifecycle, pagination, room
  membership and authentication—pass against the candidate and fresh Rust binary.
- The unmodified shared browser smoke passes with Playwright 1.63.0/Chromium in
  an isolated rootless container. The initial Mac attempt lacked its pinned browser;
  a subsequent SSH shell-syntax setup error is retained. Neither exercised an
  application failure. No assertions or asynchronous-count checks were changed.
- Read-only independent source review found no concrete correctness defect.

[`manifest.json`](manifest.json) covers sanitized samples, exact adapters/patch,
query probes, source/binary provenance and gate/build logs. Duplicate raw samples,
private fixtures/databases, cookies, signed paths and runtime/server logs are not
published. Media algorithms were not changed; media-output inventories were not
rerun. This search improvement does **not** fix or explain POST tails, the adverse
native POST-capacity sample, or the intermittent low-throughput VM observations.
Those measurements remain historical evidence, not diagnosed correctness defects
or standalone readiness gates. The current
[profile-led workflow](../../../plans/architecture.md#verification-and-design-changes)
uses demonstrated costs to choose improvements and matched end-to-end checks to
validate them; it does not require closing every historical ranking gap.
