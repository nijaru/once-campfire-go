# Fixed-rate HTTP latency

Build with `go build -o httprate ./bench/httprate`, then supply the base URL and an authenticated cookie. For example:

```sh
./httprate --base http://127.0.0.1:3000 --cookie "$COOKIE" \
  --path /users/me/sidebar --rate 3000 --duration 8 --pacing active
```

The default `--pacing timer` uses runtime timers. Sub-millisecond deadlines can coalesce into bursts. `--pacing active` spins until each deadline and yields after enqueueing so workers can send requests. It uses roughly one generator CPU in addition to request workers; allocate client CPUs separately from the server. The output records the selected mode. Do not mix the two modes in a comparison or treat a pacing improvement as an application improvement.

Arrivals follow absolute deadlines independently of response completion. The worker count and queue are bounded; drops, expired work and unscheduled arrivals remain in the accounting. Scheduled latency includes generator lateness and queueing; service latency begins when a worker takes a request. Neither measures server-only processing time. Precise enqueue times alone do not prove precise network arrivals.

The client consumes complete encoded response bodies and checks transport errors, status and nonempty successful bodies. It does not validate response semantics or persisted writes. Use the shared Campfire contracts and independent database audits for those checks. `--help` lists POST, timeout and drain options.
