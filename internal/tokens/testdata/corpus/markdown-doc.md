# Upstream timeout runbook

**Severity:** page during business hours, ticket overnight.
**Owner:** payments-platform. **Last reviewed:** 2031-04-02.

## Symptom

The `router_upstream_errors_total{upstream="catalogue"}` counter climbs while
`router_request_duration_seconds` p99 crosses the 900 ms budget. Clients see
`504 Gateway Timeout` with a `x-request-id` header they can quote back to you.

## First checks

1. Confirm the upstream is actually slow, not merely reported slow:

   ```sh
   curl -sS -o /dev/null -w '%{time_total}\n' \
     https://catalogue.internal.example.com:9443/livez
   ```

2. Look at the circuit breaker state. If it is `open`, the router is already
   shedding and the alert is telling you about recovery, not about a new fault.
3. Check queue depth. A backlog with a *falling* drain rate means the fault is
   downstream of the queue, not in the router.

## What usually causes it

| Cause | Signal | Fix |
| --- | --- | --- |
| Catalogue redeploy | error burst ends in ~90 s | wait, then close the circuit |
| Connection pool exhaustion | `pool.open == pool.max` | raise `max_conns`, restart |
| Slow query in catalogue | upstream p99 up, ours flat | escalate to catalogue owners |
| Network partition | `livez` fails from two zones | escalate to the network team |

> **Do not** raise the timeout budget as a first move. A longer budget converts
> a fast failure into a slow one and pushes the backlog upstream, where it is
> harder to see and much harder to drain.

## Recovery

- If the circuit is open and probes succeed, do nothing; it closes on its own.
- If probes fail for more than ten minutes, disable the upstream explicitly:

  ```sh
  router-admin upstream disable catalogue --reason 'runbook 0004'
  ```

- Re-enable only after the owning team confirms in the incident channel.

## Aftermath

Write the timeline into the incident document *before* you close the page.
Include the exact minute the circuit opened and the exact minute it closed;
those two numbers are what the next reader needs, and they are the two nobody
remembers an hour later.

See also: [queue backlog](queue-backlog.md), [ADR 0014](../adr/0014-drop-shared-session-state.md).
