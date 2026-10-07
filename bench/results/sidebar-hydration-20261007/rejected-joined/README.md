# Rejected flattened sidebar join

A follow-up experiment replaced the room query and scoped participant batch with one LEFT JOIN.
It preserved complete response bytes and passed focused database/web race tests. A second version
removed the position map and parsed each room timestamp only once, using contiguous membership
ordering. Nullable participant rows remained explicit; the production timestamp scanner was unchanged.

Fewer query/step calls did not justify repeated room-column decoding. On the original seed, the
second version's three interleaved pairs gave median throughput 17,366 → 18,469 requests/second,
but CPU/request stayed 140.13 → 139.89 µs and one pair regressed. Saturated p99 rose 5.387 → 5.599 ms.
The first draft's results are retained separately, not pooled with the revised implementation.

An expanded fixture added 100 users and 20 directs, each containing the viewer and eight other
participants (31 rooms/110 users total). With the same gzip, 16 clients, guest affinity, excluded
two-second warmups, three interleaved ten-second samples and exact decoded-body checks:

| Expanded sidebar | Scoped batch | Flattened join |
|---|---:|---:|
| Median requests/second | 4,734 | 3,719 |
| Median CPU µs/request | 470.42 | 602.22 |
| Median saturated p99 ms | 12.519 | 13.775 |

Every expanded-fixture pair regressed: throughput −21.4%, CPU/request +28.0%. The synthetic fixture
is a scalability probe, not a production workload claim. The full-response validation matters:
all participants must still contribute to group labels even though only four avatars render.

The join was rejected and `rooms.go` restored to `5f12ecd`. Production retains the narrow scoped
batch. The patch, raw samples, identities and seed-generation script preserve this experiment;
no driver, schema or correctness contracts were relaxed. No other work ran during timing.
