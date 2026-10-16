# Double-check: bare attachment nulls — 2026-10-16

An independent review of PR `a441e6f` caught a regression in bot attachment
selection: `?attachment` was flattened to an empty string and selected attachment-only
attributes, silently discarding nonempty raw message text. Rust treats the bare key
as null; `?attachment=` is a distinct, present empty string. A real bot-key regression
reproduced the lost body and missing FTS entry in the old Go handler.

Form intake now records scalar null metadata before `url.Values` loses it, and final
normalization applies query nulls with the same last-scalar/query-over-body precedence
as the values. Explicit values clear earlier nulls, including JSON body nulls. Bot
intake reuses the existing form parser; JSON syntax validation/raw bytes remain
unchanged. Ordinary value-only forms allocate no null metadata. Every normalizer
caller handles decoding errors. No SQL, cache budget, authority, signing, frontend,
media processing or postcommit-lifetime change is made; no speedup is claimed.

Verification:

- The permanent real-bot-key test fails the old source for bare query/form keys and
  overridden/repeated nulls. It checks creator identity, exact stored body and FTS
  membership, while explicit-empty controls remain distinct.
- A temporary extension of the maintained HTTP workflow creates the bot through the
  administrator endpoint and sends requests with no login cookie. Pinned Rust passes;
  the old Go executable loses the body. The corrected Go executable passes the same
  raw-body/FTS/source-precedence cases. Generated bot keys are redacted from traces.
- The final candidate passes `bin/check` and all fifteen existing Go/Rust HTTP/state
  workflows, **30/30**. The earlier double-check also passed all thirty on the broken
  head: those existing flows did not expose the missing bare-null distinction.
- The pristine shared Chromium smoke passes setup, two-tab live messaging/duplicate
  suppression/markup safety, permalinks, editing/search, profiles/accounts, QR, rooms,
  bots/styles, transfer/join and direct ping. The shared checkout had an unrelated
  three-line persisted-echo wait change; it was preserved, and a pristine `HEAD`
  export was used instead. Missing matching Playwright Chromium was installed locally.
- A fresh read-only review found no concrete issue in null ordering, current callers,
  multipart/nullable attributes or error/resource paths. Parent checks establish
  execution evidence; the review did not run tests.

Initial probe setup errors (cookie argument type, raw token versus the official
`id-token` bot key, and selecting a seeded room the new bot had not joined) were
corrected before valid comparisons. They are not product failures; private logs
are retained. The valid baseline report records Rust passing and old Go losing text.
Baseline local `0e65a86` has the runtime of published PR `a441e6f`. Compact proof,
reproduction source and hashes are retained here; runtime logs, tokens, generated
data/storage, complete rosters and binaries remain private. No full arbitrary Rack
parameter, browser-platform, media-output or whole-application correctness guarantee
is claimed. This double-check found and repaired a real gap; passing gates alone
had not established its absence.
