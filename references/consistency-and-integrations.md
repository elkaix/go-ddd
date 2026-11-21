# Consistency and integration decisions

## Local transaction ownership

Define the application operation that owns begin, commit, and rollback. Pass a transaction-bound repository or unit-of-work scope so every participating statement uses the same transaction. Keep transactions short and keep remote calls outside their lock-holding lifetime when possible. A rollback cannot undo a remote provider effect.

Protect concurrent writes with an appropriate database mechanism: an expected-version update, row lock, unique constraint, conditional reservation, or a justified isolation level. Check the affected-row count and distinguish stale version from missing identity where the contract requires it. Enforce race-sensitive rules at the persistence boundary as well as in the model.

Use the actual driver's documented transaction semantics. For `database/sql`, a deferred rollback provides cleanup and a successful commit establishes success; mixing `DB` calls with `Tx` calls bypasses the transaction. Treat a commit error as an outcome to classify before retrying a whole operation. See [Go's transaction guidance](https://go.dev/doc/database/execute-transactions). Cancellation and rollback behavior varies by driver, so verify it rather than generalizing from another API.

## Idempotent commands

Give retried writes a stable operation identity before dispatch. Scope it to the relevant tenant, business operation, and resource; include actor scope where required. Persist a request fingerprint so reuse of a key with different intent is rejected. Atomically claim the key with the local state change, record pending/completed/unknown outcomes as needed, and return the stored result for a matching completed request. Define retention in relation to the retry/reconciliation horizon.

Concurrent requests with the same key must converge through durable uniqueness or equivalent coordination. A process mutex or a preflight SELECT alone is insufficient across workers. An idempotency key in your database does not automatically deduplicate a provider call: verify the provider's supported mechanism, key lifetime, and lookup/reconciliation capabilities.

## Outbox and consumers

When a committed local change must cause an external event, record the change and outbox entry in the same local transaction. Publish from committed rows using a recoverable worker, acknowledge after delivery, and expect duplicate publication around crashes. Use stable event IDs, aggregate identity/version, contract version, and tenant scope where applicable.

At each consumer, coordinate duplicate detection and the local effect in one transaction. Define per-aggregate ordering or explicit stale-event handling when order matters. If the consumer makes an external call, it needs its own idempotency and recovery plan; an inbox row alone does not make that call atomic. Track retry age, failed jobs, replay state, and operator recovery. This is not an end-to-end exactly-once guarantee. Source: [AWS transactional outbox guidance](https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/transactional-outbox.html).

## Remote failures and retries

Classify the actual outcome before selecting a retry. A timeout or lost response may follow a completed external write. Preserve the same operation identity and enter a reconcilable unknown state; use supported status lookup, verified webhook processing, or an operator path rather than starting a new charge or order blindly.

Use a bounded overall deadline, attempt limit, backoff with jitter, and an explicit retry owner. Honor provider rate-limit semantics. Retry transient outcomes only when the operation is safe to repeat; validation, authorization, and permanent business rejection need another response. Avoid multiplied retries across client, adapter, queue, and orchestrator layers. Source: [AWS timeouts and retries guidance](https://aws.amazon.com/builders-library/timeouts-retries-and-backoff-with-jitter/).

## Sagas and compensation

Use a durable process manager only when a real multi-step business operation spans independent transaction boundaries. Persist workflow identity, step status, external references, and recovery decisions. After failure, stop scheduling dependent forward steps. Reconcile unknown side effects before deciding what completed.

Define compensation in business terms and in dependency order; reversing successful steps is a useful sequential starting point, not a universal law. Some actions are irreversible, some require a refund rather than erasure, and some need an explicit human decision. Make compensations idempotent and resumable; persist compensation failures and expose an actionable recovery state. Do not overwrite newer concurrent changes to restore an old snapshot. Source: [Azure compensating transactions](https://learn.microsoft.com/en-us/azure/architecture/patterns/compensating-transaction).

## Contract boundaries

Translate vendor payloads through an anti-corruption layer with explicit field, enum, identity, and error mappings. Version public APIs/events, document required versus optional fields, and test unsupported/unknown values. Authenticate and validate callbacks using the provider's current contract, then enforce tenant/resource scope. Keep provider credentials in adapters and credential services, not domain entities.

Finish with a failure table for each external write: before dispatch, acknowledged success, explicit rejection, timeout, duplicate request, local commit failure, duplicate callback, and delayed callback. State the durable evidence and next action for every applicable row.
