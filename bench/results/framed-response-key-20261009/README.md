# Length-framed private response identities

This cohort replaces generic JSON key construction with a concrete, length-framed
identity. It retains every existing input: captured authority/target, effective
URI and parsed form, authenticated user, cookie/accept/type/frame headers, browser
agent, Origin, X-Requested-With, coding and Git revision. Form names are sorted;
repeated values retain order. Nil/empty collections and arbitrary header bytes
cannot alias through framing or JSON's replacement of invalid UTF-8. The 8-KiB
serialized-key budget remains. Authentication and fresh generation checks do not
change; no identity, permission or epoch is cached by this change.

The previous build is `47cf073`. Source/binary hashes are in `metadata.json`, and
`candidate.patch` records the measured change. **All three upstreams advanced when
refreshed after timing:** Go to `c4ab53d`, Rust to `6dae2fd`, and shared verification
to `ec02deb`. These results reproduce the measured Go upstream `9b0ef06`, Rust
`9872c1d` and verification `e244051`; they do not establish acceptance or parity
against the newly advanced heads. The next integration must reconcile them and
rerun current contracts.

## Motivation and checks

The preceding complete-application mixed profile assigns 2.14 of 53.79 CPU seconds
to the response-hit JSON key call. The new helper removes temporary interface trees,
reflection and JSON escaping; it does not hash away distinct request variants.
A focused framing test protects map order, repeated values, collection boundaries,
embedded separators, raw header bytes and oversized-key bypass. Existing response
variant, foreign-write and live-authority tests remain unchanged.

Full native `bin/check`, repeated focused races, Linux vet/web races, fresh native
and Linux builds, browser flows and actual public HTTP/write checks pass against the
measured sources. The public checks include seven preflights, twenty responses,
four exact acknowledged ID/body/room/FTS audits and forwarding-disabled aliases.
The browser retains its disclosed persisted-echo wait before the unchanged exact
three-message assertion. No assertions or independent SQL oracles were weakened.

A temporary Apple M3 Max Go 1.27.1 benchmem probe uses one representative fixed
search/cookie/header input. Five repetitions record JSON at 2,106–2,151 ns/op,
1,802 B/op and 36 allocations; framing at 332–337 ns/op, 472 B/op and two allocations.
The source is archived as `probe_test.go.txt`, not added to the permanent test
suite. This isolated result does not establish whole-application allocation or
capacity improvements.

## Matched application measurements

Three sequential interleaved rounds use eight timed seconds, two excluded warmup
seconds, sixteen clients and one viewer. Four runtime workers/readers and three
workers per job queue; server CPUs 0–3, client CPUs 4–7. Completed/fragment budgets
are 64/32 MiB plus inherited gzip/public 32/64 MiB. Forward/reverse three-build
order keeps the candidate in the middle. No builds, tests, profilers or agents
run alongside timing. Complete responses and exact acknowledged writes use the
shared independent SQLite oracle; the only temporal adapter permits `go-before`.

| Warm route | Previous median req/s | Candidate median req/s | Change | Measured Rust median req/s |
|---|---:|---:|---:|---:|
| Room | 27,547 | 30,119 | +9.3% | 41,247 |
| Messages | 29,096 | 30,459 | +4.7% | 40,018 |
| Sidebar | 32,042 | 34,236 | +6.8% | 44,722 |
| Search | 30,800 | 33,613 | +9.1% | 44,822 |
| Post | 4,433 | 4,410 | −0.5% | 5,800 |

Every room/message/search pair improves. Sidebar regresses in the last pair;
posting improves two pairs but has a slower second candidate round and a slightly
lower median. Posting does not execute the changed key path. Go's warm-read
medians are 73–77% of the measured Rust population, and posting is 76%; neither
is a result for the newly advanced Rust revision.

| Reads with ten paced HQ writes/sec | Previous median req/s | Candidate median req/s | Change |
|---|---:|---:|---:|
| Room | 30,700 | 32,012 | +4.3% |
| Messages | 31,281 | 31,705 | +1.4% |
| Sidebar | 36,723 | 37,132 | +1.1% |
| Search | 35,900 | 34,325 | −4.4% |

Every mixed room pair improves; the first message/sidebar/search pairs regress.
The previous second sidebar round slows to 20,871 req/s. Rust has much larger
first-round slowdowns: room 7,463, messages 2,725 and sidebar 14,283 req/s; later
rounds reach 34,415–37,236, 22,349–33,116 and 39,999–42,932. All samples remain.
These unstable Rust medians cannot establish a Go/Rust capacity ordering or parity.
HQ writes are not same-room Watercooler fanout. Uncached runs are not repeated:
with completed-response caching disabled, this key path does not execute.

Warm runs validate 10,360,992 timed responses and audit 428,969 writes including
warmup. Mixed runs validate 9,028,424 timed reads, 2,874 timed writes and 3,594 exact
acknowledged writes including warmup. Errors and invalid responses are zero; all
write/FTS audits pass. Peak generator CPU is 53.4% warm and 62.0% mixed out of 400%
available, not proof of an unconstrained generator in every workload.

## Diagnostics and limits

A separate two-round five-second warm diagnostic records 53.11 CPU seconds over
31.56 wall seconds. Response identity construction uses 0.56 seconds (1.05%);
response-hit work including fresh observation uses 1.44 seconds (2.71%). Generic
JSON marshaling is absent from this path. The preceding mixed profile has different
requests/parallelism, so cumulative percentages are not normalized cost ratios.
Cookie verification, fresh observation, network IO and allocation remain material.
There is no complete-application allocation snapshot or matched-rate latency claim.

Compressed bundles retain every summary, raw sample, contract and audit result.
Runtime databases/ack files are consumed before normal harness cleanup; verified
counts/results remain. Profiles, listings and gate logs are included. The original
helper probe is included for reproducing isolated benchmem measurements.

These are ARM64 OrbStack/macOS measurements, not upstream's published x86 machine.
Guest affinity is not physical isolation; host/workload variance is substantial.
Tmpfs verifies transactions/FTS, not NVMe or crash durability. One viewer does not
establish multi-user, eviction or fanout capacity; throughput-run tails are not
matched-offered-rate latency. Native Intel is unavailable, the CLI forced-exit path
remains unexercised and unchanged media checks are not rerun. Aggregate resource
work and advanced-upstream integration remain unfinished; global optimality is
not claimed.
