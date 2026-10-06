# Native Intel final comparison — 2026-10-06

Desktop access returned after the VM trials. This follow-up uses the **physical Intel Core i9-13900KF**, not OrbStack and not the published AMD host. At 16 clients with gzip, retained Go has higher throughput than pinned Rust on the five warm read workloads, but worse p99, higher end-of-core-suite memory and slower posts. This is not universal parity.

Retained Go production is `7afad06` (message-form DSA `96cb2df` plus two fixed regex hoists); harness `63ad07a`. The form baseline is `3b692dc`. Rust remains pinned at `64f86353021145b63849fb1cd93adeb08f3b8dbb`, Rails at `90b330024dec3e757c79b6a7e6568f93da8e3148`. The rejected timestamp/avatar preparation is absent.

## Warm/core workload

Public HTTP/1.1, gzip, 16 clients, three alternating repetitions, **30-second samples** after two-second warmups. Each run has an isolated copy of the same seed and a fresh authenticated session. Values are medians.

| Workload | Go requests/s | Rust requests/s | Go / Rust p99 ms | Go / Rust CPU µs/request |
|---|---:|---:|---:|---:|
| Room | 32,348 | 29,693 | 1.532 / 0.947 | 116.69 / 129.09 |
| Active room | 37,429 | 28,057 | 1.522 / 1.064 | 102.04 / 136.57 |
| History | 39,671 | 31,869 | 1.213 / 0.836 | 97.27 / 119.49 |
| Search | 31,090 | 25,459 | 1.452 / 1.264 | 121.23 / 128.03 |
| Active search | 12,267 | 8,517 | 5.667 / 4.093 | 278.50 / 424.80 |
| Posts | 3,718 | 5,670 | 21.135 / 11.543 | 613.02 / 501.39 |

Go read throughput is about 9–44% higher; posts about 34% lower. This suite has no pre-optimization Go control, so it does not attribute these differences to the last micro pass. Go's smaller HTML and differing runtime/SQLite versions matter. Initial Go/Rust room, history and search documents are 374,036/416,139, 342,444/383,844 and 135,497/149,625 bytes. End-of-core-suite PSS is **210.3 / 150.2 MiB**, including allocator high-water effects and the active/write workload; it is not live-client steady-state memory.

## Edit/boost follow-up

The same affinity/encoding/concurrency/sample duration, with three rotating baseline/retained/Rust repetitions. Documents are complete authenticated GETs; no Turbo-Frame header is sent.

| Form | Baseline → retained Go requests/s | Rust requests/s | Baseline → retained / Rust p99 ms |
|---|---:|---:|---:|
| Boost index | 13,103 → 16,056 | 28,184 | 4.031 → 3.533 / 0.990 |
| Edit body | 18,351 → 22,649 | 30,131 | 2.887 → 2.591 / 0.892 |
| Edit attachment | 16,600 → 20,149 | 31,508 | 3.031 → 2.767 / 0.852 |
| New boost | 17,765 → 28,469 | 28,885 | 2.979 → 2.257 / 0.964 |

The focused views improve native throughput about **21–60%**, with lower CPU/request and p99 on these four routes. New-boost throughput approaches Rust here, not its latency; other forms remain slower. End-of-form-suite baseline/retained/Rust PSS is 54.7/53.8/51.9 MiB. Do not compare that population with the core suite.

## Validation and environment

**15 application runs, 72 HTTP samples, zero errors. 913,923 acknowledged HTTP posts and 7,680 active posts—921,603 total—are verified in messages and FTS.** Every response is fully consumed. Form Go pairs require identical full decoded bytes after replacing only their known listener origins; no nonce/token/cursor/content masking. Each process independently requires exact identity/decoded-gzip equivalence. Rust uses the same seed-grounded complete-document/form controls, including ordered existing boosts from a deactivated user. Cross-language checks are semantic, not byte equality.

Fedora 44/Linux on the i9-13900KF; user-owned rootless Podman containers, **no VM**. The common builder has Go 1.27.1 and Rust 1.98.0; Go SQLite 3.53.4/Rust SQLite 3.53.2. Four application workers run on CPUs **0,2,4,6**, the load generator on **8,10,12,14**. `lscpu` confirms eight distinct physical P-cores: no E-cores or SMT siblings in either set. Logging remains enabled; no host tuning/settings changes. Container dependency installation is confined to the benchmark images.

Native Linux vet/race and the WebSocket fork pass before timing; host `bin/check` and OrbStack gates also pass. Reviewed production changes and benchmark contracts are unchanged for the native rebuild. No agents, builds or tests run during timing. Bulk logs stay in container `/tmp`, then are gzip level 9 compressed before copying out. `checks/` retains the builder recipes and native build/test log. `manifest.json` records executable hashes, exact verified source archive hashes, image ID and native kernel. Archives/ELFs remain at the recorded desktop paths; the source and binary hashes were checked against the measured metadata after timing.

`forms/` and `core/` retain raw samples, full form/room captures, environment metadata and server/load-generator logs. `application.py` and `message_forms.py` are the exact harness/validator. `validation/positive/` contains excluded 0.1-second smoke probes. The first setup probe failed before starting a workload because extracted archives lacked Git metadata; it is retained in `validation/missing-git-metadata/`. Real shallow Git metadata at the exact source revisions was then supplied without changing source files. Archive xattr warnings did not affect compiled sources; hashes were verified.

## Scope and reproduction

Inside the native builder, repeat the core suite with:

```sh
python3 bench/application --rust-root reference \
  --seed .cache/parity-seeds/default --loadgen .cache/linux/loadgen \
  --candidate final=.cache/linux/final \
  --candidate rust=.cache/rust-target/release/campfire \
  --apps final rust --listener public --gzip 1 \
  --routes room_show active_room messages_page search active_search post_message \
  --concurrency 16 --reps 3 --seconds 30 \
  --server-cpus 0,2,4,6 --loadgen-cpus 8,10,12,14 \
  --cable-clients --upload-reps 0 --out /tmp/native-core
```

For forms, add `--candidate base=.cache/linux/render`, select `--apps base final rust`, use the four form routes and `--same-html-apps base final`.

The seed is small (169 initial messages); these are screen-specific local measurements, not production-load or language-wide claims. No identity/native 1/64-client sweep, native cold/churning-data suite, Ruby/Elixir rebuild, TLS/media/browser timing or matched Rust Cable trial was run. The earlier Go Cable index trial improved routing/throughput but **worsened saturated p99**; this HTTP follow-up neither fixes nor remeasures that trade-off. No VM ratio is transferred to this native host or to published AMD measurements.
