# Functional contracts and cursor-validation cost

Compared five HTTP workflows against Rust `64f8635` using isolated copies of the same parity
seed. Each copy uses SQLite's backup API, including WAL state, and the complete seeded storage
tree. Redirects are not followed automatically; actors retain all cookies. The checker compares
statuses, destinations, usable controls, authorization, fresh content and committed message,
rich-text and FTS state—not HTML serialization, generated tokens or implementation details.

Before the changes, Rust passed all five workflows. Go failed three:

- A signed-in user could not open the login form or sign in as another user.
- Selecting a nonexistent user alongside Kevin created a duplicate David/Kevin ping.
- A pagination cursor belonging to another room returned 204 instead of 404.

After the fixes, both implementations pass authentication/account switching/logout, direct-ping
reuse, pagination, ordinary/framed sidebar documents and message create/edit/search/delete.
The Go regressions failed before the fixes. Full macOS and Linux CGO vet/race checks and the
WebSocket-fork checks pass. `before.json.gz`, `after.json.gz` and `checks/` retain the results;
`metadata.json` records source, binary, tool and environment identities.

Run the functional checker with a prepared reference seed and same-host binaries:

```sh
bin/check-parity --seed .cache/parity-seeds/default \
  --go-binary .cache/correctness-current/campfire-go-linux-after \
  --rust-binary .cache/rust-linux-target/release/campfire \
  --out .cache/contracts
```

This is not full parity. Uploads, membership revocation, live Cable delivery, cancellation races
and crash durability are not covered by these five workflows.

## Read-performance check

A local OrbStack Linux ARM64 check used three alternating-order repetitions, gzip, 16 clients,
four workers, guest server CPUs 0–3 and generator CPUs 4–7. Each sample followed a one-second
warmup and ran for ten seconds. No builds, tests or agents ran during timing. Complete decoded
40-message responses matched after replacing only the known listener origin; every timed
response was consumed and its encoded length and 200 status checked. All 2,566,753 timed
requests succeeded. Logs were kept in container `/tmp` and compressed after timing.

| Read workload | Before req/s | After req/s | Change |
| --- | ---: | ---: | ---: |
| Latest 40 messages | 19,730.3 | 21,877.1 | +10.9% |
| 40 messages before a valid cursor | 22,367.8 | 21,449.6 | −4.1% |

These short VM trials do **not** establish a performance improvement. The cursor fix adds a
scoped lookup, and its observed median cost is about 4%. The unrelated latest-page control
moved by 11%, showing substantial noise. Saturated latency is not equal-offered-rate latency;
no memory, direct-ping performance or native-host claim is made. Raw samples are in
`history-raw.json.gz`; `history-perf.py` is the exact container script used. An aborted initial
attempt omitted listener-origin normalization and is excluded from these paired results.

## Browser verification gap

The browser smoke now waits for the message channel, completed saves and acknowledged message
DOM updates, and checks unique message IDs. It does not suppress console or page errors.
Current-head Chromium smoke passed once, then failed all three repeat runs with an uncaught
`AbortError` despite completing the functional assertions. This is not a passing browser gate.

A minimal setup/join/direct-ping reproduction also fails in Go and passes in Rust. CDP identifies
`FrameController.requestSucceededWithResponse` in bundled Turbo 8.0.13. `FetchRequest.receive`
does not await that asynchronous callback, so a cancelled cloned response-body read escapes the
existing abort handler. `checks/browser-exception.log.gz` retains that diagnostic. The frontend
cancellation issue was unfixed at this checkpoint; these backend results do not establish browser
parity. A later [frontend fix and verification](../turbo-cancellation-20261007/) records the
resolution separately.
