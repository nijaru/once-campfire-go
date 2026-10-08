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
bin/check-parity --seed /path/to/parity-seed --go-binary ./campfire \
  --rust-binary /path/to/rust/campfire --out .cache/contracts
node bin/check-browser.mjs
bin/check-screens --out .cache/screens --only '**'
bin/check-acme
bin/check-container --image once-campfire-go:verification
```

`bin/check-parity` compares authentication, direct-ping reuse, pagination, sidebar documents,
message creation/editing/deletion, room conversion/membership revocation and partial updates,
account restrictions/custom styles, retained inactive direct-room participants, involvement pages
and nullable updates, direct uploads, signed blob/disk/representation downloads, multipart attachment
replacement/purge, avatar assignment/deletion and nested direct-upload metadata against pinned Rust. Supply a prepared parity seed and
binaries built for the same host. Each workflow uses a fresh SQLite backup, including WAL state, and a copy
of seeded storage. It checks HTTP behavior, controls, authorization, fresh reads and committed
message/FTS state, not byte-identical HTML or internal implementation details. Reports include
binary, tool and seed-snapshot hashes; failures remain visible and produce a nonzero exit status.
The [upstream integration record](bench/results/upstream-integration-20261007/README.md) covers
cache corrections and focused sidebar verification; it also records the unresolved full-browser
smoke failures in the earlier local harness. The current browser command uses the shared flow
rather than maintaining a divergent copy. The [read-path record](bench/results/read-path-20261008/README.md)
documents current checks and a shared-flow synchronization defect. A local patch is available
for review; it has not been submitted. Shared flows are functional checks, not complete parity.
The paired HTTP tool does not cover malformed Rack parameter compatibility, multipart metadata,
JavaScript, live Cable delivery or crash durability.

The browser command starts a fresh disposable Go instance and runs the
[shared Campfire browser flow](https://github.com/basecamp/once-campfire-verification), the same
flow used for Rails, Elixir and Rust. Clone that harness alongside this repository, run `npm ci`
and `npx playwright install chromium` there, then run `node bin/check-browser.mjs` here.
`VERIFICATION_ROOT` selects another harness checkout; `GO_BINARY` selects another Go binary.
Use `--turbo-cancellation` only for the focused regression when changing the port's Turbo override.
The upgrade and screen tools use the Rust checkout's parity seed and installed Playwright. Screen
comparison runs both binaries on disposable copies of each seed, recording their hashes and keeping
the original parity masks. Reports distinguish screenshots, accessibility, server/live DOM, network
responses and Cable traffic. There are no Go-specific allowlists hiding failures.

Package tests use temporary databases and do not silently skip integration tests for a missing seed.
They include Rails signing/encryption and serialization vectors, 658 rich-text cases, 385 user-agent
cases, QR codes, route recognition, storage/ranges, transactions, access control, Cable, jobs,
90 Open Graph cases, 19 webhook cases and Web Push encryption/local delivery. Media metadata tests
run locally. Changes to media processing also use the reference output checks, selected with the
`media_vectors` build tag. The checked-in video goldens require AMD64 FFmpeg 7.1.5; ARM64 FFmpeg with the same
version produces different JPEG bytes. Architecture is part of the toolchain identity:

```sh
docker build --platform linux/amd64 --target toolchain -t once-campfire-go:toolchain .
mkdir -p .cache/tmp .cache/docker-go-build .cache/docker-go-mod
docker run --rm --platform linux/amd64 --user "$(id -u):$(id -g)" -v "$PWD:/src" -w /src \
  -e GOCACHE=/src/.cache/docker-go-build -e GOMODCACHE=/src/.cache/docker-go-mod \
  -e TMPDIR=/src/.cache/tmp once-campfire-go:toolchain \
  go test -race -tags 'sqlite_fts5 media_vectors' ./internal/storage
```

`CAMPFIRE_STORAGE_VECTORS` can explicitly select a separate Rails capture for another architecture;
the checked-in vectors remain the default. Originals, derivatives, attributes and metadata are
checked without changing encoder flags or weakening byte comparisons. The
[media verification record](bench/results/acceptance-corrections-20261007/README.md) distinguishes
unchanged-golden verification from the separate ARM64 Rails comparison.

`bin/build` and `bin/check` use mise when Go is absent from PATH and keep temporary build files in
`.cache/`. Live ACME is tested against a local Pebble CA, including restart with the CA offline.
Container verification exercises setup, a live SQLite backup, offline restore, and restart.

## Benchmarks

Upstream's published measurements use 16 concurrent clients on an AMD Ryzen AI MAX+ 395
with 32 GB RAM and four hardware cores allocated to each app. These are not measurements of
this fork.

| HTTP workload (requests/sec) | Rails | [Django](https://github.com/basecamp/once-campfire-django) | [Laravel](https://github.com/basecamp/once-campfire-laravel) | [Express](https://github.com/basecamp/once-campfire-express) | [Elixir](https://github.com/basecamp/once-campfire-elixir) | [Go](https://github.com/basecamp/once-campfire-go) | [Rust](https://github.com/basecamp/once-campfire-rust) | [C](https://github.com/basecamp/once-campfire-c) |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Room page | 710 | 414 | 1,696 | 42,481 | 1,126 | 52,512 | 105,909 | 137,505 |
| Messages page | 1,113 | 454 | 1,890 | 74,779 | 1,407 | 54,100 | 103,301 | 142,669 |
| Sidebar | 1,901 | 576 | 3,364 | 94,460 | 3,621 | 58,714 | 120,930 | 151,001 |
| Search | 1,332 | 549 | 2,615 | 83,493 | 2,127 | 60,444 | 121,502 | 148,766 |
| Post a message | 226 | 113 | 567 | 2,121 | 1,392 | 9,000 | 8,004 | 7,486 |

[Shared verification](https://github.com/basecamp/once-campfire-verification) · [Detailed results](https://github.com/basecamp/once-campfire-verification/blob/main/docs/performance-review.md).

See [`bench/`](bench/) for tooling. The [initial Go/Rust baseline](bench/results/current-baseline-20261007/README.md)
and [sidebar hydration comparison](bench/results/sidebar-hydration-20261007/README.md) use identical
fixtures and full-response/write validation on an ARM64 Linux VM. The [reference scan comparison](bench/results/scan-record-20261007/README.md)
reduces allocations and improves measured room throughput. The
[cache-hit setup comparison](bench/results/read-path-20261008/README.md) removes repeated route
recognition and SQLite statement preparation, improving read medians by 5.5–9.4%; writes do not
improve. It also records a rejected query-fusion experiment rather than hiding its regressions.
Earlier comparisons still favor Rust on reads. These local measurements differ from the upstream
table in machine, revisions and methodology.

## Known differences

- Bundled Turbo 8.0.13 awaits its asynchronous response delegates so cancelled frame-body reads
  reach the existing abort handler; other errors still propagate. The attributed override is in
  `assets/overrides/turbo.js`. Tests require it to match the pinned library with only two added
  `await`s and a license notice. This is an intentional frontend fix, not exact JavaScript parity.

- Session-transfer auto-submit forms explicitly close their form tag; the pinned Rails
  reference omitted it.
- Background sidebar refreshes preserve an open New Ping form and selected recipients.

- Sidebar connection refresh waits for the current Turbo frame to finish loading,
  preventing an aborted response on startup or reconnect. Obsolete connections and removed frames do not reload.

- Templates use `html/template`. Whitespace, attribute serialization, some canonical form-action
  URLs, some document titles, and response headers/validators differ from Rust. Strict server/live DOM and network layers
  therefore still fail in many inventory cells, even when screenshots, accessibility and workflows
  match. These failures remain visible in the validation report. Exact protocol parity for malformed
  parameters and every content-negotiation edge case is not claimed.
- WebSockets share serialized and compressed broadcast payloads through a small extension to
  coder/websocket v1.8.15 (see `third_party/websocket/README.campfire`). Outgoing queues hold 256
  frames; slow clients are disconnected. Authorization is checked afresh for each publication,
  batching distinct sessions per room. Rust uses different stream queues.
- Go ignores typing commands for rooms that have been deleted; Rust can still echo them to an
  already subscribed socket. The composer shows the same deleted-room message.
- Empty involvement updates preserve SQL NULL, as in the reference. Go reads that nullable value
  without a scan error; Rust can raise a nil-inquiry error on a subsequent involvement request.
- The response cache uses least-recently-used eviction instead of Rust's sampled eviction. The Go
  message-fragment cache is also independently implemented. It retains versioned message lists
  and owned sidebar frame/layout parts; current membership, permission and layout observations
  are read before cache lookup.
  Room pages also cache their surrounding HTML keyed by fresh page data, inserting freshly selected
  messages and their queried refresh cursor. Search layouts similarly surround a captured immutable
  result body; misses rerun the scoped full query. Complete GET gzip representations have a bounded
  memo keyed by immutable part identities, not by wire ETags. Authorization, sessions and request
  observations stay fresh. Responses derive validators from part lengths and hashes, so ETag values
  differ from both the original Go implementation and Rust. Message pagination uses its
  complete rendered body for ETags, including related-user and boost changes. It omits
  Last-Modified because message timestamps cannot describe external presentation edits.
- Completed room, messages, sidebar and search HTML/gzip responses share a per-process
  cache (`CAMPFIRE_RESPONSE_CACHE_MB`, default 64 MiB, 0 disables it). Every request checks
  its session and access before lookup. A dedicated SQLite reader observes local and foreign
  commits and invalidates whole responses. Message and message-list fragments also carry
  the observed generation and request origin, including external edits without timestamp
  changes and host-dependent rich-text filtering. Request
  variants remain separate; headers and cookies
  stay fresh. GET and HEAD share bodies and preserve validators; flash responses bypass it.
- The default version label and fallback VAPID subject identify `once-campfire-go`. Explicit version,
  VAPID keys and subject settings remain supported.
- Storage keys containing separators, NUL, or parent-directory shards are rejected before
  filesystem access. Valid keys retain the existing storage layout.
- Media bytes depend on toolchain architecture as well as library versions. Unchanged goldens pass
  with AMD64 FFmpeg 7.1.5 and ARM64 libvips 8.16.1 in the recorded mixed-architecture environment.
  Go and Rust also match a separate native ARM64 Rails capture. A matching version string alone
  does not guarantee matching bytes. Web Push is verified locally, not against external push providers.

As in Rust, background queues are bounded and in-process: graceful shutdown drains work, but a
process crash can lose queued jobs. See [the implementation record](plans/go-conversion.md) for
coverage and validation scope.

## License

MIT; see [MIT-LICENSE](MIT-LICENSE). Third-party notices for copied/adapted algorithms are in
[licenses](licenses/) and [internal/html/LICENSE](internal/html/LICENSE); module dependencies retain
their own licenses.
