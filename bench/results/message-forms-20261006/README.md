# Focused message-form views — 2026-10-06

Production `96cb2df` removes unused full-message rendering from edit, attachment-edit, new-boost and boost-index pages. Baseline is `3b692dc` (report HEAD `e941366`). The benchmark extension is `446ea6b`.

The required inputs now have explicit ownership: edit reads attachment presentation or editable body; boost index reads current ordered boosts; new boost uses message identity and the freshly authenticated user. Shared attachment construction is extracted without changing its URL/preview/error behavior. Account/layout observations remain before hydration. Missing creators still suppress attachment/boost presentation, as in the previous Go and pinned Rust implementation. Editable errors retain the prior Go handling. No authorization/session cache, dependency, database schema or durability change is introduced.

## Matched local results

Public HTTP/1.1, 16 clients, three rotating repetitions, five-second samples after two-second warmups. Values are medians. This measures complete authenticated **documents**, not Turbo-Frame clicks; the pinned load generator sends no Turbo-Frame header.

| Route | Identity requests/s: baseline / focused / Rust | Gzip requests/s: baseline / focused / Rust |
|---|---:|---:|
| Boost index | 10,035 / 11,471 / 26,775 | 9,967 / 12,457 / 27,498 |
| Edit body | 12,941 / 16,332 / 29,600 | 12,966 / 17,362 / 28,638 |
| Edit attachment | 11,126 / 13,608 / 29,939 | 11,091 / 13,915 / 29,706 |
| New boost | 13,785 / 21,069 / 27,915 | 13,286 / 22,767 / 28,438 |

Gzip throughput improves about 25%, 34%, 25% and 71%, respectively. CPU/request falls 13%, 17%, 16% and 32%. Rust remains faster on all four routes; this is not parity.

| Route | Gzip CPU µs/request, baseline → focused | Gzip p99 ms, baseline → focused / Rust |
|---|---:|---:|
| Boost index | 299.93 → 260.83 | 9.351 → 7.999 / 2.099 |
| Edit body | 220.31 → 183.58 | 7.739 → 5.915 / 1.934 |
| Edit attachment | 231.94 → 195.76 | 9.575 → 8.439 / 2.041 |
| New boost | 217.91 → 148.85 | 6.819 → 4.767 / 1.957 |

Identity tails are mixed: boost-index p99 worsens 9.031 → 9.575 ms despite higher throughput; edit improves 7.963 → 7.143, attachment edit is essentially unchanged 9.303 → 9.295, new boost improves 6.823 → 5.955. End-of-four-route PSS MiB is baseline/focused/Rust 53.4/52.6/48.6 identity and 52.6/52.8/48.7 gzip. No broad memory reduction is claimed, and this population differs from the six-route/Cable suites.

## Correctness and evidence

**18 application runs, 72 HTTP samples, zero errors.** Baseline and focused Go have identical full decoded form bytes after replacing only each process's known listener origin. No nonce/token masking is performed. Identity and decoded selected-encoding bodies match exactly within each process. Go initial body sizes are 29,566 / 23,461 / 22,755 / 21,938 bytes in the table's route order; Rust 30,246 / 23,657 / 22,899 / 22,242. Cross-language comparison uses seed-grounded form contracts, not byte equality.

Validation checks full document layout/completion, editor values, attachment presentation, required form fields, deletion controls and ordered boost IDs/booster identities/content. Six seed boosts include a deactivated user's existing boost. The untouched seed provides a real 39-byte `launch-notes.txt` attachment. Live Go/Rust snapshots rejected 18 corrupt-response probes: bare fragments, truncated documents and missing boost-delete controls. `validation/` retains those probes; its 0.1-second samples are verification only, excluded from performance medians.

Tests compare the previous full-hydration path's complete template bytes with the narrow views for normal/malformed/emoji bodies, text/image attachments and an orphan creator with boosts/attachment. HTTP tests verify fresh boost additions/deletions and membership revocation. Native `bin/check`, Linux vet/race and the local WebSocket fork pass. Read-only safety reviews finish before timing. No agents, builds or tests run during timing.

`matched/` contains raw JSON, full form snapshots, environment metadata and gzip-compressed server/load-generator logs. `checks/` retains gates. `manifest.json` records production/harness commits, exact staged source and executable hashes; source/binary remain at their local `.cache/` paths. `application.py` and `message_forms.py` are the exact measured harness/validator. No writes are benchmarked in this read-only suite; it does not establish message/FTS write performance or a Cable improvement.

## Environment and reproduction

Apple M3 Max's **OrbStack Linux ARM64 VM**, common Go 1.27.1/Rust 1.98.0 toolchain; Go SQLite 3.53.4/Rust SQLite 3.53.2. Four server workers on CPUs 0–3; load generator CPUs 4–7. Identical seed, logging enabled, fresh login and isolated database/storage per run. Logs stay in container `/tmp`, are gzip level 9 compressed after timing, then copied out. No host changes or upstream writes.

Repeat for `--gzip 0` and `1`:

```sh
python3 bench/application --rust-root reference \
  --seed .cache/parity-seeds/default --loadgen .cache/linux/loadgen \
  --candidate base=.cache/linux/final-render-micro \
  --candidate forms=.cache/linux/final-form-views \
  --candidate rust=.cache/rust-linux-target/release/campfire \
  --apps base forms rust --listener public --gzip 1 \
  --routes message_edit message_edit_attachment new_boost boosts_index \
  --same-html-apps base forms --concurrency 16 --reps 3 --seconds 5 \
  --server-cpus 0-3 --loadgen-cpus 4-7 --cable-clients --upload-reps 0 \
  --out /tmp/message-forms
```

These opt-in routes do not change the harness's default workload population. Short samples, one seed/admin user and virtualization constrain generalization. Native Fedora performance remains unavailable; no fresh Ruby/Elixir comparison was made. Do not transfer VM ratios to native/published AMD measurements or multiply gains from earlier phases.
