# Room involvement and sidebar publication ownership

This cohort completes a remaining room command/query/effect boundary against
`5b9bf44`. It is a verification record, not a throughput comparison.

- Involvement updates select the current active actor, membership, room metadata
  and previous setting inside the writer transaction. The receipt owns the
  previous and canonical new values; blank input still persists SQL NULL.
- GET settings select room and value in one observation. The room-index query
  moves out of HTTP without changing descending-ID navigation or no-room errors.
- `application.RoomPublications` owns bounded postcommit save/delete/visibility
  effects. Incoming request cancellation cannot suppress an already committed
  visibility publication. Cable's publication-time authorization is unchanged.
- Open-room publication remains global. Closed-room publication selects active
  recipients and renders identical shared markup once. Direct publication selects
  active recipients and retained participants together, preserving name-based
  recipient order, ID-based participant order, personalized names and self-pings.
  Inactive retained participants contribute names but are not recipients.
- `web/sidebar.go`, the old string involvement read/write methods and the obsolete
  active-participant query are removed. Runtime and test composition supply the
  concrete effect service; HTTP constructs no workers or resources.

The focused HTTP regression captures an active ping participant, deactivates that
user (whose direct membership remains), then invokes a write with the captured
identity. The original returns 302 and changes the setting; the candidate returns
404 and preserves it. This deliberate transaction-current authority correction is
recorded in the root README. A held-writer test queues another involvement command,
commits a new previous value/room name, and verifies the returned receipt plus NULL
storage while every reader is held. A real Cable subscriber receives the committed
visibility effect even with a canceled incoming context. An audience regression
checks retained inactive names and distinct recipient/participant ordering.

Affected native races repeat five times. Full native `bin/check`, fresh native and
Linux builds, Linux vet/database/application/web races, current shared browser
flows and actual public HTTP/write checks pass. The public probe verifies four
involvement PUT/GET transitions, exact SQL NULL clearing and original-value
restoration, twenty shared complete responses, seven preflights and four exact
acknowledged ID/body/room/FTS writes. Existing forwarding and coding probes still
pass. Shared assertions are unchanged; the disclosed browser persisted-echo wait
remains before exact-three. A read-only reviewer found no actionable P1/P2 issue;
its review is source-only, not independent test evidence.

No new application timing/profile population is measured here. Fewer queries and
closed-room renders are source facts, not established throughput gains. Previous
profiles and comparisons remain under their original identities; changed-workload
timing is unfinished. Retained/active resource budgets, public coding ownership,
other query/effect migrations and multi-user/fanout/matched-load acceptance remain.
No optimality or Rust parity is claimed. Effects remain process-only, not an
outbox/exactly-once guarantee. Native Intel is unavailable; forced CLI exit and
unrelated unchanged media are not rerun.

Raw gate logs, the original failure, source patch and exact identities are retained
alongside this record. The final gate runs after archiving to catch artifact errors.
