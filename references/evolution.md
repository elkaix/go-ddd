# Evolving an existing system

## Start with one business problem

Trace one command from its entry point through validation, authorization, domain decisions, storage, and external effects. Identify the precise pain: duplicated rules, ambiguous ownership, broken invariants, coupling, costly changes, or a measured operational bottleneck. Preserve existing behavior with targeted characterization tests where that behavior is intentional.

Separate a confirmed correctness fix from a proposed redesign. A folder called `services` or a struct with exported fields is evidence to investigate, not automatically a defect. Show the path by which the arrangement permits an invalid outcome or makes the requested change harder.

## Migrate incrementally

1. Establish the intended rule and its acceptance examples; distinguish current behavior from approved changes.
2. Introduce a small domain operation behind the current API, preserving request/response contracts unless the task includes changing them.
3. Move rule ownership into the new boundary and adapt existing persistence/providers behind focused ports.
4. Switch one use case to the boundary, verify behavior and failure handling, then remove the redundant rule path.
5. Expand only as subsequent use cases justify it; document the decision if the project maintains architectural records.

For schema changes, plan compatible expansion, resumable backfill, verification, cutover, and eventual contraction. Identify which application versions can coexist. Preserve identities, uniqueness, audit history, and external references. Avoid uncoordinated dual writes; explicitly identify the authoritative writer during a transition.

If rollback would encounter incompatible data or irreversible effects, describe the practical forward-recovery plan. Include rollout signals tied to the change, such as conflict rate, stuck operations, reconciliation backlog, or latency. A code rollback is not automatically a data rollback.

## Choose additional patterns deliberately

| Option | Evidence that may justify it | Cost to account for |
|---|---|---|
| Modular monolith | Shared deployment remains workable; business boundaries need clarity | Enforcing package/data ownership within one deployment |
| Separate service | Independent ownership, scaling, isolation, or deployment is demonstrably valuable | Network failure, separate operations, contract evolution, distributed consistency |
| Separate read model | Read shape or load materially differs from write-side rules | Projection lag, stale reads, rebuilds, and authorization on projections |
| Event sourcing | Historical events must be the authoritative state with useful replay semantics | Event evolution, replay safety, snapshots, privacy, recovery, and query projections |
| Saga/process manager | A business operation spans autonomous transaction boundaries | Durable orchestration, unknown outcomes, compensation, and intervention |

Keep CQRS and event sourcing separate decisions. Distinct command/query functions do not require separate databases. An outbox is a delivery mechanism, not an event store. Adopt specialized patterns for the affected context only; a whole application need not share one persistence model. See [Fowler on CQRS](https://martinfowler.com/bliki/CQRS.html).

## Design handoff

Return the requested level of detail: observed problem and evidence, chosen boundary, owned data and invariants, contracts, consistency and failure behavior, first migration slice, compatibility plan, and validation. Record rejected alternatives only when the tradeoff matters. State any policy decision the business must supply without blocking unrelated implementation work.
