# Turbo frame-body cancellation

Fixed the browser cancellation failure left visible by the earlier
[backend-contract checks](../functional-contracts-20261007/).

Turbo 8.0.13's `FetchRequest.receive` called its asynchronous success and failure delegates
without awaiting them. A frame delegate reads the cloned response body after headers arrive.
Cancelling during that read rejected a detached promise, outside `perform`'s existing abort
handler. The fix adds two `await`s, retaining the existing handling of cancellation and other
errors.

`assets/overrides/turbo.js` contains the pinned library, those two changes and its MIT notice.
Reference files remain untouched. This retains a 185 KiB library copy rather than adding a
patch processor or runtime monkeypatch. `assets.TestTurboCancellationOverride` requires exact
agreement with the pinned source plus the two changes and license; a reference update requires
reconsidering the override rather than silently retaining an outdated library.

## Verification

The browser smoke now begins with a deterministic regression using the application's served
Turbo asset and a local HTTP fixture. For both 200 and 422 responses, it sends headers and an
unfinished body, waits for the browser to receive the headers, then navigates the frame to a
replacement. Both body reads must actually be cancelled, the replacement must render, and no
page exception may escape. Separate asynchronous delegate failures must still reach the error
handler, reject the request and finish its lifecycle.

The fixture explicitly checks Chromium's two resource diagnostics for its deliberately injected
422 responses. It accepts no other fixture errors. The application's console/page error checks
are unchanged and suppress nothing.

- The final regression fails against the original binary with `Cancelled 200 frame body:
  AbortError`, before running application workflows.
- The candidate passes the regression and the full Chromium smoke once, then in three consecutive
  repeat runs. This covers setup, live two-tab messaging, duplicate suppression, markup safety,
  editing, search, profiles, account settings, room updates, bots, styles, transfers, joining and
  direct-ping autocomplete.
- The minimal paired setup/join/ping workflow passes in Go and Rust.
- Native and Linux CGO vet/race checks, including the WebSocket fork, pass.
- All five isolated Go/Rust HTTP workflows still pass.
- Asset comparison checks all 315 paths and bodies plus import-map and stylesheet tags. Only the
  declared Turbo override differs; its fingerprint and embedded bytes are checked too.

`checks/`, `contracts.json.gz` and `metadata.json` retain logs and source/binary identities. Run
`bin/check-browser.mjs` with the `GO_BINARY`, `PLAYWRIGHT_ROOT` and
`PLAYWRIGHT_BROWSERS_PATH` settings described in the main README. The deterministic regression
is part of that command, not a separate patched diagnostic script.

These results supersede the current-browser gap, not the earlier checkpoint's failed evidence.
They do not establish strict DOM/network inventory parity, Firefox/Safari support, or performance
improvements. No application timing was performed for this frontend-only change.
