# Current Go / upstream Go / Rust — desktop, 2026-10-13

The fork's fixed-load POST deficit repeats against both unchanged upstream versions.
The fork retains higher POST capacity than both and higher live delivery than upstream
Go, but trails Rust on live delivery. This is a current-source comparison, not evidence
that a particular fork change introduced the deficits.

| Median, three rounds | Fork Go | Upstream Go | Rust |
|---|---:|---:|---:|
| POST p99 at 3,000 offered/s, ms | 0.965 | 0.831 | 0.635 |
| Saturated POST capacity, requests/s | 9,765.5 | 9,094.9 | 7,155.7 |
| Saturated POST p99, ms | 6.199 | 6.519 | 3.479 |
| Live complete-delivery throughput, messages/s | 3,087.8 | 2,364.3 | 4,351.9 |
| Paced all-client delivery p99, ms | 6.507 | 7.311 | 4.951 |

Fixed-load p99 is **16.1% higher than upstream Go and 52.0% higher than Rust**.
The fork loses each of the three matched rounds to both. The ranges are
0.919–1.018 ms for the fork, 0.792–0.849 for upstream Go and 0.626–0.640 for Rust.
This supports repeatability under the tested conditions; three short rounds do not
provide a precise noise estimate or confidence interval.

Fork live-throughput rounds are 3,119.9, **2,864.3** and 3,087.8 messages/s. The slower
second round is valid and retained. No unsupported noise explanation or favorable-run
selection is used. Capacity p99 compares different achieved loads; it is not a
same-offered-rate latency ranking.

## Matched sources and conditions

- Fork application: `8351224b56c9a3c6784ed5d7510857212133ac4c`.
- Upstream Go: `c4ab53d00b0a19a55962e03d18057ca6c3da9437`.
- Upstream Rust: `6dae2fd3c1a2dce61ca11beb731bcce2b42b67ad`.
- Both upstream Git HEADs were checked before the batch. No server source changed.
- Physical Fedora desktop: Linux `7.2.9-201.fc44.x86_64`, i9-13900KF,
  32 logical CPUs; `systemd-detect-virt` reports `none`.
- Identical fresh SQLite seed backups/files; seed SHA-256
  `691c52bfdc30b38bd8867e93b0914613df20e03d4cd7b00796ab68e0a5d26c9a`.
  Provider endpoints use closed loopback destinations.
- Server affinity `0,2,4,6`, client `8,10,12,14`, runner `16,18`;
  four workers/readers, three job workers, gzip and default GC settings.
- Three balanced Latin rounds per workload: upstream/fork/Rust,
  Rust/upstream/fork, fork/Rust/upstream. **27 valid runs**, nine per workload,
  with no extension. The desktop-local batch ran 12:54:41–13:06:16.
- No hooks, profilers, process-counter samples, or overlapping task-owned builds,
  tests or agents during timing. Cold route-contract preparation and login precede
  posting warmup. Task containers are removed afterward.

Fixed POST uses 64 workers, an active-pacing queue of 8,192, two warm seconds and
eight timed seconds at 3,000 offered/s. All arms use the same maintained receipt
client, binary SHA-256
`4f8848e2a52927660d6aa031fa38c16b7f7e7cd500cf1c8aee7e3f245b73b1f9`.
It consumes complete wire bodies before timestamping completion, then validates
content type, gzip integrity and decoded Turbo/client/sequence/message identity.
Validation and ACK-retention overhead are not assumed neutral.

Capacity uses the same shared load generator and response/ACK contracts: two warm
seconds with four clients, then eight timed seconds with sixteen. Cable uses the
same trace client in every arm: 64 distinct authenticated people, six subscriptions,
300 paced messages at 100 ms intervals, then five saturated seconds with four posters.

## Outcomes and retained evidence

**1,190,787 writes** pass their applicable exact audits:

| Workload | Warm + timed audited writes |
|---|---:|
| Fixed POST | 270,000 |
| POST capacity | 772,202 |
| Cable | 148,585 |

Fixed receipts are canonically matched to exact persisted IDs, body sequences,
client identities, creator, room and normalized FTS content. Capacity keeps the
actual ACK JSONL sidecars, not just aggregate counts. Cable verifies exact ACK IDs,
body markers, creator/room/FTS and complete delivery to every client; presence returns
to zero. Database integrity passes, with the seed's one FK orphan unchanged.

[`manifest.json`](manifest.json) indexes all raw samples, receipt/ACK sidecars,
audits, full hardware and executable identities, workload adapters and declared
budget. Cookies, secrets, databases, storage and request logs are excluded. The
initial `delivered_all` reporting-key error is corrected to `all_clients`; it did
not alter or invalidate the captured outcomes. `bin/check` passes on the unchanged
application before this evidence delivery.

These results supersede the [two-way current-Go record](../current-upstream-native-20261011/README.md)
for current rankings. Earlier Rust comparisons used an older fork and different
receipt tooling. This batch does not establish read-route rankings, external-provider
performance, exclusive physical-core/SMT isolation, or universal correctness.
Source cache budgets/defaults are preserved rather than artificially equalized.

The [independent native CPU capture](../native-perf-attribution-20261013/README.md)
and [rejected association projection](../display-projection-20261013/README.md)
follow this baseline. No performance implementation is promoted by this record.
