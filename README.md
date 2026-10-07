# once-campfire-go

Campfire in Go, ported from [ONCE Campfire in Rust](https://github.com/basecamp/once-campfire-rust).
It uses the Go standard library for routing, HTTP, templates, SQL access, cryptography and process
lifecycle. There is no web framework, ORM, dependency injection container or frontend framework added
by the port. The original Turbo/Stimulus/Lexxy frontend is retained.

The application includes setup and invitations, sign-in and session transfers, account and user
administration, open/closed/direct rooms, messages and attachments, editing/deletion, boosts,
mentions, search, notifications, bot APIs/webhooks, link previews, Web Push, and Action Cable
presence, typing, read/unread and live updates. It includes media processing, the public HTTP/TLS
front server, automatic certificates, production packaging and backup/restore hooks.

SQLite schema, storage keys/layout, password hashes and Rails signed/encrypted cookies are preserved.
An upgrade test checks Go and Rust cookies in both directions and Rust reading/searching Go-written
messages. The port is functionally implemented; **strict HTML/network parity is not complete**.
See [validation](plans/validation.md) and the known differences below before replacing an installation.
The pinned Rust source is in `reference/`, with its Rails source in `reference/reference/`.

## Dependencies

- Go 1.27.1: `net/http`, `html/template`, `database/sql`, `crypto`, `encoding/json`, `embed`, `testing`.
- `github.com/mattn/go-sqlite3`: SQLite with FTS5, through CGO.
- `github.com/coder/websocket`: WebSocket transport; the Action Cable protocol and channels are local code.
- `golang.org/x/crypto`: bcrypt and ACME. The pinned revision includes the upstream ACME missing-Location fix.
- `golang.org/x/net`: HTTP/2 and HTML tokenization; a local tree-builder fork matches the reference parser.
- `golang.org/x/text`: indirect Unicode support.
- Native libvips, ffmpeg/ffprobe and libzstd: image/video processing and public HTTP zstd compression.

A local build needs a C compiler, pkg-config, libvips and libzstd development headers, ffmpeg, and
Python 3 for asset generation. Assets are built from the pinned sources without Ruby, Rust or Node.
The Dockerfile pins the reference media toolchain: libvips 8.16.1 and ffmpeg 7.1.5.

## Run

```sh
git submodule update --init --recursive
bin/build
export SECRET_KEY_BASE="$(openssl rand -hex 64)"
DISABLE_SSL=1 HTTP_PORT=8080 ./campfire server
```

Open <http://localhost:8080/first_run>. Keep the same `SECRET_KEY_BASE` across restarts; use an
existing installation's secret to retain logins when migrating its database and files.

```sh
docker build -t once-campfire-go .
docker run --rm -p 8080:80 -e DISABLE_SSL=1 -e SECRET_KEY_BASE \
  -v campfire-go-storage:/rails/storage once-campfire-go
```

The container runs as uid/gid 1000. For a bind mount, make its storage writable by that user.
`CAMPFIRE_STORAGE_PATH` defaults to `storage`, with databases in `db/`, media in `files/`, backups
in `backups/`, and certificate cache in `thruster/`. `CAMPFIRE_DATABASE_PATH` and `CAMPFIRE_FILES_PATH`
override individual locations; `RAILS_ENV` defaults to `production`. `campfire db:prepare` initializes
an empty database and checks migration versions; existing databases missing migrations are rejected.

The public listener uses `HTTP_PORT=80`. Set `TLS_DOMAIN` for automatic ACME certificates and HTTPS
on `HTTPS_PORT=443`. The internal application listener defaults to `TARGET_BIND=127.0.0.1` and
`TARGET_PORT=3000`. The front server provides HTTP/2, optional H2C, gzip/zstd with compression jitter,
a bounded response cache and graceful shutdown. `campfire backup` writes an atomic SQLite snapshot;
the image provides ONCE's `/hooks/pre-backup` and `/hooks/post-restore` hooks.

## Validation

```sh
bin/check                              # gofmt, assets, vet, all tests with the race detector
bin/check-assets                       # asset digests/importmap
bin/check-upgrade --rust-root ../once-campfire-rust
node bin/check-browser.mjs
bin/check-screens --out .cache/screens --only '**'
bin/check-acme
bin/check-container --image once-campfire-go:verification
```

The browser/upgrade tools use the Rust checkout's parity seed and installed Playwright. Screen
comparison runs both binaries on disposable copies of each seed, recording their hashes and keeping
the original parity masks. Reports distinguish screenshots, accessibility, server/live DOM, network
responses and Cable traffic. There are no Go-specific allowlists hiding failures.

Package tests use temporary databases and do not silently skip integration tests for a missing seed.
They include Rails signing/encryption and serialization vectors, 658 rich-text cases, 385 user-agent
cases, QR codes, route recognition, storage/ranges, transactions, access control, Cable, jobs,
90 Open Graph cases, 19 webhook cases and Web Push encryption/local delivery. Media metadata tests
run locally; byte-for-byte output tests require the pinned container toolchain:

```sh
docker build --target toolchain -t once-campfire-go:toolchain .
mkdir -p .cache/tmp .cache/docker-go-build .cache/docker-go-mod
docker run --rm --user "$(id -u):$(id -g)" -v "$PWD:/src" -w /src \
  -e GOCACHE=/src/.cache/docker-go-build -e GOMODCACHE=/src/.cache/docker-go-mod \
  -e TMPDIR=/src/.cache/tmp once-campfire-go:toolchain \
  go test -tags 'sqlite_fts5 media_vectors' ./internal/storage
```

`bin/build` and `bin/check` use mise when Go is absent from PATH and keep temporary build files in
`.cache/`. Live ACME is tested against a local Pebble CA, including restart with the CA offline.
Container verification exercises setup, a live SQLite backup, offline restore, and restart.

## Benchmarks

The [competing-PR comparison](bench/results/pr-comparison-20261007/README.md) measures all six
open Go alternatives, this fork and pinned Rust on the same ARM64 Linux VM. The retained,
template-derived renderer roughly doubles fragment-disabled read capacity and improves posts;
complete Go response bytes remain unchanged. Warm reads are mixed, and Cable results include
regressions rather than a claimed capacity win. Matched-rate HTTP tails expose material generator/
VM lateness; Rust-relative latency targets remain unmet. Workload-matched core and paced 1,000-client
Cable memory stay roughly unchanged. The report records exact sources, runtime differences,
raw measurements, correctness boundaries and deliberately rejected migrations.

The [native Intel follow-up](bench/results/native-final-20261006/README.md) uses distinct physical
P-cores, gzip, 16 clients and repeated 30-second samples. Go's five warm read workloads have
9–44% higher throughput than pinned Rust, but worse p99 and higher end-of-suite memory;
posts are 34% slower. Focused edit/boost views improve native throughput 21–60% over the
preceding Go build. All 72 samples have zero errors; 921,603 writes are verified in messages/FTS.
These screen-specific results are not universal parity or the published AMD comparison.

The [remaining render trials](bench/results/remaining-render-20261006/README.md) reject timestamp/
avatar preparation after inconsistent application results. Only two constant-regex hoists remain,
with lower scalar cost but no claimed HTTP speedup. The evaluated candidates reached diminishing
returns without a new dependency or broader template/database redesign.

The [message-form trials](bench/results/message-forms-20261006/README.md) remove unused
full-message work from edit and boost pages. Matched OrbStack gzip throughput improves 25–71%
on those four routes, with exact Go form-byte checks. Rust remains faster; identity tails are
mixed. These complete-document tests do not measure Turbo-Frame clicks or general read/write gains.

The [rich-text/render trials](bench/results/render-work-20261006/README.md) remove redundant
tree/link work, prepare fixed reaction contents and reuse fixed escaping/signing primitives.
Each isolated cold-render phase lowers CPU and read-route p99. The final matched six-route
OrbStack comparison improves posts about 5–6%; warm reads remain broadly unchanged or mixed.
Rust still leads room/writes and generally tails/memory. All 658 rich-text cases pass;
270 HTTP samples have zero errors and 746,570 acknowledged writes persisted and matched FTS.
These are VM results, not native or universal parity; no new library was justified.

The [Cable routing trials](bench/results/cable-routing-20261006/README.md) replace all-client
subscription scans with dense, indexed recipient lists while retaining fresh authorization.
At 10,000 clients, longer OrbStack trials improve uncompressed complete fan-out throughput 39%
and compressed throughput 3%. Paced delivery improves, but saturated p99 worsens; the report
records that trade-off, high-client workload-scoped memory, fresh profiles and validation.
No native or current cross-language Cable comparison is available.

The [search-layout/database trials](bench/results/layout-database-20261006/README.md)
retain immutable search layouts around freshly selected results. In the final six-route OrbStack
VM comparison at 16 clients, gzip search measures 26,561 Go versus 26,734 Rust requests/sec;
active search 16,591 versus 14,083. Rust still leads room (28,307 versus 22,040), writes
(3,183 versus 2,271), usually tails, and process memory. These warm-cache VM results do not
establish native or general parity. Reader-cache replacement was rejected; covering-index
write amplification and insufficient checkpoint evidence did not justify database changes.

The preceding [search/hydration trials](bench/results/search-hydration-20261006/README.md)
reuse immutable message lists after fresh matching-reference queries and share metadata reads
across fragment misses. Fragment-disabled CPU falls about 7–15%; its p99 and gzip process
memory do not consistently improve. Both reports retain raw trials, rejected candidates and checks.

The preceding [matched Go/Rust VM comparison](bench/results/rust-go-boundary-20261006/README.md)
covers room/history/search/writes with both encodings and 1/16/64 clients. Go has higher warm
active-room/history throughput; Rust leads room/search/writes and generally latency and memory.
It records a corrected search-count regression, rejected SQLite crossing trial, response-size
differences and substantial VM limitations. The native follow-up above has a narrower gzip/16-client
scope; fresh Ruby/Elixir measurements remain unavailable. Published AMD ratios are not transferable.

This fork fixes the room refresh cursor and removes repeated compression, hashing and
rendering preparation while keeping authorization, queries, headers and cookies fresh.
The [request-work comparison](bench/results/request-work-20261005/README.md) measured
22,152 gzip and 15,319 identity room requests/sec in a Linux/arm64 VM: 16% and 15% over the
preceding direct-record build. Search/static performance was essentially unchanged. The report
includes rejected candidates, raw samples, checks, memory costs and reproduction commands.
These are not native AMD or Go/Rust comparisons, and reduced allocation did not consistently
reduce resident memory. A [final ownership cleanup](bench/results/ownership-20261005/README.md)
removes duplicate retained message-list payloads without claiming a broad throughput or
process-memory reduction. [Timestamp parsing](bench/results/timestamp-20261005/README.md)
and [default PGO](bench/results/pgo-20261005/README.md) trials were not justified by their
application results; the latter report explains the remaining costs and stopping point.

Earlier same-VM iterations have separate pinned comparisons:
[prepared gzip](bench/results/prepared-gzip-20261005/README.md),
[part identities](bench/results/part-identity-20261005/README.md),
[static shell parts](bench/results/static-shell-20261005/README.md),
[binary version keys](bench/results/binary-keys-20261005/README.md), and
[direct record input](bench/results/record-input-20261005/README.md).
Do not multiply ratios from these separately timed experiments.

The [upstream comparison](bench/results/optimization-next-20261003/README.md) measures the
initial Go version, optimized upstream Go, and Rust in three rotating runs. Median
requests/sec at 16 HTTP clients, with identical seed data and four application CPUs:

| Workload | Initial Go | Optimized upstream Go | Change | Rust |
|---|---:|---:|---:|---:|
| Room page | 10,175 | 15,218 | +49.6% | 27,535 |
| Message history | 21,445 | 21,369 | -0.4% | 31,139 |
| Sidebar | 24,423 | 24,658 | +1.0% | 38,303 |
| Search | 14,817 | 14,841 | +0.2% | 30,807 |
| Post message | 4,025 | 5,036 | +25.1% | 7,740 |

The historical sidebar row is not an equivalent Go/Rust workload: that Go build returned a bare
frame where the reference wrapped a layout. The handler now renders the application or Turbo-Frame
layout, but no full-page sidebar performance comparison is established by those historical numbers.

Room-page p99 latency fell from 5.69 to 4.14 ms; message-write p99 fell from 16.03 to
12.49 ms. In that comparison, Rust was 1.81× faster on room pages and 1.54× faster on writes. All nine
application runs completed with zero HTTP errors, 345,913 acknowledged writes verified
in both messages and FTS, and nine thumbnails with identical bytes. HTTP memory use was
essentially unchanged. See the [raw report](bench/results/optimization-next-20261003/application/report.md)
for ranges, latency, resource measurements and limitations.

That upstream pass did not remeasure Cable throughput. In the
[earlier full-workload comparison](bench/results/application-optimized-20261003/report.md),
compressed broadcasts to 10,000 clients measured 21.1 complete messages/sec for Go and
39.3 for Rust. Final workload Pss was 1,023 MiB versus 408 MiB. Those measurements include
large WebSocket workloads and must not be compared directly with the upstream HTTP-only
memory figures. The earlier run had zero HTTP errors and complete Cable delivery;
[interrupted attempts](bench/results/application-optimized-20261003/CONTENTION.md) were
excluded and restarted. The [first optimization report](bench/results/optimization-20261003/comparison.md)
retains the original-port comparison.

These numbers compare these implementations on this workstation, not languages in general.

```sh
# Build both release binaries and Rust's bench/loadgen; prepare its parity seed.
bin/build
bench/application --out bench/results/my-run --reps 3 --seconds 5 \
  --concurrency 1 16 64 --cable-clients 100 1000 10000 --deflate 0 1
```

The harness alternates applications, uses fresh identical seed copies, fixes server/client CPU
sets and four application workers, and warms each HTTP workload. It validates message/room IDs,
static/avatar bytes, every successful write and FTS entry, complete Cable fan-out, and actual thumbnail
bytes. Reports include raw samples, source/binary hashes, toolchains, load averages and limitations.
HTTP measurements use the direct application listener and identity encoding; public TLS/compression
throughput is not measured in that upstream comparison. The fork harness also supports
`--listener public --gzip 1`, named candidates, concurrent active-room/search writes,
and `--fragment-cache-mb 0` for fragment-disabled rendering. Build `go build -o httprate ./bench/httprate`
and pass `--rate-loadgen ./httprate --http-rates 1000 --concurrency 16` to replace closed-loop capacity
samples with scheduled-arrival latency at a fixed offered rate. Select a rate each application can
sustain; inspect generator lateness, queueing, errors and drain time rather than treating it as
server-only latency. Keep fixed-rate and maximum-capacity comparisons separate.
`bench/health` remains available for the much narrower health-handler test.

## Known differences

- Templates use `html/template`. Whitespace, attribute serialization, some canonical form-action
  URLs, and response headers/validators differ from Rust. Strict server/live DOM and network layers
  therefore still fail in many inventory cells, even when screenshots, accessibility and workflows
  match. These failures remain visible in the validation report. Exact protocol parity for malformed
  parameters and every content-negotiation edge case is not claimed.
- WebSockets share serialized and compressed broadcast payloads through a small extension to
  coder/websocket v1.8.15 (see `third_party/websocket/README.campfire`). Outgoing queues hold 256
  frames; slow clients are disconnected. Authorization is checked afresh for each publication,
  batching distinct sessions per room. Rust uses different stream queues.
- Go ignores typing commands for rooms that have been deleted; Rust can still echo them to an
  already subscribed socket. The composer shows the same deleted-room message.
- The response cache uses least-recently-used eviction instead of Rust's sampled eviction. The Go
  message-fragment cache is also independently implemented. It retains versioned message lists
  and sidebar HTML; current membership and permission data are read before cache lookup.
  Room pages also cache their surrounding HTML keyed by fresh page data, inserting the current
  messages and the queried room's update timestamp on every request. Responses assemble cached
  message bytes with fresh page HTML and derive validators from part lengths and hashes, so ETag
  values differ from both the original Go implementation and Rust.
- Completed GET gzip bodies have a separate 32 MiB compression memo keyed by ordered part
  lengths, constructor-computed SHA-256 digests and gzip mtime, not weak ETags. Requests still run authentication, authorization and page queries;
  headers/cookies are never reused. No-store, writes, streams, bodies over 1 MiB and large compressed
  entries bypass retention. Zstd and streaming gzip retain their existing compression paths.
- The default version label and fallback VAPID subject identify `once-campfire-go`. Explicit version,
  VAPID keys and subject settings remain supported.
- Storage keys containing separators, NUL, or parent-directory shards are rejected before
  filesystem access. Rust directly joins the key's shards; valid keys retain the same storage layout.
- Native host media output can differ with installed library versions. All byte-golden media tests
  pass with the pinned container libraries. Web Push is verified locally, not against external push
  providers.

As in Rust, background queues are bounded and in-process: graceful shutdown drains work, but a
process crash can lose queued jobs. See [the implementation record](plans/go-conversion.md) for
coverage and validation scope.

## License

MIT; see [MIT-LICENSE](MIT-LICENSE). Third-party notices for copied/adapted algorithms are in
[licenses](licenses/) and [internal/html/LICENSE](internal/html/LICENSE); module dependencies retain
their own licenses.
