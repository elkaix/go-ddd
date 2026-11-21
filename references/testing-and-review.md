# Behavior verification and review

## Select checks from the risk

Derive assertions from business rules and public outcomes. Keep tests that would fail for the identified defect; avoid tests that merely repeat the implementation's control flow or verify architectural labels.

| Risk | Useful evidence |
|---|---|
| Invariant violation | Valid/invalid commands, zero values, boundaries, unchanged state after rejection |
| Incorrect value semantics | Equality, currency/unit mismatch, overflow, rounding, mutable alias tests |
| Invalid transition | Allowed and forbidden state pairs, repeat command semantics |
| Persistence mismatch | Round-trip historical state, missing rows, version conflicts, transaction rollback |
| Concurrent mutation | Two competing clients against actual persistence; uniqueness/reservation guarantees |
| Repeated command/event | Duplicate and simultaneous delivery; same key with different request; replay |
| External uncertainty | Timeout after effect, lost response, local commit failure, late callback |
| Recovery failure | Process restart between durable steps; resumed compensation/outbox work |
| Boundary leak | Invalid foreign payload, unknown enum, context translation, cross-tenant lookup |
| Contract drift | Consumer/adapter contract cases, backward-compatible API/event evolution |

Use pure unit tests for model rules, application tests for orchestration outcomes, and integration tests for database/adapter semantics. Fakes and mocks help isolate a rule but cannot establish real transaction isolation or provider guarantees. Prefer meaningful scenarios over universal coverage targets.

## Test execution

Inspect the project's commands and supported Go version first. Identify every affected module using its `go.mod`/workspace configuration; a root `go test ./...` does not automatically cover nested modules. Compile and run the relevant tests, vet/lint, and formatting checks required by that repository.

For changed concurrent code, use the [race detector](https://go.dev/doc/articles/race_detector) where supported, with a meaningful concurrent workload. It only observes executed in-process paths; a passing run does not prove freedom from database races, lost updates, or distributed duplicate effects.

Use [fuzzing](https://go.dev/doc/tutorial/fuzz) for parsers, arithmetic/value-object boundaries, and state transitions with strong properties. Keep runs bounded, seed known edge cases, and avoid network-dependent fuzz targets. Property examples include failure atomicity, preserved total/currency invariants, and valid serialization round trips.

The bundled `assets/order-example` is a separate module. From that directory, run:

```sh
go test ./...
go vet ./...
go test -race ./...
```

These are validation commands, not assertions that they have already passed. Verify that `go version` identifies the Go compiler before running them. Report a missing toolchain, unavailable database, or unsupported race runtime as a skipped/blocked check with its reason.

## Review findings

For each finding, provide severity, exact code location, triggering input/interleaving/failure, violated invariant, user-visible consequence, and the smallest defensible correction. Include a reproducer or regression scenario when feasible. Mark unverified assumptions and confidence; do not invent line numbers or claim execution from static inspection.

Separate correctness/security defects, maintainability observations, and optional architecture choices. Prioritize issues that can lose money/data, cross tenant boundaries, corrupt state, duplicate external effects, or prevent recovery. A DDD label alone is not evidence of safety or quality.

When scope is broad and agents are available, divide independent investigation by business flow or boundary, then merge duplicate findings and resolve contradictions against source. Preserve one owner for shared edits. Keep a final inventory of inspected and uninspected areas so the report's coverage is honest.

## Completion evidence

Report commands actually run, affected modules, observed outcomes, and material gaps. Connect the checks to the changed rules. Distinguish static inspection, executed tests, and proposed follow-up verification. Stop once the identified risk and required project gates are adequately covered.
